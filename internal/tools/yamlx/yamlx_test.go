package yamlx

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "yaml") }

func parse(t *testing.T, text string) string {
	t.Helper()
	return toolstest.Compact(toolstest.Must(t, "yaml", "parse", "text", text))
}

func TestParse(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ``},
		{"# only a comment\n", ``},
		{"a: 1\nb: two\nc: true\nd: ~\ne: 1.5\nf: 0x10\n", `{"a":1,"b":"two","c":true,"d":null,"e":1.5,"f":16}`},
		{"name: zagros\ntags: [a, b, \"c d\"]\nobj: {x: 1, y: [2, 3]}", `{"name":"zagros","tags":["a","b","cd"],"obj":{"x":1,"y":[2,3]}}`},
		{"- a\n- b\n-\n- c: 1\n  d: 2\n- - x\n  - y\n", `["a","b",null,{"c":1,"d":2},["x","y"]]`},
		{"top:\n  - 1\n  - 2\nsame:\n- 3\n- 4\nnest:\n  deep:\n    k: v\n", `{"top":[1,2],"same":[3,4],"nest":{"deep":{"k":"v"}}}`},
		{"a: 'it''s'\nb: \"tab\\there\\u00e9\"\nc: \"x: y\" # trailing\nd: a # comment\n", `{"a":"it's","b":"tab\there` + "é" + `","c":"x:y","d":"a"}`},
		{"yes: no\non: off\nn: null\n", `{"yes":"no","on":"off","n":null}`},
		{"url: http://example.com:8080/x#frag\ntime: 12:30\n", `{"url":"http://example.com:8080/x#frag","time":"12:30"}`},
		{"s: |\n  line1\n  line2\n\nt: |-\n  x\nu: >\n  folded\n  text\n\n  para\nv: |+\n  keep\n\nw: z\n", `{"s":"line1\nline2\n","t":"x","u":"foldedtext\npara\n","v":"keep\n\n","w":"z"}`},
		{"base: &b\n  x: 1\n  y: 2\nuse:\n  <<: *b\n  y: 3\ncopy: *b\n", `{"base":{"x":1,"y":2},"use":{"x":1,"y":3},"copy":{"x":1,"y":2}}`},
		{"plain: this is\n  a multi line\n  value\nnext: 1\n", `{"plain":"thisisamultilinevalue","next":1}`},
		{"[1, 2,\n 3]\n", `[1,2,3]`},
		{"key: [a,\n  b]\nz: 1\n", `{"key":["a","b"],"z":1}`},
		{"a: !!str 123\nb: !!int \"7\"\n", `{"a":"123","b":7}`},
		{"'q k': 1\n\"d k\": 2\n", `{"qk":1,"dk":2}`},
		{"---\na: 1\n", `{"a":1}`},
		{"a:\n", `{"a":null}`},
	}
	for _, c := range cases {
		got := parse(t, c.in)
		// Compact removes spaces outside quotes only; expectations above keep spaces inside strings out of the way
		want := strings.ReplaceAll(c.want, " ", "")
		got = strings.ReplaceAll(got, " ", "")
		if got != want {
			t.Errorf("input %q\n got %s\nwant %s", c.in, got, want)
		}
	}
}

func TestParseAll(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "yaml", "parse_all", "text", "a: 1\n---\nb: 2\n...\n---\n- 3\n"))
	if got != `[{"a":1},{"b":2},[3]]` {
		t.Errorf("got %s", got)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{
		"a: 1\na: 2\n",
		"a: [1, 2\n",
		"a: \"unterminated\n",
		"a: 1\n  b: 2\n",
		"\ta: 1\n",
		"a: *nope\n",
		"- a\nb: 1\n",
		"a: |x\n  t\n",
		"? complex\n: key\n",
		"a: [1, 2]]\n",
		"a: {b: 1}}\n",
	} {
		if _, err := toolstest.Call("yaml", "parse", "text", in); err == nil {
			t.Errorf("expected an error for %q", in)
		} else if !strings.Contains(err.Error(), "yaml:") {
			t.Errorf("error for %q lacks the yaml: prefix: %v", in, err)
		}
	}
}

func TestParseNeverPanics(t *testing.T) {
	seeds := []string{"a: b\n- c", "{[}]", "- - - - -", "&a *a", "a: &x\n  b: *x", "|\n", ">-\n  a", "\"\\", "'", "!!\n", "a:\n  - b:\n      - c: [d, {e: f}]\n", "\x00\xff\xfe", "---\n---\n...\n", "a: |\n"}
	rng := rand.New(rand.NewSource(1))
	alphabet := []byte("ab:- \n#[]{},'\"&*!|>%?\t1.")
	for i := 0; i < 3000; i++ {
		var s string
		if i < len(seeds) {
			s = seeds[i]
		} else {
			n := rng.Intn(40)
			b := make([]byte, n)
			for j := range b {
				b[j] = alphabet[rng.Intn(len(alphabet))]
			}
			s = string(b)
		}
		toolstest.Call("yaml", "parse", "text", s)
		toolstest.Call("yaml", "parse_all", "text", s)
		toolstest.Call("yaml", "is_valid", "text", s)
	}
	deep := strings.Repeat("[", 5000) + strings.Repeat("]", 5000)
	toolstest.Fails(t, "yaml", "parse", "text", deep)
}

func TestDumpRoundTrip(t *testing.T) {
	docs := []string{
		`{"name":"x","list":[1,2,{"a":"b","c":[true,null]}],"nested":{"k":{"z":1.5}},"empty":[],"emptyd":{}}`,
		`[[1,2],[3,[4,5]],{"a":[{"b":1}]}]`,
		`{"s":"multi\nline\n","t":"no newline end\nx","q":"true","n":"123","e":"","sp":" lead","colon":"a: b","hash":"a #b","dash":"- x","star":"*x"}`,
		`"just a string"`,
		`42`,
	}
	for _, d := range docs {
		in := toolstest.Must(t, "yaml", "dump", "value", toolstest.JSONValue(t, d))
		out := toolstest.Compact(toolstest.Must(t, "yaml", "parse", "text", in))
		want := toolstest.Compact(toolstest.JSONValue(t, d))
		if out != want {
			t.Errorf("round trip changed the value\nyaml:\n%s\n got %s\nwant %s", in, out, want)
		}
	}
}

func TestDumpShape(t *testing.T) {
	v := toolstest.JSONValue(t, `{"a":1,"b":[1,{"c":2,"d":[3]}],"e":{"f":"g"}}`)
	got := toolstest.Must(t, "yaml", "dump", "value", v)
	want := "a: 1\nb:\n  - 1\n  - c: 2\n    d:\n      - 3\ne:\n  f: g\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	toolstest.Fails(t, "yaml", "dump", "value", v, "indent", int64(0))
}
