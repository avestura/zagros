package regexx

import (
	"regexp"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "regex") }

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	v, err := tools.Call("regex", name, m)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return v
}

func eq(t *testing.T, got any, want string) {
	t.Helper()
	if g := tools.Format(got); g != want {
		t.Errorf("got %q, want %q", g, want)
	}
}

func TestBasics(t *testing.T) {
	eq(t, call(t, "test", "pattern", `^\d+$`, "text", "123"), "true")
	eq(t, call(t, "test", "pattern", "HELLO", "text", "say hello", "flags", "i"), "true")
	eq(t, call(t, "match", "pattern", `\d+`, "text", "a 66 b 7"), "66")
	eq(t, call(t, "find_all", "pattern", `\d+`, "text", "1 22 333", "limit", int64(2)), "1\n22")
	eq(t, call(t, "count", "pattern", "a", "text", "banana"), "3")
	eq(t, call(t, "positions", "pattern", "an", "text", "éanana"), "1-3: an\n3-5: an")
	eq(t, call(t, "replace", "pattern", `(\w+)@(\w+)`, "text", "bob@home", "replacement", "$2:$1"), "home:bob")
	eq(t, call(t, "replace", "pattern", `a`, "text", "aaa", "replacement", "b", "limit", int64(2)), "bba")
	eq(t, call(t, "replace_literal", "pattern", `\d+`, "text", "a1b22", "replacement", "$#"), "a$#b$#")
	eq(t, call(t, "split", "pattern", `[,;]\s*`, "text", "a, b;c"), "a\nb\nc")
	eq(t, call(t, "escape", "text", "1+1=2?"), `1\+1=2\?`)
	eq(t, call(t, "is_valid", "pattern", "(a"), "false")
	eq(t, call(t, "is_valid", "pattern", `(?<=a)b`), "false")
	eq(t, call(t, "validate", "pattern", "a(b"), "missing closing ): `a(b`")
	eq(t, call(t, "validate", "pattern", "a+"), "ok")
	eq(t, call(t, "lines_matching", "pattern", "^E", "text", "E1\nok\nE2"), "E1\nE2")
	eq(t, call(t, "lines_matching", "pattern", "^E", "text", "E1\nok\nE2", "invert", true), "ok")
	eq(t, call(t, "extract_between", "text", "a[1] b[22] c[", "start", "[", "end", "]"), "1\n22")
	eq(t, call(t, "named_group_names", "pattern", `(?P<y>\d+)-(\d+)-(?P<d>\d+)`), "y\nd")
	eq(t, call(t, "literal_prefix", "pattern", `abc\d+`), "abc")
}

func TestGroups(t *testing.T) {
	g := call(t, "groups", "pattern", `(?P<y>\d{4})-(\d\d)`, "text", "on 2026-10-17").(map[string]string)
	if g["y"] != "2026" || g["1"] != "2026" || g["2"] != "10" || g["0"] != "2026-10" {
		t.Errorf("groups: %v", g)
	}
	eq(t, call(t, "all_groups", "pattern", `(\w)=(\d)`, "text", "a=1 b=2"),
		`{"0":"a=1","1":"a","2":"1"}`+"\n"+`{"0":"b=2","1":"b","2":"2"}`)
	if len(call(t, "groups", "pattern", "x", "text", "y").(map[string]string)) != 0 {
		t.Error("no match must give empty dict")
	}
}

func TestErrors(t *testing.T) {
	if _, err := tools.Call("regex", "test", map[string]any{"pattern": "(", "text": "x"}); err == nil {
		t.Error("want error for bad pattern")
	}
	if _, err := tools.Call("regex", "test", map[string]any{"pattern": "a", "text": "x", "flags": "z"}); err == nil {
		t.Error("want error for bad flag")
	}
	big := strings.Repeat("a", 5<<20)
	if _, err := tools.Call("regex", "test", map[string]any{"pattern": "a", "text": big}); err == nil {
		t.Error("want error for oversized text")
	}
}

func TestWildcardAndGlob(t *testing.T) {
	cases := []struct {
		p, s string
		want bool
	}{
		{"*.txt", "notes.txt", true}, {"*.txt", "notes.txt.bak", false}, {"a?c", "abc", true},
		{"a?c", "ac", false}, {"*", "", true}, {"a*b*c", "aXXbYYc", true}, {"a*b*c", "aXXbYY", false},
	}
	for _, c := range cases {
		if got := wildcard(c.p, c.s, false); got != c.want {
			t.Errorf("wildcard(%q,%q)=%v", c.p, c.s, got)
		}
	}
	if !wildcard("*.TXT", "a.txt", true) {
		t.Error("ignore case")
	}
	re := globToRegex("*.{jpg,png}")
	if re != `^.*\.(?:jpg|png)$` {
		t.Errorf("glob: %s", re)
	}
	if !regexp.MustCompile(globToRegex("file?.[ch]")).MatchString("file1.c") {
		t.Error("glob class")
	}
	g, err := regexToGlob(`^img_.*\.png$`)
	if err != nil || g != "img_*.png" {
		t.Errorf("to_glob %q %v", g, err)
	}
	if _, err := regexToGlob(`(a|b)`); err == nil {
		t.Error("to_glob should reject groups")
	}
}

func TestSed(t *testing.T) {
	cases := [][3]string{
		{"hello world", "s/o/0/g", "hell0 w0rld"},
		{"hello world", "s/o/0/", "hell0 world"},
		{"hello world", "s/o/0/2", "hello w0rld"},
		{"hello world", "s/(l+)/[\\1]/", "he[ll]o world"},
		{"abc", "s|b|&&|", "abbc"},
		{"a/b", `s/\//-/`, "a-b"},
		{"Hello", "s/h/J/i", "Jello"},
		{"cost", "s/o/$/", "c$st"},
		{"a a", "s/a/b/\ns/b/c/g", "c a"},
	}
	for _, c := range cases {
		got, err := sed(c[0], c[1])
		if err != nil || got != c[2] {
			t.Errorf("sed(%q,%q)=%q,%v want %q", c[0], c[1], got, err, c[2])
		}
	}
	if _, err := sed("x", "bogus"); err == nil {
		t.Error("want error")
	}
}

func TestExplain(t *testing.T) {
	out := strings.Join(explain(`^(?P<n>\d{3})-[a-z]+(?:x|y)*?$`), "\n")
	for _, want := range []string{"start of text", `group named "n"`, "a digit", "exactly 3 times", "a to z", "one or more", "OR", "lazy", "end of text"} {
		if !strings.Contains(out, want) {
			t.Errorf("explain missing %q in:\n%s", want, out)
		}
	}
	if !strings.Contains(strings.Join(explain(`(?=a)b`), ""), "lookaround") {
		t.Error("lookahead")
	}
	if !strings.Contains(strings.Join(explain(`ab+`), "\n"), `literal "a"`) {
		t.Error("literal split before quantifier")
	}
	explain(`[`)
	explain(`\`)
	explain(`(((`)
	explain(`{3`)
}

func TestPresets(t *testing.T) {
	for _, p := range presetList {
		if _, err := regexp.Compile(p.pattern); err != nil {
			t.Errorf("preset %s: %v", p.name, err)
		}
	}
	ok := map[string][]string{
		"email":  {"bob.smith+tag@example.co.uk"},
		"ipv4":   {"192.168.1.1", "0.0.0.0", "255.255.255.255"},
		"ipv6":   {"::1", "2001:db8::ff00:42:8329", "fe80::1"},
		"uuid":   {"123e4567-e89b-12d3-a456-426614174000"},
		"semver": {"1.2.3", "v1.0.0-rc.1+build.5"},
		"mac":    {"aa:bb:cc:dd:ee:ff", "AA-BB-CC-DD-EE-FF"},
	}
	bad := map[string][]string{
		"ipv4":  {"256.1.1.1", "1.2.3"},
		"email": {"not-an-email", "a@b"},
		"uuid":  {"123e4567-e89b-12d3-a456-42661417400"},
	}
	for name, xs := range ok {
		for _, x := range xs {
			if !call(t, "test_preset", "name", name, "text", x, "full", true).(bool) {
				t.Errorf("%s should match %q", name, x)
			}
		}
	}
	for name, xs := range bad {
		for _, x := range xs {
			if call(t, "test_preset", "name", name, "text", x, "full", true).(bool) {
				t.Errorf("%s should not match %q", name, x)
			}
		}
	}
	eq(t, call(t, "preset", "name", "Hex Color"), presetList[8].pattern)
	if _, err := tools.Call("regex", "preset", map[string]any{"name": "nope"}); err == nil {
		t.Error("unknown preset must fail")
	}
	if len(call(t, "presets").(map[string]string)) < 40 {
		t.Error("expected many presets")
	}
}
