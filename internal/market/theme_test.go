package market

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const goodTheme = "name = \"mini\"\ndark = true\n[colors]\nbackground = \"#101010\"\n"

func themeServer(t *testing.T, themeBody string) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/registry.json":
			fmt.Fprint(w, `{"schema":1,"generated":"x","extensions":[
			 {"id":"mini","name":"Mini","repo":"o/r","kind":"theme","subdir":"themes/mini.toml","latest":"1.0.0","permissions":[],"verified":true},
			 {"id":"ext","name":"Ext","repo":"o/r","latest":"1.0.0","permissions":[],"verified":false}]}`)
		case strings.HasSuffix(r.URL.Path, "/contents/themes/mini.toml"):
			fmt.Fprint(w, themeBody)
		case strings.HasSuffix(r.URL.Path, "/t.toml"):
			fmt.Fprint(w, goodTheme)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestInstallTheme(t *testing.T) {
	srv := themeServer(t, goodTheme)
	c := newTestClient(t, srv)
	dir := filepath.Join(t.TempDir(), "themes")
	pkg, err := c.InstallTheme(context.Background(), dir, "mini")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Kind != KindTheme || pkg.ID != "mini" {
		t.Fatalf("pkg = %+v", pkg)
	}
	b, err := os.ReadFile(filepath.Join(dir, "mini.toml"))
	if err != nil || string(b) != goodTheme {
		t.Fatalf("installed file: %q %v", b, err)
	}
	// Direct URL to a .toml file.
	if p, err := c.InstallTheme(context.Background(), dir, srv.URL+"/x/t.toml"); err != nil || p.ID != "t" {
		t.Fatalf("url install: %v %+v", err, p)
	}
	if _, err := os.Stat(filepath.Join(dir, "t.toml")); err != nil {
		t.Fatal(err)
	}
}

func TestInstallThemeRejects(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"invalid": "this is = not [valid",
		"huge":    "name = \"x\"\n#" + strings.Repeat("a", MaxThemeSize),
		"badhex":  "[colors]\nbackground = \"#zzzzzz\"\n",
	} {
		srv := themeServer(t, body)
		c := newTestClient(t, srv)
		if _, err := c.InstallTheme(context.Background(), dir, "mini"); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("rejected themes must not be written, found %v", entries)
	}
	c := newTestClient(t, themeServer(t, goodTheme))
	if _, err := c.InstallTheme(context.Background(), dir, "ext"); err == nil {
		t.Error("an extension id must not install as a theme")
	}
}

func TestThemeSpec(t *testing.T) {
	s, err := ParseSpec([]byte("kind = \"theme\"\nrepo = \"a/b\"\nsubdir = \"themes/x.toml\"\n"), "x")
	if err != nil || s.Kind != KindTheme {
		t.Fatalf("%v %+v", err, s)
	}
	if _, err := ParseSpec([]byte("kind = \"theme\"\nrepo = \"a/b\"\n"), "x"); err == nil {
		t.Error("theme without a file path must fail")
	}
	if _, err := ParseSpec([]byte("kind = \"plugin\"\nrepo = \"a/b\"\n"), "x"); err == nil {
		t.Error("unknown kind must fail")
	}
	e, err := EntryFromSpec(s)
	if err != nil || e.Kind != KindTheme {
		t.Fatalf("%v %+v", err, e)
	}
}
