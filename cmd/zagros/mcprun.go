//go:build windows

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/avestura/zagros/internal/builtin"
)

// startMCP starts the HTTP MCP server inside the launcher. It returns an
// error when the address cannot be bound (the usual reason is a busy port).
func (s *runtimeState) startMCP() error {
	s.mu.Lock()
	if s.mcpCancel != nil {
		s.mu.Unlock()
		return nil
	}
	cfg := s.cfg
	s.mu.Unlock()
	srv, err := newMCPServer(cfg.MCP.Token != "")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe(ctx, mcpAddr(cfg), cfg.MCP.Token) }()
	select {
	case err := <-errCh: // failed to bind
		cancel()
		return err
	case <-time.After(250 * time.Millisecond):
	}
	s.mu.Lock()
	s.mcpCancel, s.mcpURL = cancel, "http://"+mcpAddr(cfg)+"/mcp"
	s.mu.Unlock()
	go func() { // the server stopped on its own
		<-errCh
		s.mu.Lock()
		if s.mcpCancel != nil {
			s.mcpCancel, s.mcpURL = nil, ""
		}
		s.mu.Unlock()
	}()
	return nil
}

func (s *runtimeState) stopMCP() {
	s.mu.Lock()
	cancel := s.mcpCancel
	s.mcpCancel, s.mcpURL = nil, ""
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *runtimeState) mcpOn() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.mcpCancel != nil }

func (s *runtimeState) mcpEndpoint() string { s.mu.Lock(); defer s.mu.Unlock(); return s.mcpURL }

// toggleMCP turns the HTTP server on or off for this session and returns a
// status line. The setting [mcp] http = true makes it start with Zagros.
func (s *runtimeState) toggleMCP() string {
	if s.mcpOn() {
		s.stopMCP()
		return "MCP server stopped"
	}
	if err := s.startMCP(); err != nil {
		return "MCP server could not start: " + err.Error()
	}
	return fmt.Sprintf("MCP server on at %s (set [mcp] http = true to keep it on)", s.mcpEndpoint())
}

// wireMCP connects the tray, the "mcp" keyword and the config switch.
func (s *runtimeState) wireMCP() {
	builtin.MCPHooks.Prompt = func(http bool) string { return agentPrompt(s.config(), http) }
	builtin.MCPHooks.Command = func() string { return fmt.Sprintf(`claude mcp add zagros -- "%s" mcp`, exePath()) }
	builtin.MCPHooks.HTTPURL = s.mcpEndpoint
	builtin.MCPHooks.Toggle = s.toggleMCP
	if s.config().MCP.HTTP {
		if err := s.startMCP(); err != nil {
			if a := s.getApp(); a != nil {
				a.Notify("Zagros", "MCP server could not start: "+err.Error())
			}
		}
	}
}
