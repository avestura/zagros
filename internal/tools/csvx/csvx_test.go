package csvx

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "csv") }

func TestParse(t *testing.T) {
	cases := []struct {
		args []any
		want string
	}{
		{[]any{"text", "a,b\n1,2"}, `[["a","b"],["1","2"]]`},
		{[]any{"text", "a;b\n1;2"}, `[["a","b"],["1","2"]]`},
		{[]any{"text", "a\tb\n1\t2", "header", true}, `[{"a":"1","b":"2"}]`},
		{[]any{"text", "\uFEFFa,b\r\n\"x,1\",\"y\"\"z\"\r\n", "header", true}, `[{"a":"x,1","b":"y\"z"}]`},
		{[]any{"text", "a,b\n1", "header", true}, `[{"a":"1","b":""}]`},
		{[]any{"text", "a,b\n1,2,3", "header", true}, `[{"a":"1","b":"2","col3":"3"}]`},
		{[]any{"text", "# c\na, b\n", "comment", "#", "trim", true}, `[["a","b"]]`},
		{[]any{"text", ""}, `[]`},
	}
	for _, c := range cases {
		if got := toolstest.Compact(toolstest.Must(t, "csv", "parse", c.args...)); got != c.want {
			t.Errorf("%v: got %s, want %s", c.args, got, c.want)
		}
	}
}

func TestParseMalformedNeverPanics(t *testing.T) {
	for _, s := range []string{"\"", "a,\"b", "\"a\"b\",c", "\x00\xff", strings.Repeat("\"", 99), "a,b\n\"", "\r\r\n\n,"} {
		for _, strict := range []bool{false, true} {
			toolstest.Call("csv", "parse", "text", s, "strict", strict, "header", true)
		}
	}
	toolstest.Fails(t, "csv", "parse", "text", "a,\"b", "strict", true)
	toolstest.Fails(t, "csv", "parse", "text", "a", "delimiter", "ab")
	toolstest.Fails(t, "csv", "parse", "text", "a", "comment", "\"")
}

func TestDump(t *testing.T) {
	d1 := tools.NewDict()
	d1.Set("name", "ada, l")
	d1.Set("age", int64(36))
	d2 := tools.NewDict()
	d2.Set("name", "bob")
	d2.Set("city", nil)
	got := toolstest.Must(t, "csv", "dump", "rows", []any{d1, d2}, "header", true)
	if want := "name,age,city\n\"ada, l\",36,\nbob,,\n"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	got = toolstest.Must(t, "csv", "dump", "rows", []any{[]any{"a", true, 1.5}}, "delimiter", ";")
	if got != "a;true;1.5\n" {
		t.Errorf("got %q", got)
	}
	toolstest.Fails(t, "csv", "dump", "rows", []any{[]any{[]any{}}})
	toolstest.Fails(t, "csv", "dump", "rows", "x")
	toolstest.Fails(t, "csv", "dump", "rows", []any{"x"})
}

func TestDetect(t *testing.T) {
	for in, want := range map[string]string{"a;b\n1;2": ";", "a|b|c": "|", "a\tb": "\t", "abc": ",", "\"a;b\",c": ","} {
		if got := toolstest.Must(t, "csv", "detect_delimiter", "text", in); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
}
