package market

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

const maxUnpacked = 50 << 20

type ghAsset struct {
	Name          string `json:"name"`
	URL           string `json:"browser_download_url"`
	DownloadCount int    `json:"download_count"`
}

type ghRelease struct {
	Tag    string    `json:"tag_name"`
	Assets []ghAsset `json:"assets"`
}

// Release is the part of a GitHub release the marketplace uses.
type Release struct {
	Tag       string // as published, e.g. "v1.2.0"
	Version   string // Tag without leading "v"
	Downloads int    // summed asset download counts
}

func (c *Client) getRelease(ctx context.Context, owner, repo, tag string) (*ghRelease, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiBase(), owner, repo)
	if tag != "" {
		u = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", c.apiBase(), owner, repo, url.PathEscape(tag))
	}
	resp, err := c.do(ctx, u, map[string]string{"Accept": "application/vnd.github+json"}, maxRegistrySize)
	if err != nil {
		return nil, err
	}
	switch resp.Status {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: release of %s/%s", ErrNotFound, owner, repo)
	default:
		return nil, fmt.Errorf("github: HTTP %d for %s", resp.Status, u)
	}
	var rel ghRelease
	if err := json.Unmarshal(resp.Body, &rel); err != nil {
		return nil, fmt.Errorf("github: %w", err)
	}
	return &rel, nil
}

// LatestRelease returns the latest published release of owner/repo.
func (c *Client) LatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	rel, err := c.getRelease(ctx, owner, repo, "")
	if err != nil {
		return nil, err
	}
	r := &Release{Tag: rel.Tag, Version: strings.TrimPrefix(rel.Tag, "v")}
	for _, a := range rel.Assets {
		r.Downloads += a.DownloadCount
	}
	return r, nil
}

// FetchManifest reads manifest.toml (in subdir) of owner/repo at ref via the
// contents API. ref "" means the default branch.
func (c *Client) FetchManifest(ctx context.Context, owner, repo, ref, subdir string) (Manifest, error) {
	p := path.Join(subdir, "manifest.toml")
	u := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase(), owner, repo, p)
	if ref != "" {
		u += "?ref=" + url.QueryEscape(ref)
	}
	resp, err := c.do(ctx, u, map[string]string{"Accept": "application/vnd.github.raw"}, 1<<20)
	if err != nil {
		return Manifest{}, err
	}
	if resp.Status == http.StatusNotFound {
		return Manifest{}, fmt.Errorf("%w: %s at %s", ErrNotFound, p, ref)
	}
	if resp.Status != http.StatusOK {
		return Manifest{}, fmt.Errorf("github: HTTP %d for %s", resp.Status, u)
	}
	return ParseManifest(resp.Body)
}

// Resolve turns any reference (see ParseRef) into a downloaded Package.
func (c *Client) Resolve(ctx context.Context, ref string) (*Package, error) {
	r, err := ParseRef(ref)
	if err != nil {
		return nil, err
	}
	switch r.Kind {
	case RefURL:
		if strings.HasSuffix(strings.ToLower(strings.SplitN(r.URL, "?", 2)[0]), ".toml") {
			return c.resolveThemeURL(ctx, r.URL)
		}
		return c.resolveURL(ctx, r.URL)
	case RefGitHub:
		return c.resolveGitHub(ctx, r.Owner, r.Repo, r.Tag, "", "", nil)
	}
	reg, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range reg.Extensions {
		e := reg.Extensions[i]
		if e.ID != r.ID {
			continue
		}
		owner, repo, _ := strings.Cut(e.Repo, "/")
		if e.Kind == KindTheme {
			return c.resolveTheme(ctx, owner, repo, e)
		}
		pkg, err := c.resolveGitHub(ctx, owner, repo, e.Ref, e.Subdir, e.Asset, &e)
		if err != nil {
			return nil, err
		}
		if e.SHA256 != "" && !strings.EqualFold(e.SHA256, pkg.SHA256) {
			return nil, fmt.Errorf("%w: %s expected %s, got %s", ErrChecksum, e.ID, e.SHA256, pkg.SHA256)
		}
		if pkg.ID == "" {
			pkg.ID = e.ID
		}
		return pkg, nil
	}
	return nil, fmt.Errorf("%w: extension %q is not in the registry", ErrNotFound, r.ID)
}

func (c *Client) resolveURL(ctx context.Context, rawURL string) (*Package, error) {
	data, err := c.download(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	mf, err := inspectZip(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", rawURL, err)
	}
	return &Package{ID: mf.ID, Version: mf.Version, Source: rawURL, Zip: data,
		SHA256: sum256(data), URL: rawURL, Manifest: mf}, nil
}

func (c *Client) download(ctx context.Context, rawURL string) ([]byte, error) {
	resp, err := c.do(ctx, rawURL, map[string]string{"Accept": "application/octet-stream"}, c.maxBytes())
	if err != nil {
		return nil, err
	}
	if resp.Status != http.StatusOK {
		if resp.Status == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, rawURL)
		}
		return nil, fmt.Errorf("download %s: HTTP %d", rawURL, resp.Status)
	}
	return resp.Body, nil
}

// resolveGitHub prefers a release asset (.zagros, then .zip with a manifest)
// and falls back to the repository zipball of tag (default branch if empty).
func (c *Client) resolveGitHub(ctx context.Context, owner, repo, tag, subdir, assetGlob string, entry *Entry) (*Package, error) {
	source := Ref{Kind: RefGitHub, Owner: owner, Repo: repo, Tag: tag}
	var rel *ghRelease
	relErr := error(ErrNotFound)
	// Entries living in a subfolder of a repo with no pinned tag or asset glob
	// install from the default branch; the repo's own releases are unrelated.
	if subdir == "" || tag != "" || assetGlob != "" {
		rel, relErr = c.getRelease(ctx, owner, repo, tag)
	}
	if relErr == nil {
		source.Tag = rel.Tag
		for _, a := range orderAssets(rel.Assets, assetGlob) {
			data, err := c.download(ctx, a.URL)
			if err != nil {
				return nil, err
			}
			mf, err := inspectZip(data)
			if err != nil {
				continue // e.g. a source .zip without manifest; try the next candidate
			}
			return &Package{ID: mf.ID, Version: mf.Version, Source: source.String(), Zip: data,
				SHA256: sum256(data), URL: a.URL, Manifest: mf, Entry: entry}, nil
		}
	} else if !errorsIsNotFound(relErr) {
		return nil, relErr
	}
	// Fallback: repository zipball at tag / HEAD.
	zu := fmt.Sprintf("%s/repos/%s/%s/zipball", c.apiBase(), owner, repo)
	if source.Tag != "" {
		zu += "/" + url.PathEscape(source.Tag)
	}
	raw, err := c.download(ctx, zu)
	if err != nil {
		return nil, err
	}
	packed, mf, err := repack(raw, subdir)
	if err != nil {
		return nil, fmt.Errorf("github:%s/%s: %w", owner, repo, err)
	}
	if mf.Version == "" {
		mf.Version = strings.TrimPrefix(source.Tag, "v")
	}
	return &Package{ID: mf.ID, Version: mf.Version, Source: source.String(), Zip: packed,
		SHA256: sum256(raw), URL: zu, Manifest: mf, Entry: entry}, nil
}

func errorsIsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// orderAssets returns candidate assets: glob matches first (when a glob is
// given, only those), otherwise *.zagros before *.zip.
func orderAssets(assets []ghAsset, glob string) []ghAsset {
	var out []ghAsset
	if glob != "" {
		for _, a := range assets {
			if ok, _ := path.Match(glob, a.Name); ok {
				out = append(out, a)
			}
		}
		return out
	}
	for _, a := range assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".zagros") {
			out = append(out, a)
		}
	}
	for _, a := range assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".zip") {
			out = append(out, a)
		}
	}
	return out
}

func openZip(data []byte) (*zip.Reader, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid zip archive: %w", err)
	}
	var total uint64
	for _, f := range zr.File {
		total += f.UncompressedSize64
		if total > maxUnpacked {
			return nil, fmt.Errorf("archive expands beyond %d MB", maxUnpacked>>20)
		}
	}
	return zr, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(io.LimitReader(rc, maxUnpacked+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxUnpacked {
		return nil, fmt.Errorf("%s expands beyond limit", f.Name)
	}
	return b, nil
}

// inspectZip requires manifest.toml at the archive root.
func inspectZip(data []byte) (Manifest, error) {
	zr, err := openZip(data)
	if err != nil {
		return Manifest{}, err
	}
	for _, f := range zr.File {
		if f.Name == "manifest.toml" {
			b, err := readZipFile(f)
			if err != nil {
				return Manifest{}, err
			}
			return ParseManifest(b)
		}
	}
	return Manifest{}, fmt.Errorf("archive has no manifest.toml at its root")
}

func safeName(name string) bool {
	return name != "" && !strings.HasPrefix(name, "/") && !strings.Contains(name, `\`) &&
		!strings.Contains(name, ":") && !strings.Contains(path.Clean("/"+name), "..") &&
		!strings.Contains(name, "../")
}

// repack converts a GitHub zipball (single top-level directory) into a
// .zagros archive rooted at subdir, with manifest.toml at the root.
func repack(raw []byte, subdir string) ([]byte, Manifest, error) {
	zr, err := openZip(raw)
	if err != nil {
		return nil, Manifest{}, err
	}
	// Candidate base prefixes: "" (manifest at archive root) or the single top dir.
	var bases []string
	top := ""
	for _, f := range zr.File {
		first, _, found := strings.Cut(f.Name, "/")
		if !found {
			top = ""
			break
		}
		if top == "" {
			top = first
		} else if top != first {
			top = ""
			break
		}
	}
	if top != "" {
		bases = append(bases, top+"/")
	}
	bases = append(bases, "")
	subdir = strings.Trim(subdir, "/")
	for _, base := range bases {
		root := base
		if subdir != "" {
			root += subdir + "/"
		}
		var files []*zip.File
		hasManifest := false
		for _, f := range zr.File {
			if f.FileInfo().IsDir() || !strings.HasPrefix(f.Name, root) {
				continue
			}
			rel := strings.TrimPrefix(f.Name, root)
			if !safeName(rel) {
				return nil, Manifest{}, fmt.Errorf("unsafe path %q in archive", f.Name)
			}
			if rel == "manifest.toml" {
				hasManifest = true
			}
			files = append(files, f)
		}
		if !hasManifest {
			continue
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		var mf Manifest
		for _, f := range files {
			rel := strings.TrimPrefix(f.Name, root)
			b, err := readZipFile(f)
			if err != nil {
				return nil, Manifest{}, err
			}
			w, err := zw.Create(rel)
			if err != nil {
				return nil, Manifest{}, err
			}
			if _, err := w.Write(b); err != nil {
				return nil, Manifest{}, err
			}
			if rel == "manifest.toml" {
				if mf, err = ParseManifest(b); err != nil {
					return nil, Manifest{}, err
				}
			}
		}
		if err := zw.Close(); err != nil {
			return nil, Manifest{}, err
		}
		return buf.Bytes(), mf, nil
	}
	where := "the repository root"
	if subdir != "" {
		where = "subdir " + subdir
	}
	return nil, Manifest{}, fmt.Errorf("no manifest.toml found at %s", where)
}
