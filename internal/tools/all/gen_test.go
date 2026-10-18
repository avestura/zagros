package all

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/gen"
)

func read(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("%s is missing: run `go run ./cmd/toolsgen` and commit the result (%v)", rel, err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// TestGeneratedFilesAreCurrent fails when the committed stubs and docs differ
// from what the registry generates. Fix: go run ./cmd/toolsgen
func TestGeneratedFilesAreCurrent(t *testing.T) {
	for _, g := range gen.Groups() {
		if got, want := read(t, "stubs/zagros/"+g+".pyi"), gen.Stub(g); got != want {
			t.Errorf("stubs/zagros/%s.pyi is stale: run `go run ./cmd/toolsgen`", g)
		}
		if got, want := read(t, "docs/extensions/tools-"+g+".md"), gen.DocsGroup(g); got != want {
			t.Errorf("docs/extensions/tools-%s.md is stale: run `go run ./cmd/toolsgen`", g)
		}
	}
	if got, want := read(t, "docs/extensions/tools.md"), gen.DocsIndex(); got != want {
		t.Error("docs/extensions/tools.md is stale: run `go run ./cmd/toolsgen`")
	}
}
