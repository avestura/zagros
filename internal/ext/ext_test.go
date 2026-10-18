package ext

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/core"
	"go.starlark.net/starlark"
)

const baseManifest = `id = "test.ext"
name = "Test Ext"
version = "1.0.0"
description = "d"
author = "a"
`

func writeExt(t *testing.T, dir, manifest, script string, extra map[string]string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	must := func(name, s string) {
		p := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must(ManifestFile, manifest)
	must("main.star", script)
	for k, v := range extra {
		must(k, v)
	}
}

type logSink struct {
	mu   sync.Mutex
	logs []string
}

func (l *logSink) fn(id, level, msg string) {
	l.mu.Lock()
	l.logs = append(l.logs, level+": "+msg)
	l.mu.Unlock()
}

func (l *logSink) has(sub string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, s := range l.logs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func openStarted(t *testing.T, manifest, script string, granted []Permission, o Options) *Extension {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "x")
	writeExt(t, dir, manifest, script, nil)
	e, err := OpenExtension(dir, o)
	if err != nil {
		t.Fatal(err)
	}
	e.SetGranted(granted)
	if err := e.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	return e
}

func TestManifestValidation(t *testing.T) {
	good := baseManifest + `keywords = ["t"]
permissions = ["network:api.github.com", "storage"]
[[permission_detail]]
permission = "network:api.github.com"
reason = "search"
[[settings]]
key = "limit"
type = "int"
default = 5
label = "Limit"
[[settings]]
key = "mode"
type = "choice"
options = ["a", "b"]
`
	m, err := ParseManifest([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if m.Entry != "main.star" || m.Reason("network:api.github.com") != "search" {
		t.Fatalf("bad parse: %+v", m)
	}
	s := m.ResolveSettings(map[string]any{"limit": float64(9), "mode": "zzz"})
	if s["limit"] != int64(9) || s["mode"] != "a" {
		t.Fatalf("settings: %v", s)
	}
	bad := []string{
		`id = "Bad ID"` + "\nname=\"x\"\nversion=\"1.0.0\"",
		baseManifest + `permissions = ["teleport"]`,
		baseManifest + `permissions = ["clipboard.read:foo"]`,
		strings.Replace(baseManifest, "1.0.0", "1.0", 1),
		baseManifest + "entry = \"../x.star\"",
		baseManifest + "[[settings]]\nkey=\"a\"\ntype=\"int\"\ndefault=\"str\"",
	}
	for i, b := range bad {
		if _, err := ParseManifest([]byte(b)); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestPermissionScopes(t *testing.T) {
	ps := PermissionSet{"network:api.github.com", Permission("fs.read:" + os.TempDir()), "process.run:git", "env.read:HOME"}
	cases := []struct {
		kind, target string
		want         bool
	}{
		{PermNetwork, "api.github.com", true},
		{PermNetwork, "v2.api.github.com", true},
		{PermNetwork, "evilapi.github.com", false},
		{PermNetwork, "example.com", false},
		{PermFSRead, filepath.Join(os.TempDir(), "a", "b"), true},
		{PermFSRead, filepath.Join(os.TempDir(), "..", "other"), false},
		{PermFSWrite, filepath.Join(os.TempDir(), "a"), false},
		{PermProcessRun, `C:\Program Files\Git\bin\git.exe`, true},
		{PermProcessRun, "powershell", false},
		{PermEnvRead, "home", true},
		{PermEnvRead, "PATH", false},
		{PermStorage, "", true},
		{PermClipboardRead, "", false},
	}
	for _, c := range cases {
		if got := ps.Allows(c.kind, c.target); got != c.want {
			t.Errorf("Allows(%s,%s)=%v want %v", c.kind, c.target, got, c.want)
		}
	}
}

func TestQueryRoundtrip(t *testing.T) {
	script := `
def on_load(ctx):
    log.info("loaded", ctx.id)

def on_query(ctx, query):
    out = []
    for i in range(2):
        out.append(item(
            id = "r%d" % i,
            title = "%s #%d" % (query, i),
            subtitle = ctx.version,
            action = action_call("pick", {"n": i, "q": query}),
            alt = [{"name": "Copy", "action": action_copy(query)}, ("Open", action_open("https://example.com"))],
        ))
    return out

def on_execute(ctx, item):
    storage.set("last", item.data)
    return action_copy("picked %d" % item.data["n"])
`
	var sink logSink
	var executed []core.Action
	exec := execFn(func(a core.Action) error { executed = append(executed, a); return nil })
	e := openStarted(t, baseManifest+"keywords = [\"t\"]\n", script,
		[]Permission{"clipboard.write", "open"}, Options{Log: sink.fn, Executor: exec})
	if !sink.has("loaded test.ext") {
		t.Fatalf("on_load did not run: %v", sink.logs)
	}
	p := NewProvider(e)
	if p.Name() != "Test Ext" || !p.Ranked() || len(p.Keywords()) != 1 {
		t.Fatalf("provider metadata wrong")
	}
	if got := p.Query(context.Background(), "hello"); len(got) != 0 {
		t.Fatalf("keyword gate failed: %v", got)
	}
	items := p.Query(context.Background(), "t foo bar")
	if len(items) != 2 {
		t.Fatalf("got %d items", len(items))
	}
	it := items[1]
	if it.ID != "ext:test.ext:r1" || it.Title != "foo bar #1" || it.Subtitle != "1.0.0" || it.Category != "Test Ext" {
		t.Fatalf("item: %+v", it)
	}
	if it.Action.Kind != KindExtCall || it.Action.Extra["ext"] != "test.ext" || it.Action.Extra["name"] != "pick" {
		t.Fatalf("action: %+v", it.Action)
	}
	if len(it.Alt) != 2 || it.Alt[0].Action.Kind != "copy" || it.Alt[1].Action.Target != "https://example.com" {
		t.Fatalf("alt: %+v", it.Alt)
	}
	if items[0].Score <= items[1].Score {
		t.Errorf("scores should keep order: %v %v", items[0].Score, items[1].Score)
	}
	if err := p.Execute(it.Action); err != nil {
		t.Fatal(err)
	}
	if len(executed) != 1 || executed[0].Kind != "copy" || executed[0].Target != "picked 1" {
		t.Fatalf("executed: %+v", executed)
	}
	v, err := e.Call(context.Background(), "on_query", starlark.String("z"))
	if err != nil || v == nil {
		t.Fatal(err)
	}
}

type execFn func(core.Action) error

func (f execFn) Execute(a core.Action) error { return f(a) }

func TestActionsDroppedWithoutPermission(t *testing.T) {
	script := `
def on_query(ctx, q):
    return [item("a", "A", action = action_open("https://x.y")), item("b", "B", action = action_run("calc.exe"))]
`
	var sink logSink
	e := openStarted(t, baseManifest, script, nil, Options{Log: sink.fn})
	items, err := e.Query(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.Action.Kind != "none" {
			t.Errorf("%s: action should be dropped, got %s", it.ID, it.Action.Kind)
		}
	}
	if !sink.has("permission denied") {
		t.Errorf("expected permission log")
	}
}

func TestHostFunctionsDenied(t *testing.T) {
	calls := map[string]string{
		"http":      `http.get("https://api.github.com/x")`,
		"fs.read":   `fs.read_text("C:/Windows/win.ini")`,
		"fs.write":  `fs.write_text("x.txt", "a")`,
		"clipboard": `clipboard.get()`,
		"clipset":   `clipboard.set("a")`,
		"proc":      `proc.run("cmd", ["/c", "echo"])`,
		"env":       `env.get("PATH")`,
		"notify":    `notify("a", "b")`,
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			e := openStarted(t, baseManifest, "def on_query(ctx, q):\n    "+call+"\n    return []\n", nil, Options{})
			_, err := e.Query(context.Background(), "")
			if err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("want permission denied, got %v", err)
			}
		})
	}
}

func TestScopedNetworkAndHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"items":[{"name":"zagros"}],"ua":"`+r.Header.Get("User-Agent")+`"}`)
	}))
	defer srv.Close()
	script := `
def on_query(ctx, q):
    d = http.json("` + srv.URL + `/x")
    r = http.get("` + srv.URL + `/y")
    return [item("1", d["items"][0]["name"], subtitle = str(r.status) + " " + r.json()["ua"])]
`
	// Granted for a different host: denied.
	e := openStarted(t, baseManifest, script, []Permission{"network:example.com"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("want denial, got %v", err)
	}
	// Granted for the right host.
	e = openStarted(t, baseManifest, script, []Permission{"network:127.0.0.1"}, Options{})
	items, err := e.Query(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Title != "zagros" || !strings.HasPrefix(items[0].Subtitle, "200 Zagros/test.ext/1.0.0") {
		t.Fatalf("%+v", items[0])
	}
	// Redirect to a non-granted host is blocked.
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://localhost:1/", http.StatusFound)
	}))
	defer redir.Close()
	e = openStarted(t, baseManifest, `def on_query(ctx, q):
    http.get("`+redir.URL+`")
    return []
`, []Permission{"network:127.0.0.1"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("redirect should be denied, got %v", err)
	}
}

func TestFSScope(t *testing.T) {
	allowed := t.TempDir()
	other := t.TempDir()
	os.WriteFile(filepath.Join(allowed, "a.txt"), []byte("hi"), 0o644)
	os.WriteFile(filepath.Join(other, "b.txt"), []byte("no"), 0o644)
	script := fmt.Sprintf(`
def on_query(ctx, q):
    t = fs.read_text(%q)
    fs.write_text(%q, "w")
    names = fs.list(%q)
    ex = fs.exists(%q)
    own = fs.exists(ctx.dir + "/main.star")
    if q == "bad":
        fs.read_text(%q)
    return [item("1", t, subtitle = ",".join(names) + str(ex) + str(own))]
`, filepath.ToSlash(filepath.Join(allowed, "a.txt")), filepath.ToSlash(filepath.Join(allowed, "w.txt")),
		filepath.ToSlash(allowed), filepath.ToSlash(filepath.Join(allowed, "a.txt")), filepath.ToSlash(filepath.Join(other, "b.txt")))
	e := openStarted(t, baseManifest, script, []Permission{Permission("fs.read:" + allowed), Permission("fs.write:" + allowed)}, Options{})
	items, err := e.Query(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Title != "hi" || !strings.Contains(items[0].Subtitle, "a.txt") || !strings.HasSuffix(items[0].Subtitle, "TrueTrue") {
		t.Fatalf("%+v", items[0])
	}
	if b, _ := os.ReadFile(filepath.Join(allowed, "w.txt")); string(b) != "w" {
		t.Fatal("write did not happen")
	}
	if _, err := e.Query(context.Background(), "bad"); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("out-of-scope read should be denied: %v", err)
	}
}

func TestStepLimitAndTimeout(t *testing.T) {
	loop := "def on_query(ctx, q):\n    while True:\n        pass\n"
	e := openStarted(t, baseManifest, loop, nil, Options{Limits: Limits{MaxSteps: 10_000, CallTimeout: 10 * time.Second}})
	start := time.Now()
	_, err := e.Query(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("want step limit error, got %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("step limit took too long")
	}
	e = openStarted(t, baseManifest, loop, nil, Options{Limits: Limits{MaxSteps: 1 << 60, CallTimeout: 100 * time.Millisecond}})
	_, err = e.Query(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("want timeout error, got %v", err)
	}
	// Context cancellation also stops the script.
	e = openStarted(t, baseManifest, loop, nil, Options{Limits: Limits{MaxSteps: 1 << 60, CallTimeout: time.Minute}})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := e.Query(ctx, ""); err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestStepLimitAtLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "x")
	writeExt(t, dir, baseManifest, "while True:\n    pass\n", nil)
	e, _ := OpenExtension(dir, Options{Limits: Limits{MaxSteps: 1000}})
	if err := e.Start(context.Background()); err == nil || e.Err() == nil {
		t.Fatal("load loop should fail")
	}
}

func TestStorageSettingsAndLoad(t *testing.T) {
	data := t.TempDir()
	manifest := baseManifest + "[[settings]]\nkey = \"greeting\"\ntype = \"string\"\ndefault = \"hi\"\n"
	script := `
load("lib.star", "shout")
changed = []
def on_query(ctx, q):
    n = storage.get("n", 0) + 1
    storage.set("n", n)
    return [item("1", shout(settings["greeting"]), subtitle = str(n) + " " + sha(q))]
def sha(s):
    return crypto.sha256(s)[:8] + "/" + crypto.hmac("k", s)[:4] + "/" + json.encode({"a": 1})
def on_settings_changed(ctx):
    log.info("settings changed", settings["greeting"])
`
	dir := filepath.Join(t.TempDir(), "x")
	writeExt(t, dir, manifest, script, map[string]string{"lib.star": "def shout(s):\n    return s.upper() + \"!\"\n"})
	var sink logSink
	o := Options{DataDir: data, Log: sink.fn}
	e, err := OpenExtension(dir, o)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	items, err := e.Query(context.Background(), "abc")
	if err != nil || items[0].Title != "HI!" || !strings.HasPrefix(items[0].Subtitle, "1 ba7816bf/") {
		t.Fatalf("%v %+v", err, items)
	}
	if err := e.SetSettings(context.Background(), map[string]any{"greeting": "yo"}); err != nil {
		t.Fatal(err)
	}
	if !sink.has("settings changed yo") {
		t.Fatalf("on_settings_changed missing: %v", sink.logs)
	}
	if err := e.SetSettings(context.Background(), map[string]any{"greeting": 5}); err == nil {
		t.Fatal("type error expected")
	}
	if err := e.SetSettings(context.Background(), map[string]any{"nope": "x"}); err == nil {
		t.Fatal("unknown key expected")
	}
	// A fresh instance sees persisted settings and storage.
	e2, _ := OpenExtension(dir, o)
	if err := e2.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	items, err = e2.Query(context.Background(), "abc")
	if err != nil || items[0].Title != "YO!" || !strings.HasPrefix(items[0].Subtitle, "2 ") {
		t.Fatalf("%v %+v", err, items)
	}
}

func TestStorageQuota(t *testing.T) {
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    storage.set("big", "x" * 5000)
    return []
`, nil, Options{Limits: Limits{MaxKVBytes: 100}})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "quota") {
		t.Fatalf("want quota error, got %v", err)
	}
}

// ---- archives ----

func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(body))
	}
	zw.Close()
	return buf.Bytes()
}

func newMgr(t *testing.T, p PermissionPrompter) (*Manager, string, string) {
	t.Helper()
	root := t.TempDir()
	ext, data := filepath.Join(root, "ext"), filepath.Join(root, "data")
	m, err := NewManager(Options{ExtDir: ext, DataDir: data, Prompter: p})
	if err != nil {
		t.Fatal(err)
	}
	return m, ext, data
}

const helloScript = "def on_query(ctx, q):\n    return [item(\"h\", \"Hello \" + q)]\n"

func TestZipSlipAndLimits(t *testing.T) {
	m, ext, _ := newMgr(t, AutoDeny{})
	ctx := context.Background()
	evil := []string{"../evil.txt", "a/../../evil.txt", "/abs.txt", `..\evil.txt`, "C:/evil.txt", "x/../../y"}
	for _, name := range evil {
		z := makeZip(t, map[string]string{ManifestFile: baseManifest, "main.star": helloScript, name: "pwn"})
		if _, err := m.InstallBytes(ctx, z, "t"); err == nil {
			t.Errorf("%q: expected rejection", name)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(ext), "evil.txt")); err == nil {
		t.Fatal("zip slip wrote outside")
	}
	if _, err := m.InstallBytes(ctx, makeZip(t, map[string]string{"main.star": helloScript}), "t"); err == nil {
		t.Error("missing manifest should fail")
	}
	if _, err := m.InstallBytes(ctx, makeZip(t, map[string]string{ManifestFile: baseManifest}), "t"); err == nil {
		t.Error("missing entry should fail")
	}
	if _, err := m.InstallBytes(ctx, []byte("not a zip"), "t"); err == nil {
		t.Error("garbage should fail")
	}
	big := strings.Repeat("a", MaxSingleFile+10)
	if _, err := m.InstallBytes(ctx, makeZip(t, map[string]string{ManifestFile: baseManifest, "main.star": helloScript, "big.bin": big}), "t"); err == nil {
		t.Error("oversized file should fail")
	}
	many := map[string]string{ManifestFile: baseManifest, "main.star": helloScript}
	for i := 0; i <= MaxArchiveFiles; i++ {
		many[fmt.Sprintf("f/%d.txt", i)] = "x"
	}
	if _, err := m.InstallBytes(ctx, makeZip(t, many), "t"); err == nil {
		t.Error("too many files should fail")
	}
	ents, _ := os.ReadDir(ext)
	for _, e := range ents {
		if e.Name() != "installed.json" {
			t.Errorf("leftover in ext dir: %s", e.Name())
		}
	}
}

func TestInstallLifecycle(t *testing.T) {
	var asked int
	prompter := PrompterFunc(func(m *Manifest, req []Permission) ([]Permission, error) {
		asked++
		return req, nil
	})
	m, ext, data := newMgr(t, prompter)
	ctx := context.Background()
	manifest := baseManifest + "keywords = [\"t\"]\npermissions = [\"clipboard.write\", \"storage\"]\n"
	// Wrapped top-level folder is accepted.
	z := makeZip(t, map[string]string{"pkg/" + ManifestFile: manifest, "pkg/main.star": helloScript, "pkg/assets/i.txt": "i"})
	e, err := m.InstallBytes(ctx, z, "test-source")
	if err != nil {
		t.Fatal(err)
	}
	if asked != 1 || !e.allows(PermClipboardWrite, "") || !e.Running() {
		t.Fatalf("consent/run state wrong: asked=%d", asked)
	}
	if _, err := os.Stat(filepath.Join(ext, "test.ext", "assets", "i.txt")); err != nil {
		t.Fatal(err)
	}
	rec := m.Records()["test.ext"]
	if rec.Source != "test-source" || rec.Version != "1.0.0" || len(rec.SHA256) != 64 {
		t.Fatalf("record: %+v", rec)
	}
	if got := m.Providers(); len(got) != 1 {
		t.Fatal("provider missing")
	}
	if _, err := os.Stat(filepath.Join(data, "permissions.json")); err != nil {
		t.Fatal("permissions.json not written")
	}
	// Re-discover in a new manager: grants persist, no new prompt.
	m2, err := NewManager(Options{ExtDir: ext, DataDir: data, Prompter: prompter})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m2.Discover(ctx); err != nil {
		t.Fatal(err)
	}
	e2, _ := m2.Get("test.ext")
	if asked != 1 || !e2.allows(PermClipboardWrite, "") {
		t.Fatalf("grants should persist (asked=%d)", asked)
	}
	// Upgrade with the same permission set: no re-consent. Changed set: re-consent.
	z2 := makeZip(t, map[string]string{ManifestFile: strings.Replace(manifest, "1.0.0", "1.1.0", 1), "main.star": helloScript})
	if _, err := m2.InstallBytes(ctx, z2, "s"); err != nil {
		t.Fatal(err)
	}
	if asked != 1 {
		t.Fatalf("same permission set must not re-prompt (asked=%d)", asked)
	}
	z3 := makeZip(t, map[string]string{ManifestFile: strings.Replace(manifest, "storage", "storage\", \"open", 1), "main.star": helloScript})
	e3, err := m2.InstallBytes(ctx, z3, "s")
	if err != nil {
		t.Fatal(err)
	}
	if asked != 2 || !e3.allows(PermOpen, "") {
		t.Fatalf("changed permission set must re-prompt (asked=%d)", asked)
	}
	// Disable / enable.
	if err := m2.Disable("test.ext"); err != nil {
		t.Fatal(err)
	}
	if len(m2.Providers()) != 0 {
		t.Fatal("disabled extension still provides")
	}
	m3, _ := NewManager(Options{ExtDir: ext, DataDir: data, Prompter: prompter})
	m3.Discover(ctx)
	if e, _ := m3.Get("test.ext"); e.Enabled() || e.Running() {
		t.Fatal("disabled state must persist and not run code")
	}
	if err := m3.Enable(ctx, "test.ext"); err != nil || len(m3.Providers()) != 1 {
		t.Fatalf("enable: %v", err)
	}
	// Uninstall with purge.
	if err := m3.Uninstall("test.ext", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(ext, "test.ext")); err == nil {
		t.Fatal("dir should be gone")
	}
	if _, ok := m3.Permissions().Get("test.ext"); ok {
		t.Fatal("grants should be forgotten on purge")
	}
	if len(m3.Records()) != 0 {
		t.Fatal("record should be gone")
	}
}

func TestAutoDenyGivesPermissionErrors(t *testing.T) {
	m, _, _ := newMgr(t, AutoDeny{})
	manifest := baseManifest + "permissions = [\"clipboard.read\"]\n"
	script := "def on_query(ctx, q):\n    return [item(\"a\", clipboard.get())]\n"
	e, err := m.InstallBytes(context.Background(), makeZip(t, map[string]string{ManifestFile: manifest, "main.star": script}), "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("got %v", err)
	}
	// Granted via AutoAllow + platform works.
	m2, _, _ := newMgr(t, AutoAllow{})
	m2.opts.Platform = fakePlatform{}
	e, err = m2.InstallBytes(context.Background(), makeZip(t, map[string]string{ManifestFile: manifest, "main.star": script}), "s")
	if err != nil {
		t.Fatal(err)
	}
	items, err := e.Query(context.Background(), "")
	if err != nil || items[0].Title != "clip" {
		t.Fatalf("%v %v", err, items)
	}
}

type fakePlatform struct{ NoPlatform }

func (fakePlatform) ClipboardGet() (string, error) { return "clip", nil }

func TestPackAndValidate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "src")
	writeExt(t, dir, baseManifest+"permissions = [\"network:example.com\"]\n", `def on_query(ctx, q):
    r = http.get("https://example.com")
    return [item("a", r.text)]
`, map[string]string{"assets/x.txt": "x", ".git/config": "secret", "old.zagros": "zip"})
	z, err := Pack(dir)
	if err != nil {
		t.Fatal(err)
	}
	z2, _ := Pack(dir)
	if !bytes.Equal(z, z2) {
		t.Error("pack should be deterministic")
	}
	zr, _ := zip.NewReader(bytes.NewReader(z), int64(len(z)))
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	if strings.Join(names, ",") != "assets/x.txt,main.star,manifest.toml" {
		t.Fatalf("names: %v", names)
	}
	m, _, _ := newMgr(t, AutoAllow{})
	if _, err := m.InstallBytes(context.Background(), z, "p"); err != nil {
		t.Fatal(err)
	}

	if is := Validate(dir); HasErrors(is) {
		t.Fatalf("unexpected: %v", is)
	}
	// Undeclared permission use, unknown name, syntax error.
	writeExt(t, dir, baseManifest, "def on_query(ctx, q):\n    http.get(\"x\")\n    return []\n", nil)
	var warned bool
	for _, i := range Validate(dir) {
		if !i.Error && strings.Contains(i.Msg, "not declared") {
			warned = true
		}
	}
	if !warned {
		t.Error("expected undeclared-permission warning")
	}
	writeExt(t, dir, baseManifest, "def on_query(ctx, q):\n    return [nope()]\n", nil)
	if !HasErrors(Validate(dir)) {
		t.Error("undefined name should be an error")
	}
	writeExt(t, dir, baseManifest, "def on_query(:\n", nil)
	if !HasErrors(Validate(dir)) {
		t.Error("syntax error should be an error")
	}
	writeExt(t, dir, "id = 1", "", nil)
	if !HasErrors(Validate(dir)) {
		t.Error("bad manifest should be an error")
	}
}

func TestHotReloadAndBroken(t *testing.T) {
	m, ext, _ := newMgr(t, AutoAllow{})
	ctx := context.Background()
	dir := filepath.Join(ext, "test.ext")
	writeExt(t, dir, baseManifest, helloScript, nil)
	writeExt(t, filepath.Join(ext, "broken"), "id = ", "", nil)
	if _, err := m.Discover(ctx); err != nil {
		t.Fatal(err)
	}
	if len(m.BrokenDirs()) != 1 {
		t.Fatalf("broken: %v", m.BrokenDirs())
	}
	e, _ := m.Get("test.ext")
	items, _ := e.Query(ctx, "a")
	if items[0].Title != "Hello a" {
		t.Fatal(items)
	}
	var changes int
	m.OnChange(func() { changes++ })
	time.Sleep(10 * time.Millisecond)
	writeExt(t, dir, baseManifest, "def on_query(ctx, q):\n    return [item(\"h\", \"Bye \" + q)]\n", nil)
	// Make sure mtime/size signature differs.
	os.Chtimes(filepath.Join(dir, "main.star"), time.Now().Add(time.Hour), time.Now().Add(time.Hour))
	m.pollOnce(ctx)
	e, _ = m.Get("test.ext")
	items, _ = e.Query(ctx, "a")
	if len(items) != 1 || items[0].Title != "Bye a" || changes == 0 {
		t.Fatalf("not reloaded: %v changes=%d", items, changes)
	}
	// New folder is picked up, removed folder dropped.
	writeExt(t, filepath.Join(ext, "other"), strings.Replace(baseManifest, "test.ext", "other.ext", 1), helloScript, nil)
	m.pollOnce(ctx)
	if _, ok := m.Get("other.ext"); !ok {
		t.Fatal("new extension not found")
	}
	os.RemoveAll(filepath.Join(ext, "other"))
	m.pollOnce(ctx)
	if _, ok := m.Get("other.ext"); ok {
		t.Fatal("removed extension still present")
	}
	if err := m.Reload(ctx, "test.ext"); err != nil {
		t.Fatal(err)
	}
}

func TestMinZagrosAndExecuteDispatch(t *testing.T) {
	m, ext, _ := newMgr(t, AutoAllow{})
	m.opts.AppVersion = "0.1.0"
	z := makeZip(t, map[string]string{ManifestFile: baseManifest + "min_zagros = \"9.0.0\"\n", "main.star": helloScript})
	if _, err := m.InstallBytes(context.Background(), z, "s"); err == nil {
		t.Fatal("min_zagros should block install")
	}
	_ = ext
	z = makeZip(t, map[string]string{ManifestFile: baseManifest, "main.star": "def on_query(ctx, q):\n    return [item(\"a\", \"A\", action = action_call(\"go\", {\"k\": [1, 2]}))]\ndef on_execute(ctx, item):\n    log.info(item.name, item.data[\"k\"][1])\n"})
	var sink logSink
	m.opts.Log = sink.fn
	e, err := m.InstallBytes(context.Background(), z, "s")
	if err != nil {
		t.Fatal(err)
	}
	items, _ := e.Query(context.Background(), "")
	if err := m.WrapExecutor(execFn(func(core.Action) error { t.Fatal("should not forward"); return nil })).Execute(items[0].Action); err != nil {
		t.Fatal(err)
	}
	if !sink.has("go 2") {
		t.Fatalf("on_execute did not run: %v", sink.logs)
	}
	if err := m.Execute(core.Action{Kind: "open"}); err == nil {
		t.Fatal("non ext-call must be rejected")
	}
}

func TestScopeFromExtension(t *testing.T) {
	script := `
def on_query(ctx, query):
    return [item(id = "s", title = "Search " + query, scope = scope(id = "engines", label = "Docs", placeholder = "Search docs", query = query, data = {"lang": "go"}))]

def on_scope(ctx, scope_id, query, data):
    return [item(id = "e1", title = scope_id + ":" + query + ":" + data["lang"], action = action_copy(query))]
`
	e := openStarted(t, baseManifest, script, []Permission{"clipboard.write"}, Options{})
	items, err := e.Query(context.Background(), "abc")
	if err != nil || len(items) != 1 || items[0].Scope == nil {
		t.Fatalf("item must carry a scope: %v %+v", err, items)
	}
	sc := items[0].Scope
	if sc.Label != "Docs" || sc.Placeholder != "Search docs" || sc.Query != "abc" || sc.Provider != e.Manifest.Name {
		t.Fatalf("scope: %+v", sc)
	}
	got, err := NewProvider(e).e.QueryScope(context.Background(), *sc, "xyz")
	if err != nil || len(got) != 1 || got[0].Title != "engines:xyz:go" {
		t.Fatalf("on_scope result: %v %+v", err, got)
	}
}

func TestHostModule(t *testing.T) {
	script := `
def on_query(ctx, query):
    return [item(id = "v", title = host.version() + "|" + str(host.compare_versions("0.2.0", "v0.10.0")) + "|" + str(host.compare_versions("1.0.0-rc.1", "1.0.0")) + "|" + str(host.compare_versions("2.0.0", "2.0.0")))]
`
	m := baseManifest
	e := openStarted(t, m, script, nil, Options{AppVersion: "0.2.0"})
	items, err := e.Query(context.Background(), "")
	if err != nil || len(items) != 1 || items[0].Title != "0.2.0|-1|-1|0" {
		t.Fatalf("host module: %v %+v", err, items)
	}
	bad := `
def on_query(ctx, query):
    return [item(id = "v", title = str(host.compare_versions("x", "1.0.0")))]
`
	e2 := openStarted(t, baseManifest, bad, nil, Options{})
	if _, err := e2.Query(context.Background(), ""); err == nil {
		t.Fatal("compare_versions must reject non-semver input")
	}
}
