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

type MarketplaceConfig struct {
	RegistryURL string `toml:"registry_url"` // default https://avestura.github.io/zagros/registry.json
}
