//go:build windows

package ui

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	dwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	shcore   = windows.NewLazySystemDLL("shcore.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassExW       = user32.NewProc("RegisterClassExW")
	pCreateWindowExW        = user32.NewProc("CreateWindowExW")
	pDestroyWindow          = user32.NewProc("DestroyWindow")
	pDefWindowProcW         = user32.NewProc("DefWindowProcW")
	pShowWindow             = user32.NewProc("ShowWindow")
	pSetWindowPos           = user32.NewProc("SetWindowPos")
	pGetMessageW            = user32.NewProc("GetMessageW")
	pTranslateMessage       = user32.NewProc("TranslateMessage")
	pDispatchMessageW       = user32.NewProc("DispatchMessageW")
	pPostMessageW           = user32.NewProc("PostMessageW")
	pPostQuitMessage        = user32.NewProc("PostQuitMessage")
	pBeginPaint             = user32.NewProc("BeginPaint")
	pEndPaint               = user32.NewProc("EndPaint")
	pInvalidateRect         = user32.NewProc("InvalidateRect")
	pGetClientRect          = user32.NewProc("GetClientRect")
	pGetWindowRect          = user32.NewProc("GetWindowRect")
	pSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	pSetFocus               = user32.NewProc("SetFocus")
	pGetCursorPos           = user32.NewProc("GetCursorPos")
	pMonitorFromPoint       = user32.NewProc("MonitorFromPoint")
	pGetMonitorInfoW        = user32.NewProc("GetMonitorInfoW")
	pRegisterHotKey         = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey       = user32.NewProc("UnregisterHotKey")
	pGetKeyState            = user32.NewProc("GetKeyState")
	pSetTimer               = user32.NewProc("SetTimer")
	pKillTimer              = user32.NewProc("KillTimer")
	pLoadCursorW            = user32.NewProc("LoadCursorW")
	pSetCursor              = user32.NewProc("SetCursor")
	pSetCapture             = user32.NewProc("SetCapture")
	pReleaseCapture         = user32.NewProc("ReleaseCapture")
	pSetLayeredWindowAttrs  = user32.NewProc("SetLayeredWindowAttributes")
	pCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	pAppendMenuW            = user32.NewProc("AppendMenuW")
	pTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	pDestroyMenu            = user32.NewProc("DestroyMenu")
	pRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	pFindWindowW            = user32.NewProc("FindWindowW")
	pSetProcessDpiAwareCtx  = user32.NewProc("SetProcessDpiAwarenessContext")
	pDrawIconEx             = user32.NewProc("DrawIconEx")
	pDestroyIcon            = user32.NewProc("DestroyIcon")
	pCreateIconFromResEx    = user32.NewProc("CreateIconFromResourceEx")
	pOpenClipboard          = user32.NewProc("OpenClipboard")
	pCloseClipboard         = user32.NewProc("CloseClipboard")
	pEmptyClipboard         = user32.NewProc("EmptyClipboard")
	pGetClipboardData       = user32.NewProc("GetClipboardData")
	pSetClipboardData       = user32.NewProc("SetClipboardData")
	pSetWindowRgn           = user32.NewProc("SetWindowRgn")
	pGetForegroundWindow    = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadProcID  = user32.NewProc("GetWindowThreadProcessId")
	pAttachThreadInput      = user32.NewProc("AttachThreadInput")
	pIsWindowVisible        = user32.NewProc("IsWindowVisible")
	pPeekMessageW           = user32.NewProc("PeekMessageW")
	pMsgWait                = user32.NewProc("MsgWaitForMultipleObjects")

	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pSetBkColor             = gdi32.NewProc("SetBkColor")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pSetTextColor           = gdi32.NewProc("SetTextColor")
	pExtTextOutW            = gdi32.NewProc("ExtTextOutW")
	pTextOutW               = gdi32.NewProc("TextOutW")
	pGetTextExtentPoint32W  = gdi32.NewProc("GetTextExtentPoint32W")
	pCreateFontIndirectW    = gdi32.NewProc("CreateFontIndirectW")
	pGetTextMetricsW        = gdi32.NewProc("GetTextMetricsW")
	pIntersectClipRect      = gdi32.NewProc("IntersectClipRect")
	pSaveDC                 = gdi32.NewProc("SaveDC")
	pRestoreDC              = gdi32.NewProc("RestoreDC")
	pCreateRoundRectRgn     = gdi32.NewProc("CreateRoundRectRgn")
	pGetPixel               = gdi32.NewProc("GetPixel")

	pDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")

	pShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
	pSHGetFileInfoW   = shell32.NewProc("SHGetFileInfoW")
	pShellExecuteW    = shell32.NewProc("ShellExecuteW")

	pGetDpiForMonitor = shcore.NewProc("GetDpiForMonitor")

	pGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	pGlobalLock       = kernel32.NewProc("GlobalLock")
	pGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	pGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	wsPopup = 0x80000000

	wsExToolWindow = 0x00000080
	wsExTopmost    = 0x00000008
	wsExLayered    = 0x00080000

	wmCreate      = 0x0001
	wmDestroy     = 0x0002
	wmActivate    = 0x0006
	wmPaint       = 0x000F
	wmClose       = 0x0010
	wmEraseBkgnd  = 0x0014
	wmSetCursor   = 0x0020
	wmKeyDown     = 0x0100
	wmChar        = 0x0102
	wmSysKeyDown  = 0x0104
	wmSysCommand  = 0x0112
	wmTimer       = 0x0113
	wmMouseMove   = 0x0200
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmLButtonDbl  = 0x0203
	wmRButtonUp   = 0x0205
	wmMouseWheel  = 0x020A
	wmHotkey      = 0x0312
	wmApp         = 0x8000

	swHide = 0
	swShow = 5

	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpNoActivate = 0x0010

	hwndTopmost = ^uintptr(0)

	vkBack    = 0x08
	vkTab     = 0x09
	vkReturn  = 0x0D
	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12
	vkEscape  = 0x1B
	vkPrior   = 0x21
	vkNext    = 0x22
	vkEnd     = 0x23
	vkHome    = 0x24
	vkLeft    = 0x25
	vkUp      = 0x26
	vkRight   = 0x27
	vkDown    = 0x28
	vkDelete  = 0x2E

	srcCopy     = 0x00CC0020
	transparent = 1
	etoOpaque   = 0x2
	lwaAlpha    = 0x2

	tpmReturnCmd   = 0x0100
	tpmRightButton = 0x0002
	tpmBottomAlign = 0x0020
	mfString       = 0x0
	mfSeparator    = 0x800
	mfChecked      = 0x8
	mfGrayed       = 0x1

	nimAdd     = 0
	nimModify  = 1
	nimDelete  = 2
	nifMessage = 0x1
	nifIcon    = 0x2
	nifTip     = 0x4
	nifInfo    = 0x10
	niifInfo   = 0x1

	idcArrow = 32512
	idcIBeam = 32513
	idcHand  = 32649

	cfUnicodeText = 13
	gmemMoveable  = 0x2

	shgfiIcon      = 0x100
	shgfiLargeIcon = 0x0

	monitorDefaultToNearest = 2

	dwmwaWindowCornerPreference = 33
	dwmwcpDoNotRound            = 1
	dwmwcpRound                 = 2

	diNormal = 3
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type msg struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}
type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}
type paintStruct struct {
	Hdc         windows.Handle
	Erase       int32
	Paint       rect
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}
type monitorInfo struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
}
type sizeT struct{ CX, CY int32 }
type logFont struct {
	Height, Width, Escapement, Orientation, Weight int32
	Italic, Underline, StrikeOut, CharSet          byte
	OutPrecision, ClipPrecision, Quality, PitchFam byte
	FaceName                                       [32]uint16
}
type textMetric struct {
	Height, Ascent, Descent, InternalLeading, ExternalLeading int32
	AveCharWidth, MaxCharWidth, Weight, Overhang              int32
	DigitizedAspectX, DigitizedAspectY                        int32
	FirstChar, LastChar, DefaultChar, BreakChar               uint16
	Italic, Underlined, StruckOut, PitchAndFamily, CharSet    byte
}
type notifyIconData struct {
	Size            uint32
	Hwnd            windows.Handle
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	TimeoutOrVer    uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     windows.Handle
}
type shFileInfo struct {
	Icon        windows.Handle
	IconIndex   int32
	Attributes  uint32
	DisplayName [260]uint16
	TypeName    [80]uint16
}

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func hinstance() windows.Handle {
	h, _, _ := pGetModuleHandleW.Call(0)
	return windows.Handle(h)
}

func signedX(lp uintptr) int { return int(int16(lp & 0xFFFF)) }
func signedY(lp uintptr) int { return int(int16((lp >> 16) & 0xFFFF)) }

// colorref converts 0xRRGGBB to a GDI COLORREF (0x00BBGGRR).
func colorref(c uint32) uintptr {
	return uintptr((c&0xFF)<<16 | (c & 0xFF00) | (c>>16)&0xFF)
}

func keyDown(vk int) bool {
	r, _, _ := pGetKeyState.Call(uintptr(vk))
	return int16(r) < 0
}

func postMessage(h windows.Handle, m uint32, wp, lp uintptr) {
	pPostMessageW.Call(uintptr(h), uintptr(m), wp, lp)
}

func fillRect(dc uintptr, r rect, color uint32) {
	pSetBkColor.Call(dc, colorref(color))
	pExtTextOutW.Call(dc, 0, 0, etoOpaque, uintptr(unsafe.Pointer(&r)), 0, 0, 0)
}

func utf16NoNul(s string) []uint16 {
	w, _ := syscall.UTF16FromString(s)
	return w[:len(w)-1]
}

func textWidth(dc uintptr, s string) int {
	if s == "" {
		return 0
	}
	w := utf16NoNul(s)
	var sz sizeT
	pGetTextExtentPoint32W.Call(dc, uintptr(unsafe.Pointer(&w[0])), uintptr(len(w)), uintptr(unsafe.Pointer(&sz)))
	return int(sz.CX)
}

func drawText(dc uintptr, x, y int, s string) {
	if s == "" {
		return
	}
	w := utf16NoNul(s)
	pTextOutW.Call(dc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&w[0])), uintptr(len(w)))
}

// drawTextClipped draws s at (x,y) limited to maxW pixels, adding an
// ellipsis, and returns the drawn width.
func drawTextClipped(dc uintptr, x, y, maxW int, s string) int {
	if maxW <= 0 || s == "" {
		return 0
	}
	if textWidth(dc, s) > maxW {
		rs := []rune(s)
		lo, hi := 0, len(rs)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if textWidth(dc, string(rs[:mid])+"…") <= maxW {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		s = string(rs[:lo]) + "…"
	}
	drawText(dc, x, y, s)
	return textWidth(dc, s)
}

func makeFont(face string, px int, bold bool) uintptr {
	var lf logFont
	lf.Height = int32(-px)
	lf.Weight = 400
	if bold {
		lf.Weight = 700
	}
	lf.CharSet = 1
	lf.Quality = 5 // CLEARTYPE_QUALITY
	n, _ := syscall.UTF16FromString(face)
	copy(lf.FaceName[:31], n)
	h, _, _ := pCreateFontIndirectW.Call(uintptr(unsafe.Pointer(&lf)))
	return h
}

func fontHeight(dc, font uintptr) int {
	old, _, _ := pSelectObject.Call(dc, font)
	var tm textMetric
	pGetTextMetricsW.Call(dc, uintptr(unsafe.Pointer(&tm)))
	pSelectObject.Call(dc, old)
	return int(tm.Height)
}
