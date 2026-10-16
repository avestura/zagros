// Package ext implements the Zagros extension system: .zagros archives,
// manifests, permissions, the sandboxed Starlark runtime and the extension
// manager.
package ext

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// ManifestFile is the manifest's file name inside an extension.
const ManifestFile = "manifest.toml"

// DefaultEntry is the entry script used when the manifest has no "entry".
const DefaultEntry = "main.star"

// Manifest is the parsed manifest.toml of an extension.
type Manifest struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Author      string `toml:"author"`
	Homepage    string `toml:"homepage"`
	License     string `toml:"license"`
	MinZagros   string `toml:"min_zagros"`
	Entry       string `toml:"entry"`

	// Keywords are query keywords: with keywords set, the extension is only
	// asked when the query starts with one of them.
	Keywords    []string     `toml:"keywords"`
	Permissions []Permission `toml:"permissions"`
	// PermissionDetails explains each permission to the user in the consent
	// dialog. (A separate key, because TOML cannot make "permissions" both an
	// array and a table.)
	PermissionDetails []PermissionDetail `toml:"permission_detail"`
	Settings          []Setting          `toml:"settings"`
}

// PermissionDetail is the reason shown for a requested permission.
type PermissionDetail struct {
	Permission Permission `toml:"permission"`
	Reason     string     `toml:"reason"`
}

// Setting declares one user-configurable setting.
type Setting struct {
	Key         string   `toml:"key"`
	Type        string   `toml:"type"` // string, bool, int, number, choice
	Default     any      `toml:"default"`
	Label       string   `toml:"label"`
	Description string   `toml:"description"`
	Options     []string `toml:"options"` // for type "choice"
}

var (
	idRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)?$`)
	semverRe  = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	settingRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

// ParseManifest decodes and validates manifest.toml content.
func ParseManifest(data []byte) (*Manifest, error) {
	m, _, err := parseManifest(data)
	return m, err
}

// parseManifest also returns warnings (unknown keys).
func parseManifest(data []byte) (*Manifest, []string, error) {
	var m Manifest
	md, err := toml.Decode(string(data), &m)
	if err != nil {
		return nil, nil, fmt.Errorf("manifest.toml: %w", err)
	}
	var warns []string
	for _, k := range md.Undecoded() {
		warns = append(warns, "unknown manifest key: "+k.String())
	}
	if m.Entry == "" {
		m.Entry = DefaultEntry
	}
	if err := m.Validate(); err != nil {
		return &m, warns, err
	}
	return &m, warns, nil
}

// DirName is the folder name used for the extension under ExtDir.
func (m *Manifest) DirName() string { return dirName(m.ID) }

func dirName(id string) string { return strings.ReplaceAll(id, "/", "__") }

// Validate checks required fields and formats. All problems are reported.
func (m *Manifest) Validate() error {
	var errs []string
	add := func(f string, a ...any) { errs = append(errs, fmt.Sprintf(f, a...)) }
	if !idRe.MatchString(m.ID) || len(m.ID) > 100 {
		add("id %q must be reverse-dns or owner/name (lowercase letters, digits, . _ -)", m.ID)
	}
	if strings.TrimSpace(m.Name) == "" {
		add("name is required")
	}
	if !semverRe.MatchString(m.Version) {
		add("version %q must be semver (e.g. 1.0.0)", m.Version)
	}
	if m.MinZagros != "" && !semverRe.MatchString(m.MinZagros) {
		add("min_zagros %q must be semver", m.MinZagros)
	}
	e := strings.ReplaceAll(m.Entry, "\\", "/")
	if e == "" || strings.HasPrefix(e, "/") || strings.Contains(e, "..") || !strings.HasSuffix(e, ".star") {
		add("entry %q must be a relative .star path inside the extension", m.Entry)
	}
	for _, k := range m.Keywords {
		if k == "" || strings.ContainsAny(k, " \t") {
			add("keyword %q must be a single non-empty word", k)
		}
	}
	seen := map[Permission]bool{}
	for _, p := range m.Permissions {
		if err := p.Validate(); err != nil {
			add("%v", err)
		}
		if seen[p] {
			add("duplicate permission %q", string(p))
		}
		seen[p] = true
	}
	for _, d := range m.PermissionDetails {
		if !seen[d.Permission] {
			add("permission_detail for %q which is not in permissions", string(d.Permission))
		}
	}
	keys := map[string]bool{}
	for i := range m.Settings {
		s := &m.Settings[i]
		if !settingRe.MatchString(s.Key) {
			add("setting key %q must match [a-z_][a-z0-9_]*", s.Key)
		}
		if keys[s.Key] {
			add("duplicate setting %q", s.Key)
		}
		keys[s.Key] = true
		switch s.Type {
		case "string", "bool", "int", "number":
		case "choice":
			if len(s.Options) == 0 {
				add("setting %q: choice needs options", s.Key)
			}
		default:
			add("setting %q: unknown type %q", s.Key, s.Type)
			continue
		}
		if s.Default != nil {
			v, err := coerceSetting(s, s.Default)
			if err != nil {
				add("setting %q default: %v", s.Key, err)
			} else {
				s.Default = v
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid manifest: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Reason returns the declared reason for p, or "".
func (m *Manifest) Reason(p Permission) string {
	for _, d := range m.PermissionDetails {
		if d.Permission == p {
			return d.Reason
		}
	}
	return ""
}

// coerceSetting converts v (from TOML or JSON) to the Go type for s.
func coerceSetting(s *Setting, v any) (any, error) {
	switch s.Type {
	case "string":
		if x, ok := v.(string); ok {
			return x, nil
		}
	case "bool":
		if x, ok := v.(bool); ok {
			return x, nil
		}
	case "int":
		switch x := v.(type) {
		case int64:
			return x, nil
		case int:
			return int64(x), nil
		case float64:
			if x == float64(int64(x)) {
				return int64(x), nil
			}
		}
	case "number":
		switch x := v.(type) {
		case float64:
			return x, nil
		case int64:
			return float64(x), nil
		case int:
			return float64(x), nil
		}
	case "choice":
		if x, ok := v.(string); ok {
			for _, o := range s.Options {
				if o == x {
					return x, nil
				}
			}
			return nil, fmt.Errorf("%q is not one of %v", x, s.Options)
		}
	}
	return nil, fmt.Errorf("expected %s, got %T", s.Type, v)
}

func zeroSetting(s *Setting) any {
	switch s.Type {
	case "bool":
		return false
	case "int":
		return int64(0)
	case "number":
		return float64(0)
	case "choice":
		if len(s.Options) > 0 {
			return s.Options[0]
		}
	}
	return ""
}

// ResolveSettings merges user values over the manifest defaults. Invalid user
// values fall back to the default.
func (m *Manifest) ResolveSettings(user map[string]any) map[string]any {
	out := map[string]any{}
	for i := range m.Settings {
		s := &m.Settings[i]
		v := s.Default
		if v == nil {
			v = zeroSetting(s)
		}
		if u, ok := user[s.Key]; ok {
			if c, err := coerceSetting(s, u); err == nil {
				v = c
			}
		}
		out[s.Key] = v
	}
	return out
}

// compareVersions compares the numeric major.minor.patch parts.
func compareVersions(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(v string) [3]int {
	var out [3]int
	m := semverRe.FindStringSubmatch(v)
	if m == nil {
		return out
	}
	for i := 0; i < 3; i++ {
		out[i], _ = strconv.Atoi(m[i+1])
	}
	return out
}
