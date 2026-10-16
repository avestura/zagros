// Command registrygen builds site/registry.json from registry/extensions/*.toml.
//
//	go run ./cmd/registrygen                       # query GitHub (honours GITHUB_TOKEN)
//	go run ./cmd/registrygen -offline              # no network; read specs and local marketplace/ sources
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/avestura/zagros/internal/market"
)

func main() {
	specsDir := flag.String("specs", "registry/extensions", "directory with <id>.toml files")
	out := flag.String("out", "site/registry.json", "output file")
	offline := flag.Bool("offline", false, "do not query GitHub; build entries from the spec files and the local marketplace sources")
	root := flag.String("root", ".", "repository root; with -offline, subdir paths are read from here")
	api := flag.String("api", "", "GitHub API base URL (default https://api.github.com)")
	flag.Parse()

	specs, err := market.LoadSpecs(*specsDir)
	if err != nil {
		fatal(err)
	}
	c := market.New("", "")
	c.APIBase = *api
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var entries []market.Entry
	failed := 0
	for _, s := range specs {
		var e market.Entry
		var err error
		if *offline {
			e, err = market.EntryFromLocal(s, *root)
		} else {
			e, err = c.BuildEntry(ctx, s)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", s.ID, err)
			failed++
			continue
		}
		fmt.Printf("ok   %s %s\n", e.ID, e.Latest)
		entries = append(entries, e)
	}
	if failed > 0 {
		fatal(fmt.Errorf("%d extension(s) failed validation; registry not written", failed))
	}
	reg := market.NewRegistry(entries, time.Now())
	if entries == nil {
		reg.Extensions = []market.Entry{}
	}
	b, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%d extensions)\n", *out, len(reg.Extensions))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "registrygen:", err)
	os.Exit(1)
}
