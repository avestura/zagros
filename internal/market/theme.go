package market

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/avestura/zagros/internal/theme"
)

// MaxThemeSize caps a theme file. Themes are a few hundred bytes in practice.
const MaxThemeSize = 64 << 10

// ValidateTheme checks that data is a theme file Zagros can load.
func ValidateTheme(data []byte, fallbackName string) (theme.Theme, error) {
	if len(data) == 0 {
		return theme.Theme{}, fmt.Errorf("theme file is empty")
	}
	if len(data) > MaxThemeSize {
		return theme.Theme{}, fmt.Errorf("theme file is %d bytes, limit is %d", len(data), MaxThemeSize)
	}
	t, err := theme.Parse(data, fallbackName)
	if err != nil {
		return theme.Theme{}, fmt.Errorf("invalid theme: %w", err)
	}
	return t, nil
}

// fetchRaw reads one file of owner/repo at ref through the contents API.
func (c *Client) fetchRaw(ctx context.Context, owner, repo, ref, p string, limit int64) ([]byte, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiBase(), owner, repo, strings.Trim(p, "/"))
	if ref != "" {
		u += "?ref=" + url.QueryEscape(ref)
	}
	resp, err := c.do(ctx, u, map[string]string{"Accept": "application/vnd.github.raw"}, limit)
	if err != nil {
		return nil, err
	}
	switch resp.Status {
	case http.StatusOK:
		return resp.Body, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s at %s", ErrNotFound, p, ref)
	}
	return nil, fmt.Errorf("github: HTTP %d for %s", resp.Status, u)
}

// themeBytes fetches the theme file an entry points to: a release asset
// matching e.Asset, or the file at e.Subdir.
func (c *Client) themeBytes(ctx context.Context, owner, repo string, e Entry) (data []byte, from string, err error) {
	if e.Asset != "" {
		rel, err := c.getRelease(ctx, owner, repo, e.Ref)
		if err != nil {
			return nil, "", err
		}
		for _, a := range orderAssets(rel.Assets, e.Asset) {
			b, err := c.download(ctx, a.URL)
			if err != nil {
				return nil, "", err
			}
			return b, a.URL, nil
		}
		return nil, "", fmt.Errorf("%w: no release asset matches %q", ErrNotFound, e.Asset)
	}
	b, err := c.fetchRaw(ctx, owner, repo, e.Ref, e.Subdir, MaxThemeSize+1)
	return b, e.Subdir, err
}

func (c *Client) resolveTheme(ctx context.Context, owner, repo string, e Entry) (*Package, error) {
	data, from, err := c.themeBytes(ctx, owner, repo, e)
	if err != nil {
		return nil, err
	}
	sum := sum256(data)
	if e.SHA256 != "" && !strings.EqualFold(e.SHA256, sum) {
		return nil, fmt.Errorf("%w: %s expected %s, got %s", ErrChecksum, e.ID, e.SHA256, sum)
	}
	t, err := ValidateTheme(data, e.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", e.ID, err)
	}
	return &Package{Kind: KindTheme, Theme: data, ID: e.ID, Version: e.Latest, Source: e.ID,
		SHA256: sum, URL: from, Entry: &e, Manifest: Manifest{ID: e.ID, Name: t.Name, Author: t.Author}}, nil
}

func (c *Client) buildThemeEntry(ctx context.Context, s Spec, e Entry, owner, repo string) (Entry, error) {
	e.Kind = KindTheme
	data, _, err := c.themeBytes(ctx, owner, repo, e)
	if err != nil {
		return e, fmt.Errorf("%s: %w", s.ID, err)
	}
	t, err := ValidateTheme(data, s.ID)
	if err != nil {
		return e, fmt.Errorf("%s: %w", s.ID, err)
	}
	if s.Name == "" {
		e.Name = t.Name
	}
	if s.Author == "" && t.Author != "" {
		e.Author = t.Author
	}
	e.Permissions = []string{}
	return e, ValidateEntry(e)
}

// resolveThemeURL handles a direct https URL to a .toml theme file.
func (c *Client) resolveThemeURL(ctx context.Context, rawURL string) (*Package, error) {
	data, err := c.download(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(rawURL)
	id := strings.ToLower(strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path)))
	if !idRe.MatchString(id) {
		return nil, fmt.Errorf("cannot derive a theme id from %s", rawURL)
	}
	t, err := ValidateTheme(data, id)
	if err != nil {
		return nil, err
	}
	return &Package{Kind: KindTheme, Theme: data, ID: id, Source: rawURL, SHA256: sum256(data),
		URL: rawURL, Manifest: Manifest{ID: id, Name: t.Name, Author: t.Author}}, nil
}

// InstallTheme resolves ref (a registry id of a theme entry, or an https URL
// to a .toml file), validates the theme and writes it to themeDir/<id>.toml.
func (c *Client) InstallTheme(ctx context.Context, themeDir, ref string) (*Package, error) {
	pkg, err := c.Resolve(ctx, ref)
	if err != nil {
		return nil, err
	}
	if pkg.Kind != KindTheme {
		return nil, fmt.Errorf("%s is an extension, not a theme", ref)
	}
	if !idRe.MatchString(pkg.ID) {
		return nil, fmt.Errorf("invalid theme id %q", pkg.ID)
	}
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		return nil, err
	}
	dst := filepath.Join(themeDir, pkg.ID+".toml")
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, pkg.Theme, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return nil, err
	}
	return pkg, nil
}

// EntryFromLocal builds an entry from files under root (the repository
// checkout): the theme file or the extension's manifest.toml found at
// root/spec.Subdir. Used by registrygen -offline so the full metadata
// (name, permissions, version) comes from the sources in this repo.
func EntryFromLocal(s Spec, root string) (Entry, error) {
	e, _ := EntryFromSpec(s)
	if s.Subdir == "" {
		return e, ValidateEntry(e)
	}
	p := filepath.Join(root, filepath.FromSlash(s.Subdir))
	if s.Kind == KindTheme {
		data, err := os.ReadFile(p)
		if err != nil {
			return e, fmt.Errorf("%s: %w", s.ID, err)
		}
		t, err := ValidateTheme(data, s.ID)
		if err != nil {
			return e, fmt.Errorf("%s: %w", s.ID, err)
		}
		if s.Name == "" {
			e.Name = t.Name
		}
		if s.Author == "" && t.Author != "" {
			e.Author = t.Author
		}
		e.Permissions = []string{}
		return e, ValidateEntry(e)
	}
	data, err := os.ReadFile(filepath.Join(p, "manifest.toml"))
	if err != nil {
		return e, fmt.Errorf("%s: %w", s.ID, err)
	}
	mf, err := ParseManifest(data)
	if err != nil {
		return e, fmt.Errorf("%s: %w", s.ID, err)
	}
	if mf.ID != s.ID {
		return e, fmt.Errorf("%s: manifest id %q does not match the registry id", s.ID, mf.ID)
	}
	if mf.Name != "" {
		e.Name = mf.Name
	}
	if mf.Description != "" {
		e.Description = mf.Description
	}
	if mf.Author != "" {
		e.Author = mf.Author
	}
	if mf.Version != "" {
		e.Latest = mf.Version
	}
	e.Permissions = nonNil(mf.Permissions)
	return e, ValidateEntry(e)
}
