//go:build windows

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/theme"
	"github.com/avestura/zagros/internal/ui"
)

// themeRow describes one available theme.
type themeRow struct {
	Name   string
	Author string
	Source string // "built-in", "user" or "installed"
	File   string // path of the .toml for user and installed themes
}

func themeRecordPath() string { return filepath.Join(core.DataDir(), "installed-themes.json") }

// readThemeRecords maps theme file stem to the source it was installed from.
func readThemeRecords() map[string]string {
	rec := map[string]string{}
	if b, err := os.ReadFile(themeRecordPath()); err == nil {
		_ = json.Unmarshal(b, &rec)
	}
	return rec
}

func writeThemeRecords(rec map[string]string) {
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(themeRecordPath()), 0o755)
	_ = os.WriteFile(themeRecordPath(), b, 0o644)
}

func stem(p string) string { return strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)) }

// listThemes returns built-in themes, then user and installed ones, by name.
func listThemes() []themeRow {
	rows := map[string]themeRow{}
	for _, n := range theme.List() {
		t, _ := theme.Builtin(n)
		rows[strings.ToLower(t.Name)] = themeRow{Name: t.Name, Author: t.Author, Source: "built-in"}
	}
	rec := readThemeRecords()
	dir := core.ThemeDir()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".toml") {
			continue
		}
		file := filepath.Join(dir, e.Name())
		t, err := theme.Load(file)
		if err != nil {
			continue
		}
		src := "user"
		if _, ok := rec[stem(file)]; ok {
			src = "installed"
		}
		rows[strings.ToLower(t.Name)] = themeRow{Name: t.Name, Author: t.Author, Source: src, File: file}
	}
	out := make([]themeRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].Source == "built-in") != (out[j].Source == "built-in") {
			return out[i].Source == "built-in"
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

// installTheme installs a theme and records where it came from.
func installTheme(ctx context.Context, c *market.Client, ref string) (*market.Package, error) {
	pkg, err := c.InstallTheme(ctx, core.ThemeDir(), ref)
	if err != nil {
		return nil, err
	}
	rec := readThemeRecords()
	rec[pkg.ID] = pkg.Source
	writeThemeRecords(rec)
	return pkg, nil
}

// removeTheme deletes a user or installed theme by name or file name.
func removeTheme(name string) error {
	for _, r := range listThemes() {
		if r.File == "" {
			continue
		}
		if strings.EqualFold(r.Name, name) || strings.EqualFold(stem(r.File), name) {
			if err := os.Remove(r.File); err != nil {
				return err
			}
			rec := readThemeRecords()
			delete(rec, stem(r.File))
			writeThemeRecords(rec)
			if strings.EqualFold(mustConfig().Theme, r.Name) {
				fmt.Fprintf(os.Stderr, "note: config still selects %q; it now falls back to the default\n", r.Name)
			}
			return nil
		}
	}
	for _, n := range theme.List() {
		if strings.EqualFold(n, name) {
			return fmt.Errorf("%s is a built-in theme and cannot be removed", name)
		}
	}
	return fmt.Errorf("no user or installed theme named %q", name)
}

func cmdTheme(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	switch args[0] {
	case "list", "ls":
		cur := mustConfig().Theme
		for _, r := range listThemes() {
			mark := ""
			if cur != "" && strings.EqualFold(cur, r.Name) {
				mark = "  (in use)"
			}
			fmt.Printf("%-24s %-10s%s\n", r.Name, r.Source, mark)
		}
	case "install":
		if len(args) != 2 {
			return fail(errors.New("usage: zagros theme install <registry-id|https url to .toml>"))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		c := market.New(core.DataDir(), mustConfig().Marketplace.RegistryURL)
		pkg, err := installTheme(ctx, c, args[1])
		if err != nil {
			return fail(err)
		}
		fmt.Printf("installed theme %s\n", pkg.Manifest.Name)
		fmt.Printf("use it by setting theme = %q in %s\n", pkg.Manifest.Name, core.ConfigPath())
		if ui.SignalReload() {
			fmt.Println("asked the running Zagros to reload")
		}
	case "remove", "rm", "uninstall":
		if len(args) != 2 {
			return fail(errors.New("usage: zagros theme remove <name>"))
		}
		if err := removeTheme(args[1]); err != nil {
			return fail(err)
		}
		fmt.Println("removed", args[1])
		ui.SignalReload()
	default:
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	return 0
}

func kindLabel(e market.Entry) string {
	if e.Kind == market.KindTheme {
		return "theme"
	}
	return "ext"
}

func cmdMarket(args []string) int {
	if len(args) == 0 || (args[0] != "list" && args[0] != "ls" && args[0] != "search") {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	query := ""
	if args[0] == "search" {
		if len(args) < 2 {
			return fail(errors.New("usage: zagros market search <query>"))
		}
		query = strings.Join(args[1:], " ")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	c := market.New(core.DataDir(), mustConfig().Marketplace.RegistryURL)
	reg, err := c.List(ctx)
	if err != nil {
		return fail(err)
	}
	if reg.Stale {
		fmt.Fprintln(os.Stderr, "warning: offline, showing cached registry")
	}
	entries := market.SearchEntries(reg.Extensions, query)
	if len(entries) == 0 {
		fmt.Println("no matches")
		return 0
	}
	for _, e := range entries {
		fmt.Printf("%-6s %-28s %s\n", kindLabel(e), e.ID, e.Description)
	}
	return 0
}

// ---- ui.Themes ---------------------------------------------------------------

type themeAdmin struct{}

func (themeAdmin) List() []ui.ThemeInfo {
	cur := mustConfig().Theme
	var out []ui.ThemeInfo
	for _, r := range listThemes() {
		out = append(out, ui.ThemeInfo{Name: r.Name, Author: r.Author, Source: r.Source,
			Current: cur != "" && strings.EqualFold(cur, r.Name)})
	}
	return out
}

func (themeAdmin) Use(name string) error { return core.SetTheme(core.ConfigPath(), name) }
