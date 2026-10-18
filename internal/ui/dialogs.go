//go:build windows

package ui

import (
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	comdlg32 = windows.NewLazySystemDLL("comdlg32.dll")
	ole32    = windows.NewLazySystemDLL("ole32.dll")

	pGetWindowLongW       = user32.NewProc("GetWindowLongW")
	pGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	pSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	pSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	pDragQueryFileW       = shell32.NewProc("DragQueryFileW")
	pDragFinish           = shell32.NewProc("DragFinish")
	pCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
)

const (
	wmDropFiles = 0x0233

	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800
	ofnNoChangeDir   = 0x00000008
	bifReturnDirs    = 0x0001
)

type openFileName struct {
	StructSize      uint32
	Owner           windows.Handle
	Instance        windows.Handle
	Filter          *uint16
	CustomFilter    *uint16
	MaxCustomFilter uint32
	FilterIndex     uint32
	File            *uint16
	MaxFile         uint32
	FileTitle       *uint16
	MaxFileTitle    uint32
	InitialDir      *uint16
	Title           *uint16
	Flags           uint32
	FileOffset      uint16
	FileExtension   uint16
	DefExt          *uint16
	CustData        uintptr
	Hook            uintptr
	Template        *uint16
	Reserved        unsafe.Pointer
	Reserved2       uint32
	FlagsEx         uint32
}

type browseInfo struct {
	Owner       windows.Handle
	Root        uintptr
	DisplayName *uint16
	Title       *uint16
	Flags       uint32
	Fn          uintptr
	Param       uintptr
	Image       int32
}

// pickExtensionFile shows the Open dialog for a .zagros archive. It returns
// "" when the user cancels.
func pickExtensionFile(owner windows.Handle) string {
	var filter []uint16
	for _, part := range []string{"Zagros extension (*.zagros)", "*.zagros", "All files", "*.*"} {
		filter = append(filter, utf16.Encode([]rune(part))...)
		filter = append(filter, 0)
	}
	filter = append(filter, 0)
	buf := make([]uint16, 1024)
	ofn := openFileName{
		Owner:   owner,
		Filter:  &filter[0],
		File:    &buf[0],
		MaxFile: uint32(len(buf)),
		Title:   u16("Install extension from file"),
		Flags:   ofnFileMustExist | ofnPathMustExist | ofnNoChangeDir,
		DefExt:  u16("zagros"),
	}
	ofn.StructSize = uint32(unsafe.Sizeof(ofn))
	if r, _, _ := pGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn))); r == 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
}

// pickExtensionFolder shows a folder browser. It returns "" on cancel.
func pickExtensionFolder(owner windows.Handle) string {
	name := make([]uint16, 260)
	bi := browseInfo{Owner: owner, DisplayName: &name[0], Title: u16("Choose the extension folder (the one with manifest.toml)"), Flags: bifReturnDirs}
	pidl, _, _ := pSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return ""
	}
	defer pCoTaskMemFree.Call(pidl)
	path := make([]uint16, 520)
	if r, _, _ := pSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&path[0]))); r == 0 {
		return ""
	}
	return windows.UTF16ToString(path)
}

// droppedFiles returns the paths carried by a WM_DROPFILES handle and
// releases it.
func droppedFiles(h uintptr) []string {
	defer pDragFinish.Call(h)
	n, _, _ := pDragQueryFileW.Call(h, ^uintptr(0), 0, 0)
	var out []string
	for i := uintptr(0); i < n; i++ {
		buf := make([]uint16, 1024)
		l, _, _ := pDragQueryFileW.Call(h, i, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if l > 0 {
			out = append(out, windows.UTF16ToString(buf[:l]))
		}
	}
	return out
}
