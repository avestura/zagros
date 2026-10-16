package ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Permission kinds.
const (
	PermNetwork        = "network"
	PermFSRead         = "fs.read"
	PermFSWrite        = "fs.write"
	PermClipboardRead  = "clipboard.read"
	PermClipboardWrite = "clipboard.write"
	PermProcessRun     = "process.run"
	PermEnvRead        = "env.read"
	PermNotifications  = "notifications"
	PermOpen           = "open"
	PermStorage        = "storage"
)

// Permission is "kind" or "kind:scope", e.g. "network:api.github.com".
type Permission string

// scoped lists the kinds that accept a ":scope" suffix.
var scoped = map[string]bool{PermNetwork: true, PermFSRead: true, PermFSWrite: true, PermProcessRun: true, PermEnvRead: true}

var allKinds = map[string]bool{
	PermNetwork: true, PermFSRead: true, PermFSWrite: true, PermClipboardRead: true, PermClipboardWrite: true,
	PermProcessRun: true, PermEnvRead: true, PermNotifications: true, PermOpen: true, PermStorage: true,
}

// Kind returns the part before the first colon.
func (p Permission) Kind() string {
	k, _, _ := strings.Cut(string(p), ":")
	return k
}

// Scope returns the part after the first colon ("" when unscoped).
func (p Permission) Scope() string {
	_, s, _ := strings.Cut(string(p), ":")
	return s
}

// Validate checks that the kind is known and a scope is only given where allowed.
func (p Permission) Validate() error {
	if !allKinds[p.Kind()] {
		return fmt.Errorf("unknown permission %q", string(p))
	}
	if strings.Contains(string(p), ":") {
		if !scoped[p.Kind()] {
			return fmt.Errorf("permission %q does not take a scope", p.Kind())
		}
		if p.Scope() == "" {
			return fmt.Errorf("permission %q has an empty scope", string(p))
		}
	}
	return nil
}

// Implicit reports whether the permission is always granted (storage).
func (p Permission) Implicit() bool { return p.Kind() == PermStorage }

// PermissionError is returned (as a Starlark error) when a host function is
// called without the needed permission.
type PermissionError struct {
	Ext    string
	Kind   string
	Target string
}

func (e *PermissionError) Error() string {
	t := ""
	if e.Target != "" {
		t = " for " + e.Target
	}
	return fmt.Sprintf("permission denied: %s requires %q%s, which the user has not granted", e.Ext, e.Kind, t)
}

// PermissionSet is a set of granted permissions with scope-aware checks.
type PermissionSet []Permission

// Allows reports whether an operation of the given kind on target is covered.
// target is a host, absolute path, executable or variable name; "" for unscoped kinds.
func (s PermissionSet) Allows(kind, target string) bool {
	if kind == PermStorage {
		return true
	}
	for _, p := range s {
		if p.Kind() != kind {
			continue
		}
		sc := p.Scope()
		if sc == "" || scopeMatches(kind, sc, target) {
			return true
		}
	}
	return false
}

func scopeMatches(kind, scope, target string) bool {
	switch kind {
	case PermNetwork:
		scope, target = strings.ToLower(scope), strings.ToLower(target)
		return target == scope || strings.HasSuffix(target, "."+scope)
	case PermFSRead, PermFSWrite:
		return pathWithin(expandHome(scope), target)
	case PermProcessRun:
		if strings.ContainsAny(scope, `/\`) {
			return samePath(filepath.Clean(scope), filepath.Clean(target))
		}
		return strings.EqualFold(exeBase(target), exeBase(scope))
	case PermEnvRead:
		return strings.EqualFold(scope, target)
	}
	return false
}

func exeBase(s string) string {
	b := filepath.Base(s)
	return strings.TrimSuffix(strings.ToLower(b), ".exe")
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, p[1:])
		}
	}
	return p
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// pathWithin reports whether target is root or inside root.
func pathWithin(root, target string) bool {
	root, target = filepath.Clean(root), filepath.Clean(target)
	if runtime.GOOS == "windows" {
		root, target = strings.ToLower(root), strings.ToLower(target)
	}
	if root == target {
		return true
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// declares reports whether granted permission g stays within the declared set.
func (m *Manifest) declares(g Permission) bool {
	if g.Implicit() {
		return true
	}
	for _, d := range m.Permissions {
		if d.Kind() == g.Kind() && (d.Scope() == "" || d.Scope() == g.Scope()) {
			return true
		}
	}
	return false
}

// Requestable returns the declared permissions that need user consent
// (everything except implicit ones), sorted.
func (m *Manifest) Requestable() []Permission {
	var out []Permission
	for _, p := range m.Permissions {
		if !p.Implicit() {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// PermissionPrompter asks the user which permissions to grant. The UI
// implements it. The returned list must be a subset of requested.
type PermissionPrompter interface {
	Request(ext *Manifest, requested []Permission) (granted []Permission, err error)
}

// AutoDeny grants nothing.
type AutoDeny struct{}

func (AutoDeny) Request(*Manifest, []Permission) ([]Permission, error) { return nil, nil }

// AutoAllow grants everything requested. Use in tests and trusted dev setups.
type AutoAllow struct{}

func (AutoAllow) Request(_ *Manifest, req []Permission) ([]Permission, error) {
	return append([]Permission(nil), req...), nil
}

// PrompterFunc adapts a function to PermissionPrompter.
type PrompterFunc func(ext *Manifest, requested []Permission) ([]Permission, error)

func (f PrompterFunc) Request(ext *Manifest, req []Permission) ([]Permission, error) {
	return f(ext, req)
}

// Grant is the persisted consent record of one extension.
type Grant struct {
	Declared []Permission `json:"declared"` // the set the user was asked about
	Granted  []Permission `json:"granted"`
	Decided  time.Time    `json:"decided"`
}

// PermissionStore persists grants in permissions.json, keyed by extension id
// (not version: an update with the same permission set keeps its grants).
type PermissionStore struct {
	path string
	mu   sync.Mutex
	data map[string]Grant
}

// OpenPermissionStore loads the store at path. An empty path keeps it in memory.
func OpenPermissionStore(path string) (*PermissionStore, error) {
	s := &PermissionStore{path: path, data: map[string]Grant{}}
	if path == "" {
		return s, nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s.data); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// Get returns the stored grant for id.
func (s *PermissionStore) Get(id string) (Grant, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.data[id]
	return g, ok
}

// NeedsConsent is true when the manifest declares a permission set different
// from the one the user last decided on.
func (s *PermissionStore) NeedsConsent(m *Manifest) bool {
	want := m.Requestable()
	g, ok := s.Get(m.ID)
	if len(want) == 0 {
		return false
	}
	if !ok {
		return true
	}
	return !sameSet(g.Declared, want)
}

func sameSet(a, b []Permission) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]Permission(nil), a...)
	y := append([]Permission(nil), b...)
	sort.Slice(x, func(i, j int) bool { return x[i] < x[j] })
	sort.Slice(y, func(i, j int) bool { return y[i] < y[j] })
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// Save records a decision and writes the file atomically.
func (s *PermissionStore) Save(id string, g Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = g
	return s.flushLocked()
}

// Forget removes id's record.
func (s *PermissionStore) Forget(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
	return s.flushLocked()
}

func (s *PermissionStore) flushLocked() error {
	if s.path == "" {
		return nil
	}
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(s.path, b)
}

func writeFileAtomic(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Consent makes sure the user has decided on m's permissions and returns the
// effective grants. With a nil prompter nothing new is granted or saved.
func (s *PermissionStore) Consent(m *Manifest, p PermissionPrompter) ([]Permission, error) {
	if !s.NeedsConsent(m) {
		g, _ := s.Get(m.ID)
		var out []Permission
		for _, x := range g.Granted {
			if m.declares(x) {
				out = append(out, x)
			}
		}
		return out, nil
	}
	if p == nil {
		return nil, nil
	}
	req := m.Requestable()
	granted, err := p.Request(m, req)
	if err != nil {
		return nil, err
	}
	var ok []Permission
	for _, g := range granted {
		if g.Validate() == nil && m.declares(g) {
			ok = append(ok, g)
		}
	}
	err = s.Save(m.ID, Grant{Declared: req, Granted: ok, Decided: time.Now().UTC()})
	return ok, err
}
