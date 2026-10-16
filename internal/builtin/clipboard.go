package builtin

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
)

const (
	clipRing    = 50
	clipMaxSize = 64 * 1024
)

// clipboard keeps the last clipRing text entries in memory only (never on disk).
// It is opt-in: All creates it only when cfg.Providers["clipboard"] is true.
type clipboard struct {
	base
	mu      sync.Mutex
	entries []string // newest first
	stop    chan struct{}
	once    sync.Once
}

func newClipboard() *clipboard {
	c := &clipboard{base: base{"clipboard"}, stop: make(chan struct{})}
	go c.poll()
	return c
}

func (c *clipboard) Ranked() bool       { return true }
func (c *clipboard) Keywords() []string { return []string{"cb"} }

// Close stops the polling goroutine.
func (c *clipboard) Close() { c.once.Do(func() { close(c.stop) }) }

func (c *clipboard) poll() {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	last := clipSeq()
	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
		}
		seq := clipSeq()
		if seq == last {
			continue
		}
		last = seq
		if clipExcluded() {
			continue // password managers ask not to be monitored
		}
		if s, ok := clipText(); ok {
			c.add(s)
		}
	}
}

func (c *clipboard) add(s string) {
	if strings.TrimSpace(s) == "" || len(s) > clipMaxSize {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, e := range c.entries {
		if e == s {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			break
		}
	}
	c.entries = append([]string{s}, c.entries...)
	if len(c.entries) > clipRing {
		c.entries = c.entries[:clipRing]
	}
}

func (c *clipboard) Query(ctx context.Context, q string) []core.Item {
	rest, _, ok := trigger(q, "cb")
	if !ok {
		return nil
	}
	c.mu.Lock()
	snapshot := append([]string(nil), c.entries...)
	c.mu.Unlock()
	var out []core.Item
	for i, e := range snapshot {
		if rest != "" && !containsFold(e, rest) {
			continue
		}
		first := strings.TrimSpace(strings.SplitN(strings.TrimSpace(e), "\n", 2)[0])
		out = append(out, core.Item{
			ID: "clipboard:" + strconv.Itoa(i), Title: truncate(first, 100), Subtitle: strconv.Itoa(len([]rune(e))) + " chars", Category: "Clipboard",
			Action: noHistoryAction(copyAction(e)),
		})
		if len(out) >= 20 {
			break
		}
	}
	if len(out) == 0 {
		return []core.Item{hint("clipboard", "empty", "Clipboard history is empty", "")}
	}
	return out
}
