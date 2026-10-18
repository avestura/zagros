package ext

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/mcp"
	"github.com/avestura/zagros/stubs"
	"go.starlark.net/starlark"
)

// Names an example may use without defining them: the hook arguments, and
// placeholders for values an earlier step produces (a key pair from
// cipher.rsa_generate, bytes read from a file). Add to the list sparingly.
var exampleNames = map[string]bool{
	"ctx": true, "query": true, "scope_id": true, "data": true,
	"blob": true, "pem": true, "pem_text": true, "private_pem": true, "public_pem": true,
	"signature_b64": true, "signature_hex": true, "ciphertext_b64": true,
	"public_key_hex": true, "my_private_hex": true, "peer_public_hex": true,
}

// TestStubExamplesParse parses every `Example:` of every stub (a one-line
// expression, or an indented or fenced block) with the interpreter extensions
// run on, against the predeclared names. A syntax error or a misspelled name
// in an example fails here, so an agent copying one gets working code.
func TestStubExamplesParse(t *testing.T) {
	idx, err := mcp.BuildIndex(stubs.FS())
	if err != nil {
		t.Fatal(err)
	}
	e := &Extension{Manifest: &Manifest{ID: "t.t", Name: "t", Version: "1.0.0"}, opts: &Options{}, kv: newKV("", 0)}
	e.opts.defaults()
	pre := e.predeclared()
	known := func(name string) bool { return pre.Has(name) || exampleNames[name] }

	count := 0
	for _, en := range idx.Entries {
		if en.Example == "" {
			continue
		}
		count++
		if _, _, err := starlark.SourceProgramOptions(fileOpts, en.Qualified+".star", en.Example, known); err != nil {
			t.Errorf("%s (%s): example does not parse: %v\n%s", en.Qualified, en.File, err, en.Example)
			continue
		}
		if (en.Kind == "function") && !strings.Contains(en.Example, en.Name) {
			t.Errorf("%s (%s): the example never uses %s", en.Qualified, en.File, en.Name)
		}
	}
	if count < 100 {
		t.Fatalf("only %d examples found; the Example: convention is broken", count)
	}
}
