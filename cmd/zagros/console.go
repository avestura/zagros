//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// attachConsole connects a GUI-subsystem process (-H=windowsgui) to the
// console it was started from so CLI output and prompts are visible. Redirected
// handles are left alone.
func attachConsole() {
	attach := windows.NewLazySystemDLL("kernel32.dll").NewProc("AttachConsole")
	attach.Call(^uintptr(0)) // ATTACH_PARENT_PROCESS
	if h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE); err != nil || h == 0 || h == windows.InvalidHandle {
		if f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
			os.Stdout, os.Stderr = f, f
		}
	} else if h, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE); err != nil || h == 0 || h == windows.InvalidHandle {
		os.Stderr = os.Stdout
	}
	if h, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE); err != nil || h == 0 || h == windows.InvalidHandle {
		if f, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
			os.Stdin = f
		}
	}
}
