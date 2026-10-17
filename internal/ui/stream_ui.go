//go:build windows

package ui

import (
	"context"
	"strings"
	"time"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
)

const (
	// applyInterval is the minimum time between two applied result updates
	// (about one display frame).
	applyInterval = 16 * time.Millisecond
	// idleTrimMS is how long the launcher stays hidden before caches are
	// trimmed.
	idleTrimMS = 5 * 60 * 1000
	// typingQuiet is how long keys must pause before the list repaints during a
	// fast burst; typingMaxDefer bounds the postponement.
	typingQuiet    = 50 * time.Millisecond
	typingMaxDefer = 160 * time.Millisecond
	// fastKeyGap: a key arriving sooner than this after the previous one is
	// part of a burst.
	fastKeyGap = 120 * time.Millisecond
)

// snapMsg is one progressive result update handed from the query goroutine to
// the UI thread.
type snapMsg struct {
	seq  uint64
	snap core.Snapshot
}

// traceMark is a latency measurement waiting for the paint that shows it.
type traceMark struct {
	name   string
	since  time.Time
	detail string
}

// ---- queries ------------------------------------------------------------

// startQuery starts a new query for the current editor text (the user typed).
func (a *App) startQuery() {
	a.userMoved = false
	a.runQuery()
}

// rerunQuery repeats the current query after a provider's data changed. It
// keeps the user's selection.
func (a *App) rerunQuery() {
	if a.visible && a.ed.Text() != "" {
		a.runQuery()
	}
}

func (a *App) runQuery() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.opt.Catalog == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.qseq++
	seq := a.qseq
	q := a.ed.Text()
	a.query = q
	a.qKeyAt = a.keyAt
	a.qFirst, a.qItems = false, false
	cat := a.opt.Catalog
	hwnd := a.hwnd
	emit := func(s core.Snapshot) {
		if ctx.Err() != nil {
			return
		}
		if a.snap.Put(snapMsg{seq: seq, snap: s}) {
			postMessage(hwnd, msgResults, 0, 0)
		}
	}
	go func() {
		defer func() { _ = recover() }()
		if sc, ok := cat.(StreamCatalog); ok {
			sc.QueryStream(ctx, q, emit)
			return
		}
		items := cat.Query(ctx, q)
		emit(core.Snapshot{Items: items, Done: true})
	}()
}

// scheduleApply runs on msgResults: apply the newest snapshot now, or after
// the rest of the current frame interval so bursts cost one repaint.
func (a *App) scheduleApply() {
	if !a.visible {
		a.snap.Take()
		return
	}
	if a.applyTimer {
		return
	}
	now := time.Now()
	if a.fastTyping && !a.keyAt.IsZero() {
		// Burst typing: let the input echo first and repaint the list once the
		// keys pause (bounded, so a long burst still shows progress).
		if q := typingQuiet - now.Sub(a.keyAt); q > 0 && (a.gate.last.IsZero() || now.Sub(a.gate.last) < typingMaxDefer) {
			a.applyTimer = true
			pSetTimer.Call(uintptr(a.hwnd), timerApply, uintptr((q+time.Millisecond-1)/time.Millisecond), 0)
			return
		}
	}
	w := a.gate.Wait(now)
	if w == 0 {
		a.applyPending()
		return
	}
	a.applyTimer = true
	ms := (w + time.Millisecond - 1) / time.Millisecond
	pSetTimer.Call(uintptr(a.hwnd), timerApply, uintptr(ms), 0)
}

func (a *App) applyPending() {
	m, ok := a.snap.Take()
	if !ok || m.seq != a.qseq || !a.visible {
		return // stale
	}
	a.applySnapshot(m.snap)
}

// applySnapshot shows one merged result list. The selection follows its item
// when the user moved it; the window is only resized when the number of rows
// changed, and nothing is repainted when nothing changed.
func (a *App) applySnapshot(s core.Snapshot) {
	items := s.Items
	status := ""
	if len(s.NotReady) > 0 && strings.TrimSpace(a.query) != "" {
		status = "Indexing " + strings.Join(s.NotReady, ", ") + "..."
	}
	a.lastDone = s.Done
	sel := SelectionAfterUpdate(a.items, a.sc.Sel, a.userMoved, items)
	changed := !SameItems(a.items, items) || status != a.status || sel != a.sc.Sel
	if !a.qKeyAt.IsZero() && core.Trace.On() {
		if !a.qFirst {
			a.qFirst = true
			a.marks = append(a.marks, traceMark{"key_to_first_paint", a.qKeyAt, ""})
		}
		if !a.qItems && len(items) > 0 {
			a.qItems = true
			a.marks = append(a.marks, traceMark{"key_to_items_paint", a.qKeyAt, ""})
		}
		if s.Done {
			a.marks = append(a.marks, traceMark{"key_to_final_paint", a.qKeyAt, ""})
		}
	}
	if changed {
		a.items = items
		a.status = status
		a.sc.Reset(len(items), a.m.VisibleRows(len(items)))
		a.sc.Select(sel)
		a.resizeToResults()
		a.invalidateList()
	} else {
		a.flushMarks(time.Now()) // nothing to repaint: what is on screen is current
	}
	if a.bench != nil {
		a.bench.noteApply()
	}
	if a.onApplied != nil {
		a.onApplied()
	}
}

func (a *App) flushMarks(now time.Time) {
	for _, m := range a.marks {
		core.Trace.Event(m.name, now.Sub(m.since), "%s", m.detail)
	}
	a.marks = a.marks[:0]
}

// ---- paint ----------------------------------------------------------------

// renderFrame draws the whole launcher into the back buffer and copies it to
// hdc. Everything it uses (fonts, back buffer, icon bitmaps) is cached; it
// never calls into the shell or the file system.
func (a *App) renderFrame(hdc uintptr) {
	w := a.m.Width
	h := a.windowHeight()
	a.ensureBack(hdc, w, a.m.Height(a.m.MaxRows)+a.m.StatusH)
	if a.skipInput {
		// Only the list area is dirty: keep the input band in the back buffer.
		ir := a.m.InputRect()
		state, _, _ := pSaveDC.Call(a.back.dc)
		pIntersectClipRect.Call(a.back.dc, 0, uintptr(ir.B), uintptr(w), uintptr(h))
		a.render(a.back.dc, w, h)
		pRestoreDC.Call(a.back.dc, state)
	} else {
		a.render(a.back.dc, w, h)
	}
	pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), a.back.dc, 0, 0, srcCopy)
}

// paintInputNow draws just the input band straight to the window, bypassing
// WM_PAINT and the result list entirely, so typing echoes immediately no
// matter how busy the list is. It returns false when it cannot (no back
// buffer yet); the caller then falls back to a full invalidate.
func (a *App) paintInputNow() bool {
	if !a.visible || a.back.dc == 0 || a.fonts.main == 0 {
		return false
	}
	ir := a.m.InputRect()
	dc := a.back.dc
	pSetBkMode.Call(dc, transparent)
	fillRect(dc, rect{int32(ir.L), int32(ir.T), int32(ir.R), int32(ir.B)}, uint32(a.th.Colors.InputBg))
	a.drawInput(dc, ir)
	hdc, _, _ := pGetDC.Call(uintptr(a.hwnd))
	if hdc == 0 {
		return false
	}
	pBitBlt.Call(hdc, uintptr(ir.L), uintptr(ir.T), uintptr(ir.R-ir.L), uintptr(ir.B-ir.T), dc, uintptr(ir.L), uintptr(ir.T), srcCopy)
	pReleaseDC.Call(uintptr(a.hwnd), hdc)
	return true
}

// invalidateList marks only the area below the input dirty.
func (a *App) invalidateList() {
	ir := a.m.InputRect()
	r := rect{0, int32(ir.B), int32(a.m.Width), int32(a.m.Height(a.m.MaxRows) + a.m.StatusH)}
	pInvalidateRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&r)), 0)
}

// refreshInput repaints the input immediately, or everything as a fallback.
func (a *App) refreshInput() {
	if !a.paintInputNow() {
		invalidate(a.hwnd)
	}
}

func (a *App) paint() {
	var t0 time.Time
	tracing := core.Trace.On()
	if tracing {
		t0 = time.Now()
	}
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&ps)))
	if !a.visible || a.fonts.main == 0 {
		return
	}
	a.skipInput = ps.Paint.Top >= int32(a.m.InputRect().B)
	a.renderFrame(hdc)
	a.skipInput = false
	if tracing {
		now := time.Now()
		core.Trace.Event("paint", now.Sub(t0), "rows=%d", len(a.items))
		a.flushMarks(now)
	}
}

// ---- background warm-up -----------------------------------------------------

func (a *App) ensureLoader() *iconLoader {
	if a.loader == nil {
		a.loader = newIconLoader(a.hwnd)
	}
	return a.loader
}

// startBackground runs after the window and tray exist. Nothing here blocks
// the message loop: the shell icon subsystem, fonts and back buffer are warmed
// on other goroutines, and providers load their indexes.
func (a *App) startBackground() {
	a.ensureLoader()
	if ic, ok := a.opt.Catalog.(interface{ OnInvalidate(func()) }); ok {
		ic.OnInvalidate(func() { a.Post(a.rerunQuery) })
	}
	go a.warmUI()
}

func (a *App) warmUI() {
	defer func() { _ = recover() }()
	t0 := time.Now()
	_, dpi := monitorAtCursor()
	th := a.theme()
	px := (th.FontSize*dpi + 36) / 72
	f := makeFont(th.Font, px, false)
	dc := screenDC()
	fontHeight(dc, f)
	old, _, _ := pSelectObject.Call(dc, f)
	textWidth(dc, "Warm-up 0123 Äb")
	pSelectObject.Call(dc, old)
	pDeleteObject.Call(f)
	pReleaseDC.Call(0, dc)
	core.Trace.Event("warm_fonts", time.Since(t0), "dpi=%d", dpi)
	a.Post(func() {
		if a.visible {
			return
		}
		t1 := time.Now()
		a.prepare(dpi)
		ref := screenDC()
		a.ensureBack(ref, a.m.Width, a.m.Height(a.m.MaxRows)+a.m.StatusH)
		pReleaseDC.Call(0, ref)
		core.Trace.Event("warm_backbuf", time.Since(t1), "%dx%d", a.m.Width, a.m.Height(a.m.MaxRows))
	})
}
