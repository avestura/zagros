//go:build windows

package ui

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/avestura/zagros/internal/core"
)

// SelfTest creates the windows, renders frames of the launcher, the
// permission dialog and the extension manager into offscreen bitmaps, checks a
// few pixels and tears everything down. It never shows a window.
func (a *App) SelfTest() error {
	runtime.LockOSThread() // the window and its message queue belong to this thread
	defer runtime.UnlockOSThread()
	if err := a.create(); err != nil {
		return err
	}
	defer pDestroyWindow.Call(uintptr(a.hwnd))
	defer func() {
		if a.loader != nil {
			a.loader.Close()
		}
	}()
	a.noShow = true

	for _, dpi := range []int{96, 168} {
		a.prepare(dpi)
		a.items = []core.Item{
			{ID: "t:1", Title: "Notepad", Subtitle: `C:\Windows\notepad.exe`, Category: "Apps", Icon: `C:\Windows\notepad.exe`,
				Action: core.Action{Kind: "run", Target: `C:\Windows\notepad.exe`}, Alt: []core.NamedAction{{Name: "Open folder"}}},
			{ID: "t:2", Title: "Zagros settings", Subtitle: "Open zagros.toml", Category: "Zagros", Action: core.Action{Kind: "open", Target: "zagros.toml"}},
			{ID: "t:3", Title: "A very long title that must be truncated with an ellipsis because it cannot possibly fit in the row width of the window", Category: "Test"},
		}
		a.query = "note"
		a.ed.SetText("note")
		a.caretOn = true
		a.sc.Reset(len(a.items), a.m.VisibleRows(len(a.items)))
		w, h := a.m.Width, a.m.Height(len(a.items))
		dc, bmp, old := offscreen(w, h)
		a.render(dc, w, h)
		c := a.th.Colors
		if got, want := getPixel(dc, w/2, a.m.Border+3), uint32(c.InputBg); got != want {
			return fmt.Errorf("dpi %d: input pixel %06x, want %06x", dpi, got, want)
		}
		if a.m.Border > 0 {
			if got, want := getPixel(dc, 0, 0), uint32(c.Border); got != want {
				return fmt.Errorf("dpi %d: border pixel %06x, want %06x", dpi, got, want)
			}
		}
		selRow := a.m.RowRect(0)
		if got, want := getPixel(dc, selRow.R-2, selRow.B-2), uint32(c.SelBg); got != want {
			return fmt.Errorf("dpi %d: selection pixel %06x, want %06x", dpi, got, want)
		}
		freeOffscreen(dc, bmp, old)
		a.freeResources()
	}

	if err := a.scopeTest(); err != nil {
		return err
	}
	if err := a.burstTest(); err != nil {
		return err
	}
	a.freeResources()

	// permission dialog
	ch := make(chan PermissionResult, 1)
	a.openPrompt(PermissionRequest{Extension: "Demo", Publisher: "someone", Source: "github:someone/demo",
		Asks: []PermissionAsk{{Name: "net", Scope: "api.example.com"}, {Name: "fs.read", Scope: `C:\Users\me\Documents`}}}, ch)
	if len(a.panels) != 1 {
		return fmt.Errorf("prompt panel not created")
	}
	if err := renderPanel(a.panels[0]); err != nil {
		return err
	}
	a.panels[0].close()
	if r := <-ch; !r.Cancelled {
		return fmt.Errorf("closing the prompt must cancel")
	}

	// manager
	a.opt.Ext = demoExt{}
	a.openManager()
	if a.mgr == nil {
		return fmt.Errorf("manager not created")
	}
	if err := renderPanel(a.mgr.p); err != nil {
		return err
	}
	a.mgr.p.close()

	// marketplace and themes tabs
	a.opt.Market = demoMarket{}
	a.opt.Themes = demoThemes{}
	a.openManager()
	if a.mgr == nil {
		return fmt.Errorf("manager not created")
	}
	a.mgr.mk = []MarketEntry{
		{Ref: "hello", Name: "Hello", Description: "Says hello", Author: "someone", Version: "1.0.0"},
		{Ref: "dusk", Name: "Dusk", Description: "A dark theme", Author: "someone (verified)", Version: "0.2.0", Kind: "theme"},
	}
	a.mgr.mkLoaded = true
	a.mgr.refreshThemes()
	for tab := 0; tab < 3; tab++ {
		a.mgr.setTab(tab)
		a.mgr.mkLoaded = true
		if err := renderPanel(a.mgr.p); err != nil {
			return fmt.Errorf("manager tab %d: %w", tab, err)
		}
	}
	if len(a.mgr.mk) != 2 || len(a.mgr.themes) != 2 {
		return fmt.Errorf("manager lists: %d market, %d themes", len(a.mgr.mk), len(a.mgr.themes))
	}
	// Regression: clicking a row must repaint the window straight away (the
	// highlight used to wait for the next mouse move).
	a.mgr.setTab(2)
	if err := renderPanel(a.mgr.p); err != nil {
		return err
	}
	hw := uintptr(a.mgr.p.hwnd)
	before := a.mgr.p.redraws
	a.mgr.scT.Select(0)
	x, y := a.mgr.list.L+a.mgr.p.s(20), a.mgr.list.T+a.mgr.rowH+a.mgr.rowH/2
	pSendMessageW.Call(hw, wmLButtonDown, 1, uintptr(uint32(y)<<16|uint32(x)&0xFFFF))
	if a.mgr.p.redraws == before {
		return fmt.Errorf("clicking a theme row did not invalidate the window")
	}
	if a.mgr.scT.Sel != 1 {
		return fmt.Errorf("clicking the second theme row selected %d", a.mgr.scT.Sel)
	}
	a.mgr.p.close()

	if err := a.managerBench(); err != nil {
		return err
	}

	// tray add/remove (informational: needs a shell)
	a.addTray()
	ok := a.trayAdded
	a.removeTray()
	fmt.Fprintf(os.Stderr, "selftest: tray icon added=%v\n", ok)
	return nil
}

// managerBench times frames of the extension manager with a long installed
// list (what a scroll or hover repaints) and fails when a frame is too slow.
func (a *App) managerBench() error {
	a.opt.Ext = manyExt{}
	a.openManager()
	if a.mgr == nil {
		return fmt.Errorf("manager not created")
	}
	defer a.mgr.p.close()
	p := a.mgr.p
	dc, bmp, old := offscreen(p.w, p.h)
	defer freeOffscreen(dc, bmp, old)
	const frames = 200
	t0 := time.Now()
	for i := 0; i < frames; i++ {
		a.mgr.pxs().jump(float64(i % 400))
		p.render(dc)
	}
	per := time.Since(t0) / frames
	if err := a.managerScrollTest(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "selftest: manager frame %v (%d rows listed)\n", per, len(a.mgr.inst))
	if per > 8*time.Millisecond {
		return fmt.Errorf("manager frame takes %v, want under 8ms", per)
	}
	return nil
}

// managerScrollTest drives the scrollbar, the precision-wheel accumulation and
// checks that a slow action leaves the UI thread free.
func (a *App) managerScrollTest() error {
	m := a.mgr
	p := m.p
	m.px[0] = smooth{}
	m.scI.Reset(len(m.inst), 1)
	if err := renderPanel(p); err != nil {
		return err
	}
	if m.bar.W() == 0 {
		return fmt.Errorf("a long list must show a scrollbar")
	}
	hw := uintptr(p.hwnd)
	at := func(x, y int) uintptr { return uintptr(uint32(y)<<16 | uint32(x)&0xFFFF) }
	sm := m.pxs()
	settle := func() {
		for i := 0; i < 200 && m.tick(); i++ {
			time.Sleep(5 * time.Millisecond)
		}
	}
	// small wheel steps (touchpad) add up: three 40/120 notches = 2 rows' worth
	for i := 0; i < 3; i++ {
		pSendMessageW.Call(hw, wmMouseWheel, uintptr(uint32(0xFFD8)<<16), at(10, 10)) // delta -40
	}
	if want := float64(2 * m.rowH); math.Abs(sm.target-want) > 1 {
		return fmt.Errorf("three -40 wheel steps aim at %.1f px, want %.1f", sm.target, want)
	}
	// the list glides: after one frame it is between the start and the target
	time.Sleep(16 * time.Millisecond)
	m.tick()
	if sm.pos <= 0 || sm.pos >= sm.target {
		return fmt.Errorf("scroll must glide, but pos=%.1f target=%.1f after one frame", sm.pos, sm.target)
	}
	settle()
	if sm.pos != sm.target {
		return fmt.Errorf("glide did not settle: pos=%.1f target=%.1f", sm.pos, sm.target)
	}
	// click below the thumb pages down
	before := sm.target
	pSendMessageW.Call(hw, wmLButtonDown, 1, at(m.bar.L+2, m.bar.B-2))
	if sm.target <= before {
		return fmt.Errorf("clicking the track below the thumb did not page down")
	}
	settle()
	// dragging the thumb to the bottom reaches the end exactly
	renderPanel(p)
	th := m.thumb()
	pSendMessageW.Call(hw, wmLButtonDown, 1, at(th.L+2, th.T+2))
	pSendMessageW.Call(hw, wmMouseMove, 1, at(th.L+2, m.bar.B+50))
	pSendMessageW.Call(hw, wmLButtonUp, 0, at(th.L+2, m.bar.B))
	if end := float64(m.content() - m.view()); sm.pos != end || p.dragging {
		return fmt.Errorf("dragging the thumb down left pos=%.1f, want %.1f (dragging=%v)", sm.pos, end, p.dragging)
	}
	// keyboard selection scrolls the row into view
	m.scI.Select(0)
	m.reveal()
	settle()
	if sm.pos != 0 {
		return fmt.Errorf("selecting the first row left the list at %.1f", sm.pos)
	}
	// a slow toggle must not block the UI thread
	a.opt.Ext = slowExt{}
	t0 := time.Now()
	m.onBtn(idToggle)
	if d := time.Since(t0); d > 50*time.Millisecond || !m.busy {
		return fmt.Errorf("toggle blocked the UI thread for %v (busy=%v)", d, m.busy)
	}
	time.Sleep(250 * time.Millisecond) // let the worker finish before teardown
	return nil
}

type slowExt struct{ manyExt }

func (slowExt) SetEnabled(string, bool) error { time.Sleep(200 * time.Millisecond); return nil }

type manyExt struct{ demoExt }

func (manyExt) List() []ExtInfo {
	var out []ExtInfo
	for i := 0; i < 60; i++ {
		out = append(out, ExtInfo{ID: fmt.Sprint("e", i), Name: fmt.Sprint("Extension number ", i, " with a fairly long display name"), Version: "1.2.3",
			Publisher: "someone", Enabled: i%2 == 0, Permissions: []string{"net", "fs.read", "clipboard", "notify"}})
	}
	return out
}

func renderPanel(p *panel) error {
	dc, bmp, old := offscreen(p.w, p.h)
	defer freeOffscreen(dc, bmp, old)
	p.render(dc)
	if got, want := getPixel(dc, p.w/2, p.h-p.s(2)-1), uint32(p.th.Colors.Background); got != want {
		return fmt.Errorf("panel %q: background pixel %06x, want %06x", p.title, got, want)
	}
	return nil
}

func offscreen(w, h int) (dc, bmp, old uintptr) {
	sdc := screenDC()
	dc, _, _ = pCreateCompatibleDC.Call(sdc)
	bmp, _, _ = pCreateCompatibleBitmap.Call(sdc, uintptr(w), uintptr(h))
	old, _, _ = pSelectObject.Call(dc, bmp)
	pReleaseDC.Call(0, sdc)
	return
}

func freeOffscreen(dc, bmp, old uintptr) {
	pSelectObject.Call(dc, old)
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(dc)
}

// getPixel returns the pixel as 0xRRGGBB.
func getPixel(dc uintptr, x, y int) uint32 {
	r, _, _ := pGetPixel.Call(dc, uintptr(x), uintptr(y))
	v := uint32(r)
	return (v&0xFF)<<16 | v&0xFF00 | (v>>16)&0xFF
}

type demoExt struct{}

func (demoExt) List() []ExtInfo {
	return []ExtInfo{{ID: "demo", Name: "Demo", Version: "1.0", Publisher: "someone", Enabled: true, Permissions: []string{"net", "fs.read"}}}
}
func (demoExt) SetEnabled(string, bool) error { return nil }
func (demoExt) Revoke(string, string) error   { return nil }
func (demoExt) Uninstall(string) error        { return nil }

type demoMarket struct{}

func (demoMarket) List(context.Context) ([]MarketEntry, error) { return nil, nil }
func (demoMarket) Install(context.Context, MarketEntry) error  { return nil }

type demoThemes struct{}

func (demoThemes) List() []ThemeInfo {
	return []ThemeInfo{{Name: "nord", Source: "built-in", Current: true}, {Name: "dusk", Source: "installed", Author: "someone"}}
}
func (demoThemes) Use(string) error { return nil }

var (
	pSendMessageW = user32.NewProc("SendMessageW")
)

// scopeTest checks the Tab scope: entering shows the chip, Backspace on the
// empty input and Esc leave it and restore the text typed before.
func (a *App) scopeTest() error {
	a.prepare(96)
	a.ed.SetText("note")
	a.items = []core.Item{{ID: "t:1", Title: "Notepad", Action: core.Action{Kind: "run", Target: "notepad.exe"},
		Alt: []core.NamedAction{{Name: "Open folder", Action: core.Action{Kind: "open", Target: `C:\Windows`}}}}}
	a.sc.Reset(1, a.m.VisibleRows(1))
	a.sc.Select(0)
	a.onTab()
	if a.scope == nil || a.scope.ID != "actions" || len(a.scope.Items) != 2 || a.ed.Text() != "" {
		return fmt.Errorf("scope: Tab on an item with actions must enter the Actions scope, got %+v text %q", a.scope, a.ed.Text())
	}
	w, h := a.m.Width, a.m.Height(0)
	dc, bmp, old := offscreen(w, h)
	a.render(dc, w, h)
	ir := a.m.InputRect()
	got := getPixel(dc, ir.L+a.m.Pad+3, ir.T+ir.H()/2)
	freeOffscreen(dc, bmp, old)
	if want := uint32(a.th.Colors.Accent); got != want {
		return fmt.Errorf("scope chip pixel %06x, want accent %06x", got, want)
	}
	if a.chipW == 0 {
		return fmt.Errorf("scope chip width not recorded")
	}
	a.onKeyDown(vkBack)
	if a.scope != nil || a.ed.Text() != "note" {
		return fmt.Errorf("scope: Backspace on empty input must leave the scope and restore %q, got scope %v text %q", "note", a.scope, a.ed.Text())
	}
	a.sc.Select(0)
	a.onTab()
	a.ed.SetText("op")
	a.onKeyDown(vkEscape)
	if a.scope != nil || a.ed.Text() != "note" || a.visible {
		return fmt.Errorf("scope: Esc must leave the scope first, got scope %v text %q", a.scope, a.ed.Text())
	}
	a.freeResources()
	return nil
}
