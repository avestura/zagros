//go:build windows

package ui

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/theme"
	"golang.org/x/sys/windows"
)

var (
	pGetDC     = user32.NewProc("GetDC")
	pReleaseDC = user32.NewProc("ReleaseDC")
)

const (
	swpShowWindow = 0x0040

	msgFuncs   = wmApp + 1 // run queued funcs
	msgResults = wmApp + 2 // query results ready (wParam = seq)
	msgTray    = wmApp + 3 // tray icon callback
	msgShow    = wmApp + 4 // show request (second instance)
	msgToggle  = wmApp + 5
	msgQuit    = wmApp + 6
	msgReload  = wmApp + 7
	msgIcons   = wmApp + 8 // icon worker finished something

	timerCaret = 1
	timerFade  = 2
	timerApply = 3 // coalesced result application
	timerIdle  = 4 // trim memory after a long idle period
)

// App is the launcher: one hidden-by-default window, the tray icon, the global
// hotkey and the dialogs. Everything except the exported thread-safe methods
// runs on the single UI thread started by Run.
type App struct {
	opt  Options
	hwnd windows.Handle

	mu    sync.Mutex
	queue []func()
	snap  latestSlot[snapMsg]

	quitCh chan struct{}
	once   sync.Once

	// UI-thread state.
	th         theme.Theme
	dpi        int
	m          Metrics
	fonts      fonts
	items      []core.Item
	sc         Scroll
	ed         Editor
	query      string
	qseq       uint64
	cancel     context.CancelFunc
	visible    bool
	caretOn    bool
	hscroll    int
	dragging   bool
	status     string
	back       backBuf
	icons      *iconLRU
	iconDC     uintptr
	loader     *iconLoader
	frameIcons []iconKey

	// streaming results
	gate       rateGate
	applyTimer bool
	userMoved  bool
	curH       int
	fontKey    string
	keyAt      time.Time
	qKeyAt     time.Time
	qFirst     bool
	qItems     bool
	lastDone   bool // the newest applied snapshot was final
	marks      []traceMark
	onApplied  func() // self-test hook
	inMenu     bool
	alpha      int
	fadeStep   int
	highSur    rune
	listDirty  bool // a key changed the selection; repaint the list after echoing
	skipInput  bool // render the list only; the input band is painted separately
	fastTyping bool // keys arrive faster than results should repaint the list

	taskbarMsg    uint32
	trayAdded     bool
	trayIcon      windows.Handle
	hotkeyErr     string
	trayStateText string
	roundRegion   bool

	noShow bool
	panels []*panel
	mgr    *manager
}

// New creates an App. Call Run on any goroutine; it locks its own OS thread.
func New(opt Options) *App {
	if opt.Highlight == nil {
		opt.Highlight = SubsequenceMatch
	}
	if opt.FadeMS > 100 {
		opt.FadeMS = 100
	}
	return &App{opt: opt, quitCh: make(chan struct{}), icons: newIconLRU(iconCacheCap, func(h uintptr) { pDeleteObject.Call(h) }), gate: rateGate{min: applyInterval}}
}

func (a *App) theme() theme.Theme {
	if a.opt.Theme != nil {
		return NormalizeTheme(a.opt.Theme())
	}
	return NormalizeTheme(theme.Theme{Colors: theme.Colors{
		Background: 0x1b1b1b, Border: 0x393939, InputBg: 0x262626, InputText: 0xf4f4f4,
		Placeholder: 0x8d8d8d, Text: 0xf4f4f4, TextDim: 0x8d8d8d, SelBg: 0x0f62fe,
		SelText: 0xffffff, SelTextDim: 0xd0e2ff, Accent: 0x78a9ff, Tag: 0xc6c6c6, TagBg: 0x393939,
	}, ShowIcons: true, ShowTags: true, Dark: true})
}

// Run creates the window and tray icon and pumps messages until Quit.
func (a *App) Run() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := a.create(); err != nil {
		return err
	}
	a.registerHotkey()
	a.addTray()
	time.AfterFunc(8*time.Second, trimMemory)
	a.startBackground()
	if a.opt.OnReady != nil {
		go a.opt.OnReady()
	}
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

func (a *App) create() error {
	setDPIAware()
	a.hwnd = createWindow(classLauncher, "Zagros", wsPopup, wsExToolWindow|wsExTopmost|wsExLayered, 0, 0, 1, 1, a)
	if a.hwnd == 0 {
		return errors.New("ui: CreateWindowEx failed")
	}
	pSetLayeredWindowAttrs.Call(uintptr(a.hwnd), 0, 255, lwaAlpha)
	r, _, _ := pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(u16("TaskbarCreated"))))
	a.taskbarMsg = uint32(r)
	return nil
}

// ---- thread-safe API -------------------------------------------------

// Show asks the launcher to appear.
func (a *App) Show() { postMessage(a.hwnd, msgShow, 0, 0) }

// Toggle shows or hides the launcher.
func (a *App) Toggle() { postMessage(a.hwnd, msgToggle, 0, 0) }

// Quit ends Run.
func (a *App) Quit() {
	a.once.Do(func() { close(a.quitCh) })
	postMessage(a.hwnd, msgQuit, 0, 0)
}

// Notify shows a tray balloon notification (for extension notify()).
func (a *App) Notify(title, text string) {
	a.Post(func() { a.balloon(title, text) })
}

// SetTrayState updates the tray tooltip text suffix, e.g. "3 extensions".
func (a *App) SetTrayState(state string) {
	a.Post(func() { a.trayStateText = state; a.updateTray() })
}

// Post runs fn on the UI thread.
func (a *App) Post(fn func()) {
	a.mu.Lock()
	a.queue = append(a.queue, fn)
	a.mu.Unlock()
	postMessage(a.hwnd, msgFuncs, 0, 0)
}

// ---- messages ----------------------------------------------------------

func (a *App) wndProc(hwnd windows.Handle, m uint32, wp, lp uintptr) uintptr {
	switch m {
	case wmPaint:
		a.paint()
		return 0
	case wmEraseBkgnd:
		return 1
	case wmKeyDown, wmSysKeyDown:
		if a.visible {
			a.onKeyDown(int(wp))
			return 0
		}
	case wmChar:
		if a.visible {
			a.onChar(rune(wp))
		}
		return 0
	case wmSysCommand:
		if wp&0xFFF0 == 0xF100 { // SC_KEYMENU: Alt/Alt+Space must not open the system menu
			return 0
		}
	case wmLButtonDown:
		a.onMouseDown(signedX(lp), signedY(lp))
		return 0
	case wmLButtonDbl:
		a.onMouseDbl(signedX(lp), signedY(lp))
		return 0
	case wmLButtonUp:
		if a.dragging {
			a.dragging = false
			pReleaseCapture.Call()
		}
		return 0
	case wmMouseMove:
		a.onMouseMove(signedX(lp), signedY(lp), wp&1 != 0)
		return 0
	case wmSetCursor:
		if lp&0xFFFF == 1 {
			return 1
		}
	case wmMouseWheel:
		d := int(int16(wp>>16)) / 120
		a.sc.Wheel(-d * 3)
		a.invalidateList()
		return 0
	case wmTimer:
		a.onTimer(wp)
		return 0
	case wmActivate:
		if wp&0xFFFF == 0 && a.visible && !a.inMenu {
			a.hide()
		}
		return 0
	case wmHotkey:
		a.toggle()
		return 0
	case msgFuncs:
		a.runQueue()
		return 0
	case msgResults:
		a.scheduleApply()
		return 0
	case msgIcons:
		a.onIconsReady()
		return 0
	case msgTray:
		a.onTray(uint32(lp & 0xFFFF))
		return 0
	case msgShow:
		a.showLauncher()
		return 0
	case msgToggle:
		a.toggle()
		return 0
	case msgQuit:
		a.shutdown()
		return 0
	case msgReload:
		a.runTrayCommand(cmdReload)
		return 0
	case wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	if m == a.taskbarMsg && m != 0 {
		a.trayAdded = false
		a.addTray()
		return 0
	}
	return defProc(hwnd, m, wp, lp)
}

func (a *App) runQueue() {
	a.mu.Lock()
	q := a.queue
	a.queue = nil
	a.mu.Unlock()
	for _, f := range q {
		f()
	}
}

func (a *App) shutdown() {
	if a.opt.OnQuit != nil {
		a.opt.OnQuit()
	}
	a.hide()
	if a.loader != nil {
		a.loader.Close()
		a.loader = nil
	}
	for _, p := range append([]*panel(nil), a.panels...) {
		p.close()
	}
	pUnregisterHotKey.Call(uintptr(a.hwnd), 1)
	a.removeTray()
	pDestroyWindow.Call(uintptr(a.hwnd))
}

func (a *App) registerHotkey() {
	a.hotkeyErr = ""
	pUnregisterHotKey.Call(uintptr(a.hwnd), 1)
	hk := a.opt.Config.Hotkey
	if hk == "" {
		hk = "Alt+Space"
	}
	mods, vk, err := ParseHotkey(hk)
	if err != nil {
		a.hotkeyErr = err.Error()
		return
	}
	r, _, e := pRegisterHotKey.Call(uintptr(a.hwnd), 1, uintptr(mods|modNoRepeat), uintptr(vk))
	if r == 0 {
		a.hotkeyErr = "hotkey " + hk + " is unavailable: " + e.Error()
	}
}

// SetConfig applies a new configuration (hotkey, max results). Thread-safe.
func (a *App) SetConfig(c core.Config) {
	a.Post(func() {
		a.opt.Config = c
		a.registerHotkey()
		a.updateTray()
		if a.hotkeyErr != "" {
			a.balloon("Zagros", a.hotkeyErr)
		}
	})
}

const modNoRepeat = 0x4000
