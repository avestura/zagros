//go:build windows

package main

import (
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestWaitForPIDBlocksUntilExit(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "ping -n 3 127.0.0.1 >nul")
	if err := cmd.Start(); err != nil {
		t.Skip(err)
	}
	start := time.Now()
	waitForPID(uint32(cmd.Process.Pid))
	if d := time.Since(start); d < 500*time.Millisecond || d > 15*time.Second {
		t.Fatalf("waited %v for a process that lives about 2s", d)
	}
	_ = cmd.Wait()
	start = time.Now()
	waitForPID(uint32(cmd.Process.Pid)) // already gone
	if time.Since(start) > time.Second {
		t.Fatal("waiting for a finished process must return at once")
	}
}

func TestHandleRestartStripsTheFlag(t *testing.T) {
	if got := handleRestart([]string{"--wait-pid", "99999999", "--trace"}); !reflect.DeepEqual(got, []string{"--trace"}) {
		t.Fatalf("got %v", got)
	}
	if got := handleRestart([]string{"ext", "list"}); !reflect.DeepEqual(got, []string{"ext", "list"}) {
		t.Fatalf("got %v", got)
	}
	if got := handleRestart(nil); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}
