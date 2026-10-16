//go:build windows

package ui

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// ShellOpen calls ShellExecuteW. verb may be "" for the default verb.
func ShellOpen(verb, file, params, dir string) error {
	var vp, pp, dp *uint16
	if verb != "" {
		vp = u16(verb)
	}
	if params != "" {
		pp = u16(params)
	}
	if dir != "" {
		dp = u16(dir)
	}
	r, _, _ := pShellExecuteW.Call(0, uintptr(unsafe.Pointer(vp)), uintptr(unsafe.Pointer(u16(file))),
		uintptr(unsafe.Pointer(pp)), uintptr(unsafe.Pointer(dp)), 1)
	if r <= 32 {
		return fmt.Errorf("cannot open %q (ShellExecute code %d)", file, r)
	}
	return nil
}

// SetClipboard puts text on the clipboard.
func SetClipboard(s string) error {
	w, err := windows.UTF16FromString(s)
	if err != nil {
		return err
	}
	for i := 0; i < 5; i++ {
		if r, _, _ := pOpenClipboard.Call(0); r != 0 {
			break
		} else if i == 4 {
			return errors.New("clipboard is busy")
		}
		windows.SleepEx(10, false)
	}
	defer pCloseClipboard.Call()
	pEmptyClipboard.Call()
	mem, _, _ := pGlobalAlloc.Call(gmemMoveable, uintptr(len(w)*2))
	if mem == 0 {
		return errors.New("out of memory")
	}
	p, _, _ := pGlobalLock.Call(mem)
	copy(unsafe.Slice((*uint16)(*(*unsafe.Pointer)(unsafe.Pointer(&p))), len(w)), w)
	pGlobalUnlock.Call(mem)
	if r, _, _ := pSetClipboardData.Call(cfUnicodeText, mem); r == 0 {
		return errors.New("SetClipboardData failed")
	}
	return nil
}

// GetClipboard returns the clipboard text, or "".
func GetClipboard() string {
	if r, _, _ := pOpenClipboard.Call(0); r == 0 {
		return ""
	}
	defer pCloseClipboard.Call()
	h, _, _ := pGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return ""
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		return ""
	}
	defer pGlobalUnlock.Call(h)
	return windows.UTF16PtrToString((*uint16)(*(*unsafe.Pointer)(unsafe.Pointer(&p))))
}

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// AutostartEnabled reports whether the HKCU Run entry exists.
func AutostartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	_, _, err = k.GetStringValue("Zagros")
	return err == nil
}

// SetAutostart adds or removes the HKCU Run entry for the current executable.
func SetAutostart(on bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		err := k.DeleteValue("Zagros")
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue("Zagros", `"`+exe+`"`)
}

// AcquireSingleInstance takes the app-wide mutex. When another instance
// already runs it asks that instance to show itself and returns ok=false.
func AcquireSingleInstance() (release func(), ok bool) {
	h, err := windows.CreateMutex(nil, false, u16("Local\\ZagrosLauncherSingleton"))
	if err != nil && h == 0 {
		return func() {}, true // cannot tell; run anyway
	}
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		windows.CloseHandle(h)
		SignalRunningInstance()
		return func() {}, false
	}
	return func() { windows.CloseHandle(h) }, true
}

// SignalReload asks a running launcher to reload extensions and themes.
func SignalReload() bool {
	w, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(u16(classLauncher))), 0)
	if w == 0 {
		return false
	}
	postMessage(windows.Handle(w), msgReload, 0, 0)
	return true
}

// SignalRunningInstance asks a running launcher to show itself.
func SignalRunningInstance() bool {
	w, _, _ := pFindWindowW.Call(uintptr(unsafe.Pointer(u16(classLauncher))), 0)
	if w == 0 {
		return false
	}
	postMessage(windows.Handle(w), msgShow, 0, 0)
	return true
}

var pSetProcessWorkingSetSize = kernel32.NewProc("SetProcessWorkingSetSize")

// trimMemory returns freed heap to the OS and trims the working set. It is
// called after the launcher hides so the idle footprint stays small.
func trimMemory() {
	debug.FreeOSMemory()
	h := windows.CurrentProcess()
	pSetProcessWorkingSetSize.Call(uintptr(h), ^uintptr(0), ^uintptr(0))
}
