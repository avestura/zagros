package xmlx

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "xml") }

const lib = `<?xml version="1.0"?>
<!DOCTYPE library>
<!-- c -->
<library xmlns:x="urn:x">
  <book id="1" lang="en"><title>Go &amp; C</title><price>10</price></book>
  <book id="2"><title><![CDATA[<Raw>]]></title><price>20</price></book>
  <x:magazine id="3"><title>Mag</title></x:magazine>
</library>`

func TestParse(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "xml", "parse", "text", `<a id="1"><b>hi &lt;there&gt;</b><c/></a>`))
	want := `{"tag":"a","attrs":{"id":"1"},"text":"","children":[{"tag":"b","attrs":{},"text":"hi <there>","children":[]},{"tag":"c","attrs":{},"text":"","children":[]}]}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestQuery(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/library/book/title/text()", `["Go&C","<Raw>"]`},
		{"//book[@id='2']/price/text()", `["20"]`},
		{"//book[2]/@id", `["2"]`},
		{"/library/book[last()]/@id", `["2"]`},
		{"//book[@lang]/@id", `["1"]`},
		{"//book[price='10']/@id", `["1"]`},
		{"//book[price!='10']/@id", `["2"]`},
		{"//@id", `["1","2","3"]`},
		{"/library/*/@id", `["1","2","3"]`},
		{"//x:magazine/title/text()", `["Mag"]`},
		{"/nothing", `[]`},
	}
	for _, c := range cases {
		got := toolstest.Compact(toolstest.Must(t, "xml", "query", "source", lib, "path", c.path))
		if strings.ReplaceAll(got, " ", "") != strings.ReplaceAll(c.want, " ", "") {
			t.Errorf("%s: got %s want %s", c.path, got, c.want)
		}
	}
	got := toolstest.Must(t, "xml", "query", "source", lib, "path", "//book[1]").([]any)
	if len(got) != 1 {
		t.Fatalf("got %d nodes", len(got))
	}
	// a parsed tree can be queried again
	tree := toolstest.Must(t, "xml", "parse", "text", lib)
	again := toolstest.Compact(toolstest.Must(t, "xml", "query", "source", tree, "path", "//title/text()"))
	if again != `["Go & C","<Raw>","Mag"]` {
		t.Errorf("tree query got %s", again)
	}
	for _, bad := range []string{"", "//", "a[", "a[0]", "a[b=c]", "@id/x", "a//", "a[@]", "a[()]"} {
		toolstest.Fails(t, "xml", "query", "source", lib, "path", bad)
	}
}

func TestText(t *testing.T) {
	if got := toolstest.Must(t, "xml", "text", "source", "<a>one <b>two</b> <c>three</c></a>"); got != "one two three" {
		t.Errorf("got %q", got)
	}
	if got := toolstest.Must(t, "xml", "text", "source", lib, "path", "//book[2]/title"); got != "<Raw>" {
		t.Errorf("got %q", got)
	}
}

func TestDumpRoundTrip(t *testing.T) {
	tree := toolstest.Must(t, "xml", "parse", "text", lib)
	for _, indent := range []string{"  ", ""} {
		text := toolstest.Must(t, "xml", "dump", "node", tree, "indent", indent, "declaration", true).(string)
		back := toolstest.Must(t, "xml", "parse", "text", text)
		if toolstest.Compact(back) != toolstest.Compact(tree) {
			t.Errorf("indent %q: round trip differs\n%s", indent, text)
		}
	}
	d := toolstest.JSONValue(t, `{"tag":"a","attrs":{"q":"1\"<2>&\n"},"text":"x<y","children":[{"tag":"b"}]}`)
	got := toolstest.Must(t, "xml", "dump", "node", d, "indent", "")
	if want := "<a q=\"1&quot;&lt;2&gt;&amp;&#10;\">x&lt;y<b/></a>"; got != want {
		// text precedes children on one line; accept the exact form below
		t.Errorf("got %q want %q", got, want)
	}
	for _, bad := range []string{`"x"`, `{}`, `{"tag":"1bad"}`, `{"tag":"a","attrs":{"b c":"1"}}`, `{"tag":"a","children":"x"}`, `{"tag":"a","attrs":{"k":[1]}}`} {
		toolstest.Fails(t, "xml", "dump", "node", toolstest.JSONValue(t, bad))
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "text", "<a>", "<a></b>", "<a/><b/>", "<a>&nope;</a>", "<a b=1/>", "</a>", "<a><b></a></b>", "<a>x</a>trailing"} {
		if _, err := toolstest.Call("xml", "parse", "text", in); err == nil {
			t.Errorf("expected an error for %q", in)
		}
	}
	toolstest.Fails(t, "xml", "parse", "text", strings.Repeat("<a>", 300)+strings.Repeat("</a>", 300))
	if v := toolstest.Must(t, "xml", "is_valid", "text", "<a><b/></a>"); v != true {
		t.Error("valid XML reported invalid")
	}
}

func TestParseNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	alphabet := []byte("<>/=\"'&;!-?[]a b\n:#x0")
	for i := 0; i < 4000; i++ {
		b := make([]byte, rng.Intn(40))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		s := string(b)
		toolstest.Call("xml", "parse", "text", s)
		toolstest.Call("xml", "query", "source", "<a b='1'><c/></a>", "path", s)
		toolstest.Call("xml", "is_valid", "text", s)
	}
}
