//go:build windows

package ui

import (
	"fmt"
	"math"
	"strings"
)

// dpane lays out the scrolling body of the detail pane. Buttons that scroll
// out of the viewport are not registered, and partly visible ones are
// hit-tested only inside it.
type dpane struct {
	p    *panel
	dc   uintptr
	vp   Rect
	x, w int
	y    int
}

func (d *dpane) vis(r Rect) bool { _, ok := intersect(r, d.vp); return ok }

func (d *dpane) fix(n int) {
	for i := n; i < len(d.p.btns); i++ {
		d.p.btns[i].abs = d.p.btns[i].r
		d.p.btns[i].r, _ = intersect(d.p.btns[i].r, d.vp)
	}
}

// hidden registers a control that is scrolled out of view so keyboard focus
// can still reach it (and scroll it into view); it has no hit area.
func (d *dpane) hidden(r Rect, id int, label string, disabled bool) {
	d.p.btns = append(d.p.btns, btn{id: id, label: label, disabled: disabled, abs: r})
}

func (d *dpane) ghost(r Rect, id int, label string, disabled, primary, danger bool) {
	if !d.vis(r) {
		d.hidden(r, id, label, disabled)
		return
	}
	n := len(d.p.btns)
	d.p.ghost(d.dc, r, id, label, disabled, primary, danger)
	d.fix(n)
}

func (d *dpane) toggle(r Rect, id int, on, disabled bool) {
	if !d.vis(r) {
		d.hidden(r, id, "", disabled)
		return
	}
	n := len(d.p.btns)
	d.p.toggle(d.dc, r, id, on, disabled)
	d.fix(n)
}

func (d *dpane) gap(v int) { d.y += d.p.s(v) }

// heading draws a small section label.
func (d *dpane) heading(s string) {
	p := d.p
	d.gap(12)
	p.text(d.dc, p.fSmall, d.x, d.y, strings.ToUpper(s), uint32(p.th.Colors.TextDim))
	d.y += p.hSmall + p.s(8)
}

// para draws wrapped body text.
func (d *dpane) para(s string, color uint32) {
	p := d.p
	d.y += p.wrap(d.dc, p.fMain, p.hMain, d.x, d.y, d.w, 12, s, color) + p.s(4)
}

func (d *dpane) note(s string) {
	p := d.p
	d.y += p.wrap(d.dc, p.fSmall, p.hSmall, d.x, d.y, d.w, 6, s, uint32(p.th.Colors.TextDim)) + p.s(4)
}

func (m *manager) draw(p *panel, dc uintptr) {
	pad := p.s(16)
	b := max(p.s(p.th.BorderWidth), 1)
	t := p.titleH()

	// tab strip, filter on the right
	th := p.s(40)
	stripB := t + th
	x := b + p.s(8)
	labels := [3]string{fmt.Sprintf("Installed (%d)", len(m.inst)), "Marketplace", fmt.Sprintf("Themes (%d)", len(m.themes))}
	ids := [3]int{idTabInstalled, idTabMarket, idTabThemes}
	for i, l := range labels {
		w := textWidthFont(dc, p.fBold, l) + p.s(32)
		p.tab(dc, Rect{x, t, x + w, stripB}, ids[i], l, !m.tabOK(i), m.tab == i)
		x += w
	}
	iw, ih := p.s(260), p.s(28)
	p.input(dc, Rect{p.w - b - pad - iw, t + (th-ih)/2, p.w - b - pad, t + (th-ih)/2 + ih},
		m.filter, "Type to filter, Esc clears", m.filter != "", false)
	p.rule(dc, Rect{b, stripB, p.w - b, stripB + 1})

	footH := p.s(52)
	footT := p.h - b - footH
	bodyT := stripB + 1
	p.rule(dc, Rect{b, footT, p.w - b, footT + 1})

	listW := min(p.s(330), (p.w-2*b)*2/5)
	m.list = Rect{b, bodyT, b + listW, footT}
	m.dpane = Rect{m.list.R + 1, bodyT, p.w - b, footT}
	p.rule(dc, Rect{m.list.R, bodyT, m.list.R + 1, footT})
	m.rowH = max(p.hMain+p.hSmall+p.s(22), p.s(56))

	m.drawList(p, dc, pad)
	m.drawDetail(p, dc, pad)
	m.drawFooter(p, dc, pad, footT+1, p.h-b)
}

func (m *manager) drawList(p *panel, dc uintptr, pad int) {
	c := p.th.Colors
	ix := m.idx(m.tab)
	n := len(ix)
	s := m.sc()
	vis := max(m.list.H()/m.rowH, 1)
	s.SetCount(n, vis)
	sm := m.pxs()
	sm.clamp(n*m.rowH, m.list.H())
	rowR := m.list.R
	m.bar = Rect{}
	if n*m.rowH > m.list.H() {
		bw := p.s(barW)
		m.bar = Rect{m.list.R - bw, m.list.T, m.list.R, m.list.B}
		rowR = m.bar.L
	}

	empty := ""
	switch {
	case m.filter != "" && n == 0:
		empty = fmt.Sprintf("Nothing matches %q. Esc clears the filter.", m.filter)
	case m.tab == 0 && n == 0:
		empty = "No extensions installed yet."
	case m.tab == 2 && n == 0:
		empty = "No themes found."
	case m.tab == 1 && m.mkLoading:
		empty = "Loading marketplace..."
	case m.tab == 1 && m.mkErr != "":
		empty = "Could not load the marketplace: " + m.mkErr
	case m.tab == 1 && n == 0:
		empty = "The marketplace is empty."
	}
	if empty != "" {
		p.wrap(dc, p.fMain, p.hMain, m.list.L+pad, m.list.T+pad, m.list.W()-2*pad, 5, empty, uint32(c.TextDim))
	}

	restore := p.clip(dc, m.list)
	pos := int(math.Round(sm.pos))
	first, off := pos/m.rowH, pos%m.rowH
	for i := 0; first+i < n && i*m.rowH-off < m.list.H(); i++ {
		abs := first + i
		src := ix[abs]
		y0 := m.list.T + i*m.rowH - off
		r := Rect{m.list.L, y0, rowR, y0 + m.rowH}
		fg, dim := uint32(c.Text), uint32(c.TextDim)
		switch {
		case abs == s.Sel:
			p.fill(dc, r, uint32(c.SelBg))
			fg, dim = uint32(c.SelText), uint32(c.SelTextDim)
		case r.Contains(p.mx, p.my):
			p.fill(dc, r, p.hoverBg(uint32(c.Background)))
		}
		var l1, l2, tag string
		mono := false
		tagC := dim
		switch m.tab {
		case 0:
			e := m.inst[src]
			l1, l2 = e.Name, e.Description
			if l2 == "" {
				l2 = orDash(e.Publisher)
			}
			switch {
			case e.Err != "":
				tag, tagC = "error", dangerBg
			case !e.Enabled:
				tag = "off"
				if abs != s.Sel {
					fg = dim
				}
			case len(e.Keywords) > 0:
				tag, mono = e.Keywords[0], true
			}
		case 2:
			e := m.themes[src]
			l1, l2 = e.Name, e.Source
			if e.Author != "" {
				l2 += "  -  " + e.Author
			}
			if e.Current {
				tag, tagC = "in use", uint32(c.Accent)
				if abs == s.Sel {
					tagC = fg
				}
			}
		default:
			e := m.mk[src]
			l1, l2, tag = e.Name, e.Description, kindTag(e.Kind)
		}
		x := r.L + pad
		right := r.R - pad
		tf, tw := p.fSmall, 0
		if mono {
			tf = p.fMono
		}
		if tag != "" {
			tw = textWidthFont(dc, tf, tag)
			p.text(dc, tf, right-tw, r.T+p.s(10), tag, tagC)
			tw += p.s(8)
		}
		p.textFit(dc, p.fBold, x, r.T+p.s(8), right-x-tw, l1, fg)
		p.textFit(dc, p.fSmall, x, r.T+p.s(8)+p.hMain+p.s(2), right-x, l2, dim)
	}
	restore()
	if m.bar.W() > 0 {
		p.fill(dc, m.bar, uint32(c.Background))
		tc := uint32(c.TagBg)
		if m.p.dragging {
			tc = uint32(c.TextDim)
		}
		p.fill(dc, m.thumb(), tc)
	}
}

func (m *manager) drawFooter(p *panel, dc uintptr, pad, top, bottom int) {
	c := p.th.Colors
	bh := p.s(32)
	fy := top + (bottom-top-bh)/2
	right := p.w - pad
	add := func(id int, label string, disabled bool) {
		w := p.ghostWidth(dc, label)
		p.ghost(dc, Rect{right - w, fy, right, fy + bh}, id, label, disabled, false, false)
		right -= w + p.s(8)
	}
	switch m.tab {
	case 0:
		add(idRefreshInst, "Refresh", false)
		add(idInstFolder, "Install folder...", m.busy)
		add(idInstFile, "Install file...", m.busy)
	case 1:
		add(idRefreshMkt, "Refresh", false)
	default:
		add(idRefreshThm, "Refresh", false)
	}
	msg, col := m.status, uint32(c.Text)
	if msg == "" {
		msg, col = "Type to filter  -  Enter "+map[int]string{0: "toggles", 1: "installs", 2: "uses"}[m.tab]+"  -  Tab focus, Space press  -  Ctrl+Tab tabs", uint32(c.TextDim)
	}
	p.textFit(dc, p.fSmall, pad, fy+(bh-p.hSmall)/2, right-pad, msg, col)
}

// detailHead returns the identity and header lines of the selected item.
func (m *manager) drawDetail(p *panel, dc uintptr, pad int) {
	c := p.th.Colors
	pane := m.dpane
	p.fill(dc, pane, uint32(c.Background))
	hh := p.s(72)
	m.dview = Rect{pane.L, pane.T + hh + 1, pane.R, pane.B}
	p.rule(dc, Rect{pane.L, pane.T + hh, pane.R, pane.T + hh + 1})
	x, w := pane.L+pad, pane.W()-2*pad

	var id, title, ver, sub string
	var inst ExtInfo
	var mk MarketEntry
	var thm ThemeInfo
	have := false
	switch m.tab {
	case 0:
		if inst, have = m.selInst(); have {
			id, title, ver, sub = inst.ID, inst.Name, inst.Version, orDash(inst.Publisher)
		}
	case 1:
		if mk, have = m.selMarket(); have {
			id, title, ver, sub = mk.Ref, mk.Name, mk.Version, kindTag(mk.Kind)+"  -  "+orDash(mk.Author)
		}
	default:
		if thm, have = m.selTheme(); have {
			id, title, sub = thm.Name, thm.Name, thm.Source
			if thm.Author != "" {
				sub += "  -  " + thm.Author
			}
		}
	}
	if key := fmt.Sprint(m.tab, ":", id); key != m.dkey {
		m.dkey, m.px[3], m.editKey, m.editErr = key, smooth{}, "", ""
	}
	if !have {
		m.detH = 0
		msg := "Select an item to see what it does and how to use it."
		switch {
		case m.tab == 0 && len(m.inst) == 0:
			msg = "Nothing installed yet. Open the Marketplace tab, use Install file... or Install folder... below, or drop a .zagros file or a project folder on this window. Every extension you install is described here: what to type, what it can access, and its settings."
		case m.filter != "":
			msg = "No item matches the filter."
		}
		p.wrap(dc, p.fMain, p.hMain, x, pane.T+pad, w, 8, msg, uint32(c.TextDim))
		return
	}

	// header: name, version, source, and the one control that matters
	ctlW := 0
	hy := pane.T + p.s(14)
	switch m.tab {
	case 0:
		lbl := "Disabled"
		if inst.Enabled {
			lbl = "Enabled"
		}
		sw := Rect{pane.R - pad - p.s(36), hy + p.s(2), pane.R - pad, hy + p.s(20)}
		p.toggle(dc, sw, idToggle, inst.Enabled, m.busy)
		lw := textWidthFont(dc, p.fMain, lbl)
		p.text(dc, p.fMain, sw.L-p.s(8)-lw, hy+p.s(2)+(sw.H()-p.hMain)/2, lbl, uint32(c.Text))
		ctlW = sw.W() + lw + p.s(16)
	case 1:
		lbl := "Install"
		ctlW = p.ghostWidth(dc, lbl)
		p.ghost(dc, Rect{pane.R - pad - ctlW, hy, pane.R - pad, hy + p.s(32)}, idInstall, lbl, m.busy, true, false)
	default:
		lbl := "Use this theme"
		if thm.Current {
			lbl = "In use"
		}
		ctlW = p.ghostWidth(dc, lbl)
		p.ghost(dc, Rect{pane.R - pad - ctlW, hy, pane.R - pad, hy + p.s(32)}, idUseTheme, lbl, m.busy || thm.Current, !thm.Current, false)
	}
	name := title
	if ver != "" {
		name += "  v" + ver
	}
	p.textFit(dc, p.fBold, x, hy, w-ctlW-p.s(12), name, uint32(c.Text))
	p.textFit(dc, p.fSmall, x, hy+p.hMain+p.s(4), w-ctlW-p.s(12), sub, uint32(c.TextDim))

	// scrolling body
	scroll := int(math.Round(m.px[3].pos))
	restore := p.clip(dc, m.dview)
	d := &dpane{p: p, dc: dc, vp: m.dview, x: x, w: w, y: m.dview.T + pad - scroll}
	switch m.tab {
	case 0:
		m.bodyInstalled(d, inst)
	case 1:
		m.bodyMarket(d, mk)
	default:
		m.bodyTheme(d, thm)
	}
	restore()
	m.detH = d.y + scroll - m.dview.T + pad
	m.px[3].clamp(m.detH, m.dview.H())
	if m.detH > m.dview.H() {
		bw := p.s(4)
		th := max(m.dview.H()*m.dview.H()/m.detH, p.s(24))
		span := m.detH - m.dview.H()
		top := m.dview.T + int(float64(m.dview.H()-th)*m.px[3].pos/float64(span))
		p.fill(dc, Rect{m.dview.R - bw - p.s(2), top, m.dview.R - p.s(2), top + th}, uint32(c.TagBg))
	}
}

func (m *manager) bodyInstalled(d *dpane, e ExtInfo) {
	p, dc := d.p, d.dc
	c := p.th.Colors
	dim := uint32(c.TextDim)
	if e.Err != "" {
		d.para("Could not start: "+e.Err, dangerBg)
	}
	if e.Description != "" {
		d.para(e.Description, uint32(c.Text))
	}
	bh := p.s(28)
	ax := d.x
	for _, a := range []struct {
		id    int
		label string
		on    bool
	}{{idOpenDir, "Open folder", e.Dir != ""}, {idHomepage, "Homepage", strings.HasPrefix(e.Homepage, "https://")}} {
		if !a.on {
			continue
		}
		w := p.ghostWidth(dc, a.label)
		d.ghost(Rect{ax, d.y, ax + w, d.y + bh}, a.id, a.label, false, false, false)
		ax += w + p.s(8)
	}
	if ax > d.x {
		d.y += bh
	}

	// how to use
	d.heading("How to use")
	if e.Usage != "" {
		d.para(e.Usage, uint32(c.Text))
	}
	if len(e.Keywords) > 0 {
		lbl := "Keyword"
		if len(e.Keywords) > 1 {
			lbl = "Keywords"
		}
		p.text(dc, p.fSmall, d.x, d.y+p.s(5), lbl+": type", dim)
		tx := d.x + textWidthFont(dc, p.fSmall, lbl+": type") + p.s(8)
		for _, k := range e.Keywords {
			tx += p.tag(dc, tx, d.y, k, true) + p.s(6)
		}
		p.text(dc, p.fSmall, tx, d.y+p.s(5), "then a space and your text", dim)
		d.y += p.hMono + p.s(16)
	} else if e.Usage == "" {
		d.para("No keyword. Its results appear in the launcher's normal search.", dim)
	}
	if tr := tryQueries(e); len(tr) > 0 {
		d.gap(4)
		p.text(dc, p.fSmall, d.x, d.y, "Click Try to type it into the launcher", dim)
		d.y += p.hSmall + p.s(6)
		for i, q := range tr {
			rh := p.s(32)
			row := Rect{d.x, d.y, d.x + d.w, d.y + rh}
			p.hairline(dc, row, uint32(c.Border))
			p.textFit(dc, p.fMono, row.L+p.s(10), row.T+(rh-p.hMono)/2, row.W()-p.s(100), q, uint32(c.Text))
			bw := p.ghostWidth(dc, "Try")
			d.ghost(Rect{row.R - bw - p.s(2), row.T + p.s(2), row.R - p.s(2), row.B - p.s(2)}, idTryBase+i, "Try", false, false, false)
			d.y += rh + p.s(4)
		}
	}

	// settings
	d.heading("Settings")
	if len(e.Settings) == 0 {
		d.note("This extension has no settings.")
	} else {
		file := ""
		if ad := m.a.opt.Ext; ad != nil {
			file = ad.SettingsFile(e.ID)
		}
		msg := "Click a value to change it. Changes apply immediately."
		if file != "" {
			msg += " Stored in " + file + ", not in zagros.toml."
		}
		d.note(msg)
		d.gap(4)
		for i, st := range e.Settings {
			m.drawSetting(d, i, st)
		}
	}

	// permissions
	d.heading("Permissions")
	perms := append([]string(nil), e.Declared...)
	for _, g := range e.Permissions {
		found := false
		for _, q := range perms {
			found = found || q == g
		}
		if !found {
			perms = append(perms, g)
		}
	}
	if len(perms) == 0 {
		d.note("None. This extension cannot reach files, the network or the clipboard.")
	}
	for _, perm := range perms {
		gi := -1
		for i, g := range e.Permissions {
			if g == perm {
				gi = i
			}
		}
		rh := p.s(28)
		bw := p.ghostWidth(dc, "Revoke")
		p.textFit(dc, p.fMono, d.x, d.y+(rh-p.hMono)/2, d.w-bw-p.s(12), perm, uint32(c.Text))
		if gi >= 0 {
			d.ghost(Rect{d.x + d.w - bw, d.y, d.x + d.w, d.y + rh}, idRevokeBase+gi, "Revoke", m.busy, false, true)
		} else {
			lw := textWidthFont(dc, p.fSmall, "not granted")
			p.text(dc, p.fSmall, d.x+d.w-lw, d.y+(rh-p.hSmall)/2, "not granted", dim)
		}
		d.y += rh
		if why := e.Reasons[perm]; why != "" {
			d.note(why)
		}
		d.gap(4)
	}

	// uninstall
	d.gap(16)
	p.rule(dc, Rect{d.x, d.y, d.x + d.w, d.y + 1})
	d.gap(16)
	bw := p.ghostWidth(dc, "Uninstall")
	d.ghost(Rect{d.x, d.y, d.x + bw, d.y + bh}, idUninstall, "Uninstall", m.busy, false, true)
	d.y += bh
}

func (m *manager) drawSetting(d *dpane, i int, st SettingInfo) {
	p, dc := d.p, d.dc
	c := p.th.Colors
	rh := p.s(32)
	cw := p.s(180)
	label := st.Label
	p.textFit(dc, p.fMain, d.x, d.y+(rh-p.hMain)/2, d.w-cw-p.s(12), label, uint32(c.Text))
	ctl := Rect{d.x + d.w - cw, d.y, d.x + d.w, d.y + rh}
	editing := m.editKey == st.Key
	switch {
	case st.Type == "bool":
		sw := Rect{ctl.R - p.s(36), d.y + (rh-p.s(18))/2, ctl.R, d.y + (rh-p.s(18))/2 + p.s(18)}
		d.toggle(sw, idSetBase+i, st.Value == "true", m.busy)
	case editing:
		p.input(dc, ctl, m.editBuf, "", true, true)
	default:
		v := st.Value
		if v == "" {
			v = "(empty)"
		}
		d.ghost(ctl, idSetBase+i, v, m.busy, false, false)
	}
	d.y += rh
	if desc := st.Description; desc != "" {
		d.note(desc)
	}
	switch {
	case editing:
		hint := "Enter saves, Esc cancels."
		if m.editErr != "" {
			p.text(dc, p.fSmall, d.x, d.y, m.editErr, dangerBg)
			d.y += p.hSmall + p.s(3)
		}
		d.note(hint)
	case st.Type == "choice":
		d.note("Click to cycle: " + strings.Join(st.Options, ", "))
	case st.Default != "" && st.Default != st.Value && st.Type != "bool":
		d.note("Default: " + st.Default)
	}
	d.gap(6)
}

func (m *manager) bodyMarket(d *dpane, e MarketEntry) {
	p := d.p
	dim := uint32(p.th.Colors.TextDim)
	if e.Description != "" {
		d.para(e.Description, uint32(p.th.Colors.Text))
	}
	if e.Kind == "theme" {
		d.heading("After installing")
		d.para("Open the Themes tab and choose Use this theme.", uint32(p.th.Colors.Text))
		return
	}
	d.heading("It will ask for")
	if len(e.Permissions) == 0 {
		d.note("No permissions.")
	}
	for _, perm := range e.Permissions {
		p.textFit(d.dc, p.fMono, d.x, d.y, d.w, perm, uint32(p.th.Colors.Text))
		d.y += p.hMono + p.s(6)
	}
	d.note("You approve each permission before the extension runs, and can revoke it later.")
	d.heading("After installing")
	d.para("Open the Installed tab: it shows what to type, the settings and why each permission is needed.", dim)
}

func (m *manager) bodyTheme(d *dpane, t ThemeInfo) {
	p := d.p
	d.para("Applies to the launcher and these windows as soon as you choose it.", uint32(p.th.Colors.Text))
	d.heading("Make your own")
	d.note("A theme is a TOML file in the themes folder of your config folder. Missing keys fall back to Boxy. See docs/themes.md.")
}
