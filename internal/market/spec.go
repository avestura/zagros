package market

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	idRe     = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	repoRe   = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
	sha256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// ValidID reports whether s is a legal extension id.
func ValidID(s string) bool { return idRe.MatchString(s) }

// Spec is one registry/extensions/<id>.toml contribution file. Name,
// Description, Author, Permissions and Latest are optional overrides that are
// normally taken from the extension's manifest; they matter for -offline runs.
type Spec struct {
	ID          string
	Repo        string
	Ref         string
	Subdir      string
	Asset       string
	Tags        []string
	SHA256      string
	Verified    bool
	Kind        string // "extension" (default) or "theme"
	Name        string
	Description string
	Author      string
	Latest      string
	Permissions []string
}

// ParseSpec parses a registry TOML file. defaultID is normally the file name
// without extension; it is used when the file has no explicit id and must
// match it when it does.
func ParseSpec(data []byte, defaultID string) (Spec, error) {
	m, err := parseTOML(string(data))
	if err != nil {
		return Spec{}, err
	}
	s := Spec{
		ID:          tomlString(m, "id"),
		Repo:        tomlString(m, "repo"),
		Ref:         tomlString(m, "ref"),
		Subdir:      strings.Trim(tomlString(m, "subdir"), "/"),
		Asset:       tomlString(m, "asset"),
		Tags:        tomlStrings(m, "tags"),
		SHA256:      strings.ToLower(tomlString(m, "sha256")),
		Kind:        strings.ToLower(tomlString(m, "kind")),
		Name:        tomlString(m, "name"),
		Description: tomlString(m, "description"),
		Author:      tomlString(m, "author"),
		Latest:      strings.TrimPrefix(tomlString(m, "latest"), "v"),
		Permissions: tomlStrings(m, "permissions"),
	}
	s.Verified, _ = m["verified"].(bool)
	if s.ID == "" {
		s.ID = defaultID
	}
	if defaultID != "" && s.ID != defaultID {
		return s, fmt.Errorf("id %q must match the file name %q", s.ID, defaultID)
	}
	if !idRe.MatchString(s.ID) {
		return s, fmt.Errorf("invalid id %q (lowercase letters, digits, . _ -)", s.ID)
	}
	if !repoRe.MatchString(s.Repo) {
		return s, fmt.Errorf("repo %q must look like owner/name", s.Repo)
	}
	switch s.Kind {
	case "":
		s.Kind = KindExtension
	case KindExtension, KindTheme:
	default:
		return s, fmt.Errorf("kind %q must be \"extension\" or \"theme\"", s.Kind)
	}
	if s.Kind == KindTheme && !strings.HasSuffix(strings.ToLower(s.Subdir), ".toml") && s.Asset == "" {
		return s, fmt.Errorf("theme entries need subdir = path to the .toml file, or an asset glob")
	}
	if s.SHA256 != "" && !sha256Re.MatchString(s.SHA256) {
		return s, fmt.Errorf("sha256 must be 64 hex characters")
	}
	if strings.Contains(s.Subdir, "..") {
		return s, fmt.Errorf("subdir must not contain ..")
	}
	return s, nil
}

// ValidateEntry checks a finished registry entry.
func ValidateEntry(e Entry) error {
	switch {
	case !idRe.MatchString(e.ID):
		return fmt.Errorf("invalid id %q", e.ID)
	case e.Name == "":
		return fmt.Errorf("%s: name is empty", e.ID)
	case e.Kind != "" && e.Kind != KindExtension && e.Kind != KindTheme:
		return fmt.Errorf("%s: invalid kind %q", e.ID, e.Kind)
	case !repoRe.MatchString(e.Repo):
		return fmt.Errorf("%s: invalid repo %q", e.ID, e.Repo)
	case e.Latest != "" && !parseSemver(e.Latest).ok:
		return fmt.Errorf("%s: latest %q is not semver", e.ID, e.Latest)
	}
	return nil
}
