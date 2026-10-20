package core

import (
	"context"
	"strings"
	"time"
)

type providerResult struct {
	p      Provider
	items  []Item
	routed bool
	tier   int
}

// Query runs q against all enabled providers and returns ranked, deduplicated
// items with Score set, best first. It waits for every provider (within its
// tier budget); interactive callers should use QueryStream.
func (c *Catalog) Query(ctx context.Context, q string) []Item {
	var last []Item
	c.stream(ctx, q, 0, func(s Snapshot) { last = s.Items })
	return last
}

// QueryStream runs q and calls emit with a merged, re-ranked snapshot every
// time a provider's results arrive. Tier-0 providers start immediately;
// tier-1 providers start after a short debounce and have a longer budget.
// emit is called from a single goroutine (the caller's), never after ctx is
// cancelled, and the last call has Done set (unless ctx was cancelled first).
// QueryStream returns when the query is complete.
func (c *Catalog) QueryStream(ctx context.Context, q string, emit func(Snapshot)) {
	c.mu.RLock()
	d := c.debounce
	c.mu.RUnlock()
	c.stream(ctx, q, d, emit)
}

// SetSlowBudget sets the budget of tier-1 providers (default 1.5 s) and the
// debounce before they start (default 40 ms).
func (c *Catalog) SetSlowBudget(budget, debounce time.Duration) {
	c.mu.Lock()
	c.slowTimeout, c.debounce = budget, debounce
	c.mu.Unlock()
}

// OnInvalidate registers fn, called (from any goroutine) when a provider's
// data changed so the current query should be run again.
func (c *Catalog) OnInvalidate(fn func()) {
	c.mu.Lock()
	c.invalidate = fn
	ps := append([]Provider(nil), c.providers...)
	c.mu.Unlock()
	for _, p := range ps {
		c.hookReady(p)
	}
}

func (c *Catalog) hookReady(p Provider) {
	if rn, ok := p.(ReadyNotifier); ok {
		rn.OnReady(func() {
			c.mu.RLock()
			fn := c.invalidate
			c.mu.RUnlock()
			if fn != nil {
				fn()
			}
		})
	}
}

// Warm asks every provider that has an index to load it, in the background.
func (c *Catalog) Warm() {
	for _, p := range c.Providers() {
		if w, ok := p.(Warmer); ok {
			go func() {
				defer func() { _ = recover() }()
				t := time.Now()
				w.Warm()
				Trace.Event("warm", time.Since(t), "%s", p.Name())
			}()
		}
	}
}

func (c *Catalog) stream(ctx context.Context, q string, debounce time.Duration, emit func(Snapshot)) {
	raw := q
	q = strings.TrimSpace(q)

	c.mu.RLock()
	var active []Provider
	for _, p := range c.providers {
		if !c.disabled[p.Name()] {
			active = append(active, p)
		}
	}
	limit, timeout, slow := c.limit, c.timeout, c.slowTimeout
	c.mu.RUnlock()

	routedName, routedQ, _ := c.SplitKeyword(q)

	var notReady []string
	var pending [2]int
	for _, p := range active {
		pending[TierOf(p)]++
		if w, ok := p.(Warmer); ok && !w.Ready() {
			notReady = append(notReady, p.Name())
		}
	}

	ctx, cancelAll := context.WithCancel(context.WithValue(ctx, rawQueryKey{}, raw))
	defer cancelAll()

	results := make(chan providerResult, len(active))
	run := func(p Provider, tier int) {
		routed := p.Name() == routedName
		pq := q
		if routed {
			pq = routedQ
		}
		budget := timeout
		if tier == 1 {
			budget = slow
		}
		pctx := ctx
		if budget > 0 {
			var cancel context.CancelFunc
			pctx, cancel = context.WithTimeout(ctx, budget)
			defer cancel()
		}
		var items []Item
		start := time.Now()
		defer func() {
			_ = recover() // a misbehaving provider must not take the app down
			Trace.Event("provider", time.Since(start), "%s tier=%d n=%d q=%q", p.Name(), tier, len(items), pq)
			results <- providerResult{p: p, items: items, routed: routed, tier: tier}
		}()
		items = p.Query(pctx, pq)
	}
	for _, p := range active {
		if TierOf(p) == 0 {
			go run(p, 0)
			continue
		}
		go func() {
			if debounce > 0 {
				t := time.NewTimer(debounce)
				select {
				case <-t.C:
				case <-ctx.Done():
					t.Stop()
					results <- providerResult{p: p, tier: 1}
					return
				}
			}
			run(p, 1)
		}()
	}

	var t0, t1 <-chan time.Time
	if timeout > 0 && pending[0] > 0 {
		t := time.NewTimer(timeout)
		defer t.Stop()
		t0 = t.C
	}
	if slow > 0 && pending[1] > 0 {
		t := time.NewTimer(debounce + slow)
		defer t.Stop()
		t1 = t.C
	}

	running := map[string]bool{} // tier-1 providers that have not answered
	for _, p := range active {
		if TierOf(p) == 1 {
			running[p.Name()] = true
		}
	}
	var tShow <-chan time.Time
	if pending[1] > 0 {
		t := time.NewTimer(debounce + slowShow)
		defer t.Stop()
		tShow = t.C
	}
	var problems []string
	showSlow := false
	names := func() []string {
		var out []string
		for _, p := range active {
			if running[p.Name()] {
				out = append(out, p.Name())
			}
		}
		return out
	}

	var all []providerResult
	timedOut := false // a budget expired: absent items prove nothing
	send := func(done bool) {
		if ctx.Err() != nil {
			return
		}
		s := Snapshot{
			Items:    c.withFallbacks(c.rank(append([]providerResult(nil), all...), q, routedQ, limit), active, q, routedName != "", limit),
			Done:     done,
			Waiting:  !done && pending[1] > 0,
			NotReady: notReady,
		}
		if showSlow && !done {
			s.Slow = names()
		}
		if done {
			s.Problems = problems
		}
		emit(s)
	}
	if pending[0] == 0 && pending[1] > 0 {
		send(false) // nothing instant: drop the previous query's items now
	}
	for pending[0]+pending[1] > 0 {
		select {
		case r := <-results:
			pending[r.tier]--
			if r.tier == 0 && pending[0] == 0 {
				t0 = nil
			}
			if r.tier == 1 {
				delete(running, r.p.Name())
				if f, ok := r.p.(Failer); ok && len(r.items) == 0 {
					if msg := f.Failure(); msg != "" {
						problems = append(problems, r.p.Name()+": "+msg)
					}
				}
			}
			if len(r.items) > 0 {
				all = append(all, r)
			}
			rest := pending[0] + pending[1]
			// Finishing tier 0 always yields a snapshot so stale items go away
			// even when nothing matched.
			if rest > 0 && (len(r.items) > 0 || (r.tier == 0 && pending[0] == 0)) {
				send(false)
			}
		case <-t0:
			pending[0], t0, timedOut = 0, nil, true
			if pending[1] > 0 {
				send(false)
			}
		case <-tShow:
			tShow = nil
			if pending[1] > 0 {
				showSlow = true
				send(false)
			}
		case <-t1:
			for _, n := range names() {
				problems = append(problems, n+" did not answer in time")
			}
			pending[1], t1, timedOut = 0, nil, true
		case <-ctx.Done():
			return
		}
	}
	send(true)
	if !timedOut && ctx.Err() == nil {
		c.validateChoices(q, all, active)
	}
}

// validateChoices drops remembered picks for q that the complete answer no
// longer contains, e.g. an app that was uninstalled. Only items of providers
// that ran and are ready count, so a disabled provider or a slow index does
// not erase anything.
func (c *Catalog) validateChoices(q string, all []providerResult, active []Provider) {
	ch := c.Choices()
	if !ch.HasQuery(q) {
		return
	}
	seen := map[string]bool{}
	for _, r := range all {
		for _, it := range r.items {
			seen[it.ID] = true
		}
	}
	ready := map[string]bool{}
	for _, p := range active {
		w, isW := p.(Warmer)
		ready[p.Name()] = !isW || w.Ready()
	}
	ch.Validate(q,
		func(id string) bool { return Learnable(id) && ready[providerOf(id)] },
		func(id string) bool { return seen[id] })
}

type rawQueryKey struct{}

// RawQuery returns the query exactly as typed (trailing space included) for
// providers that care whether the user has typed a space after a keyword. It
// returns "" outside a catalog query.
func RawQuery(ctx context.Context) string {
	s, _ := ctx.Value(rawQueryKey{}).(string)
	return s
}

// WithRawQuery attaches a raw query to ctx (the catalog does this itself; the
// helper exists for tests of providers that read it).
func WithRawQuery(ctx context.Context, raw string) context.Context {
	return context.WithValue(ctx, rawQueryKey{}, raw)
}
