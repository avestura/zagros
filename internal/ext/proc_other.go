//go:build !windows

package ext

import "os/exec"

func hideWindow(*exec.Cmd) {}
