// Package mcp serves the Zagros extension API to coding agents over the Model
// Context Protocol (stdio or HTTP). It indexes the embedded .pyi stubs, so
// the answers are exactly the API the runtime exposes (a test keeps stubs and
// runtime in sync) and it never reads user data.
package mcp

import (
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Entry is one searchable API item: a function, class, variable or module.
type Entry struct {
	Module    string `json:"module"`    // "text", "http", or "" for the global names
	Name      string `json:"name"`      // "lower", "item"
	Qualified string `json:"qualified"` // "text.lower", "item"
	Kind      string `json:"kind"`      // function | class | variable | module
	Signature string `json:"signature"`
	Doc       string `json:"doc"`
	Needs     string `json:"needs"`   // permission text from "Needs: ..." ("none" when pure)
	Example   string `json:"example"` // from "Example: ..." when present
	File      string `json:"file"`    // stub file, e.g. zagros/text.pyi
}

// Summary is the first sentence of the doc.
func (e Entry) Summary() string {
	d := strings.TrimSpace(e.Doc)
	if i := strings.IndexAny(d, ".\n"); i > 0 {
		d = d[:i]
	}
	return d
}

// Index holds every entry of the stubs plus module summaries.
type Index struct {
	Entries []Entry
	Modules map[string]Entry // module-level docs by name
	files   map[string]string
}

var (
	defStartRe = regexp.MustCompile(`^def (\w+)\(`)
	classRe    = regexp.MustCompile(`^class (\w+)`)
	varRe      = regexp.MustCompile(`^(\w+): `)
	needsRe    = regexp.MustCompile(`(?m)Needs: ([^\n]+)`)
	exampleRe  = regexp.MustCompile(`(?m)^\s*Example: (.+)$`)
)

// BuildIndex parses every zagros/*.pyi in fsys.
func BuildIndex(fsys fs.FS) (*Index, error) {
	idx := &Index{Modules: map[string]Entry{}, files: map[string]string{}}
	ents, err := fs.ReadDir(fsys, "zagros")
	if err != nil {
		return nil, err
	}
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pyi") {
			continue
		}
		b, err := fs.ReadFile(fsys, "zagros/"+e.Name())
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(e.Name(), ".pyi")
		idx.files[e.Name()] = string(b)
		module := name
		if name == "__init__" || name == "hooks" || name == "permissions" {
			module = "" // global names
		}
		idx.parse(string(b), module, "zagros/"+e.Name(), name)
	}
	sort.SliceStable(idx.Entries, func(i, j int) bool { return idx.Entries[i].Qualified < idx.Entries[j].Qualified })
	return idx, nil
}

// File returns the stub text by name ("text.pyi").
func (x *Index) File(name string) (string, bool) { s, ok := x.files[name]; return s, ok }

// FileNames lists the stub files.
func (x *Index) FileNames() []string {
	var out []string
	for n := range x.files {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (x *Index) parse(src, module, file, stubName string) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	// module docstring
	if d, _ := docAt(lines, 0); d != "" {
		qual := module
		if qual == "" {
			qual = stubName
		}
		en := Entry{Module: module, Name: stubName, Qualified: qual, Kind: "module", Doc: d, File: file}
		en.Needs, en.Example = extras(d)
		if module != "" {
			x.Modules[module] = en
		}
		x.Entries = append(x.Entries, en)
	}
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		switch {
		case defStartRe.MatchString(l):
			name := defStartRe.FindStringSubmatch(l)[1]
			sig := l
			j := i
			for !signatureDone(sig) && j+1 < len(lines) {
				j++
				sig += " " + strings.TrimSpace(lines[j])
			}
			doc, _ := docAt(lines, j+1)
			en := Entry{Module: module, Name: name, Qualified: qualify(module, name), Kind: "function",
				Signature: strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(sig, "def ")), ":"), Doc: doc, File: file}
			en.Needs, en.Example = extras(doc)
			x.Entries = append(x.Entries, en)
			i = j
		case classRe.MatchString(l):
			name := classRe.FindStringSubmatch(l)[1]
			doc, _ := docAt(lines, i+1)
			x.Entries = append(x.Entries, Entry{Module: module, Name: name, Qualified: qualify(module, name), Kind: "class",
				Signature: strings.TrimSuffix(strings.TrimSpace(l), ":"), Doc: doc, File: file})
		case varRe.MatchString(l) && !strings.HasPrefix(l, "from ") && !strings.HasPrefix(l, "import "):
			name := varRe.FindStringSubmatch(l)[1]
			doc, _ := docAt(lines, i+1)
			en := Entry{Module: module, Name: name, Qualified: qualify(module, name), Kind: "variable", Signature: strings.TrimSpace(l), Doc: doc, File: file}
			en.Needs, en.Example = extras(doc)
			x.Entries = append(x.Entries, en)
		}
	}
}

func qualify(module, name string) string {
	if module == "" {
		return name
	}
	return module + "." + name
}

func signatureDone(sig string) bool {
	if strings.Count(sig, "(") > strings.Count(sig, ")") {
		return false
	}
	return strings.HasSuffix(strings.TrimSpace(sig), ":")
}

// docAt reads a triple-quoted docstring starting at the first non-blank line
// from index i. It returns the dedented text and the index after it.
func docAt(lines []string, i int) (string, int) {
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) {
		return "", i
	}
	t := strings.TrimSpace(lines[i])
	var q string
	switch {
	case strings.HasPrefix(t, `"""`):
		q = `"""`
	case strings.HasPrefix(t, `'''`):
		q = `'''`
	default:
		return "", i
	}
	t = t[3:]
	if k := strings.Index(t, q); k >= 0 {
		return strings.TrimSpace(t[:k]), i + 1
	}
	var b strings.Builder
	b.WriteString(t)
	for i++; i < len(lines); i++ {
		if k := strings.Index(lines[i], q); k >= 0 {
			b.WriteString("\n" + strings.TrimSpace(lines[i][:k]))
			return dedent(b.String()), i + 1
		}
		b.WriteString("\n" + lines[i])
	}
	return dedent(b.String()), i
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	min := -1
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " "))
		if min < 0 || n < min {
			min = n
		}
	}
	if min > 0 {
		for i := 1; i < len(lines); i++ {
			if len(lines[i]) >= min {
				lines[i] = lines[i][min:]
			}
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func extras(doc string) (needs, example string) {
	if m := needsRe.FindStringSubmatch(doc); m != nil {
		needs = strings.TrimRight(strings.TrimSpace(m[1]), ".")
		if strings.HasPrefix(needs, "none") {
			needs = "none"
		}
	}
	if m := exampleRe.FindStringSubmatch(doc); m != nil {
		example = strings.TrimSpace(m[1])
	}
	return
}

// ---- search ----

// Match is a scored entry.
type Match struct {
	Entry
	Score int `json:"score"`
}

func tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// Search ranks entries for a free-text query: exact names first, then name and
// module words, then words in the documentation. module narrows to one module
// ("" for any, "global" for the host-level names).
func (x *Index) Search(query, module string, limit int) []Match {
	if limit <= 0 {
		limit = 10
	}
	q := strings.ToLower(strings.TrimSpace(query))
	qt := tokens(query)
	if q == "" {
		return nil
	}
	var out []Match
	for _, e := range x.Entries {
		if e.Kind == "module" && module != "" {
			continue
		}
		if module != "" && !(strings.EqualFold(e.Module, module) || (module == "global" && e.Module == "")) {
			continue
		}
		score := 0
		name, qual := strings.ToLower(e.Name), strings.ToLower(e.Qualified)
		switch {
		case q == qual || q == name:
			score += 100
		case strings.Contains(qual, q):
			score += 40
		}
		nameTok := tokens(strings.ReplaceAll(e.Qualified, "_", " ") + " " + e.Module)
		docTok := map[string]bool{}
		for _, w := range tokens(e.Doc + " " + e.Signature) {
			docTok[w] = true
		}
		covered := 0
		for _, t := range qt {
			hit := false
			for _, n := range nameTok {
				if n == t {
					score += 20
					hit = true
				} else if len(t) >= 4 && strings.HasPrefix(n, t) {
					score += 8
					hit = true
				}
			}
			if docTok[t] {
				score += 5
				hit = true
			}
			if hit {
				covered++
			}
		}
		// Entries that account for every query word beat those that match one
		// word loosely; this is what makes task-style queries work.
		if len(qt) > 1 && covered == len(qt) {
			score += 30 * covered
		}
		if score > 0 {
			if e.Kind == "module" {
				score += 3
			}
			out = append(out, Match{Entry: e, Score: score})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Qualified < out[j].Qualified
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Get finds an entry by qualified name ("http.get", "item") or, when the name
// is a bare function of a single module ("lower"), by its unique short name.
func (x *Index) Get(name string) (Entry, bool) {
	name = strings.TrimSpace(name)
	for _, e := range x.Entries { // exact case first: Item (class) and item() differ
		if e.Qualified == name && e.Kind != "module" {
			return e, true
		}
	}
	for _, e := range x.Entries {
		if strings.EqualFold(e.Qualified, name) && e.Kind == "function" {
			return e, true
		}
	}
	for _, e := range x.Entries {
		if strings.EqualFold(e.Qualified, name) && e.Kind != "module" {
			return e, true
		}
	}
	var hit []Entry
	for _, e := range x.Entries {
		if strings.EqualFold(e.Name, name) && e.Kind != "module" {
			hit = append(hit, e)
		}
	}
	if len(hit) == 1 {
		return hit[0], true
	}
	return Entry{}, false
}

// ModuleNames lists modules (with their function counts) sorted by name.
func (x *Index) ModuleNames() []string {
	var out []string
	for m := range x.Modules {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

// InModule lists the non-module entries of one module ("global" for the
// host-level names).
func (x *Index) InModule(module string) []Entry {
	var out []Entry
	for _, e := range x.Entries {
		if e.Kind == "module" {
			continue
		}
		if strings.EqualFold(e.Module, module) || (module == "global" && e.Module == "") {
			out = append(out, e)
		}
	}
	return out
}
