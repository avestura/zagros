//go:build windows

package ui

import (
	"fmt"
	"golang.org/x/sys/windows"
	"image"
	"image/png"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"sync"
	"time"
	"unsafe"
)

// This file is the real-world typing benchmark (zagros --bench-input). It runs
// the full launcher (real window, real providers and indexes) and types into
// it with real key messages, measuring how long the UI thread takes to echo
// each key and which messages keep it busy.

type procStat struct {
	n          int
	total, max time.Duration
}

type benchState struct {
	mu      sync.Mutex
	echo    []time.Time
	apply   []time.Time
	resizes int
	proc    map[uint32]*procStat
}

func newBench() *benchState { return &benchState{proc: map[uint32]*procStat{}} }

func (b *benchState) reset() {
	b.mu.Lock()
	b.echo, b.apply, b.resizes = nil, nil, 0
	b.proc = map[uint32]*procStat{}
	b.mu.Unlock()
}

func (b *benchState) noteEcho()  { b.mu.Lock(); b.echo = append(b.echo, time.Now()); b.mu.Unlock() }
func (b *benchState) noteApply() { b.mu.Lock(); b.apply = append(b.apply, time.Now()); b.mu.Unlock() }
func (b *benchState) recProc(m uint32, d time.Duration) {
	b.mu.Lock()
	s := b.proc[m]
	if s == nil {
		s = &procStat{}
		b.proc[m] = s
	}
	s.n++
	s.total += d
	s.max = max(s.max, d)
	b.mu.Unlock()
}

func msgName(m uint32) string {
	switch m {
	case wmPaint:
		return "WM_PAINT"
	case wmChar:
		return "WM_CHAR"
	case wmKeyDown:
		return "WM_KEYDOWN"
	case wmTimer:
		return "WM_TIMER"
	case msgResults:
		return "msgResults"
	case msgIcons:
		return "msgIcons"
	case msgFuncs:
		return "msgFuncs"
	}
	return fmt.Sprintf("0x%X", m)
}

func pct(d []time.Duration, p float64) time.Duration {
	if len(d) == 0 {
		return 0
	}
	i := int(float64(len(d)-1) * p)
	return d[i]
}

type inputRec struct {
	Type uint32
	_    uint32
	KI   struct {
		VK, Scan uint16
		Flags    uint32
		Time     uint32
		Extra    uintptr
	}
	_ [8]byte
}

var pSendInput = user32.NewProc("SendInput")

func sendUnicode(r rune) {
	const keyUnicode, keyUp = 0x0004, 0x0002
	var in [2]inputRec
	for i := range in {
		in[i].Type = 1
		in[i].KI.Scan = uint16(r)
		in[i].KI.Flags = keyUnicode
	}
	in[1].KI.Flags |= keyUp
	pSendInput.Call(2, uintptr(unsafe.Pointer(&in[0])), unsafe.Sizeof(in[0]))
}

// RunInputBench types several strings into the running launcher and writes a
// latency report to out. ready blocks until providers have loaded their
// indexes. It quits the app when done. Call it from a goroutine.
func (a *App) RunInputBench(out io.Writer, ready func()) {
	defer a.Quit()
	b := newBench()
	a.Post(func() { a.bench = b })
	ready()
	time.Sleep(300 * time.Millisecond)
	a.Post(a.showLauncher)
	time.Sleep(700 * time.Millisecond)

	// Keep the launcher above other (e.g. always-on-top terminal) windows so the
	// screen sampler sees it.
	stopTop := make(chan struct{})
	defer close(stopTop)
	go func() {
		t := time.NewTicker(15 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stopTop:
				return
			case <-t.C:
				pSetWindowPos.Call(uintptr(a.hwnd), hwndTopmost, 0, 0, 0, 0, 0x0001|0x0002|0x0010)
			}
		}
	}()
	fg := make(chan bool, 1)
	a.Post(func() { r, _, _ := pGetForegroundWindow.Call(); fg <- r == uintptr(a.hwnd) })
	haveFocus := <-fg
	dch := make(chan string, 1)
	a.Post(func() {
		// Bench-only: a console-started process may not get to the top of the z-order.
		pSetWindowPos.Call(uintptr(a.hwnd), hwndTopmost, 0, 0, 0, 0, 0x0001|0x0002|0x0010) // NOSIZE|NOMOVE|NOACTIVATE
		pBringWindowToTop.Call(uintptr(a.hwnd))
		dch <- a.describeWindow()
	})
	fmt.Fprintln(out, "window:", <-dch)
	if dir := os.Getenv("ZAGROS_BENCH_SHOT"); dir != "" {
		whole := newRegion(0, 0, 1920, 1080)
		whole.savePNG(dir + "/start.png")
		whole.free()
		pw := newRegion(0, 0, 800, 67)
		pPrintWindow.Call(uintptr(a.hwnd), pw.mem, 2)
		pw.savePNGNoGrab(dir + "/printwindow.png")
		pw.free()
	}
	fmt.Fprintf(out, "launcher has foreground focus: %v\n", haveFocus)

	type scenario struct {
		name, text string
		gap        time.Duration
		input      bool
		cold       bool // hide, trim memory like the idle timer, then show again
	}
	sc := []scenario{
		{"post  70ms  'chrome'", "chrome", 70 * time.Millisecond, false, false},
		{"post  70ms  'visual studio code'", "visual studio code", 70 * time.Millisecond, false, false},
		{"post  25ms  'notepad calc settings'", "notepad calc settings", 25 * time.Millisecond, false, false},
		{"post  8ms   'windows terminal powershell'", "windows terminal powershell", 8 * time.Millisecond, false, false},
	}
	sc = append(sc,
		scenario{"COLD after idle trim  70ms 'chrome'", "chrome", 70 * time.Millisecond, false, true},
		scenario{"COLD after idle trim  70ms 'visual studio code'", "visual studio code", 70 * time.Millisecond, false, true})
	if haveFocus {
		sc = append(sc,
			scenario{"input 70ms  'chrome' (SendInput)", "chrome", 70 * time.Millisecond, true, false},
			scenario{"input 25ms  'notepad calc settings' (SendInput)", "notepad calc settings", 25 * time.Millisecond, true, false})
	}
	// Screen geometry, read on the UI thread; the caret blink is switched off so
	// only typing changes pixels.
	type geo struct{ inx, iny, inw, inh, lx, ly, lw, lh int }
	gch := make(chan geo, 1)
	a.Post(func() {
		pKillTimer.Call(uintptr(a.hwnd), timerCaret)
		a.caretOn = false
		var wr rect
		pGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&wr)))
		ir := a.m.InputRect()
		gch <- geo{int(wr.Left) + ir.L, int(wr.Top) + ir.T, ir.R - ir.L, ir.B - ir.T,
			int(wr.Left) + ir.L, int(wr.Top) + ir.B + 2, ir.R - ir.L, a.m.Height(a.m.MaxRows) - ir.B - 2}
	})
	g := <-gch
	worst, worstScreen := time.Duration(0), time.Duration(0)
	for _, s := range sc {
		if s.cold {
			a.Post(a.hide)
			time.Sleep(200 * time.Millisecond)
			a.Post(a.trimIdle) // what the idle timer does (icons, back buffer, memory)
			time.Sleep(1500 * time.Millisecond)
			a.Post(a.showLauncher)
			time.Sleep(300 * time.Millisecond)
		} else {
			a.Post(func() { a.ed.Clear(); a.startQuery(); a.refreshInput() })
			time.Sleep(500 * time.Millisecond)
		}
		b.reset()
		var gcBefore, gcAfter debug.GCStats
		debug.ReadGCStats(&gcBefore)
		inR, liR := newRegion(g.inx, g.iny, g.inw, g.inh), newRegion(g.lx, g.ly, g.lw, g.lh)
		watch := startWatch(inR, liR)
		var sent []time.Time
		for _, r := range s.text {
			sent = append(sent, time.Now())
			if s.input {
				sendUnicode(r)
			} else {
				postMessage(a.hwnd, wmChar, uintptr(r), 0)
			}
			time.Sleep(s.gap)
		}
		time.Sleep(700 * time.Millisecond)
		debug.ReadGCStats(&gcAfter)
		d2 := make(chan string, 1)
		a.Post(func() { d2 <- a.describeWindow() })
		fmt.Fprintln(out, "window at end:", <-d2)
		inEv, liEv := watch.finish()
		if dir := os.Getenv("ZAGROS_BENCH_SHOT"); dir != "" {
			fmt.Fprintf(out, "geometry %+v\n", g)
			whole := newRegion(0, 0, 2560, 1440)
			whole.savePNG(fmt.Sprintf("%s/screen.png", dir))
			whole.free()
			full := newRegion(g.inx, g.iny, g.inw, g.lh+g.inh+4)
			full.savePNG(fmt.Sprintf("%s/shot-%d.png", dir, len(s.text)))
			full.free()
		}
		inR.free()
		liR.free()

		b.mu.Lock()
		echo := append([]time.Time(nil), b.echo...)
		apply := len(b.apply)
		resizes := b.resizes
		procs := map[uint32]procStat{}
		for k, v := range b.proc {
			procs[k] = *v
		}
		b.mu.Unlock()
		var lat []time.Duration
		for i := range sent {
			if i < len(echo) {
				lat = append(lat, echo[i].Sub(sent[i]))
			}
		}
		sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
		fmt.Fprintf(out, "\n== %s ==\n", s.name)
		fmt.Fprintf(out, "keys sent %d, echoed %d, list updates applied %d, window resizes %d\n", len(sent), len(echo), apply, resizes)
		fmt.Fprintf(out, "key->input painted: p50 %v  p95 %v  max %v\n", pct(lat, .5).Round(10*time.Microsecond), pct(lat, .95).Round(10*time.Microsecond), pct(lat, 1).Round(10*time.Microsecond))
		worst = max(worst, pct(lat, 1))
		// what the screen showed
		var slat []time.Duration
		for i, t := range sent {
			if e, ok := firstAfter(inEv, t); ok {
				d := e.Sub(t)
				if i+1 < len(sent) && e.After(sent[i+1]) {
					continue // belongs to a later key
				}
				slat = append(slat, d)
			}
		}
		sort.Slice(slat, func(i, j int) bool { return slat[i] < slat[j] })
		fmt.Fprintf(out, "SCREEN key->input visible: seen %d/%d  p50 %v  p95 %v  max %v\n", len(slat), len(sent),
			pct(slat, .5).Round(100*time.Microsecond), pct(slat, .95).Round(100*time.Microsecond), pct(slat, 1).Round(100*time.Microsecond))
		worstScreen = max(worstScreen, pct(slat, 1))
		if len(liEv) > 0 {
			first := time.Duration(-1)
			if e, ok := firstAfter(liEv, sent[0]); ok {
				first = e.Sub(sent[0])
			}
			last := liEv[len(liEv)-1].Sub(sent[len(sent)-1])
			fmt.Fprintf(out, "SCREEN list: %d changes, first %v after first key, settled %v after last key\n", len(liEv), first.Round(time.Millisecond), last.Round(time.Millisecond))
		} else {
			fmt.Fprintf(out, "SCREEN list: no visible change\n")
		}
		var keys []uint32
		for k := range procs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return procs[keys[i]].max > procs[keys[j]].max })
		for i, k := range keys {
			if i == 4 {
				break
			}
			p := procs[k]
			fmt.Fprintf(out, "  UI thread %-10s n=%-4d avg %-8v max %v\n", msgName(k), p.n, (p.total / time.Duration(p.n)).Round(10*time.Microsecond), p.max.Round(10*time.Microsecond))
		}
		var maxPause time.Duration
		for _, p := range gcAfter.Pause[:min(len(gcAfter.Pause), int(gcAfter.NumGC-gcBefore.NumGC))] {
			maxPause = max(maxPause, p)
		}
		fmt.Fprintf(out, "  GC cycles %d, max pause %v\n", gcAfter.NumGC-gcBefore.NumGC, maxPause)
	}
	if dir := os.Getenv("ZAGROS_BENCH_SHOT"); dir != "" {
		a.scopeShot(dir)
	}
	fmt.Fprintf(out, "\nworst key->input painted over all scenarios: %v\n", worst.Round(10*time.Microsecond))
}

// ---- screen sampling: what is actually visible, after the compositor ----------

var pGetDIBits = gdi32.NewProc("GetDIBits")

type bmiHeader struct {
	Size          uint32
	Width, Height int32
	Planes, Bits  uint16
	Compression   uint32
	SizeImage     uint32
	XPels, YPels  int32
	ClrUsed, ClrI uint32
}

type screenRegion struct {
	x, y, w, h    int
	mem, bmp, old uintptr
	buf           []byte
}

func newRegion(x, y, w, h int) *screenRegion {
	sdc := screenDC()
	defer pReleaseDC.Call(0, sdc)
	r := &screenRegion{x: x, y: y, w: w, h: h, buf: make([]byte, w*h*4)}
	r.mem, _, _ = pCreateCompatibleDC.Call(sdc)
	r.bmp, _, _ = pCreateCompatibleBitmap.Call(sdc, uintptr(w), uintptr(h))
	r.old, _, _ = pSelectObject.Call(r.mem, r.bmp)
	return r
}

func (r *screenRegion) free() {
	pSelectObject.Call(r.mem, r.old)
	pDeleteObject.Call(r.bmp)
	pDeleteDC.Call(r.mem)
}

// hash grabs the region from the composed desktop and returns a fingerprint.
func (r *screenRegion) hash() uint64 {
	sdc := screenDC()
	pBitBlt.Call(r.mem, 0, 0, uintptr(r.w), uintptr(r.h), sdc, uintptr(r.x), uintptr(r.y), srcCopy|0x40000000)
	pReleaseDC.Call(0, sdc)
	bi := bmiHeader{Size: 40, Width: int32(r.w), Height: -int32(r.h), Planes: 1, Bits: 32}
	pGetDIBits.Call(r.mem, r.bmp, 0, uintptr(r.h), uintptr(unsafe.Pointer(&r.buf[0])), uintptr(unsafe.Pointer(&bi)), 0)
	h := uint64(14695981039346656037)
	for _, c := range r.buf {
		h = (h ^ uint64(c)) * 1099511628211
	}
	return h
}

type screenWatch struct {
	mu          sync.Mutex
	input, list []time.Time
	stop        chan struct{}
	done        chan struct{}
}

func startWatch(in, li *screenRegion) *screenWatch {
	w := &screenWatch{stop: make(chan struct{}), done: make(chan struct{})}
	hi, hl := in.hash(), li.hash()
	go func() {
		defer close(w.done)
		for {
			select {
			case <-w.stop:
				return
			default:
			}
			if n := in.hash(); n != hi {
				hi = n
				w.mu.Lock()
				w.input = append(w.input, time.Now())
				w.mu.Unlock()
			}
			if n := li.hash(); n != hl {
				hl = n
				w.mu.Lock()
				w.list = append(w.list, time.Now())
				w.mu.Unlock()
			}
		}
	}()
	return w
}

func (w *screenWatch) finish() (input, list []time.Time) {
	close(w.stop)
	<-w.done
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.input, w.list
}

// firstAfter returns the first event at or after t.
func firstAfter(ev []time.Time, t time.Time) (time.Time, bool) {
	for _, e := range ev {
		if !e.Before(t) {
			return e, true
		}
	}
	return time.Time{}, false
}

// savePNG writes the region as a PNG (debug aid for --bench-input).
func (r *screenRegion) savePNG(path string) {
	r.hash()
	img := image.NewNRGBA(image.Rect(0, 0, r.w, r.h))
	for i := 0; i < r.w*r.h; i++ {
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = r.buf[i*4+2], r.buf[i*4+1], r.buf[i*4], 255
	}
	if f, err := os.Create(path); err == nil {
		_ = png.Encode(f, img)
		f.Close()
	}
}

var (
	pBringWindowToTop      = user32.NewProc("BringWindowToTop")
	pDwmGetWindowAttribute = windows.NewLazySystemDLL("dwmapi.dll").NewProc("DwmGetWindowAttribute")
	pWindowFromPoint       = user32.NewProc("WindowFromPoint")
	pGetClassNameW         = user32.NewProc("GetClassNameW")
	pIsWindowVisibleB      = user32.NewProc("IsWindowVisible")
	pPrintWindow           = user32.NewProc("PrintWindow")
	pGetLayeredAttrs       = user32.NewProc("GetLayeredWindowAttributes")
)

// describeWindow reports the real window state (UI thread only).
func (a *App) describeWindow() string {
	vis, _, _ := pIsWindowVisibleB.Call(uintptr(a.hwnd))
	var wr rect
	pGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&wr)))
	var key, alpha, flags uint32
	pGetLayeredAttrs.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&key)), uintptr(unsafe.Pointer(&alpha)), uintptr(unsafe.Pointer(&flags)))
	var cloaked uint32
	r, _, _ := pDwmGetWindowAttribute.Call(uintptr(a.hwnd), 14, uintptr(unsafe.Pointer(&cloaked)), 4)
	at, _, _ := pWindowFromPoint.Call(uintptr(int64(wr.Left+20) | int64(wr.Top+20)<<32))
	var cls [64]uint16
	pGetClassNameW.Call(at, uintptr(unsafe.Pointer(&cls[0])), 64)
	return fmt.Sprintf("a.visible=%v IsWindowVisible=%d rect=%v alpha(attr)=%d a.alpha=%d curH=%d cloaked=%d(ret %d) windowAtPoint=%#x(%s) self=%#x", a.visible, vis, wr, alpha, a.alpha, a.curH, cloaked, r, at, windows.UTF16ToString(cls[:]), a.hwnd)
}

func (r *screenRegion) savePNGNoGrab(path string) {
	bi := bmiHeader{Size: 40, Width: int32(r.w), Height: -int32(r.h), Planes: 1, Bits: 32}
	pGetDIBits.Call(r.mem, r.bmp, 0, uintptr(r.h), uintptr(unsafe.Pointer(&r.buf[0])), uintptr(unsafe.Pointer(&bi)), 0)
	img := image.NewNRGBA(image.Rect(0, 0, r.w, r.h))
	for i := 0; i < r.w*r.h; i++ {
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = r.buf[i*4+2], r.buf[i*4+1], r.buf[i*4], 255
	}
	if f, err := os.Create(path); err == nil {
		_ = png.Encode(f, img)
		f.Close()
	}
}

// scopeShot types a query, selects the web-search row, presses Tab and saves
// a picture of the launcher inside the Web search scope.
func (a *App) scopeShot(dir string) {
	a.Post(func() { a.ed.SetText("something"); a.startQuery(); a.refreshInput() })
	time.Sleep(600 * time.Millisecond)
	a.Post(func() {
		a.sc.End()
		a.onTab()
	})
	time.Sleep(600 * time.Millisecond)
	done := make(chan struct{})
	a.Post(func() {
		var wr rect
		pGetWindowRect.Call(uintptr(a.hwnd), uintptr(unsafe.Pointer(&wr)))
		w, h := int(wr.Right-wr.Left), int(wr.Bottom-wr.Top)
		pw := newRegion(0, 0, w, h)
		pPrintWindow.Call(uintptr(a.hwnd), pw.mem, 2)
		pw.savePNGNoGrab(dir + "/scope.png")
		pw.free()
		close(done)
	})
	<-done
}
