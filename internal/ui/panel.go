//go:build windows

package ui

import (
	"strings"
	"unsafe"

	"github.com/avestura/zagros/internal/theme"
	"golang.org/x/sys/windows"
)

var pMessageBoxW = user32.NewProc("MessageBoxW")

// btn is a clickable region registered during draw.
type btn struct {
	r        Rect
	id       int
	label    string
	disabled bool
	primary  bool
	on       bool // toggled-on look
}

// panel is a themed borderless top-level window used for dialogs. Content is
// drawn by the owner through draw; buttons are registered with p.button so
// hit-testing and hover work uniformly.
type panel struct {
	a    *App
	hwnd windows.Handle
	th   theme.Theme
	dpi  int
	w, h int

	fMain, fBold, fSmall uintptr
	hMain, hSmall        int

	btns     []btn
	hover    int
	mx, my   int  // last mouse position (client), -1 when outside
	tracking bool // a WM_MOUSELEAVE is pending
	title    string
	closed   bool
	redraws  int // refresh calls (the selftest checks clicks repaint)

	dragging bool // the mouse is captured and moves go to onDrag
	ticking  bool // the animation timer runs

	draw        func(p *panel, dc uintptr)
	onBtn       func(id int)
	onClick     func(x, y int)
	onWheel     func(delta int) // wheel movement in 1/120 notches, positive = down
	onTick      func() bool     // animation frame; true while more frames are needed
	onDrag      func(x, y int)  // mouse moved while captured (see capture)
	onKey       func(vk int) bool
	onDrop      func(paths []string) // files or folders dropped on the window
	onTint      uint32               // fill of the next toggled-on button; 0 uses the selection colour
	onHoverMove func()               // repaint on every mouse move (the owner draws row hover)
	onClose     func()
}

func (p *panel) s(v int) int { return Scale(v, p.dpi) }

// newPanel creates and shows a panel of logical size lw x lh on the monitor
// holding the cursor.
func (a *App) newPanel(title string, lw, lh int) *panel {
	p := &panel{a: a, title: title, hover: -1, mx: -1, my: -1}
	work, dpi := monitorAtCursor()
	p.dpi = dpi
	p.th = a.theme()
	p.w, p.h = p.s(lw), p.s(lh)
	if maxH := int(work.Bottom-work.Top) * 9 / 10; p.h > maxH {
		p.h = maxH
	}
	px := func(pt int) int { return (pt*dpi + 36) / 72 }
	p.fMain = makeFont(p.th.Font, px(p.th.FontSize), false)
	p.fBold = makeFont(p.th.Font, px(p.th.FontSize), true)
	p.fSmall = makeFont(p.th.Font, px(max(p.th.FontSize-2, 7)), false)
	dc := screenDC()
	p.hMain = fontHeight(dc, p.fMain)
	p.hSmall = fontHeight(dc, p.fSmall)
	pReleaseDC.Call(0, dc)
	x := int(work.Left) + (int(work.Right-work.Left)-p.w)/2
	y := int(work.Top) + (int(work.Bottom-work.Top)-p.h)/2
	p.hwnd = createWindow(classPanel, title, wsPopup, wsExToolWindow|wsExTopmost|wsExAcceptFiles, x, y, p.w, p.h, p)
	a.panels = append(a.panels, p)
	return p
}

func (p *panel) show() {
	if p.a.noShow {
		return
	}
	pShowWindow.Call(uintptr(p.hwnd), swShow)
	forceForeground(p.hwnd)
	invalidate(p.hwnd)
}

func (p *panel) close() {
	if p.closed {
		return
	}
	p.closed = true
	a := p.a
	for i, q := range a.panels {
		if q == p {
			a.panels = append(a.panels[:i], a.panels[i+1:]...)
			break
		}
	}
	for _, f := range []uintptr{p.fMain, p.fBold, p.fSmall} {
		pDeleteObject.Call(f)
	}
	pDestroyWindow.Call(uintptr(p.hwnd))
	if p.onClose != nil {
		p.onClose()
	}
}

func (p *panel) titleH() int { return p.hMain + p.s(18) }

func (p *panel) closeRect() Rect {
	t := p.titleH()
	return Rect{p.w - t, 0, p.w, t}
}

func (p *panel) wndProc(hwnd windows.Handle, m uint32, wp, lp uintptr) uintptr {
	switch m {
	case wmPaint:
		p.paint()
		return 0
	case wmEraseBkgnd:
		return 1
	case 0x0084: // WM_NCHITTEST: drag by the title bar
		var wr rect
		pGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&wr)))
		x, y := signedX(lp)-int(wr.Left), signedY(lp)-int(wr.Top)
		if y >= 0 && y < p.titleH() && !p.closeRect().Contains(x, y) {
			return 2 // HTCAPTION
		}
		return 1 // HTCLIENT
	case wmMouseMove:
		if p.dragging && p.onDrag != nil {
			p.onDrag(signedX(lp), signedY(lp))
			invalidate(hwnd)
			return 0
		}
		p.mx, p.my = signedX(lp), signedY(lp)
		p.trackLeave()
		i := p.hit(p.mx, p.my)
		if i != p.hover || p.onHoverMove != nil {
			p.hover = i
			invalidate(hwnd)
		}
		return 0
	case wmLButtonDown:
		x, y := signedX(lp), signedY(lp)
		if p.closeRect().Contains(x, y) {
			p.close()
			return 0
		}
		if i := p.hit(x, y); i >= 0 {
			if !p.btns[i].disabled && p.onBtn != nil {
				p.onBtn(p.btns[i].id)
			}
		} else if p.onClick != nil {
			p.onClick(x, y)
		}
		if !p.closed {
			p.refresh() // show the new selection / state right away
		}
		return 0
	case 0x02A3: // WM_MOUSELEAVE
		p.tracking = false
		p.mx, p.my = -1, -1
		p.hover = -1
		invalidate(hwnd)
		return 0
	case wmSetCursor:
		if lp&0xFFFF == 1 {
			cur := uintptr(idcArrow)
			if i := p.hover; i > 0 && i < len(p.btns) && !p.btns[i].disabled {
				cur = idcHand
			}
			c, _, _ := pLoadCursorW.Call(0, cur)
			pSetCursor.Call(c)
			return 1
		}
	case wmLButtonUp:
		p.capture(false)
		return 0
	case wmMouseWheel:
		if p.onWheel != nil {
			p.onWheel(-int(int16(wp >> 16)))
			invalidate(hwnd)
		}
		return 0
	case wmTimer:
		if wp == panelTimer {
			if p.onTick == nil || !p.onTick() {
				p.stopTick()
			}
			invalidate(hwnd)
			return 0
		}
	case wmDropFiles:
		paths := droppedFiles(wp)
		if p.onDrop != nil && len(paths) > 0 {
			p.onDrop(paths)
			if !p.closed {
				p.refresh()
			}
		}
		return 0
	case wmKeyDown:
		if p.onKey != nil && p.onKey(int(wp)) {
			invalidate(hwnd)
			return 0
		}
		if wp == vkEscape {
			p.close()
			return 0
		}
	case wmSysCommand:
		if wp&0xFFF0 == 0xF100 {
			return 0
		}
	}
	return defProc(hwnd, m, wp, lp)
}

const panelTimer = 1

// startTick runs onTick every frame until it returns false.
func (p *panel) startTick() {
	if !p.ticking {
		p.ticking = true
		pSetTimer.Call(uintptr(p.hwnd), panelTimer, smoothTickMS, 0)
	}
}

func (p *panel) stopTick() {
	if p.ticking {
		p.ticking = false
		pKillTimer.Call(uintptr(p.hwnd), panelTimer)
	}
}

// capture routes mouse moves to onDrag until the button is released.
func (p *panel) capture(on bool) {
	if on == p.dragging {
		return
	}
	p.dragging = on
	if on {
		pSetCapture.Call(uintptr(p.hwnd))
	} else {
		pReleaseCapture.Call()
	}
}

func (p *panel) hit(x, y int) int {
	for i := len(p.btns) - 1; i >= 0; i-- {
		if p.btns[i].r.Contains(x, y) {
			return i
		}
	}
	return -1
}

func (p *panel) paint() {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(uintptr(p.hwnd), uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(uintptr(p.hwnd), uintptr(unsafe.Pointer(&ps)))
	dc, _, _ := pCreateCompatibleDC.Call(hdc)
	bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(p.w), uintptr(p.h))
	old, _, _ := pSelectObject.Call(dc, bmp)
	p.render(dc)
	pBitBlt.Call(hdc, 0, 0, uintptr(p.w), uintptr(p.h), dc, 0, 0, srcCopy)
	pSelectObject.Call(dc, old)
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(dc)
}

// render draws the frame, title bar and the owner's content into dc.
func (p *panel) render(dc uintptr) {
	c := p.th.Colors
	p.btns = p.btns[:0]
	pSetBkMode.Call(dc, transparent)
	p.fill(dc, Rect{0, 0, p.w, p.h}, uint32(c.Border))
	b := max(p.s(p.th.BorderWidth), 1)
	p.fill(dc, Rect{b, b, p.w - b, p.h - b}, uint32(c.Background))
	// title bar
	t := p.titleH()
	p.fill(dc, Rect{b, b, p.w - b, t}, uint32(c.InputBg))
	p.text(dc, p.fBold, b+p.s(12), b+(t-b-p.hMain)/2, p.title, uint32(c.Text))
	cr := p.closeRect()
	p.btns = append(p.btns, btn{r: cr, id: -1})
	cross := uint32(c.TextDim)
	if p.hover == 0 {
		p.fill(dc, cr, dangerBg)
		cross = 0xFFFFFF
	}
	drawCross(dc, cr.L+cr.W()/2, cr.T+cr.H()/2, max(p.s(5), 3), max(p.s(1), 1), cross)
	if p.draw != nil {
		p.draw(p, dc)
	}
}

func textWidthFont(dc, f uintptr, s string) int {
	old, _, _ := pSelectObject.Call(dc, f)
	w := textWidth(dc, s)
	pSelectObject.Call(dc, old)
	return w
}

func (p *panel) fill(dc uintptr, r Rect, color uint32) {
	fillRect(dc, rect{int32(r.L), int32(r.T), int32(r.R), int32(r.B)}, color)
}

func (p *panel) text(dc, font uintptr, x, y int, s string, color uint32) {
	pSelectObject.Call(dc, font)
	pSetTextColor.Call(dc, colorref(color))
	drawText(dc, x, y, s)
}

// textFit draws s clipped with an ellipsis to maxW.
func (p *panel) textFit(dc, font uintptr, x, y, maxW int, s string, color uint32) {
	pSelectObject.Call(dc, font)
	pSetTextColor.Call(dc, colorref(color))
	drawText(dc, x, y, fitText(dc, s, maxW))
}

// button draws and registers a button. clip, when non-nil, limits hit-testing
// and drawing to a viewport (for scrolled content).
func (p *panel) button(dc uintptr, r Rect, id int, label string, disabled, primary, on bool) {
	c := p.th.Colors
	idx := len(p.btns)
	hov := p.hover == idx && !disabled
	bg, fg := uint32(c.TagBg), uint32(c.Text)
	switch {
	case disabled:
		fg = uint32(c.TextDim)
	case primary || on:
		bg, fg = uint32(c.SelBg), uint32(c.SelText)
		if on && p.onTint != 0 {
			bg, fg = p.onTint, 0xFFFFFF
		}
		if hov {
			bg = p.hoverBg(bg)
		}
	case hov:
		bg = p.hoverBg(bg)
	}
	p.fill(dc, r, bg)
	p.btns = append(p.btns, btn{r: r, id: id, label: label, disabled: disabled, primary: primary, on: on})
	w := textWidthFont(dc, p.fMain, label)
	p.text(dc, p.fMain, r.L+(r.W()-w)/2, r.T+(r.H()-p.hMain)/2, label, fg)
}

// clip limits drawing to r until the returned restore func is called. Buttons
// registered while clipped are still hit-testable only inside r because the
// owner passes already-clipped rectangles.
func (p *panel) clip(dc uintptr, r Rect) func() {
	st, _, _ := pSaveDC.Call(dc)
	pIntersectClipRect.Call(dc, uintptr(r.L), uintptr(r.T), uintptr(r.R), uintptr(r.B))
	return func() { pRestoreDC.Call(dc, st) }
}

// intersect returns a and b overlapped; ok is false when empty.
func intersect(a, b Rect) (Rect, bool) {
	r := Rect{max(a.L, b.L), max(a.T, b.T), min(a.R, b.R), min(a.B, b.B)}
	return r, r.R > r.L && r.B > r.T
}

// wrapText breaks s into at most maxLines lines no wider than maxW using the
// font currently selected in dc. The last line gets an ellipsis if cut.
func wrapText(dc uintptr, s string, maxW, maxLines int) []string {
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for i, w := range words {
		try := w
		if cur != "" {
			try = cur + " " + w
		}
		if textWidth(dc, try) <= maxW || cur == "" {
			cur = try
			continue
		}
		lines = append(lines, cur)
		cur = w
		if len(lines) == maxLines-1 {
			cur = strings.Join(words[i:], " ")
			break
		}
	}
	if cur != "" {
		lines = append(lines, fitText(dc, cur, maxW))
	}
	return lines
}

func messageBox(owner windows.Handle, title, text string, yesNo bool) bool {
	flags := uintptr(0x40) // MB_ICONINFORMATION
	if yesNo {
		flags = 0x4 | 0x20 // MB_YESNO | MB_ICONQUESTION
	}
	r, _, _ := pMessageBoxW.Call(uintptr(owner), uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(title))), flags|0x00040000)
	return r == 6 || !yesNo
}

// refresh repaints the panel now-ish.
func (p *panel) refresh() {
	p.redraws++
	invalidate(p.hwnd)
}

// mix blends colour a toward b by pct percent (channel order does not matter).
func mix(a, b uint32, pct int) uint32 {
	ch := func(sh uint) uint32 {
		x, y := int(a>>sh&0xFF), int(b>>sh&0xFF)
		return uint32(x+(y-x)*pct/100) << sh
	}
	return ch(0) | ch(8) | ch(16)
}

// hoverBg is the fill of a hovered button or row; themes often give TagBg and
// SelBg the same value, so it is derived instead of read.
func (p *panel) hoverBg(base uint32) uint32 { return mix(base, uint32(p.th.Colors.Text), 14) }

func (p *panel) trackLeave() {
	if p.tracking {
		return
	}
	p.tracking = true
	t := struct {
		Size, Flags uint32
		Wnd         windows.Handle
		Hover       uint32
	}{Size: 16, Flags: 2, Wnd: p.hwnd} // TME_LEAVE
	pTrackMouseEvent.Call(uintptr(unsafe.Pointer(&t)))
}
