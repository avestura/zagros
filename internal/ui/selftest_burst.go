//go:build windows

package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
)

const (
	burstKeys       = 50
	burstKeyGap     = 20 * time.Millisecond
	burstSlowDelay  = 200 * time.Millisecond
	maxPaint        = 50 * time.Millisecond
	maxFastLatency  = 30 * time.Millisecond
	pmRemove        = 0x0001
	qsAllInput      = 0x04FF
	burstFinalLimit = 1500 * time.Millisecond
)

// burstProvider is a synthetic provider for the self-test.
type burstProvider struct {
	name   string
	tier   int
	delay  time.Duration
	n      int
	ranked bool
}

func (p *burstProvider) Name() string { return p.name }
func (p *burstProvider) Ranked() bool { return p.ranked }
func (p *burstProvider) Tier() int    { return p.tier }
func (p *burstProvider) Query(ctx context.Context, q string) []core.Item {
	if p.delay > 0 {
		t := time.NewTimer(p.delay)
		defer t.Stop()
		select {
		case <-t.C:
		case <-ctx.Done():
			return nil
		}
	}
	icons := []string{`C:\Windows\notepad.exe`, `C:\Windows\explorer.exe`, `C:\Windows\System32\cmd.exe`, `C:\Windows\regedit.exe`}
	items := make([]core.Item, 0, p.n)
	for i := 0; i < p.n; i++ {
		items = append(items, core.Item{
			ID: fmt.Sprintf("%s:%d", p.name, i), Title: fmt.Sprintf("%s result %d", q, i),
			Subtitle: "synthetic", Icon: icons[i%len(icons)], Category: p.name,
			Action: core.Action{Kind: "none"},
		})
	}
	return items
}

// pumpMessages dispatches queued window messages for d (the self-test has no
// message loop of its own). It blocks in MsgWaitForMultipleObjects so a posted
// message is handled immediately, like GetMessage would.
func pumpMessages(d time.Duration) {
	end := time.Now().Add(d)
	var m msg
	for {
		for {
			r, _, _ := pPeekMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
			if r == 0 {
				break
			}
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
			pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
		}
		left := time.Until(end)
		if left <= 0 {
			return
		}
		ms := uintptr(min((left+time.Millisecond-1)/time.Millisecond, 2)) // short waits: expired timers are only noticed when the queue is polled
		pMsgWait.Call(0, 0, 0, ms, qsAllInput)
	}
}

// burstTest types 50 characters at 20 ms intervals through the real query
// pipeline (catalog streaming, coalescing, icon worker) with an instant
// provider and a deliberately slow one (200 ms, tier 1), and fails when a
// frame takes longer than 50 ms to paint or when instant results take longer
// than 30 ms to appear after a keystroke.
func (a *App) burstTest() error {
	if !core.Trace.On() {
		core.Trace.EnableMemory()
	}
	cat := core.NewCatalog(nil)
	cat.Register(&burstProvider{name: "fast", n: 6})
	cat.Register(&burstProvider{name: "slow", tier: 1, delay: burstSlowDelay, n: 3, ranked: true})
	cat.SetLimit(8)

	// What the launcher did before: wait for every provider.
	t0 := time.Now()
	n := len(cat.Query(context.Background(), "abc"))
	legacy := time.Since(t0)

	a.opt.Catalog = cat
	a.prepare(96)
	a.ed.Clear()
	a.visible = true
	a.curH = a.m.Height(0)
	a.items, a.status, a.userMoved = nil, "", false
	w, hmax := a.m.Width, a.m.Height(a.m.MaxRows)+a.m.StatusH
	dc, bmp, old := offscreen(w, hmax)
	defer freeOffscreen(dc, bmp, old)
	defer func() {
		a.visible = false
		a.onApplied = nil
		if a.cancel != nil {
			a.cancel()
			a.cancel = nil
		}
		a.snap.Take()
		a.items = nil
		a.icons.Trim(0)
	}()

	var paints []time.Duration
	a.renderFrame(dc) // warm-up frame, not measured
	a.onApplied = func() {
		t := time.Now()
		a.renderFrame(dc)
		d := time.Since(t)
		paints = append(paints, d)
		core.Trace.Event("paint", d, "rows=%d", len(a.items))
		a.flushMarks(time.Now())
	}

	letters := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"
	for i := 0; i < burstKeys; i++ {
		a.onChar(rune(letters[i%len(letters)]))
		pumpMessages(burstKeyGap)
	}
	deadline := time.Now().Add(burstFinalLimit)
	for !a.lastDone && time.Now().Before(deadline) {
		pumpMessages(5 * time.Millisecond)
	}
	if !a.lastDone {
		a.snap.mu.Lock()
		has := a.snap.has
		a.snap.mu.Unlock()
		return fmt.Errorf("burst: final results never arrived (qseq=%d timer=%v slotHas=%v items=%d )", a.qseq, a.applyTimer, has, len(a.items))
	}
	slow := 0
	for _, it := range a.items {
		if strings.HasPrefix(it.ID, "slow:") {
			slow++
		}
	}
	if slow == 0 || len(a.items) == 0 || a.items[0].Category == "" {
		return fmt.Errorf("burst: final snapshot did not merge the slow provider (%d slow of %d items)", slow, len(a.items))
	}
	// Let the icon worker deliver; rows asked for icons while painting.
	for end := time.Now().Add(3 * time.Second); core.Trace.Stat("icon_load").Count == 0 && time.Now().Before(end); {
		pumpMessages(10 * time.Millisecond)
	}
	pumpMessages(50 * time.Millisecond)

	var worst time.Duration
	for _, d := range paints {
		worst = max(worst, d)
	}
	items := core.Trace.Stat("key_to_items_paint")
	fmt.Printf("selftest burst: %d keystrokes @%v, slow provider %v\n", burstKeys, burstKeyGap, burstSlowDelay)
	fmt.Printf("  before (wait for all providers, %d items): %v\n", n, legacy.Round(time.Millisecond))
	fmt.Printf("  after  (stream): key->instant results painted max %v, %d paints, worst paint %v\n",
		items.Max.Round(10*time.Microsecond), len(paints), worst.Round(10*time.Microsecond))
	fmt.Print(core.Trace.Summary())

	if worst > maxPaint {
		return fmt.Errorf("burst: a paint took %v (limit %v)", worst, maxPaint)
	}
	if items.Count == 0 {
		return fmt.Errorf("burst: no instant results were painted")
	}
	if items.Max > maxFastLatency {
		return fmt.Errorf("burst: instant results took %v after a keystroke (limit %v)", items.Max, maxFastLatency)
	}
	if core.Trace.Stat("icon_load").Count == 0 {
		return fmt.Errorf("burst: the icon worker loaded nothing")
	}
	fmt.Fprintln(os.Stderr, "selftest: burst ok")
	return nil
}
