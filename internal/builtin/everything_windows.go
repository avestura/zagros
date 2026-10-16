//go:build windows

package builtin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// evAPI wraps the Everything SDK DLL. The DLL keeps global query state, so
// searches are serialised.
type evAPI struct {
	mu                                                                sync.Mutex
	reset, setSearch, setMax, setFlags, query, num, fullPath, lastErr *syscall.Proc
}

func everythingDLLName() string {
	switch runtime.GOARCH {
	case "386":
		return "Everything32.dll"
	case "arm64":
		return "EverythingARM64.dll"
	}
	return "Everything64.dll"
}

// loadEverything searches for the SDK DLL: explicit path, next to the exe,
// Program Files install dirs, then PATH. Only absolute paths are loaded.
func loadEverything(explicit string) (*evAPI, string, error) {
	name := everythingDLLName()
	var cands []string
	if explicit != "" {
		cands = append(cands, explicit)
	}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), name))
	}
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
		if pf := os.Getenv(env); pf != "" {
			cands = append(cands, filepath.Join(pf, "Everything", name), filepath.Join(pf, "Everything 1.5a", name))
		}
	}
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if d != "" {
			cands = append(cands, filepath.Join(d, name))
		}
	}
	for _, c := range cands {
		if !filepath.IsAbs(c) {
			continue
		}
		if _, err := os.Stat(c); err != nil {
			continue
		}
		dll, err := syscall.LoadDLL(c)
		if err != nil {
			continue
		}
		a := &evAPI{}
		ok := true
		for dst, fn := range map[**syscall.Proc]string{
			&a.reset: "Everything_Reset", &a.setSearch: "Everything_SetSearchW", &a.setMax: "Everything_SetMax",
			&a.setFlags: "Everything_SetRequestFlags", &a.query: "Everything_QueryW", &a.num: "Everything_GetNumResults",
			&a.fullPath: "Everything_GetResultFullPathNameW", &a.lastErr: "Everything_GetLastError",
		} {
			p, err := dll.FindProc(fn)
			if err != nil {
				ok = false
				break
			}
			*dst = p
		}
		if ok {
			return a, c, nil
		}
	}
	return nil, "", errors.New("Everything SDK DLL not found")
}

// search runs q and returns up to max full paths. The returned code is the
// Everything error (2 = IPC unavailable, i.e. Everything is not running).
func (a *evAPI) search(ctx context.Context, q string, max int) ([]string, uint32, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if ctx.Err() != nil { // superseded while waiting for the DLL
		return nil, 0, ctx.Err()
	}
	p, err := syscall.UTF16PtrFromString(q)
	if err != nil {
		return nil, 0, err
	}
	a.reset.Call()
	a.setSearch.Call(uintptr(unsafe.Pointer(p)))
	a.setMax.Call(uintptr(max))
	a.setFlags.Call(0x4) // EVERYTHING_REQUEST_FULL_PATH_AND_FILE_NAME
	if r, _, _ := a.query.Call(1); r == 0 {
		code, _, _ := a.lastErr.Call()
		return nil, uint32(code), errors.New("Everything query failed")
	}
	n, _, _ := a.num.Call()
	out := make([]string, 0, n)
	buf := make([]uint16, 32768)
	for i := uintptr(0); i < n; i++ {
		l, _, _ := a.fullPath.Call(i, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if l > 0 {
			out = append(out, syscall.UTF16ToString(buf[:l]))
		}
	}
	return out, 0, nil
}
