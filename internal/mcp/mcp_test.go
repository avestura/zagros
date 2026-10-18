package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/avestura/zagros/stubs"
)

func newServer(t *testing.T) *Server {
	t.Helper()
	idx, err := BuildIndex(stubs.FS())
	if err != nil {
		t.Fatal(err)
	}
	return New(Options{Index: idx, Guide: string(stubs.AgentGuide()), Version: "0.0.0-test"})
}

func call(t *testing.T, s *Server, method string, params any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	out := s.Handle(body)
	if out == nil {
		t.Fatalf("%s: no response", method)
	}
	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func toolText(t *testing.T, s *Server, name string, args map[string]any) (string, bool) {
	t.Helper()
	resp := call(t, s, "tools/call", map[string]any{"name": name, "arguments": args})
	res, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("%s: %v", name, resp)
	}
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	isErr, _ := res["isError"].(bool)
	return content, isErr
}

func TestIndexParsesStubs(t *testing.T) {
	idx, err := BuildIndex(stubs.FS())
	if err != nil {
		t.Fatal(err)
	}
	item, ok := idx.Get("item")
	if !ok || item.Kind != "function" || !strings.Contains(item.Signature, "scope: Optional[Scope]") || item.Needs != "none" {
		t.Fatalf("item: %+v", item)
	}
	get, ok := idx.Get("http.get")
	if !ok || !strings.Contains(get.Needs, "network") || get.Module != "http" {
		t.Fatalf("http.get: %+v", get)
	}
	if _, ok := idx.Get("on_scope"); !ok {
		t.Fatal("hooks must be indexed")
	}
	if len(idx.ModuleNames()) < 8 {
		t.Fatalf("modules: %v", idx.ModuleNames())
	}
}

func TestSearchRanksExactAndDocMatches(t *testing.T) {
	idx, _ := BuildIndex(stubs.FS())
	m := idx.Search("clipboard", "", 5)
	if len(m) == 0 || !strings.HasPrefix(m[0].Qualified, "clipboard") {
		t.Fatalf("clipboard: %+v", m)
	}
	m = idx.Search("sub context tab", "", 5)
	found := false
	for _, x := range m {
		if x.Qualified == "scope" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a task description should find scope(): %+v", m)
	}
	if len(idx.Search("zzzzqqqq", "", 5)) != 0 {
		t.Fatal("nonsense must not match")
	}
	if got := idx.Search("get", "http", 10); len(got) == 0 || got[0].Module != "http" {
		t.Fatalf("module filter: %+v", got)
	}
}

func TestProtocolHandshakeAndTools(t *testing.T) {
	s := newServer(t)
	init := call(t, s, "initialize", map[string]any{"protocolVersion": "2025-03-26"})
	res := init["result"].(map[string]any)
	if res["protocolVersion"] != "2025-03-26" || res["serverInfo"].(map[string]any)["name"] != "zagros" {
		t.Fatalf("initialize: %v", res)
	}
	if s.Handle([]byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)) != nil {
		t.Fatal("notifications get no response")
	}
	list := call(t, s, "tools/list", nil)["result"].(map[string]any)["tools"].([]any)
	names := map[string]bool{}
	for _, x := range list {
		names[x.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"get_guide", "list_modules", "list_module", "search_api", "get_api", "get_permissions", "get_hooks"} {
		if !names[want] {
			t.Errorf("missing tool %s", want)
		}
	}
	if g, _ := toolText(t, s, "get_guide", nil); !strings.Contains(g, "Workflow") {
		t.Fatalf("guide: %.80s", g)
	}
	if txt, _ := toolText(t, s, "list_modules", nil); !strings.Contains(txt, "http") || !strings.Contains(txt, "global") {
		t.Fatalf("modules: %s", txt)
	}
	if txt, _ := toolText(t, s, "list_module", map[string]any{"module": "global"}); !strings.Contains(txt, "item(") || !strings.Contains(txt, "scope(") {
		t.Fatalf("global names: %s", txt)
	}
	txt, isErr := toolText(t, s, "get_api", map[string]any{"name": "http.get"})
	if isErr || !strings.Contains(txt, "Permission: network") {
		t.Fatalf("get_api: %v %s", isErr, txt)
	}
	if txt, isErr := toolText(t, s, "get_api", map[string]any{"name": "http.gett"}); !isErr || !strings.Contains(txt, "Did you mean") {
		t.Fatalf("typo should suggest: %v %s", isErr, txt)
	}
	if txt, _ := toolText(t, s, "search_api", map[string]any{"query": "notification toast"}); !strings.Contains(txt, "notify") {
		t.Fatalf("search: %s", txt)
	}
	if txt, _ := toolText(t, s, "get_permissions", nil); !strings.Contains(txt, "clipboard.read") {
		t.Fatalf("permissions: %.60s", txt)
	}
	if call(t, s, "nope", nil)["error"] == nil {
		t.Fatal("unknown method must be an error")
	}
}

func TestResourcesAndPrompts(t *testing.T) {
	s := newServer(t)
	res := call(t, s, "resources/list", nil)["result"].(map[string]any)["resources"].([]any)
	if len(res) < 5 {
		t.Fatalf("resources: %d", len(res))
	}
	r := call(t, s, "resources/read", map[string]any{"uri": "zagros://stubs/http.pyi"})["result"].(map[string]any)
	if !strings.Contains(r["contents"].([]any)[0].(map[string]any)["text"].(string), "def get(") {
		t.Fatal("stub text missing")
	}
	if call(t, s, "resources/read", map[string]any{"uri": "zagros://stubs/../secret"})["error"] == nil {
		t.Fatal("unknown resources must fail (no path traversal)")
	}
	p := call(t, s, "prompts/get", map[string]any{"name": "write_extension", "arguments": map[string]any{"goal": "convert units"}})["result"].(map[string]any)
	if !strings.Contains(p["messages"].([]any)[0].(map[string]any)["content"].(map[string]any)["text"].(string), "convert units") {
		t.Fatal("prompt must include the goal")
	}
}

func TestStdioTransport(t *testing.T) {
	s := newServer(t)
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n" +
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")
	var out bytes.Buffer
	if err := s.ServeStdio(context.Background(), in, &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 responses (the notification has none), got %d: %q", len(lines), out.String())
	}
}

func TestHTTPTransportSecurity(t *testing.T) {
	s := newServer(t)
	post := func(h http.Handler, host, origin, auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
		req.Host = host
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	open := s.HTTPHandler("")
	if r := post(open, "127.0.0.1:7777", "", ""); r.Code != 200 || !strings.Contains(r.Body.String(), `"result"`) {
		t.Fatalf("local request: %d %s", r.Code, r.Body.String())
	}
	if r := post(open, "evil.example:7777", "", ""); r.Code != http.StatusForbidden {
		t.Fatalf("a foreign Host (DNS rebinding) must be refused: %d", r.Code)
	}
	if r := post(open, "localhost:7777", "https://evil.example", ""); r.Code != http.StatusForbidden {
		t.Fatalf("a foreign Origin must be refused: %d", r.Code)
	}
	if r := post(open, "localhost:7777", "http://localhost:3000", ""); r.Code != 200 {
		t.Fatalf("a local Origin is fine: %d", r.Code)
	}
	locked := s.HTTPHandler("s3cret")
	if r := post(locked, "127.0.0.1:7777", "", ""); r.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: %d", r.Code)
	}
	if r := post(locked, "127.0.0.1:7777", "", "Bearer wrong"); r.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", r.Code)
	}
	if r := post(locked, "127.0.0.1:7777", "", "Bearer s3cret"); r.Code != 200 {
		t.Fatalf("right token: %d", r.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Host = "127.0.0.1:7777"
	rec := httptest.NewRecorder()
	open.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET: %d", rec.Code)
	}
	if err := s.ListenAndServe(context.Background(), "0.0.0.0:7777", ""); err == nil {
		t.Fatal("non-loopback listen must be refused")
	}
}

func TestAgentPrompt(t *testing.T) {
	p := AgentPrompt(PromptOptions{Exe: `C:\Program Files\Zagros\zagros.exe`, URL: "http://127.0.0.1:7777/mcp", Token: "tok"})
	for _, want := range []string{"get_guide", "search_api", "list_module", `"C:\Program Files\Zagros\zagros.exe" mcp`, "http://127.0.0.1:7777/mcp", "Bearer tok", "zagros ext init"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt lacks %q", want)
		}
	}
	if strings.Contains(AgentPrompt(PromptOptions{Exe: "z.exe"}), "HTTP") {
		t.Error("HTTP section must be omitted when the HTTP server is off")
	}
}
