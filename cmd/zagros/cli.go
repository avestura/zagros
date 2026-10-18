//go:build windows

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/ui"
)

// cliPrompter asks for consent in the terminal.
type cliPrompter struct{ in *bufio.Reader }

func (c cliPrompter) Request(m *ext.Manifest, requested []ext.Permission) ([]ext.Permission, error) {
	fmt.Printf("\n%s %s by %s requests permissions:\n", m.Name, m.Version, orUnknown(m.Author))
	var granted []ext.Permission
	for _, p := range requested {
		fmt.Printf("  %s", p)
		if r := m.Reason(p); r != "" {
			fmt.Printf("  (%s)", r)
		}
		fmt.Printf("\n    %s\n", ui.DefaultExplanation(p.Kind()))
	}
	fmt.Print("Grant all of them? [y/N/each] ")
	ans, _ := c.in.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "y", "yes":
		return append(granted, requested...), nil
	case "each", "e":
		for _, p := range requested {
			fmt.Printf("  allow %s? [y/N] ", p)
			a, _ := c.in.ReadString('\n')
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(a)), "y") {
				granted = append(granted, p)
			}
		}
	}
	return granted, nil
}

func orUnknown(s string) string {
	if s == "" {
		return "an unknown author"
	}
	return s
}

func newCLIManager(interactive bool) (*ext.Manager, error) {
	o := ext.Options{ExtDir: core.ExtDir(), DataDir: core.DataDir(), Log: func(string, string, string) {}}
	if interactive {
		o.Prompter = cliPrompter{bufio.NewReader(os.Stdin)}
	}
	return ext.NewManager(o)
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}

func cmdExt(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	switch args[0] {
	case "install":
		if len(args) < 2 {
			return fail(fmt.Errorf("usage: zagros ext install <path|github:owner/repo[@tag]|url|id>"))
		}
		return extInstall(ctx, args[1])
	case "list", "ls":
		m, err := newCLIManager(false)
		if err != nil {
			return fail(err)
		}
		exts, err := m.Discover(ctx)
		if err != nil {
			return fail(err)
		}
		if len(exts) == 0 {
			fmt.Println("no extensions installed")
			return 0
		}
		recs := m.Records()
		for _, e := range exts {
			state := "enabled"
			if !e.Enabled() {
				state = "disabled"
			}
			var perms []string
			for _, p := range e.Granted() {
				perms = append(perms, string(p))
			}
			src := recs[e.ID()].Source
			fmt.Printf("%-32s %-10s %-9s %s\n", e.ID(), e.Manifest.Version, state, src)
			if len(perms) > 0 {
				fmt.Printf("    granted: %s\n", strings.Join(perms, ", "))
			}
			if err := e.Err(); err != nil {
				fmt.Printf("    error: %v\n", err)
			}
		}
	case "remove", "rm", "uninstall":
		fs := flag.NewFlagSet("remove", flag.ContinueOnError)
		keep := fs.Bool("keep-data", false, "keep the extension's storage, settings and grants")
		if err := fs.Parse(reorder(args[1:])); err != nil || fs.NArg() != 1 {
			return fail(fmt.Errorf("usage: zagros ext remove <id> [--keep-data]"))
		}
		m, err := newCLIManager(false)
		if err != nil {
			return fail(err)
		}
		if _, err := m.Discover(ctx); err != nil {
			return fail(err)
		}
		if err := m.Uninstall(fs.Arg(0), !*keep); err != nil {
			return fail(err)
		}
		fmt.Println("removed", fs.Arg(0))
		ui.SignalReload()
	case "pack":
		fs := flag.NewFlagSet("pack", flag.ContinueOnError)
		out := fs.String("o", "", "output file")
		if err := fs.Parse(reorder(args[1:])); err != nil || fs.NArg() != 1 {
			return fail(fmt.Errorf("usage: zagros ext pack <dir> [-o file]"))
		}
		p, err := ext.PackFile(fs.Arg(0), *out)
		if err != nil {
			return fail(err)
		}
		fmt.Println("wrote", p)
	case "init":
		fs := flag.NewFlagSet("init", flag.ContinueOnError)
		id := fs.String("id", "", "extension id (default local.<folder>)")
		name := fs.String("name", "", "display name (default from the folder name)")
		if err := fs.Parse(reorder(args[1:])); err != nil || fs.NArg() != 1 {
			return fail(fmt.Errorf("usage: zagros ext init <dir> [--id id] [--name Name]"))
		}
		if err := ext.Scaffold(fs.Arg(0), *id, *name); err != nil {
			return fail(err)
		}
		fmt.Println("created", fs.Arg(0), "(manifest.toml, main.star, AGENTS.md, API_INDEX.md, zagros-stubs/)")
		fmt.Println("next: cd", fs.Arg(0), "&& zagros ext validate . && zagros --dev")
	case "stubs":
		if len(args) != 2 {
			return fail(fmt.Errorf("usage: zagros ext stubs <dir>   (refresh zagros-stubs/, AGENTS.md and API_INDEX.md)"))
		}
		if err := ext.WriteStubs(args[1]); err != nil {
			return fail(err)
		}
		fmt.Println("updated", filepath.Join(args[1], ext.StubsDir))
	case "validate":
		if len(args) != 2 {
			return fail(fmt.Errorf("usage: zagros ext validate <dir>"))
		}
		issues := ext.Validate(args[1])
		for _, i := range issues {
			fmt.Println(i)
		}
		if ext.HasErrors(issues) {
			return 1
		}
		fmt.Println("ok")
	default:
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	return 0
}

// reorder moves flags before positional args so "remove id --keep-data" works.
func reorder(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if args[i] == "-o" && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			pos = append(pos, args[i])
		}
	}
	return append(flags, pos...)
}

func extInstall(ctx context.Context, ref string) int {
	m, err := newCLIManager(true)
	if err != nil {
		return fail(err)
	}
	if _, err := m.Discover(ctx); err != nil {
		return fail(err)
	}
	var e *ext.Extension
	if _, serr := os.Stat(ref); serr == nil {
		e, err = m.InstallPath(ctx, ref)
	} else {
		var pkg *market.Package
		c := market.New(core.DataDir(), mustConfig().Marketplace.RegistryURL)
		pkg, err = c.Resolve(ctx, ref)
		if err != nil {
			return fail(err)
		}
		if pkg.Kind == market.KindTheme {
			return fail(fmt.Errorf("%s is a theme, not an extension; run: zagros theme install %s", ref, ref))
		}
		e, err = m.InstallBytes(ctx, pkg.Zip, pkg.Source)
	}
	if err != nil {
		return fail(err)
	}
	if e == nil {
		return fail(fmt.Errorf("installation failed"))
	}
	fmt.Printf("installed %s %s\n", e.ID(), e.Manifest.Version)
	if ui.SignalReload() {
		fmt.Println("asked the running Zagros to reload")
	}
	return 0
}

func mustConfig() core.Config {
	c, _ := core.LoadConfig(core.ConfigPath())
	return c
}
