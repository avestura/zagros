package ext

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/avestura/zagros/internal/core"
)

// Provider adapts an Extension to core.Provider. It also implements the
// optional keyword interface (Keywords() []string) and Execute for "ext-call".
type Provider struct {
	e *Extension
	// StripKeyword removes a leading query keyword before calling on_query
	// (default true). Set false if the catalog already strips it.
	StripKeyword bool
	fail         atomic.Pointer[string]
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
			p.e.sweepTasks(p.e.beginQuery())
			return nil
		}
		if p.StripKeyword {
			q = strings.TrimLeft(rest, " ")
		}
	}
	items, err := p.e.Query(ctx, q)
	p.report(ctx, err)
	return items
}

// QueryScope implements core.ScopeProvider for scopes this extension offered.
func (p *Provider) QueryScope(ctx context.Context, s core.Scope, q string) []core.Item {
	items, err := p.e.QueryScope(ctx, s, q)
	p.report(ctx, err)
	return items
}

// Execute handles Kind "ext-call" actions created by this provider.
func (p *Provider) Execute(a core.Action) error {
	err := p.e.Execute(context.Background(), a)
	if err != nil {
		p.e.opts.Log(p.e.Manifest.ID, "error", "action: "+err.Error())
	}
	return err
}

var _ core.ReadyNotifier = (*Provider)(nil)

// OnReady lets the catalog re-run the query when a task() run finishes.
func (p *Provider) OnReady(fn func()) { p.e.setTaskNotify(fn) }

var _ core.Failer = (*Provider)(nil)

// report logs a failed query and remembers a short reason for Failure.
func (p *Provider) report(ctx context.Context, err error) {
	if err == nil || ctx.Err() != nil {
		p.fail.Store(nil)
		return
	}
	p.e.opts.Log(p.e.Manifest.ID, "error", err.Error())
	msg, _, _ := strings.Cut(err.Error(), "\n")
	if r := []rune(msg); len(r) > 100 {
		msg = string(r[:100]) + "..."
	}
	p.fail.Store(&msg)
}

// Failure implements core.Failer: why the latest query failed, or "".
func (p *Provider) Failure() string {
	if m := p.fail.Load(); m != nil {
		return *m
	}
	return ""
}
