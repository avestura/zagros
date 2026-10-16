//go:build windows

package builtin

import (
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")

	pOpenClipboard      = user32.NewProc("OpenClipboard")
	pCloseClipboard     = user32.NewProc("CloseClipboard")
	pGetClipboardData   = user32.NewProc("GetClipboardData")
	pClipSeq            = user32.NewProc("GetClipboardSequenceNumber")
	pIsClipFormat       = user32.NewProc("IsClipboardFormatAvailable")
	pRegClipFormat      = user32.NewProc("RegisterClipboardFormatW")
	pGlobalLock         = kernel32.NewProc("GlobalLock")
	pGlobalUnlock       = kernel32.NewProc("GlobalUnlock")
	pGlobalSize         = kernel32.NewProc("GlobalSize")
	pEnumWindows        = user32.NewProc("EnumWindows")
	pIsWindowVisible    = user32.NewProc("IsWindowVisible")
	pGetWindowTextLen   = user32.NewProc("GetWindowTextLengthW")
	pGetWindowText      = user32.NewProc("GetWindowTextW")
	pGetWindow          = user32.NewProc("GetWindow")
	pGetWindowLong      = user32.NewProc("GetWindowLongW")
	pGetWindowThreadPID = user32.NewProc("GetWindowThreadProcessId")
	pIsIconic           = user32.NewProc("IsIconic")
	pShowWindow         = user32.NewProc("ShowWindow")
	pSetForeground      = user32.NewProc("SetForegroundWindow")
	pDwmGetAttr         = dwmapi.NewProc("DwmGetWindowAttribute")
	pSendMsgTimeout     = user32.NewProc("SendMessageTimeoutW")
)

// ptr converts a uintptr returned by the OS into an unsafe.Pointer without
// tripping vet's unsafeptr check.
func ptr(u uintptr) unsafe.Pointer { return *(*unsafe.Pointer)(unsafe.Pointer(&u)) }

func clipSeq() uint32 {
	r, _, _ := pClipSeq.Call()
	return uint32(r)
}

var (
	excludeFmtOnce sync.Once
	excludeFmts    []uintptr
)

// clipExcluded reports whether the current clipboard content asks not to be
// monitored (password managers set these formats).
func clipExcluded() bool {
	excludeFmtOnce.Do(func() {
		for _, n := range []string{"ExcludeClipboardContentFromMonitorProcessing", "Clipboard Viewer Ignore"} {
			p, _ := syscall.UTF16PtrFromString(n)
			if id, _, _ := pRegClipFormat.Call(uintptr(unsafe.Pointer(p))); id != 0 {
				excludeFmts = append(excludeFmts, id)
			}
		}
	})
	for _, id := range excludeFmts {
		if r, _, _ := pIsClipFormat.Call(id); r != 0 {
			return true
		}
	}
	return false
}

// clipText reads the clipboard's unicode text.
func clipText() (string, bool) {
	opened := false
	for i := 0; i < 5; i++ {
		if r, _, _ := pOpenClipboard.Call(0); r != 0 {
			opened = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !opened {
		return "", false
	}
	defer pCloseClipboard.Call()
	h, _, _ := pGetClipboardData.Call(13) // CF_UNICODETEXT
	if h == 0 {
		return "", false
	}
	p, _, _ := pGlobalLock.Call(h)
	if p == 0 {
		return "", false
	}
	defer pGlobalUnlock.Call(h)
	size, _, _ := pGlobalSize.Call(h)
	n := int(size / 2)
	if n == 0 {
		return "", false
	}
	u := unsafe.Slice((*uint16)(ptr(p)), n)
	for i, c := range u {
		if c == 0 {
			u = u[:i]
			break
		}
	}
	return syscall.UTF16ToString(u), true
}

type winInfo struct {
	HWND  uintptr
	Title string
	PID   uint32
}

var (
	enumMu  sync.Mutex
	enumOut []winInfo
	enumCB  uintptr
	enumCBO sync.Once
)

func enumProc(h, _ uintptr) uintptr {
	if r, _, _ := pIsWindowVisible.Call(h); r == 0 {
		return 1
	}
	n, _, _ := pGetWindowTextLen.Call(h)
	if n == 0 {
		return 1
	}
	if o, _, _ := pGetWindow.Call(h, 4); o != 0 { // GW_OWNER
		return 1
	}
	ex, _, _ := pGetWindowLong.Call(h, ^uintptr(19)) // GWL_EXSTYLE
	if ex&0x80 != 0 && ex&0x40000 == 0 {             // TOOLWINDOW without APPWINDOW
		return 1
	}
	var cloaked uint32
	pDwmGetAttr.Call(h, 14, uintptr(unsafe.Pointer(&cloaked)), 4)
	if cloaked != 0 {
		return 1
	}
	buf := make([]uint16, n+1)
	pGetWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
	title := syscall.UTF16ToString(buf)
	if title == "" || title == "Program Manager" {
		return 1
	}
	var pid uint32
	pGetWindowThreadPID.Call(h, uintptr(unsafe.Pointer(&pid)))
	enumOut = append(enumOut, winInfo{h, title, pid})
	return 1
}

func listWindows() []winInfo {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumCBO.Do(func() { enumCB = syscall.NewCallback(enumProc) })
	enumOut = nil
	pEnumWindows.Call(enumCB, 0)
	out := enumOut
	enumOut = nil
	return out
}

func focusWindow(h uintptr) {
	if r, _, _ := pIsIconic.Call(h); r != 0 {
		pShowWindow.Call(h, 9) // SW_RESTORE
	}
	pSetForeground.Call(h)
}

type procInfo struct {
	PID  uint32
	Name string
}

func listProcs() []procInfo {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil
	}
	defer syscall.CloseHandle(snap)
	var pe syscall.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	var out []procInfo
	if err := syscall.Process32First(snap, &pe); err != nil {
		return nil
	}
	for {
		out = append(out, procInfo{pe.ProcessID, syscall.UTF16ToString(pe.ExeFile[:])})
		if err := syscall.Process32Next(snap, &pe); err != nil {
			break
		}
	}
	return out
}

func killPID(pid uint32) error {
	h, err := syscall.OpenProcess(1, false, pid) // PROCESS_TERMINATE
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(h)
	return syscall.TerminateProcess(h, 1)
}

func hideCmd(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

func broadcastThemeChange() {
	p, _ := syscall.UTF16PtrFromString("ImmersiveColorSet")
	var res uintptr
	pSendMsgTimeout.Call(0xFFFF, 0x1A, 0, uintptr(unsafe.Pointer(p)), 2, 1000, uintptr(unsafe.Pointer(&res)))
}

// toggleDarkMode flips the apps/system light-theme registry values.
func toggleDarkMode() error {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	cur := exec.Command("reg", "query", key, "/v", "AppsUseLightTheme")
	hideCmd(cur)
	out, _ := cur.Output()
	light := !strings.Contains(string(out), "0x0")
	v := "0"
	if !light {
		v = "1"
	}
	for _, name := range []string{"AppsUseLightTheme", "SystemUsesLightTheme"} {
		c := exec.Command("reg", "add", key, "/v", name, "/t", "REG_DWORD", "/d", v, "/f")
		hideCmd(c)
		if err := c.Run(); err != nil {
			return err
		}
	}
	broadcastThemeChange()
	return nil
}
