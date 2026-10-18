package ai

import (
	"context"
	"encoding/json"
	"net/http"
)

func completeOpenAI(ctx context.Context, hc *http.Client, p Provider, r Request, maxBody int64) (*Response, error) {
	body := map[string]any{"model": r.Model, "max_tokens": r.MaxTokens}
	if r.Temperature != nil {
		body["temperature"] = *r.Temperature
	}
	var msgs []map[string]any
	if r.System != "" {
		msgs = append(msgs, map[string]any{"role": "system", "content": r.System})
	}
	for _, m := range r.Messages {
		switch m.Role {
		case "tool":
			msgs = append(msgs, map[string]any{"role": "tool", "tool_call_id": m.ToolCallID, "content": m.Content})
		case "assistant":
			msg := map[string]any{"role": "assistant", "content": m.Content}
			if len(m.ToolCalls) > 0 {
				var tcs []any
				for _, c := range m.ToolCalls {
					args := []byte("{}")
					if c.Args != nil {
						args, _ = json.Marshal(c.Args)
					}
					tcs = append(tcs, map[string]any{"id": c.ID, "type": "function", "function": map[string]any{"name": c.Name, "arguments": string(args)}})
				}
				msg["tool_calls"] = tcs
			}
			msgs = append(msgs, msg)
		default:
			msgs = append(msgs, map[string]any{"role": "user", "content": m.Content})
		}
	}
	body["messages"] = msgs
	if len(r.Tools) > 0 {
		var ts []any
		for _, t := range r.Tools {
			ts = append(ts, map[string]any{"type": "function", "function": map[string]any{"name": t.Name, "description": t.Description, "parameters": t.Schema}})
		}
		body["tools"] = ts
		switch r.ToolChoice {
		case "":
		case "auto", "none", "required":
			body["tool_choice"] = r.ToolChoice
		default:
			body["tool_choice"] = map[string]any{"type": "function", "function": map[string]any{"name": r.ToolChoice}}
		}
	}
	var out struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content   *string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			In  int `json:"prompt_tokens"`
			Out int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := post(ctx, hc, p, "chat/completions", body, &out, maxBody); err != nil {
		return nil, err
	}
	res := &Response{Model: out.Model, Usage: Usage{Input: out.Usage.In, Output: out.Usage.Out}}
	if len(out.Choices) == 0 {
		return res, nil
	}
	c := out.Choices[0]
	if c.Message.Content != nil {
		res.Text = *c.Message.Content
	}
	for _, tc := range c.Message.ToolCalls {
		res.ToolCalls = append(res.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Args: decodeArgs(json.RawMessage(tc.Function.Arguments))})
	}
	switch c.FinishReason {
	case "stop":
		res.StopReason = "end"
	case "tool_calls", "function_call":
		res.StopReason = "tool_use"
	case "length":
		res.StopReason = "max_tokens"
	default:
		res.StopReason = c.FinishReason
	}
	if len(res.ToolCalls) > 0 {
		res.StopReason = "tool_use"
	}
	return res, nil
}
