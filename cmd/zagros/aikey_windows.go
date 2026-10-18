//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// readSecretLine reads a line from stdin, without echo when stdin is a console.
func readSecretLine() (string, error) {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if windows.GetConsoleMode(h, &mode) == nil {
		if windows.SetConsoleMode(h, mode&^windows.ENABLE_ECHO_INPUT) == nil {
			defer windows.SetConsoleMode(h, mode)
		}
	}
	return readLine()
}
