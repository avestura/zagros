package ext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLogBufferKeepsNewestPerExtension(t *testing.T) {
	b := NewLogBuffer()
	for i := 0; i < logKeep+5; i++ {
		b.Add("a", "info", fmt.Sprint("line ", i))
	}
	b.Add("b", "error", "boom")
	es := b.Entries("a")
	if len(es) != logKeep || es[0].Msg != "line 5" || es[len(es)-1].Msg != fmt.Sprint("line ", logKeep+4) {
		t.Fatalf("kept %d, first %q last %q", len(es), es[0].Msg, es[len(es)-1].Msg)
	}
	if got := b.Entries("b"); len(got) != 1 || got[0].Level != "error" {
		t.Fatalf("b: %+v", got)
	}
	b.Clear("a")
	if len(b.Entries("a")) != 0 || len(b.Entries("b")) != 1 {
		t.Fatal("Clear must only drop one extension")
	}
}

func TestOptionsLogReachesBufferAndOutput(t *testing.T) {
	var out []string
	o := Options{Log: func(id, level, msg string) { out = append(out, id+level+msg) }}
	o.defaults()
	o.defaults() // idempotent: no double entries
	o.Log("x", "warn", "careful")
	if got := o.Logs.Entries("x"); len(got) != 1 || got[0].Msg != "careful" || len(out) != 1 {
		t.Fatalf("buffer %+v output %v", got, out)
	}
}

func TestDevBuildIsNeverTooOld(t *testing.T) {
	for _, c := range []struct {
		app, min string
		want     bool
	}{{"dev", "0.1.0", false}, {"", "0.1.0", false}, {"0.0.9", "0.1.0", true}, {"0.1.0", "0.1.0", false}, {"1.0.0", "0.1.0", false}} {
		if got := tooOld(c.app, c.min); got != c.want {
			t.Errorf("tooOld(%q, %q) = %v", c.app, c.min, got)
		}
	}
}

func TestBrokenFolderIsListedAndRemovable(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "oops"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "oops", "manifest.toml"), []byte("not = [valid"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(Options{ExtDir: dir, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Discover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.BrokenDir(BrokenPrefix + "oops"); !ok {
		t.Fatalf("broken folder not found: %v", m.BrokenDirs())
	}
	if es := m.Logs("oops"); len(es) == 0 || es[0].Level != "error" {
		t.Fatalf("load failure must be logged: %+v", es)
	}
	if _, ok := m.BrokenDir(BrokenPrefix + "../x"); ok {
		t.Fatal("path traversal accepted")
	}
	if err := m.Uninstall(BrokenPrefix+"oops", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "oops")); err == nil {
		t.Fatal("folder still there")
	}
	if _, ok := m.BrokenDir(BrokenPrefix + "oops"); ok {
		t.Fatal("still listed as broken")
	}
}

func TestFromFolderMarksLocalInstalls(t *testing.T) {
	m, err := NewManager(Options{ExtDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	m.records["mine"] = InstalledRecord{ID: "mine", Source: "local"}
	m.records["theirs"] = InstalledRecord{ID: "theirs", Source: "github:a/b"}
	if !m.FromFolder("mine") || m.FromFolder("theirs") || m.FromFolder("unknown") {
		t.Fatal("only folder installs are the author's own")
	}
}
