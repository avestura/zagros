// Package tools is the registry of Zagros' pure utility functions: string,
// number, encoding, hash, cipher, random, regex, network, date/time and
// extraction helpers.
//
// One registration feeds three consumers, so they can never drift apart:
//   - the launcher (internal/builtin/tools.go shows families of variants),
//   - the extension API (internal/ext exposes each group as a Starlark module),
//   - the generated stubs, docs and MCP search (cmd/toolsgen, internal/mcp).
//
// Functions here are pure: no I/O, no clipboard, no network, no clock except
// where the signature says so (dt.now). That is why extensions can call them
// without any permission.
package tools

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Param describes one parameter of a Func.
type Param struct {
	Name string
	// Type is the stub type: "str", "int", "float", "bool", "list[str]",
	// "list[float]" or "dict[str, str]".
	Type    string
	Doc     string
	Default any // used when the caller omits the parameter; nil means the zero value of Type
	// Required parameters have no default and must be passed.
	Required bool
	// Choices lists the allowed values of a str parameter (documentation and
	// launcher completion only; Fn must still validate).
	Choices []string
}

// Func is one function of the API.
type Func struct {
	// Group is the Starlark module and stub file: text, lines, num, math,
	// encode, hash, cipher, rand, uuid, regex, net, dt, extract.
	Group string
	// Name is the snake_case function name, unique inside Group.
	Name string
	// Title is the human label shown in the launcher ("Base64 encode").
	Title string
	// Doc explains what it does, in one or two sentences, including edge
	// cases. It becomes the stub docstring and is searched by the MCP server.
	Doc    string
	Params []Param
	// Returns is the stub return type: "str", "int", "float", "bool",
	// "list[str]", "list[float]", "dict[str, str]" or "None".
	Returns string
	// Fn does the work. Args are already converted and defaulted per Params.
	// The result must be string, int64, float64, bool, []string, []float64,
	// map[string]string or nil, matching Returns.
	Fn func(a Args) (any, error)
	// Example is one runnable line of Starlark, e.g. text.lower("ABC").
	Example string
	// Tags are extra search words ("base64", "encode", "rfc4648").
	Tags []string

	// ---- launcher metadata (all optional) ----

	// Family is the launcher family this variant is listed under (usually the
	// same as Group). Empty keeps the function API-only.
	Family string
	// Order sorts variants inside a family (lower first).
	Order int
	// Input is the name of the str parameter that receives the text the user
	// typed. Default: the first str parameter.
	Input string
	// Accepts, when set, is checked against the input before the variant is
	// offered (decoders return false for text they cannot decode).
	Accepts func(input string) bool
	// Keywords are extra launcher keywords that select just this variant or a
	// small set of variants (e.g. "b64" for the base64 variants).
	Keywords []string
}

// Key is "group.name".
func (f Func) Key() string { return f.Group + "." + f.Name }

// InputParam returns the name of the launcher input parameter.
func (f Func) InputParam() string {
	if f.Input != "" {
		return f.Input
	}
	for _, p := range f.Params {
		if p.Type == "str" {
			return p.Name
		}
	}
	return ""
}

// Args holds converted arguments by parameter name.
type Args map[string]any

func (a Args) Str(n string) string {
	if s, ok := a[n].(string); ok {
		return s
	}
	return ""
}

func (a Args) Int(n string) int64 {
	switch v := a[n].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	}
	return 0
}

func (a Args) Float(n string) float64 {
	switch v := a[n].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	}
	return 0
}

func (a Args) Bool(n string) bool { b, _ := a[n].(bool); return b }

func (a Args) List(n string) []string { l, _ := a[n].([]string); return l }

func (a Args) Floats(n string) []float64 { l, _ := a[n].([]float64); return l }

var (
	mu    sync.RWMutex
	funcs []Func
	index = map[string]int{}
)

// Register adds f. It panics on a duplicate or malformed registration so
// mistakes surface in tests, not at runtime.
func Register(f Func) {
	mu.Lock()
	defer mu.Unlock()
	if f.Group == "" || f.Name == "" || f.Fn == nil || f.Title == "" || f.Doc == "" || f.Returns == "" {
		panic(fmt.Sprintf("tools: incomplete registration %q (group, name, title, doc, returns and Fn are required)", f.Key()))
	}
	if _, dup := index[f.Key()]; dup {
		panic("tools: duplicate " + f.Key())
	}
	seen := map[string]bool{}
	for _, p := range f.Params {
		if p.Name == "" || p.Type == "" || seen[p.Name] {
			panic(fmt.Sprintf("tools: %s has a bad parameter %q", f.Key(), p.Name))
		}
		seen[p.Name] = true
	}
	index[f.Key()] = len(funcs)
	funcs = append(funcs, f)
}

// All returns every function ordered by group, then Order, then name.
func All() []Func {
	mu.RLock()
	out := append([]Func(nil), funcs...)
	mu.RUnlock()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Lookup finds group.name.
func Lookup(group, name string) (Func, bool) {
	mu.RLock()
	defer mu.RUnlock()
	i, ok := index[group+"."+name]
	if !ok {
		return Func{}, false
	}
	return funcs[i], true
}

// Groups returns the distinct group names, sorted.
func Groups() []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range All() {
		if !seen[f.Group] {
			seen[f.Group] = true
			out = append(out, f.Group)
		}
	}
	return out
}

// Call runs group.name with named arguments, applying defaults and checking
// required parameters. Values must already have the Go types of the Params.
func Call(group, name string, named map[string]any) (any, error) {
	f, ok := Lookup(group, name)
	if !ok {
		return nil, fmt.Errorf("tools: unknown function %s.%s", group, name)
	}
	return f.Invoke(named)
}

// Invoke runs f with named arguments (see Call).
func (f Func) Invoke(named map[string]any) (any, error) {
	a := Args{}
	for _, p := range f.Params {
		v, ok := named[p.Name]
		switch {
		case ok:
			a[p.Name] = v
		case p.Required:
			return nil, fmt.Errorf("%s: missing argument %q", f.Key(), p.Name)
		default:
			a[p.Name] = p.Default
			if p.Default == nil {
				a[p.Name] = zeroOf(p.Type)
			}
		}
	}
	return f.Fn(a)
}

func zeroOf(t string) any {
	switch t {
	case "str":
		return ""
	case "int":
		return int64(0)
	case "float":
		return float64(0)
	case "bool":
		return false
	case "list[str]":
		return []string(nil)
	case "list[float]":
		return []float64(nil)
	case "dict[str, str]":
		return map[string]string(nil)
	}
	return nil
}

// Format renders a result for display and copying: strings as is, lists one
// element per line, maps as "key: value" lines, numbers without noise.
func Format(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case int64:
		return fmt.Sprintf("%d", x)
	case int:
		return fmt.Sprintf("%d", x)
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.12f", x), "0"), ".")
	case bool:
		if x {
			return "true"
		}
		return "false"
	case []string:
		return strings.Join(x, "\n")
	case []float64:
		parts := make([]string, len(x))
		for i, f := range x {
			parts[i] = Format(f)
		}
		return strings.Join(parts, "\n")
	case map[string]string:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines := make([]string, len(keys))
		for i, k := range keys {
			lines[i] = k + ": " + x[k]
		}
		return strings.Join(lines, "\n")
	}
	return fmt.Sprint(v)
}
