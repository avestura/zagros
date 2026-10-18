package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/core"
)

const aiKey = "sk-secret-123456"

// fakeProvider is a scripted Anthropic or OpenAI-compatible server.
type fakeProvider struct {
	mu      sync.Mutex
	reqs    []map[string]any
	headers []http.Header
	paths   []string
	reply   func(n int, req map[string]any) (status int, body string)
	delay   time.Duration
}

func (f *fakeProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	var req map[string]any
	_ = json.Unmarshal(b, &req)
	f.mu.Lock()
	f.reqs = append(f.reqs, req)
	f.headers = append(f.headers, r.Header.Clone())
	f.paths = append(f.paths, r.URL.Path)
	n := len(f.reqs)
	f.mu.Unlock()
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	status, body := f.reply(n, req)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	io.WriteString(w, body)
}

func (f *fakeProvider) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.reqs) }

func anthropicText(text string, in, out int) string {
	return fmt.Sprintf(`{"model":"fake-1","content":[{"type":"text","text":%q}],"stop_reason":"end_turn","usage":{"input_tokens":%d,"output_tokens":%d}}`, text, in, out)
}

func aiOptions(srv *httptest.Server, kind string) Options {
	return Options{AI: func() core.AIConfig {
		return core.AIConfig{Providers: []core.AIProvider{
			{Name: "fake", Kind: kind, BaseURL: srv.URL, Model: "fake-1", EmbedModel: "embed-1"},
			{Name: "other", Kind: kind, BaseURL: srv.URL, Model: "fake-2"},
		}}
	}}
}

// aiScript runs body inside on_query of an extension declaring manifest extras.
func aiScript(t *testing.T, manifest, body string, granted []Permission, o Options) (string, error) {
	t.Helper()
	useFakeSecrets(t).m[aiKeyTarget("fake")] = aiKey
	if o.Log == nil {
		o.Log = func(string, string, string) {}
	}
	script := "def on_query(ctx, q):\n"
	for _, l := range strings.Split(strings.TrimSpace(body), "\n") {
		script += "    " + l + "\n"
	}
	script += "    return [item(\"1\", str(out))]\n"
	e := openStarted(t, manifest, script, granted, o)
	t.Cleanup(e.Stop)
	items, err := e.Query(context.Background(), "")
	if err != nil {
		return "", err
	}
	return items[0].Title, nil
}

const aiManifest = baseManifest + `permissions = ["ai"]
`

func TestAICompleteAnthropic(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) { return 200, anthropicText("hello there", 10, 5) }}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
r = ai.complete("hi", system="Be brief.", max_tokens=50, temperature=0.2)
out = [r["text"], r["usage"]["total_tokens"], r["stop_reason"], r["model"], r["provider"], r["tool_calls"]]
`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err != nil {
		t.Fatal(err)
	}
	if want := `["hello there", 15, "end", "fake-1", "fake", []]`; got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	req := fp.reqs[0]
	if fp.paths[0] != "/v1/messages" || fp.headers[0].Get("x-api-key") != aiKey || fp.headers[0].Get("anthropic-version") == "" {
		t.Errorf("anthropic request: %s %v", fp.paths[0], fp.headers[0])
	}
	if req["model"] != "fake-1" || req["system"] != "Be brief." || req["max_tokens"] != float64(50) || req["temperature"] != 0.2 {
		t.Errorf("body: %v", req)
	}
}

func TestAICompleteOpenAICompatible(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) {
		return 200, `{"model":"llama","choices":[{"message":{"content":"salam"},"finish_reason":"stop"}],"usage":{"prompt_tokens":7,"completion_tokens":3}}`
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
r = ai.complete([{"role": "system", "content": "Persian only"}, {"role": "user", "content": "hello"}])
out = [r["text"], r["usage"]["total_tokens"], r["stop_reason"]]
`, []Permission{"ai"}, aiOptions(srv, "openai"))
	if err != nil {
		t.Fatal(err)
	}
	if got != `["salam", 10, "end"]` {
		t.Fatalf("got %s", got)
	}
	if fp.paths[0] != "/v1/chat/completions" || fp.headers[0].Get("Authorization") != "Bearer "+aiKey {
		t.Errorf("openai request: %s %v", fp.paths[0], fp.headers[0])
	}
	msgs := fp.reqs[0]["messages"].([]any)
	if msgs[0].(map[string]any)["role"] != "system" || len(msgs) != 2 {
		t.Errorf("messages: %v", msgs)
	}
}

func TestAIRunExecutesToolsWithExtensionPermissions(t *testing.T) {
	fp := &fakeProvider{reply: func(n int, req map[string]any) (int, string) {
		if n == 1 {
			return 200, `{"model":"fake-1","content":[{"type":"text","text":"let me add"},{"type":"tool_use","id":"tu_1","name":"add","input":{"a":2.5,"b":4}},{"type":"tool_use","id":"tu_2","name":"read_file","input":{"path":"C:/Windows/win.ini"}}],"stop_reason":"tool_use","usage":{"input_tokens":20,"output_tokens":10}}`
		}
		return 200, anthropicText("the answer is 6.5", 40, 8)
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
def add(a, b):
    return a + b
def read_file(path):
    return fs.read_text(path)
tools = [
    ai.tool(add, description="Add two numbers", params={"a": "number", "b": "number"}),
    ai.tool(read_file, description="Read a file", params={"path": "string"}),
]
r = ai.run("add 2.5 and 4", tools=tools)
t = r["tool_calls"]
out = [r["text"], r["steps"], r["stop_reason"], r["usage"]["total_tokens"], t[0]["name"], t[0]["result"], t[1]["name"], t[1]["error"] != None]
`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, `["the answer is 6.5", 2, "end", 78, "add", "6.5", "read_file", True`) {
		t.Fatalf("got %s", got)
	}
	// The second request carries both tool results, the failing one as an error.
	second := fp.reqs[1]["messages"].([]any)
	last := second[len(second)-1].(map[string]any)["content"].([]any)
	if len(last) != 2 {
		t.Fatalf("tool results: %v", last)
	}
	denied := last[1].(map[string]any)
	if denied["is_error"] != true || !strings.Contains(denied["content"].(string), "permission denied") {
		t.Errorf("fs.read from a tool must hit the extension's own permission check: %v", denied)
	}
	tools := fp.reqs[0]["tools"].([]any)
	schema := tools[0].(map[string]any)["input_schema"].(map[string]any)
	if schema["type"] != "object" || len(schema["required"].([]any)) != 2 {
		t.Errorf("shorthand params became %v", schema)
	}
}

func TestAIRunOpenAIToolLoopAndMaxSteps(t *testing.T) {
	fp := &fakeProvider{reply: func(n int, req map[string]any) (int, string) {
		return 200, `{"model":"m","choices":[{"message":{"content":null,"tool_calls":[{"id":"c1","type":"function","function":{"name":"ping","arguments":"{\"n\": 1}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":5,"completion_tokens":5}}`
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
def ping(n):
    return {"pong": n}
r = ai.run("loop", tools=[ai.tool(ping, description="Ping", params={"n": "integer"})], max_steps=3)
out = [r["steps"], r["stop_reason"], r["usage"]["total_tokens"], r["tool_calls"][0]["result"]]
`, []Permission{"ai"}, aiOptions(srv, "openai"))
	if err != nil {
		t.Fatal(err)
	}
	if got != `[3, "max_steps", 30, "{\"pong\":1}"]` {
		t.Fatalf("got %s", got)
	}
	msgs := fp.reqs[1]["messages"].([]any)
	if m := msgs[len(msgs)-1].(map[string]any); m["role"] != "tool" || m["tool_call_id"] != "c1" {
		t.Errorf("tool message: %v", m)
	}
}

func TestAIRunRegistryTools(t *testing.T) {
	fp := &fakeProvider{reply: func(n int, req map[string]any) (int, string) {
		if n == 1 {
			return 200, `{"model":"fake-1","content":[{"type":"tool_use","id":"t1","name":"hash_hash","input":{"algo":"sha256","text":"hello"}}],"stop_reason":"tool_use","usage":{"input_tokens":5,"output_tokens":5}}`
		}
		return 200, anthropicText("done", 5, 5)
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
r = ai.run("hash hello", tools=ai.registry_tools(["hash.hash"]))
out = [r["text"], r["tool_calls"][0]["result"]]
`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824") {
		t.Fatalf("got %s", got)
	}
	tl := fp.reqs[0]["tools"].([]any)
	if len(tl) != 1 || tl[0].(map[string]any)["name"] != "hash_hash" {
		t.Errorf("tools sent: %v", tl)
	}
	if _, err := aiScript(t, aiManifest, `out = ai.registry_tools(["nope.nothing"])`, []Permission{"ai"}, aiOptions(srv, "anthropic")); err == nil || !strings.Contains(err.Error(), "no tools function or group") {
		t.Errorf("unknown registry name: %v", err)
	}
}

func TestAIPermissionScope(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) { return 200, anthropicText("ok", 1, 1) }}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	o := aiOptions(srv, "anthropic")
	cases := []struct {
		name    string
		granted []Permission
		body    string
		ok      bool
	}{
		{"none", nil, `out = ai.complete("x")["text"]`, false},
		{"scoped default", []Permission{"ai:fake"}, `out = ai.complete("x")["text"]`, true},
		{"scoped explicit", []Permission{"ai:fake"}, `out = ai.complete("x", provider="fake")["text"]`, true},
		{"wrong scope default", []Permission{"ai:other"}, `out = ai.complete("x")["text"]`, false},
		{"wrong scope explicit", []Permission{"ai:other"}, `out = ai.complete("x", provider="fake")["text"]`, false},
		{"unscoped", []Permission{"ai"}, `out = ai.complete("x", provider="other")["text"]`, true},
		{"embed needs it too", nil, `out = ai.embed("x")`, false},
		{"run needs it too", nil, `out = ai.run("x")`, false},
	}
	for _, c := range cases {
		_, err := aiScript(t, aiManifest, c.body, c.granted, o)
		if c.ok && err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
		if !c.ok && (err == nil || !strings.Contains(err.Error(), "permission denied")) {
			t.Errorf("%s: want permission denied, got %v", c.name, err)
		}
	}
	got, err := aiScript(t, aiManifest, `out = [p["name"] for p in ai.providers()]`, []Permission{"ai:other"}, o)
	if err != nil || got != `["other"]` {
		t.Errorf("providers() must list only permitted ones: %q %v", got, err)
	}
	for _, p := range []Permission{"ai", "ai:claude", "ai:my-local.2"} {
		if err := p.Validate(); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if (Permission("ai:")).Validate() == nil {
		t.Error("an empty scope must be invalid")
	}
}

func TestAIBudget(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) { return 200, anthropicText("ok", 60, 60) }}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	manifest := aiManifest + "ai_budget_tokens = 150\n"
	dir := t.TempDir()
	o := aiOptions(srv, "anthropic")
	o.DataDir = dir
	script := `
def on_query(ctx, q):
    r = ai.complete("hi", max_tokens=4000)
    return [item("1", r["text"])]
`
	useFakeSecrets(t).m[aiKeyTarget("fake")] = aiKey
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	old := aiNow
	aiNow = func() time.Time { return now }
	defer func() { aiNow = old }()

	e := openStarted(t, manifest, script, []Permission{"ai"}, o)
	defer e.Stop()
	q := func() error { _, err := e.Query(context.Background(), ""); return err }
	if err := q(); err != nil {
		t.Fatal(err)
	}
	if mt := fp.reqs[0]["max_tokens"].(float64); mt > 150 {
		t.Errorf("max_tokens %v exceeds the budget", mt)
	}
	if err := q(); err != nil { // 120 spent, 30 left, the call is allowed and overshoots
		t.Fatal(err)
	}
	err := q()
	if err == nil || !strings.Contains(err.Error(), "token budget used up") || !strings.Contains(err.Error(), "ai_budget_tokens") {
		t.Fatalf("third call must hit the budget, got %v", err)
	}
	n := fp.count()
	if q(); fp.count() != n {
		t.Error("a call over budget must not reach the provider")
	}
	// The count survives a restart (same data dir) and resets the next day.
	e2 := openStarted(t, manifest, script, []Permission{"ai"}, o)
	defer e2.Stop()
	if _, err := e2.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "token budget used up") {
		t.Fatalf("budget must persist: %v", err)
	}
	now = now.Add(24 * time.Hour)
	if _, err := e2.Query(context.Background(), ""); err != nil {
		t.Fatalf("budget must reset the next day: %v", err)
	}
}

func TestAIManifestBudget(t *testing.T) {
	m, err := ParseManifest([]byte(aiManifest + "ai_budget_tokens = 1234\n"))
	if err != nil || m.AIBudget() != 1234 {
		t.Fatalf("%v %+v", err, m)
	}
	m, _ = ParseManifest([]byte(aiManifest))
	if m.AIBudget() != aiDefaultBudget {
		t.Errorf("default budget %d", m.AIBudget())
	}
	if _, err := ParseManifest([]byte(baseManifest + "ai_budget_tokens = 10\n")); err == nil {
		t.Error("a budget without the ai permission must be rejected")
	}
	if _, err := ParseManifest([]byte(aiManifest + "ai_budget_tokens = -1\n")); err == nil {
		t.Error("a negative budget must be rejected")
	}
}

func TestAIErrorsNeverContainTheKey(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) {
		return 401, `{"error":{"type":"authentication_error","message":"invalid x-api-key ` + aiKey + `"}}`
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	logs := &logSink{}
	o := aiOptions(srv, "anthropic")
	o.Log = logs.fn
	_, err := aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, o)
	if err == nil || !strings.Contains(err.Error(), "HTTP 401") {
		t.Fatalf("want an HTTP 401 error, got %v", err)
	}
	if strings.Contains(err.Error(), aiKey) || !strings.Contains(err.Error(), "***") {
		t.Errorf("key leaked into the error: %v", err)
	}
	if logs.has(aiKey) {
		t.Error("key leaked into the log")
	}
}

func TestAIMissingKeyAndConfig(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) { return 200, anthropicText("ok", 1, 1) }}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	// no providers at all
	_, err := aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, Options{})
	if err == nil || !strings.Contains(err.Error(), "no AI provider is configured") {
		t.Errorf("no config: %v", err)
	}
	// unknown provider
	_, err = aiScript(t, aiManifest, `out = ai.complete("x", provider="nope")`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err == nil || !strings.Contains(err.Error(), "unknown AI provider") {
		t.Errorf("unknown: %v", err)
	}
	// remote host without a key
	o := Options{AI: func() core.AIConfig {
		return core.AIConfig{Providers: []core.AIProvider{{Name: "cloud", Kind: "anthropic", BaseURL: "https://api.example.com"}}}
	}}
	_, err = aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, o)
	if err == nil || !strings.Contains(err.Error(), "zagros ai key set cloud") {
		t.Errorf("missing key: %v", err)
	}
	// plain http to a remote host is refused before any key is sent
	o = Options{AI: func() core.AIConfig {
		return core.AIConfig{Providers: []core.AIProvider{{Name: "fake", Kind: "openai", BaseURL: "http://api.example.com"}}}
	}}
	_, err = aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, o)
	if err == nil || !strings.Contains(err.Error(), "https") {
		t.Errorf("http remote: %v", err)
	}
	// a local server needs no key
	local := httptest.NewServer(fp)
	defer local.Close()
	f := &fakeSecrets{m: map[string]string{}}
	old := secretBackend
	secretBackend = f
	defer func() { secretBackend = old }()
	e := openStarted(t, aiManifest, "def on_query(ctx, q):\n    return [item('1', ai.complete('x')['text'])]\n", []Permission{"ai"}, aiOptions(local, "anthropic"))
	defer e.Stop()
	if _, err := e.Query(context.Background(), ""); err != nil {
		t.Errorf("local without key: %v", err)
	}
	if fp.headers[len(fp.headers)-1].Get("x-api-key") != "" {
		t.Error("no key must mean no auth header")
	}
}

func TestAIEmbed(t *testing.T) {
	fp := &fakeProvider{reply: func(n int, req map[string]any) (int, string) {
		return 200, `{"data":[{"index":1,"embedding":[0.5,0.25]},{"index":0,"embedding":[1,2]}],"usage":{"prompt_tokens":4,"total_tokens":4}}`
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `out = ai.embed(["a", "b"])`, []Permission{"ai"}, aiOptions(srv, "openai"))
	if err != nil || got != `[[1.0, 2.0], [0.5, 0.25]]` {
		t.Fatalf("%q %v", got, err)
	}
	if fp.paths[0] != "/v1/embeddings" || fp.reqs[0]["model"] != "embed-1" {
		t.Errorf("request: %s %v", fp.paths[0], fp.reqs[0])
	}
	_, err = aiScript(t, aiManifest, `out = ai.embed("a")`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err == nil || !strings.Contains(err.Error(), "no embeddings API") {
		t.Errorf("anthropic embeddings must say so: %v", err)
	}
}

func TestAIWaitDoesNotCountAgainstTheHookTimeout(t *testing.T) {
	fp := &fakeProvider{delay: 700 * time.Millisecond, reply: func(int, map[string]any) (int, string) { return 200, anthropicText("slow", 1, 1) }}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	useFakeSecrets(t).m[aiKeyTarget("fake")] = aiKey
	o := aiOptions(srv, "anthropic")
	o.Limits = Limits{CallTimeout: 300 * time.Millisecond, MaxSteps: 1 << 60}
	e := openStarted(t, aiManifest, `
def on_query(ctx, q):
    r = ai.complete("x")
    return [item("1", r["text"])]
def spin(ctx):
    for i in range(100000000):
        pass
def wait_then_spin(ctx):
    ai.complete("x")
    spin(ctx)
`, []Permission{"ai"}, o)
	defer e.Stop()
	items, err := e.Query(context.Background(), "")
	if err != nil || items[0].Title != "slow" {
		t.Fatalf("a 700 ms provider wait must survive a 300 ms hook limit: %v %v", items, err)
	}
	if _, err := e.Call(context.Background(), "spin"); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Errorf("the hook limit still bounds extension work: %v", err)
	}
	if _, err := e.Call(context.Background(), "wait_then_spin"); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Errorf("work after the wait is still limited: %v", err)
	}
}

func TestAIRunCannotNestAndNeedsDescribedTools(t *testing.T) {
	fp := &fakeProvider{reply: func(n int, req map[string]any) (int, string) {
		if n == 1 {
			return 200, `{"model":"m","content":[{"type":"tool_use","id":"t","name":"inner","input":{}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`
		}
		return 200, anthropicText("done", 1, 1)
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	got, err := aiScript(t, aiManifest, `
def inner():
    return ai.run("again")
r = ai.run("go", tools=[ai.tool(inner, description="Nest")])
out = r["tool_calls"][0]["error"]
`, []Permission{"ai"}, aiOptions(srv, "anthropic"))
	if err != nil || !strings.Contains(got, "cannot be called from a tool") {
		t.Errorf("nested run: %q %v", got, err)
	}
	for body, want := range map[string]string{
		`def f():
    pass
out = ai.tool(f)`: "needs a description",
		`out = ai.run("x", tools=[len])`:                                                             "ai.tool",
		`out = ai.tool(len, description="d", params={"a": "text"})`:                                  "JSON type name",
		`out = ai.complete("x", tool_choice="zzz", tools=[ai.tool(len, name="n", description="d")])`: "not one of the tools",
		`out = ai.run("x", max_steps=99)`:                                                            "max_steps",
		`out = ai.complete("x", max_tokens=0)`:                                                       "max_tokens",
		`out = ai.complete([{"role": "robot", "content": "x"}])`:                                     "role must be",
	} {
		_, err := aiScript(t, aiManifest, body, []Permission{"ai"}, aiOptions(srv, "anthropic"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: want %q, got %v", body, want, err)
		}
	}
}

func TestAIResponseCap(t *testing.T) {
	fp := &fakeProvider{reply: func(int, map[string]any) (int, string) {
		return 200, anthropicText(strings.Repeat("x", 3000), 1, 1)
	}}
	srv := httptest.NewServer(fp)
	defer srv.Close()
	o := aiOptions(srv, "anthropic")
	o.Limits = Limits{MaxHTTPBytes: 1000}
	_, err := aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, o)
	if err == nil || !strings.Contains(err.Error(), "larger than 1000 bytes") {
		t.Errorf("oversized reply: %v", err)
	}
}

func TestAIRedirectsAreNotFollowed(t *testing.T) {
	var hits int
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.Header.Get("x-api-key") != "" {
			t.Error("the key followed a redirect")
		}
	}))
	defer evil.Close()
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, evil.URL+"/steal", http.StatusTemporaryRedirect)
	}))
	defer redir.Close()
	_, err := aiScript(t, aiManifest, `out = ai.complete("x")`, []Permission{"ai"}, aiOptions(redir, "anthropic"))
	if err == nil || hits != 0 {
		t.Errorf("redirect must be an error and must not be followed: hits=%d err=%v", hits, err)
	}
}

func TestAIStubScriptValidation(t *testing.T) {
	dir := t.TempDir()
	writeExt(t, dir, aiManifest, "def on_query(ctx, q):\n    return [item('1', ai.complete('x')['text'])]\n", nil)
	for _, is := range Validate(dir) {
		if strings.Contains(is.Msg, "ai.complete") {
			t.Errorf("declared permission must not warn: %v", is)
		}
	}
	writeExt(t, dir, baseManifest, "def on_query(ctx, q):\n    return [item('1', ai.complete('x')['text'])]\n", nil)
	found := false
	for _, is := range Validate(dir) {
		found = found || strings.Contains(is.Msg, `"ai"`)
	}
	if !found {
		t.Error("validate must flag ai use without the permission")
	}
}
