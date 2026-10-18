package builtin

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/update"
)

// updateProvider answers the "update" keyword with the result of a GitHub
// release check. The check runs in the background (never inside Query), is
// cached for a few hours, and the list refreshes itself when it finishes.
type updateProvider struct {
	base
	mu    sync.Mutex
	res   *update.Result
	err   error
	busy  bool
	cbs   []func()
	check func(ctx context.Context, force bool) (update.Result, error)
}

const updateMaxAge = 6 * time.Hour

func newUpdate() *updateProvider {
	p := &updateProvider{base: base{"update"}}
	p.check = func(ctx context.Context, force bool) (update.Result, error) {
		age := updateMaxAge
		if force {
			age = 0
		}
		return update.CheckCached(ctx, core.DataDir(), core.AppVersion, age, update.Options{})
	}
	return p
}

func (p *updateProvider) Ranked() bool       { return true }
func (p *updateProvider) Keywords() []string { return []string{"update", "updates", "upgrade"} }

// OnReady implements core.ReadyNotifier.
func (p *updateProvider) OnReady(fn func()) {
	p.mu.Lock()
	p.cbs = append(p.cbs, fn)
	p.mu.Unlock()
}

func (p *updateProvider) start(force bool) {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return
	}
	p.busy = true
	if force {
		p.res, p.err = nil, nil
	}
	p.mu.Unlock()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		r, err := p.check(ctx, force)
		p.mu.Lock()
		p.busy = false
		if err != nil {
			p.res, p.err = nil, err
		} else {
			p.res, p.err = &r, nil
		}
		cbs := append([]func(){}, p.cbs...)
		p.mu.Unlock()
		for _, cb := range cbs {
			cb()
		}
	}()
}

func (p *updateProvider) Query(ctx context.Context, q string) []core.Item {
	if _, _, ok := trigger(q, p.Keywords()...); !ok {
		return nil
	}
	p.mu.Lock()
	res, err, busy := p.res, p.err, p.busy
	fresh := res != nil && time.Since(res.Checked) < updateMaxAge
	p.mu.Unlock()
	if !fresh && err == nil && !busy {
		p.start(false)
		busy = true
	}
	recheck := core.NamedAction{Name: "Check again", Action: core.Action{Kind: "func", Func: func() error { p.start(true); return nil }}}
	releases := "https://github.com/" + update.Repo + "/releases"
	item := func(id, title, sub string, act core.Action, alt ...core.NamedAction) core.Item {
		return core.Item{ID: "update:" + id, Title: title, Subtitle: sub, Category: "Update", Action: noHistoryAction(act), Alt: alt}
	}
	switch {
	case err != nil:
		return []core.Item{item("error", "Could not check for updates", err.Error(), core.Action{Kind: "open", Target: releases}, recheck)}
	case res == nil || busy && !fresh:
		return []core.Item{item("checking", "Checking for updates...", "asking github.com/"+update.Repo, core.Action{Kind: "none"})}
	case res.Available:
		alts := []core.NamedAction{}
		if a, ok := update.AssetFor(res.Latest, runtime.GOARCH); ok {
			alts = append(alts,
				core.NamedAction{Name: fmt.Sprintf("Download Windows %s zip (%d MB)", runtime.GOARCH, a.Size>>20), Action: core.Action{Kind: "open", Target: a.URL}},
				core.NamedAction{Name: "Copy download link", Action: core.Action{Kind: "copy", Target: a.URL}})
		}
		alts = append(alts, core.NamedAction{Name: "Copy release link", Action: core.Action{Kind: "copy", Target: res.Latest.URL}}, recheck)
		return []core.Item{item("available", "Zagros "+res.Latest.Version+" is available",
			"You have "+orDev(res.Current)+" - Enter opens the release page, Tab for downloads", core.Action{Kind: "open", Target: res.Latest.URL}, alts...)}
	case res.Dev:
		return []core.Item{item("dev", "Latest release: Zagros "+res.Latest.Version, "This is a development build - Enter opens the release page",
			core.Action{Kind: "open", Target: res.Latest.URL}, recheck)}
	case res.Ahead:
		return []core.Item{item("ahead", "You have "+res.Current+", newer than the latest release "+res.Latest.Version,
			"Enter opens the releases page", core.Action{Kind: "open", Target: releases}, recheck)}
	}
	return []core.Item{item("uptodate", "Zagros is up to date", orDev(res.Current)+" - checked "+ago(res.Checked)+" - Enter opens the releases page",
		core.Action{Kind: "open", Target: releases}, recheck)}
}

func orDev(v string) string {
	if v == "" {
		return "dev"
	}
	return v
}

func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%d h ago", int(d.Hours()))
}
