package ui

import (
	"context"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/theme"
)

// Catalog is the query side of core.Catalog.
type Catalog interface {
	Query(ctx context.Context, q string) []core.Item
}

// ExtInfo describes one installed extension for the manager window.
type ExtInfo struct {
	ID          string
	Name        string
	Version     string
	Publisher   string
	Enabled     bool
	Permissions []string // granted permissions
}

// ExtAdmin is what the extension manager window needs from the ext package.
type ExtAdmin interface {
	List() []ExtInfo
	SetEnabled(id string, enabled bool) error
	Revoke(id, permission string) error
	Uninstall(id string) error
}

// MarketEntry is one installable extension in the marketplace.
type MarketEntry struct {
	Ref         string // passed to Install, e.g. "github:owner/repo"
	Name        string
	Description string
	Author      string
	Version     string
	Kind        string // "extension" (or empty) or "theme"
}

// ThemeInfo is one theme listed on the manager's Themes tab.
type ThemeInfo struct {
	Name    string
	Author  string
	Source  string // "built-in", "user" or "installed"
	Current bool
}

// ThemeAdmin lists themes and selects one. Use only has to persist the choice
// in the config; the manager calls Options.Reload afterwards.
type ThemeAdmin interface {
	List() []ThemeInfo
	Use(name string) error
}

// Market is the marketplace side used by the extension manager.
type Market interface {
	List(ctx context.Context) ([]MarketEntry, error)
	Install(ctx context.Context, e MarketEntry) error
}

// PermissionAsk is one requested permission.
type PermissionAsk struct {
	Name        string // machine name, returned in the grant list
	Label       string // optional display name; defaults to Name
	Scope       string // e.g. a host list or folder
	Explanation string // plain English; if empty a default is derived
}

// PermissionRequest asks the user to approve an extension's permissions.
type PermissionRequest struct {
	Extension string
	Publisher string
	Source    string
	Asks      []PermissionAsk
}

// PermissionResult is the user's decision. Granted holds Name values.
type PermissionResult struct {
	Granted   []string
	Cancelled bool
}

// Options configures the App. Zero values are valid except Catalog and Exec.
type Options struct {
	Config  core.Config
	Theme   func() theme.Theme // called on every show so reloads apply
	Catalog Catalog
	Exec    core.Executor
	Record  func(it core.Item, query string) // history hook, called after a successful execute
	// Choose is called after a successful execute with the typed query and the
	// Tab scope the item was run in (nil at the root); Failed when it errored.
	// Together they feed the remembered picks (core.Choices).
	Choose      func(it core.Item, query string, scope *core.Scope)
	Failed      func(it core.Item)
	Highlight   HighlightFunc // fuzzy match indices; defaults to SubsequenceMatch
	Reload      func() error  // reload extensions and themes
	ReloadTheme func() error  // cheap: re-read the config and theme only (optional)
	ConfigPath  string        // zagros.toml, opened by tray "Settings"
	ConfigDir   string
	Ext         ExtAdmin   // nil disables the manager's Installed tab
	Market      Market     // nil disables the Marketplace tab
	Themes      ThemeAdmin // nil disables the Themes tab
	Version     string
	FadeMS      int // fade-in duration, capped to 100; 0 disables
	// OnAutostart is called after the Start with Windows toggle changes.
	OnAutostart func(enabled bool)
	// CheckUpdate is run (off the UI thread) by the tray item "Check for
	// updates"; it returns the text for the balloon notification.
	CheckUpdate func() string
	// MCPOn reports whether the HTTP MCP server runs; MCPToggle flips it
	// (returns a status line for the balloon); AgentPrompt is the text the tray
	// item "Copy agent prompt" puts on the clipboard. All optional.
	MCPOn       func() bool
	MCPToggle   func() string
	AgentPrompt func() string
	// OnQuit is called when Exit is chosen, before the loop ends.
	OnQuit func()
	// OnReady runs on its own goroutine once the window and tray icon exist;
	// start extension discovery here so permission prompts can be shown.
	OnReady func()
	// IconICO is the app icon (.ico bytes) used for the tray and dialogs.
	IconICO []byte
}

// ScopeCatalog is implemented by catalogs that can answer queries inside a
// core.Scope (the context Tab enters).
type ScopeCatalog interface {
	QueryScope(ctx context.Context, s core.Scope, q string) []core.Item
}

// StreamCatalog is implemented by catalogs that deliver results
// progressively; the launcher uses it when available.
type StreamCatalog interface {
	QueryStream(ctx context.Context, q string, emit func(core.Snapshot))
}
