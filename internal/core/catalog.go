package core

import (
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

	best := map[string]int{} // ID -> index in out
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
