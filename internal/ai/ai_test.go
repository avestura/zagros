package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpointRules(t *testing.T) {
	cases := []struct {
		p    Provider
		path string
		want string
		ok   bool
	}{
		{Provider{Name: "a", Kind: KindAnthropic}, "messages", "https://api.anthropic.com/v1/messages", true},
		{Provider{Name: "o", Kind: KindOpenAI}, "chat/completions", "https://api.openai.com/v1/chat/completions", true},
		{Provider{Name: "l", Kind: KindOpenAI, BaseURL: "http://localhost:11434/v1"}, "chat/completions", "http://localhost:11434/v1/chat/completions", true},
		{Provider{Name: "l", Kind: KindOpenAI, BaseURL: "http://192.168.1.5:8080/"}, "embeddings", "http://192.168.1.5:8080/v1/embeddings", true},
		{Provider{Name: "r", Kind: KindOpenAI, BaseURL: "http://example.com"}, "x", "", false},
		{Provider{Name: "r", Kind: KindOpenAI, BaseURL: "ftp://example.com"}, "x", "", false},
		{Provider{Name: "r", Kind: KindOpenAI, BaseURL: "https://user:pw@example.com"}, "x", "", false},
		{Provider{Name: "r", Kind: "gemini"}, "x", "", false},
	}
	for _, c := range cases {
		got, err := c.p.endpoint(c.path)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("%+v: got %q, %v", c.p, got, err)
		}
	}
	if !(Provider{BaseURL: "http://127.0.0.1:1234"}).KeyOptional() || (Provider{BaseURL: "https://api.openai.com"}).KeyOptional() || (Provider{}).KeyOptional() {
		t.Error("KeyOptional is wrong")
	}
}

func TestRedact(t *testing.T) {
	p := Provider{APIKey: "sk-abcdef"}
	if got := p.Redact("bad key sk-abcdef!"); got != "bad key ***!" {
		t.Errorf("got %q", got)
	}
}

func TestAnthropicToolRoundTripShape(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		io.WriteString(w, `{"model":"m","content":[{"type":"text","text":"a"},{"type":"text","text":"b"},{"type":"tool_use","id":"x","name":"f","input":{"k":1}}],"stop_reason":"tool_use","usage":{"input_tokens":3,"output_tokens":4}}`)
	}))
	defer srv.Close()
	p := Provider{Name: "a", Kind: KindAnthropic, BaseURL: srv.URL, Model: "m", APIKey: "k"}
	res, err := Complete(context.Background(), srv.Client(), p, Request{
		Messages: []Message{
			{Role: "user", Content: "go"},
			{Role: "assistant", ToolCalls: []ToolCall{{ID: "1", Name: "f"}, {ID: "2", Name: "g"}}},
			{Role: "tool", ToolCallID: "1", Content: "r1"},
			{Role: "tool", ToolCallID: "2", Content: "r2", IsError: true},
		},
		Tools:      []Tool{{Name: "f", Description: "d", Schema: map[string]any{"type": "object"}}},
		ToolChoice: "f",
	}, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "ab" || len(res.ToolCalls) != 1 || res.ToolCalls[0].Args["k"] != float64(1) || res.StopReason != "tool_use" || res.Usage.Total() != 7 {
		t.Errorf("response: %+v", res)
	}
	msgs := body["messages"].([]any)
	if len(msgs) != 3 {
		t.Fatalf("the two tool results must share one user turn: %v", msgs)
	}
	if n := len(msgs[2].(map[string]any)["content"].([]any)); n != 2 {
		t.Errorf("tool_result blocks: %d", n)
	}
	if tc := body["tool_choice"].(map[string]any); tc["type"] != "tool" || tc["name"] != "f" {
		t.Errorf("tool_choice: %v", tc)
	}
}

func TestErrorBodiesAreClippedAndRedacted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		io.WriteString(w, "oops sk-zzzzzz "+strings.Repeat("x", 1000))
	}))
	defer srv.Close()
	p := Provider{Name: "o", Kind: KindOpenAI, BaseURL: srv.URL, Model: "m", APIKey: "sk-zzzzzz"}
	_, err := Complete(context.Background(), srv.Client(), p, Request{Messages: []Message{{Role: "user", Content: "x"}}}, 1<<20)
	if err == nil || strings.Contains(err.Error(), "sk-zzzzzz") || len(err.Error()) > 450 {
		t.Errorf("error: %v", err)
	}
}

func TestCompleteNeedsModel(t *testing.T) {
	p := Provider{Name: "o", Kind: KindOpenAI, BaseURL: "http://localhost:1"}
	if _, err := Complete(context.Background(), http.DefaultClient, p, Request{}, 10); err == nil || !strings.Contains(err.Error(), "default model") {
		t.Errorf("got %v", err)
	}
	if _, _, err := Embed(context.Background(), http.DefaultClient, Provider{Name: "a", Kind: KindAnthropic}, "m", []string{"x"}, 10); err != ErrNoEmbeddings {
		t.Errorf("got %v", err)
	}
}
