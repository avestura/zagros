package builtin

import (
	"context"
	"fmt"

	"github.com/avestura/zagros/internal/core"
)

// reindexProvider answers "reindex": one row that rescans the app index now
// instead of waiting for the 12 hour cache to expire.
type reindexProvider struct{ base }

func newReindex() *reindexProvider { return &reindexProvider{base{"reindex"}} }

func (p *reindexProvider) Ranked() bool       { return true }
func (p *reindexProvider) Keywords() []string { return []string{"reindex", "rescan"} }

func (p *reindexProvider) Query(ctx context.Context, q string) []core.Item {
	if _, _, ok := trigger(q, p.Keywords()...); !ok {
		return nil
	}
	a := theApps.Load()
	if a == nil {
		return nil
	}
	n, updated, scanning := a.indexInfo()
	if scanning {
		return []core.Item{{ID: "reindex:apps", Title: "Rebuilding the app index...", Category: "Apps",
			Subtitle: "Scanning the Start Menu and Store apps in the background", Action: core.Action{Kind: "none"}}}
	}
	sub := "Rescan the Start Menu, scan_dirs and Store apps now"
	if !updated.IsZero() {
		sub = fmt.Sprintf("%d apps, scanned %s - Enter rescans now", n, ago(updated))
	}
	return []core.Item{{ID: "reindex:apps", Title: "Rebuild the app index", Category: "Apps", Subtitle: sub,
		Action: noHistoryAction(core.Action{Kind: "func", Func: func() error { a.Reindex(); return nil }})}}
}
