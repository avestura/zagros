//go:build windows

package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/ext"
)

func TestLogServerSendsBacklogThenLive(t *testing.T) {
	buf := ext.NewLogBuffer()
	buf.Add("a", "info", "old")
	path := filepath.Join(os.TempDir(), "zg-test-logs.sock")
	stop := serveLogsAt(path, buf)
	defer stop()
	c, err := net.DialTimeout("unix", path, 2*time.Second)
	if err != nil {
		t.Skip("no AF_UNIX here:", err)
	}
	defer c.Close()
	dec := json.NewDecoder(c)
	read := func() wireLine {
		_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
		var l wireLine
		if err := dec.Decode(&l); err != nil {
			t.Fatal(err)
		}
		return l
	}
	if l := read(); l.Msg != "old" || l.ID != "a" {
		t.Fatalf("backlog: %+v", l)
	}
	if l := read(); !l.Live {
		t.Fatalf("want the live marker, got %+v", l)
	}
	buf.Add("b", "error", "boom\nsecond line")
	if l := read(); l.ID != "b" || l.Level != "error" || l.Msg != "boom\nsecond line" {
		t.Fatalf("live: %+v", l)
	}
}

func TestLogWriterFeedsBuffer(t *testing.T) {
	buf := ext.NewLogBuffer()
	_, _ = logWriter{buf}.Write([]byte("hello\r\n"))
	if es := buf.Entries("zagros"); len(es) != 1 || es[0].Msg != "hello" {
		t.Fatalf("%+v", es)
	}
}
