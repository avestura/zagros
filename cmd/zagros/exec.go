//go:build windows

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ui"
	"golang.org/x/sys/windows"
)

// executor implements core.Executor for the built-in action kinds. The
// extension manager wraps it to handle "ext-call".
type executor struct {
	// setQuery is used when an action of kind "setquery" reaches the executor
	// (the UI normally handles it before).
	setQuery func(string)
}

func (e executor) Execute(a core.Action) error {
	switch a.Kind {
	case "open":
		return ui.ShellOpen(a.Extra["verb"], a.Target, joinArgs(a.Args), a.Extra["dir"])
	case "run":
		return runAction(a)
	case "copy":
		return ui.SetClipboard(a.Target)
	case "func":
		if a.Func == nil {
			return errors.New("action has no function")
		}
		return a.Func()
	case "setquery":
		if e.setQuery != nil {
			e.setQuery(a.Target)
		}
		return nil
	case "none", "":
		return nil
	}
	return fmt.Errorf("unknown action kind %q", a.Kind)
}

func joinArgs(args []string) string {
	q := make([]string, len(args))
	for i, a := range args {
		q[i] = windows.EscapeArg(a)
	}
	return strings.Join(q, " ")
}

// runAction starts a program. Target may be a bare command line when Args is
// empty ("notepad C:\x.txt"); a ShellExecute verb such as "runas" can be given
// in Extra["verb"].
func runAction(a core.Action) error {
	file, params := a.Target, joinArgs(a.Args)
	if len(a.Args) == 0 && strings.ContainsAny(file, " \t") {
		if parts, err := windows.DecomposeCommandLine(file); err == nil && len(parts) > 1 {
			if !fileExists(file) {
				file, params = parts[0], joinArgs(parts[1:])
			}
		}
	}
	return ui.ShellOpen(a.Extra["verb"], file, params, a.Extra["dir"])
}
