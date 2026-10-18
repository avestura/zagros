// Command toolsgen writes the .pyi stubs and docs generated from the tools
// registry: stubs/zagros/<group>.pyi and docs/extensions/tools*.md.
//
//	go run ./cmd/toolsgen
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/avestura/zagros/internal/tools/all"
	"github.com/avestura/zagros/internal/tools/gen"
)

var _ = all.Doc

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	for _, g := range gen.Groups() {
		write(filepath.Join(root, "stubs", "zagros", g+".pyi"), gen.Stub(g))
		write(filepath.Join(root, "docs", "extensions", "tools-"+g+".md"), gen.DocsGroup(g))
	}
	write(filepath.Join(root, "docs", "extensions", "tools.md"), gen.DocsIndex())
}

func write(path, text string) {
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "toolsgen:", err)
		os.Exit(1)
	}
}
