package market

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EntryFromSpec builds a registry entry without touching the network, using
// only what the spec file states. Used by registrygen -offline.
func EntryFromSpec(s Spec) (Entry, error) {
	e := Entry{
		ID: s.ID, Name: s.Name, Description: s.Description, Author: s.Author,
		Repo: s.Repo, Tags: nonNil(s.Tags), Latest: s.Latest, Permissions: nonNil(s.Permissions),
		Verified: s.Verified, Ref: s.Ref, Subdir: s.Subdir, Asset: s.Asset, SHA256: s.SHA256,
	}
	if s.Kind == KindTheme {
		e.Kind = KindTheme
	}
	if e.Name == "" {
		e.Name = s.ID
	}
	if e.Author == "" {
		e.Author, _, _ = strings.Cut(s.Repo, "/")
	}
	return e, ValidateEntry(e)
}

// BuildEntry queries GitHub for the spec's repository: the latest release (or
// spec.Ref) and the manifest.toml at that ref. Spec fields fill in anything
// the manifest leaves out.
func (c *Client) BuildEntry(ctx context.Context, s Spec) (Entry, error) {
	e, _ := EntryFromSpec(s)
	owner, repo, _ := strings.Cut(s.Repo, "/")
	ref := s.Ref
	if s.Kind == KindTheme {
		return c.buildThemeEntry(ctx, s, e, owner, repo)
	}
	var rel *Release
	if ref == "" {
		r, err := c.LatestRelease(ctx, owner, repo)
		if err == nil {
			rel, ref = r, r.Tag
		}
	}
	mf, err := c.FetchManifest(ctx, owner, repo, ref, s.Subdir)
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
	if mf.Permissions != nil {
		e.Permissions = mf.Permissions
	}
	switch {
	case mf.Version != "":
		e.Latest = mf.Version
	case rel != nil:
		e.Latest = rel.Version
	}
	if rel != nil {
		e.Downloads = rel.Downloads
	}
	e.Permissions = nonNil(e.Permissions)
	return e, ValidateEntry(e)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// LoadSpecs reads every *.toml in dir, sorted by id.
func LoadSpecs(dir string) ([]Spec, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.toml"))
	if err != nil {
		return nil, err
	}
	var specs []Spec
	seen := map[string]string{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		id := strings.TrimSuffix(filepath.Base(f), ".toml")
		s, err := ParseSpec(b, id)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		if prev, dup := seen[s.Repo+"|"+s.Subdir]; dup {
			return nil, fmt.Errorf("%s: repo %s already registered by %s", f, s.Repo, prev)
		}
		seen[s.Repo+"|"+s.Subdir] = f
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
	return specs, nil
}

// NewRegistry assembles a registry document.
func NewRegistry(entries []Entry, now time.Time) Registry {
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return Registry{Schema: SchemaVersion, Generated: now.UTC().Format(time.RFC3339), Extensions: entries}
}
