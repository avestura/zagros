//go:build windows

package builtin

import (
	"errors"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	ole32           = syscall.NewLazyDLL("ole32.dll")
	pCoInitializeEx = ole32.NewProc("CoInitializeEx")
	pCoUninitialize = ole32.NewProc("CoUninitialize")
	pCoCreateInst   = ole32.NewProc("CoCreateInstance")
)

var (
	clsidShellLink = syscall.GUID{Data1: 0x00021401, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidShellLinkW  = syscall.GUID{Data1: 0x000214F9, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
	iidPersistFile = syscall.GUID{Data1: 0x0000010B, Data4: [8]byte{0xC0, 0, 0, 0, 0, 0, 0, 0x46}}
)

// vcall invokes method idx of the COM object obj.
func vcall(obj uintptr, idx int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(ptr(obj))
	fn := *(*uintptr)(ptr(vtbl + uintptr(idx)*unsafe.Sizeof(uintptr(0))))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{obj}, args...)...)
	return r
}

// lnkResolver resolves .lnk files through IShellLinkW. It locks the calling
// goroutine to its OS thread; use and Close it from that goroutine only.
type lnkResolver struct {
	link, file uintptr
}

func newLnkResolver() (*lnkResolver, error) {
	runtime.LockOSThread()
	pCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED; S_FALSE is fine
	r := &lnkResolver{}
	hr, _, _ := pCoCreateInst.Call(uintptr(unsafe.Pointer(&clsidShellLink)), 0, 1,
		uintptr(unsafe.Pointer(&iidShellLinkW)), uintptr(unsafe.Pointer(&r.link)))
	if hr != 0 || r.link == 0 {
		pCoUninitialize.Call()
		runtime.UnlockOSThread()
		return nil, errors.New("CoCreateInstance(ShellLink) failed")
	}
	if hr := vcall(r.link, 0, uintptr(unsafe.Pointer(&iidPersistFile)), uintptr(unsafe.Pointer(&r.file))); hr != 0 {
		vcall(r.link, 2)
		pCoUninitialize.Call()
		runtime.UnlockOSThread()
		return nil, errors.New("QueryInterface(IPersistFile) failed")
	}
	return r, nil
}

func (r *lnkResolver) Close() {
	vcall(r.file, 2)
	vcall(r.link, 2)
	pCoUninitialize.Call()
	runtime.UnlockOSThread()
}

// Resolve returns the target path and arguments of a shortcut (target may be
// empty for MSI-advertised shortcuts).
func (r *lnkResolver) Resolve(path string) (target, args string, err error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", "", err
	}
	if hr := vcall(r.file, 5, uintptr(unsafe.Pointer(p)), 0); hr != 0 { // IPersistFile::Load
		return "", "", errors.New("load failed")
	}
	buf := make([]uint16, 1024)
	if hr := vcall(r.link, 3, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0, 0); hr == 0 { // GetPath
		target = syscall.UTF16ToString(buf)
	}
	abuf := make([]uint16, 1024)
	if hr := vcall(r.link, 10, uintptr(unsafe.Pointer(&abuf[0])), uintptr(len(abuf))); hr == 0 { // GetArguments
		args = syscall.UTF16ToString(abuf)
	}
	return target, args, nil
}
