//go:build windows

package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	idTabInstalled = 10
	idTabMarket    = 11
	idTabThemes    = 12
	idToggle       = 20
	idUninstall    = 21
	idRefreshInst  = 22
	idInstFile     = 23
	idInstFolder   = 24
	idOpenDir      = 25
	idHomepage     = 26
	idLogOpen      = 27
	idLogCopy      = 28
	idLogClear     = 29
	idInstall      = 30
	idRefreshMkt   = 31
	idUseTheme     = 40
	idRefreshThm   = 41
	idRevokeBase   = 100  // + index into ExtInfo.Permissions
	idTryBase      = 500  // + index of the example query
	idSetBase      = 1000 // + index into ExtInfo.Settings
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
	px        [4]smooth // pixel scroll of each tab's list; [3] is the detail pane
	status    string
	busy      bool

	filter  string // narrows the list of the current tab
	editKey string // setting being typed into, "" when none
	editBuf string
	editErr string
	dkey    string // identity of the item shown in the detail pane

	list  Rect // list viewport from the last draw
	dpane Rect // detail pane
	dview Rect // scrolling part of the detail pane
	detH  int  // height of the detail content
	rowH  int
	bar   Rect // scrollbar track from the last draw (empty when the list fits)
	// grab is where inside the thumb the drag started (px from its top).
	grab int
}

// pxs is the pixel scroll of the current tab.
func (m *manager) pxs() *smooth { return &m.px[m.tab] }

func (a *App) openManager() {
	if a.mgr != nil {
		forceForeground(a.mgr.p.hwnd)
		return
	}
	m := &manager{a: a}
	m.p = a.newPanel("Zagros extensions", 920, 600)
	m.p.draw = m.draw
	m.p.onBtn = m.onBtn
	m.p.onClick = m.onClick
	m.p.onHoverMove = func() {}
	m.p.onWheel = m.onWheel
	m.p.onChar = m.onChar
	m.p.onTick = m.tick
	m.p.onDrag = m.onDrag
	m.p.onKey = m.onKey
	m.p.onDrop = m.installLocal
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

func (m *manager) sc() *Scroll { return m.scrollOf(m.tab) }

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
	t, ok := m.selTheme()
	if th == nil || !ok || m.busy || t.Current {
		return
	}
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
	m.p.focusID = 0
	m.filter, m.editKey, m.editErr = "", "", ""
	if t == 1 && !m.mkLoaded {
		m.refreshMarket()
	}
	if t == 2 {
		m.refreshThemes()
	}
}

// run does slow work (disk, extension start/stop, re-indexing) off the UI
// thread so the window keeps painting, then applies the result on it.
func (m *manager) run(status string, work func() error, done func(err error)) {
	if m.busy {
		return
	}
	m.busy, m.status = true, status
	go func() {
		err := work()
		m.a.Post(func() {
			m.busy = false
			done(err)
			if m.p != nil && !m.p.closed {
				m.p.refresh()
			}
		})
	}()
}

func (m *manager) scrollOf(tab int) *Scroll {
	switch tab {
	case 0:
		return &m.scI
	case 2:
		return &m.scT
	}
	return &m.scM
}

// idx lists the source indexes of the rows of tab that match the filter.
func (m *manager) idx(tab int) []int {
	n := len(m.themes)
	switch tab {
	case 0:
		n = len(m.inst)
	case 1:
		n = len(m.mk)
	}
	words := strings.Fields(strings.ToLower(m.filter))
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if len(words) == 0 {
			out = append(out, i)
			continue
		}
		var hay string
		switch tab {
		case 0:
			e := m.inst[i]
			hay = e.Name + " " + e.Description + " " + strings.Join(e.Keywords, " ") + " " + e.Publisher + " " + e.ID
		case 1:
			e := m.mk[i]
			hay = e.Name + " " + e.Description + " " + e.Author + " " + kindTag(e.Kind)
		default:
			e := m.themes[i]
			hay = e.Name + " " + e.Author + " " + e.Source
		}
		hay = strings.ToLower(hay)
		ok := true
		for _, w := range words {
			ok = ok && strings.Contains(hay, w)
		}
		if ok {
			out = append(out, i)
		}
	}
	return out
}

// sel is the source index of the selected row of tab.
func (m *manager) sel(tab int) (int, bool) {
	ix := m.idx(tab)
	if s := m.scrollOf(tab); s.Sel >= 0 && s.Sel < len(ix) {
		return ix[s.Sel], true
	}
	return 0, false
}

func (m *manager) selInst() (ExtInfo, bool) {
	if i, ok := m.sel(0); ok {
		return m.inst[i], true
	}
	return ExtInfo{}, false
}

func (m *manager) selMarket() (MarketEntry, bool) {
	if i, ok := m.sel(1); ok {
		return m.mk[i], true
	}
	return MarketEntry{}, false
}

func (m *manager) selTheme() (ThemeInfo, bool) {
	if i, ok := m.sel(2); ok {
		return m.themes[i], true
	}
	return ThemeInfo{}, false
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
			word := map[bool]string{true: "disabled", false: "enabled"}[e.Enabled]
			m.run(fmt.Sprintf("Turning %s %s...", e.Name, map[bool]string{true: "off", false: "on"}[e.Enabled]),
				func() error { return opt.Ext.SetEnabled(e.ID, !e.Enabled) },
				func(err error) {
					m.status = fmt.Sprintf("%s %s", e.Name, word)
					if err != nil {
						m.status = "Error: " + err.Error()
					}
					m.refreshInstalled()
				})
		}
	case id == idUninstall:
		if e, ok := m.selInst(); ok && opt.Ext != nil {
			if !messageBox(m.p.hwnd, "Uninstall extension", "Uninstall "+e.Name+"? Its permissions and data are removed.", true) {
				return
			}
			m.run("Uninstalling "+e.Name+"...", func() error { return opt.Ext.Uninstall(e.ID) }, func(err error) {
				m.status = e.Name + " uninstalled"
				if err != nil {
					m.status = "Error: " + err.Error()
				}
				m.refreshInstalled()
			})
		}
	case id == idInstall:
		m.install()
	case id == idInstFile:
		if p := pickExtensionFile(m.p.hwnd); p != "" {
			m.installLocal([]string{p})
		}
	case id == idInstFolder:
		if p := pickExtensionFolder(m.p.hwnd); p != "" {
			m.installLocal([]string{p})
		}
	case id == idOpenDir:
		if e, ok := m.selInst(); ok && e.Dir != "" {
			go func() { _ = ShellOpen("open", e.Dir, "", "") }()
		}
	case id == idLogOpen:
		if f := opt.Ext.LogFile(); f != "" {
			go func() { _ = ShellOpen("open", f, "", "") }()
		}
	case id == idLogCopy:
		if e, ok := m.selInst(); ok {
			var b strings.Builder
			for _, l := range e.Log {
				b.WriteString(l.Time + " " + l.Level + " " + l.Msg + "\r\n")
			}
			_ = SetClipboard(b.String())
			m.status = "Copied the log of " + e.Name
		}
	case id == idLogClear:
		if e, ok := m.selInst(); ok {
			opt.Ext.ClearLog(e.ID)
			m.refreshInstalled()
		}
	case id == idHomepage:
		if e, ok := m.selInst(); ok && strings.HasPrefix(e.Homepage, "https://") {
			go func() { _ = ShellOpen("open", e.Homepage, "", "") }()
		}
	case id >= idSetBase:
		if e, ok := m.selInst(); ok && id-idSetBase < len(e.Settings) {
			m.setting(e, e.Settings[id-idSetBase])
		}
	case id >= idTryBase && id < idSetBase:
		if e, ok := m.selInst(); ok {
			if tr := tryQueries(e); id-idTryBase < len(tr) {
				m.a.tryQuery(tr[id-idTryBase])
				m.status = "Typed " + strconv.Quote(tr[id-idTryBase]) + " into the launcher"
			}
		}
	case id >= idRevokeBase && id < idTryBase:
		if e, ok := m.selInst(); ok && opt.Ext != nil && id-idRevokeBase < len(e.Permissions) {
			perm := e.Permissions[id-idRevokeBase]
			m.run("Revoking "+perm+"...", func() error { return opt.Ext.Revoke(e.ID, perm) }, func(err error) {
				m.status = "Revoked " + perm
				if err != nil {
					m.status = "Error: " + err.Error()
				}
				m.refreshInstalled()
			})
		}
	}
}

// installLocal installs dropped or picked folders and .zagros files, then
// reloads so the permission dialog appears for each.
func (m *manager) installLocal(paths []string) {
	ad := m.a.opt.Ext
	if ad == nil || m.busy || len(paths) == 0 {
		return
	}
	m.tab = 0
	m.busy, m.status = true, "Installing..."
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		var names []string
		var err error
		for _, p := range paths {
			var name string
			if name, err = ad.InstallLocal(ctx, p); err != nil {
				err = fmt.Errorf("%s: %w", p, err)
				break
			}
			names = append(names, name)
		}
		if len(names) > 0 && m.a.opt.Reload != nil {
			if rerr := m.a.opt.Reload(); rerr != nil && err == nil {
				err = rerr
			}
		}
		m.a.Post(func() {
			m.busy = false
			switch {
			case err != nil:
				m.status = "Install failed: " + err.Error()
			default:
				m.status = strings.Join(names, ", ") + " installed"
			}
			m.refreshInstalled()
			if m.p != nil && !m.p.closed {
				invalidate(m.p.hwnd)
			}
		})
	}()
}

func (m *manager) install() {
	mk := m.a.opt.Market
	e, ok := m.selMarket()
	if mk == nil || !ok || m.busy {
		return
	}
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

const barW = 10 // scrollbar width, logical px

// content and view are the list height in pixels and the visible part of it.
func (m *manager) content() int { return m.sc().N * m.rowH }
func (m *manager) view() int    { return m.list.H() }

func (m *manager) scrollBy(px float64) {
	sm := m.pxs()
	sm.by(px)
	sm.clamp(m.content(), m.view())
	m.p.startTick()
}

// tick advances the glide of the list and the detail pane.
func (m *manager) tick() bool {
	now := time.Now()
	a := m.pxs().step(now)
	b := m.px[3].step(now)
	return a || b
}

func (m *manager) scrollDetail(px float64) {
	sm := &m.px[3]
	sm.by(px)
	sm.clamp(m.detH, m.dview.H())
	m.p.startTick()
}

func (m *manager) onWheel(d int) {
	step := float64(d) * 2 * float64(m.rowH) / 120
	if m.p.mx >= m.dpane.L {
		m.scrollDetail(step)
		return
	}
	m.scrollBy(step)
}

// reveal glides the list so the selected row is fully visible.
func (m *manager) reveal() {
	if m.rowH == 0 {
		return
	}
	sm := m.pxs()
	top := float64(m.sc().Sel * m.rowH)
	if top < sm.target {
		sm.target = top
	} else if v := float64(m.view()); top+float64(m.rowH) > sm.target+v {
		sm.target = top + float64(m.rowH) - v
	}
	sm.clamp(m.content(), m.view())
	if sm.moving() {
		sm.last = time.Now()
		m.p.startTick()
	}
}

// thumb is the draggable part of the scrollbar track.
func (m *manager) thumb() Rect {
	content := max(m.content(), 1)
	h := min(max(m.bar.H()*m.view()/content, m.bar.W()*2), m.bar.H())
	t := m.bar.T
	if span := content - m.view(); span > 0 {
		t += int(float64(m.bar.H()-h) * m.pxs().pos / float64(span))
	}
	return Rect{m.bar.L, t, m.bar.R, t + h}
}

// onDrag moves the list while the thumb is held.
func (m *manager) onDrag(x, y int) {
	room := m.bar.H() - m.thumb().H()
	span := m.content() - m.view()
	if room <= 0 || span <= 0 {
		return
	}
	sm := m.pxs()
	sm.jump(float64(min(max((y-m.grab-m.bar.T)*span/room, 0), span)))
}

func (m *manager) onClick(x, y int) {
	if m.bar.Contains(x, y) {
		th := m.thumb()
		page := float64(m.view() - m.rowH)
		switch {
		case th.Contains(x, y):
			m.grab = y - th.T
			m.p.capture(true)
		case y < th.T:
			m.scrollBy(-page)
		default:
			m.scrollBy(page)
		}
		return
	}
	if !m.list.Contains(x, y) || m.rowH == 0 {
		return
	}
	s := m.sc()
	if i := int((float64(y-m.list.T) + m.pxs().pos) / float64(m.rowH)); i >= 0 && i < s.N {
		s.Select(i)
		m.reveal()
	}
}

func (m *manager) onKey(vk int) bool {
	if m.editKey != "" && vk != vkReturn && vk != vkEscape {
		return true // typing into a setting: arrows and Tab do not move the list
	}
	s := m.sc()
	switch vk {
	case vkUp:
		s.Move(-1, false)
		m.reveal()
	case vkDown:
		s.Move(1, false)
		m.reveal()
	case vkPrior:
		s.PageUp()
		m.reveal()
	case vkNext:
		s.PageDown()
		m.reveal()
	case vkHome:
		s.Home()
		m.reveal()
	case vkEnd:
		s.End()
		m.reveal()
	case vkReturn:
		switch {
		case m.editKey != "":
			m.commitEdit()
		case m.p.focusID != 0:
			m.onBtn(m.p.focusID)
		default:
			m.primary()
		}
	case vkEscape:
		switch {
		case m.editKey != "":
			m.editKey, m.editErr = "", ""
		case m.filter != "":
			m.filter = ""
		default:
			return false
		}
	case vkTab:
		back := keyDown(vkShift)
		if keyDown(vkControl) {
			m.nextTab(back)
			break
		}
		m.focusMove(back)
	case '1', '2', '3':
		if !keyDown(vkControl) || !m.tabOK(vk-'1') {
			return false
		}
		m.setTab(vk - '1')
	default:
		return false
	}
	return true
}

func idIsTab(id int) bool { return id >= idTabInstalled && id <= idTabThemes }

// nextTab switches to the next tab that has a backend.
func (m *manager) nextTab(back bool) {
	for i := 1; i <= 3; i++ {
		t := (m.tab + i) % 3
		if back {
			t = (m.tab - i + 6) % 3
		}
		if m.tabOK(t) {
			m.setTab(t)
			return
		}
	}
}

// focusMove moves keyboard focus through the controls and scrolls the detail
// pane so the focused one is visible.
func (m *manager) focusMove(back bool) {
	b, ok := m.p.moveFocus(idIsTab, back)
	if !ok || b.abs.W() == 0 {
		return
	}
	pad := m.p.s(16)
	switch {
	case b.abs.T < m.dview.T:
		m.scrollDetail(float64(b.abs.T - m.dview.T - pad))
	case b.abs.B > m.dview.B:
		m.scrollDetail(float64(b.abs.B - m.dview.B + pad))
	}
}

// primary runs the main action of the selected row (Enter).
func (m *manager) primary() {
	switch m.tab {
	case 0:
		if _, ok := m.selInst(); ok {
			m.onBtn(idToggle)
		}
	case 1:
		m.install()
	default:
		m.useTheme()
	}
}

// onChar types into the setting being edited, otherwise into the filter.
func (m *manager) onChar(r rune) bool {
	dst := &m.filter
	if m.editKey != "" {
		dst = &m.editBuf
	} else if r == ' ' && m.p.focusID != 0 {
		m.onBtn(m.p.focusID) // Space presses the focused control
		return true
	} else if r >= 0x20 {
		m.p.focusID = 0 // typing goes to the filter
	}
	switch {
	case r == 8:
		if rs := []rune(*dst); len(rs) > 0 {
			*dst = string(rs[:len(rs)-1])
		}
	case r < 0x20 || r == 0x7f || (r >= 0xD800 && r < 0xE000):
		return false
	default:
		*dst += string(r)
	}
	if m.editKey == "" {
		m.sc().Sel, m.sc().Top = 0, 0
		m.pxs().jump(0)
	}
	return true
}

// setting acts on a click on a setting: booleans flip, choices cycle, other
// types open an inline editor.
func (m *manager) setting(e ExtInfo, st SettingInfo) {
	switch st.Type {
	case "bool":
		m.saveSetting(e, st, strconv.FormatBool(st.Value != "true"))
	case "choice":
		if len(st.Options) == 0 {
			return
		}
		next := st.Options[0]
		for i, o := range st.Options {
			if o == st.Value {
				next = st.Options[(i+1)%len(st.Options)]
			}
		}
		m.saveSetting(e, st, next)
	default:
		m.editKey, m.editBuf, m.editErr = st.Key, st.Value, ""
	}
}

func (m *manager) commitEdit() {
	e, ok := m.selInst()
	if !ok {
		m.editKey = ""
		return
	}
	for _, st := range e.Settings {
		if st.Key == m.editKey {
			m.saveSetting(e, st, m.editBuf)
			return
		}
	}
}

func (m *manager) saveSetting(e ExtInfo, st SettingInfo, text string) {
	ad := m.a.opt.Ext
	if ad == nil {
		return
	}
	m.run("Saving "+st.Label+"...", func() error { return ad.SetSetting(e.ID, st.Key, text) }, func(err error) {
		if err != nil {
			m.status = "Error: " + err.Error()
			if m.editKey != "" {
				m.editErr = err.Error()
			}
			return
		}
		m.editKey, m.editErr = "", ""
		m.status = st.Label + " saved"
		m.refreshInstalled()
	})
}

// tryQueries are the queries offered as "Try" buttons for e: the manifest
// examples, or each keyword followed by a space.
func tryQueries(e ExtInfo) []string {
	if len(e.Examples) > 0 {
		return e.Examples
	}
	var out []string
	for _, k := range e.Keywords {
		out = append(out, k+" ")
	}
	return out
}

// tryQuery opens the launcher with q typed in.
func (a *App) tryQuery(q string) {
	a.showLauncher()
	a.ed.SetText(q)
	a.edited("")
}
