//go:build windows

package ui

// Flat Boxy-style controls shared by the windows drawn with panel: hairline
// outlines, no bevels, accent only for the active/selected state.

// hairline draws a 1px rectangle outline.
func (p *panel) hairline(dc uintptr, r Rect, color uint32) {
	b := max(p.s(1), 1)
	p.fill(dc, Rect{r.L, r.T, r.R, r.T + b}, color)
	p.fill(dc, Rect{r.L, r.B - b, r.R, r.B}, color)
	p.fill(dc, Rect{r.L, r.T, r.L + b, r.B}, color)
	p.fill(dc, Rect{r.R - b, r.T, r.R, r.B}, color)
}

// rule draws a horizontal (h) or vertical hairline across r.
func (p *panel) rule(dc uintptr, r Rect) { p.fill(dc, r, uint32(p.th.Colors.Border)) }

// tab draws a text tab: dim when idle, full text with an accent underline when
// active. The hovered tab gets a faint fill.
func (p *panel) tab(dc uintptr, r Rect, id int, label string, disabled, active bool) {
	c := p.th.Colors
	hov := p.hover == len(p.btns) && !disabled
	fg := uint32(c.TextDim)
	if active {
		fg = uint32(c.Text)
	}
	if hov {
		p.fill(dc, r, p.hoverBg(uint32(c.Background)))
		fg = uint32(c.Text)
	}
	if active {
		u := max(p.s(2), 2)
		p.fill(dc, Rect{r.L, r.B - u, r.R, r.B}, uint32(c.Accent))
	}
	p.btns = append(p.btns, btn{r: r, id: id, label: label, disabled: disabled, on: active})
	f := p.fMain
	if active {
		f = p.fBold
	}
	w := textWidthFont(dc, f, label)
	p.text(dc, f, r.L+(r.W()-w)/2, r.T+(r.H()-p.hMain)/2, label, fg)
}

// ghost is a hairline button on the window background. danger turns it red on
// hover. primary fills it with the selection colour (one per view).
func (p *panel) ghost(dc uintptr, r Rect, id int, label string, disabled, primary, danger bool) {
	c := p.th.Colors
	hov := p.hover == len(p.btns) && !disabled
	fg, edge := uint32(c.Text), uint32(c.Border)
	switch {
	case disabled:
		fg = uint32(c.TextDim)
	case primary:
		bg := uint32(c.SelBg)
		if hov {
			bg = p.hoverBg(bg)
		}
		p.fill(dc, r, bg)
		edge, fg = bg, uint32(c.SelText)
	case hov && danger:
		p.fill(dc, r, dangerBg)
		edge, fg = dangerBg, 0xFFFFFF
	case hov:
		p.fill(dc, r, p.hoverBg(uint32(c.Background)))
		edge = uint32(c.TextDim)
	}
	if !primary {
		p.hairline(dc, r, edge)
	}
	p.btns = append(p.btns, btn{r: r, id: id, label: label, disabled: disabled, primary: primary})
	w := textWidthFont(dc, p.fMain, label)
	p.text(dc, p.fMain, r.L+(r.W()-w)/2, r.T+(r.H()-p.hMain)/2, label, fg)
	p.focusRing(dc, r, id, disabled)
}

// focusRing outlines the control that has keyboard focus.
func (p *panel) focusRing(dc uintptr, r Rect, id int, disabled bool) {
	if id != p.focusID || disabled {
		return
	}
	a, w := uint32(p.th.Colors.Accent), max(p.s(2), 2)
	p.fill(dc, Rect{r.L, r.T, r.R, r.T + w}, a)
	p.fill(dc, Rect{r.L, r.B - w, r.R, r.B}, a)
	p.fill(dc, Rect{r.L, r.T, r.L + w, r.B}, a)
	p.fill(dc, Rect{r.R - w, r.T, r.R, r.B}, a)
}

// focusable lists the ids of the enabled controls that take keyboard focus, in
// drawing order. skip excludes ids (tabs have their own keys).
func (p *panel) focusable(skip func(id int) bool) []int {
	var out []int
	for _, b := range p.btns {
		if b.id > 0 && !b.disabled && !(skip != nil && skip(b.id)) {
			out = append(out, b.id)
		}
	}
	return out
}

// moveFocus moves focus to the next (or previous) control and returns its
// button, wrapping around.
func (p *panel) moveFocus(skip func(id int) bool, back bool) (btn, bool) {
	ids := p.focusable(skip)
	if len(ids) == 0 {
		return btn{}, false
	}
	at := -1
	for i, id := range ids {
		if id == p.focusID {
			at = i
		}
	}
	switch {
	case at < 0 && back:
		at = len(ids) - 1
	case at < 0:
		at = 0
	case back:
		at = (at - 1 + len(ids)) % len(ids)
	default:
		at = (at + 1) % len(ids)
	}
	p.focusID = ids[at]
	for _, b := range p.btns {
		if b.id == p.focusID {
			return b, true
		}
	}
	return btn{}, false
}

// ghostWidth is the width a ghost button needs for label.
func (p *panel) ghostWidth(dc uintptr, label string) int {
	return textWidthFont(dc, p.fMain, label) + p.s(24)
}

// toggle draws a switch (track and knob) with the state as accent fill when on.
func (p *panel) toggle(dc uintptr, r Rect, id int, on, disabled bool) {
	c := p.th.Colors
	hov := p.hover == len(p.btns) && !disabled
	track := uint32(c.TagBg)
	edge := uint32(c.Border)
	knob := uint32(c.TextDim)
	if on {
		track, edge, knob = uint32(c.Accent), uint32(c.Accent), 0xFFFFFF
	}
	if hov && !on {
		edge = uint32(c.TextDim)
	}
	if disabled {
		track = uint32(c.TagBg)
	}
	p.fill(dc, r, track)
	p.hairline(dc, r, edge)
	in := p.s(3)
	k := r.H() - 2*in
	x := r.L + in
	if on {
		x = r.R - in - k
	}
	p.fill(dc, Rect{x, r.T + in, x + k, r.T + in + k}, knob)
	p.btns = append(p.btns, btn{r: r, id: id, disabled: disabled, on: on})
	p.focusRing(dc, r, id, disabled)
}

// input draws a text box. focused shows the caret and an accent edge.
func (p *panel) input(dc uintptr, r Rect, text, placeholder string, focused bool, mono bool) {
	c := p.th.Colors
	p.fill(dc, r, uint32(c.InputBg))
	edge := uint32(c.Border)
	if focused {
		edge = uint32(c.Accent)
	}
	p.hairline(dc, r, edge)
	f, h := p.fMain, p.hMain
	if mono {
		f, h = p.fMono, p.hMono
	}
	pad := p.s(8)
	y := r.T + (r.H()-h)/2
	maxW := r.W() - 2*pad
	if text == "" && !focused {
		p.textFit(dc, f, r.L+pad, y, maxW, placeholder, uint32(c.Placeholder))
		return
	}
	shown := text
	for textWidthFont(dc, f, shown) > maxW-p.s(2) && len(shown) > 0 { // keep the end visible
		shown = shown[1:]
	}
	p.text(dc, f, r.L+pad, y, shown, uint32(c.InputText))
	if focused {
		cx := r.L + pad + textWidthFont(dc, f, shown) + p.s(1)
		p.fill(dc, Rect{cx, y, cx + max(p.s(1), 1), y + h}, uint32(c.Accent))
	}
}

// tag draws a small outlined label (keywords, state) and returns its width.
func (p *panel) tag(dc uintptr, x, y int, label string, mono bool) int {
	c := p.th.Colors
	f, h := p.fSmall, p.hSmall
	if mono {
		f, h = p.fMono, p.hMono
	}
	w := textWidthFont(dc, f, label) + p.s(16)
	r := Rect{x, y, x + w, y + h + p.s(8)}
	p.fill(dc, r, uint32(c.TagBg))
	p.text(dc, f, x+p.s(8), y+p.s(4), label, uint32(c.Text))
	return w
}

// wrap draws s wrapped to maxW and returns the height used.
func (p *panel) wrap(dc, font uintptr, h, x, y, maxW, maxLines int, s string, color uint32) int {
	pSelectObject.Call(dc, font)
	lines := wrapText(dc, s, maxW, maxLines)
	lh := h + p.s(3)
	for i, l := range lines {
		p.text(dc, font, x, y+i*lh, l, color)
	}
	return len(lines) * lh
}
