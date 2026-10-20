package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackgroundUpdatesOnByDefaultAndSwitchable(t *testing.T) {
	if !DefaultConfig().Updates.Background {
		t.Fatal("background update checks must default to on")
	}
	p := filepath.Join(t.TempDir(), "zagros.toml")
	if err := os.WriteFile(p, []byte("[updates]\nbackground = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(p)
	if err != nil || c.Updates.Background {
		t.Fatalf("background = false must turn it off: %v %+v", err, c.Updates)
	}
	// the written default file keeps it on
	fresh := filepath.Join(t.TempDir(), "new.toml")
	if c, err := LoadConfig(fresh); err != nil || !c.Updates.Background {
		t.Fatalf("fresh config: %v %+v", err, c.Updates)
	}
	if c, err := LoadConfig(fresh); err != nil || !c.Updates.Background {
		t.Fatalf("reading the default file back: %v %+v", err, c.Updates)
	}
}
