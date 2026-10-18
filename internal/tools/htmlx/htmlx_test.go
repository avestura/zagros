package htmlx

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "html") }

const page = `<!DOCTYPE html>
<html><head><title>T</title><style>p{}</style></head>
<body>
<div id="main" class="box wide">
  <h1>Head <em>line</em></h1>
  <p class="x">one</p>
  <p>two <a href="https://a.test/x" class="ext">link</a></p>
  <p class="x last">three</p>
  <ul><li>a</li><li class="on">b</li><li>c</li><li>d</li></ul>
  <script>var s = 1;</script>
  <input type="checkbox" checked><input type="text" disabled>
</div>
<a href="/rel" data-k="Pre-Fix">rel</a>
</body></html>`

func texts(t *testing.T, sel string) string {
	t.Helper()
	nodes := toolstest.Must(t, "html", "select", "text", page, "selector", sel).([]any)
	var parts []string
	for _, n := range nodes {
		d := n.(interface{ Get(string) (any, bool) })
		v, _ := d.Get("text")
		parts = append(parts, v.(string))
	}
	return strings.Join(parts, "|")
}

func TestSelect(t *testing.T) {
	cases := []struct{ sel, want string }{
		{"p", "one|two|three"},
		{"p.x", "one|three"},
		{"p.x.last", "three"},
		{"#main > p", "one|two|three"},
		{"div p a", "link"},
		{"a[href^='https']", "link"},
		{"a[href$=x]", "link"},
		{"a[href*='.test']", "link"},
		{"a[data-k='pre-fix' i]", "rel"},
		{"a[href]", "link|rel"},
		{"li:first-child", "a"},
		{"li:last-child", "d"},
		{"li:nth-child(2)", "b"},
		{"li:nth-child(odd)", "a|c"},
		{"li:nth-child(2n+2)", "b|d"},
		{"li:nth-last-child(1)", "d"},
		{"li.on + li", "c"},
		{"li.on ~ li", "c|d"},
		{"li:not(.on)", "a|c|d"},
		{"p:first-of-type", "one"},
		{"h1, li.on", "Head|b"},
		{"p:is(.last, .ext)", "three"},
		{"title", "T"},
		{"input:checked", ""},
		{"input:disabled", ""},
		{"input[type=text]", ""},
		{"nothing", ""},
	}
	for _, c := range cases {
		if got := texts(t, c.sel); got != c.want {
			t.Errorf("%s: got %q want %q", c.sel, got, c.want)
		}
	}
	nodes := toolstest.Must(t, "html", "select", "text", page, "selector", "li", "limit", int64(2)).([]any)
	if len(nodes) != 2 {
		t.Errorf("limit gave %d", len(nodes))
	}
	if got := toolstest.Must(t, "html", "select", "text", page, "selector", "input:checked").([]any); len(got) != 1 {
		t.Errorf(":checked matched %d", len(got))
	}
}

func TestSelectorErrors(t *testing.T) {
	for _, s := range []string{"", " ", "a[", "a[=x]", ">", "a >", ".", "#", "a:unknown", "a::before", "a:not(", "a:nth-child(x)", "a b c d e f g h i", "a[b=\"x]", "p,", strings.Repeat("a", 3000)} {
		toolstest.Fails(t, "html", "select", "text", "<p>", "selector", s)
	}
}

func TestText(t *testing.T) {
	got := toolstest.Must(t, "html", "text", "text", page)
	if want := "Head line\none\ntwo link\nthree\na\nb\nc\nd\nrel"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if got := toolstest.Must(t, "html", "text", "text", page, "selector", "p.x"); got != "one\nthree" {
		t.Errorf("got %q", got)
	}
	if got := toolstest.Must(t, "html", "text", "text", "a<br>b&amp;c &lt;d&gt;"); got != "a\nb&c <d>" {
		t.Errorf("got %q", got)
	}
}

func TestAttrAndOuter(t *testing.T) {
	got := toolstest.Must(t, "html", "attr", "text", page, "selector", "a", "name", "href").([]string)
	if len(got) != 2 || got[0] != "https://a.test/x" || got[1] != "/rel" {
		t.Errorf("got %v", got)
	}
	out := toolstest.Must(t, "html", "outer_html", "text", page, "selector", "li.on").([]string)
	if len(out) != 1 || out[0] != "<li class=\"on\">b</li>" {
		t.Errorf("got %q", out)
	}
}

func TestParse(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "html", "parse", "text", "<ul><li>a<li>b</ul>", "fragment", true))
	want := `{"tag":"#fragment","attrs":{},"text":"","children":[{"tag":"ul","attrs":{},"text":"","children":[{"tag":"li","attrs":{},"text":"a","children":[]},{"tag":"li","attrs":{},"text":"b","children":[]}]}]}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	whole := toolstest.Compact(toolstest.Must(t, "html", "parse", "text", "<p>x"))
	if !strings.HasPrefix(whole, `{"tag":"html"`) || !strings.Contains(whole, `"tag":"body"`) {
		t.Errorf("got %s", whole)
	}
}

func TestNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(4))
	alphabet := []byte("<>/=\"'&;!-a b\n[]#.:()*+~,0")
	for i := 0; i < 3000; i++ {
		b := make([]byte, rng.Intn(50))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		s := string(b)
		toolstest.Call("html", "parse", "text", s)
		toolstest.Call("html", "parse", "text", s, "fragment", true)
		toolstest.Call("html", "select", "text", page, "selector", s)
		toolstest.Call("html", "select", "text", s, "selector", "*")
		toolstest.Call("html", "text", "text", s)
	}
	toolstest.Call("html", "parse", "text", strings.Repeat("<div>", 20000))
	toolstest.Call("html", "text", "text", strings.Repeat("<div>", 20000))
}
