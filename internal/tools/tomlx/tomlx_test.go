package tomlx

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "toml") }

func TestParse(t *testing.T) {
	in := `title = "x"
zeta = 1
alpha = 2.5
flag = true
list = [1, 2, 3]
when = 2024-01-02T03:04:05Z
day = 2024-01-02

[owner]
name = "Tom"
  [owner.contact]
  email = "t@example.com"

[[products]]
name = "Hammer"
sku = 7

[[products]]
name = "Nail"
`
	got := toolstest.Compact(toolstest.Must(t, "toml", "parse", "text", in))
	want := `{"title":"x","zeta":1,"alpha":2.5,"flag":true,"list":[1,2,3],"when":"2024-01-02T03:04:05Z","day":"2024-01-02",` +
		`"owner":{"name":"Tom","contact":{"email":"t@example.com"}},"products":[{"name":"Hammer","sku":7},{"name":"Nail"}]}`
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"a = ", "a = 1\na = 2", "[x\n", "a = \"unterminated", "a = [1, \"x\"", "= 1"} {
		_, err := toolstest.Call("toml", "parse", "text", in)
		if err == nil {
			t.Errorf("expected an error for %q", in)
		} else if !strings.HasPrefix(err.Error(), "toml.parse: toml: ") && !strings.HasPrefix(err.Error(), "toml: ") {
			t.Errorf("unexpected error form for %q: %v", in, err)
		}
	}
	if v := toolstest.Must(t, "toml", "is_valid", "text", "a = ["); v != false {
		t.Error("is_valid must be false")
	}
}

func TestParseNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	alphabet := []byte("a=\"'[]{}.,\n #1-:tT\\")
	for i := 0; i < 3000; i++ {
		b := make([]byte, rng.Intn(40))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		toolstest.Call("toml", "parse", "text", string(b))
	}
	toolstest.Call("toml", "parse", "text", strings.Repeat("[", 5000))
	toolstest.Call("toml", "parse", "text", "a = "+strings.Repeat("[", 5000))
}

func TestDumpRoundTrip(t *testing.T) {
	docs := []string{
		`{"title":"x","n":1,"f":1.5,"b":false,"list":[1,2],"mixed":[[1],["a"]],"s":"q\"uote\nline\ttab"}`,
		`{"a":1,"db":{"host":"h","ports":[1,2],"opts":{"x":true}},"servers":[{"name":"a","tags":["x"]},{"name":"b"}]}`,
		`{"weird key":1,"t":{"a.b":{"c":2}},"empty":{}}`,
		`{}`,
	}
	for _, d := range docs {
		v := toolstest.JSONValue(t, d)
		text := toolstest.Must(t, "toml", "dump", "value", v)
		out := toolstest.Compact(toolstest.Must(t, "toml", "parse", "text", text))
		want := toolstest.Compact(v)
		if out != want {
			t.Errorf("round trip changed the value\ntoml:\n%s\n got %s\nwant %s", text, out, want)
		}
	}
}

func TestDumpErrors(t *testing.T) {
	toolstest.Fails(t, "toml", "dump", "value", toolstest.JSONValue(t, `[1,2]`))
	toolstest.Fails(t, "toml", "dump", "value", toolstest.JSONValue(t, `{"a":null}`))
	toolstest.Fails(t, "toml", "dump", "value", toolstest.JSONValue(t, `{"a":[1,null]}`))
}
