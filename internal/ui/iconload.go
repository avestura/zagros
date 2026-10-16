//go:build windows

package ui

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
	"golang.org/x/sys/windows"
)

var (
	msimg32       = windows.NewLazySystemDLL("msimg32.dll")
	pAlphaBlend   = msimg32.NewProc("AlphaBlend")
	pCreateDIBSec = gdi32.NewProc("CreateDIBSection")
	pGetDriveType = kernel32.NewProc("GetDriveTypeW")
)

const (
	shgfiUseFileAttributes = 0x10
	fileAttributeNormal    = 0x80
	driveRemote            = 4

	iconQueueMax = 24
	iconCacheCap = 128
	iconIdleKeep = 32
)

// iconLoader loads shell icons on one background goroutine so that painting
// never calls into the shell. Results are premultiplied 32-bit DIB sections
// that are blitted with AlphaBlend. The UI thread owns the cache (iconLRU);
// the worker only appends finished bitmaps to a hand-off list.
type iconLoader struct {
	hwnd   windows.Handle
	q      *iconQueue
	wake   chan struct{}
	stop   chan struct{}
	mu     sync.Mutex
	done   []iconResult
	posted atomic.Bool
}

type iconResult struct {
	key iconKey
	bmp uintptr // 0: no icon
}

func newIconLoader(hwnd windows.Handle) *iconLoader {
	l := &iconLoader{hwnd: hwnd, q: newIconQueue(iconQueueMax), wake: make(chan struct{}, 1), stop: make(chan struct{})}
	go l.run()
	return l
}

func (l *iconLoader) Close() { close(l.stop) }

// Request queues k; it never blocks.
func (l *iconLoader) Request(k iconKey) {
	if l.q.Push(k) {
		select {
		case l.wake <- struct{}{}:
		default:
		}
	}
}

// Drain returns the finished icons since the last call.
func (l *iconLoader) Drain() []iconResult {
	l.mu.Lock()
	out := l.done
	l.done = nil
	l.mu.Unlock()
	l.posted.Store(false)
	return out
}

func (l *iconLoader) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	_ = windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED)
	defer windows.CoUninitialize()
	warmShellIcons()
	for {
		select {
		case <-l.stop:
			return
		case <-l.wake:
		}
		for {
			k, ok := l.q.Pop()
			if !ok {
				break
			}
			t := time.Now()
			bmp := loadIconBitmap(k)
			core.Trace.Event("icon_load", time.Since(t), "%s size=%d ok=%v", k.Path, k.Size, bmp != 0)
			l.q.Done(k)
			l.mu.Lock()
			l.done = append(l.done, iconResult{k, bmp})
			l.mu.Unlock()
			if l.posted.CompareAndSwap(false, true) {
				postMessage(l.hwnd, msgIcons, 0, 0)
			}
		}
	}
}

// warmShellIcons pays the one-time shell icon initialisation cost off the UI
// thread.
func warmShellIcons() {
	var sfi shFileInfo
	pSHGetFileInfoW.Call(uintptr(unsafe.Pointer(u16(".txt"))), fileAttributeNormal,
		uintptr(unsafe.Pointer(&sfi)), unsafe.Sizeof(sfi), shgfiIcon|shgfiLargeIcon|shgfiUseFileAttributes)
	if sfi.Icon != 0 {
		pDestroyIcon.Call(uintptr(sfi.Icon))
	}
}

// specificIcon reports whether a file's icon depends on its contents, so the
// shell has to open it (as opposed to a per-extension icon).
func specificIcon(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case "", ".exe", ".lnk", ".ico", ".dll", ".url", ".scr", ".cpl", ".msc":
		return true
	}
	return false
}

// slowPath reports whether touching path may block: UNC shares and network
// drives. Their icons come from the file extension only.
func slowPath(path string) bool {
	if strings.HasPrefix(path, `\\`) && !strings.HasPrefix(path, `\\?\`) && !strings.HasPrefix(path, `\\.\`) {
		return true
	}
	if len(path) >= 2 && path[1] == ':' {
		root := path[:2] + `\`
		if t, _, _ := pGetDriveType.Call(uintptr(unsafe.Pointer(u16(root)))); t == driveRemote {
			return true
		}
	}
	return false
}

func shellIcon(path string) windows.Handle {
	get := func(extra uintptr) windows.Handle {
		var sfi shFileInfo
		r, _, _ := pSHGetFileInfoW.Call(uintptr(unsafe.Pointer(u16(path))), fileAttributeNormal,
			uintptr(unsafe.Pointer(&sfi)), unsafe.Sizeof(sfi), shgfiIcon|shgfiLargeIcon|extra)
		if r == 0 {
			return 0
		}
		return sfi.Icon
	}
	if slowPath(path) || !specificIcon(path) {
		return get(shgfiUseFileAttributes)
	}
	if h := get(0); h != 0 {
		return h
	}
	return get(shgfiUseFileAttributes) // missing file: the generic icon for its type
}

// loadIconBitmap returns a premultiplied-alpha size x size DIB section with
// the file's icon, or 0.
func loadIconBitmap(k iconKey) uintptr {
	h := shellIcon(k.Path)
	if h == 0 {
		return 0
	}
	defer pDestroyIcon.Call(uintptr(h))
	return iconToBitmap(uintptr(h), k.Size)
}

func newDIB(ref uintptr, sz int) (bmp uintptr, px []byte) {
	bi := struct {
		Size                   uint32
		W, H                   int32
		Planes, Bits           uint16
		Compression, ImageSize uint32
		XPels, YPels           int32
		ClrUsed, ClrImportant  uint32
	}{Size: 40, W: int32(sz), H: -int32(sz), Planes: 1, Bits: 32}
	var bits uintptr
	bmp, _, _ = pCreateDIBSec.Call(ref, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	if bmp == 0 || bits == 0 {
		return 0, nil
	}
	return bmp, unsafe.Slice((*byte)(*(*unsafe.Pointer)(unsafe.Pointer(&bits))), sz*sz*4)
}

// iconToBitmap renders the icon onto black and onto white; the difference
// gives the alpha channel and the black rendering is the premultiplied colour.
func iconToBitmap(hicon uintptr, sz int) uintptr {
	if sz <= 0 {
		return 0
	}
	ref := screenDC()
	defer pReleaseDC.Call(0, ref)
	dc, _, _ := pCreateCompatibleDC.Call(ref)
	defer pDeleteDC.Call(dc)

	onBlack, black := newDIB(ref, sz)
	onWhite, white := newDIB(ref, sz)
	if onBlack == 0 || onWhite == 0 {
		if onBlack != 0 {
			pDeleteObject.Call(onBlack)
		}
		if onWhite != 0 {
			pDeleteObject.Call(onWhite)
		}
		return 0
	}
	defer pDeleteObject.Call(onWhite)
	for i := range black {
		black[i] = 0
		white[i] = 0xFF
	}
	old, _, _ := pSelectObject.Call(dc, onBlack)
	pDrawIconEx.Call(dc, 0, 0, hicon, uintptr(sz), uintptr(sz), 0, 0, diNormal)
	pSelectObject.Call(dc, onWhite)
	pDrawIconEx.Call(dc, 0, 0, hicon, uintptr(sz), uintptr(sz), 0, 0, diNormal)
	pSelectObject.Call(dc, old)
	opaque := false
	for i := 0; i+3 < len(black); i += 4 {
		d := (int(white[i]) - int(black[i]) + int(white[i+1]) - int(black[i+1]) + int(white[i+2]) - int(black[i+2])) / 3
		a := 255 - d
		if a < 0 {
			a = 0
		}
		for c := 0; c < 3; c++ {
			if int(black[i+c]) > a {
				black[i+c] = byte(a)
			}
		}
		black[i+3] = byte(a)
		if a != 0 {
			opaque = true
		}
	}
	if !opaque {
		pDeleteObject.Call(onBlack)
		return 0
	}
	return onBlack
}

// ---- UI-thread side ---------------------------------------------------------

// blitIcon draws the cached bitmap bmp at (x,y).
func (a *App) blitIcon(dc, bmp uintptr, x, y, sz int) {
	if a.iconDC == 0 {
		a.iconDC, _, _ = pCreateCompatibleDC.Call(dc)
	}
	old, _, _ := pSelectObject.Call(a.iconDC, bmp)
	const blend = 0x01000000 | 255<<16 // AC_SRC_OVER, SourceConstantAlpha 255, AC_SRC_ALPHA
	pAlphaBlend.Call(dc, uintptr(x), uintptr(y), uintptr(sz), uintptr(sz), a.iconDC, 0, 0, uintptr(sz), uintptr(sz), blend)
	pSelectObject.Call(a.iconDC, old)
}

// onIconsReady takes finished icons into the cache and repaints.
func (a *App) onIconsReady() {
	if a.loader == nil {
		return
	}
	res := a.loader.Drain()
	for _, r := range res {
		a.icons.Put(r.key, r.bmp)
	}
	if a.visible && len(res) > 0 {
		invalidate(a.hwnd)
	}
}
