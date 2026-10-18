package ext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Limits bound what one extension call may consume.
type Limits struct {
	MaxSteps     uint64        // Starlark execution steps per call (default 20M)
	CallTimeout  time.Duration // wall-clock per call (default 5s)
	MaxHTTPBytes int64         // response body cap (default 8 MiB)
	MaxFileBytes int64         // fs.read_text cap (default 4 MiB)
	MaxKVBytes   int64         // total storage per extension (default 1 MiB)
	MaxBulkBytes int64         // download, copy, archive and decompress totals (default 256 MiB)
	MaxDBBytes   int64         // size of one sqlite database (default 64 MiB)
	MaxEntries   int           // files per walk, glob, copy or archive and rows per query (default 10000)
}

func (l Limits) withDefaults() Limits {
	if l.MaxSteps == 0 {
		l.MaxSteps = 20_000_000
	}
	if l.CallTimeout == 0 {
		l.CallTimeout = 5 * time.Second
	}
	if l.MaxHTTPBytes == 0 {
		l.MaxHTTPBytes = 8 << 20
	}
	if l.MaxFileBytes == 0 {
		l.MaxFileBytes = 4 << 20
	}
	if l.MaxKVBytes == 0 {
		l.MaxKVBytes = 1 << 20
	}
	if l.MaxBulkBytes == 0 {
		l.MaxBulkBytes = 256 << 20
	}
	if l.MaxDBBytes == 0 {
		l.MaxDBBytes = 64 << 20
	}
	if l.MaxEntries == 0 {
		l.MaxEntries = 10000
	}
	return l
}

// LogFunc receives extension log output. level is "info", "warn" or "error".
type LogFunc func(extID, level, msg string)

// Platform provides the OS-bound capabilities that need Win32 (clipboard,
// toast notifications). The app/UI supplies the real one.
type Platform interface {
	ClipboardGet() (string, error)
	ClipboardSet(text string) error
	Notify(title, body string) error
}

// NoPlatform is the default: every call reports "not available".
type NoPlatform struct{}

var errNoPlatform = errors.New("not available on this platform")

func (NoPlatform) ClipboardGet() (string, error) { return "", errNoPlatform }
func (NoPlatform) ClipboardSet(string) error     { return errNoPlatform }
func (NoPlatform) Notify(string, string) error   { return errNoPlatform }

// Options configures a Manager or a standalone Extension.
type Options struct {
	ExtDir     string // installed extensions live in <ExtDir>/<dirname>/
	DataDir    string // permissions.json and per-extension data; "" = in memory
	Prompter   PermissionPrompter
	Platform   Platform
	Executor   core.Executor // runs actions returned from on_execute; optional
	Limits     Limits
	Log        LogFunc
	AppVersion string // enforced against min_zagros when set
	HTTPClient *http.Client
}

func (o *Options) defaults() {
	o.Limits = o.Limits.withDefaults()
	if o.Platform == nil {
		o.Platform = NoPlatform{}
	}
	if o.Log == nil {
		o.Log = func(id, level, msg string) { log.Printf("ext %s [%s] %s", id, level, msg) }
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
}

var fileOpts = &syntax.FileOptions{Set: true, While: true, TopLevelControl: true, GlobalReassign: true, Recursion: true}

// Extension is one loaded extension. It is safe for concurrent use.
type Extension struct {
	Manifest *Manifest
	Dir      string // absolute directory containing manifest.toml

	opts *Options

	mu       sync.RWMutex
	rt       *vm // nil until Start
	enabled  bool
	loadErr  error
	granted  PermissionSet
	settings map[string]any // user values (unresolved)
	kv       *kvStore
	dataDir  string
	host     hostState // http cookie jar and open sqlite databases
}

type vm struct {
	globals starlark.StringDict
	pre     starlark.StringDict
	ctxVal  starlark.Value
}

const ctxKey = "zagros.ctx"

// OpenExtension reads the manifest in dir without running any code. Call
// SetGranted and Start afterwards. Manager does this for you.
func OpenExtension(dir string, o Options) (*Extension, error) {
	o.defaults()
	return openExtension(dir, &o)
}

func openExtension(dir string, o *Options) (*Extension, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(abs, ManifestFile))
	if err != nil {
		return nil, err
	}
	m, err := ParseManifest(b)
	if err != nil {
		return nil, err
	}
	if o.AppVersion != "" && m.MinZagros != "" && compareVersions(o.AppVersion, m.MinZagros) < 0 {
		return nil, fmt.Errorf("%s needs Zagros %s or newer (running %s)", m.ID, m.MinZagros, o.AppVersion)
	}
	e := &Extension{Manifest: m, Dir: abs, opts: o, enabled: true}
	if o.DataDir != "" {
		e.dataDir = filepath.Join(o.DataDir, "ext", m.DirName())
	}
	e.kv = newKV(e.dataDir, o.Limits.MaxKVBytes)
	e.settings = e.loadSettings()
	return e, nil
}

func (e *Extension) loadSettings() map[string]any {
	out := map[string]any{}
	if e.dataDir == "" {
		return out
	}
	if b, err := os.ReadFile(filepath.Join(e.dataDir, "settings.json")); err == nil {
		_ = json.Unmarshal(b, &out)
	}
	return out
}

// ID returns the extension id.
func (e *Extension) ID() string { return e.Manifest.ID }

// Enabled reports whether the extension is active.
func (e *Extension) Enabled() bool { e.mu.RLock(); defer e.mu.RUnlock(); return e.enabled }

// Err returns the last load/runtime-start error (nil when healthy).
func (e *Extension) Err() error { e.mu.RLock(); defer e.mu.RUnlock(); return e.loadErr }

// Running reports whether the Starlark module is loaded.
func (e *Extension) Running() bool { e.mu.RLock(); defer e.mu.RUnlock(); return e.rt != nil }

// SetGranted replaces the permissions the extension may use.
func (e *Extension) SetGranted(p []Permission) {
	e.mu.Lock()
	e.granted = append(PermissionSet(nil), p...)
	e.mu.Unlock()
}

// Granted returns the permissions currently granted.
func (e *Extension) Granted() []Permission {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]Permission(nil), e.granted...)
}

func (e *Extension) allows(kind, target string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.granted.Allows(kind, target)
}

func (e *Extension) require(kind, target string) error {
	if e.allows(kind, target) {
		return nil
	}
	return &PermissionError{Ext: e.Manifest.ID, Kind: kind, Target: target}
}

// Settings returns the effective settings (defaults merged with user values).
func (e *Extension) Settings() map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Manifest.ResolveSettings(e.settings)
}

// Keywords returns the manifest keywords.
func (e *Extension) Keywords() []string { return append([]string(nil), e.Manifest.Keywords...) }

// Start loads the Starlark module and runs on_load. On failure the extension
// stays stopped and Err reports why.
func (e *Extension) Start(ctx context.Context) error {
	rt, err := e.build(ctx)
	e.mu.Lock()
	e.loadErr = err
	if err == nil {
		e.rt = rt
	} else {
		e.rt = nil
	}
	e.mu.Unlock()
	if err != nil {
		return err
	}
	if _, err := e.callHook(ctx, "on_load"); err != nil && !errors.Is(err, errNoHook) {
		e.mu.Lock()
		e.loadErr = err
		e.mu.Unlock()
		return err
	}
	return nil
}

// Stop drops the loaded module. Queries return nothing until Start.
func (e *Extension) Stop() {
	e.mu.Lock()
	e.rt = nil
	e.mu.Unlock()
	e.closeDBs()
}

// SetEnabled flips the enabled flag (the Manager starts/stops accordingly).
func (e *Extension) SetEnabled(v bool) { e.mu.Lock(); e.enabled = v; e.mu.Unlock() }

// SetSettings validates and stores user setting values, rebuilds the module so
// the `settings` global reflects them, and calls on_settings_changed.
func (e *Extension) SetSettings(ctx context.Context, values map[string]any) error {
	clean := map[string]any{}
	for i := range e.Manifest.Settings {
		s := &e.Manifest.Settings[i]
		v, ok := values[s.Key]
		if !ok {
			continue
		}
		c, err := coerceSetting(s, v)
		if err != nil {
			return fmt.Errorf("setting %q: %w", s.Key, err)
		}
		clean[s.Key] = c
	}
	for k := range values {
		found := false
		for _, s := range e.Manifest.Settings {
			if s.Key == k {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("unknown setting %q", k)
		}
	}
	if e.dataDir != "" {
		b, _ := json.MarshalIndent(clean, "", "  ")
		if err := writeFileAtomic(filepath.Join(e.dataDir, "settings.json"), b); err != nil {
			return err
		}
	}
	e.mu.Lock()
	e.settings = clean
	running := e.rt != nil
	e.mu.Unlock()
	if !running {
		return nil
	}
	rt, err := e.build(ctx)
	if err != nil {
		e.mu.Lock()
		e.loadErr = err
		e.mu.Unlock()
		return err
	}
	e.mu.Lock()
	e.rt = rt
	e.mu.Unlock()
	if _, err := e.callHook(ctx, "on_settings_changed"); err != nil && !errors.Is(err, errNoHook) {
		return err
	}
	return nil
}

// ---- execution ----

var errNoHook = errors.New("hook not defined")

func (e *Extension) runtime() *vm {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.rt
}

func (e *Extension) newThread(ctx context.Context, name string, load func(*starlark.Thread, string) (starlark.StringDict, error)) *starlark.Thread {
	th := &starlark.Thread{
		Name: e.Manifest.ID + ":" + name,
		Print: func(_ *starlark.Thread, msg string) {
			e.opts.Log(e.Manifest.ID, "info", msg)
		},
		Load: load,
	}
	th.SetMaxExecutionSteps(e.opts.Limits.MaxSteps)
	th.SetLocal(ctxKey, ctx)
	return th
}

// withLimits runs f, cancelling the thread when ctx ends or the timeout hits.
func (e *Extension) withLimits(ctx context.Context, th *starlark.Thread, f func() error) error {
	cctx, cancel := context.WithTimeout(ctx, e.opts.Limits.CallTimeout)
	defer cancel()
	th.SetLocal(ctxKey, cctx)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-cctx.Done():
			th.Cancel(cctx.Err().Error())
		case <-done:
		}
	}()
	return f()
}

func callCtx(th *starlark.Thread) context.Context {
	if c, ok := th.Local(ctxKey).(context.Context); ok {
		return c
	}
	return context.Background()
}

func formatErr(err error) error {
	var ee *starlark.EvalError
	if errors.As(err, &ee) {
		return fmt.Errorf("%s\n%s", ee.Msg, ee.Backtrace())
	}
	return err
}

// build executes the entry script (and its load()s) into a frozen module.
func (e *Extension) build(ctx context.Context) (*vm, error) {
	pre := e.predeclared()
	ld := &loader{ext: e, pre: pre, cache: map[string]*loadEntry{}}
	entry := filepath.ToSlash(e.Manifest.Entry)
	src, err := ld.read(entry)
	if err != nil {
		return nil, err
	}
	th := e.newThread(ctx, "load", ld.load)
	var globals starlark.StringDict
	err = e.withLimits(ctx, th, func() error {
		var err error
		globals, err = starlark.ExecFileOptions(fileOpts, th, entry, src, pre)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", entry, formatErr(err))
	}
	globals.Freeze()
	return &vm{globals: globals, pre: pre, ctxVal: e.ctxValue()}, nil
}

type loadEntry struct {
	globals starlark.StringDict
	err     error
	loading bool
}

type loader struct {
	ext   *Extension
	pre   starlark.StringDict
	mu    sync.Mutex
	cache map[string]*loadEntry
}

func (l *loader) read(rel string) ([]byte, error) {
	rel = strings.ReplaceAll(rel, "\\", "/")
	if strings.HasPrefix(rel, "/") || strings.Contains(rel, "..") || strings.Contains(rel, ":") || !strings.HasSuffix(rel, ".star") {
		return nil, fmt.Errorf("invalid module path %q", rel)
	}
	return os.ReadFile(filepath.Join(l.ext.Dir, filepath.FromSlash(rel)))
}

func (l *loader) load(th *starlark.Thread, module string) (starlark.StringDict, error) {
	module = strings.TrimPrefix(strings.ReplaceAll(module, "\\", "/"), "./")
	l.mu.Lock()
	ent, ok := l.cache[module]
	if ok && ent.loading {
		l.mu.Unlock()
		return nil, fmt.Errorf("cycle in load of %s", module)
	}
	if ok {
		l.mu.Unlock()
		return ent.globals, ent.err
	}
	ent = &loadEntry{loading: true}
	l.cache[module] = ent
	l.mu.Unlock()

	src, err := l.read(module)
	if err == nil {
		sub := &starlark.Thread{Name: module, Load: l.load, Print: th.Print}
		sub.SetMaxExecutionSteps(l.ext.opts.Limits.MaxSteps)
		sub.SetLocal(ctxKey, callCtx(th))
		ent.globals, err = starlark.ExecFileOptions(fileOpts, sub, module, src, l.pre)
		if err == nil {
			ent.globals.Freeze()
		}
	}
	l.mu.Lock()
	ent.err, ent.loading = err, false
	l.mu.Unlock()
	return ent.globals, err
}

// callHook calls a global function with ctx as first argument plus args.
func (e *Extension) callHook(ctx context.Context, name string, args ...starlark.Value) (starlark.Value, error) {
	rt := e.runtime()
	if rt == nil {
		return nil, errors.New("extension is not running")
	}
	fn, ok := rt.globals[name].(starlark.Callable)
	if !ok {
		return nil, errNoHook
	}
	th := e.newThread(ctx, name, nil)
	full := append(starlark.Tuple{rt.ctxVal}, args...)
	var res starlark.Value
	err := e.withLimits(ctx, th, func() error {
		var err error
		res, err = starlark.Call(th, fn, full, nil)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, formatErr(err))
	}
	return res, nil
}

// Call invokes an arbitrary global function (used by tests and tooling).
func (e *Extension) Call(ctx context.Context, name string, args ...starlark.Value) (starlark.Value, error) {
	return e.callHook(ctx, name, args...)
}

// ---- key/value storage ----

type kvStore struct {
	path string
	max  int64
	mu   sync.Mutex
	data map[string]json.RawMessage
	size int64
}

func newKV(dir string, max int64) *kvStore {
	k := &kvStore{max: max, data: map[string]json.RawMessage{}}
	if dir != "" {
		k.path = filepath.Join(dir, "storage.json")
		if b, err := os.ReadFile(k.path); err == nil {
			_ = json.Unmarshal(b, &k.data)
		}
		for key, v := range k.data {
			k.size += int64(len(key) + len(v))
		}
	}
	return k
}

func (k *kvStore) get(key string) (json.RawMessage, bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	v, ok := k.data[key]
	return v, ok
}

func (k *kvStore) keys() []string {
	k.mu.Lock()
	defer k.mu.Unlock()
	out := make([]string, 0, len(k.data))
	for key := range k.data {
		out = append(out, key)
	}
	return out
}

func (k *kvStore) set(key string, v json.RawMessage) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	old, had := k.data[key]
	size := k.size + int64(len(key)+len(v))
	if had {
		size -= int64(len(key) + len(old))
	}
	if size > k.max {
		return fmt.Errorf("storage quota exceeded (%d bytes)", k.max)
	}
	k.data[key] = v
	k.size = size
	return k.flushLocked()
}

func (k *kvStore) delete(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if old, ok := k.data[key]; ok {
		k.size -= int64(len(key) + len(old))
		delete(k.data, key)
		return k.flushLocked()
	}
	return nil
}

func (k *kvStore) flushLocked() error {
	if k.path == "" {
		return nil
	}
	b, err := json.Marshal(k.data)
	if err != nil {
		return err
	}
	return writeFileAtomic(k.path, b)
}

func (e *Extension) hasHook(name string) bool {
	rt := e.runtime()
	if rt == nil {
		return false
	}
	_, ok := rt.globals[name].(starlark.Callable)
	return ok
}
