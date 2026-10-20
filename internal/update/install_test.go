package update

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func releaseZip(t *testing.T, exe string) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for name, body := range map[string]string{"README.md": "hi", "zagros.exe": exe} {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// fakeRelease serves a zip and a checksums file; badSum corrupts the entry.
func fakeRelease(t *testing.T, zipData []byte, badSum, noSums bool) (Release, Options) {
	t.Helper()
	name := "zagros_9.9.9_windows_amd64.zip"
	sums := sum(zipData) + "  " + name + "\n"
	if badSum {
		sums = strings.Repeat("0", 64) + "  " + name + "\n"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/zip", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(zipData) })
	mux.HandleFunc("/sums", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(sums)) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	rel := Release{Tag: "v9.9.9", Version: "9.9.9", Assets: []Asset{{Name: name, URL: srv.URL + "/zip"}}}
	if !noSums {
		rel.Assets = append(rel.Assets, Asset{Name: "checksums.txt", URL: srv.URL + "/sums"})
	}
	return rel, Options{BaseURL: srv.URL}
}

func TestStageVerifiesAndUnpacks(t *testing.T) {
	data := t.TempDir()
	rel, o := fakeRelease(t, releaseZip(t, "MZ-new-build"), false, false)
	st, err := Stage(context.Background(), rel, "amd64", data, o)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(st.Path); string(b) != "MZ-new-build" || st.Version != "9.9.9" {
		t.Fatalf("staged %+v content %q", st, b)
	}
	if got, ok := LoadStaged(data, "0.8.1"); !ok || got.Version != "9.9.9" {
		t.Fatalf("LoadStaged = %+v %v", got, ok)
	}
	if _, ok := LoadStaged(data, "dev"); ok {
		t.Fatal("a dev build must not get a staged update")
	}
	if _, ok := LoadStaged(data, "9.9.9"); ok {
		t.Fatal("already running it")
	}
	if _, err := os.Stat(stagedFile(data)); err == nil {
		t.Fatal("staged files of a version we run must be removed")
	}
}

func TestStageRefusesBadDownloads(t *testing.T) {
	for name, c := range map[string]struct {
		zip           []byte
		badSum, noSum bool
		want          string
	}{
		"wrong checksum": {releaseZip(t, "MZ-x"), true, false, "does not match"},
		"no checksums":   {releaseZip(t, "MZ-x"), false, true, "checksums.txt"},
		"not an exe":     {releaseZip(t, "text"), false, false, "not a Windows executable"},
	} {
		rel, o := fakeRelease(t, c.zip, c.badSum, c.noSum)
		_, err := Stage(context.Background(), rel, "amd64", t.TempDir(), o)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want %q", name, err, c.want)
		}
	}
}

func TestStageOnlyFromGitHub(t *testing.T) {
	rel := Release{Version: "9.9.9", Assets: []Asset{
		{Name: "zagros_9.9.9_windows_amd64.zip", URL: "https://evil.example/z.zip"},
		{Name: "checksums.txt", URL: "https://evil.example/sums"}}}
	if _, err := Stage(context.Background(), rel, "amd64", t.TempDir(), Options{}); err == nil || !strings.Contains(err.Error(), "github.com") {
		t.Fatalf("err = %v", err)
	}
}

func TestApplySwapsAndRestores(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "zagros.exe")
	_ = os.WriteFile(exe, []byte("MZ-old"), 0o755)
	staged := filepath.Join(dir, "new.exe")
	_ = os.WriteFile(staged, []byte("MZ-new"), 0o755)
	st := Staged{Version: "9.9.9", Path: staged, SHA256: sum([]byte("MZ-new"))}
	if err := Apply(st, exe); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "MZ-new" {
		t.Fatalf("exe = %q", b)
	}
	if b, _ := os.ReadFile(exe + ".old"); string(b) != "MZ-old" {
		t.Fatalf("old = %q", b)
	}
	CleanupOld(exe)
	if _, err := os.Stat(exe + ".old"); err == nil {
		t.Fatal(".old left behind")
	}
	// a tampered staged file is refused and the exe is untouched
	_ = os.WriteFile(staged, []byte("MZ-evil"), 0o755)
	if err := Apply(st, exe); err == nil {
		t.Fatal("tampered file accepted")
	}
	if b, _ := os.ReadFile(exe); string(b) != "MZ-new" {
		t.Fatalf("exe changed to %q", b)
	}
	// a missing executable cannot be replaced
	if err := Apply(Staged{Path: staged, SHA256: sum([]byte("MZ-evil"))}, filepath.Join(dir, "none.exe")); err == nil {
		t.Fatal("replacing a missing exe should fail")
	}
}

func TestChecksumFor(t *testing.T) {
	h := strings.Repeat("ab", 32)
	if got := checksumFor("junk\n"+h+"  a.zip\n"+strings.ToUpper(h)+" *b.zip\n", "b.zip"); got != h {
		t.Fatalf("got %q", got)
	}
	if checksumFor(h+"  a.zip", "c.zip") != "" {
		t.Fatal("matched the wrong name")
	}
}
