package ext

import (
	"strings"
	"testing"

	_ "github.com/avestura/zagros/internal/tools/all"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// runStar executes a script with the tool modules predeclared and returns its
// globals.
func runStar(t *testing.T, src string) (starlark.StringDict, error) {
	t.Helper()
	th := &starlark.Thread{Name: "test"}
	return starlark.ExecFileOptions(&syntax.FileOptions{}, th, "t.star", src, toolModules())
}

func TestStructuredToolsFromStarlark(t *testing.T) {
	g, err := runStar(t, `
y = yaml.parse("name: zagros\nports: [80, 443]\nnested:\n  on: true\n  none: ~\n")
y_out = yaml.dump(y)
y_back = yaml.parse(y_out)
y_keys = list(y.keys())
docs = yaml.parse_all("a: 1\n---\nb: 2")

tm = toml.parse("title = 'x'\n[db]\nport = 5432\n")
tm_out = toml.dump({"z": 1, "a": {"b": [1, 2]}})

rows = csv.parse("name,age\nada,36\nbob,41", header=True)
csv_out = csv.dump(rows)
csv_out_lists = csv.dump([["a", 1, True, None]], delimiter=";")

tree = xml.parse("<l><b id='1'>x</b><b id='2'>y</b></l>")
xml_ids = xml.query(tree, "/l/b/@id")
xml_text = xml.query("<l><b>x</b></l>", "//b/text()")
xml_out = xml.dump({"tag": "a", "attrs": {"k": "v"}, "children": [{"tag": "b", "text": "t"}]}, indent="")

links = html.attr("<p><a href='/x'>x</a><a href='/y'>y</a></p>", "p > a", "href")
sel = html.select("<ul><li class=on>a</li><li>b</li></ul>", "li.on")
ptext = html.text("<h1>Hi</h1><p>there <b>you</b></p>")

u = url.parse("https://u@example.com:8080/a?x=1&x=2#f")
u_built = url.build({"scheme": "https", "host": "example.com", "path": "/p", "query": {"q": "a b", "n": 2}})
u_join = url.join("https://e.com/a/b", "../c")
u_q = url.query_parse("a=1&b=2")
u_norm = url.normalize("HTTP://Example.com:80/a/../b")

p_join = path.join(["/a", "b", "../c"])
p_win = path.base("C:\\Users\\me\\a.txt", flavor="windows")
p_rel = path.rel("/a/b", "/a/c/d")
p_parts = path.parts("/a/b")

t_out = template.render("Hi {{ .name }}: {{ range .xs }}{{ . | upper }}{{ end }}", {"name": "Ada", "xs": ["a", "b"]})

tok = jwt.sign({"sub": "1"}, "secret", expires_in=60, now=1000)
claims = jwt.verify(tok, "secret", alg="HS256", now=1010)
dec = jwt.decode(tok)
ok_expired = jwt.is_valid(tok, "secret", now=2000)
`)
	if err != nil {
		t.Fatal(err)
	}
	str := func(name string) string {
		s, ok := g[name].(starlark.String)
		if !ok {
			t.Fatalf("%s is %v (%s)", name, g[name], g[name].Type())
		}
		return string(s)
	}
	repr := func(name string) string { return g[name].String() }
	eq := func(name, want string) {
		t.Helper()
		if got := repr(name); got != want {
			t.Errorf("%s = %s, want %s", name, got, want)
		}
	}

	eq("y", `{"name": "zagros", "ports": [80, 443], "nested": {"on": True, "none": None}}`)
	eq("y_back", repr("y"))
	eq("y_keys", `["name", "ports", "nested"]`)
	eq("docs", `[{"a": 1}, {"b": 2}]`)
	eq("tm", `{"title": "x", "db": {"port": 5432}}`)
	if got := str("tm_out"); got != "z = 1\n\n[a]\nb = [1, 2]\n" {
		t.Errorf("toml.dump = %q", got)
	}
	eq("rows", `[{"name": "ada", "age": "36"}, {"name": "bob", "age": "41"}]`)
	if got := str("csv_out"); got != "name,age\nada,36\nbob,41\n" {
		t.Errorf("csv.dump = %q", got)
	}
	if got := str("csv_out_lists"); got != "a;1;true;\n" {
		t.Errorf("csv.dump lists = %q", got)
	}
	eq("xml_ids", `["1", "2"]`)
	eq("xml_text", `["x"]`)
	if got := str("xml_out"); got != `<a k="v"><b>t</b></a>` {
		t.Errorf("xml.dump = %q", got)
	}
	eq("links", `["/x", "/y"]`)
	if s := repr("sel"); !strings.Contains(s, `"tag": "li"`) || !strings.Contains(s, `"text": "a"`) {
		t.Errorf("html.select = %s", s)
	}
	if got := str("ptext"); got != "Hi\nthere you" {
		t.Errorf("html.text = %q", got)
	}
	if s := repr("u"); !strings.Contains(s, `"query": {"x": ["1", "2"]}`) || !strings.Contains(s, `"user": "u"`) {
		t.Errorf("url.parse = %s", s)
	}
	if got := str("u_built"); got != "https://example.com/p?q=a+b&n=2" {
		t.Errorf("url.build = %q", got)
	}
	if got := str("u_join"); got != "https://e.com/c" {
		t.Errorf("url.join = %q", got)
	}
	eq("u_q", `{"a": "1", "b": "2"}`)
	if got := str("u_norm"); got != "http://example.com/b" {
		t.Errorf("url.normalize = %q", got)
	}
	if got := str("p_join"); got != "/a/c" {
		t.Errorf("path.join = %q", got)
	}
	if got := str("p_win"); got != "a.txt" {
		t.Errorf("path.base = %q", got)
	}
	if got := str("p_rel"); got != "../c/d" {
		t.Errorf("path.rel = %q", got)
	}
	eq("p_parts", `["/", "a", "b"]`)
	if got := str("t_out"); got != "Hi Ada: AB" {
		t.Errorf("template.render = %q", got)
	}
	if s := repr("claims"); !strings.Contains(s, `"payload": {"sub": "1", "exp": 1060}`) {
		t.Errorf("jwt.verify = %s", s)
	}
	if s := repr("dec"); !strings.Contains(s, `"alg": "HS256"`) {
		t.Errorf("jwt.decode = %s", s)
	}
	eq("ok_expired", "False")
}

func TestStructuredToolsErrorsAreCleanStarlarkErrors(t *testing.T) {
	cases := map[string]string{
		`yaml.parse("a: [1, 2")`:                                "yaml:",
		`toml.parse("a = ")`:                                    "toml:",
		`xml.parse("<a><b></a>")`:                               "xml:",
		`csv.dump([[[1]]])`:                                     "CSV cell",
		`csv.dump({"a": 1})`:                                    "rows must be",
		`html.select("<p>", "a[")`:                              "bad selector",
		`url.build({"port": "x"})`:                              "port",
		`template.render("{{ .x", {})`:                          "unclosed action",
		`template.render("{{ template \"a\" }}", {})`:           "not allowed",
		`jwt.verify("a.b", "k")`:                                "three",
		`jwt.sign({"a": 1}, "k", alg="none")`:                   "unsupported algorithm",
		`yaml.dump({1: 2})`:                                     "keys must be str",
		`yaml.dump(lambda: 1)`:                                  "unsupported value",
		`path.rel("/a", "b")`:                                   "absolute",
		`jwt.verify(jwt.sign({"a": 1}, "k"), "other")`:          "signature",
		`toml.dump({"a": None})`:                                "None cannot be written",
		`xml.dump({"tag": "1x"})`:                               "invalid or missing tag",
		`url.query_encode({"a": {"b": 1}})`:                     "values must be",
		`yaml.parse(5)`:                                         "want str",
		`template.render("x", {}, left="<<")`:                   "", // right defaults, so this is fine
		`html.select("<p>", "p", limit="x")`:                    "want int",
		`path.clean("/a", flavor="amiga")`:                      "flavor must be",
		`csv.parse("a,b", delimiter="ab")`:                      "single character",
		`xml.query("<a/>", "a[")`:                               "bad path",
		`jwt.sign({"a": 1}, "k", alg="RS256")`:                  "not PEM",
		`yaml.parse("a: 1\na: 2")`:                              "duplicate key",
		`template.render("{{ div 1 0 }}", {})`:                  "division by zero",
		`url.parse("http://a b")`:                               "whitespace",
		`jwt.decode("x.y.z")`:                                   "base64url",
		`xml.dump([1])`:                                         "must be a dict",
		`toml.dump([1])`:                                        "must be a dict",
		`template.render("x", data={"f": lambda: 1})`:           "unsupported value",
		`html.text("<p>", selector="p:nope")`:                   "unsupported pseudo-class",
		`url.query_parse("a=%zz")`:                              "percent",
		`path.with_ext("/a/b", "x/y")`:                          "invalid extension",
		`csv.parse("a", comment='"')`:                           "single character",
		`jwt.sign({"a": 1}, "k", header="x")`:                   "header must be a dict",
		`template.render("{{ repeat 99999999 \"x\" }}", {})`:    "not allowed",
		`yaml.dump({"a": 1}, indent=0)`:                         "indent must be",
		`html.attr("<p>", "p")`:                                 "missing",
		`xml.text("not xml")`:                                   "xml:",
		`url.join("http://a/", "http://[bad")`:                  "ref",
		`jwt.verify(jwt.sign({"a": 1}, "k"), "k", alg="ES256")`: "HS256",
	}
	for src, want := range cases {
		_, err := runStar(t, "x = "+src)
		if want == "" {
			if err != nil {
				t.Errorf("%s: unexpected error %v", src, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%s: expected an error containing %q", src, want)
			continue
		}
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error %q does not contain %q", src, err, want)
		}
	}
}
