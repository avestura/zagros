package ext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
)

// InstalledRecord is one entry of <ExtDir>/installed.json.
type InstalledRecord struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"` // zip path, "github:owner/repo", or "local"
	Version     string    `json:"version"`
	SHA256      string    `json:"sha256"` // of the installed archive
	InstalledAt time.Time `json:"installed_at"`
	Disabled    bool      `json:"disabled,omitempty"`
}

// Manager owns the installed extensions.
type Manager struct {
	opts  Options
	store *PermissionStore

	mu       sync.RWMutex
	exts     map[string]*Extension // by id
	broken   map[string]error      // dirname -> manifest/load error
	records  map[string]InstalledRecord
	onChange func()
	sigs     map[string]string // dirname -> content signature (dev watcher)
}

// NewManager prepares a manager. Call Discover to load extensions.
func NewManager(o Options) (*Manager, error) {
	o.defaults()
	if o.ExtDir == "" {
		return nil, errors.New("ext: ExtDir is required")
	}
	if err := os.MkdirAll(o.ExtDir, 0o755); err != nil {
		return nil, err
	}
	permPath := ""
	if o.DataDir != "" {
		permPath = filepath.Join(o.DataDir, "permissions.json")
	}
	st, err := OpenPermissionStore(permPath)
	if err != nil {
		return nil, err
	}
	m := &Manager{opts: o, store: st, exts: map[string]*Extension{}, broken: map[string]error{}, sigs: map[string]string{}}
	m.records = m.readRecords()
	return m, nil
}

// OnChange registers a callback fired after installs, removals, reloads and
// enable/disable (the UI refreshes its lists / invalidates caches).
func (m *Manager) OnChange(f func()) { m.mu.Lock(); m.onChange = f; m.mu.Unlock() }

func (m *Manager) changed() {
	m.mu.RLock()
	f := m.onChange
	m.mu.RUnlock()
	if f != nil {
		f()
	}
}

// Permissions exposes the grant store (for the settings UI).
func (m *Manager) Permissions() *PermissionStore { return m.store }

// ---- installed.json ----

func (m *Manager) recordsPath() string { return filepath.Join(m.opts.ExtDir, "installed.json") }

func (m *Manager) readRecords() map[string]InstalledRecord {
	out := map[string]InstalledRecord{}
	if b, err := os.ReadFile(m.recordsPath()); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	return out
}

// Records returns a copy of installed.json.
func (m *Manager) Records() map[string]InstalledRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]InstalledRecord, len(m.records))
	for k, v := range m.records {
		out[k] = v
	}
	return out
}

func (m *Manager) saveRecord(r InstalledRecord) error {
	m.mu.Lock()
	m.records[r.ID] = r
	b, _ := json.MarshalIndent(m.records, "", "  ")
	m.mu.Unlock()
	return writeFileAtomic(m.recordsPath(), b)
}

func (m *Manager) dropRecord(id string) error {
	m.mu.Lock()
	delete(m.records, id)
	b, _ := json.MarshalIndent(m.records, "", "  ")
	m.mu.Unlock()
	return writeFileAtomic(m.recordsPath(), b)
}

// ---- discovery ----

// Discover scans ExtDir for unpacked extensions (<ExtDir>/<dirname>/manifest.toml),
// asks for consent where the permission set is new or changed, and starts every
// enabled one. Broken extensions are kept in the list with Err set / BrokenDirs.
func (m *Manager) Discover(ctx context.Context) ([]*Extension, error) {
	ents, err := os.ReadDir(m.opts.ExtDir)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	for _, e := range m.exts {
		e.Stop()
	}
	m.exts = map[string]*Extension{}
	m.broken = map[string]error{}
	m.mu.Unlock()
	for _, d := range ents {
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			continue
		}
		dir := filepath.Join(m.opts.ExtDir, d.Name())
		if _, err := os.Stat(filepath.Join(dir, ManifestFile)); err != nil {
			continue
		}
		m.loadDir(ctx, dir)
	}
	m.changed()
	return m.List(), nil
}

// loadDir opens, consents and starts one extension directory.
func (m *Manager) loadDir(ctx context.Context, dir string) *Extension {
	name := filepath.Base(dir)
	e, err := openExtension(dir, &m.opts)
	if err != nil {
		m.mu.Lock()
		m.broken[name] = err
		m.mu.Unlock()
		m.opts.Log(name, "error", err.Error())
		return nil
	}
	m.mu.Lock()
	rec, hasRec := m.records[e.Manifest.ID]
	m.exts[e.Manifest.ID] = e
	delete(m.broken, name)
	m.sigs[name] = dirSignature(dir)
	m.mu.Unlock()
	if hasRec && rec.Disabled {
		e.SetEnabled(false)
		return e
	}
	m.startExt(ctx, e)
	return e
}

func (m *Manager) startExt(ctx context.Context, e *Extension) {
	granted, err := m.store.Consent(e.Manifest, m.opts.Prompter)
	if err != nil {
		m.opts.Log(e.Manifest.ID, "error", "permission prompt: "+err.Error())
	}
	e.SetGranted(granted)
	if err := e.Start(ctx); err != nil {
		m.opts.Log(e.Manifest.ID, "error", err.Error())
	}
}

// BrokenDirs lists extension folders whose manifest failed to load.
func (m *Manager) BrokenDirs() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[string]error{}
	for k, v := range m.broken {
		out[k] = v
	}
	return out
}

// List returns all known extensions sorted by name.
func (m *Manager) List() []*Extension {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Extension, 0, len(m.exts))
	for _, e := range m.exts {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := strings.ToLower(out[i].Manifest.Name), strings.ToLower(out[j].Manifest.Name)
		if a != b {
			return a < b
		}
		return out[i].Manifest.ID < out[j].Manifest.ID
	})
	return out
}

// Get returns the extension with the given id.
func (m *Manager) Get(id string) (*Extension, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.exts[id]
	return e, ok
}

// Providers returns a core.Provider for every running, enabled extension.
func (m *Manager) Providers() []core.Provider {
	var out []core.Provider
	for _, e := range m.List() {
		if e.Enabled() && e.Running() {
			out = append(out, NewProvider(e))
		}
	}
	return out
}

// ---- install / uninstall ----

// InstallPath installs an extension from a project folder (packed on the fly,
// recorded as source "local") or from a .zagros archive.
func (m *Manager) InstallPath(ctx context.Context, path string) (*Extension, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return m.InstallFile(ctx, path, "")
	}
	data, err := Pack(path)
	if err != nil {
		return nil, err
	}
	e, err := m.InstallBytes(ctx, data, "local")
	if err == nil && e == nil {
		err = errors.New("installation failed")
	}
	return e, err
}

// InstallFile installs a .zagros archive from disk. source is recorded in
// installed.json (the path when empty).
func (m *Manager) InstallFile(ctx context.Context, zipPath, source string) (*Extension, error) {
	if st, err := os.Stat(zipPath); err != nil {
		return nil, err
	} else if st.Size() > MaxArchiveBytes {
		return nil, fmt.Errorf("archive is %d bytes, limit is %d", st.Size(), MaxArchiveBytes)
	}
	b, err := os.ReadFile(zipPath)
	if err != nil {
		return nil, err
	}
	if source == "" {
		source, _ = filepath.Abs(zipPath)
	}
	return m.InstallBytes(ctx, b, source)
}

// InstallBytes installs an archive held in memory. An existing extension with
// the same id is replaced (its storage, settings and grants are kept).
func (m *Manager) InstallBytes(ctx context.Context, data []byte, source string) (*Extension, error) {
	tmp, err := os.MkdirTemp(m.opts.ExtDir, ".install-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	mb, err := extractZip(data, tmp)
	if err != nil {
		return nil, err
	}
	man, err := ParseManifest(mb)
	if err != nil {
		return nil, err
	}
	if tooOld(m.opts.AppVersion, man.MinZagros) {
		return nil, fmt.Errorf("%s needs Zagros %s or newer (running %s)", man.ID, man.MinZagros, m.opts.AppVersion)
	}
	if _, err := os.Stat(filepath.Join(tmp, filepath.FromSlash(man.Entry))); err != nil {
		return nil, fmt.Errorf("entry %s missing from archive", man.Entry)
	}
	final := filepath.Join(m.opts.ExtDir, man.DirName())
	if old, ok := m.Get(man.ID); ok {
		old.Stop()
	}
	backup := ""
	if _, err := os.Stat(final); err == nil {
		backup = final + ".old-" + fmt.Sprint(time.Now().UnixNano())
		if err := os.Rename(final, backup); err != nil {
			return nil, fmt.Errorf("replacing existing install: %w", err)
		}
	}
	if err := os.Rename(tmp, final); err != nil {
		if backup != "" {
			_ = os.Rename(backup, final)
		}
		return nil, err
	}
	if backup != "" {
		os.RemoveAll(backup)
	}
	sum := sha256.Sum256(data)
	if source == "" {
		source = "bytes"
	}
	if err := m.saveRecord(InstalledRecord{ID: man.ID, Source: source, Version: man.Version, SHA256: hex.EncodeToString(sum[:]), InstalledAt: time.Now().UTC()}); err != nil {
		return nil, err
	}
	e := m.loadDir(ctx, final)
	m.changed()
	if e == nil {
		return nil, m.BrokenDirs()[man.DirName()]
	}
	return e, nil
}

// Uninstall stops and deletes the extension. purge also removes its storage,
// settings and permission grants.
func (m *Manager) Uninstall(id string, purge bool) error {
	if dir, ok := m.BrokenDir(id); ok {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		m.mu.Lock()
		delete(m.broken, filepath.Base(dir))
		delete(m.sigs, filepath.Base(dir))
		m.mu.Unlock()
		m.changed()
		return nil
	}
	e, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("extension %q not installed", id)
	}
	e.Stop()
	if err := os.RemoveAll(e.Dir); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.exts, id)
	delete(m.sigs, filepath.Base(e.Dir))
	m.mu.Unlock()
	if err := m.dropRecord(id); err != nil {
		return err
	}
	if purge {
		if e.dataDir != "" {
			os.RemoveAll(e.dataDir)
		}
		_ = m.store.Forget(id)
	}
	m.changed()
	return nil
}

// SetEnabled enables or disables an extension and persists the choice.
func (m *Manager) SetEnabled(ctx context.Context, id string, enabled bool) error {
	e, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("extension %q not installed", id)
	}
	m.mu.RLock()
	rec, has := m.records[id]
	m.mu.RUnlock()
	if !has {
		rec = InstalledRecord{ID: id, Source: "local", Version: e.Manifest.Version, InstalledAt: time.Now().UTC()}
	}
	rec.Disabled = !enabled
	if err := m.saveRecord(rec); err != nil {
		return err
	}
	e.SetEnabled(enabled)
	if enabled {
		m.startExt(ctx, e)
	} else {
		e.Stop()
	}
	m.changed()
	return e.Err()
}

// Enable and Disable are shorthands for SetEnabled.
func (m *Manager) Enable(ctx context.Context, id string) error { return m.SetEnabled(ctx, id, true) }
func (m *Manager) Disable(id string) error {
	return m.SetEnabled(context.Background(), id, false)
}

// Reload re-reads the extension from disk (manifest and scripts) and restarts it.
func (m *Manager) Reload(ctx context.Context, id string) error {
	e, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("extension %q not installed", id)
	}
	dir := e.Dir
	e.Stop()
	m.mu.Lock()
	delete(m.exts, id)
	m.mu.Unlock()
	ne := m.loadDir(ctx, dir)
	m.changed()
	if ne == nil {
		return m.BrokenDirs()[filepath.Base(dir)]
	}
	return ne.Err()
}

// GrantPermissions re-prompts the user for id's permissions regardless of the
// stored decision (settings UI "review permissions").
func (m *Manager) GrantPermissions(id string) error {
	e, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("extension %q not installed", id)
	}
	if m.opts.Prompter == nil {
		return errors.New("no permission prompter configured")
	}
	req := e.Manifest.Requestable()
	granted, err := m.opts.Prompter.Request(e.Manifest, req)
	if err != nil {
		return err
	}
	var ok2 []Permission
	for _, g := range granted {
		if g.Validate() == nil && e.Manifest.declares(g) {
			ok2 = append(ok2, g)
		}
	}
	if err := m.store.Save(id, Grant{Declared: req, Granted: ok2, Decided: time.Now().UTC()}); err != nil {
		return err
	}
	e.SetGranted(ok2)
	m.changed()
	return nil
}

// ---- execution ----

// Execute runs an "ext-call" action by dispatching to the owning extension.
func (m *Manager) Execute(a core.Action) error {
	if a.Kind != KindExtCall {
		return fmt.Errorf("ext: not an %s action: %q", KindExtCall, a.Kind)
	}
	e, ok := m.Get(a.Extra["ext"])
	if !ok {
		return fmt.Errorf("extension %q not found", a.Extra["ext"])
	}
	return e.Execute(context.Background(), a)
}

type chainExecutor struct {
	m    *Manager
	next core.Executor
}

func (c chainExecutor) Execute(a core.Action) error {
	if a.Kind == KindExtCall {
		return c.m.Execute(a)
	}
	return c.next.Execute(a)
}

// WrapExecutor returns an Executor that handles "ext-call" and forwards all
// other kinds to next. It is also usable as Options.Executor's counterpart in
// the app wiring.
func (m *Manager) WrapExecutor(next core.Executor) core.Executor { return chainExecutor{m, next} }

// ---- dev mode hot reload ----

// dirSignature changes whenever a file in dir is added, removed or modified.
func dirSignature(dir string) string {
	var sb strings.Builder
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			rel, _ := filepath.Rel(dir, p)
			fmt.Fprintf(&sb, "%s|%d|%d;", rel, info.Size(), info.ModTime().UnixNano())
		}
		return nil
	})
	return sb.String()
}

// WatchDev polls ExtDir every interval (default 500ms) and reloads extensions
// whose files changed, loads new folders and drops removed ones, until ctx
// ends. Intended for `zagros --dev`.
func (m *Manager) WatchDev(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.pollOnce(ctx)
		}
	}
}

func (m *Manager) pollOnce(ctx context.Context) {
	ents, err := os.ReadDir(m.opts.ExtDir)
	if err != nil {
		return
	}
	seen := map[string]bool{}
	dirty := false
	for _, d := range ents {
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			continue
		}
		dir := filepath.Join(m.opts.ExtDir, d.Name())
		if _, err := os.Stat(filepath.Join(dir, ManifestFile)); err != nil {
			continue
		}
		seen[d.Name()] = true
		sig := dirSignature(dir)
		m.mu.RLock()
		old, known := m.sigs[d.Name()]
		m.mu.RUnlock()
		if known && old == sig {
			continue
		}
		// Replace any extension that lives in this folder.
		m.mu.Lock()
		for id, e := range m.exts {
			if filepath.Base(e.Dir) == d.Name() {
				e.Stop()
				delete(m.exts, id)
			}
		}
		m.mu.Unlock()
		m.opts.Log(d.Name(), "info", "change detected, reloading")
		m.loadDir(ctx, dir)
		dirty = true
	}
	m.mu.Lock()
	for id, e := range m.exts {
		if !seen[filepath.Base(e.Dir)] {
			e.Stop()
			delete(m.exts, id)
			delete(m.sigs, filepath.Base(e.Dir))
			dirty = true
		}
	}
	m.mu.Unlock()
	if dirty {
		m.changed()
	}
}
