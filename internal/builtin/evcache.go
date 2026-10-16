package builtin

import (
	"strings"
	"sync"
	"time"
)

// evCache remembers recent Everything searches so typing on ("foo" -> "foob")
// and backspacing ("foob" -> "foo") do not go back to Everything.
type evCache struct {
	mu      sync.Mutex
	entries []evEntry // oldest first
}

type evEntry struct {
	q        string // lower-cased search text
	limit    int
	paths    []string
	complete bool // fewer results than limit: the full match set
	at       time.Time
}

const (
	evCacheSize = 16
	evCacheTTL  = 10 * time.Second
)

// evSimple reports whether q is plain words that Everything matches against
// file names only, so a local substring filter gives the same answer.
func evSimple(q string) bool {
	return q != "" && !strings.ContainsAny(q, "*?\"|!<>:;,~%^()[]{}=+&`"+`\/`)
}

func (c *evCache) put(q string, limit int, paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	q = strings.ToLower(q)
	for i, e := range c.entries {
		if e.q == q && e.limit == limit {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			break
		}
	}
	c.entries = append(c.entries, evEntry{q: q, limit: limit, paths: paths, complete: len(paths) < limit, at: time.Now()})
	if len(c.entries) > evCacheSize {
		c.entries = c.entries[1:]
	}
}

// reuse returns results for q without asking Everything, if possible: an
// identical earlier search (also the backspace case), or a plain-word search
// whose complete result set is a superset of q's.
func (c *evCache) reuse(q string, limit int) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	lq := strings.ToLower(q)
	keep := c.entries[:0]
	for _, e := range c.entries { // the index changes: do not trust old answers
		if time.Since(e.at) < evCacheTTL {
			keep = append(keep, e)
		}
	}
	c.entries = keep
	for i := len(c.entries) - 1; i >= 0; i-- {
		if e := c.entries[i]; e.q == lq && (e.limit == limit || (e.complete && e.limit > limit)) {
			return head(e.paths, limit), true
		}
	}
	if !evSimple(lq) {
		return nil, false
	}
	// The longest cached prefix that is complete.
	best := -1
	for i, e := range c.entries {
		if e.complete && evSimple(e.q) && strings.HasPrefix(lq, e.q) && (best < 0 || len(e.q) > len(c.entries[best].q)) {
			best = i
		}
	}
	if best < 0 {
		return nil, false
	}
	terms := strings.Fields(lq)
	var out []string
	for _, p := range c.entries[best].paths {
		name := strings.ToLower(p[strings.LastIndexAny(p, `\/`)+1:])
		ok := true
		for _, t := range terms {
			if !strings.Contains(name, t) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, p)
		}
	}
	return head(out, limit), true
}

func head(s []string, n int) []string {
	if n > 0 && len(s) > n {
		return s[:n]
	}
	return s
}
