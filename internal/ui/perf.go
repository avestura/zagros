package ui

import (
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
)

// This file holds the pure (no Win32) parts of the performance architecture so
// they can be unit-tested: selection preservation, update coalescing and the
// icon request queue and cache.

// SelectionAfterUpdate returns the selected index in next after the list
// changed from prev. If the user moved the selection on this query and the
// selected item is still present (matched by ID), it stays selected; otherwise
// the first row is selected.
func SelectionAfterUpdate(prev []core.Item, sel int, userMoved bool, next []core.Item) int {
	if !userMoved || sel < 0 || sel >= len(prev) || prev[sel].ID == "" {
		return 0
	}
	id := prev[sel].ID
	for i, it := range next {
		if it.ID == id {
			return i
		}
	}
	return 0
}

// SameItems reports whether two lists show identical rows, so a repaint can
// be skipped.
func SameItems(a, b []core.Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID || a[i].Title != b[i].Title || a[i].Subtitle != b[i].Subtitle ||
			a[i].Category != b[i].Category || a[i].Icon != b[i].Icon {
			return false
		}
	}
	return true
}

// rateGate limits how often updates are applied: Wait returns how long to hold
// back an update arriving at now so that applications are at least min apart.
type rateGate struct {
	min  time.Duration
	last time.Time
}

// Wait returns 0 when the update may be applied now (and records it), else the
// time left to wait.
func (g *rateGate) Wait(now time.Time) time.Duration {
	if g.last.IsZero() || now.Sub(g.last) >= g.min {
		g.last = now
		return 0
	}
	return g.min - now.Sub(g.last)
}

// Applied records that an update was applied at now (after a delayed wait).
func (g *rateGate) Applied(now time.Time) { g.last = now }

// latestSlot holds the newest value produced by a worker goroutine. Put
// reports whether the consumer needs to be woken (nothing was pending), so any
// number of Puts between two Takes cause a single wake-up.
type latestSlot[T any] struct {
	mu  sync.Mutex
	v   T
	has bool
}

func (s *latestSlot[T]) Put(v T) (wake bool) {
	s.mu.Lock()
	wake = !s.has
	s.v, s.has = v, true
	s.mu.Unlock()
	return wake
}

func (s *latestSlot[T]) Take() (v T, ok bool) {
	s.mu.Lock()
	v, ok = s.v, s.has
	var zero T
	s.v, s.has = zero, false
	s.mu.Unlock()
	return
}

// ---- icons ---------------------------------------------------------------

// iconKey identifies one icon bitmap: the file it comes from and its size in
// pixels.
type iconKey struct {
	Path string
	Size int
}

// iconQueue is the work list of the icon worker. It is small, deduplicated and
// only keeps requests for rows that are still visible.
type iconQueue struct {
	mu       sync.Mutex
	pending  []iconKey // newest last; the worker takes the newest first
	inflight map[iconKey]bool
	max      int
}

func newIconQueue(max int) *iconQueue {
	return &iconQueue{inflight: map[iconKey]bool{}, max: max}
}

// Push queues k. It reports whether k was added (false when it is already
// queued or being loaded). When full the oldest request is dropped.
func (q *iconQueue) Push(k iconKey) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.inflight[k] {
		return false
	}
	for _, p := range q.pending {
		if p == k {
			return false
		}
	}
	q.pending = append(q.pending, k)
	if len(q.pending) > q.max {
		q.pending = q.pending[1:]
	}
	return true
}

// SetVisible drops queued requests that are not in keys (rows scrolled or
// typed away).
func (q *iconQueue) SetVisible(keys []iconKey) {
	q.mu.Lock()
	defer q.mu.Unlock()
	keep := q.pending[:0]
	for _, p := range q.pending {
		for _, k := range keys {
			if p == k {
				keep = append(keep, p)
				break
			}
		}
	}
	q.pending = keep
}

// Pop takes the newest queued request and marks it in flight.
func (q *iconQueue) Pop() (iconKey, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := len(q.pending)
	if n == 0 {
		return iconKey{}, false
	}
	k := q.pending[n-1]
	q.pending = q.pending[:n-1]
	q.inflight[k] = true
	return k, true
}

// Done marks k as no longer in flight.
func (q *iconQueue) Done(k iconKey) {
	q.mu.Lock()
	delete(q.inflight, k)
	q.mu.Unlock()
}

// Len is the number of queued (not yet started) requests.
func (q *iconQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// iconLRU maps iconKey to a GDI handle with a size cap. It is confined to the
// UI thread (no locking); destroy releases an evicted handle.
type iconLRU struct {
	cap     int
	m       map[iconKey]*lruEntry
	tick    uint64
	destroy func(h uintptr)
}

type lruEntry struct {
	h    uintptr // 0 = known to have no icon
	used uint64
}

func newIconLRU(cap int, destroy func(uintptr)) *iconLRU {
	return &iconLRU{cap: cap, m: map[iconKey]*lruEntry{}, destroy: destroy}
}

func (c *iconLRU) Get(k iconKey) (h uintptr, ok bool) {
	e, ok := c.m[k]
	if !ok {
		return 0, false
	}
	c.tick++
	e.used = c.tick
	return e.h, true
}

func (c *iconLRU) Put(k iconKey, h uintptr) {
	if old, ok := c.m[k]; ok && old.h != 0 && old.h != h {
		c.destroy(old.h)
	}
	c.tick++
	c.m[k] = &lruEntry{h: h, used: c.tick}
	c.Trim(c.cap)
}

// Trim evicts least recently used entries until at most n remain.
func (c *iconLRU) Trim(n int) {
	for len(c.m) > n {
		var oldest iconKey
		var min uint64 = ^uint64(0)
		for k, e := range c.m {
			if e.used < min {
				min, oldest = e.used, k
			}
		}
		if e := c.m[oldest]; e.h != 0 {
			c.destroy(e.h)
		}
		delete(c.m, oldest)
	}
}

func (c *iconLRU) Len() int { return len(c.m) }
