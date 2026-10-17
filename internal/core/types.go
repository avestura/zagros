// Package core holds the types shared by every Zagros component.
package core

import "context"

// Action is what happens when the user confirms an item.
type Action struct {
	Kind   string            // "open", "run", "copy", "func", "setquery", "none"
	Target string            // path, URL, command line, or text to copy
	Args   []string          // optional arguments for "run"
	Extra  map[string]string // free-form data for the owning provider
	// Func is used for Kind "func" (built-in providers only).
	Func func() error `json:"-"`
}

// Item is one row in the result list.
type Item struct {
	ID       string // stable id, used for frecency; "<provider>:<key>"
	Title    string
	Subtitle string
	Icon     string // optional: file path to take the shell icon from, or a glyph name
	Category string // provider display name, shown as a tag
	Action   Action
	Score    float64 // filled by the catalog after ranking
	// Alt lists secondary actions shown in the actions menu (Tab).
	Alt []NamedAction
	// Complete is the query Tab puts in the input box (Keypirinha-style
	// completion), e.g. "g " so the user can type a web search. Empty means
	// Tab opens the actions menu instead.
	Complete string
	// Refresh makes Tab re-run the same query instead of completing it, for
	// items whose results are generated (passwords).
	Refresh bool
}

type NamedAction struct {
	Name   string
	Action Action
}

// Provider produces items for a query. Query must honour ctx cancellation
// and return quickly (the UI calls it on every keystroke).
type Provider interface {
	Name() string
	// Query returns candidates for q. Providers do coarse filtering; the
	// catalog applies fuzzy ranking on Title when Ranked() is false.
	Query(ctx context.Context, q string) []Item
	// Ranked reports whether the provider already scored/ordered its results.
	Ranked() bool
}

// Executor runs an Action. Implemented in the app (shell-open, clipboard, ...).
// Fallbacker is implemented by providers that offer a last-resort item for
// queries nothing else handled well (e.g. "search the web for ..."). The
// catalog appends these items after the ranked results.
type Fallbacker interface {
	Fallback(q string) []Item
}

type Executor interface {
	Execute(a Action) error
}
