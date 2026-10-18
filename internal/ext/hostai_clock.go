package ext

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.starlark.net/starlark"
)

const clockKey = "zagros.clock"

// callClock is the wall-clock limit of one hook call. It can be paused while
// the extension waits for a language-model provider, so the 5 s limit bounds
// the extension's own work and the provider's time is bounded by the ai
// module instead (per request, per ai.run and per call in total).
type callClock struct {
	mu        sync.Mutex
	timer     *time.Timer
	left      time.Duration
	since     time.Time
	pauses    int
	paused    time.Duration // total paused so far
	pauseFrom time.Time
	cancel    context.CancelCauseFunc
}

// newClockCtx returns a context cancelled with context.DeadlineExceeded after
// d of un-paused time, and the clock controlling it.
func newClockCtx(parent context.Context, d time.Duration) (context.Context, context.CancelFunc, *callClock) {
	ctx, cancel := context.WithCancelCause(parent)
	c := &callClock{left: d, since: time.Now(), cancel: cancel}
	c.timer = time.AfterFunc(d, func() { cancel(context.DeadlineExceeded) })
	return ctx, func() { c.timer.Stop(); cancel(nil) }, c
}

func clockOf(th *starlark.Thread) *callClock {
	c, _ := th.Local(clockKey).(*callClock)
	return c
}

// pause stops the clock until the returned function runs. total caps the time
// one call may spend paused. A nil clock (no limit running) pauses nothing.
func (c *callClock) pause(total time.Duration) (func(), error) {
	if c == nil {
		return func() {}, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.paused >= total {
		return nil, fmt.Errorf("ai: this call has already waited %s for providers, the most a single call may", total)
	}
	if c.pauses == 0 {
		if c.timer.Stop() {
			c.left -= time.Since(c.since)
		}
		c.pauseFrom = time.Now()
	}
	c.pauses++
	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.pauses--
			if c.pauses > 0 {
				return
			}
			c.paused += time.Since(c.pauseFrom)
			if c.left < time.Millisecond {
				c.left = time.Millisecond
			}
			c.since = time.Now()
			c.timer.Reset(c.left)
		})
	}, nil
}

var errAINested = errors.New("ai.run cannot be called from a tool that ai.run is running")
