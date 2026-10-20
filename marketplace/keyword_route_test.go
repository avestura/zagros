package marketplace

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
)

// The catalog strips the keyword before calling the provider; "define word"
// typed in the launcher must still reach the extension.
func TestKeywordQueryThroughCatalog(t *testing.T) {
	web := rt{"entries/en/something": `[{"word":"something","meanings":[{"partOfSpeech":"noun","definitions":[{"definition":"A thing."}]}]}]`}
	e, err := ext.OpenExtension(filepath.Join("extensions", "dictionary"), ext.Options{DataDir: t.TempDir(), HTTPClient: &http.Client{Transport: web}})
	if err != nil {
		t.Fatal(err)
	}
	e.SetGranted(e.Manifest.Permissions)
	if err := e.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	c := core.NewCatalog(nil)
	c.Register(ext.NewProvider(e))
	last := func(q string) core.Snapshot {
		var s core.Snapshot
		c.QueryStream(context.Background(), q, func(x core.Snapshot) { s = x })
		return s
	}
	for _, q := range []string{"define something", "dict something", "define   something"} {
		got := false
		for i := 0; i < 400 && !got; i++ { // a finished task re-runs the query in the launcher
			for _, it := range last(q).Items {
				got = got || strings.Contains(it.Title, "A thing.")
			}
		}
		if !got {
			t.Errorf("%q gave no definition: %+v", q, last(q).Items)
		}
	}
	if items := last("define").Items; len(items) != 1 || !strings.Contains(items[0].Subtitle, "Tab") {
		t.Errorf("bare keyword must show the hint row: %+v", items)
	}
	if items := last("something").Items; len(items) != 0 {
		t.Errorf("without the keyword the extension must stay silent: %+v", items)
	}
}
