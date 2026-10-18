package core

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// KeywordProvider is optionally implemented by providers that are addressed
// with a leading keyword ("g golang"). When the query is "<kw> <rest>" the
// provider is queried with rest (other providers still see the full query)
// and its items are ranked first.
type KeywordProvider interface {
	Keywords() []string
}

const (
	defaultQueryTimeout = 250 * time.Millisecond
	defaultLimit        = 8
	defaultSlowTimeout  = 1500 * time.Millisecond
	defaultDebounce     = 40 * time.Millisecond
	keywordBonus        = 10000.0
	rankedBase          = 200.0
)

// Catalog fans a query out to all enabled providers and ranks the results.
type Catalog struct {
	history *History
	choices *Choices

	mu          sync.RWMutex
	providers   []Provider
	disabled    map[string]bool
	limit       int
	slowTimeout time.Duration
	debounce    time.Duration
	invalidate  func()
	timeout     time.Duration
}

// NewCatalog creates a catalog; history may be nil.
func NewCatalog(history *History) *Catalog {
	return &Catalog{
		history:     history,
		disabled:    map[string]bool{},
		limit:       defaultLimit,
		timeout:     defaultQueryTimeout,
		slowTimeout: defaultSlowTimeout,
		debounce:    defaultDebounce,
	}
}

// History returns the frecency store (possibly nil).
func (c *Catalog) History() *History { return c.history }

// SetChoices attaches the store of remembered picks (may be nil).
func (c *Catalog) SetChoices(ch *Choices) {
	c.mu.Lock()
	c.choices = ch
	c.mu.Unlock()
}

// Choices returns the store of remembered picks (possibly nil).
func (c *Catalog) Choices() *Choices {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.choices
}

// Register adds a provider. A provider with the same name replaces the old one.
func (c *Catalog) Register(p Provider) {
	c.mu.Lock()
	replaced := false
	for i, old := range c.providers {
		if old.Name() == p.Name() {
			c.providers[i] = p
			replaced = true
			break
		}
	}
	if !replaced {
		c.providers = append(c.providers, p)
	}
	c.mu.Unlock()
	c.hookReady(p)
}

// Unregister removes the provider with the given name.
func (c *Catalog) Unregister(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, p := range c.providers {
		if p.Name() == name {
			c.providers = append(c.providers[:i:i], c.providers[i+1:]...)
			return
		}
	}
}

// SetEnabled enables or disables one provider by name.
func (c *Catalog) SetEnabled(name string, on bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if on {
		delete(c.disabled, name)
	} else {
		c.disabled[name] = true
	}
}

// ApplyConfig takes provider enablement and result count from cfg.
func (c *Catalog) ApplyConfig(cfg Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = map[string]bool{}
	for name, on := range cfg.Providers {
		if !on {
			c.disabled[name] = true
		}
	}
	if cfg.MaxResults > 0 {
		c.limit = cfg.MaxResults
	}
}

// SetLimit caps the number of returned items.
func (c *Catalog) SetLimit(n int) {
	c.mu.Lock()
	c.limit = n
	c.mu.Unlock()
}

// SetTimeout sets the per-query deadline; providers still running when it
// expires are ignored.
func (c *Catalog) SetTimeout(d time.Duration) {
	c.mu.Lock()
	c.timeout = d
	c.mu.Unlock()
}

// Providers returns the registered providers.
func (c *Catalog) Providers() []Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Provider(nil), c.providers...)
}

// SplitKeyword splits "<kw> <rest>" and reports whether kw is a keyword of
// some provider. The keyword is matched case-insensitively and must be
// followed by a space.
func (c *Catalog) SplitKeyword(q string) (provider, rest string, ok bool) {
	kw, rest, found := strings.Cut(q, " ")
	if !found || kw == "" {
		return "", "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.providers {
		if c.disabled[p.Name()] {
			continue
		}
		if kp, isKP := p.(KeywordProvider); isKP {
			for _, k := range kp.Keywords() {
				if strings.EqualFold(k, kw) {
					return p.Name(), strings.TrimLeft(rest, " "), true
				}
			}
		}
	}
	return "", "", false
}

func (c *Catalog) rank(results []providerResult, q, routedQ string, limit int) []Item {
	// Order providers by registration so ties are stable regardless of which
	// goroutine finished first.
	order := map[string]int{}
	for i, p := range c.Providers() {
		order[p.Name()] = i
	}
	sort.SliceStable(results, func(i, j int) bool {
		return order[results[i].p.Name()] < order[results[j].p.Name()]
	})

	bonus := c.Choices().QueryBonuses(q) // what the user picked for this query before
	best := map[string]int{}             // ID -> index in out
	var out []Item
	for _, r := range results {
		pat := q
		if r.routed {
			pat = routedQ
		}
		n := len(r.items)
		for pos, it := range r.items {
			if r.p.Ranked() {
				it.Score = rankedBase + float64(n-pos)
			} else {
				s, ok := FuzzyScore(pat, it.Title)
				if !ok {
					continue
				}
				// Shift so that weak matches stay positive and the boost
				// (a multiplier) always helps.
				it.Score = float64(s+40) * c.history.Boost(it.ID)
				if it.Score < 1 {
					it.Score = 1
				}
			}
			if r.routed {
				it.Score += keywordBonus
			}
			it.Score += bonus[it.ID]
			if it.Category == "" {
				it.Category = r.p.Name()
			}
			if it.ID != "" {
				if i, dup := best[it.ID]; dup {
					if it.Score > out[i].Score {
						out[i] = it
					}
					continue
				}
				best[it.ID] = len(out)
			}
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// maxFallbackRows caps the fallback rows (web search, tool families) so they
// never push more than half of a short list's real results out.
const maxFallbackRows = 4

// withFallbacks appends the providers' fallback items (web search, the tool
// families) after the ranked results. They are skipped for keyword-routed
// queries, and they take the last rows when the list is full so the window
// never grows past limit. Each provider gets its first row before any provider
// gets a second, up to maxFallbackRows.
func (c *Catalog) withFallbacks(out []Item, active []Provider, q string, routed bool, limit int) []Item {
	if routed || q == "" {
		return out
	}
	taken := map[string]bool{}
	for _, o := range out {
		taken[o.ID] = true
	}
	var lists [][]Item
	for _, p := range active {
		f, ok := p.(Fallbacker)
		if !ok {
			continue
		}
		var l []Item
		for _, it := range f.Fallback(q) {
			if taken[it.ID] {
				continue
			}
			if it.Category == "" {
				it.Category = p.Name()
			}
			l = append(l, it)
		}
		if len(l) > 0 {
			lists = append(lists, l)
		}
	}
	room := maxFallbackRows
	if limit > 0 {
		room = min(room, max(limit/2, 1))
	}
	take := make([]int, len(lists))
	for n := 0; n < room; {
		progressed := false
		for i := range lists {
			if n < room && take[i] < len(lists[i]) {
				take[i]++
				n++
				progressed = true
			}
		}
		if !progressed {
			break
		}
	}
	var fb []Item
	for i, l := range lists {
		fb = append(fb, l[:take[i]]...)
	}
	if len(fb) == 0 {
		return out
	}
	if limit > 0 && len(out)+len(fb) > limit {
		keep := max(limit-len(fb), 0)
		if keep < len(out) {
			out = out[:keep]
		}
	}
	return append(out, fb...)
}

// QueryScope returns the items of scope s for the input text q: the owning
// provider's answer, or the static item list filtered by q. It is fast and
// not streamed.
func (c *Catalog) QueryScope(ctx context.Context, s Scope, q string) []Item {
	q = strings.TrimSpace(q)
	var items []Item
	if s.Provider == "" {
		type scored struct {
			it Item
			sc float64
		}
		var keep []scored
		for _, it := range s.Items {
			if q == "" {
				keep = append(keep, scored{it, 0})
			} else if sc, ok := FuzzyScore(q, it.Title); ok {
				keep = append(keep, scored{it, float64(sc)})
			}
		}
		sort.SliceStable(keep, func(i, j int) bool { return keep[i].sc > keep[j].sc })
		for _, k := range keep {
			items = append(items, k.it)
		}
		return items
	}
	for _, p := range c.Providers() {
		if p.Name() != s.Provider {
			continue
		}
		if sp, ok := p.(ScopeProvider); ok {
			items = sp.QueryScope(ctx, s, q)
		}
		break
	}
	for i := range items {
		if items[i].Category == "" {
			items[i].Category = s.Provider
		}
	}
	sortByBonus(items, c.Choices().ScopeBonuses(s.ID)) // the tool picked here last time leads
	const maxScopeItems = 400                          // scope lists scroll, so they are not cut to max_results
	if len(items) > maxScopeItems {
		items = items[:maxScopeItems]
	}
	return items
}
