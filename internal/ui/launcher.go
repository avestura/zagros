//go:build windows

package ui

import (
	"fmt"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
)

type fonts struct {
	main, bold, small, input uintptr
	hMain, hSmall, hInput    int
}

type backBuf struct {
	dc, bmp, old uintptr
	w, h         int
}

func screenDC() uintptr { r, _, _ := pGetDC.Call(0); return r }

// prepare builds fonts and metrics for the given theme/DPI. Fonts are GDI
// objects created once per (face, size, DPI) and kept across hide/show; only
// the cheap metrics are recomputed. It is used by show, the background
// pre-warm and the offscreen self-test.
func (a *App) prepare(dpi int) {
	a.th = a.theme()
	a.dpi = dpi
	key := fmt.Sprintf("%s|%d|%d", a.th.Font, a.th.FontSize, dpi)
	if a.fonts.main == 0 || key != a.fontKey {
		t0 := time.Now()
		a.freeFonts()
		px := func(pt int) int { return (pt*dpi + 36) / 72 }
		a.fonts.main = makeFont(a.th.Font, px(a.th.FontSize), false)
		a.fonts.bold = makeFont(a.th.Font, px(a.th.FontSize), true)
		a.fonts.small = makeFont(a.th.Font, px(max(a.th.FontSize-2, 7)), false)
		a.fonts.input = makeFont(a.th.Font, px(a.th.FontSize+4), false)
		dc := screenDC()
		a.fonts.hMain = fontHeight(dc, a.fonts.main)
		a.fonts.hSmall = fontHeight(dc, a.fonts.small)
		a.fonts.hInput = fontHeight(dc, a.fonts.input)
		pReleaseDC.Call(0, dc)
		a.fontKey = key
		core.Trace.Event("fonts", time.Since(t0), "%s", key)
	}
	a.m = ComputeMetrics(a.th, dpi, a.opt.Config.MaxResults, a.fonts.hSmall+Scale(10, dpi))
	a.sc.Reset(0, a.m.VisibleRows(0))
}

func (a *App) freeFonts() {
	for _, f := range []*uintptr{&a.fonts.main, &a.fonts.bold, &a.fonts.small, &a.fonts.input} {
		if *f != 0 {
			pDeleteObject.Call(*f)
			*f = 0
		}
	}
}

// freeResources drops per-session state when the launcher hides. GDI objects
// (fonts, back buffer) and the icon cache are kept; trimIdle releases them
// after a long idle period.
func (a *App) freeResources() {
	a.items = nil
	a.sc.Reset(0, 0)
}

// trimIdle runs after the launcher stayed hidden for a while.
func (a *App) trimIdle() {
	if a.visible {
		return
	}
	a.icons.Trim(iconIdleKeep)
	a.freeBack()
	go trimMemory()
}

func (a *App) freeBack() {
	if a.back.dc != 0 {
		pSelectObject.Call(a.back.dc, a.back.old)
		pDeleteObject.Call(a.back.bmp)
		pDeleteDC.Call(a.back.dc)
		a.back = backBuf{}
	}
}

func (a *App) ensureBack(ref uintptr, w, h int) {
	if a.back.dc != 0 && a.back.w >= w && a.back.h >= h {
		return
	}
	a.freeBack()
	dc, _, _ := pCreateCompatibleDC.Call(ref)
	bmp, _, _ := pCreateCompatibleBitmap.Call(ref, uintptr(w), uintptr(h))
	old, _, _ := pSelectObject.Call(dc, bmp)
	a.back = backBuf{dc: dc, bmp: bmp, old: old, w: w, h: h}
}

// ---- show / hide ----------------------------------------------------------

func (a *App) toggle() {
	if a.visible {
		a.hide()
	} else {
		a.showLauncher()
	}
}

func (a *App) showLauncher() {
	if a.visible {
		forceForeground(a.hwnd)
		return
	}
	work, dpi := monitorAtCursor()
	pKillTimer.Call(uintptr(a.hwnd), timerIdle)
	a.prepare(dpi)
	a.ed.Clear()
	a.hscroll = 0
	a.status = ""
	a.items = nil
	a.sc.Reset(0, 0)
	h := a.m.Height(0)
	x := int(work.Left) + (int(work.Right-work.Left)-a.m.Width)/2
	y := int(work.Top) + int(work.Bottom-work.Top)/5
	a.applyCorners()
	a.alpha = 255
	if a.opt.FadeMS > 0 {
		a.alpha = 0
	}
	if a.opt.FadeMS > 0 {
		pSetLayeredWindowAttrs.Call(uintptr(a.hwnd), 0, uintptr(a.alpha), lwaAlpha)
	}
	a.userMoved = false
	a.curH = h
	pSetWindowPos.Call(uintptr(a.hwnd), hwndTopmost, uintptr(int32(x)), uintptr(int32(y)),
		uintptr(a.m.Width), uintptr(h), swpShowWindow)
	a.setRegion(h)
	a.visible = true
	a.caretOn = true
	forceForeground(a.hwnd)
	pSetTimer.Call(uintptr(a.hwnd), timerCaret, 530, 0)
	if a.opt.FadeMS > 0 {
		a.fadeStep = 0
		pSetTimer.Call(uintptr(a.hwnd), timerFade, 15, 0)
	}
	a.keyAt = time.Time{}
	a.startQuery()
	invalidate(a.hwnd)
}

func (a *App) hide() {
	if !a.visible {
		return
	}
	a.visible = false
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	a.qseq++ // drop in-flight results
	pKillTimer.Call(uintptr(a.hwnd), timerApply)
	a.applyTimer = false
	a.snap.Take()
	pKillTimer.Call(uintptr(a.hwnd), timerCaret)
	pKillTimer.Call(uintptr(a.hwnd), timerFade)
	pShowWindow.Call(uintptr(a.hwnd), swHide)
	a.freeResources()
	pSetTimer.Call(uintptr(a.hwnd), timerIdle, idleTrimMS, 0)
}

func (a *App) applyCorners() {
	pref := uint32(dwmwcpDoNotRound)
	if a.th.Radius > 0 {
		pref = dwmwcpRound
	}
	r, _, _ := pDwmSetWindowAttribute.Call(uintptr(a.hwnd), dwmwaWindowCornerPreference,
		uintptr(unsafe.Pointer(&pref)), 4)
	if r != 0 && a.th.Radius > 0 { // pre-Windows 11: clip with a region
		a.roundRegion = true
	} else {
		a.roundRegion = false
		pSetWindowRgn.Call(uintptr(a.hwnd), 0, 0)
	}
}

// windowHeight is the height the window needs now. It only depends on the
// row count (capped by max_results), so it changes in whole-row steps.
func (a *App) windowHeight() int {
	h := a.m.Height(len(a.items))
	if len(a.items) == 0 && a.status != "" {
		h += a.m.StatusH
	}
	return h
}

// resizeToResults resizes the window only when the height tier changed.
func (a *App) resizeToResults() {
	h := a.windowHeight()
	if h == a.curH {
		return
	}
	a.curH = h
	if a.bench != nil {
		a.bench.mu.Lock()
		a.bench.resizes++
		a.bench.mu.Unlock()
	}
	pSetWindowPos.Call(uintptr(a.hwnd), 0, 0, 0, uintptr(a.m.Width), uintptr(h), swpNoMove|swpNoActivate|0x0004)
	a.setRegion(h)
}

func (a *App) setRegion(h int) {
	if a.roundRegion {
		d := uintptr(Scale(a.th.Radius*2, a.dpi))
		rgn, _, _ := pCreateRoundRectRgn.Call(0, 0, uintptr(a.m.Width+1), uintptr(h+1), d, d)
		pSetWindowRgn.Call(uintptr(a.hwnd), rgn, 1)
	}
}

func (a *App) onTimer(id uintptr) {
	switch id {
	case timerCaret:
		a.caretOn = !a.caretOn
		a.refreshInput()
	case timerApply:
		pKillTimer.Call(uintptr(a.hwnd), timerApply)
		a.applyTimer = false
		a.gate.Applied(time.Now())
		a.applyPending()
	case timerIdle:
		pKillTimer.Call(uintptr(a.hwnd), timerIdle)
		a.trimIdle()
	case timerFade:
		a.fadeStep++
		total := max(a.opt.FadeMS/15, 1)
		if a.fadeStep >= total {
			a.alpha = 255
			pKillTimer.Call(uintptr(a.hwnd), timerFade)
		} else {
			a.alpha = 255 * a.fadeStep / total
		}
		pSetLayeredWindowAttrs.Call(uintptr(a.hwnd), 0, uintptr(a.alpha), lwaAlpha)
	}
}

// ---- keyboard -----------------------------------------------------------

func (a *App) selected() (core.Item, bool) {
	if a.sc.Sel < 0 || a.sc.Sel >= len(a.items) {
		return core.Item{}, false
	}
	return a.items[a.sc.Sel], true
}

func (a *App) edited(before string) {
	changed := a.ed.Text() != before
	a.caretOn = true
	// Echo first: the input band is painted before the query even starts, so
	// typing never waits on providers, the result list or icons.
	a.refreshInput()
	if a.bench != nil && changed {
		a.bench.noteEcho()
	}
	if changed {
		now := time.Now()
		a.fastTyping = !a.keyAt.IsZero() && now.Sub(a.keyAt) < fastKeyGap
		a.keyAt = now
		a.startQuery()
	}
	if a.listDirty {
		a.listDirty = false
		a.invalidateList()
	}
}

func (a *App) onKeyDown(vk int) {
	ctrl, shift := keyDown(vkControl), keyDown(vkShift)
	before := a.ed.Text()
	defer func() { a.edited(before) }()
	switch vk {
	case vkEscape:
		a.hide()
		return
	case vkUp:
		a.userMoved = true
		a.listDirty = true
		a.sc.Move(-1, true)
	case vkDown:
		a.userMoved = true
		a.listDirty = true
		a.sc.Move(1, true)
	case vkPrior:
		a.userMoved = true
		a.listDirty = true
		a.sc.PageUp()
	case vkNext:
		a.userMoved = true
		a.listDirty = true
		a.sc.PageDown()
	case vkHome:
		if ctrl || a.ed.Len() == 0 {
			a.userMoved = true
			a.listDirty = true
			a.sc.Home()
		} else {
			a.ed.Home(shift)
		}
	case vkEnd:
		if ctrl || a.ed.Len() == 0 {
			a.userMoved = true
			a.listDirty = true
			a.sc.End()
		} else {
			a.ed.End(shift)
		}
	case vkLeft:
		a.ed.Move(-1, ctrl, shift)
	case vkRight:
		a.ed.Move(1, ctrl, shift)
	case vkBack:
		if ctrl {
			a.ed.CtrlBackspace()
		} else {
			a.ed.Backspace()
		}
	case vkDelete:
		if ctrl {
			a.ed.CtrlDelete()
		} else {
			a.ed.Delete()
		}
	case vkReturn:
		if it, ok := a.selected(); ok {
			act := it.Action
			if ctrl && len(it.Alt) > 0 {
				act = it.Alt[0].Action
			}
			a.activate(it, act)
		}
	case vkTab:
		a.onTab()
	case 'A':
		if ctrl {
			a.ed.SelectAll()
		}
	case 'C':
		if ctrl && a.ed.HasSelection() {
			SetClipboard(a.ed.SelectedText())
		}
	case 'X':
		if ctrl && a.ed.HasSelection() {
			SetClipboard(a.ed.SelectedText())
			a.ed.Backspace()
		}
	case 'V':
		if ctrl {
			a.ed.Insert(GetClipboard())
		}
	default:
		if ctrl && vk >= '1' && vk <= '9' {
			if i := a.sc.Abs(vk - '1'); i >= 0 {
				it := a.items[i]
				a.activate(it, it.Action)
			}
		}
	}
}

func (a *App) onChar(r rune) {
	switch {
	case r >= 0xD800 && r < 0xDC00:
		a.highSur = r
		return
	case r >= 0xDC00 && r < 0xE000:
		if a.highSur != 0 {
			r = utf16.DecodeRune(a.highSur, r)
			a.highSur = 0
		} else {
			return
		}
	case r < 0x20 || r == 0x7f:
		return
	}
	a.highSur = 0
	before := a.ed.Text()
	a.ed.Insert(string(r))
	a.edited(before)
}

// activate runs an action: setquery stays in the launcher, everything else
// hides it first and runs off the UI thread.
func (a *App) activate(it core.Item, act core.Action) {
	if act.Kind == "setquery" {
		a.ed.SetText(act.Target)
		a.startQuery()
		invalidate(a.hwnd)
		return
	}
	a.hide()
	exec, rec, query := a.opt.Exec, a.opt.Record, a.query
	go func() {
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
			}()
			if exec != nil {
				err = exec.Execute(act)
			}
		}()
		if err != nil {
			a.Notify("Zagros: "+it.Title, err.Error())
			return
		}
		if rec != nil {
			rec(it, query)
		}
	}()
}

// onTab is Keypirinha-style completion: when the selected item offers a
// completion (a keyword such as "g " for web search) Tab puts it in the input
// so the user can type the query. Otherwise Tab opens the item's actions menu.
func (a *App) onTab() {
	if it, ok := a.selected(); ok && it.Complete != "" && it.Complete != a.ed.Text() {
		a.ed.SetText(it.Complete)
		return // edited() re-runs the query
	}
	a.showActionsMenu()
}

func (a *App) showActionsMenu() {
	it, ok := a.selected()
	if !ok || len(it.Alt) == 0 {
		a.status = "No other actions for this item"
		a.listDirty = true
		return
	}
	menu, _, _ := pCreatePopupMenu.Call()
	defer pDestroyMenu.Call(menu)
	for i, na := range it.Alt {
		pAppendMenuW.Call(menu, mfString, uintptr(i+1), uintptr(unsafe.Pointer(u16(na.Name))))
	}
	row := max(a.sc.Rel(a.sc.Sel), 0)
	rr := a.m.RowRect(row)
	var wr rect
	pGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&wr)))
	a.inMenu = true
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightButton,
		uintptr(int(wr.Left)+rr.R-Scale(200, a.dpi)), uintptr(int(wr.Top)+rr.B), 0, uintptr(a.hwnd), 0)
	a.inMenu = false
	if cmd > 0 && int(cmd) <= len(it.Alt) {
		a.activate(it, it.Alt[cmd-1].Action)
	}
}

// ---- mouse ----------------------------------------------------------------

// prefixWidths measures the input text prefixes with the input font.
func (a *App) prefixWidths() []int {
	if a.back.dc == 0 {
		return []int{0}
	}
	old, _, _ := pSelectObject.Call(a.back.dc, a.fonts.input)
	defer pSelectObject.Call(a.back.dc, old)
	rs := []rune(a.ed.Text())
	w := make([]int, len(rs)+1)
	for i := 1; i <= len(rs); i++ {
		w[i] = textWidth(a.back.dc, string(rs[:i]))
	}
	return w
}

func (a *App) caretFromX(x int) int {
	ir := a.m.InputRect()
	return CaretIndexAt(a.prefixWidths(), x-(ir.L+a.m.Pad)+a.hscroll)
}

func (a *App) onMouseDown(x, y int) {
	if a.m.InputRect().Contains(x, y) {
		a.ed.SetCaret(a.caretFromX(x), keyDown(vkShift))
		a.dragging = true
		pSetCapture.Call(uintptr(a.hwnd))
		a.caretOn = true
		invalidate(a.hwnd)
		return
	}
	if r := a.m.HitRow(x, y, len(a.items)); r >= 0 {
		if i := a.sc.Abs(r); i >= 0 {
			a.userMoved = true
			a.listDirty = true
			a.sc.Select(i)
			invalidate(a.hwnd)
		}
	}
}

func (a *App) onMouseDbl(x, y int) {
	if r := a.m.HitRow(x, y, len(a.items)); r >= 0 {
		if i := a.sc.Abs(r); i >= 0 {
			it := a.items[i]
			a.activate(it, it.Action)
		}
	}
}

func (a *App) onMouseMove(x, y int, left bool) {
	cur := uintptr(idcArrow)
	if a.m.InputRect().Contains(x, y) {
		cur = idcIBeam
	}
	c, _, _ := pLoadCursorW.Call(0, cur)
	pSetCursor.Call(c)
	if a.dragging && left {
		a.ed.SetCaret(a.caretFromX(x), true)
		invalidate(a.hwnd)
	}
}
