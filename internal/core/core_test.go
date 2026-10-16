package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeProvider struct {
	name   string
	ranked bool
	items  []Item
	delay  time.Duration
	kws    []string
	got    string
}

func (f *fakeProvider) Name() string       { return f.name }
func (f *fakeProvider) Ranked() bool       { return f.ranked }
func (f *fakeProvider) Keywords() []string { return f.kws }
func (f *fakeProvider) Query(ctx context.Context, q string) []Item {
	f.got = q
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil
		}
	}
	return f.items
}

func TestCatalogRankDedupeLimit(t *testing.T) {
	c := NewCatalog(nil)
	c.SetLimit(3)
	c.Register(&fakeProvider{name: "a", items: []Item{
		{ID: "a:1", Title: "Visual Studio Code"},
		{ID: "a:2", Title: "Notepad"},
		{ID: "dup", Title: "Vim"},
	}})
	c.Register(&fakeProvider{name: "b", items: []Item{
		{ID: "dup", Title: "Vim"},
		{ID: "b:1", Title: "Vivaldi"},
	}})
	got := c.Query(context.Background(), "vi")
	if len(got) != 3 {
		t.Fatalf("len=%d want 3: %+v", len(got), got)
	}
	seen := map[string]bool{}
	for _, it := range got {
		if seen[it.ID] {
			t.Errorf("duplicate %s", it.ID)
		}
		seen[it.ID] = true
		if it.Score <= 0 {
			t.Errorf("score not set for %s", it.ID)
		}
		if it.Title == "Notepad" {
			t.Error("non-matching item returned")
		}
	}
}

func TestCatalogDisabledAndTimeout(t *testing.T) {
	c := NewCatalog(nil)
	c.SetTimeout(50 * time.Millisecond)
	c.Register(&fakeProvider{name: "fast", items: []Item{{ID: "f", Title: "Fast"}}})
	c.Register(&fakeProvider{name: "slow", delay: 2 * time.Second, items: []Item{{ID: "s", Title: "Slow"}}})
	c.Register(&fakeProvider{name: "off", items: []Item{{ID: "o", Title: "Off"}}})
	cfg := DefaultConfig()
	cfg.Providers["off"] = false
	c.ApplyConfig(cfg)

	start := time.Now()
	got := c.Query(context.Background(), "")
	if time.Since(start) > time.Second {
		t.Error("query did not honour the deadline")
	}
	if len(got) != 1 || got[0].ID != "f" {
		t.Errorf("got %+v", got)
	}
}

func TestCatalogKeywordRouting(t *testing.T) {
	c := NewCatalog(nil)
	web := &fakeProvider{name: "web", ranked: true, kws: []string{"g"}, items: []Item{{ID: "w", Title: "Search Google"}}}
	apps := &fakeProvider{name: "apps", items: []Item{{ID: "a", Title: "g golang app"}}}
	c.Register(apps)
	c.Register(web)
	got := c.Query(context.Background(), "g golang")
	if web.got != "golang" {
		t.Errorf("routed provider got %q", web.got)
	}
	if apps.got != "g golang" {
		t.Errorf("other provider got %q", apps.got)
	}
	if len(got) == 0 || got[0].ID != "w" {
		t.Errorf("routed item should be first: %+v", got)
	}
}

func TestCatalogRankedOrderKept(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProvider{name: "r", ranked: true, items: []Item{{ID: "1", Title: "zzz"}, {ID: "2", Title: "aaa"}}})
	got := c.Query(context.Background(), "a")
	if len(got) != 2 || got[0].ID != "1" {
		t.Errorf("ranked order not kept: %+v", got)
	}
}

func TestHistoryBoostAndPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	h := NewHistory(path)
	if h.Boost("x") != 1 {
		t.Error("unknown id should have boost 1")
	}
	for i := 0; i < 5; i++ {
		h.Record("x", "xx")
	}
	b := h.Boost("x")
	if b <= 1 || b > maxBoost {
		t.Errorf("boost %v out of range", b)
	}
	if err := h.Flush(); err != nil {
		t.Fatal(err)
	}
	h2 := NewHistory(path)
	if h2.Uses("x") != 5 || h2.Boost("x") <= 1 {
		t.Error("history not persisted")
	}
	// Decay: two half-lives later the boost shrinks.
	h2.now = func() time.Time { return time.Now().Add(2 * frecencyHalfLife) }
	if h2.Boost("x") >= b {
		t.Error("boost should decay")
	}
}

func TestHistoryBoostChangesRanking(t *testing.T) {
	h := NewHistory("")
	c := NewCatalog(h)
	c.Register(&fakeProvider{name: "a", items: []Item{{ID: "1", Title: "Firefox"}, {ID: "2", Title: "Fire Alarm"}}})
	h.Record("2", "fi")
	h.Record("2", "fi")
	h.Record("2", "fi")
	got := c.Query(context.Background(), "fi")
	if got[0].ID != "2" {
		t.Errorf("frequent item should lead: %+v", got)
	}
}

func TestConfigLoadDefaultsAndRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "zagros.toml")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("default config not written")
	}
	if !cfg.Everything.Enabled || cfg.Everything.Limit != 20 || len(cfg.WebSearch) != 5 {
		t.Errorf("bad defaults: %+v", cfg)
	}
	if !cfg.ProviderEnabled("anything") {
		t.Error("missing provider should be enabled")
	}
	// The written file parses back to the defaults.
	cfg2, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg2.WebSearch) != 5 || cfg2.Marketplace.RegistryURL != DefaultRegistryURL || cfg2.MaxResults != 8 {
		t.Errorf("reloaded defaults differ: %+v", cfg2)
	}

	os.WriteFile(path, []byte("max_results = 5\n[providers]\ncalc = false\n[[web_search]]\nkeyword=\"x\"\nname=\"X\"\nurl=\"u{q}\"\n"), 0o644)
	cfg3, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg3.MaxResults != 5 || cfg3.ProviderEnabled("calc") || len(cfg3.WebSearch) != 1 || !cfg3.Everything.Enabled {
		t.Errorf("partial config: %+v", cfg3)
	}

	if err := SaveConfig(path, cfg3); err != nil {
		t.Fatal(err)
	}
	cfg4, err := LoadConfig(path)
	if err != nil || cfg4.MaxResults != 5 || cfg4.ProviderEnabled("calc") {
		t.Errorf("round trip: %+v %v", cfg4, err)
	}
}

func TestPathsOverride(t *testing.T) {
	t.Setenv("ZAGROS_HOME", `C:\z`)
	if ConfigDir() != filepath.Join(`C:\z`, "config") || DataDir() != filepath.Join(`C:\z`, "data") {
		t.Error("ZAGROS_HOME not honoured")
	}
	if filepath.Base(ExtDir()) != "extensions" || filepath.Base(ThemeDir()) != "themes" {
		t.Error("sub dirs wrong")
	}
}

func TestSetThemeLine(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"replace", "# c\ntheme = \"\"  \nmax_results = 8\n", "# c\ntheme = \"nord\"\nmax_results = 8\n"},
		{"replace crlf", "a = 1\r\ntheme=\"x\"\r\n", "a = 1\r\ntheme = \"nord\"\r\n"},
		{"append", "a = 1\n", "a = 1\ntheme = \"nord\"\n"},
		{"append no newline", "a = 1", "a = 1\ntheme = \"nord\"\n"},
		{"empty", "", "theme = \"nord\"\n"},
		{"before table", "a = 1\n[x]\ntheme = \"keep\"\n", "a = 1\ntheme = \"nord\"\n\n[x]\ntheme = \"keep\"\n"},
		{"similar key", "theme_x = 1\n", "theme_x = 1\ntheme = \"nord\"\n"},
	}
	for _, c := range cases {
		if got := SetThemeLine(c.in, "nord"); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestSetTheme(t *testing.T) {
	p := filepath.Join(t.TempDir(), "zagros.toml")
	if err := os.WriteFile(p, []byte("# keep me\nhotkey = \"Alt+Space\"\ntheme = \"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetTheme(p, "my theme"); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "# keep me") {
		t.Fatalf("comment lost: %s", b)
	}
	cfg, err := LoadConfig(p)
	if err != nil || cfg.Theme != "my theme" {
		t.Fatalf("theme %q err %v", cfg.Theme, err)
	}
}
