package mcp

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Supported MCP protocol revisions, newest first.
var protocolVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// Options configure a Server.
type Options struct {
	Index   *Index
	Guide   string // the extension-author guide (stubs/AGENTS.md)
	Version string // Zagros version reported to clients
	// Install installs an extension from a project folder or .zagros file and
	// returns a status line. Nil leaves the install_extension tool out.
	Install func(ctx context.Context, path string) (string, error)
}

// Server answers MCP requests. It is safe for concurrent use (it has no
// per-client state).
type Server struct {
	o Options
}

// New creates a server.
func New(o Options) *Server { return &Server{o: o} }

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// Handle processes one JSON-RPC message or batch and returns the response
// bytes, or nil when nothing is to be sent (notifications).
func (s *Server) Handle(msg []byte) []byte {
	msg = []byte(strings.TrimSpace(string(msg)))
	if len(msg) == 0 {
		return nil
	}
	if msg[0] == '[' {
		var batch []json.RawMessage
		if err := json.Unmarshal(msg, &batch); err != nil {
			return s.errResp(nil, -32700, "parse error")
		}
		var out []json.RawMessage
		for _, m := range batch {
			if r := s.handleOne(m); r != nil {
				out = append(out, r)
			}
		}
		if len(out) == 0 {
			return nil
		}
		b, _ := json.Marshal(out)
		return b
	}
	return s.handleOne(msg)
}

func (s *Server) errResp(id json.RawMessage, code int, msg string) []byte {
	if id == nil {
		id = json.RawMessage("null")
	}
	b, _ := json.Marshal(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
	return b
}

func (s *Server) handleOne(raw []byte) []byte {
	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil || req.Method == "" {
		return s.errResp(nil, -32600, "invalid request")
	}
	notification := len(req.ID) == 0
	result, rerr := s.dispatch(req)
	if notification {
		return nil
	}
	if rerr != nil {
		return s.errResp(req.ID, rerr.Code, rerr.Message)
	}
	b, _ := json.Marshal(rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result})
	return b
}

func (s *Server) dispatch(req rpcRequest) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		ver := protocolVersions[0]
		for _, v := range protocolVersions {
			if v == p.ProtocolVersion {
				ver = v
			}
		}
		return map[string]any{
			"protocolVersion": ver,
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
			"serverInfo":   map[string]any{"name": "zagros", "version": s.o.Version},
			"instructions": "Search and browse the Zagros extension API. Start with get_guide, then list_modules, search_api and get_api. Never guess function names.",
		}, nil
	case "notifications/initialized", "notifications/cancelled":
		return nil, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": toolDefs(s.o.Install != nil)}, nil
	case "tools/call":
		return s.callTool(req.Params)
	case "resources/list":
		return map[string]any{"resources": s.resourceList()}, nil
	case "resources/read":
		return s.readResource(req.Params)
	case "prompts/list":
		return map[string]any{"prompts": []any{map[string]any{
			"name":        "write_extension",
			"description": "How to write a Zagros extension for a goal, using this server to look up the API.",
			"arguments":   []any{map[string]any{"name": "goal", "description": "What the extension should do", "required": false}},
		}}}, nil
	case "prompts/get":
		var p struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &p)
		if p.Name != "write_extension" {
			return nil, &rpcError{-32602, "unknown prompt " + p.Name}
		}
		text := WriteExtensionPrompt(p.Arguments["goal"])
		return map[string]any{"description": "Write a Zagros extension", "messages": []any{
			map[string]any{"role": "user", "content": map[string]any{"type": "text", "text": text}}}}, nil
	}
	return nil, &rpcError{-32601, "method not found: " + req.Method}
}

// ---- tools ----

func schema(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func toolDefs(install bool) []any {
	str := func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
	defs := []any{
		map[string]any{"name": "get_guide", "description": "The guide for writing a Zagros extension: workflow, Starlark rules, permissions, UI mechanisms. Read this first.",
			"inputSchema": schema(map[string]any{})},
		map[string]any{"name": "list_modules", "description": "List every API module with what it covers and how many functions it has. Start here to traverse the API.",
			"inputSchema": schema(map[string]any{})},
		map[string]any{"name": "list_module", "description": "List all functions, classes and variables of one module with their signatures. Use module \"global\" for the host-level names (item, scope, action_open, hooks...).",
			"inputSchema": schema(map[string]any{"module": str("Module name, for example text, http, hash, or global")}, "module")},
		map[string]any{"name": "search_api", "description": "Search the extension API by keyword or task (\"base64 decode\", \"cidr subnet\", \"read clipboard\"). Returns ranked signatures with the permission each needs.",
			"inputSchema": schema(map[string]any{"query": str("What you are looking for"), "module": str("Optional module to restrict to"),
				"limit": map[string]any{"type": "integer", "description": "Maximum results (default 10)"}}, "query")},
		map[string]any{"name": "get_api", "description": "Full details of one API item: signature, parameters, the permission it needs, documentation and an example. Name it like http.get, text.lower or item.",
			"inputSchema": schema(map[string]any{"name": str("Qualified name such as hash.sha256, or a unique short name")}, "name")},
		map[string]any{"name": "get_permissions", "description": "Every permission an extension can declare, what it allows, and which API functions need it.",
			"inputSchema": schema(map[string]any{})},
		map[string]any{"name": "get_hooks", "description": "The hook functions an extension defines (on_query, on_scope, on_execute...) with their signatures.",
			"inputSchema": schema(map[string]any{})},
	}
	if install {
		defs = append(defs, map[string]any{"name": "install_extension", "description": "Install the extension you built into the user's Zagros, without the marketplace. Takes the project folder (with manifest.toml) or a .zagros archive. Permissions are not granted by this call: Zagros asks the user to approve them. Run validate first.",
			"inputSchema": schema(map[string]any{"path": str("Absolute path of the extension folder or .zagros file")}, "path")})
	}
	return defs
}

type toolResult struct {
	Content []map[string]any `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

func text(s string) toolResult {
	return toolResult{Content: []map[string]any{{"type": "text", "text": s}}}
}

func toolErr(s string) toolResult {
	r := text(s)
	r.IsError = true
	return r
}

func (s *Server) callTool(params json.RawMessage) (any, *rpcError) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{-32602, "bad params"}
	}
	arg := func(k string) string {
		if v, ok := p.Arguments[k].(string); ok {
			return v
		}
		return ""
	}
	idx := s.o.Index
	switch p.Name {
	case "get_guide":
		return text(s.o.Guide), nil
	case "list_modules":
		var b strings.Builder
		b.WriteString("Modules (use list_module, search_api or get_api to go deeper):\n")
		g := len(idx.InModule("global"))
		fmt.Fprintf(&b, "- global (%d names): item, scope, action_open/copy/run/call, notify, settings, hooks, permissions\n", g)
		for _, m := range idx.ModuleNames() {
			n := len(idx.InModule(m))
			fmt.Fprintf(&b, "- %s (%d): %s\n", m, n, firstLine(idx.Modules[m].Doc))
		}
		return text(b.String()), nil
	case "list_module":
		m := arg("module")
		if m == "" {
			return toolErr("module is required"), nil
		}
		ents := idx.InModule(m)
		if len(ents) == 0 {
			return toolErr(fmt.Sprintf("no module %q; call list_modules", m)), nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s: %s\n\n", m, firstLine(idx.Modules[m].Doc))
		for _, e := range ents {
			fmt.Fprintf(&b, "%s\n    %s  [needs: %s]\n", signatureLine(e), e.Summary(), orNone(e.Needs))
		}
		return text(b.String()), nil
	case "search_api":
		q := arg("query")
		if q == "" {
			return toolErr("query is required"), nil
		}
		limit := 10
		if v, ok := p.Arguments["limit"].(float64); ok && v > 0 {
			limit = int(v)
		}
		ms := idx.Search(q, arg("module"), limit)
		if len(ms) == 0 {
			return text(fmt.Sprintf("No API matches %q. Try other words or call list_modules.", q)), nil
		}
		var b strings.Builder
		for i, m := range ms {
			fmt.Fprintf(&b, "%d. %s\n    %s  [needs: %s]\n", i+1, signatureLine(m.Entry), m.Summary(), orNone(m.Needs))
		}
		b.WriteString("\nCall get_api with a name above for parameters, documentation and an example.")
		return text(b.String()), nil
	case "get_api":
		name := arg("name")
		e, ok := idx.Get(name)
		if !ok {
			ms := idx.Search(name, "", 5)
			var b strings.Builder
			fmt.Fprintf(&b, "No API item named %q.", name)
			if len(ms) > 0 {
				b.WriteString(" Did you mean:\n")
				for _, m := range ms {
					fmt.Fprintf(&b, "- %s\n", m.Qualified)
				}
			}
			return toolErr(b.String()), nil
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s\n\nPermission: %s\nStub file: %s\n\n%s\n", signatureLine(e), orNone(e.Needs), e.File, e.Doc)
		return text(b.String()), nil
	case "get_permissions":
		f, _ := idx.File("permissions.pyi")
		return text(f), nil
	case "get_hooks":
		f, _ := idx.File("hooks.pyi")
		return text(f), nil
	case "install_extension":
		if s.o.Install == nil {
			break
		}
		path := arg("path")
		if path == "" {
			return toolErr("path is required"), nil
		}
		msg, err := s.o.Install(context.Background(), path)
		if err != nil {
			return toolErr("install failed: " + err.Error()), nil
		}
		return text(msg), nil
	}
	return nil, &rpcError{-32602, "unknown tool " + p.Name}
}

func signatureLine(e Entry) string {
	switch e.Kind {
	case "function":
		return e.Qualified + signatureTail(e)
	default:
		return e.Qualified + "  (" + e.Kind + ")"
	}
}

// signatureTail returns "(args) -> ret" from "name(args) -> ret".
func signatureTail(e Entry) string {
	if i := strings.Index(e.Signature, "("); i >= 0 {
		return e.Signature[i:]
	}
	return "()"
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	return s
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// ---- resources ----

func (s *Server) resourceList() []any {
	out := []any{map[string]any{"uri": "zagros://guide", "name": "Extension author guide", "mimeType": "text/markdown"}}
	for _, n := range s.o.Index.FileNames() {
		out = append(out, map[string]any{"uri": "zagros://stubs/" + n, "name": "API stub " + n, "mimeType": "text/x-python"})
	}
	return out
}

func (s *Server) readResource(params json.RawMessage) (any, *rpcError) {
	var p struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{-32602, "bad params"}
	}
	var body, mime string
	switch {
	case p.URI == "zagros://guide":
		body, mime = s.o.Guide, "text/markdown"
	case strings.HasPrefix(p.URI, "zagros://stubs/"):
		f, ok := s.o.Index.File(strings.TrimPrefix(p.URI, "zagros://stubs/"))
		if !ok {
			return nil, &rpcError{-32002, "resource not found"}
		}
		body, mime = f, "text/x-python"
	default:
		return nil, &rpcError{-32002, "resource not found"}
	}
	return map[string]any{"contents": []any{map[string]any{"uri": p.URI, "mimeType": mime, "text": body}}}, nil
}

// ---- transports ----

// ServeStdio reads newline-delimited JSON-RPC messages from r and writes
// responses to w until r ends or ctx is cancelled. Nothing else may write to
// w (stdout), so diagnostics go to stderr.
func (s *Server) ServeStdio(ctx context.Context, r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4<<20)
	bw := bufio.NewWriter(w)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if resp := s.Handle(sc.Bytes()); resp != nil {
			bw.Write(resp)
			bw.WriteByte('\n')
			if err := bw.Flush(); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

func localHost(host string) bool {
	h := host
	if hh, _, err := net.SplitHostPort(host); err == nil {
		h = hh
	}
	h = strings.Trim(h, "[]")
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

// HTTPHandler serves MCP over HTTP at /mcp (POST, JSON in and out). It only
// answers requests addressed to localhost (a defence against DNS rebinding),
// rejects foreign Origin headers, and requires "Authorization: Bearer
// <token>" when token is set.
func (s *Server) HTTPHandler(token string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if !localHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if o := r.Header.Get("Origin"); o != "" {
			if u, err := url.Parse(o); err != nil || !localHost(u.Host) {
				http.Error(w, "forbidden origin", http.StatusForbidden)
				return
			}
		}
		if token != "" {
			got := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "use POST", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		resp := s.Handle(body)
		if resp == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	})
	return mux
}

// ListenAndServe runs the HTTP transport until ctx is cancelled. addr must be a
// loopback address ("127.0.0.1:7777"); anything else is refused.
func (s *Server) ListenAndServe(ctx context.Context, addr, token string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || !localHost(host) {
		return fmt.Errorf("mcp: refusing to listen on %q: use a loopback address such as 127.0.0.1:7777", addr)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: s.HTTPHandler(token)}
	go func() { <-ctx.Done(); _ = srv.Close() }()
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ---- prompts ----

// PromptOptions describe how the agent reaches the server.
type PromptOptions struct {
	Exe   string // path of zagros.exe, for the stdio command
	URL   string // http endpoint, for example http://127.0.0.1:7777/mcp ("" when HTTP is off)
	Token string // bearer token, if one is configured
}

// AgentPrompt is the text the user pastes into a coding agent so it can find
// and use the extension API.
func AgentPrompt(o PromptOptions) string {
	var b strings.Builder
	b.WriteString("You are helping me build a Zagros extension. Zagros is a keyboard launcher for Windows; extensions are small Starlark scripts (a Python dialect) plus a manifest.toml, and they run in a sandbox that needs approved permissions.\n\n")
	b.WriteString("Use the \"zagros\" MCP server to look up the extension API. Do not guess function names or signatures; look them up.\n\n")
	b.WriteString("How to use it:\n")
	b.WriteString("1. Call get_guide once. It explains the workflow, the Starlark rules and the permission model.\n")
	b.WriteString("2. Call list_modules to see the API areas: the host-level names (item, scope, actions, hooks) and modules such as http, fs, clipboard, text, encode, hash, net, dt, extract.\n")
	b.WriteString("3. Find what you need with search_api (for example \"base64 decode\" or \"cidr subnets\"), or traverse a module with list_module. Use get_api for the exact signature, parameters, the permission it needs, and an example.\n")
	b.WriteString("4. Declare only the permissions the functions you call actually need, as narrowly as possible (for example network:api.example.com). The tool modules (text, lines, num, math, encode, hash, cipher, rand, uuid, regex, net, dt, extract) are pure and need none.\n")
	b.WriteString("5. Start a project with `zagros ext init <folder>`, check it with `zagros ext validate .`, try it with `zagros --dev`, and when it works install it with the install_extension tool (or `zagros ext install .`). Zagros asks me to approve its permissions; do not try to grant them yourself. `zagros ext pack .` builds a .zagros file if I want to share it.\n\n")
	if o.Exe != "" {
		fmt.Fprintf(&b, "Connect over stdio (the agent starts the server itself):\n    %s mcp\n", quoteArg(o.Exe))
		fmt.Fprintf(&b, "For example with Claude Code:\n    claude mcp add zagros -- %s mcp\n\n", quoteArg(o.Exe))
	}
	if o.URL != "" {
		fmt.Fprintf(&b, "Or over HTTP (Zagros is running with its MCP server on):\n    %s\n", o.URL)
		if o.Token != "" {
			fmt.Fprintf(&b, "    send the header: Authorization: Bearer %s\n", o.Token)
		}
		fmt.Fprintf(&b, "For example with Claude Code:\n    claude mcp add --transport http zagros %s\n\n", o.URL)
	}
	b.WriteString("What I want to build: ")
	return b.String()
}

func quoteArg(s string) string {
	if strings.ContainsAny(s, " \t") {
		return `"` + s + `"`
	}
	return s
}

// WriteExtensionPrompt is the MCP prompt body: the same workflow plus the goal.
func WriteExtensionPrompt(goal string) string {
	p := AgentPrompt(PromptOptions{})
	if goal == "" {
		goal = "(ask me what the extension should do)"
	}
	return p + goal
}

// SortedKeys is a small helper for deterministic output.
func SortedKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
