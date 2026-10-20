package builtin

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/update"
)

// ExtUpdate is a newer version of an installed extension.
type ExtUpdate struct {
	ID, Name, Installed, Latest string
}

// UpdateHooks connects the "update" keyword to what the application owns: the
// downloaded Zagros update, the restart, and the extensions' registry. Unset
// hooks hide the matching rows.
var UpdateHooks struct {
	Staged   func() (version string, ok bool) // a verified newer Zagros is ready to install
	Download func(ctx context.Context) error  // download and verify the latest release now
	Restart  func() error                     // swap the executable and start the new one
	// ExtUpdates returns the newer extension versions known from the last
	// check and when it ran; CheckExts asks the registry again.
	ExtUpdates func() ([]ExtUpdate, time.Time)
	CheckExts  func(ctx context.Context) error
	UpdateExt  func(ctx context.Context, id string) error
}

// updateProvider answers the "update" keyword: the state of Zagros itself
// (checked, downloaded, ready to restart) and of the installed extensions.
// Checks and downloads run in the background, never inside Query, and the
// list refreshes itself when they finish.
type updateProvider struct {
	base
	mu     sync.Mutex
	res    *update.Result
	err    error
	busy   bool
	dl     bool   // a download is running
	dlErr  error  // the last download failed
	exBusy string // what the extension update is doing, "" when idle
	exErr  error
	cbs    []func()
	check  func(ctx context.Context, force bool) (update.Result, error)
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

func (p *updateProvider) notify() {
	p.mu.Lock()
	cbs := append([]func(){}, p.cbs...)
	p.mu.Unlock()
	for _, cb := range cbs {
		cb()
	}
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
		p.mu.Unlock()
		if force && UpdateHooks.CheckExts != nil {
			_ = UpdateHooks.CheckExts(ctx)
		}
		p.notify()
	}()
}

// download fetches and verifies the latest release in the background.
func (p *updateProvider) download() {
	p.mu.Lock()
	if p.dl || UpdateHooks.Download == nil {
		p.mu.Unlock()
		return
	}
	p.dl, p.dlErr = true, nil
	p.mu.Unlock()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := UpdateHooks.Download(ctx)
		p.mu.Lock()
		p.dl, p.dlErr = false, err
		p.mu.Unlock()
		p.notify()
	}()
}

// updateExts installs the newer versions of the given extensions in turn.
func (p *updateProvider) updateExts(ups []ExtUpdate) {
	p.mu.Lock()
	if p.exBusy != "" || UpdateHooks.UpdateExt == nil {
		p.mu.Unlock()
		return
	}
	p.exBusy, p.exErr = "starting", nil
	p.mu.Unlock()
	p.notify()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		var firstErr error
		for _, u := range ups {
			p.mu.Lock()
			p.exBusy = u.Name
			p.mu.Unlock()
			p.notify()
			if err := UpdateHooks.UpdateExt(ctx, u.ID); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", u.Name, err)
			}
		}
		if UpdateHooks.CheckExts != nil {
			_ = UpdateHooks.CheckExts(ctx)
		}
		p.mu.Lock()
		p.exBusy, p.exErr = "", firstErr
		p.mu.Unlock()
		p.notify()
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
	items := p.zagrosItems(res, err, busy, fresh)
	return append(items, p.extensionItems()...)
}

func updItem(id, title, sub string, act core.Action, alt ...core.NamedAction) core.Item {
	return core.Item{ID: "update:" + id, Title: title, Subtitle: sub, Category: "Update", Action: noHistoryAction(act), Alt: alt}
}

func (p *updateProvider) zagrosItems(res *update.Result, err error, busy, fresh bool) []core.Item {
	recheck := core.NamedAction{Name: "Check again", Action: core.Action{Kind: "func", Func: func() error { p.start(true); return nil }}}
	releases := "https://github.com/" + update.Repo + "/releases"
	item := updItem

	p.mu.Lock()
	dl, dlErr := p.dl, p.dlErr
	p.mu.Unlock()
	if UpdateHooks.Staged != nil {
		if v, ok := UpdateHooks.Staged(); ok {
			return []core.Item{item("restart", "Restart to update to Zagros "+v, "✔ Update installed - Enter restarts Zagros now",
				core.Action{Kind: "func", Func: func() error {
					if UpdateHooks.Restart == nil {
						return fmt.Errorf("restart is not available")
					}
					return UpdateHooks.Restart()
				}}, recheck)}
		}
	}
	if dl {
		v := ""
		if res != nil {
			v = " " + res.Latest.Version
		}
		return []core.Item{item("downloading", "Downloading Zagros"+v+"...", "Checking its checksum, then it is ready to restart into", core.Action{Kind: "none"})}
	}
	switch {
	case dlErr != nil:
		return []core.Item{item("dlerror", "Could not install the update", dlErr.Error()+" - Enter tries again",
			core.Action{Kind: "func", Func: func() error { p.download(); return nil }}, recheck)}
	case err != nil:
		return []core.Item{item("error", "Could not check for updates", err.Error(), core.Action{Kind: "open", Target: releases}, recheck)}
	case res == nil || busy && !fresh:
		return []core.Item{item("checking", "Checking for updates...", "asking github.com/"+update.Repo, core.Action{Kind: "none"})}
	case res.Available:
		alts := []core.NamedAction{{Name: "Open the release page", Action: core.Action{Kind: "open", Target: res.Latest.URL}}}
		if a, ok := update.AssetFor(res.Latest, runtime.GOARCH); ok {
			alts = append(alts,
				core.NamedAction{Name: fmt.Sprintf("Download the Windows %s zip yourself (%d MB)", runtime.GOARCH, a.Size>>20), Action: core.Action{Kind: "open", Target: a.URL}},
				core.NamedAction{Name: "Copy download link", Action: core.Action{Kind: "copy", Target: a.URL}})
		}
		alts = append(alts, core.NamedAction{Name: "Copy release link", Action: core.Action{Kind: "copy", Target: res.Latest.URL}}, recheck)
		act := core.Action{Kind: "open", Target: res.Latest.URL}
		sub := "You have " + orDev(res.Current) + " - Enter opens the release page, Tab for downloads"
		if UpdateHooks.Download != nil {
			act = core.Action{Kind: "func", Func: func() error { p.download(); return nil }}
			sub = "You have " + orDev(res.Current) + " - Enter downloads and installs it (then restarts), Tab for more"
		}
		return []core.Item{item("available", "Zagros "+res.Latest.Version+" is available", sub, act, alts...)}
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

func (p *updateProvider) extensionItems() []core.Item {
	if UpdateHooks.ExtUpdates == nil {
		return nil
	}
	p.mu.Lock()
	busy, exErr := p.exBusy, p.exErr
	p.mu.Unlock()
	if busy != "" {
		return []core.Item{updItem("extbusy", "Updating extensions...", busy, core.Action{Kind: "none"})}
	}
	if exErr != nil {
		return []core.Item{updItem("exterror", "An extension could not be updated", exErr.Error(), core.Action{Kind: "none"})}
	}
	ups, checked := UpdateHooks.ExtUpdates()
	if checked.IsZero() {
		return nil
	}
	if len(ups) == 0 {
		return []core.Item{updItem("exts-ok", "Extensions are up to date", "checked "+ago(checked), core.Action{Kind: "func", Func: func() error {
			if UpdateHooks.CheckExts != nil {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					_ = UpdateHooks.CheckExts(ctx)
					p.notify()
				}()
			}
			return nil
		}})}
	}
	var names []string
	var alts []core.NamedAction
	for _, u := range ups {
		names = append(names, fmt.Sprintf("%s %s to %s", u.Name, u.Installed, u.Latest))
		u := u
		alts = append(alts, core.NamedAction{Name: fmt.Sprintf("Update only %s (%s to %s)", u.Name, u.Installed, u.Latest),
			Action: core.Action{Kind: "func", Func: func() error { p.updateExts([]ExtUpdate{u}); return nil }}})
	}
	title := fmt.Sprintf("%d extension updates available", len(ups))
	if len(ups) == 1 {
		title = "1 extension update available"
	}
	return []core.Item{updItem("exts", title, strings.Join(names, ", ")+" - Enter updates all, Tab for one",
		core.Action{Kind: "func", Func: func() error { p.updateExts(ups); return nil }}, alts...)}
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
