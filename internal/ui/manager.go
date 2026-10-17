//go:build windows

package ui

import (
	"context"
	"fmt"
	"time"
)

const (
	idTabInstalled = 10
	idTabMarket    = 11
	idTabThemes    = 12
	idToggle       = 20
	idUninstall    = 21
	idRefreshInst  = 22
	idInstall      = 30
	idRefreshMkt   = 31
	idUseTheme     = 40
	idRefreshThm   = 41
	idRevokeBase   = 100
)

// manager is the extension manager window: installed extensions on one tab,
// marketplace and themes on the others.
type manager struct {
	a   *App
	p   *panel
	tab int

	inst      []ExtInfo
	mk        []MarketEntry
	mkErr     string
	mkLoading bool
	mkLoaded  bool
	themes    []ThemeInfo
	scI, scM  Scroll
	scT       Scroll
	status    string
	busy      bool

	list Rect // list viewport from the last draw
	rowH int
}

func (a *App) openManager() {
	if a.mgr != nil {
		forceForeground(a.mgr.p.hwnd)
		return
	}
	m := &manager{a: a}
	m.p = a.newPanel("Zagros extensions", 760, 540)
	m.p.draw = m.draw
	m.p.onBtn = m.onBtn
	m.p.onClick = m.onClick
	m.p.onWheel = func(d int) { m.sc().Wheel(d * 2) }
	m.p.onKey = m.onKey
	m.p.onClose = func() { a.mgr = nil }
	a.mgr = m
	if a.opt.Ext == nil && a.opt.Market != nil {
		m.tab = 1
	}
	m.refreshInstalled()
	m.refreshThemes()
	if m.tab == 1 {
		m.refreshMarket()
	}
	m.p.show()
}

func (m *manager) sc() *Scroll {
	switch m.tab {
	case 0:
		return &m.scI
	case 2:
		return &m.scT
	}
	return &m.scM
}

func (m *manager) refreshThemes() {
	if m.a.opt.Themes != nil {
		m.themes = m.a.opt.Themes.List()
	}
	m.scT.SetCount(len(m.themes), max(m.scT.Vis, 1))
}

// tabOK reports whether tab t has a backend.
func (m *manager) tabOK(t int) bool {
	switch t {
	case 0:
		return m.a.opt.Ext != nil
	case 1:
		return m.a.opt.Market != nil
	}
	return m.a.opt.Themes != nil
}

// kindTag is the short label shown before a marketplace entry's name.
func kindTag(kind string) string {
	if kind == "theme" {
		return "THEME"
	}
	return "EXT"
}

func (m *manager) useTheme() {
	th := m.a.opt.Themes
	if th == nil || m.scT.Sel >= len(m.themes) || m.busy {
		return
	}
	t := m.themes[m.scT.Sel]
	if err := th.Use(t.Name); err != nil {
		m.status = "Error: " + err.Error()
		return
	}
	m.busy, m.status = true, "Switching to "+t.Name+"..."
	go func() {
		var err error
		switch {
		case m.a.opt.ReloadTheme != nil:
			err = m.a.opt.ReloadTheme() // a theme change needs no extension discovery or re-indexing
		case m.a.opt.Reload != nil:
			err = m.a.opt.Reload()
		}
		m.a.Post(func() {
			m.busy = false
			if err != nil {
				m.status = "Theme set, reload failed: " + err.Error()
			} else {
				m.status = "Using theme " + t.Name
			}
			m.refreshThemes()
			if m.p != nil && !m.p.closed {
				invalidate(m.p.hwnd)
			}
		})
	}()
}

func (m *manager) refreshInstalled() {
	if m.a.opt.Ext != nil {
		m.inst = m.a.opt.Ext.List()
	}
	m.scI.SetCount(len(m.inst), max(m.scI.Vis, 1))
}

func (m *manager) refreshMarket() {
	mk := m.a.opt.Market
	if mk == nil || m.mkLoading {
		return
	}
	m.mkLoading, m.mkErr = true, ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		list, err := mk.List(ctx)
		m.a.Post(func() {
			m.mkLoading, m.mkLoaded = false, true
			if err != nil {
				m.mkErr = err.Error()
			} else {
				m.mk = list
				m.scM.SetCount(len(list), max(m.scM.Vis, 1))
			}
			if m.p != nil && !m.p.closed {
				invalidate(m.p.hwnd)
			}
		})
	}()
}

func (m *manager) setTab(t int) {
	m.tab = t
	if t == 1 && !m.mkLoaded {
		m.refreshMarket()
	}
	if t == 2 {
		m.refreshThemes()
	}
}

func (m *manager) selInst() (ExtInfo, bool) {
	if m.scI.Sel < len(m.inst) {
		return m.inst[m.scI.Sel], true
	}
	return ExtInfo{}, false
}

func (m *manager) onBtn(id int) {
	opt := m.a.opt
	switch {
	case id == idTabInstalled:
		m.setTab(0)
	case id == idTabMarket:
		m.setTab(1)
	case id == idTabThemes:
		m.setTab(2)
	case id == idUseTheme:
		m.useTheme()
	case id == idRefreshThm:
		m.refreshThemes()
		m.status = "Refreshed"
	case id == idRefreshInst:
		m.refreshInstalled()
		m.status = "Refreshed"
	case id == idRefreshMkt:
		m.mkLoaded = false
		m.refreshMarket()
	case id == idToggle:
		if e, ok := m.selInst(); ok && opt.Ext != nil {
			if err := opt.Ext.SetEnabled(e.ID, !e.Enabled); err != nil {
				m.status = "Error: " + err.Error()
			} else {
				m.status = fmt.Sprintf("%s %s", e.Name, map[bool]string{true: "disabled", false: "enabled"}[e.Enabled])
			}
			m.refreshInstalled()
		}
	case id == idUninstall:
		if e, ok := m.selInst(); ok && opt.Ext != nil {
			if !messageBox(m.p.hwnd, "Uninstall extension", "Uninstall "+e.Name+"? Its permissions and data are removed.", true) {
				return
			}
			if err := opt.Ext.Uninstall(e.ID); err != nil {
				m.status = "Error: " + err.Error()
			} else {
				m.status = e.Name + " uninstalled"
			}
			m.refreshInstalled()
		}
	case id == idInstall:
		m.install()
	case id >= idRevokeBase:
		if e, ok := m.selInst(); ok && opt.Ext != nil && id-idRevokeBase < len(e.Permissions) {
			perm := e.Permissions[id-idRevokeBase]
			if err := opt.Ext.Revoke(e.ID, perm); err != nil {
				m.status = "Error: " + err.Error()
			} else {
				m.status = "Revoked " + perm
			}
			m.refreshInstalled()
		}
	}
}

func (m *manager) install() {
	mk := m.a.opt.Market
	if mk == nil || m.scM.Sel >= len(m.mk) || m.busy {
		return
	}
	e := m.mk[m.scM.Sel]
	m.busy, m.status = true, "Installing "+e.Name+"..."
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		err := mk.Install(ctx, e)
		if err == nil && m.a.opt.Reload != nil {
			err = m.a.opt.Reload()
		}
		m.a.Post(func() {
			m.busy = false
			if err != nil {
				m.status = "Install failed: " + err.Error()
			} else {
				m.status = e.Name + " installed"
				if e.Kind == "theme" {
					m.status = e.Name + " installed. Use it from the Themes tab"
				}
				m.refreshInstalled()
				m.refreshThemes()
			}
			if m.p != nil && !m.p.closed {
				invalidate(m.p.hwnd)
			}
		})
	}()
}

func (m *manager) onClick(x, y int) {
	if !m.list.Contains(x, y) || m.rowH == 0 {
		return
	}
	s := m.sc()
	if i := s.Abs((y - m.list.T) / m.rowH); i >= 0 {
		s.Select(i)
	}
}

func (m *manager) onKey(vk int) bool {
	s := m.sc()
	switch vk {
	case vkUp:
		s.Move(-1, false)
	case vkDown:
		s.Move(1, false)
	case vkPrior:
		s.PageUp()
	case vkNext:
		s.PageDown()
	case vkHome:
		s.Home()
	case vkEnd:
		s.End()
	case vkTab:
		for i := 1; i <= 3; i++ {
			if t := (m.tab + i) % 3; m.tabOK(t) {
				m.setTab(t)
				break
			}
		}
	default:
		return false
	}
	return true
}

func (m *manager) draw(p *panel, dc uintptr) {
	c := p.th.Colors
	pad := p.s(14)
	t := p.titleH()
	ty := t + p.s(10)
	th := p.s(32)
	tw := p.s(130)
	p.button(dc, Rect{pad, ty, pad + tw, ty + th}, idTabInstalled, fmt.Sprintf("Installed (%d)", len(m.inst)), m.a.opt.Ext == nil, false, m.tab == 0)
	p.button(dc, Rect{pad + tw + p.s(6), ty, pad + 2*tw + p.s(6), ty + th}, idTabMarket, "Marketplace", m.a.opt.Market == nil, false, m.tab == 1)
	p.button(dc, Rect{pad + 2*(tw+p.s(6)), ty, pad + 3*tw + 2*p.s(6), ty + th}, idTabThemes, fmt.Sprintf("Themes (%d)", len(m.themes)), m.a.opt.Themes == nil, false, m.tab == 2)

	footerH := p.s(56)
	detailH := 0
	if m.tab == 0 {
		detailH = p.s(64)
	}
	m.rowH = p.hMain + p.hSmall + p.s(20)
	top := ty + th + p.s(10)
	m.list = Rect{pad, top, p.w - pad, p.h - footerH - detailH}
	p.fill(dc, m.list, uint32(c.InputBg))
	vis := max(m.list.H()/m.rowH, 1)
	s := m.sc()
	n := len(m.inst)
	switch m.tab {
	case 1:
		n = len(m.mk)
	case 2:
		n = len(m.themes)
	}
	s.SetCount(n, vis)
	empty := ""
	switch {
	case m.tab == 0 && n == 0:
		empty = "No extensions installed. Browse the Marketplace tab or run: zagros ext install <path>"
	case m.tab == 2 && n == 0:
		empty = "No themes found. Install one from the Marketplace tab or run: zagros theme install <id>"
	case m.tab == 1 && m.mkLoading:
		empty = "Loading marketplace..."
	case m.tab == 1 && m.mkErr != "":
		empty = "Could not load the marketplace: " + m.mkErr
	case m.tab == 1 && n == 0:
		empty = "The marketplace is empty."
	}
	if empty != "" {
		p.textFit(dc, p.fMain, m.list.L+pad, m.list.T+pad, m.list.W()-2*pad, empty, uint32(c.TextDim))
	}
	restore := p.clip(dc, m.list)
	for i := 0; i < vis && s.Top+i < n; i++ {
		abs := s.Top + i
		r := Rect{m.list.L, m.list.T + i*m.rowH, m.list.R, m.list.T + (i+1)*m.rowH}
		fg, dim := uint32(c.Text), uint32(c.TextDim)
		if abs == s.Sel {
			p.fill(dc, r, uint32(c.SelBg))
			fg, dim = uint32(c.SelText), uint32(c.SelTextDim)
		}
		x := r.L + p.s(10)
		right := r.R - p.s(10)
		var l1, l2, tag string
		tagC := dim
		if m.tab == 0 {
			e := m.inst[abs]
			l1 = e.Name
			if e.Version != "" {
				l1 += "  v" + e.Version
			}
			l2 = fmt.Sprintf("%s  -  %d permission(s) granted", orDash(e.Publisher), len(e.Permissions))
			tag = "Disabled"
			if e.Enabled {
				tag, tagC = "Enabled", uint32(c.Accent)
				if abs == s.Sel {
					tagC = fg
				}
			}
		} else if m.tab == 2 {
			e := m.themes[abs]
			l1 = e.Name
			l2 = e.Source
			if e.Author != "" {
				l2 += "  -  " + e.Author
			}
			if e.Current {
				tag, tagC = "In use", uint32(c.Accent)
				if abs == s.Sel {
					tagC = fg
				}
			}
		} else {
			e := m.mk[abs]
			l1 = kindTag(e.Kind) + "  " + e.Name
			if e.Version != "" {
				l1 += "  v" + e.Version
			}
			l2, tag = e.Description, e.Author
		}
		tw := textWidthFont(dc, p.fSmall, tag)
		p.text(dc, p.fSmall, right-tw, r.T+(r.H()-p.hSmall)/2, tag, tagC)
		avail := right - tw - p.s(12) - x
		p.textFit(dc, p.fBold, x, r.T+p.s(8), avail, l1, fg)
		p.textFit(dc, p.fSmall, x, r.T+p.s(8)+p.hMain+p.s(2), avail, l2, dim)
	}
	restore()

	if m.tab == 0 {
		dy := m.list.B + p.s(8)
		p.text(dc, p.fSmall, pad, dy, "Granted permissions (click one to revoke it):", uint32(c.TextDim))
		if e, ok := m.selInst(); ok {
			x := pad
			cy := dy + p.hSmall + p.s(6)
			ch := p.hSmall + p.s(10)
			for i, perm := range e.Permissions {
				w := textWidthFont(dc, p.fSmall, perm+"  x") + p.s(16)
				if x+w > p.w-pad {
					break
				}
				p.button(dc, Rect{x, cy, x + w, cy + ch}, idRevokeBase+i, perm+"  x", false, false, false)
				x += w + p.s(6)
			}
			if len(e.Permissions) == 0 {
				p.text(dc, p.fSmall, pad, cy, "None", uint32(c.TextDim))
			}
		}
	}

	fy := p.h - footerH + p.s(12)
	bh := p.s(32)
	bw := p.s(110)
	right := p.w - pad
	if m.tab == 0 {
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idRefreshInst, "Refresh", false, false, false)
		right -= bw + p.s(8)
		_, ok := m.selInst()
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idUninstall, "Uninstall", !ok, false, false)
		right -= bw + p.s(8)
		lbl := "Disable"
		if e, ok2 := m.selInst(); ok2 && !e.Enabled {
			lbl = "Enable"
		}
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idToggle, lbl, !ok, false, false)
		right -= bw + p.s(8)
	} else if m.tab == 2 {
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idRefreshThm, "Refresh", false, false, false)
		right -= bw + p.s(8)
		use := len(m.themes) == 0 || m.busy || (m.scT.Sel < len(m.themes) && m.themes[m.scT.Sel].Current)
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idUseTheme, "Use", use, true, false)
		right -= bw + p.s(8)
	} else {
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idRefreshMkt, "Refresh", false, false, false)
		right -= bw + p.s(8)
		p.button(dc, Rect{right - bw, fy, right, fy + bh}, idInstall, "Install", len(m.mk) == 0 || m.busy, true, false)
		right -= bw + p.s(8)
	}
	if m.status != "" {
		p.textFit(dc, p.fSmall, pad, fy+(bh-p.hSmall)/2, right-pad, m.status, uint32(c.TextDim))
	}
}
