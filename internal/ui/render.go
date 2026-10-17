//go:build windows

package ui

import (
	"strings"
	"unicode"

	"github.com/avestura/zagros/internal/core"
)

const placeholderText = "Search apps, files and commands"

// fitText truncates s with an ellipsis so it fits in maxW pixels using the
// font currently selected in dc.
func fitText(dc uintptr, s string, maxW int) string {
	if textWidth(dc, s) <= maxW {
		return s
	}
	rs := []rune(s)
	lo, hi := 0, len(rs)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if textWidth(dc, string(rs[:mid])+"…") <= maxW {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(rs[:lo]) + "…"
}

func (a *App) setColor(dc uintptr, c uint32) { pSetTextColor.Call(dc, colorref(c)) }

func (a *App) selectFont(dc, f uintptr) { pSelectObject.Call(dc, f) }

// render draws one full frame of the launcher into dc (w x h pixels).
func (a *App) render(dc uintptr, w, h int) {
	c := a.th.Colors
	m := a.m
	pSetBkMode.Call(dc, transparent)

	fillRect(dc, rect{0, 0, int32(w), int32(h)}, uint32(c.Border))
	b := m.Border
	fillRect(dc, rect{int32(b), int32(b), int32(w - b), int32(h - b)}, uint32(c.Background))

	ir := m.InputRect()
	if !a.skipInput { // the input band is painted on its own path (paintInputNow)
		fillRect(dc, rect{int32(ir.L), int32(ir.T), int32(ir.R), int32(ir.B)}, uint32(c.InputBg))
		a.drawInput(dc, ir)
	}

	n := len(a.items)
	vis := m.VisibleRows(n)
	if vis == 0 {
		if a.status != "" { // e.g. "Indexing apps..." before any result exists
			fillRect(dc, rect{int32(ir.L), int32(ir.B), int32(ir.R), int32(ir.B + max(1, Scale(1, a.dpi)))}, uint32(c.Border))
			a.drawStatus(dc, m.StatusRect(0))
		}
		return
	}
	// separator under the input
	fillRect(dc, rect{int32(ir.L), int32(ir.B), int32(ir.R), int32(ir.B + max(1, Scale(1, a.dpi)))}, uint32(c.Border))
	a.frameIcons = a.frameIcons[:0]
	for i := 0; i < vis; i++ {
		abs := a.sc.Top + i
		if abs >= n {
			break
		}
		a.drawRow(dc, m.RowRect(i), a.items[abs], abs == a.sc.Sel)
	}
	if a.loader != nil {
		a.loader.q.SetVisible(a.frameIcons) // rows that scrolled or typed away are not worth loading
	}
	a.drawStatus(dc, m.StatusRect(n))
}

func (a *App) drawInput(dc uintptr, ir Rect) {
	c := a.th.Colors
	pad := a.m.Pad
	state, _, _ := pSaveDC.Call(dc)
	defer pRestoreDC.Call(dc, state)
	viewL, viewR := ir.L+pad, ir.R-pad
	pIntersectClipRect.Call(dc, uintptr(viewL), uintptr(ir.T), uintptr(viewR), uintptr(ir.B))
	a.selectFont(dc, a.fonts.input)
	th := a.fonts.hInput
	ty := ir.T + (ir.H()-th)/2
	text := a.ed.Text()
	if text == "" {
		a.setColor(dc, uint32(c.Placeholder))
		drawText(dc, viewL, ty, placeholderText)
		if a.caretOn {
			fillRect(dc, rect{int32(viewL), int32(ty), int32(viewL + max(1, Scale(1, a.dpi))), int32(ty + th)}, uint32(c.InputText))
		}
		return
	}
	rs := []rune(text)
	caretX := textWidth(dc, string(rs[:a.ed.Caret()]))
	a.hscroll = HScroll(caretX, textWidth(dc, text), viewR-viewL, a.hscroll)
	x := viewL - a.hscroll
	if a.ed.HasSelection() {
		lo, hi := a.ed.Selection()
		x1 := x + textWidth(dc, string(rs[:lo]))
		x2 := x + textWidth(dc, string(rs[:hi]))
		bg := uint32(c.SelBg)
		if bg == uint32(c.InputBg) {
			bg = uint32(c.Accent)
		}
		fillRect(dc, rect{int32(x1), int32(ty - 2), int32(x2), int32(ty + th + 2)}, bg)
	}
	a.setColor(dc, uint32(c.InputText))
	drawText(dc, x, ty, text)
	if a.caretOn {
		cx := x + caretX
		fillRect(dc, rect{int32(cx), int32(ty), int32(cx + max(1, Scale(1, a.dpi))), int32(ty + th)}, uint32(c.InputText))
	}
}

func (a *App) iconPath(it core.Item) string {
	p := it.Icon
	if p == "" && (it.Action.Kind == "open" || it.Action.Kind == "run") {
		p = it.Action.Target
	}
	if strings.Contains(p, "://") || !strings.ContainsAny(p, `\/.:`) {
		return ""
	}
	return p
}

func (a *App) drawRow(dc uintptr, r Rect, it core.Item, selected bool) {
	c := a.th.Colors
	dpi := a.dpi
	pad := a.m.Pad
	if selected {
		fillRect(dc, rect{int32(r.L), int32(r.T), int32(r.R), int32(r.B)}, uint32(c.SelBg))
	}
	titleC, dimC := uint32(c.Text), uint32(c.TextDim)
	if selected {
		titleC, dimC = uint32(c.SelText), uint32(c.SelTextDim)
	}
	x := r.L + pad
	if a.th.ShowIcons {
		sz := r.H() * 5 / 9
		iy := r.T + (r.H()-sz)/2
		a.drawIcon(dc, it, x, iy, sz, selected)
		x += sz + pad*3/4
	}
	right := r.R - pad
	if a.th.ShowTags && it.Category != "" {
		a.selectFont(dc, a.fonts.small)
		px := Scale(6, dpi)
		tw := textWidth(dc, it.Category) + 2*px
		th := a.fonts.hSmall + Scale(4, dpi)
		tx := right - tw
		ty := r.T + (r.H()-th)/2
		tagBg, tagFg := uint32(c.TagBg), uint32(c.Tag)
		fillRect(dc, rect{int32(tx), int32(ty), int32(tx + tw), int32(ty + th)}, tagBg)
		a.setColor(dc, tagFg)
		drawText(dc, tx+px, ty+(th-a.fonts.hSmall)/2, it.Category)
		right = tx - pad
	}
	avail := right - x
	if avail <= 0 {
		return
	}
	gap := Scale(2, dpi)
	total := a.fonts.hMain
	if it.Subtitle != "" {
		total += gap + a.fonts.hSmall
	}
	ty := r.T + (r.H()-total)/2

	// title with highlight
	a.selectFont(dc, a.fonts.main)
	title := fitText(dc, it.Title, avail)
	idx := a.opt.Highlight(a.query, title)
	cx := x
	for _, run := range SplitRuns(title, idx) {
		if run.Hit {
			a.selectFont(dc, a.fonts.bold)
			if selected {
				a.setColor(dc, uint32(c.SelText))
			} else {
				a.setColor(dc, uint32(c.Accent))
			}
		} else {
			a.selectFont(dc, a.fonts.main)
			a.setColor(dc, titleC)
		}
		drawText(dc, cx, ty, run.Text)
		cx += textWidth(dc, run.Text)
	}
	if it.Subtitle != "" {
		a.selectFont(dc, a.fonts.small)
		a.setColor(dc, dimC)
		drawText(dc, x, ty+a.fonts.hMain+gap, fitText(dc, it.Subtitle, avail))
	}
}

func (a *App) drawIcon(dc uintptr, it core.Item, x, y, sz int, selected bool) {
	if p := a.iconPath(it); p != "" {
		k := iconKey{Path: p, Size: sz}
		a.frameIcons = append(a.frameIcons, k)
		if bmp, ok := a.icons.Get(k); ok {
			if bmp != 0 {
				a.blitIcon(dc, bmp, x, y, sz)
				return
			}
		} else {
			a.ensureLoader().Request(k) // never loads on the UI thread
		}
	}
	// letter tile fallback
	c := a.th.Colors
	bg, fg := uint32(c.TagBg), uint32(c.Tag)
	if selected {
		bg, fg = uint32(c.Accent), uint32(c.Background)
	}
	fillRect(dc, rect{int32(x), int32(y), int32(x + sz), int32(y + sz)}, bg)
	letter := "?"
	for _, r := range it.Title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			letter = strings.ToUpper(string(r))
			break
		}
	}
	a.selectFont(dc, a.fonts.bold)
	a.setColor(dc, fg)
	lw := textWidth(dc, letter)
	drawText(dc, x+(sz-lw)/2, y+(sz-a.fonts.hMain)/2, letter)
}

func (a *App) drawStatus(dc uintptr, r Rect) {
	c := a.th.Colors
	pad := a.m.Pad
	a.selectFont(dc, a.fonts.small)
	a.setColor(dc, uint32(c.TextDim))
	ty := r.T + (r.H()-a.fonts.hSmall)/2
	hint := "Enter run   Tab actions   Esc close"
	if it, ok := a.selected(); ok && it.Complete != "" {
		hint = "Enter run   Tab complete   Esc close"
	}
	hw := textWidth(dc, hint)
	left := a.status
	if left == "" {
		if it, ok := a.selected(); ok {
			left = it.Action.Target
		}
	}
	avail := r.W() - 3*pad - hw
	if avail > 0 {
		drawText(dc, r.L+pad, ty, fitText(dc, left, avail))
	}
	drawText(dc, r.R-pad-hw, ty, hint)
}
