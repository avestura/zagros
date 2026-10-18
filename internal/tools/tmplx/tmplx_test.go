package tmplx

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "template") }

func rend(t *testing.T, tpl, data string, kv ...any) string {
	t.Helper()
	args := append([]any{"template", tpl, "data", toolstest.JSONValue(t, data)}, kv...)
	return toolstest.Must(t, "template", "render", args...).(string)
}

func TestRender(t *testing.T) {
	cases := []struct{ tpl, data, want string }{
		{"Hello {{ .name }}!", `{"name":"Ada"}`, "Hello Ada!"},
		{"{{ range .tags }}[{{ . | upper }}]{{ end }}", `{"tags":["a","b"]}`, "[A][B]"},
		{"{{ if .ok }}yes{{ else }}no{{ end }}", `{"ok":false}`, "no"},
		{"{{ with .user }}{{ .name }}{{ end }}", `{"user":{"name":"Bo"}}`, "Bo"},
		{"{{ .missing }}|{{ .a.b }}", `{"a":{}}`, "|"},
		{"{{ range $i, $e := .xs }}{{ $i }}={{ $e }};{{ end }}", `{"xs":["p","q"]}`, "0=p;1=q;"},
		{`{{ .s | trim | title }}`, `{"s":"  hello world  "}`, "Hello World"},
		{`{{ replace "a" "b" .s }} {{ split "," .s | join "-" }}`, `{"s":"a,a"}`, "b,b a-a"},
		{`{{ add .n 2 }} {{ div .n 4 }} {{ mul 1.5 2 }} {{ mod 7 3 }}`, `{"n":10}`, "12 2.5 3 1"},
		{`{{ default "x" .none }}{{ default "x" .set }}`, `{"none":"","set":"y"}`, "xy"},
		{`{{ .m | json }}`, `{"m":{"a":[1,2]}}`, `{"a":[1,2]}`},
		{`{{ keys .m | join "," }}`, `{"m":{"b":1,"a":2}}`, "a,b"},
		{`{{ dict "k" 1 | json }} {{ list 1 2 | json }}`, `{}`, `{"k":1} [1,2]`},
		{`{{ printf "%05.1f|%s" .f .s }} {{ len .s }}`, `{"f":3.14159,"s":"abc"}`, "003.1|abc 3"},
		{`{{ if and .a (gt .n 3) }}big{{ end }}`, `{"a":true,"n":5}`, "big"},
		{`{{ repeat 3 "ab" }}{{ quote .s }}`, `{"s":"q\"x"}`, `ababab"q\"x"`},
		{`{{/* note */ -}}  x`, `{}`, "x"},
		{`{{ .data }}`, `{"data":null}`, ""},
	}
	for _, c := range cases {
		if got := rend(t, c.tpl, c.data); got != c.want {
			t.Errorf("%s\n got %q\nwant %q", c.tpl, got, c.want)
		}
	}
	if got := toolstest.Must(t, "template", "render", "template", "no data {{ 1 }}"); got != "no data 1" {
		t.Errorf("got %q", got)
	}
	if got := toolstest.Must(t, "template", "render", "template", "<<.x>>-{{.x}}", "data", toolstest.JSONValue(t, `{"x":1}`), "left", "<<", "right", ">>"); got != "1-{{.x}}" {
		t.Errorf("custom delimiters: got %q", got)
	}
}

func TestStrictAndErrors(t *testing.T) {
	toolstest.Fails(t, "template", "render", "template", "{{ .nope }}", "data", toolstest.JSONValue(t, `{}`), "strict", true)
	for _, bad := range []string{
		"{{ .x", "{{ if }}x{{ end }}", "{{ end }}", "{{ nofunc 1 }}", "{{ template \"x\" }}", "{{ define \"x\" }}y{{ end }}",
		"{{ block \"b\" . }}z{{ end }}", "{{ call .f }}", "{{ range 99999999 }}x{{ end }}", "{{ .x | missingfn }}",
		"{{ div 1 0 }}", "{{ repeat 99999999 \"x\" }}", "{{ index .l 99 }}", "{{ join \",\" 5 }}",
	} {
		if _, err := toolstest.Call("template", "render", "template", bad, "data", toolstest.JSONValue(t, `{"l":[1]}`)); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
	toolstest.Fails(t, "template", "render", "template", strings.Repeat("x", 300<<10))
	for in, want := range map[string]bool{"{{ if .a }}x{{ end }}": true, "{{ if .a }}": false, "{{ template \"a\" }}": false, "plain": true} {
		if got := toolstest.Must(t, "template", "is_valid", "template", in); got != want {
			t.Errorf("is_valid(%q) = %v", in, got)
		}
	}
}

func TestLimits(t *testing.T) {
	// nested ranges over data produce output until the cap, without hanging
	big := make([]any, 3000)
	for i := range big {
		big[i] = "x"
	}
	_, err := toolstest.Call("template", "render", "template", "{{ range $.a }}{{ range $.a }}{{ . }}{{ end }}{{ end }}", "data", mkData(big))
	if err == nil || !strings.Contains(err.Error(), "output is larger") {
		t.Errorf("expected the output cap, got %v", err)
	}
	// a data-driven silent loop that finishes quickly is fine
	if _, err := toolstest.Call("template", "render", "template", "{{ range .a }}{{ end }}done", "data", mkData(big)); err != nil {
		t.Error(err)
	}
}

func mkData(l []any) any {
	d := tools.NewDict()
	d.Set("a", l)
	return d
}

func TestNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	pieces := []string{"{{", "}}", ".a", " ", "range", "if", "end", "else", "with", "$x", ":=", "|", "\"s\"", "1", "upper", "len", "index", "printf", "-", "/*", "*/", "(", ")", "template", "x"}
	for i := 0; i < 4000; i++ {
		var b strings.Builder
		for j := rng.Intn(12); j > 0; j-- {
			b.WriteString(pieces[rng.Intn(len(pieces))])
		}
		toolstest.Call("template", "render", "template", b.String(), "data", toolstest.JSONValue(t, `{"a":[1,"x",{"b":2}]}`))
		toolstest.Call("template", "is_valid", "template", b.String())
	}
}
