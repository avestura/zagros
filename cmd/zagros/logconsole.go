//go:build windows

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
	"golang.org/x/sys/windows"
)

// Logs are kept in memory only (ext.LogBuffer). The running launcher serves
// them on a local socket in its data directory and `zagros logs` streams them
// into a console. The tray item "Log console" starts that command in its own
// window, so closing the window never touches the launcher.

func logSockPath() string { return filepath.Join(core.DataDir(), "logs.sock") }

// wireLine is one log entry on the socket. A line with Live set marks the end
// of the buffered history.
type wireLine struct {
	Time  string `json:"t,omitempty"`
	ID    string `json:"id,omitempty"`
	Level string `json:"l,omitempty"`
	Msg   string `json:"m,omitempty"`
	Live  bool   `json:"live,omitempty"`
}

func toWire(e ext.LogEntry) wireLine {
	return wireLine{Time: e.Time.Format("15:04:05"), ID: e.ExtID, Level: e.Level, Msg: e.Msg}
}

// logWriter feeds the standard logger into the buffer as source "zagros".
type logWriter struct{ b *ext.LogBuffer }

func (w logWriter) Write(p []byte) (int, error) {
	if msg := strings.TrimRight(string(p), "\r\n"); msg != "" {
		w.b.Add("zagros", "info", msg)
	}
	return len(p), nil
}

// serveLogs listens for log viewers until the returned function is called.
// Without the socket (path too long, no AF_UNIX) the launcher still runs.
func serveLogs(buf *ext.LogBuffer) (stop func()) { return serveLogsAt(logSockPath(), buf) }

func serveLogsAt(path string, buf *ext.LogBuffer) (stop func()) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.Remove(path) // left by a crashed instance
	ln, err := net.Listen("unix", path)
	if err != nil {
		buf.Add("zagros", "warn", "log console unavailable: "+err.Error())
		return func() {}
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go streamLogs(c, buf)
		}
	}()
	return func() { ln.Close(); _ = os.Remove(path) }
}

func streamLogs(c net.Conn, buf *ext.LogBuffer) {
	defer c.Close()
	defer func() { _ = recover() }()
	ch := make(chan wireLine, 512)
	backlog, cancel := buf.Subscribe(func(e ext.LogEntry) {
		select {
		case ch <- toWire(e):
		default: // a viewer that cannot keep up misses lines instead of slowing logging
		}
	})
	defer cancel()
	gone := make(chan struct{})
	go func() { _, _ = io.Copy(io.Discard, c); close(gone) }()
	enc := json.NewEncoder(c)
	for _, e := range backlog {
		if enc.Encode(toWire(e)) != nil {
			return
		}
	}
	if enc.Encode(wireLine{Live: true}) != nil {
		return
	}
	for {
		select {
		case l := <-ch:
			if enc.Encode(l) != nil {
				return
			}
		case <-gone:
			return
		}
	}
}

func (s *runtimeState) logsOn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.logWin != nil
}

// toggleLogs opens the log console window, or closes it when it is open.
func (s *runtimeState) toggleLogs() {
	s.mu.Lock()
	if p := s.logWin; p != nil {
		s.mu.Unlock()
		_ = p.Kill()
		return
	}
	s.mu.Unlock()
	self, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(self, "logs", "--window")
	if err := cmd.Start(); err != nil {
		s.logs.Add("zagros", "error", "cannot open the log console: "+err.Error())
		return
	}
	s.mu.Lock()
	s.logWin = cmd.Process
	s.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		s.logWin = nil
		s.mu.Unlock()
	}()
}

// cmdLogs implements `zagros logs [id] [-n lines] [--window]`.
func cmdLogs(args []string) int {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	n := fs.Int("n", 200, "buffered lines to show before going live")
	window := fs.Bool("window", false, "open in a console window of its own (used by the tray item)")
	var id string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		id, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *window {
		newConsole("Zagros logs")
	}
	color := enableColor()
	pause := func() {
		if *window {
			fmt.Print("\nPress Enter to close...")
			_, _ = fmt.Scanln()
		}
	}
	conn, err := net.DialTimeout("unix", logSockPath(), 2*time.Second)
	if err != nil {
		fmt.Println("Cannot reach a running Zagros with a log console (is Zagros running, in a version that has the log console?):", err)
		pause()
		return 1
	}
	defer conn.Close()
	fmt.Println("Zagros logs. Kept in memory only (the last 2000 lines); nothing is written to disk. Ctrl+C or closing this window stops watching.")
	dec := json.NewDecoder(conn)
	var backlog []wireLine
	live := false
	for {
		var l wireLine
		if dec.Decode(&l) != nil {
			break
		}
		switch {
		case l.Live:
			live = true
			if len(backlog) > *n {
				backlog = backlog[len(backlog)-*n:]
			}
			for _, b := range backlog {
				printLog(b, color)
			}
			backlog = nil
		case !live:
			if id == "" || l.ID == id {
				backlog = append(backlog, l)
			}
		case id == "" || l.ID == id:
			printLog(l, color)
		}
	}
	fmt.Println("\nZagros closed the log stream.")
	pause()
	return 0
}

func printLog(l wireLine, color bool) {
	col, reset := "", ""
	if color {
		reset = "\x1b[0m"
		switch l.Level {
		case "error":
			col = "\x1b[31m"
		case "warn":
			col = "\x1b[33m"
		default:
			col = "\x1b[90m"
		}
	}
	msg := strings.ReplaceAll(strings.TrimRight(l.Msg, "\r\n"), "\n", "\n    ")
	fmt.Printf("%s %s%-5s%s %-24s %s\n", l.Time, col, l.Level, reset, l.ID, msg)
}

// newConsole gives this GUI-subsystem process a console of its own.
func newConsole(title string) {
	k := windows.NewLazySystemDLL("kernel32.dll")
	k.NewProc("AllocConsole").Call()
	k.NewProc("SetConsoleOutputCP").Call(65001)
	if p, err := windows.UTF16PtrFromString(title); err == nil {
		k.NewProc("SetConsoleTitleW").Call(uintptr(unsafe.Pointer(p)))
	}
	if f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		os.Stdout, os.Stderr = f, f
	}
	if f, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
		os.Stdin = f
	}
}

// enableColor turns on ANSI colours for stdout and reports whether it worked.
func enableColor() bool {
	h := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return false
	}
	return windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}
