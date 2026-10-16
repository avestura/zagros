//go:build windows

package builtin

import (
	"context"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/core"
)

func TestSmokeWindows(t *testing.T) {
	t.Setenv("ZAGROS_HOME", t.TempDir())
	cfg := &core.Config{Everything: core.EverythingConfig{Enabled: true}}

	a := newApps(cfg)
	a.once.Do(func() {})
	entries := a.scanShortcuts()
	t.Logf("start menu entries: %d", len(entries))
	withTarget := 0
	for _, e := range entries {
		if e.T != "" {
			withTarget++
		}
	}
	t.Logf("resolved targets: %d", withTarget)
	if len(entries) > 0 && withTarget == 0 {
		t.Error("COM shortcut resolution resolved nothing")
	}

	if n := len(listProcs()); n == 0 {
		t.Error("no processes listed")
	}
	t.Logf("windows: %d", len(listWindows()))

	// Everything must degrade quietly whether or not it is installed.
	ev := newEverything(cfg)
	_ = ev.Query(context.Background(), "f hello")
	t.Logf("everything status: %s", ev.Status())

	// clipboard read must not crash
	_, _ = clipText()
	_ = clipSeq()
	_ = clipExcluded()

	// All wires up every provider and Query is safe for arbitrary input.
	cfg.Providers = map[string]bool{"clipboard": true}
	ps := All(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, p := range ps {
		for _, q := range []string{"", "x", "pw", "?", "1+1", "w ", "kill zzzz", ":smile", "10 km to mi"} {
			_ = p.Query(ctx, q)
		}
	}
	for _, p := range ps {
		if c, ok := p.(*clipboard); ok {
			c.Close()
		}
	}
	t.Logf("providers: %d", len(ps))
}
