package builtin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/tools"
)

var registerFakes sync.Once

// fake tools, registered once under unique names so they cannot clash with
// the real registry (this package does not link internal/tools/all).
func fakeTools() {
	registerFakes.Do(func() {
		str := func(name, doc string) tools.Param {
			return tools.Param{Name: name, Type: "str", Doc: doc, Required: true}
		}
		reg := func(f tools.Func) {
			if f.Doc == "" {
				f.Doc = "fake tool used by the launcher tests"
			}
			if f.Returns == "" {
				f.Returns = "str"
			}
			f.Example = f.Group + "." + f.Name + "(x)"
			f.Tags = []string{"fake"}
			tools.Register(f)
		}
		reg(tools.Func{Group: "text", Name: "zt_lower", Title: "Lowercase", Family: "text", Order: 1, Keywords: []string{"zlower"},
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) { return strings.ToLower(a.Str("text")), nil }})
		reg(tools.Func{Group: "text", Name: "zt_upper", Title: "Uppercase", Family: "text", Order: 2,
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) { return strings.ToUpper(a.Str("text")), nil }})
		reg(tools.Func{Group: "text", Name: "zt_cases", Title: "All cases", Family: "text", Order: 3, Returns: "dict[str, str]",
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) {
				return map[string]string{"lower": strings.ToLower(a.Str("text")), "upper": strings.ToUpper(a.Str("text"))}, nil
			}})
		reg(tools.Func{Group: "text", Name: "zt_fail", Title: "Always fails", Family: "text", Order: 9,
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) { return nil, fmt.Errorf("nope") }})
		reg(tools.Func{Group: "hash", Name: "zt_hmac", Title: "HMAC fake", Family: "hash", Order: 1,
			Params: []tools.Param{str("text", "message"), str("key", "secret")}, Input: "text",
			Fn: func(a tools.Args) (any, error) { return a.Str("key") + ":" + a.Str("text"), nil }})
		reg(tools.Func{Group: "hash", Name: "zt_bcrypt_hash", Title: "Slow hash", Family: "hash", Order: 2,
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) { return "slow(" + a.Str("text") + ")", nil }})
		reg(tools.Func{Group: "hash", Name: "zt_sha", Title: "Fake SHA", Family: "hash", Order: 0, Keywords: []string{"zsha"},
			Params: []tools.Param{str("text", "input")}, Fn: func(a tools.Args) (any, error) { return "sha(" + a.Str("text") + ")", nil }})
		reg(tools.Func{Group: "regex", Name: "zt_test", Title: "Regex match", Family: "regex", Returns: "bool", Input: "text",
			Params: []tools.Param{str("pattern", "regex"), str("text", "subject")},
			Fn:     func(a tools.Args) (any, error) { return strings.Contains(a.Str("text"), a.Str("pattern")), nil }})
		reg(tools.Func{Group: "num", Name: "zt_hexonly", Title: "Hex only", Family: "num", Params: []tools.Param{str("value", "hex")},
			Accepts: func(in string) bool { return strings.HasPrefix(in, "0x") },
			Fn:      func(a tools.Args) (any, error) { return "hex:" + a.Str("value"), nil }})
		reg(tools.Func{Group: "uuid", Name: "zt_make", Title: "Make id", Family: "uuid",
			Fn: func(a tools.Args) (any, error) { return "id-1", nil }})
	})
}

func queryTools(t *testing.T, q string) []core.Item {
	t.Helper()
	fakeTools()
	return newTools().Query(context.Background(), q)
}

func titles(items []core.Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Title)
	}
	return out
}

func TestToolsFamilyRowsScopeAndNoHistory(t *testing.T) {
	items := queryTools(t, "text Hello")
	got := strings.Join(titles(items), "|")
	for _, want := range []string{"hello", "HELLO"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "nope") {
		t.Fatal("failing variants must be hidden")
	}
	for _, it := range items {
		if it.Scope == nil || it.Scope.Provider != "tools" || it.Scope.Query != "Hello" || it.Scope.Label != "Text" {
			t.Fatalf("rows must carry the family scope seeded with the input: %+v", it.Scope)
		}
		if it.Action.Kind != "copy" || it.Action.Extra["nohistory"] != "1" {
			t.Fatalf("rows copy and never enter history: %+v", it.Action)
		}
	}
	// dict results become one row per key
	if !strings.Contains(got, "hello") || len(items) != 2 {
		t.Fatalf("lower and upper once each (the dict rows repeat them and are deduplicated): %v", titles(items))
	}
}

func TestToolsUnchangedTextIsHidden(t *testing.T) {
	items := queryTools(t, "text already lower")
	for _, it := range items {
		if it.Title == "already lower" {
			t.Fatalf("a variant that returns the input unchanged is noise: %+v", it)
		}
	}
}

func TestToolsFunctionKeywordAndUniqueName(t *testing.T) {
	if items := queryTools(t, "zsha abc"); len(items) != 1 || items[0].Title != "sha(abc)" {
		t.Fatalf("function keyword must select just that variant: %v", titles(items))
	}
	if items := queryTools(t, "zt_lower ABC"); len(items) != 1 || items[0].Title != "abc" {
		t.Fatalf("a unique function name is a keyword: %v", titles(items))
	}
}

func TestToolsParamsMultiArgAndSlow(t *testing.T) {
	// missing required key: only variants that need nothing else run, plus a hint
	items := queryTools(t, "hash hello")
	got := strings.Join(titles(items), "|")
	if !strings.Contains(got, "sha(hello)") || strings.Contains(got, "slow(") {
		t.Fatalf("slow variants stay out of the live list: %s", got)
	}
	items = queryTools(t, "hash hello key=k")
	got = strings.Join(titles(items), "|")
	if !strings.Contains(got, "k:hello") {
		t.Fatalf("name=value must set parameters: %s", got)
	}
	// "a :: b" fills the string params in order
	items = queryTools(t, "re cat :: the cat sat")
	if len(items) != 1 || items[0].Title != "true" {
		t.Fatalf("pattern :: text: %v", titles(items))
	}
}

func TestToolsAcceptsAndEmptyInputHint(t *testing.T) {
	if items := queryTools(t, "num 0xff"); len(items) != 1 || items[0].Title != "hex:0xff" {
		t.Fatalf("Accepts passes: %v", titles(items))
	}
	if items := queryTools(t, "num 12"); len(items) != 0 {
		t.Fatalf("text that no variant accepts stays quiet: %+v", items)
	}
	// "num" alone is just a word; with a trailing space the clipboard (whatever
	// it holds) is the input, so only check that nothing panics or stalls.
	fakeTools()
	_ = newTools().Query(core.WithRawQuery(context.Background(), "num "), "num")
	if items := queryTools(t, "num"); len(items) != 0 {
		t.Fatalf("a bare keyword stays quiet: %+v", items)
	}
}

func TestToolsScopeListsEverythingAndFilters(t *testing.T) {
	fakeTools()
	p := newTools()
	root := p.Query(context.Background(), "hash hello key=k")
	sc := root[0].Scope
	all := p.QueryScope(context.Background(), *sc, "hello key=k")
	got := strings.Join(titles(all), "|")
	if !strings.Contains(got, "slow(hello)") {
		t.Fatalf("the scope includes slow variants: %s", got)
	}
	filtered := p.QueryScope(context.Background(), *sc, "@sha hello")
	if len(filtered) != 1 || filtered[0].Title != "sha(hello)" {
		t.Fatalf("@word narrows the scope: %v", titles(filtered))
	}
	// the variants of the keyword you typed come first
	pref := core.Scope{Provider: "tools", Data: map[string]string{"family": "hash", "prefer": "zsha"}}
	first := p.QueryScope(context.Background(), pref, "hello key=k")
	if len(first) == 0 || first[0].Title != "sha(hello)" {
		t.Fatalf("preferred variant first: %v", titles(first))
	}
}

func TestToolsGeneratorsNeedNoInput(t *testing.T) {
	items := queryTools(t, "uuid")
	if len(items) != 1 || items[0].Title != "id-1" {
		t.Fatalf("generators run without input: %v", titles(items))
	}
}

func TestToolsRegisteredAsProviderWhenRegistryHasFunctions(t *testing.T) {
	fakeTools()
	cfg := core.DefaultConfig()
	var names []string
	for _, p := range All(&cfg) {
		names = append(names, p.Name())
	}
	if !strings.Contains(strings.Join(names, ","), "tools") {
		t.Fatalf("tools provider missing: %v", names)
	}
	// the old text provider hands its overlapping keywords over
	for _, p := range All(&cfg) {
		if tp, ok := p.(*text); ok {
			for _, k := range tp.Keywords() {
				if k == "b64" || k == "hash" || k == "uuid" {
					t.Fatalf("text provider still answers %q", k)
				}
			}
		}
	}
}

func TestToolsFallbackOffersFamilyScopes(t *testing.T) {
	fakeTools() // only text, hash, regex, num and uuid exist here; missing families are skipped
	p := newTools()
	famOf := func(items []core.Item) []string {
		var out []string
		for _, it := range items {
			out = append(out, it.Scope.Data["family"])
		}
		return out
	}
	num := p.Fallback("14")
	if got := strings.Join(famOf(num), ","); got != "num,hash" || num[0].Title != "Number 14" || num[0].Scope.Query != "14" {
		t.Fatalf("a number offers Number then the everyday families: %s %+v", got, num)
	}
	for _, q := range []string{"-3.5", "0xff", "1,000", "1e9"} {
		if fb := p.Fallback(q); fb[0].Scope.Data["family"] != "num" {
			t.Fatalf("%q is a number: %v", q, famOf(fb))
		}
	}
	str := p.Fallback("hello world")
	if got := strings.Join(famOf(str), ","); got != "text,hash" || str[0].Title != "String hello world" {
		t.Fatalf("other input offers String then Hash: %s", got)
	}
	if got := famOf(p.Fallback("123e4567-e89b-12d3-a456-426614174000")); got[0] != "uuid" {
		t.Fatalf("a UUID leads with the IDs family: %v", got)
	}
	if p.Fallback("  ") != nil || p.Fallback(strings.Repeat("x", 500)) != nil {
		t.Fatal("empty and very long input get no row")
	}
	all := p.QueryScope(context.Background(), *num[0].Scope, "0xff")
	if len(all) == 0 {
		t.Fatal("the scope lists the family variants")
	}
}
