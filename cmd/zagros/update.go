package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/update"
)

// cmdUpdate implements `zagros update [check] [--json] [--prerelease]`: it
// reports whether a newer release exists on GitHub. It never installs.
func cmdUpdate(args []string) int {
	if len(args) > 0 && args[0] == "check" {
		args = args[1:]
	}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "print the result as JSON")
	pre := fs.Bool("prerelease", false, "include pre-releases")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: zagros update [check] [--json] [--prerelease]")
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := update.Check(ctx, version, update.Options{IncludePrerelease: *pre})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	update.SaveCache(core.DataDir(), res)
	if *asJSON {
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
		return 0
	}
	fmt.Println(res.Summary())
	fmt.Println("release page:", res.Latest.URL)
	if a, ok := update.AssetFor(res.Latest, runtime.GOARCH); ok && (res.Available || res.Dev) {
		fmt.Printf("download:     %s (%d bytes)\n", a.URL, a.Size)
	}
	return 0
}

// updateMessage runs a check for the tray item and returns the balloon text.
func updateMessage() string {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := update.CheckCached(ctx, core.DataDir(), version, 0, update.Options{})
	if err != nil {
		return err.Error()
	}
	return res.Summary()
}
