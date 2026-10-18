// Package apidoc renders documentation derived from the extension API stubs:
// the module table and task cheat sheet inside stubs/AGENTS.md, and the
// API_INDEX.md that `zagros ext init` writes into new projects. Nothing here
// is written by hand, so a module added to stubs/zagros shows up everywhere
// after `go run ./cmd/toolsgen`, and a test fails until that is done.
package apidoc

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/mcp"
)

// Section names inside stubs/AGENTS.md, written as
// <!-- BEGIN GENERATED: name --> ... <!-- END GENERATED: name -->.
const (
	SectionModules    = "modules"
	SectionCheatSheet = "cheatsheet"
)

func marker(kind, name string) string { return "<!-- " + kind + " GENERATED: " + name + " -->" }

// Build indexes the stubs in fsys (a tree with a zagros/ folder).
func Build(fsys fs.FS) (*mcp.Index, error) { return mcp.BuildIndex(fsys) }

func cell(s string) string {
	return strings.TrimSpace(strings.NewReplacer("|", "/", "\n", " ").Replace(s))
}

func firstLine(doc string) string {
	if i := strings.IndexByte(doc, '\n'); i >= 0 {
		doc = doc[:i]
	}
	return strings.TrimSpace(doc)
}

// shortNeeds cuts a module-level "Needs:" text to the permission names.
func shortNeeds(n string) string {
	n = strings.TrimSpace(n)
	if n == "" {
		return "per function"
	}
	for _, sep := range []string{", scoped", " (", ", "} {
		if i := strings.Index(n, sep); i > 0 {
			n = n[:i]
		}
	}
	return strings.TrimRight(n, ". ")
}

func covers(e mcp.Entry) string {
	d := firstLine(e.Doc)
	d = strings.TrimPrefix(d, e.Name+": ")
	d = strings.TrimPrefix(d, e.Name+" module: ")
	if d != "" {
		d = strings.ToUpper(d[:1]) + d[1:]
	}
	return d
}

// briefNeeds drops a trailing parenthetical from a function's Needs text.
func briefNeeds(n string) string {
	if i := strings.Index(n, " ("); i > 0 {
		n = n[:i]
	}
	return strings.TrimRight(n, ". ")
}

func functions(idx *mcp.Index, module string) int {
	n := 0
	for _, e := range idx.InModule(module) {
		if e.Kind == "function" {
			n++
		}
	}
	return n
}

// ModuleTable is the Markdown table of every stub file.
func ModuleTable(idx *mcp.Index) string {
	var b strings.Builder
	b.WriteString("| Stub file | Permission | What it covers | Functions |\n|---|---|---|---|\n")
	var globals []mcp.Entry
	for _, e := range idx.Entries {
		if e.Kind == "module" && e.Module == "" {
			globals = append(globals, e)
		}
	}
	order := map[string]int{"__init__": 0, "hooks": 1, "permissions": 2}
	sort.SliceStable(globals, func(i, j int) bool {
		a, bb := order[globals[i].Name], order[globals[j].Name]
		if a != bb {
			return a < bb
		}
		return globals[i].Name < globals[j].Name
	})
	for _, e := range globals {
		file := strings.TrimPrefix(e.File, "zagros/")
		count := "-"
		if e.Name == "__init__" {
			count = fmt.Sprint(globalFunctions(idx, "__init__"))
		} else if e.Name == "hooks" {
			count = fmt.Sprint(globalFunctions(idx, "hooks"))
		}
		fmt.Fprintf(&b, "| `%s` (global names) | %s | %s | %s |\n", file, cell(shortNeeds(e.Needs)), cell(covers(e)), count)
	}
	for _, m := range idx.ModuleNames() {
		e := idx.Modules[m]
		fmt.Fprintf(&b, "| `%s.pyi` (module `%s`) | %s | %s | %d |\n", m, m, cell(shortNeeds(e.Needs)), cell(covers(e)), functions(idx, m))
	}
	return b.String()
}

func globalFunctions(idx *mcp.Index, stub string) int {
	n := 0
	for _, e := range idx.Entries {
		if e.Kind == "function" && e.Module == "" && strings.TrimPrefix(e.File, "zagros/") == stub+".pyi" {
			n++
		}
	}
	return n
}

// CheatSheet lists every function with a Task line, sorted by the task.
func CheatSheet(idx *mcp.Index) string {
	type row struct{ task, name, needs string }
	var rows []row
	for _, e := range idx.Entries {
		if e.Task == "" || e.Kind != "function" {
			continue
		}
		rows = append(rows, row{e.Task, e.Qualified, e.Needs})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := strings.ToLower(rows[i].task), strings.ToLower(rows[j].task)
		if a != b {
			return a < b
		}
		return rows[i].name < rows[j].name
	})
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "- %s: `%s`", r.task, r.name)
		if r.needs != "" && r.needs != "none" {
			fmt.Fprintf(&b, " (needs %s)", briefNeeds(r.needs))
		}
		b.WriteString("\n")
	}
	return b.String()
}

var (
	sigOpen  = regexp.MustCompile(`\(\s+`)
	sigClose = regexp.MustCompile(`,?\s+\)`)
	sigSpace = regexp.MustCompile(`\s+`)
)

func cleanSignature(s string) string {
	s = sigSpace.ReplaceAllString(s, " ")
	s = sigOpen.ReplaceAllString(s, "(")
	s = sigClose.ReplaceAllString(s, ")")
	return strings.ReplaceAll(s, ",)", ")")
}

// IndexLine is one API_INDEX.md line:
// module.name(args) -> ret | Needs: perm | summary | Keywords: a, b
func IndexLine(e mcp.Entry) string {
	sig := cleanSignature(e.Signature)
	if e.Kind == "function" {
		if i := strings.Index(sig, "("); i >= 0 {
			sig = e.Qualified + sig[i:]
		}
	} else if e.Module != "" {
		sig = e.Module + "." + sig
	}
	needs := e.Needs
	if needs == "" {
		needs = "none"
	}
	line := fmt.Sprintf("%s | Needs: %s | %s", cell(sig), cell(needs), cell(e.Summary()))
	if len(e.Keywords) > 0 {
		line += " | Keywords: " + cell(strings.Join(e.Keywords, ", "))
	}
	return line
}

// APIIndex renders API_INDEX.md: one line per function (and global variable).
func APIIndex(idx *mcp.Index) string {
	var b strings.Builder
	b.WriteString("# Zagros API index\n\n")
	b.WriteString("One line per function: `module.name(args) -> ret | Needs: permission | summary | Keywords: synonyms and task phrases`.\n")
	b.WriteString("Generated from zagros-stubs/zagros/*.pyi; do not edit. Find a function by task with `grep -i \"unzip\" API_INDEX.md`,\n")
	b.WriteString("then read its stub for parameters and an example. Global names (item, scope, action_*, notify, hooks) have no module prefix.\n\n")
	for _, e := range idx.Entries {
		if e.Kind != "function" && e.Kind != "variable" {
			continue
		}
		b.WriteString(IndexLine(e))
		b.WriteString("\n")
	}
	return b.String()
}

// Render replaces the generated sections of the agent guide.
func Render(guide string, idx *mcp.Index) (string, error) {
	crlf := strings.Contains(guide, "\r\n")
	guide = strings.ReplaceAll(guide, "\r\n", "\n")
	sections := map[string]string{
		SectionModules:    ModuleTable(idx),
		SectionCheatSheet: CheatSheet(idx),
	}
	for _, name := range []string{SectionModules, SectionCheatSheet} {
		begin, end := marker("BEGIN", name), marker("END", name)
		i := strings.Index(guide, begin)
		j := strings.Index(guide, end)
		if i < 0 || j < i {
			return "", fmt.Errorf("AGENTS.md lacks the %s ... %s markers", begin, end)
		}
		guide = guide[:i+len(begin)] + "\n\n" + sections[name] + "\n" + guide[j:]
	}
	if crlf {
		guide = strings.ReplaceAll(guide, "\n", "\r\n")
	}
	return guide, nil
}
