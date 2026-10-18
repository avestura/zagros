// Package ai talks to language-model providers: the Anthropic Messages API and
// OpenAI-compatible chat endpoints (OpenAI, Ollama, LM Studio, vLLM, ...). It
// knows nothing about extensions, permissions or budgets; internal/ext adds
// those around it.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// Provider kinds.
const (
	KindAnthropic = "anthropic"
	KindOpenAI    = "openai"
)

// Provider is one configured endpoint with its resolved API key.
type Provider struct {
	Name       string
	Kind       string
	BaseURL    string
	Model      string // default model
	EmbedModel string // default embedding model (openai kind only)
	APIKey     string // empty for local servers
}

// Message is one turn. Role is user, assistant or tool.
type Message struct {
	Role    string
	Content string
	// ToolCalls are the calls an assistant turn made.
	ToolCalls []ToolCall
	// ToolCallID and IsError describe a tool turn: the result of one call.
	ToolCallID string
	IsError    bool
}

// ToolCall is a function the model asked to run.
type ToolCall struct {
	ID   string
	Name string
	Args map[string]any
}

// Tool describes a function the model may call. Schema is a JSON schema of
// type object.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]any
}

// Request is a provider-neutral completion request.
type Request struct {
	Model       string
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature *float64
	Tools       []Tool
	// ToolChoice is "", "auto", "none", "required" or a tool name.
	ToolChoice string
}

// Usage counts tokens.
type Usage struct {
	Input  int
	Output int
}

func (u Usage) Total() int { return u.Input + u.Output }

// Response is a provider-neutral completion.
type Response struct {
	Text       string
	ToolCalls  []ToolCall
	Usage      Usage
	StopReason string // end, tool_use, max_tokens or the provider's own word
	Model      string
}

// ErrNoEmbeddings is returned by Embed for providers without an embeddings API.
var ErrNoEmbeddings = errors.New("this provider has no embeddings API (use an openai-kind provider)")

// Validate checks the provider's static settings.
func (p Provider) Validate() error {
	if p.Name == "" {
		return errors.New("provider has no name")
	}
	if p.Kind != KindAnthropic && p.Kind != KindOpenAI {
		return fmt.Errorf("provider %q: kind must be %q or %q, got %q", p.Name, KindAnthropic, KindOpenAI, p.Kind)
	}
	_, err := p.endpoint("")
	return err
}

// endpoint returns the URL for path (e.g. "messages"). The key travels in a
// header, so plain http is only accepted for loopback and private hosts.
func (p Provider) endpoint(path string) (string, error) {
	base := p.BaseURL
	if base == "" {
		switch p.Kind {
		case KindAnthropic:
			base = "https://api.anthropic.com"
		case KindOpenAI:
			base = "https://api.openai.com"
		}
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("provider %q: base_url %q must be an http:// or https:// URL", p.Name, p.BaseURL)
	}
	if u.Scheme == "http" && !localHost(u.Hostname()) {
		return "", fmt.Errorf("provider %q: base_url must use https unless the host is local", p.Name)
	}
	if u.User != nil {
		return "", fmt.Errorf("provider %q: base_url must not contain credentials", p.Name)
	}
	u.RawQuery, u.Fragment = "", ""
	root := strings.TrimRight(u.String(), "/")
	if path == "" {
		return root, nil
	}
	if !strings.HasSuffix(root, "/v1") {
		root += "/v1"
	}
	return root + "/" + path, nil
}

func localHost(h string) bool {
	if strings.EqualFold(h, "localhost") || strings.HasSuffix(strings.ToLower(h), ".localhost") {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate())
}

// KeyOptional reports whether the provider may be used without an API key:
// it is a local server.
func (p Provider) KeyOptional() bool {
	u, err := url.Parse(p.BaseURL)
	return err == nil && p.BaseURL != "" && localHost(u.Hostname())
}

// Redact removes the provider's API key from s.
func (p Provider) Redact(s string) string {
	if len(p.APIKey) >= 4 {
		s = strings.ReplaceAll(s, p.APIKey, "***")
	}
	return s
}

// Complete sends req to p. maxBody caps the response body. Redirects are
// never followed, so the key cannot be sent to another host.
func Complete(ctx context.Context, hc *http.Client, p Provider, req Request, maxBody int64) (*Response, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if req.Model == "" {
		req.Model = p.Model
	}
	if req.Model == "" {
		return nil, fmt.Errorf("provider %q has no default model: pass model= or set model in the config", p.Name)
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = 1024
	}
	if p.Kind == KindAnthropic {
		return completeAnthropic(ctx, hc, p, req, maxBody)
	}
	return completeOpenAI(ctx, hc, p, req, maxBody)
}

// Embed returns one vector per text.
func Embed(ctx context.Context, hc *http.Client, p Provider, model string, texts []string, maxBody int64) ([][]float64, Usage, error) {
	if err := p.Validate(); err != nil {
		return nil, Usage{}, err
	}
	if p.Kind != KindOpenAI {
		return nil, Usage{}, ErrNoEmbeddings
	}
	if model == "" {
		model = p.EmbedModel
	}
	if model == "" {
		return nil, Usage{}, fmt.Errorf("provider %q has no embed_model: pass model= or set embed_model in the config", p.Name)
	}
	var out struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			Prompt int `json:"prompt_tokens"`
			Total  int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := post(ctx, hc, p, "embeddings", map[string]any{"model": model, "input": texts}, &out, maxBody); err != nil {
		return nil, Usage{}, err
	}
	vecs := make([][]float64, len(texts))
	for _, d := range out.Data {
		if d.Index >= 0 && d.Index < len(vecs) {
			vecs[d.Index] = d.Embedding
		}
	}
	for i, v := range vecs {
		if v == nil {
			return nil, Usage{}, fmt.Errorf("provider %q returned no embedding for text %d", p.Name, i)
		}
	}
	return vecs, Usage{Input: max(out.Usage.Prompt, out.Usage.Total)}, nil
}

// post sends a JSON body and decodes a JSON reply. Errors never contain the key.
func post(ctx context.Context, hc *http.Client, p Provider, path string, body any, out any, maxBody int64) error {
	u, err := p.endpoint(path)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return errors.New(p.Redact(err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Zagros")
	if p.Kind == KindAnthropic {
		req.Header.Set("anthropic-version", "2023-06-01")
		if p.APIKey != "" {
			req.Header.Set("x-api-key", p.APIKey)
		}
	} else if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	c := *hc
	c.Timeout = 0 // the context carries the deadline
	c.Jar = nil
	c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := c.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("provider %q: %w", p.Name, ctx.Err())
		}
		return fmt.Errorf("provider %q: %s", p.Name, p.Redact(err.Error()))
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return fmt.Errorf("provider %q: reading reply: %s", p.Name, p.Redact(err.Error()))
	}
	if int64(len(b)) > maxBody {
		return fmt.Errorf("provider %q: reply larger than %d bytes", p.Name, maxBody)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("provider %q: HTTP %d: %s", p.Name, resp.StatusCode, p.Redact(errorText(b)))
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("provider %q: reply is not the expected JSON: %v", p.Name, err)
	}
	return nil
}

// errorText pulls the message out of a provider error body, else a short
// excerpt of it.
func errorText(b []byte) string {
	var e struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(b, &e) == nil && len(e.Error) > 0 {
		var o struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(e.Error, &o) == nil && o.Message != "" {
			return clip(o.Message)
		}
		var s string
		if json.Unmarshal(e.Error, &s) == nil && s != "" {
			return clip(s)
		}
	}
	return clip(strings.TrimSpace(string(b)))
}

func clip(s string) string {
	const n = 300
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
