package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestColorText(t *testing.T) {
	var c Color
	for in, want := range map[string]Color{"#0f62fe": 0x0f62fe, "ABCDEF": 0xabcdef, "#fa0": 0xffaa00} {
		if err := c.UnmarshalText([]byte(in)); err != nil || c != want {
			t.Errorf("%q: got %v err %v", in, c, err)
		}
	}
	for _, bad := range []string{"", "#12", "#gggggg", "#1234567"} {
		if err := c.UnmarshalText([]byte(bad)); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
	b, _ := Color(0x0f62fe).MarshalText()
	if string(b) != "#0f62fe" {
		t.Errorf("marshal %s", b)
	}
	if Color(0x0f62fe).COLORREF() != 0xfe620f {
		t.Error("COLORREF")
	}
}

func TestBuiltins(t *testing.T) {
	for _, n := range []string{"boxy-light", "boxy-dark", "classic", "nord"} {
		th, ok := Builtin(n)
		if !ok || th.Name != n {
			t.Errorf("missing %s", n)
		}
	}
	l, _ := Builtin("BOXY-LIGHT")
	if l.Radius != 0 || l.BorderWidth != 1 || l.Colors.Accent != 0x0f62fe || l.Colors.Background != 0xffffff {
		t.Errorf("boxy-light wrong: %+v", l)
	}
	d, _ := Builtin("boxy-dark")
	if !d.Dark || d.Colors.Accent != 0x78a9ff || d.Colors.Background != 0x16191c {
		t.Errorf("boxy-dark wrong: %+v", d)
	}
	if len(List()) != 4 {
		t.Error("List")
	}
}

func TestParsePartialUsesDefaults(t *testing.T) {
	th, err := Parse([]byte("dark = true\n[colors]\naccent = \"#ff0000\"\n"), "mine")
	if err != nil {
		t.Fatal(err)
	}
	if th.Name != "mine" || th.Colors.Accent != 0xff0000 {
		t.Errorf("%+v", th)
	}
	if th.Colors.Background != 0x16191c || th.Width != 640 || th.FontSize != 11 {
		t.Errorf("defaults not applied: %+v", th)
	}
	th, err = Parse([]byte("name = \"Sharp\"\nradius = 0\nborder_width = 0\nwidth = -4\n"), "x")
	if err != nil || th.Name != "Sharp" || th.BorderWidth != 0 || th.Width != 640 {
		t.Errorf("%+v %v", th, err)
	}
	if _, err := Parse([]byte("[colors]\nbackground = \"nope\"\n"), "x"); err == nil {
		t.Error("bad color should fail")
	}
}

func TestLoadAllOverrides(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "custom.toml"), []byte("[colors]\ntext = \"#112233\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "over.toml"), []byte("name = \"nord\"\nradius = 2\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "broken.toml"), []byte("= ="), 0o644)
	all, err := LoadAll(dir)
	if err == nil {
		t.Error("expected error for broken file")
	}
	if len(all) != 5 {
		t.Fatalf("got %d themes", len(all))
	}
	byName := map[string]Theme{}
	for _, th := range all {
		byName[th.Name] = th
	}
	if byName["custom"].Colors.Text != 0x112233 || byName["nord"].Radius != 2 {
		t.Errorf("overrides: %+v", byName)
	}
	if got := Resolve("custom", dir, true); got.Name != "custom" {
		t.Error("Resolve custom")
	}
	if got := Resolve("", dir, true); got.Name != DefaultDark {
		t.Error("Resolve default")
	}
	if all, err := LoadAll(filepath.Join(dir, "nope")); err != nil || len(all) != 4 {
		t.Error("missing dir")
	}
}

func TestRoundTrip(t *testing.T) {
	th, _ := Builtin("nord")
	var buf = new(bytesBuf)
	if err := toml.NewEncoder(buf).Encode(th); err != nil {
		t.Fatal(err)
	}
	back, err := Parse(buf.b, "")
	if err != nil || back != th {
		t.Errorf("round trip differs: %v\n%+v\n%+v", err, back, th)
	}
}

type bytesBuf struct{ b []byte }

func (w *bytesBuf) Write(p []byte) (int, error) { w.b = append(w.b, p...); return len(p), nil }
