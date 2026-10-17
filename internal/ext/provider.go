package ext

import (
	"context"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

// Provider adapts an Extension to core.Provider. It also implements the
// optional keyword interface (Keywords() []string) and Execute for "ext-call".
type Provider struct {
	e *Extension
	// StripKeyword removes a leading query keyword before calling on_query
	// (default true). Set false if the catalog already strips it.
	StripKeyword bool
}

var (
	_ core.Provider                    = (*Provider)(nil)
	_ core.Tiered                      = (*Provider)(nil)
	_ core.ScopeProvider               = (*Provider)(nil)
	_ interface{ Keywords() []string } = (*Provider)(nil)
)

// NewProvider wraps e.
func NewProvider(e *Extension) *Provider { return &Provider{e: e, StripKeyword: true} }

// Extension returns the wrapped extension.
func (p *Provider) Extension() *Extension { return p.e }

func (p *Provider) Name() string { return p.e.Manifest.Name }

func (p *Provider) Ranked() bool { return true }

// Tier makes the catalog treat extensions as slow I/O: they are queried after
// a short debounce and never delay built-in results.
func (p *Provider) Tier() int { return 1 }

func (p *Provider) Keywords() []string { return p.e.Keywords() }

// Query implements core.Provider. Errors are logged, never returned to the UI.
func (p *Provider) Query(ctx context.Context, q string) []core.Item {
	if kws := p.e.Manifest.Keywords; len(kws) > 0 {
		first, rest, _ := strings.Cut(strings.TrimLeft(q, " "), " ")
		matched := false
		for _, k := range kws {
			if strings.EqualFold(k, first) {
				matched = true
			}
		}
		if !matched {
			return nil
		}
		if p.StripKeyword {
			q = strings.TrimLeft(rest, " ")
		}
	}
	items, err := p.e.Query(ctx, q)
	if err != nil && ctx.Err() == nil {
		p.e.opts.Log(p.e.Manifest.ID, "error", err.Error())
	}
	return items
}

// QueryScope implements core.ScopeProvider for scopes this extension offered.
func (p *Provider) QueryScope(ctx context.Context, s core.Scope, q string) []core.Item {
	items, err := p.e.QueryScope(ctx, s, q)
	if err != nil && ctx.Err() == nil {
		p.e.opts.Log(p.e.Manifest.ID, "error", err.Error())
	}
	return items
}

// Execute handles Kind "ext-call" actions created by this provider.
func (p *Provider) Execute(a core.Action) error {
	return p.e.Execute(context.Background(), a)
}
