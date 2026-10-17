//go:build windows

package ui

import (
	"context"
	"fmt"
	"os"
	"runtime"

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

	// tray add/remove (informational: needs a shell)
	a.addTray()
	ok := a.trayAdded
	a.removeTray()
	fmt.Fprintf(os.Stderr, "selftest: tray icon added=%v\n", ok)
	return nil
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
