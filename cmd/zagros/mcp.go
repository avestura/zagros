//go:build windows

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/mcp"
	"github.com/avestura/zagros/internal/ui"
	"github.com/avestura/zagros/stubs"
)

// newMCPServer builds the server. install adds the install_extension tool; it
// is only offered where the caller is trusted (stdio, or HTTP with a token).
func newMCPServer(install bool) (*mcp.Server, error) {
	idx, err := mcp.BuildIndex(stubs.FS())
	if err != nil {
		return nil, err
	}
	o := mcp.Options{Index: idx, Guide: string(stubs.AgentGuide()), Version: version}
	if install {
		o.Install = mcpInstall
	}
	return mcp.New(o), nil
}

// mcpInstall installs without a prompter, so nothing is granted here: the
// running launcher is asked to reload and shows the permission dialog.
func mcpInstall(ctx context.Context, path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	m, err := newCLIManager(false)
	if err != nil {
		return "", err
	}
	if _, err := m.Discover(ctx); err != nil {
		return "", err
	}
	e, err := m.InstallPath(ctx, path)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Installed %s %s.", e.ID(), e.Manifest.Version)
	if ui.SignalReload() {
		return msg + " Asked the running Zagros to reload; the user approves any permissions in its dialog.", nil
	}
	return msg + " Zagros is not running; the user is asked to approve any permissions the next time it starts.", nil
}

func exePath() string {
	p, err := os.Executable()
	if err != nil {
		return "zagros"
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// agentPrompt builds the text to give a coding agent. includeHTTP adds the
// HTTP endpoint (and token) of the configured server.
func agentPrompt(cfg core.Config, includeHTTP bool) string {
	o := mcp.PromptOptions{Exe: exePath()}
	if includeHTTP {
		o.URL = "http://" + mcpAddr(cfg) + "/mcp"
		o.Token = cfg.MCP.Token
	}
	return mcp.AgentPrompt(o)
}

func mcpAddr(cfg core.Config) string {
	if cfg.MCP.Addr == "" {
		return "127.0.0.1:7777"
	}
	return cfg.MCP.Addr
}

const mcpUsage = `usage:
  zagros mcp                         serve the MCP API search over stdio (an agent starts this itself)
  zagros mcp --http [--addr a] [--token t]   serve over HTTP in the foreground (loopback only)
  zagros mcp prompt [--http]         print the prompt to give your coding agent
  zagros mcp config                  print ready-to-paste client configuration
`

// cmdMCP implements "zagros mcp". On stdio, stdout carries only protocol
// messages, so nothing else may print there.
func cmdMCP(args []string) int {
	cfg, _ := core.LoadConfig(core.ConfigPath())
	if len(args) > 0 && args[0] == "prompt" {
		fs := flag.NewFlagSet("mcp prompt", flag.ContinueOnError)
		http := fs.Bool("http", false, "include the HTTP endpoint")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		fmt.Println(agentPrompt(cfg, *http || cfg.MCP.HTTP))
		return 0
	}
	if len(args) > 0 && args[0] == "config" {
		exe := exePath()
		fmt.Printf("Claude Code (stdio):\n  claude mcp add zagros -- %q mcp\n\n", exe)
		fmt.Printf("Claude Code (HTTP, needs \"zagros mcp --http\" or [mcp] http = true):\n  claude mcp add --transport http zagros http://%s/mcp\n\n", mcpAddr(cfg))
		fmt.Printf("JSON for other MCP clients (stdio):\n  {\"mcpServers\": {\"zagros\": {\"command\": %q, \"args\": [\"mcp\"]}}}\n", exe)
		return 0
	}
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	http := fs.Bool("http", false, "serve over HTTP instead of stdio")
	addr := fs.String("addr", mcpAddr(cfg), "loopback address for --http")
	token := fs.String("token", cfg.MCP.Token, "bearer token for --http")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprint(os.Stderr, mcpUsage)
		return 2
	}
	srv, err := newMCPServer(!*http || *token != "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcp:", err)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *http {
		fmt.Fprintf(os.Stderr, "zagros mcp: serving http://%s/mcp (Ctrl+C to stop)\n", *addr)
		if err := srv.ListenAndServe(ctx, *addr, *token); err != nil {
			fmt.Fprintln(os.Stderr, "mcp:", err)
			return 1
		}
		return 0
	}
	if err := srv.ServeStdio(ctx, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
		fmt.Fprintln(os.Stderr, "mcp:", err)
		return 1
	}
	return 0
}
