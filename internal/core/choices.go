package core

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Choices remembers what the user picked for a query, so the same query puts
// that item first next time. Typing "fir" and running Firefox makes Firefox the
// first answer to "fir" (and to "fire", which extends it). Picks made inside a
// Tab scope are remembered per scope, so the tool you chose in the String scope
// leads the String scope the next time, whatever text is in it.
//
// The store is bounded and cleans up after itself:
//   - at most maxChoiceContexts queries (least recently used go first) and
//     maxChoicesPerContext items per query or scope;
//   - a pick loses half its weight every choiceHalfLife and is deleted when it
//     falls below choiceMinWeight, so unused picks age out;
//   - an item that no longer shows up for the exact query it was learned for
//     is deleted after choiceStrikes misses in a row (see Validate), and an
//     item whose action failed is deleted at once (Forget);
//   - queries longer than maxChoiceKeyLen are never stored.
//
// Only item IDs and the typed query are stored. Rows that must never be stored
// (tool results carry the user's text) are the caller's to skip.
type Choices struct {
	path string
	now  func() time.Time

	mu      sync.Mutex
	queries map[string]map[string]*choice
	scopes  map[string]map[string]*choice
	timer   *time.Timer
	dirty   bool
}

const (
	choiceHalfLife       = 30 * 24 * time.Hour
	choiceMinWeight      = 0.15
	maxChoiceContexts    = 300
	maxChoicesPerContext = 6
	maxChoiceKeyLen      = 40
	choiceStrikes        = 3

	// The bonus must beat the fuzzy score of any alternative but stay below
	// keywordBonus, so a keyword-routed provider still answers first.
	choiceBonusBase   = 600.0
	choiceBonusPerUse = 40.0
	choiceBonusCap    = 10 // uses counted for the per-use part
)

type choice struct {
	W    float64 `json:"w"`           // decayed pick count as of Last
	Last int64   `json:"t"`           // unix seconds
	Miss int     `json:"m,omitempty"` // consecutive times the item was absent for its exact query
}

type choicesFile struct {
	Queries map[string]map[string]*choice `json:"queries,omitempty"`
	Scopes  map[string]map[string]*choice `json:"scopes,omitempty"`
}

// DefaultChoicesPath is DataDir()/choices.json.
func DefaultChoicesPath() string { return filepath.Join(DataDir(), "choices.json") }

// NewChoices loads the store at path (missing or corrupt files give an empty
// one) and drops entries that have already aged out. An empty path makes an
// in-memory store that never saves.
func NewChoices(path string) *Choices {
	c := &Choices{path: path, now: time.Now, queries: map[string]map[string]*choice{}, scopes: map[string]map[string]*choice{}}
	if path == "" {
		return c
	}
	if data, err := os.ReadFile(path); err == nil {
		var f choicesFile
		if json.Unmarshal(data, &f) == nil {
			if f.Queries != nil {
				c.queries = f.Queries
			}
			if f.Scopes != nil {
				c.scopes = f.Scopes
			}
		}
	}
	if c.pruneLocked() {
		c.dirty = true
		_ = c.Flush()
	}
	return c
}

// QueryKey normalizes a typed query for storage; "" means "do not store".
func QueryKey(q string) string {
	k := strings.ToLower(strings.Join(strings.Fields(q), " "))
	if k == "" || len(k) > maxChoiceKeyLen {
		return ""
	}
	return k
}

// unlearnable lists the providers whose item IDs are not stable (an index into
// the clipboard history, a window handle, the one calculator result), so
// remembering them would only boost the wrong row later.
var unlearnable = map[string]bool{"calc": true, "clipboard": true, "password": true, "processes": true, "windows": true, "runner": true}

// providerOf is the part of an item ID before the first colon.
func providerOf(id string) string {
	p, _, _ := strings.Cut(id, ":")
	return p
}

// Learnable reports whether picks of the item id are remembered.
func Learnable(id string) bool { return id != "" && !unlearnable[providerOf(id)] }

func (c *Choices) decayed(e *choice, now time.Time) float64 {
	dt := now.Sub(time.Unix(e.Last, 0))
	if dt <= 0 {
		return e.W
	}
	return e.W * math.Exp2(-float64(dt)/float64(choiceHalfLife))
}

func (c *Choices) record(m map[string]map[string]*choice, ctx, id string) {
	now := c.now()
	b := m[ctx]
	if b == nil {
		b = map[string]*choice{}
		m[ctx] = b
	}
	e := b[id]
	if e == nil {
		e = &choice{}
		b[id] = e
	}
	e.W = c.decayed(e, now) + 1
	e.Last, e.Miss = now.Unix(), 0
	if len(b) > maxChoicesPerContext {
		var weakest string
		low := 0.0
		for k, v := range b {
			if w := c.decayed(v, now); weakest == "" || w < low {
				weakest, low = k, w
			}
		}
		delete(b, weakest)
	}
	if len(m) > maxChoiceContexts {
		c.evictOldestLocked(m)
	}
	c.dirty = true
	c.scheduleLocked()
}

// evictOldestLocked drops the context of m whose newest pick is the oldest.
func (c *Choices) evictOldestLocked(m map[string]map[string]*choice) {
	oldest, at := "", int64(0)
	for ctx, b := range m {
		var newest int64
		for _, e := range b {
			newest = max(newest, e.Last)
		}
		if oldest == "" || newest < at {
			oldest, at = ctx, newest
		}
	}
	delete(m, oldest)
}

// RecordQuery notes that the user ran item id for the typed query q.
func (c *Choices) RecordQuery(q, id string) {
	k := QueryKey(q)
	if c == nil || k == "" || !Learnable(id) {
		return
	}
	c.mu.Lock()
	c.record(c.queries, k, id)
	c.mu.Unlock()
}

// RecordScope notes that the user ran item id inside the Tab scope scopeID.
func (c *Choices) RecordScope(scopeID, id string) {
	if c == nil || scopeID == "" || !Learnable(id) {
		return
	}
	c.mu.Lock()
	c.record(c.scopes, scopeID, id)
	c.mu.Unlock()
}

func (c *Choices) bonus(w float64) float64 {
	return choiceBonusBase + choiceBonusPerUse*min(w, choiceBonusCap)
}

// QueryBonuses returns the score bonus per item ID for query q: the picks made
// for exactly q count fully, picks made for a shorter query that q extends
// ("fir" for "fire") count half. nil when nothing was learned.
func (c *Choices) QueryBonuses(q string) map[string]float64 {
	k := QueryKey(q)
	if c == nil || k == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queries) == 0 {
		return nil
	}
	now := c.now()
	var out map[string]float64
	for n := 1; n <= len(k); n++ {
		b := c.queries[k[:n]]
		if b == nil {
			continue
		}
		f := 0.5
		if n == len(k) {
			f = 1
		}
		for id, e := range b {
			w := c.decayed(e, now)
			if w < choiceMinWeight {
				continue
			}
			if out == nil {
				out = map[string]float64{}
			}
			out[id] = max(out[id], f*c.bonus(w))
		}
	}
	return out
}

// ScopeBonuses returns the score bonus per item ID for the Tab scope scopeID.
func (c *Choices) ScopeBonuses(scopeID string) map[string]float64 {
	if c == nil || scopeID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.scopes[scopeID]
	if len(b) == 0 {
		return nil
	}
	now := c.now()
	out := map[string]float64{}
	for id, e := range b {
		if w := c.decayed(e, now); w >= choiceMinWeight {
			out[id] = c.bonus(w)
		}
	}
	return out
}

// Validate checks the items learned for the exact query q against what the
// query returned. known reports whether the item's provider answered fully (so
// absence means something); an item that is known but absent gains a strike,
// one that is present loses its strikes, and choiceStrikes in a row delete it.
func (c *Choices) Validate(q string, known, found func(id string) bool) {
	k := QueryKey(q)
	if c == nil || k == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	b := c.queries[k]
	changed := false
	for id, e := range b {
		if !known(id) {
			continue
		}
		switch {
		case found(id):
			if e.Miss != 0 {
				e.Miss, changed = 0, true
			}
		default:
			e.Miss++
			changed = true
			if e.Miss >= choiceStrikes {
				delete(b, id)
			}
		}
	}
	if len(b) == 0 {
		delete(c.queries, k)
	}
	if changed {
		c.dirty = true
		c.scheduleLocked()
	}
}

// Forget deletes every pick of the item id (its action failed or it is gone).
func (c *Choices) Forget(id string) {
	if c == nil || id == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	changed := false
	for _, m := range []map[string]map[string]*choice{c.queries, c.scopes} {
		for ctx, b := range m {
			if _, ok := b[id]; ok {
				delete(b, id)
				changed = true
				if len(b) == 0 {
					delete(m, ctx)
				}
			}
		}
	}
	if changed {
		c.dirty = true
		c.scheduleLocked()
	}
}

// HasQuery reports whether anything is remembered for exactly the query q.
func (c *Choices) HasQuery(q string) bool {
	k := QueryKey(q)
	if c == nil || k == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.queries[k]) > 0
}

// Len is the number of stored queries and scopes (for tests and diagnostics).
func (c *Choices) Len() (queries, scopes int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.queries), len(c.scopes)
}

// pruneLocked deletes picks that aged out and empty contexts; it reports
// whether anything was removed.
func (c *Choices) pruneLocked() bool {
	now := c.now()
	removed := false
	for _, m := range []map[string]map[string]*choice{c.queries, c.scopes} {
		for ctx, b := range m {
			for id, e := range b {
				if c.decayed(e, now) < choiceMinWeight {
					delete(b, id)
					removed = true
				}
			}
			if len(b) == 0 {
				delete(m, ctx)
				removed = true
			}
		}
		for len(m) > maxChoiceContexts {
			c.evictOldestLocked(m)
			removed = true
		}
	}
	return removed
}

func (c *Choices) scheduleLocked() {
	if c.path == "" || c.timer != nil {
		return
	}
	c.timer = time.AfterFunc(saveDebounce, func() { _ = c.Flush() })
}

// Flush prunes and writes pending changes now. Call it on shutdown. An empty
// store removes its file instead of leaving an empty one behind.
func (c *Choices) Flush() error {
	c.mu.Lock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	if !c.dirty || c.path == "" {
		c.mu.Unlock()
		return nil
	}
	c.pruneLocked()
	empty := len(c.queries) == 0 && len(c.scopes) == 0
	data, err := json.Marshal(choicesFile{Queries: c.queries, Scopes: c.scopes})
	c.dirty = false
	c.mu.Unlock()
	if err != nil {
		return err
	}
	if empty {
		if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := writeAtomic(c.path, data); err != nil {
		c.mu.Lock()
		c.dirty = true
		c.mu.Unlock()
		return err
	}
	return nil
}

// sortByBonus stably moves items with a bonus ahead of the rest, highest
// first; items without one keep their order.
func sortByBonus(items []Item, bonus map[string]float64) {
	if len(bonus) == 0 {
		return
	}
	sort.SliceStable(items, func(i, j int) bool { return bonus[items[i].ID] > bonus[items[j].ID] })
}
