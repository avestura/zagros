//go:build windows

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// richExt is a fixture with the fields a real manifest fills in.
type richExt struct{ demoExt }

func (richExt) SettingsFile(string) string {
	return `C:\Users\you\AppData\Roaming\Zagros\data\ext\dev.zagros.github-repos\settings.json`
}

func (richExt) List() []ExtInfo {
	return []ExtInfo{
		{ID: "dev.zagros.github-repos", Name: "GitHub Repositories", Version: "1.0.0", Publisher: "Zagros", Enabled: true,
			Log:         []LogLine{{"09:12:01", "info", "loaded"}, {"09:14:40", "warn", "rate limit: 3 requests left"}, {"09:14:52", "error", "http.get https://api.github.com/search/repositories: HTTP 403 rate limit exceeded; try again in 41 seconds"}},
			Description: "Search GitHub repositories with 'gh <query>' and open them in the browser.",
			Homepage:    "https://github.com/avestura/zagros", Dir: `C:\Users\you\AppData\Roaming\Zagros\ext\dev.zagros.github-repos`,
			Update: "1.1.0", Keywords: []string{"gh"}, Usage: "Type 'gh <query>'. Enter opens the repository, Tab shows more actions.",
			Examples: []string{"gh starlark", "gh zagros launcher"},
			Declared: []string{"network:api.github.com", "open", "clipboard.write"}, Permissions: []string{"network:api.github.com", "open"},
			Reasons: map[string]string{"network:api.github.com": "Calls the GitHub search API. No other host is contacted.",
				"open": "Opens the selected repository page in your browser.", "clipboard.write": "Lets you copy the clone URL from the actions menu."},
			Settings: []SettingInfo{
				{Label: "Maximum results", Key: "max_results", Type: "int", Value: "8", Default: "8"},
				{Label: "Sort by", Key: "sort", Type: "choice", Value: "stars", Default: "stars", Options: []string{"stars", "forks", "updated", "best-match"}},
				{Label: "Include forks", Key: "forks", Type: "bool", Value: "true", Default: "false", Description: "Also list repositories that are forks."},
			}},
		{ID: "dev.zagros.hello", Name: "Hello", Version: "1.0.0", Publisher: "Zagros", Enabled: false,
			Description: "Minimal example.", Keywords: []string{"hello"}, Examples: []string{"hello world"}, Declared: []string{"clipboard.write"}},
		{ID: "x.broken", Name: "A broken extension with a really long display name that will not fit", Version: "0.1.0", Enabled: true,
			Err: "main.star:12: undefined: foo"},
	}
}

// managerShots saves one PNG per manager state into ShotsDir.
func (a *App) managerShots() error {
	if err := os.MkdirAll(a.ShotsDir, 0o755); err != nil {
		return err
	}
	a.opt.Ext = richExt{}
	a.opt.Market = demoMarket{}
	a.opt.Themes = demoThemes{}
	a.openManager()
	m := a.mgr
	if m == nil {
		return fmt.Errorf("manager not created")
	}
	defer m.p.close()
	m.mk = []MarketEntry{
		{Ref: "hello", Name: "Weather", Description: "Forecasts in the launcher.", Author: "someone (verified)", Version: "1.2.0", Permissions: []string{"network:api.open-meteo.com", "storage"}},
		{Ref: "dusk", Name: "Dusk", Description: "A dark theme", Author: "someone", Version: "0.2.0", Kind: "theme"},
	}
	m.mkLoaded = true
	m.refreshThemes()
	shot := func(name string) error {
		p := m.p
		dc, bmp, old := offscreen(p.w, p.h)
		defer freeOffscreen(dc, bmp, old)
		p.render(dc)
		r := &screenRegion{w: p.w, h: p.h, mem: dc, bmp: bmp, buf: make([]byte, p.w*p.h*4)}
		r.savePNGNoGrab(filepath.Join(a.ShotsDir, name+".png"))
		return nil
	}
	steps := []struct {
		name string
		f    func()
	}{
		{"installed", func() {}},
		{"installed-scrolled", func() { m.px[3].jump(480) }},
		{"installed-log", func() { m.px[3].jump(1200) }},
		{"installed-editing", func() { m.px[3].jump(0); m.editKey, m.editBuf = "max_results", "12x" }},
		{"installed-focus", func() { m.px[3].jump(0); m.p.focusID = idTryBase + 1 }},
		{"installed-disabled", func() { m.editKey = ""; m.scI.Select(1) }},
		{"installed-error", func() { m.p.focusID = 0; m.scI.Select(2) }},
		{"installed-filter", func() { m.filter = "gh" }},
		{"market", func() { m.filter = ""; m.setTab(1) }},
		{"themes", func() { m.setTab(2) }},
	}
	for _, s := range steps {
		s.f()
		if err := renderPanel(m.p); err != nil {
			return fmt.Errorf("shot %s: %w", s.name, err)
		}
		if err := shot(s.name); err != nil {
			return err
		}
	}
	return nil
}

// settingsExt records what the manager asks the extension layer to do.
type settingsExt struct {
	richExt
	got *[3]string
}

func (s settingsExt) SetSetting(id, key, text string) error {
	*s.got = [3]string{id, key, text}
	if text == "bad" {
		return fmt.Errorf("expected a whole number")
	}
	return nil
}

// managerBehaviour checks filtering, setting edits and Try on the manager.
func (a *App) managerBehaviour() error {
	var got [3]string
	ext, mk, thm := a.opt.Ext, a.opt.Market, a.opt.Themes
	defer func() { a.opt.Ext, a.opt.Market, a.opt.Themes = ext, mk, thm }()
	a.opt.Ext = settingsExt{got: &got}
	a.opt.Market, a.opt.Themes = nil, nil
	a.openManager()
	m := a.mgr
	if m == nil {
		return fmt.Errorf("manager not created")
	}
	defer m.p.close()
	if err := renderPanel(m.p); err != nil {
		return err
	}
	for _, r := range "hel" {
		m.onChar(r)
	}
	if ix := m.idx(0); len(ix) != 1 || m.inst[ix[0]].ID != "dev.zagros.hello" {
		return fmt.Errorf("filter %q matched %v", m.filter, ix)
	}
	if e, ok := m.selInst(); !ok || e.ID != "dev.zagros.hello" {
		return fmt.Errorf("selection must follow the filter, got %v", e.ID)
	}
	if !m.onKey(vkEscape) || m.filter != "" {
		return fmt.Errorf("Esc must clear the filter first")
	}
	if err := renderPanel(m.p); err != nil {
		return err
	}
	e, _ := m.selInst()
	if e.ID != "dev.zagros.github-repos" || len(tryQueries(e)) != 2 {
		return fmt.Errorf("first extension %q, tries %v", e.ID, tryQueries(e))
	}
	if q := tryQueries(ExtInfo{Keywords: []string{"gh"}}); len(q) != 1 || q[0] != "gh " {
		return fmt.Errorf("keyword fallback for Try: %v", q)
	}
	// clicking an int setting opens the editor, typing and Enter save it
	st := e.Settings[0]
	m.setting(e, st)
	if m.editKey != st.Key || m.editBuf != "8" {
		return fmt.Errorf("editor not opened: %q %q", m.editKey, m.editBuf)
	}
	m.onChar(8)
	m.onChar('9')
	m.onKey(vkReturn)
	if !m.busy {
		return fmt.Errorf("saving a setting must run in the background")
	}
	for i := 0; i < 100 && m.busy; i++ {
		time.Sleep(5 * time.Millisecond)
		a.runQueue()
	}
	if got != [3]string{"dev.zagros.github-repos", "max_results", "9"} {
		return fmt.Errorf("SetSetting got %v", got)
	}
	// a rejected value keeps the editor open and shows the error
	m.setting(e, st)
	m.editBuf = "bad"
	m.onKey(vkReturn)
	for i := 0; i < 100 && m.busy; i++ {
		time.Sleep(5 * time.Millisecond)
		a.runQueue()
	}
	if m.editKey == "" || m.editErr == "" {
		return fmt.Errorf("an invalid value must keep the editor open with an error (%q %q)", m.editKey, m.editErr)
	}
	m.onKey(vkEscape)
	if err := m.focusTest(&got); err != nil {
		return err
	}
	// booleans flip and choices cycle without an editor
	m.setting(e, e.Settings[2])
	for i := 0; i < 100 && m.busy; i++ {
		time.Sleep(5 * time.Millisecond)
		a.runQueue()
	}
	if got[1] != "forks" || got[2] != "false" {
		return fmt.Errorf("bool toggle sent %v", got)
	}
	m.setting(e, e.Settings[1])
	for i := 0; i < 100 && m.busy; i++ {
		time.Sleep(5 * time.Millisecond)
		a.runQueue()
	}
	if got[1] != "sort" || got[2] != "forks" {
		return fmt.Errorf("choice cycle sent %v", got)
	}
	return nil
}

// focusTest walks keyboard focus through the Installed tab and presses a
// switch with Space.
func (m *manager) focusTest(got *[3]string) error {
	step := func() int {
		renderPanel(m.p)
		m.onKey(vkTab)
		return m.p.focusID
	}
	m.p.focusID = 0
	var order []int
	for i := 0; i < 40; i++ {
		id := step()
		if len(order) > 0 && id == order[0] {
			break
		}
		order = append(order, id)
	}
	has := func(want int) bool {
		for _, id := range order {
			if id == want {
				return true
			}
		}
		return false
	}
	for name, want := range map[string]int{"switch": idToggle, "try": idTryBase, "setting": idSetBase, "bool setting": idSetBase + 2,
		"revoke": idRevokeBase, "uninstall": idUninstall, "refresh": idRefreshInst} {
		if !has(want) {
			return fmt.Errorf("Tab never focuses %s (%d); order %v", name, want, order)
		}
	}
	for _, id := range order {
		if idIsTab(id) {
			return fmt.Errorf("Tab must not stop on the tab strip: %v", order)
		}
	}
	// the focused control scrolls into view
	m.p.focusID = idUninstall
	renderPanel(m.p)
	if m.detH <= m.dview.H() {
		return fmt.Errorf("fixture must overflow the detail pane (%d <= %d)", m.detH, m.dview.H())
	}
	m.px[3] = smooth{}
	m.p.focusID = idRevokeBase + 1
	renderPanel(m.p)
	m.focusMove(false) // -> next after revoke 1
	if m.px[3].target == 0 {
		return fmt.Errorf("focusing a control below the fold did not scroll the pane")
	}
	// Space presses the focused switch
	m.px[3] = smooth{}
	m.p.focusID = idSetBase + 2
	renderPanel(m.p)
	*got = [3]string{}
	m.onChar(' ')
	for i := 0; i < 100 && m.busy; i++ {
		time.Sleep(5 * time.Millisecond)
		m.a.runQueue()
	}
	if got[1] != "forks" {
		return fmt.Errorf("Space on the focused switch sent %v", *got)
	}
	// reverse order starts from the end
	m.p.focusID = 0
	renderPanel(m.p)
	m.onKey(vkTab) // forward from none: first
	first := m.p.focusID
	if first != order[0] {
		return fmt.Errorf("first focus %d, want %d", first, order[0])
	}
	m.p.focusID = 0
	if b, ok := m.p.moveFocus(idIsTab, true); !ok || b.id != order[len(order)-1] {
		return fmt.Errorf("Shift+Tab from none should reach the last control, got %v", b.id)
	}
	m.p.focusID = 0
	return nil
}
