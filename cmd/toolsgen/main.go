// Command toolsgen writes the .pyi stubs and docs generated from the tools
// registry (stubs/zagros/<group>.pyi, docs/extensions/tools*.md) and then
// refreshes the generated module table and task cheat sheet inside
// stubs/AGENTS.md from all the stubs, hand-written ones included.
//
//	go run ./cmd/toolsgen
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/avestura/zagros/internal/apidoc"
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

	stubs := filepath.Join(root, "stubs")
	idx, err := apidoc.Build(os.DirFS(stubs))
	if err != nil {
		fail(err)
	}
	guidePath := filepath.Join(stubs, "AGENTS.md")
	guide, err := os.ReadFile(guidePath)
	if err != nil {
		fail(err)
	}
	out, err := apidoc.Render(string(guide), idx)
	if err != nil {
		fail(err)
	}
	write(guidePath, out)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "toolsgen:", err)
	os.Exit(1)
}

func write(path, text string) {
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		fail(err)
	}
}
