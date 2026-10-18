//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/ext"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/ui"
)

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// ---- ext.Platform -----------------------------------------------------------

// platform gives extensions clipboard and tray notifications.
type platform struct{ app func() *ui.App }

func (platform) ClipboardGet() (string, error) { return ui.GetClipboard(), nil }
func (platform) ClipboardSet(s string) error   { return ui.SetClipboard(s) }
func (p platform) Notify(title, body string) error {
	if a := p.app(); a != nil {
		a.Notify(title, body)
	}
	return nil
}

// ---- ext.PermissionPrompter ---------------------------------------------------

// guiPrompter shows the native consent dialog.
type guiPrompter struct {
	app func() *ui.App
	mgr func() *ext.Manager
}

func (g guiPrompter) Request(m *ext.Manifest, requested []ext.Permission) ([]ext.Permission, error) {
	a := g.app()
	if a == nil {
		return nil, nil
	}
	source := m.Homepage
	if mgr := g.mgr(); mgr != nil {
		if rec, ok := mgr.Records()[m.ID]; ok && rec.Source != "" {
			source = rec.Source
		}
	}
	if source == "" {
		source = "local folder"
	}
	req := ui.PermissionRequest{Extension: m.Name, Publisher: m.Author, Source: source}
	byName := map[string]ext.Permission{}
	for _, p := range requested {
		expl := ui.DefaultExplanation(p.Kind())
		if r := m.Reason(p); r != "" {
			expl += " The extension says: " + r
		}
		if p.Kind() == ext.PermAI {
			expl += fmt.Sprintf(" Daily limit: %d tokens.", m.AIBudget())
		}
		req.Asks = append(req.Asks, ui.PermissionAsk{Name: string(p), Label: p.Kind(), Scope: p.Scope(), Explanation: expl})
		byName[string(p)] = p
	}
	res := a.PromptPermissions(req)
	if res.Cancelled {
		return nil, nil
	}
	var out []ext.Permission
	for _, n := range res.Granted {
		if p, ok := byName[n]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

// ---- ui.ExtAdmin -----------------------------------------------------------

type extAdmin struct{ m *ext.Manager }

func (x extAdmin) List() []ui.ExtInfo {
	var out []ui.ExtInfo
	for _, e := range x.m.List() {
		info := ui.ExtInfo{ID: e.ID(), Name: e.Manifest.Name, Version: e.Manifest.Version,
			Publisher: e.Manifest.Author, Enabled: e.Enabled()}
		for _, p := range e.Granted() {
			info.Permissions = append(info.Permissions, string(p))
		}
		out = append(out, info)
	}
	return out
}

func (x extAdmin) SetEnabled(id string, on bool) error {
	return x.m.SetEnabled(context.Background(), id, on)
}

func (x extAdmin) Revoke(id, perm string) error {
	e, ok := x.m.Get(id)
	if !ok {
		return os.ErrNotExist
	}
	g, _ := x.m.Permissions().Get(id)
	var keep []ext.Permission
	for _, p := range e.Granted() {
		if string(p) != perm {
			keep = append(keep, p)
		}
	}
	if g.Declared == nil {
		g.Declared = e.Manifest.Requestable()
	}
	g.Granted, g.Decided = keep, time.Now().UTC()
	if err := x.m.Permissions().Save(id, g); err != nil {
		return err
	}
	e.SetGranted(keep)
	return nil
}

func (x extAdmin) Uninstall(id string) error { return x.m.Uninstall(id, true) }

func (x extAdmin) InstallLocal(ctx context.Context, path string) (string, error) {
	e, err := x.m.InstallPath(ctx, path)
	if err != nil {
		return "", err
	}
	return e.Manifest.Name, nil
}

// ---- ui.Market and market.Installer ------------------------------------------

type marketAdapter struct {
	c   *market.Client
	mgr *ext.Manager
}

func (m marketAdapter) List(ctx context.Context) ([]ui.MarketEntry, error) {
	reg, err := m.c.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []ui.MarketEntry
	for _, e := range reg.Extensions {
		if e.Kind == market.KindTheme {
			e.Permissions = nil
		}
		desc := e.Description
		if len(e.Permissions) > 0 {
			desc += "  [" + strings.Join(e.Permissions, ", ") + "]"
		}
		author := e.Author
		if e.Verified {
			author += " (verified)"
		}
		out = append(out, ui.MarketEntry{Ref: e.ID, Name: e.Name, Description: desc, Author: author, Version: e.Latest, Kind: e.Kind})
	}
	return out, nil
}

func (m marketAdapter) Install(ctx context.Context, e ui.MarketEntry) error {
	if e.Kind == market.KindTheme {
		_, err := installTheme(ctx, m.c, e.Ref)
		return err
	}
	_, err := m.c.Install(ctx, installer{m.mgr}, e.Ref)
	return err
}

type installer struct{ m *ext.Manager }

func (i installer) Install(ctx context.Context, zip []byte, source string) (string, error) {
	e, err := i.m.InstallBytes(ctx, zip, source)
	if e != nil {
		return e.ID(), err
	}
	return "", err
}

func (i installer) Installed() map[string]string {
	out := map[string]string{}
	for _, e := range i.m.List() {
		out[e.ID()] = e.Manifest.Version
	}
	return out
}
