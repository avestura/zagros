//go:build windows

package ui

var (
	pCreatePen = gdi32.NewProc("CreatePen")
	pMoveToEx  = gdi32.NewProc("MoveToEx")
	pLineTo    = gdi32.NewProc("LineTo")
)

// drawCross draws an X of the given half-size centred on (cx, cy).
func drawCross(dc uintptr, cx, cy, half, width int, color uint32) {
	pen, _, _ := pCreatePen.Call(0, uintptr(width), colorref(color))
	old, _, _ := pSelectObject.Call(dc, pen)
	pMoveToEx.Call(dc, uintptr(cx-half), uintptr(cy-half), 0)
	pLineTo.Call(dc, uintptr(cx+half+1), uintptr(cy+half+1))
	pMoveToEx.Call(dc, uintptr(cx+half), uintptr(cy-half), 0)
	pLineTo.Call(dc, uintptr(cx-half-1), uintptr(cy+half+1))
	pSelectObject.Call(dc, old)
	pDeleteObject.Call(pen)
}

const dangerBg = 0xC42B1C

// chipWidth is the width of a removable chip for label, capped at maxW.
func (p *panel) chipWidth(dc uintptr, label string, maxW int) int {
	return min(textWidthFont(dc, p.fSmall, label)+p.s(34), maxW)
}

// chip draws a removable tag: the label (ellipsised to fit) and a cross. The
// whole chip is the button; hovering it turns it red to show it is destructive.
func (p *panel) chip(dc uintptr, r Rect, id int, label string, disabled bool) {
	c := p.th.Colors
	hov := p.hover == len(p.btns) && !disabled
	bg, fg, x := uint32(c.TagBg), uint32(c.Text), uint32(c.TextDim)
	if hov {
		bg, fg, x = dangerBg, 0xFFFFFF, 0xFFFFFF
	}
	if disabled {
		fg = uint32(c.TextDim)
	}
	p.fill(dc, r, bg)
	p.btns = append(p.btns, btn{r: r, id: id, label: label, disabled: disabled})
	pad := p.s(10)
	p.textFit(dc, p.fSmall, r.L+pad, r.T+(r.H()-p.hSmall)/2, r.W()-pad-p.s(26), label, fg)
	drawCross(dc, r.R-p.s(14), r.T+r.H()/2, max(p.s(3), 2), max(p.s(1), 1), x)
}
