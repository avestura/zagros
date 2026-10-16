//go:build !windows

package builtin

import (
	"errors"
	"os/exec"
)

var errUnsupported = errors.New("not supported on this platform")

func clipSeq() uint32          { return 0 }
func clipExcluded() bool       { return false }
func clipText() (string, bool) { return "", false }

type winInfo struct {
	HWND  uintptr
	Title string
	PID   uint32
}

func listWindows() []winInfo { return nil }
func focusWindow(uintptr)    {}

type procInfo struct {
	PID  uint32
	Name string
}

func listProcs() []procInfo { return nil }
func killPID(uint32) error  { return errUnsupported }
func hideCmd(*exec.Cmd)     {}
func toggleDarkMode() error { return errUnsupported }

// lnkResolver is a no-op off Windows.
type lnkResolver struct{}

func newLnkResolver() (*lnkResolver, error) { return nil, errUnsupported }
func (r *lnkResolver) Close()               {}
func (r *lnkResolver) Resolve(string) (string, string, error) {
	return "", "", errUnsupported
}

// everything stubs.
type evAPI struct{}

func loadEverything(string) (*evAPI, string, error) { return nil, "", errUnsupported }
func (a *evAPI) search(string, int) ([]string, uint32, error) {
	return nil, 0, errUnsupported
}
