package urlx

import (
	"math/rand"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "url") }

func TestParse(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "url", "parse", "url", "https://u:p%40@Example.com:8080/a/b?x=1&x=2&y=a%20b#top"))
	want := `{"scheme":"https","user":"u","password":"p@","host":"Example.com:8080","hostname":"Example.com","port":"8080","path":"/a/b",` +
		`"query":{"x":["1","2"],"y":["a b"]},"query_string":"x=1&x=2&y=a%20b","fragment":"top","origin":"https://example.com:8080"}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	got = toolstest.Compact(toolstest.Must(t, "url", "parse", "url", "example.com/a?q=1", "default_scheme", "https"))
	if want := `"scheme":"https","user":"","password":"","host":"example.com"`; !contains(got, want) {
		t.Errorf("default scheme: %s", got)
	}
	got = toolstest.Compact(toolstest.Must(t, "url", "parse", "url", "localhost:3000/x", "default_scheme", "http"))
	if !contains(got, `"host":"localhost:3000"`) {
		t.Errorf("host:port: %s", got)
	}
	for _, bad := range []string{"", "  ", "http://a b/", "http://[::1", "%zz", "http://x/\x00", "http://a/%"} {
		toolstest.Fails(t, "url", "parse", "url", bad)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestBuildRoundTrip(t *testing.T) {
	parts := toolstest.JSONValue(t, `{"scheme":"https","user":"bob","password":"p w","hostname":"example.com","port":8443,"path":"search/here","query":{"q":"go lang","tag":["a","b"],"n":2,"skip":null},"fragment":"f"}`)
	got := toolstest.Must(t, "url", "build", "parts", parts)
	if want := "https://bob:p%20w@example.com:8443/search/here?q=go+lang&tag=a&tag=b&n=2#f"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if got := toolstest.Must(t, "url", "build", "parts", toolstest.JSONValue(t, `{"path":"/x","query":"?a=1"}`)); got != "/x?a=1" {
		t.Errorf("got %q", got)
	}
	for _, bad := range []string{`[1]`, `{"port":"x1"}`, `{"query":5}`, `{"query":{"a":{"b":1}}}`, `{"host":["x"]}`} {
		toolstest.Fails(t, "url", "build", "parts", toolstest.JSONValue(t, bad))
	}
}

func TestJoin(t *testing.T) {
	for _, c := range [][3]string{
		{"https://e.com/a/b/c", "../d?x=1", "https://e.com/a/d?x=1"},
		{"https://e.com/a/b", "/root", "https://e.com/root"},
		{"https://e.com/a/b", "?q=1", "https://e.com/a/b?q=1"},
		{"https://e.com/a/b", "//o.org/x", "https://o.org/x"},
		{"https://e.com/a/b", "http://z.net/", "http://z.net/"},
		{"https://e.com/a/b/", "./c", "https://e.com/a/b/c"},
	} {
		if got := toolstest.Must(t, "url", "join", "base", c[0], "ref", c[1]); got != c[2] {
			t.Errorf("join(%s, %s) = %v, want %s", c[0], c[1], got, c[2])
		}
	}
	toolstest.Fails(t, "url", "join", "base", "", "ref", "x")
	toolstest.Fails(t, "url", "join", "base", "http://a/", "ref", "http://[bad")
}

func TestQuery(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "url", "query_parse", "query", "?a=1&b=x%20y+z&a=2&flag&&c="))
	if want := `{"a":"2","b":"x y z","flag":"","c":""}`; got != want {
		t.Errorf("got %s", got)
	}
	got = toolstest.Compact(toolstest.Must(t, "url", "query_parse", "query", "a=1&a=2&b=3", "multi", true))
	if got != `{"a":["1","2"],"b":["3"]}` {
		t.Errorf("got %s", got)
	}
	toolstest.Fails(t, "url", "query_parse", "query", "a=%zz")
	toolstest.Fails(t, "url", "query_parse", "query", "%g1=2")
	d := toolstest.JSONValue(t, `{"z":"a b&c","a":[1,true,null,"x"],"e":""}`)
	if got := toolstest.Must(t, "url", "query_encode", "params", d); got != "z=a+b%26c&a=1&a=true&a=x&e=" {
		t.Errorf("got %q", got)
	}
	if got := toolstest.Must(t, "url", "query_encode", "params", d, "sort", true, "plus", false); got != "a=1&a=true&a=x&e=&z=a%20b%26c" {
		t.Errorf("got %q", got)
	}
	toolstest.Fails(t, "url", "query_encode", "params", "x")
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		"HTTP://Example.COM:80/a/./b/../c?b=2&a=1#F": "http://example.com/a/c?a=1&b=2#F",
		"https://example.com":                        "https://example.com/",
		"https://example.com:443/%7euser/%2f/%e4":    "https://example.com/~user/%2F/%E4",
		"https://example.com:8443/a//b/":             "https://example.com:8443/a/b/",
		"http://[::1]:80/x":                          "http://[::1]/x",
		"https://example.com/a?z=1&a=2&z=0":          "https://example.com/a?a=2&z=1&z=0",
		"mailto:Someone@Example.com":                 "mailto:Someone@Example.com",
		"https://example.com/../../a":                "https://example.com/a",
	} {
		if got := toolstest.Must(t, "url", "normalize", "url", in); got != want {
			t.Errorf("%s: got %v want %s", in, got, want)
		}
	}
	if got := toolstest.Must(t, "url", "normalize", "url", "https://e.com/a/?x=2&x=1#f", "sort_query", false, "strip_fragment", true, "strip_trailing_slash", true); got != "https://e.com/a?x=2&x=1" {
		t.Errorf("got %v", got)
	}
}

func TestNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	alphabet := []byte("htp:/@?#&=%[]a1.-~ \x00\xff")
	for i := 0; i < 5000; i++ {
		b := make([]byte, rng.Intn(40))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		s := string(b)
		toolstest.Call("url", "parse", "url", s)
		toolstest.Call("url", "parse", "url", s, "default_scheme", "https")
		toolstest.Call("url", "normalize", "url", s)
		toolstest.Call("url", "join", "base", "http://a/b/c", "ref", s)
		toolstest.Call("url", "query_parse", "query", s)
	}
}
