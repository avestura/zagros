package ai

import (
	"context"
	"encoding/json"
	"net/http"
)

func completeAnthropic(ctx context.Context, hc *http.Client, p Provider, r Request, maxBody int64) (*Response, error) {
	body := map[string]any{"model": r.Model, "max_tokens": r.MaxTokens}
	if r.System != "" {
		body["system"] = r.System
	}
	if r.Temperature != nil {
		body["temperature"] = *r.Temperature
	}
	var msgs []map[string]any
	for _, m := range r.Messages {
		switch m.Role {
		case "tool":
			blk := map[string]any{"type": "tool_result", "tool_use_id": m.ToolCallID, "content": m.Content}
			if m.IsError {
				blk["is_error"] = true
			}
			// Consecutive results share one user turn.
			if n := len(msgs); n > 0 && msgs[n-1]["role"] == "user" {
				if cs, ok := msgs[n-1]["content"].([]any); ok && len(cs) > 0 {
					if first, ok := cs[0].(map[string]any); ok && first["type"] == "tool_result" {
						msgs[n-1]["content"] = append(cs, blk)
						continue
					}
				}
			}
			msgs = append(msgs, map[string]any{"role": "user", "content": []any{blk}})
		case "assistant":
			var cs []any
			if m.Content != "" {
				cs = append(cs, map[string]any{"type": "text", "text": m.Content})
			}
			for _, c := range m.ToolCalls {
				in := c.Args
				if in == nil {
					in = map[string]any{}
				}
				cs = append(cs, map[string]any{"type": "tool_use", "id": c.ID, "name": c.Name, "input": in})
			}
			if len(cs) == 0 {
				cs = []any{map[string]any{"type": "text", "text": " "}}
			}
			msgs = append(msgs, map[string]any{"role": "assistant", "content": cs})
		default:
			msgs = append(msgs, map[string]any{"role": "user", "content": m.Content})
		}
	}
	body["messages"] = msgs
	if len(r.Tools) > 0 {
		var ts []any
		for _, t := range r.Tools {
			ts = append(ts, map[string]any{"name": t.Name, "description": t.Description, "input_schema": t.Schema})
		}
		body["tools"] = ts
		switch r.ToolChoice {
		case "", "auto":
		case "none":
			body["tool_choice"] = map[string]any{"type": "none"}
		case "required", "any":
			body["tool_choice"] = map[string]any{"type": "any"}
		default:
			body["tool_choice"] = map[string]any{"type": "tool", "name": r.ToolChoice}
		}
	}
	var out struct {
		Model   string `json:"model"`
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			In  int `json:"input_tokens"`
			Out int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := post(ctx, hc, p, "messages", body, &out, maxBody); err != nil {
		return nil, err
	}
	res := &Response{Model: out.Model, Usage: Usage{Input: out.Usage.In, Output: out.Usage.Out}}
	for _, c := range out.Content {
		switch c.Type {
		case "text":
			res.Text += c.Text
		case "tool_use":
			res.ToolCalls = append(res.ToolCalls, ToolCall{ID: c.ID, Name: c.Name, Args: decodeArgs(c.Input)})
		}
	}
	res.StopReason = out.StopReason
	if out.StopReason == "end_turn" || out.StopReason == "stop_sequence" {
		res.StopReason = "end"
	}
	return res, nil
}

func decodeArgs(raw json.RawMessage) map[string]any {
	m := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &m)
	}
	if m == nil {
		m = map[string]any{}
	}
	return m
}
