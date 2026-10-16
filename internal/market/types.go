// Package market implements the Zagros extension marketplace: a Git-based
// registry index plus direct GitHub references. There is no backend server.
package market

import "context"

// SchemaVersion is the registry.json schema version this package reads.
const SchemaVersion = 1

// DefaultRegistryURL is where the generated registry is published.
const DefaultRegistryURL = "https://avestura.github.io/zagros/registry.json"

// Entry is one extension in registry.json.
type Entry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Repo        string   `json:"repo"` // "owner/name"
	Tags        []string `json:"tags"`
	Latest      string   `json:"latest"` // semver without leading "v"
	Permissions []string `json:"permissions"`
	Verified    bool     `json:"verified"`
	// Kind is "extension" (default, omitted) or "theme".
	Kind      string `json:"kind,omitempty"`
	Downloads int    `json:"downloads,omitempty"`
	// Optional install hints copied from the registry TOML.
	Ref    string `json:"ref,omitempty"`    // tag/branch/commit to install instead of the latest release
	Subdir string `json:"subdir,omitempty"` // manifest.toml location inside the repo
	Asset  string `json:"asset,omitempty"`  // glob selecting the release asset
	SHA256 string `json:"sha256,omitempty"` // optional pin of the downloaded bytes
}

// Registry is the parsed registry.json.
type Registry struct {
	Schema     int     `json:"schema"`
	Generated  string  `json:"generated"`
	Extensions []Entry `json:"extensions"`
	// Stale is true when the data came from the offline cache because the
	// network fetch failed.
	Stale bool `json:"-"`
}

// Manifest holds the subset of an extension's manifest.toml the marketplace
// needs. The full manifest is parsed and validated by internal/ext.
type Manifest struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	Permissions []string
}

// Package is a downloaded, ready-to-install extension archive.
type Package struct {
	Kind     string // KindExtension or KindTheme
	Theme    []byte // KindTheme: the validated .toml file (Zip is nil)
	ID       string // manifest id (falls back to the registry id)
	Version  string // manifest version (falls back to the release tag)
	Source   string // canonical ref, e.g. "github:owner/repo@v1.2.0" or the URL
	Zip      []byte // .zagros archive with manifest.toml at its root
	SHA256   string // hex SHA-256 of the bytes as downloaded
	URL      string // URL the bytes came from
	Manifest Manifest
	Entry    *Entry // set when resolved through the registry
}

// Installer is satisfied (through an adapter) by ext.Manager.
type Installer interface {
	// Install installs a .zagros archive and returns the extension id.
	// source is the canonical ref recorded for later updates.
	Install(ctx context.Context, zipBytes []byte, source string) (id string, err error)
	// Installed maps extension id to installed version.
	Installed() map[string]string
}

// Update describes an available upgrade.
type Update struct {
	ID        string
	Installed string
	Latest    string
	Entry     Entry
}

// Entry/Package kinds.
const (
	KindExtension = "extension"
	KindTheme     = "theme"
)
