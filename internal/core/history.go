package core

import (
	"encoding/json"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// frecencyHalfLife is how long it takes a use to count half as much.
	frecencyHalfLife = 14 * 24 * time.Hour
	// maxBoost is the multiplier reached by very frequent items.
	maxBoost        = 2.0
	saturationScore = 12.0
	saveDebounce    = 2 * time.Second
	maxQueriesKept  = 8
)

type historyEntry struct {
	Score   float64        `json:"score"` // decayed use count as of Last
	Last    int64          `json:"last"`  // unix seconds
	Uses    int            `json:"uses"`
	Queries map[string]int `json:"queries,omitempty"`
}

// History is a frecency store: each use adds 1 to a score that halves every
// two weeks. It is safe for concurrent use.
type History struct {
	path string
	now  func() time.Time

	mu      sync.Mutex
	entries map[string]*historyEntry
	timer   *time.Timer
	dirty   bool
}

// NewHistory loads the history stored at path (missing or corrupt files give
// an empty history). An empty path makes an in-memory store that never saves.
func NewHistory(path string) *History {
	h := &History{path: path, now: time.Now, entries: map[string]*historyEntry{}}
	if path == "" {
		return h
	}
	if data, err := os.ReadFile(path); err == nil {
		var m map[string]*historyEntry
		if json.Unmarshal(data, &m) == nil && m != nil {
			h.entries = m
		}
	}
	return h
}

// DefaultHistoryPath is DataDir()/history.json.
func DefaultHistoryPath() string { return filepath.Join(DataDir(), "history.json") }

func (h *History) decayed(e *historyEntry, now time.Time) float64 {
	dt := now.Sub(time.Unix(e.Last, 0))
	if dt <= 0 {
		return e.Score
	}
	return e.Score * math.Exp2(-float64(dt)/float64(frecencyHalfLife))
}

// Boost returns a ranking multiplier in [1, 2] for the item id.
func (h *History) Boost(id string) float64 {
	if h == nil {
		return 1
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	e := h.entries[id]
	if e == nil {
		return 1
	}
	s := h.decayed(e, h.now())
	return 1 + (maxBoost-1)*math.Min(s, saturationScore)/saturationScore
}

// Record notes that the item id was executed for query. The history is saved
// shortly afterwards.
func (h *History) Record(id, query string) {
	if h == nil || id == "" {
		return
	}
	h.mu.Lock()
	now := h.now()
	e := h.entries[id]
	if e == nil {
		e = &historyEntry{}
		h.entries[id] = e
	}
	e.Score = h.decayed(e, now) + 1
	e.Last = now.Unix()
	e.Uses++
	if query != "" {
		if e.Queries == nil {
			e.Queries = map[string]int{}
		}
		if _, ok := e.Queries[query]; ok || len(e.Queries) < maxQueriesKept {
			e.Queries[query]++
		}
	}
	h.dirty = true
	h.scheduleLocked()
	h.mu.Unlock()
}

// Forget drops the history of one item.
func (h *History) Forget(id string) {
	h.mu.Lock()
	delete(h.entries, id)
	h.dirty = true
	h.scheduleLocked()
	h.mu.Unlock()
}

// Uses returns how many times id was executed.
func (h *History) Uses(id string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if e := h.entries[id]; e != nil {
		return e.Uses
	}
	return 0
}

func (h *History) scheduleLocked() {
	if h.path == "" || h.timer != nil {
		return
	}
	h.timer = time.AfterFunc(saveDebounce, func() { _ = h.Flush() })
}

// Flush writes pending changes to disk now. Call it on shutdown.
func (h *History) Flush() error {
	h.mu.Lock()
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	if !h.dirty || h.path == "" {
		h.mu.Unlock()
		return nil
	}
	h.pruneLocked()
	data, err := json.Marshal(h.entries)
	h.dirty = false
	h.mu.Unlock()
	if err != nil {
		return err
	}
	if err := writeAtomic(h.path, data); err != nil {
		h.mu.Lock()
		h.dirty = true
		h.mu.Unlock()
		return err
	}
	return nil
}

// pruneLocked removes entries whose score decayed to nothing.
func (h *History) pruneLocked() {
	now := h.now()
	for id, e := range h.entries {
		if h.decayed(e, now) < 0.01 {
			delete(h.entries, id)
		}
	}
}

var errNoPath = errors.New("history has no path")

// Reload re-reads the file, discarding in-memory state.
func (h *History) Reload() error {
	if h.path == "" {
		return errNoPath
	}
	data, err := os.ReadFile(h.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var m map[string]*historyEntry
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	h.mu.Lock()
	h.entries = m
	h.mu.Unlock()
	return nil
}
