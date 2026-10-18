package core

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultRegistryURL = "https://avestura.github.io/zagros/registry.json"

// DefaultConfig returns the configuration used when no file exists.
func DefaultConfig() Config {
	return Config{
		Hotkey:     "Alt+Space",
		Theme:      "",
		MaxResults: 8,
		Providers:  map[string]bool{},
		Aliases:    map[string]string{},
		WebSearch: []WebSearch{
			{Keyword: "g", Name: "Google", URL: "https://www.google.com/search?q={q}"},
			{Keyword: "ddg", Name: "DuckDuckGo", URL: "https://duckduckgo.com/?q={q}"},
			{Keyword: "gh", Name: "GitHub", URL: "https://github.com/search?q={q}"},
			{Keyword: "yt", Name: "YouTube", URL: "https://www.youtube.com/results?search_query={q}"},
			{Keyword: "so", Name: "Stack Overflow", URL: "https://stackoverflow.com/search?q={q}"},
		},
		Everything:  EverythingConfig{Enabled: true, Limit: 20},
		MCP:         MCPConfig{Addr: "127.0.0.1:7777"},
		Marketplace: MarketplaceConfig{RegistryURL: DefaultRegistryURL},
	}
}

// ProviderEnabled reports whether the named provider is enabled; providers
// absent from the map are enabled.
func (c Config) ProviderEnabled(name string) bool {
	on, ok := c.Providers[name]
	return !ok || on
}

// LoadConfig reads path on top of the defaults. If the file does not exist it
// is created with a commented default and the defaults are returned.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if werr := writeAtomic(path, []byte(defaultConfigTOML)); werr != nil {
			return cfg, fmt.Errorf("write default config: %w", werr)
		}
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	// A user-supplied web_search list replaces the default list rather than
	// being merged into it, so decode it separately.
	var user Config
	md, err := toml.Decode(string(data), &user)
	if err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.WebSearch = nil
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	if !md.IsDefined("web_search") {
		cfg.WebSearch = DefaultConfig().WebSearch
	}
	if cfg.Providers == nil {
		cfg.Providers = map[string]bool{}
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}
	if cfg.MaxResults <= 0 {
		cfg.MaxResults = 8
	}
	if cfg.Everything.Limit <= 0 {
		cfg.Everything.Limit = 20
	}
	if cfg.Marketplace.RegistryURL == "" {
		cfg.Marketplace.RegistryURL = DefaultRegistryURL
	}
	if cfg.Hotkey == "" {
		cfg.Hotkey = "Alt+Space"
	}
	return cfg, nil
}

// SaveConfig writes cfg to path atomically. Comments in an existing file are
// not preserved.
func SaveConfig(path string, cfg Config) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	return writeAtomic(path, buf.Bytes())
}

func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

const defaultConfigTOML = `# Zagros configuration. Changes are picked up on restart.

# Global hotkey that shows the launcher.
hotkey = "Alt+Space"

# Theme name ("boxy-light", "boxy-dark", "classic", "nord" or a file in the
# themes folder). Empty follows the Windows light/dark setting.
theme = ""

# Number of result rows shown.
max_results = 8

# Start Zagros when you sign in to Windows.
start_with_windows = false

# Extra folders scanned by the apps provider.
scan_dirs = []

# Providers are enabled unless listed here as false.
[providers]
# apps = true
# calc = true
# passwords = false

# Keyword aliases: type the keyword to open the target.
[aliases]
# docs = "https://pkg.go.dev"

[everything]
enabled = true
# dll_path = "C:\\Program Files\\Everything\\Everything64.dll"
limit = 20

[marketplace]
registry_url = "https://avestura.github.io/zagros/registry.json"

# Update check: type "update" in the launcher or use the tray menu any time.
# Set check_on_start to also look at github.com/avestura/zagros/releases at startup.
[updates]
check_on_start = false

# MCP server so coding agents (Claude Code, ...) can search the extension API.
# "zagros mcp" works without this (stdio). http = true also serves it at
# http://<addr>/mcp while Zagros runs. Loopback only; token is optional.
[mcp]
http = false
addr = "127.0.0.1:7777"
# token = ""

# Language-model providers for extensions that hold the "ai" permission.
# Keys are not kept here: run "zagros ai key set <name>" (Credential Manager).
# kind is "anthropic" or "openai" (OpenAI, Ollama, LM Studio and other
# OpenAI-compatible servers). Plain http is accepted for local hosts only.
# [ai]
# default = "claude"
#
# [[ai.provider]]
# name = "claude"
# kind = "anthropic"
# model = "claude-sonnet-5-5"
#
# [[ai.provider]]
# name = "local"
# kind = "openai"
# base_url = "http://localhost:11434/v1"
# model = "llama3.2"

# Web searches: type "<keyword> <terms>". {q} is replaced by the terms.
[[web_search]]
keyword = "g"
name = "Google"
url = "https://www.google.com/search?q={q}"

[[web_search]]
keyword = "ddg"
name = "DuckDuckGo"
url = "https://duckduckgo.com/?q={q}"

[[web_search]]
keyword = "gh"
name = "GitHub"
url = "https://github.com/search?q={q}"

[[web_search]]
keyword = "yt"
name = "YouTube"
url = "https://www.youtube.com/results?search_query={q}"

[[web_search]]
keyword = "so"
name = "Stack Overflow"
url = "https://stackoverflow.com/search?q={q}"
`

// SetThemeLine returns src with the top-level `theme = ...` line set to name.
// Everything else, comments included, is kept. When there is no such line it
// is added before the first [table] (or at the end when there are no tables),
// because a key after a table header would belong to that table.
func SetThemeLine(src, name string) string {
	line := "theme = " + strconv.Quote(name)
	nl := "\n"
	if strings.Contains(src, "\r\n") {
		nl = "\r\n"
	}
	lines := strings.SplitAfter(src, "\n")
	firstTable := -1
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "[") {
			firstTable = i
			break
		}
		rest, ok := strings.CutPrefix(t, "theme")
		if ok && strings.HasPrefix(strings.TrimSpace(rest), "=") {
			end := ""
			if strings.HasSuffix(l, "\r\n") {
				end = "\r\n"
			} else if strings.HasSuffix(l, "\n") {
				end = "\n"
			}
			lines[i] = line + end
			return strings.Join(lines, "")
		}
	}
	if firstTable >= 0 {
		out := append([]string{}, lines[:firstTable]...)
		out = append(out, line+nl+nl)
		return strings.Join(append(out, lines[firstTable:]...), "")
	}
	if src != "" && !strings.HasSuffix(src, "\n") {
		src += nl
	}
	return src + line + nl
}

// SetTheme sets `theme = "<name>"` in the config file at path with a minimal
// edit, creating the default file first if it does not exist.
func SetTheme(path, name string) error {
	if _, err := LoadConfig(path); err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return writeAtomic(path, []byte(SetThemeLine(string(b), name)))
}
