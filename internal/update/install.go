package update

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The installer works in two steps so a running Zagros can update itself.
// Stage downloads the release zip, checks it against the release's
// checksums.txt and unpacks zagros.exe into the data directory. Apply swaps
// that file in for the running executable (Windows lets a running exe be
// renamed, not overwritten); the caller then starts the new one.

const (
	maxZip      = 200 << 20 // download cap
	maxExe      = 300 << 20 // unpacked cap
	maxChecksum = 1 << 20
	exeName     = "zagros.exe"
)

// Staged is an update that is downloaded, verified and ready to apply.
type Staged struct {
	Version string    `json:"version"`
	Path    string    `json:"path"`   // the unpacked zagros.exe
	SHA256  string    `json:"sha256"` // of Path
	Staged  time.Time `json:"staged"`
}

func stageRoot(dataDir string) string { return filepath.Join(dataDir, "update") }

func stagedFile(dataDir string) string { return filepath.Join(stageRoot(dataDir), "staged.json") }

// Stage fetches release rel for goarch and unpacks it. It refuses a release
// that has no checksums.txt or whose zip does not match it.
func Stage(ctx context.Context, rel Release, goarch, dataDir string, o Options) (Staged, error) {
	asset, ok := AssetFor(rel, goarch)
	if !ok {
		return Staged{}, fmt.Errorf("release %s has no Windows %s zip", rel.Version, goarch)
	}
	var sums *Asset
	for i := range rel.Assets {
		if rel.Assets[i].Name == "checksums.txt" {
			sums = &rel.Assets[i]
		}
	}
	if sums == nil {
		return Staged{}, fmt.Errorf("release %s has no checksums.txt, not installing an unverified download", rel.Version)
	}
	repo := o.Repo
	if repo == "" {
		repo = Repo
	}
	hc := o.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Minute}
	}
	trusted := func(u string) bool {
		return o.BaseURL != "" || strings.HasPrefix(u, "https://github.com/"+repo+"/releases/download/")
	}
	if !trusted(asset.URL) || !trusted(sums.URL) {
		return Staged{}, errors.New("release assets are not on github.com, not installing")
	}
	sumText, err := download(ctx, hc, sums.URL, maxChecksum)
	if err != nil {
		return Staged{}, fmt.Errorf("checksums: %w", err)
	}
	want := checksumFor(string(sumText), asset.Name)
	if want == "" {
		return Staged{}, fmt.Errorf("checksums.txt has no entry for %s", asset.Name)
	}
	data, err := download(ctx, hc, asset.URL, maxZip)
	if err != nil {
		return Staged{}, fmt.Errorf("download: %w", err)
	}
	if got := sum(data); got != want {
		return Staged{}, fmt.Errorf("%s does not match checksums.txt (got %s...)", asset.Name, got[:12])
	}
	exe, err := unzipExe(data)
	if err != nil {
		return Staged{}, err
	}
	dir := filepath.Join(stageRoot(dataDir), rel.Version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Staged{}, err
	}
	path := filepath.Join(dir, exeName)
	if err := os.WriteFile(path+".part", exe, 0o755); err != nil {
		return Staged{}, err
	}
	if err := os.Rename(path+".part", path); err != nil {
		return Staged{}, err
	}
	st := Staged{Version: rel.Version, Path: path, SHA256: sum(exe), Staged: time.Now()}
	b, _ := json.Marshal(st)
	if err := os.WriteFile(stagedFile(dataDir), b, 0o644); err != nil {
		return Staged{}, err
	}
	pruneStaged(dataDir, rel.Version)
	return st, nil
}

// LoadStaged returns the staged update when it is newer than current and its
// file is intact. A development build (not a release version) never has one.
func LoadStaged(dataDir, current string) (Staged, bool) {
	b, err := os.ReadFile(stagedFile(dataDir))
	if err != nil {
		return Staged{}, false
	}
	var st Staged
	if json.Unmarshal(b, &st) != nil || st.Path == "" {
		return Staged{}, false
	}
	if c, ok := CompareVersions(st.Version, current); !ok || c <= 0 {
		if ok {
			ClearStaged(dataDir) // already running it, or something newer
		}
		return Staged{}, false
	}
	data, err := os.ReadFile(st.Path)
	if err != nil || sum(data) != st.SHA256 {
		return Staged{}, false
	}
	return st, true
}

// ClearStaged deletes every staged download.
func ClearStaged(dataDir string) { _ = os.RemoveAll(stageRoot(dataDir)) }

func pruneStaged(dataDir, keep string) {
	ents, _ := os.ReadDir(stageRoot(dataDir))
	for _, e := range ents {
		if e.IsDir() && e.Name() != keep {
			_ = os.RemoveAll(filepath.Join(stageRoot(dataDir), e.Name()))
		}
	}
}

// Apply replaces the executable at exe with the staged one. The old file is
// kept as exe+".old" (CleanupOld removes it on the next start) and restored
// when the new one cannot be written.
func Apply(st Staged, exe string) error {
	data, err := os.ReadFile(st.Path)
	if err != nil {
		return fmt.Errorf("staged update: %w", err)
	}
	if sum(data) != st.SHA256 {
		return errors.New("staged update changed on disk, not installing")
	}
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("cannot replace %s (is the folder writable?): %w", exe, err)
	}
	if err := os.WriteFile(exe, data, 0o755); err != nil {
		_ = os.Remove(exe)
		if rerr := os.Rename(old, exe); rerr != nil {
			return fmt.Errorf("update failed (%v) and the old version could not be restored: %w", err, rerr)
		}
		return fmt.Errorf("writing the new version: %w", err)
	}
	return nil
}

// CleanupOld removes the executable Apply set aside.
func CleanupOld(exe string) { _ = os.Remove(exe + ".old") }

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// checksumFor reads "<hex>  <name>" lines (sha256sum output).
func checksumFor(text, name string) string {
	for _, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name && len(f[0]) == 64 {
			return strings.ToLower(f[0])
		}
	}
	return ""
}

func download(ctx context.Context, hc *http.Client, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "zagros-update")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server answered %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("larger than %d MB", limit>>20)
	}
	return b, nil
}

// unzipExe returns zagros.exe from the release zip, wherever it sits.
func unzipExe(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("release zip: %w", err)
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !strings.EqualFold(filepath.Base(f.Name), exeName) {
			continue
		}
		if f.UncompressedSize64 > maxExe {
			return nil, errors.New("zagros.exe in the release zip is too large")
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, maxExe+1))
		if err != nil {
			return nil, err
		}
		if len(b) < 2 || b[0] != 'M' || b[1] != 'Z' {
			return nil, errors.New("zagros.exe in the release zip is not a Windows executable")
		}
		return b, nil
	}
	return nil, errors.New("the release zip has no zagros.exe")
}
