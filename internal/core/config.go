package core

// Config is the user configuration, stored as TOML at <ConfigDir>/zagros.toml.
// Loading/saving/defaults live in config_load.go (owned by the core workstream).
type Config struct {
	Hotkey           string            `toml:"hotkey"`      // e.g. "Alt+Space"
	Theme            string            `toml:"theme"`       // theme name; "" = follow Windows light/dark
	MaxResults       int               `toml:"max_results"` // rows shown, default 8
	Providers        map[string]bool   `toml:"providers"`   // provider name -> enabled (missing = true)
	ScanDirs         []string          `toml:"scan_dirs"`   // extra folders scanned by the apps provider
	Aliases          map[string]string `toml:"aliases"`     // keyword -> target (path/URL/command)
	WebSearch        []WebSearch       `toml:"web_search"`
	Everything       EverythingConfig  `toml:"everything"`
	Marketplace      MarketplaceConfig `toml:"marketplace"`
	Updates          UpdatesConfig     `toml:"updates"`
	MCP              MCPConfig         `toml:"mcp"`
	AI               AIConfig          `toml:"ai"`
	StartWithWindows bool              `toml:"start_with_windows"`
}

type WebSearch struct {
	Keyword string `toml:"keyword"` // "g"
	Name    string `toml:"name"`    // "Google"
	URL     string `toml:"url"`     // "https://www.google.com/search?q={q}"
}

type EverythingConfig struct {
	Enabled bool   `toml:"enabled"`
	DLLPath string `toml:"dll_path"` // optional explicit Everything64.dll
	Limit   int    `toml:"limit"`    // default 20
}

// UpdatesConfig controls the GitHub release check. Zagros never checks in the
// background unless CheckOnStart is set; the "update" keyword and the tray item
// check on demand.
type UpdatesConfig struct {
	CheckOnStart bool `toml:"check_on_start"` // check for a newer release when Zagros starts (default false)
}

// MCPConfig controls the Model Context Protocol server that lets coding agents
// search the extension API. "zagros mcp" (stdio) needs no configuration; the
// HTTP server inside the running launcher is off unless enabled here or from
// the tray menu. It only serves the public API docs, binds to loopback and
// can require a bearer token.
type MCPConfig struct {
	HTTP  bool   `toml:"http"`  // run the HTTP server while Zagros is running (default false)
	Addr  string `toml:"addr"`  // loopback address, default 127.0.0.1:7777
	Token string `toml:"token"` // optional bearer token clients must send
}

// AIConfig lists the language-model providers extensions with the "ai"
// permission may use. API keys are not stored here: "zagros ai key set <name>"
// puts them in Windows Credential Manager.
type AIConfig struct {
	Default   string       `toml:"default"` // provider used when an extension names none (default: the first)
	Providers []AIProvider `toml:"provider"`
}

// AIProvider is one [[ai.provider]] entry.
type AIProvider struct {
	Name       string `toml:"name"`        // what extensions and "ai:<name>" permissions refer to
	Kind       string `toml:"kind"`        // "anthropic" or "openai" (any OpenAI-compatible server)
	BaseURL    string `toml:"base_url"`    // e.g. http://localhost:11434/v1 for Ollama
	Model      string `toml:"model"`       // default model
	EmbedModel string `toml:"embed_model"` // default embedding model (openai kind)
}

type MarketplaceConfig struct {
	RegistryURL string `toml:"registry_url"` // default https://avestura.github.io/zagros/registry.json
}
