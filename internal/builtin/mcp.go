package builtin

import (
	"context"
	"fmt"

	"github.com/avestura/zagros/internal/core"
)

// MCPHooks connects the "mcp" keyword to the MCP server owned by the
// application (cmd/zagros). Unset hooks hide the matching rows.
var MCPHooks struct {
	Prompt  func(http bool) string // the text to give a coding agent
	Command func() string          // `claude mcp add ...` for the stdio server
	HTTPURL func() string          // "" when the HTTP server is off
	Toggle  func() string          // turns the HTTP server on/off for this session, returns a status line
}

type mcpProvider struct{ base }

func newMCP() *mcpProvider { return &mcpProvider{base{"mcp"}} }

func (p *mcpProvider) Ranked() bool       { return true }
func (p *mcpProvider) Keywords() []string { return []string{"mcp", "agent"} }

// Query shows one row; Tab opens the Actions scope with the other choices.
func (p *mcpProvider) Query(ctx context.Context, q string) []core.Item {
	if _, _, ok := trigger(q, p.Keywords()...); !ok || MCPHooks.Prompt == nil {
		return nil
	}
	cp := func(text string) core.Action { return noHistoryAction(copyAction(text)) }
	it := core.Item{ID: "mcp:prompt", Title: "Copy the coding-agent prompt", Category: "MCP",
		Subtitle: "Paste it into Claude Code or another agent so it can search the Zagros extension API - Tab for more",
		Action:   cp(MCPHooks.Prompt(false))}
	if MCPHooks.Command != nil {
		it.Alt = append(it.Alt, core.NamedAction{Name: "Copy the command that adds the stdio server", Action: cp(MCPHooks.Command())})
	}
	if url := httpURL(); url != "" {
		it.Alt = append(it.Alt,
			core.NamedAction{Name: "Copy the prompt for the HTTP server", Action: cp(MCPHooks.Prompt(true))},
			core.NamedAction{Name: "Copy the HTTP endpoint " + url, Action: cp(url)})
	}
	if MCPHooks.Toggle != nil {
		label := "Turn the HTTP MCP server on (this session)"
		if httpURL() != "" {
			label = "Turn the HTTP MCP server off"
		}
		it.Alt = append(it.Alt, core.NamedAction{Name: label, Action: core.Action{Kind: "func", Func: func() error { MCPHooks.Toggle(); return nil }}})
	}
	it.Alt = append(it.Alt, core.NamedAction{Name: "Open the documentation", Action: core.Action{Kind: "open", Target: "https://github.avestura.dev/zagros/docs/extensions-agents.html"}})
	state := "HTTP server off"
	if u := httpURL(); u != "" {
		state = fmt.Sprintf("HTTP server on at %s", u)
	}
	it.Subtitle += " (" + state + ")"
	return []core.Item{it}
}

func httpURL() string {
	if MCPHooks.HTTPURL == nil {
		return ""
	}
	return MCPHooks.HTTPURL()
}
