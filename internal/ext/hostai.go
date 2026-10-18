package ext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/ai"
	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/tools"
	"go.starlark.net/starlark"
)

const (
	aiRequestTimeout = 60 * time.Second  // one request to a provider
	aiRunTimeout     = 120 * time.Second // one ai.run, requests only
	aiCallTotal      = 180 * time.Second // all provider waits inside one hook call
	aiDefaultBudget  = 50_000            // tokens per day when the manifest declares none
	aiMaxBudget      = 100_000_000
	aiMaxSteps       = 20
	aiMaxMessages    = 200
	aiMaxTokensArg   = 65536
	aiMaxTools       = 64
	aiMaxToolResult  = 64 << 10
	aiMaxEmbedTexts  = 256
	aiRunKey         = "zagros.ai.run"
)

var aiNameRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,40}$`)
var aiToolNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// aiNow is the clock the daily budget uses (tests replace it).
var aiNow = time.Now

// aiUsage counts the tokens an extension spent today. It is kept in
// <data>/ext/<id>/ai-usage.json so a restart does not reset the budget.
type aiUsage struct {
	mu     sync.Mutex
	loaded bool
	Day    string `json:"day"`
	Tokens int64  `json:"tokens"`
}

// AIBudget returns the daily token budget of the ai module for this extension.
func (m *Manifest) AIBudget() int64 {
	if m.AIBudgetTokens > 0 {
		return m.AIBudgetTokens
	}
	return aiDefaultBudget
}

func (e *Extension) aiBudget() int64 { return e.Manifest.AIBudget() }

func (u *aiUsage) roll(dir string) {
	if !u.loaded {
		u.loaded = true
		if dir != "" {
			if b, err := os.ReadFile(filepath.Join(dir, "ai-usage.json")); err == nil {
				_ = json.Unmarshal(b, u)
			}
		}
	}
	if day := aiNow().UTC().Format("2006-01-02"); u.Day != day {
		u.Day, u.Tokens = day, 0
	}
}

// aiRemaining returns the tokens left today, or the budget error.
func (e *Extension) aiRemaining() (int64, error) {
	u := &e.host.ai
	u.mu.Lock()
	defer u.mu.Unlock()
	u.roll(e.dataDir)
	left := e.aiBudget() - u.Tokens
	if left <= 0 {
		return 0, fmt.Errorf("ai: token budget used up: %s spent %d of %d tokens today (ai_budget_tokens in manifest.toml); it resets at midnight UTC", e.Manifest.ID, u.Tokens, e.aiBudget())
	}
	return left, nil
}

func (e *Extension) aiCharge(n int) {
	u := &e.host.ai
	u.mu.Lock()
	defer u.mu.Unlock()
	u.roll(e.dataDir)
	u.Tokens += int64(n)
	if e.dataDir != "" {
		b, _ := json.Marshal(u)
		_ = writeFileAtomic(filepath.Join(e.dataDir, "ai-usage.json"), b)
	}
}

// ---- providers ----

func aiKeyTarget(name string) string { return "zagros-ai|" + name }

// SetAIKey stores the API key of a configured provider in Credential Manager.
func SetAIKey(provider, key string) error {
	if !aiNameRe.MatchString(provider) {
		return fmt.Errorf("invalid provider name %q", provider)
	}
	if key == "" || len(key) > maxSecretValue {
		return fmt.Errorf("key must be 1 to %d bytes", maxSecretValue)
	}
	return secretBackend.Set(aiKeyTarget(provider), key)
}

// DeleteAIKey removes a provider's stored key.
func DeleteAIKey(provider string) error { return secretBackend.Delete(aiKeyTarget(provider)) }

// HasAIKey reports whether a key is stored for provider.
func HasAIKey(provider string) bool {
	_, ok, err := secretBackend.Get(aiKeyTarget(provider))
	return err == nil && ok
}

// ResolveAIProvider builds the provider named name from cfg with its stored
// key. It does no permission check.
func ResolveAIProvider(cfg core.AIConfig, name string) (ai.Provider, error) {
	if len(cfg.Providers) == 0 {
		return ai.Provider{}, errors.New("no AI provider is configured: add an [[ai.provider]] section to zagros.toml (tray menu, Settings)")
	}
	if name == "" {
		name = cfg.Default
	}
	if name == "" {
		name = cfg.Providers[0].Name
	}
	for _, c := range cfg.Providers {
		if c.Name != name {
			continue
		}
		p := ai.Provider{Name: c.Name, Kind: strings.ToLower(c.Kind), BaseURL: c.BaseURL, Model: c.Model, EmbedModel: c.EmbedModel}
		if !aiNameRe.MatchString(p.Name) {
			return p, fmt.Errorf("provider name %q may only use letters, digits, . _ -", p.Name)
		}
		if err := p.Validate(); err != nil {
			return p, err
		}
		key, ok, err := secretBackend.Get(aiKeyTarget(p.Name))
		switch {
		case err != nil && !errors.Is(err, errSecretsUnsupported):
			return p, fmt.Errorf("provider %q: reading the API key: %w", p.Name, err)
		case ok:
			p.APIKey = key
		case !p.KeyOptional():
			return p, fmt.Errorf("provider %q has no API key: run `zagros ai key set %s`", p.Name, p.Name)
		}
		return p, nil
	}
	return ai.Provider{}, fmt.Errorf("unknown AI provider %q", name)
}

func (e *Extension) aiConfig() core.AIConfig {
	if e.opts.AI == nil {
		return core.AIConfig{}
	}
	return e.opts.AI()
}

// aiProvider picks the provider and checks the permission for it.
func (e *Extension) aiProvider(name string) (ai.Provider, error) {
	cfg := e.aiConfig()
	if name != "" {
		if err := e.require(PermAI, name); err != nil {
			return ai.Provider{}, err
		}
	} else if len(cfg.Providers) > 0 {
		name = cfg.Default
		if name == "" {
			name = cfg.Providers[0].Name
		}
		if err := e.require(PermAI, name); err != nil {
			return ai.Provider{}, err
		}
	}
	return ResolveAIProvider(cfg, name)
}

// ---- module ----

func (e *Extension) aiModule() starlark.Value {
	return module("ai", starlark.StringDict{
		"complete":       e.fn("ai.complete", e.biAIComplete),
		"run":            e.fn("ai.run", e.biAIRun),
		"embed":          e.fn("ai.embed", e.biAIEmbed),
		"providers":      e.fn("ai.providers", e.biAIProviders),
		"tool":           e.fn("ai.tool", biAITool),
		"registry_tools": e.fn("ai.registry_tools", biAIRegistryTools),
	})
}

func (e *Extension) biAIProviders(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("ai.providers", args, kw); err != nil {
		return nil, err
	}
	if err := e.require(PermAI, ""); err != nil && !e.allowsAnyAI() {
		return nil, err
	}
	cfg := e.aiConfig()
	var out []starlark.Value
	for _, c := range cfg.Providers {
		if !e.allows(PermAI, c.Name) {
			continue
		}
		def := c.Name == cfg.Default || (cfg.Default == "" && c.Name == cfg.Providers[0].Name)
		d := starlark.NewDict(5)
		_ = d.SetKey(starlark.String("name"), starlark.String(c.Name))
		_ = d.SetKey(starlark.String("kind"), starlark.String(strings.ToLower(c.Kind)))
		_ = d.SetKey(starlark.String("model"), starlark.String(c.Model))
		_ = d.SetKey(starlark.String("embeddings"), starlark.Bool(strings.EqualFold(c.Kind, ai.KindOpenAI)))
		_ = d.SetKey(starlark.String("default"), starlark.Bool(def))
		out = append(out, d)
	}
	return starlark.NewList(out), nil
}

// allowsAnyAI reports whether any ai permission, scoped or not, is granted.
func (e *Extension) allowsAnyAI() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, p := range e.granted {
		if p.Kind() == PermAI {
			return true
		}
	}
	return false
}

func optFloat(v starlark.Value, what string) (*float64, error) {
	if v == nil || v == starlark.None {
		return nil, nil
	}
	switch x := v.(type) {
	case starlark.Float:
		f := float64(x)
		return &f, nil
	case starlark.Int:
		f := float64(x.Float())
		return &f, nil
	}
	return nil, fmt.Errorf("%s must be a number, got %s", what, v.Type())
}

func (e *Extension) biAIComplete(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var msgs, toolsV, tempV starlark.Value
	var model, system, choice, provider string
	maxTokens := 1024
	if err := starlark.UnpackArgs("ai.complete", args, kw, "messages", &msgs, "model?", &model, "system?", &system,
		"max_tokens?", &maxTokens, "temperature?", &tempV, "tools?", &toolsV, "tool_choice?", &choice, "provider?", &provider); err != nil {
		return nil, err
	}
	if maxTokens < 1 || maxTokens > aiMaxTokensArg {
		return nil, fmt.Errorf("ai.complete: max_tokens must be 1 to %d", aiMaxTokensArg)
	}
	temp, err := optFloat(tempV, "temperature")
	if err != nil {
		return nil, fmt.Errorf("ai.complete: %w", err)
	}
	p, err := e.aiProvider(provider)
	if err != nil {
		return nil, err
	}
	ms, sys, err := e.aiMessages(th, msgs)
	if err != nil {
		return nil, fmt.Errorf("ai.complete: %w", err)
	}
	if system != "" {
		sys = system + "\n\n" + sys
	}
	tl, err := aiToolList(toolsV)
	if err != nil {
		return nil, fmt.Errorf("ai.complete: %w", err)
	}
	if choice != "" && choice != "auto" && choice != "none" && choice != "required" {
		found := false
		for _, t := range tl {
			found = found || t.def.Name == choice
		}
		if !found {
			return nil, fmt.Errorf("ai.complete: tool_choice %q is not one of the tools", choice)
		}
	}
	req := ai.Request{Model: model, System: strings.TrimSpace(sys), Messages: ms, MaxTokens: maxTokens, Temperature: temp, ToolChoice: choice}
	for _, t := range tl {
		req.Tools = append(req.Tools, t.def)
	}
	res, err := e.aiCall(th, p, req)
	if err != nil {
		return nil, err
	}
	return aiResult(res, p, nil)
}

// aiCall runs one provider request under the budget and the clock rules.
func (e *Extension) aiCall(th *starlark.Thread, p ai.Provider, req ai.Request) (*ai.Response, error) {
	left, err := e.aiRemaining()
	if err != nil {
		return nil, err
	}
	est := int64(aiEstimate(req) / 4)
	if est >= left {
		return nil, fmt.Errorf("ai: this request is about %d tokens and only %d are left in today's budget (ai_budget_tokens)", est, left)
	}
	if room := left - est; int64(req.MaxTokens) > room {
		req.MaxTokens = int(room)
	}
	resume, err := clockOf(th).pause(aiCallTotal)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(callCtx(th), aiRequestTimeout)
	res, err := ai.Complete(ctx, e.opts.HTTPClient, p, req, e.opts.Limits.MaxHTTPBytes)
	cancel()
	resume()
	if err != nil {
		return nil, fmt.Errorf("ai: %w", err)
	}
	used := res.Usage.Total()
	if used == 0 { // some local servers report no usage
		used = int(est) + len(res.Text)/4 + 1
		res.Usage = ai.Usage{Input: int(est), Output: used - int(est)}
	}
	e.aiCharge(used)
	e.opts.Log(e.Manifest.ID, "info", fmt.Sprintf("ai: provider=%s model=%s tokens=%d", p.Name, res.Model, used))
	return res, nil
}

func aiEstimate(req ai.Request) int {
	n := len(req.System)
	for _, m := range req.Messages {
		n += len(m.Content)
		for _, c := range m.ToolCalls {
			b, _ := json.Marshal(c.Args)
			n += len(c.Name) + len(b)
		}
	}
	for _, t := range req.Tools {
		b, _ := json.Marshal(t.Schema)
		n += len(t.Name) + len(t.Description) + len(b)
	}
	return n
}

func aiResult(res *ai.Response, p ai.Provider, extra map[string]starlark.Value) (starlark.Value, error) {
	d := starlark.NewDict(8)
	set := func(k string, v starlark.Value) { _ = d.SetKey(starlark.String(k), v) }
	set("text", starlark.String(res.Text))
	calls := make([]starlark.Value, 0, len(res.ToolCalls))
	for _, c := range res.ToolCalls {
		cd := starlark.NewDict(3)
		_ = cd.SetKey(starlark.String("id"), starlark.String(c.ID))
		_ = cd.SetKey(starlark.String("name"), starlark.String(c.Name))
		_ = cd.SetKey(starlark.String("args"), goToStarlark(c.Args))
		calls = append(calls, cd)
	}
	set("tool_calls", starlark.NewList(calls))
	set("usage", aiUsageValue(res.Usage))
	set("stop_reason", starlark.String(res.StopReason))
	set("model", starlark.String(res.Model))
	set("provider", starlark.String(p.Name))
	for k, v := range extra {
		set(k, v)
	}
	return d, nil
}

func aiUsageValue(u ai.Usage) starlark.Value {
	d := starlark.NewDict(3)
	_ = d.SetKey(starlark.String("input_tokens"), starlark.MakeInt(u.Input))
	_ = d.SetKey(starlark.String("output_tokens"), starlark.MakeInt(u.Output))
	_ = d.SetKey(starlark.String("total_tokens"), starlark.MakeInt(u.Total()))
	return d
}

// ---- messages ----

func dictStr(d *starlark.Dict, key string) (string, bool, error) {
	v, ok, _ := d.Get(starlark.String(key))
	if !ok || v == starlark.None {
		return "", false, nil
	}
	s, ok := starlark.AsString(v)
	if !ok {
		return "", false, fmt.Errorf("%q must be a string, got %s", key, v.Type())
	}
	return s, true, nil
}

// aiMessages converts a prompt string or a list of message dicts. System
// messages are folded into the returned system text.
func (e *Extension) aiMessages(th *starlark.Thread, v starlark.Value) ([]ai.Message, string, error) {
	limit := int(e.opts.Limits.MaxFileBytes)
	if s, ok := starlark.AsString(v); ok {
		if s == "" {
			return nil, "", errors.New("the prompt is empty")
		}
		if len(s) > limit {
			return nil, "", fmt.Errorf("the prompt is larger than %d bytes", limit)
		}
		return []ai.Message{{Role: "user", Content: s}}, "", nil
	}
	seq, ok := v.(starlark.Sequence)
	if !ok {
		return nil, "", fmt.Errorf("messages must be a string or a list of {role, content} dicts, got %s", v.Type())
	}
	if seq.Len() == 0 || seq.Len() > aiMaxMessages {
		return nil, "", fmt.Errorf("messages must hold 1 to %d entries", aiMaxMessages)
	}
	var out []ai.Message
	var sys []string
	size := 0
	it := seq.Iterate()
	defer it.Done()
	var x starlark.Value
	for i := 0; it.Next(&x); i++ {
		d, ok := x.(*starlark.Dict)
		if !ok {
			return nil, "", fmt.Errorf("message %d must be a dict, got %s", i, x.Type())
		}
		role, _, err := dictStr(d, "role")
		if err != nil {
			return nil, "", fmt.Errorf("message %d: %w", i, err)
		}
		content, _, err := dictStr(d, "content")
		if err != nil {
			return nil, "", fmt.Errorf("message %d: %w", i, err)
		}
		size += len(content)
		if size > limit {
			return nil, "", fmt.Errorf("the messages are larger than %d bytes", limit)
		}
		m := ai.Message{Role: role, Content: content}
		switch role {
		case "system":
			sys = append(sys, content)
			continue
		case "user":
		case "assistant":
			if tc, ok, _ := d.Get(starlark.String("tool_calls")); ok && tc != starlark.None {
				calls, err := aiCallsFrom(th, tc)
				if err != nil {
					return nil, "", fmt.Errorf("message %d: %w", i, err)
				}
				m.ToolCalls = calls
			}
		case "tool":
			id, ok, err := dictStr(d, "tool_call_id")
			if err != nil || !ok || id == "" {
				return nil, "", fmt.Errorf("message %d: a tool message needs tool_call_id", i)
			}
			m.ToolCallID = id
			if b, ok, _ := d.Get(starlark.String("is_error")); ok {
				m.IsError = bool(b.Truth())
			}
		default:
			return nil, "", fmt.Errorf("message %d: role must be user, assistant, tool or system, got %q", i, role)
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil, "", errors.New("messages hold no user or assistant turn")
	}
	return out, strings.Join(sys, "\n\n"), nil
}

func aiCallsFrom(th *starlark.Thread, v starlark.Value) ([]ai.ToolCall, error) {
	seq, ok := v.(starlark.Sequence)
	if !ok {
		return nil, errors.New("tool_calls must be a list")
	}
	var out []ai.ToolCall
	it := seq.Iterate()
	defer it.Done()
	var x starlark.Value
	for it.Next(&x) {
		d, ok := x.(*starlark.Dict)
		if !ok {
			return nil, errors.New("each tool call must be a dict with id, name and args")
		}
		id, _, err1 := dictStr(d, "id")
		name, _, err2 := dictStr(d, "name")
		if err1 != nil || err2 != nil || id == "" || name == "" {
			return nil, errors.New("each tool call needs id and name")
		}
		args := map[string]any{}
		if a, ok, _ := d.Get(starlark.String("args")); ok && a != starlark.None {
			enc, err := jsonCall(th, "encode", a)
			if err != nil {
				return nil, fmt.Errorf("tool call args: %w", err)
			}
			if err := json.Unmarshal([]byte(string(enc.(starlark.String))), &args); err != nil {
				return nil, errors.New("tool call args must be a dict")
			}
		}
		out = append(out, ai.ToolCall{ID: id, Name: name, Args: args})
	}
	return out, nil
}

// ---- tools ----

// aiTool is the value ai.tool returns: a Starlark function plus the schema
// the model sees.
type aiTool struct {
	def ai.Tool
	fn  starlark.Callable
}

func (t *aiTool) String() string        { return "<ai.tool " + t.def.Name + ">" }
func (t *aiTool) Type() string          { return "ai.tool" }
func (t *aiTool) Freeze()               {}
func (t *aiTool) Truth() starlark.Bool  { return true }
func (t *aiTool) Hash() (uint32, error) { return 0, errors.New("unhashable type: ai.tool") }
func (t *aiTool) AttrNames() []string   { return []string{"description", "name", "params"} }
func (t *aiTool) Attr(n string) (starlark.Value, error) {
	switch n {
	case "name":
		return starlark.String(t.def.Name), nil
	case "description":
		return starlark.String(t.def.Description), nil
	case "params":
		return goToStarlark(t.def.Schema), nil
	}
	return nil, nil
}

var jsonTypes = map[string]bool{"string": true, "number": true, "integer": true, "boolean": true, "array": true, "object": true}

// biAITool is ai.tool(fn, name="", description="", params=None).
func biAITool(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var fnV, paramsV starlark.Value = nil, starlark.None
	var name, desc string
	if err := starlark.UnpackArgs("ai.tool", args, kw, "fn", &fnV, "name?", &name, "description?", &desc, "params?", &paramsV); err != nil {
		return nil, err
	}
	fn, ok := fnV.(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("ai.tool: fn must be a function, got %s", fnV.Type())
	}
	if name == "" {
		name = fn.Name()
	}
	if !aiToolNameRe.MatchString(name) {
		return nil, fmt.Errorf("ai.tool: name %q must be 1 to 64 letters, digits, _ or - (pass name=)", name)
	}
	if desc == "" {
		if f, ok := fn.(*starlark.Function); ok {
			desc = strings.TrimSpace(f.Doc())
		}
	}
	if desc == "" {
		return nil, fmt.Errorf("ai.tool: %q needs a description (pass description= or give the function a docstring); the model decides from it when to call the tool", name)
	}
	schema := map[string]any{"type": "object", "properties": map[string]any{}}
	if paramsV != starlark.None {
		d, ok := paramsV.(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("ai.tool: params must be a dict, got %s", paramsV.Type())
		}
		enc, err := jsonCall(th, "encode", d)
		if err != nil {
			return nil, fmt.Errorf("ai.tool: params: %w", err)
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(string(enc.(starlark.String))), &raw); err != nil {
			return nil, fmt.Errorf("ai.tool: params: %w", err)
		}
		if raw["type"] == "object" {
			schema = raw
			if _, ok := schema["properties"]; !ok {
				schema["properties"] = map[string]any{}
			}
		} else { // shorthand: {"city": "string", "days": "integer"}, all required
			props := map[string]any{}
			var req []string
			for k, v := range raw {
				t, ok := v.(string)
				if !ok || !jsonTypes[t] {
					return nil, fmt.Errorf("ai.tool: params[%q] must be a JSON type name (string, number, integer, boolean, array, object); for a full schema pass {\"type\": \"object\", \"properties\": {...}}", k)
				}
				props[k] = map[string]any{"type": t}
				req = append(req, k)
			}
			sort.Strings(req)
			schema["properties"] = props
			if len(req) > 0 {
				schema["required"] = req
			}
		}
	}
	return &aiTool{def: ai.Tool{Name: name, Description: desc, Schema: schema}, fn: fn}, nil
}

func aiToolList(v starlark.Value) ([]*aiTool, error) {
	if v == nil || v == starlark.None {
		return nil, nil
	}
	seq, ok := v.(starlark.Sequence)
	if !ok {
		return nil, fmt.Errorf("tools must be a list of ai.tool(...) values, got %s", v.Type())
	}
	if seq.Len() > aiMaxTools {
		return nil, fmt.Errorf("at most %d tools", aiMaxTools)
	}
	var out []*aiTool
	seen := map[string]bool{}
	it := seq.Iterate()
	defer it.Done()
	var x starlark.Value
	for it.Next(&x) {
		t, ok := x.(*aiTool)
		if !ok {
			if fn, isFn := x.(starlark.Callable); isFn {
				return nil, fmt.Errorf("tools holds a bare function (%s): wrap it with ai.tool(fn, description=..., params=...)", fn.Name())
			}
			return nil, fmt.Errorf("tools must hold ai.tool(...) values, got %s", x.Type())
		}
		if seen[t.def.Name] {
			return nil, fmt.Errorf("two tools are named %q", t.def.Name)
		}
		seen[t.def.Name] = true
		out = append(out, t)
	}
	return out, nil
}

var toolParamTypes = map[string]map[string]any{
	"str":            {"type": "string"},
	"int":            {"type": "integer"},
	"float":          {"type": "number"},
	"bool":           {"type": "boolean"},
	"list[str]":      {"type": "array", "items": map[string]any{"type": "string"}},
	"list[float]":    {"type": "array", "items": map[string]any{"type": "number"}},
	"dict[str, str]": {"type": "object", "additionalProperties": map[string]any{"type": "string"}},
	"any":            {},
}

// registryTool turns a pure tools-registry function into a model tool.
func registryTool(f tools.Func) *aiTool {
	props := map[string]any{}
	var req []string
	for _, p := range f.Params {
		s := map[string]any{}
		for k, v := range toolParamTypes[p.Type] {
			s[k] = v
		}
		if p.Doc != "" {
			s["description"] = p.Doc
		}
		if len(p.Choices) > 0 {
			s["enum"] = p.Choices
		}
		props[p.Name] = s
		if p.Required {
			req = append(req, p.Name)
		}
	}
	schema := map[string]any{"type": "object", "properties": props}
	if len(req) > 0 {
		schema["required"] = req
	}
	desc := f.Doc
	if desc == "" {
		desc = f.Title
	}
	return &aiTool{def: ai.Tool{Name: f.Group + "_" + f.Name, Description: desc, Schema: schema}, fn: toolBuiltin(f)}
}

// biAIRegistryTools is ai.registry_tools(names): "group.name" or a whole group.
func biAIRegistryTools(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var namesV starlark.Value
	if err := starlark.UnpackArgs("ai.registry_tools", args, kw, "names", &namesV); err != nil {
		return nil, err
	}
	names, err := stringList(namesV)
	if err != nil {
		return nil, fmt.Errorf("ai.registry_tools: %w", err)
	}
	all := tools.All()
	var out []starlark.Value
	seen := map[string]bool{}
	add := func(f tools.Func) {
		if !seen[f.Key()] {
			seen[f.Key()] = true
			out = append(out, registryTool(f))
		}
	}
	for _, n := range names {
		matched := false
		for _, f := range all {
			if f.Key() == n || f.Group == n {
				add(f)
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("ai.registry_tools: no tools function or group %q (use group.name such as hash.sha256, or a group such as hash)", n)
		}
	}
	if len(out) > aiMaxTools {
		return nil, fmt.Errorf("ai.registry_tools: %d functions selected, at most %d (name them individually)", len(out), aiMaxTools)
	}
	return starlark.NewList(out), nil
}

// ---- ai.run ----

func (e *Extension) biAIRun(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var msgs, toolsV, tempV starlark.Value
	var model, system, provider string
	maxSteps, maxTokens := 8, 1024
	if err := starlark.UnpackArgs("ai.run", args, kw, "prompt", &msgs, "tools?", &toolsV, "model?", &model, "system?", &system,
		"max_steps?", &maxSteps, "max_tokens?", &maxTokens, "temperature?", &tempV, "provider?", &provider); err != nil {
		return nil, err
	}
	if th.Local(aiRunKey) != nil {
		return nil, errAINested
	}
	if maxSteps < 1 || maxSteps > aiMaxSteps {
		return nil, fmt.Errorf("ai.run: max_steps must be 1 to %d", aiMaxSteps)
	}
	if maxTokens < 1 || maxTokens > aiMaxTokensArg {
		return nil, fmt.Errorf("ai.run: max_tokens must be 1 to %d", aiMaxTokensArg)
	}
	temp, err := optFloat(tempV, "temperature")
	if err != nil {
		return nil, fmt.Errorf("ai.run: %w", err)
	}
	p, err := e.aiProvider(provider)
	if err != nil {
		return nil, err
	}
	ms, sys, err := e.aiMessages(th, msgs)
	if err != nil {
		return nil, fmt.Errorf("ai.run: %w", err)
	}
	if system != "" {
		sys = system + "\n\n" + sys
	}
	tl, err := aiToolList(toolsV)
	if err != nil {
		return nil, fmt.Errorf("ai.run: %w", err)
	}
	byName := map[string]*aiTool{}
	req := ai.Request{Model: model, System: strings.TrimSpace(sys), MaxTokens: maxTokens, Temperature: temp}
	for _, t := range tl {
		byName[t.def.Name] = t
		req.Tools = append(req.Tools, t.def)
	}
	th.SetLocal(aiRunKey, true)
	defer th.SetLocal(aiRunKey, nil)

	var trace []starlark.Value
	var total ai.Usage
	var last *ai.Response
	started := time.Now()
	stop := "max_steps"
	steps := 0
	for steps < maxSteps {
		if time.Since(started) > aiRunTimeout {
			return nil, fmt.Errorf("ai.run: stopped after %s (limit %s)", time.Since(started).Round(time.Second), aiRunTimeout)
		}
		req.Messages = ms
		res, err := e.aiCall(th, p, req)
		if err != nil {
			return nil, err
		}
		steps++
		last = res
		total.Input += res.Usage.Input
		total.Output += res.Usage.Output
		if len(res.ToolCalls) == 0 || len(byName) == 0 {
			stop = res.StopReason
			break
		}
		ms = append(ms, ai.Message{Role: "assistant", Content: res.Text, ToolCalls: res.ToolCalls})
		for _, c := range res.ToolCalls {
			out, errMsg, err := e.aiRunTool(th, byName[c.Name], c)
			if err != nil {
				return nil, err
			}
			ms = append(ms, ai.Message{Role: "tool", ToolCallID: c.ID, Content: out, IsError: errMsg != ""})
			td := starlark.NewDict(4)
			_ = td.SetKey(starlark.String("name"), starlark.String(c.Name))
			_ = td.SetKey(starlark.String("args"), goToStarlark(c.Args))
			_ = td.SetKey(starlark.String("result"), starlark.String(out))
			if errMsg != "" {
				_ = td.SetKey(starlark.String("error"), starlark.String(errMsg))
			} else {
				_ = td.SetKey(starlark.String("error"), starlark.None)
			}
			trace = append(trace, td)
		}
	}
	last.Usage = total
	return aiResult(last, p, map[string]starlark.Value{
		"stop_reason": starlark.String(stop),
		"steps":       starlark.MakeInt(steps),
		"tool_calls":  starlark.NewList(trace),
	})
}

// aiRunTool calls one tool the model asked for. A failing tool is reported to
// the model; only cancellation (timeout, step limit) aborts the run. The tool
// runs on the caller's thread, so it has the extension's permissions and no
// others.
func (e *Extension) aiRunTool(th *starlark.Thread, t *aiTool, c ai.ToolCall) (out, errMsg string, abort error) {
	if t == nil {
		m := fmt.Sprintf("there is no tool named %q", c.Name)
		return m, m, nil
	}
	raw, _ := json.Marshal(c.Args)
	dv, err := jsonCall(th, "decode", starlark.String(raw))
	d, ok := dv.(*starlark.Dict)
	if err != nil || !ok {
		m := "tool arguments must be a JSON object"
		return m, m, nil
	}
	var kw []starlark.Tuple
	for _, kv := range d.Items() {
		kw = append(kw, starlark.Tuple{kv[0], kv[1]})
	}
	v, err := starlark.Call(th, t.fn, nil, kw)
	if err != nil {
		msg := err.Error()
		var ee *starlark.EvalError
		if errors.As(err, &ee) {
			msg = ee.Msg
		}
		if callCtx(th).Err() != nil || strings.Contains(msg, "Starlark computation cancelled") {
			return "", "", err
		}
		msg = clipText(msg, 2000)
		return "error: " + msg, msg, nil
	}
	return clipText(aiToolText(th, v), aiMaxToolResult), "", nil
}

func aiToolText(th *starlark.Thread, v starlark.Value) string {
	switch x := v.(type) {
	case starlark.String:
		return string(x)
	case starlark.NoneType:
		return "ok"
	}
	if enc, err := jsonCall(th, "encode", v); err == nil {
		return string(enc.(starlark.String))
	}
	return v.String()
}

func clipText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n[truncated]"
}

// ---- ai.embed ----

func (e *Extension) biAIEmbed(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var textsV starlark.Value
	var model, provider string
	if err := starlark.UnpackArgs("ai.embed", args, kw, "texts", &textsV, "model?", &model, "provider?", &provider); err != nil {
		return nil, err
	}
	var texts []string
	if s, ok := starlark.AsString(textsV); ok {
		texts = []string{s}
	} else {
		l, err := stringList(textsV)
		if err != nil {
			return nil, fmt.Errorf("ai.embed: %w", err)
		}
		texts = l
	}
	size := 0
	for _, t := range texts {
		size += len(t)
	}
	if len(texts) == 0 || len(texts) > aiMaxEmbedTexts || size > int(e.opts.Limits.MaxFileBytes) {
		return nil, fmt.Errorf("ai.embed: pass 1 to %d texts of at most %d bytes in total", aiMaxEmbedTexts, e.opts.Limits.MaxFileBytes)
	}
	p, err := e.aiProvider(provider)
	if err != nil {
		return nil, err
	}
	left, err := e.aiRemaining()
	if err != nil {
		return nil, err
	}
	if est := int64(size / 4); est >= left {
		return nil, fmt.Errorf("ai: these texts are about %d tokens and only %d are left in today's budget (ai_budget_tokens)", est, left)
	}
	resume, err := clockOf(th).pause(aiCallTotal)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(callCtx(th), aiRequestTimeout)
	vecs, usage, err := ai.Embed(ctx, e.opts.HTTPClient, p, model, texts, e.opts.Limits.MaxHTTPBytes)
	cancel()
	resume()
	if err != nil {
		return nil, fmt.Errorf("ai.embed: %w", err)
	}
	if usage.Total() == 0 {
		usage.Input = size/4 + 1
	}
	e.aiCharge(usage.Total())
	out := make([]starlark.Value, len(vecs))
	for i, v := range vecs {
		l := make([]starlark.Value, len(v))
		for j, f := range v {
			l[j] = starlark.Float(f)
		}
		out[i] = starlark.NewList(l)
	}
	return starlark.NewList(out), nil
}
