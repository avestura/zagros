package apidoc

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/mcp"
	"github.com/avestura/zagros/stubs"
)

func index(t *testing.T) *mcp.Index {
	t.Helper()
	idx, err := Build(stubs.FS())
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

// TestAgentGuideSectionsAreCurrent fails when stubs/AGENTS.md has a stale
// module table or cheat sheet. Fix: go run ./cmd/toolsgen
func TestAgentGuideSectionsAreCurrent(t *testing.T) {
	guide := strings.ReplaceAll(string(stubs.AgentGuide()), "\r\n", "\n")
	want, err := Render(guide, index(t))
	if err != nil {
		t.Fatal(err)
	}
	if guide != want {
		t.Fatal("stubs/AGENTS.md generated sections are stale: run `go run ./cmd/toolsgen` and commit the result")
	}
	for _, m := range index(t).ModuleNames() {
		if !strings.Contains(guide, "`"+m+".pyi`") {
			t.Errorf("module %s is missing from the AGENTS.md module table", m)
		}
	}
}

func TestRenderIsDeterministicAndKeepsProse(t *testing.T) {
	idx := index(t)
	guide := "intro\n\n" + marker("BEGIN", SectionModules) + "\nold\n" + marker("END", SectionModules) + "\n\nmid\n\n" + marker("BEGIN", SectionCheatSheet) + "\n" + marker("END", SectionCheatSheet) + "\n\noutro\n"
	a, err := Render(guide, idx)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Render(a, idx)
	if a != b {
		t.Fatal("rendering twice changed the text")
	}
	if !strings.HasPrefix(a, "intro\n") || !strings.Contains(a, "\n\nmid\n\n") || !strings.HasSuffix(a, "\noutro\n") || strings.Contains(a, "\nold\n") {
		t.Fatalf("prose or markers damaged:\n%.300s", a)
	}
	if _, err := Render("no markers", idx); err == nil {
		t.Fatal("a guide without markers must be an error")
	}
	crlf, err := Render(strings.ReplaceAll(guide, "\n", "\r\n"), idx)
	if err != nil || strings.Contains(strings.ReplaceAll(crlf, "\r\n", ""), "\n") {
		t.Fatalf("CRLF guide must stay CRLF: %v", err)
	}
}

func TestModuleTableAndCheatSheet(t *testing.T) {
	idx := index(t)
	table := ModuleTable(idx)
	for _, want := range []string{"`__init__.pyi`", "`hooks.pyi`", "`permissions.pyi`", "`http.pyi`", "`dt.pyi`", "`extract.pyi`", "`encode.pyi`"} {
		if !strings.Contains(table, want) {
			t.Errorf("module table lacks %s", want)
		}
	}
	for _, l := range strings.Split(strings.TrimSpace(table), "\n") {
		if strings.Count(l, "|") != 5 {
			t.Errorf("table row has the wrong number of cells: %q", l)
		}
	}
	sheet := CheatSheet(idx)
	for _, want := range []string{"`item`", "`scope`", "`clipboard.get`", "`http.get` (needs network)", "`encode.base64_encode`"} {
		if !strings.Contains(sheet, want) {
			t.Errorf("cheat sheet lacks %s", want)
		}
	}
}

var indexLineRe = regexp.MustCompile(`^[\w.]+\(.*\) -> [^|]+ \| Needs: [^|]+ \| [^|]+( \| Keywords: [^|]+)?$`)

func TestAPIIndexHasOneLinePerFunction(t *testing.T) {
	idx := index(t)
	out := APIIndex(idx)
	lines := map[string]bool{}
	for _, l := range strings.Split(out, "\n") {
		lines[l] = true
	}
	n := 0
	for _, e := range idx.Entries {
		if e.Kind != "function" {
			continue
		}
		n++
		l := IndexLine(e)
		if !lines[l] {
			t.Errorf("%s: no line in API_INDEX.md", e.Qualified)
		}
		if !indexLineRe.MatchString(l) {
			t.Errorf("%s: line does not follow `module.name(args) -> ret | Needs: perm | summary`: %q", e.Qualified, l)
		}
		if !strings.HasPrefix(l, e.Qualified+"(") {
			t.Errorf("%s: line must start with the qualified name: %q", e.Qualified, l)
		}
	}
	if n < 100 {
		t.Fatalf("only %d functions indexed", n)
	}
	if !strings.Contains(out, "item(id: str, title: str") || strings.Contains(out, "(\n") {
		t.Error("multi-line signatures must be joined onto one line")
	}
}

// Every function carries what agents search by: Keywords, an Example and the
// permission it needs; every stub says it is not runnable Starlark.
func TestStubConventions(t *testing.T) {
	idx := index(t)
	for _, e := range idx.Entries {
		switch e.Kind {
		case "function":
			if len(e.Keywords) == 0 {
				t.Errorf("%s: add a `Keywords:` line (synonyms and task phrases) to its docstring", e.Qualified)
			}
			if e.Example == "" {
				t.Errorf("%s: add an `Example:` to its docstring", e.Qualified)
			}
			if e.Needs == "" {
				t.Errorf("%s: add a `Needs:` line (permission, or none)", e.Qualified)
			}
		case "module":
			if len(e.Keywords) == 0 {
				t.Errorf("stub %s: add a `Keywords:` line to the module docstring", e.File)
			}
			if !strings.Contains(e.Doc, "not runnable Starlark") {
				t.Errorf("stub %s: the module docstring must say it is a type description, not runnable Starlark (no import)", e.File)
			}
		}
	}
	ents, _ := fs.ReadDir(stubs.FS(), "zagros")
	if len(ents) < 20 {
		t.Fatalf("expected the stubs, found %d files", len(ents))
	}
}
