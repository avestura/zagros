//go:build windows

package ui

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// handler receives the messages of one native window. All windows live on the
// single UI thread, so the registry needs no locking.
type handler interface {
	wndProc(hwnd windows.Handle, m uint32, wp, lp uintptr) uintptr
}

var (
	handlers  = map[windows.Handle]handler{}
	creating  handler
	classOnce sync.Once
	procPtr   uintptr
)

const (
	classLauncher = "ZagrosLauncher"
	classPanel    = "ZagrosPanel"
)

func globalProc(hwnd, m, wp, lp uintptr) uintptr {
	h := windows.Handle(hwnd)
	if m == 0x0081 && creating != nil { // WM_NCCREATE
		handlers[h] = creating
	}
	if hd, ok := handlers[h]; ok {
		r := hd.wndProc(h, uint32(m), wp, lp)
		if m == 0x0082 { // WM_NCDESTROY
			delete(handlers, h)
		}
		return r
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
	return r
}

func registerClasses() {
	classOnce.Do(func() {
		procPtr = windows.NewCallback(globalProc)
		cur, _, _ := pLoadCursorW.Call(0, idcArrow)
		for _, name := range []string{classLauncher, classPanel} {
			wc := wndClassEx{
				Size:      uint32(unsafe.Sizeof(wndClassEx{})),
				Style:     0x0008, // CS_DBLCLKS
				WndProc:   procPtr,
				Instance:  hinstance(),
				Cursor:    windows.Handle(cur),
				ClassName: u16(name),
			}
			pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		}
	})
}

func createWindow(class, title string, style, ex uint32, x, y, w, h int, hd handler) windows.Handle {
	registerClasses()
	creating = hd
	r, _, _ := pCreateWindowExW.Call(
		uintptr(ex), uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(title))),
		uintptr(style), uintptr(int32(x)), uintptr(int32(y)), uintptr(int32(w)), uintptr(int32(h)),
		0, 0, uintptr(hinstance()), 0)
	creating = nil
	return windows.Handle(r)
}

func defProc(hwnd windows.Handle, m uint32, wp, lp uintptr) uintptr {
	r, _, _ := pDefWindowProcW.Call(uintptr(hwnd), uintptr(m), wp, lp)
	return r
}

func invalidate(h windows.Handle) { pInvalidateRect.Call(uintptr(h), 0, 0) }

func setDPIAware() {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 == -4
	if pSetProcessDpiAwareCtx.Find() == nil {
		v := ^uintptr(3)
		pSetProcessDpiAwareCtx.Call(v)
	}
}

// monitorAtCursor returns the work area and effective DPI of the monitor that
// holds the mouse cursor.
func monitorAtCursor() (work rect, dpi int) {
	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	mon, _, _ := pMonitorFromPoint.Call(uintptr(*(*int64)(unsafe.Pointer(&pt))), monitorDefaultToNearest)
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	pGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	work = mi.Work
	dpi = 96
	if pGetDpiForMonitor.Find() == nil {
		var dx, dy uint32
		if r, _, _ := pGetDpiForMonitor.Call(mon, 0, uintptr(unsafe.Pointer(&dx)), uintptr(unsafe.Pointer(&dy))); r == 0 && dx > 0 {
			dpi = int(dx)
		}
	}
	return
}

func forceForeground(h windows.Handle) {
	fg, _, _ := pGetForegroundWindow.Call()
	if fg != 0 && fg != uintptr(h) {
		fgThread, _, _ := pGetWindowThreadProcID.Call(fg, 0)
		me := windows.GetCurrentThreadId()
		if uint32(fgThread) != me {
			pAttachThreadInput.Call(uintptr(me), fgThread, 1)
			pSetForegroundWindow.Call(uintptr(h))
			pAttachThreadInput.Call(uintptr(me), fgThread, 0)
			pSetFocus.Call(uintptr(h))
			return
		}
	}
	pSetForegroundWindow.Call(uintptr(h))
	pSetFocus.Call(uintptr(h))
}
