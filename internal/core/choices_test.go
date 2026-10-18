package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestChoices(t *testing.T) (*Choices, *time.Time) {
	t.Helper()
	now := time.Now()
	c := NewChoices(filepath.Join(t.TempDir(), "choices.json"))
	c.now = func() time.Time { return now }
	return c, &now
}

func TestChoicesPickedItemLeadsItsQuery(t *testing.T) {
	cat := NewCatalog(nil)
	ch := NewChoices("")
	cat.SetChoices(ch)
	cat.Register(&fakeProvider{name: "apps", items: []Item{
		{ID: "apps:fire alarm", Title: "Fire Alarm"},
		{ID: "apps:firefox", Title: "Firefox"},
		{ID: "apps:file explorer", Title: "File Explorer"},
	}})
	top := func(q string) string { return cat.Query(context.Background(), q)[0].ID }
	ch.RecordQuery("fir", "apps:file explorer")
	if top("fir") != "apps:file explorer" {
		t.Fatalf("the picked item must lead the query it was picked for, got %s", top("fir"))
	}
	if top("fire") != "apps:file explorer" {
		t.Fatalf("a query that extends the learned one keeps the pick, got %s", top("fire"))
	}
	if b := ch.QueryBonuses("fi"); len(b) != 0 {
		t.Fatalf("a shorter query must not inherit the pick: %v", b)
	}
	ch.RecordQuery("fir", "apps:firefox")
	ch.RecordQuery("fir", "apps:firefox")
	if top("fir") != "apps:firefox" {
		t.Fatalf("the more often picked item leads, got %s", top("fir"))
	}
}

func TestChoicesAreBoundedAndAgeOut(t *testing.T) {
	c, now := newTestChoices(t)
	for i := 0; i < maxChoiceContexts+50; i++ {
		*now = now.Add(time.Minute)
		c.RecordQuery(fmt.Sprintf("q%d", i), "apps:x")
	}
	if q, _ := c.Len(); q != maxChoiceContexts {
		t.Fatalf("contexts must be capped at %d, have %d", maxChoiceContexts, q)
	}
	if c.HasQuery("q0") || !c.HasQuery(fmt.Sprintf("q%d", maxChoiceContexts+49)) {
		t.Fatal("the least recently used queries go first")
	}
	for i := 0; i < 20; i++ {
		c.RecordQuery("one", fmt.Sprintf("apps:item%d", i))
		*now = now.Add(time.Second)
	}
	if n := len(c.queries["one"]); n != maxChoicesPerContext {
		t.Fatalf("items per query must be capped at %d, have %d", maxChoicesPerContext, n)
	}
	before, _ := c.Len()
	c.RecordQuery(string(make([]byte, 200)), "apps:x")
	c.RecordQuery("   ", "apps:x")
	if q, _ := c.Len(); q != before {
		t.Fatalf("empty and overlong queries are not stored: %d -> %d", before, q)
	}
	// untouched picks decay below the floor and are deleted from memory and disk
	*now = now.Add(400 * 24 * time.Hour)
	if err := c.Flush(); err != nil {
		t.Fatal(err)
	}
	if q, s := c.Len(); q != 0 || s != 0 {
		t.Fatalf("aged-out picks must be deleted, have %d queries %d scopes", q, s)
	}
	if _, err := os.Stat(c.path); !os.IsNotExist(err) {
		t.Fatalf("an empty store must not leave a file behind: %v", err)
	}
}

func TestChoicesUnlearnableAndForget(t *testing.T) {
	c, _ := newTestChoices(t)
	c.RecordQuery("cl", "clipboard:0")
	c.RecordQuery("calc", "calc:result")
	c.RecordQuery("", "apps:x")
	c.RecordScope("", "apps:x")
	if q, s := c.Len(); q != 0 || s != 0 {
		t.Fatalf("unstable IDs and empty contexts are not remembered: %d %d", q, s)
	}
	c.RecordQuery("fir", "apps:firefox")
	c.RecordQuery("ff", "apps:firefox")
	c.RecordScope("tools.text", "tools:text.upper")
	c.RecordScope("tools.hash", "apps:firefox")
	c.Forget("apps:firefox")
	if q, s := c.Len(); q != 0 || s != 1 {
		t.Fatalf("Forget removes the item everywhere and empty contexts with it: %d %d", q, s)
	}
}

func TestChoicesValidateDeletesVanishedItemsAfterStrikes(t *testing.T) {
	c, _ := newTestChoices(t)
	c.RecordQuery("fir", "apps:gone")
	c.RecordQuery("fir", "apps:here")
	c.RecordQuery("fir", "ext:other")
	known := func(id string) bool { return providerOf(id) == "apps" }
	found := func(id string) bool { return id == "apps:here" }
	for i := 1; i < choiceStrikes; i++ {
		c.Validate("fir", known, found)
		if len(c.queries["fir"]) != 3 {
			t.Fatalf("strike %d must not delete yet", i)
		}
	}
	c.Validate("fir", known, found)
	if _, ok := c.queries["fir"]["apps:gone"]; ok {
		t.Fatal("an item missing choiceStrikes times in a row must be deleted")
	}
	if _, ok := c.queries["fir"]["apps:here"]; !ok {
		t.Fatal("a present item is kept")
	}
	if _, ok := c.queries["fir"]["ext:other"]; !ok {
		t.Fatal("items of providers that did not answer are never struck")
	}
	// reappearing resets the count
	c.RecordQuery("x", "apps:flaky")
	c.Validate("x", known, func(string) bool { return false })
	c.Validate("x", known, func(string) bool { return true })
	c.Validate("x", known, func(string) bool { return false })
	c.Validate("x", known, func(string) bool { return false })
	if _, ok := c.queries["x"]["apps:flaky"]; !ok {
		t.Fatal("misses must be consecutive")
	}
}

func TestCatalogDropsPicksOfUninstalledApps(t *testing.T) {
	cat := NewCatalog(nil)
	ch := NewChoices("")
	cat.SetChoices(ch)
	p := &fakeProvider{name: "apps", items: []Item{{ID: "apps:firefox", Title: "Firefox"}}}
	cat.Register(p)
	ch.RecordQuery("fir", "apps:firefox")
	p.items = nil // uninstalled
	for i := 0; i < choiceStrikes; i++ {
		cat.Query(context.Background(), "fir")
	}
	if ch.HasQuery("fir") {
		t.Fatal("the pick of an app that no longer shows up must be deleted")
	}
	// a provider that is off must not cost the user their picks
	ch.RecordQuery("fir", "apps:firefox")
	cat.SetEnabled("apps", false)
	for i := 0; i < choiceStrikes+1; i++ {
		cat.Query(context.Background(), "fir")
	}
	if !ch.HasQuery("fir") {
		t.Fatal("a disabled provider must not cost the user their picks")
	}
}

func TestChoicesScopeOrderAndPersistence(t *testing.T) {
	cat := NewCatalog(nil)
	ch, now := newTestChoices(t)
	cat.SetChoices(ch)
	cat.Register(&toolsScopeProv{fakeProvider{name: "tools"}})
	sc := Scope{ID: "tools.text", Provider: "tools"}
	ids := func(s Scope) []string {
		var out []string
		for _, it := range cat.QueryScope(context.Background(), s, "hello") {
			out = append(out, it.ID)
		}
		return out
	}
	if got := ids(sc); got[0] != "tools:lower" {
		t.Fatalf("unlearned order is the provider's: %v", got)
	}
	ch.RecordScope("tools.text", "tools:upper")
	if got := ids(sc); got[0] != "tools:upper" || got[1] != "tools:lower" {
		t.Fatalf("the tool picked in this scope leads it, the rest keep their order: %v", got)
	}
	if got := ids(Scope{ID: "tools.hash", Provider: "tools"}); got[0] != "tools:lower" {
		t.Fatalf("picks are per scope: %v", got)
	}
	*now = now.Add(time.Hour)
	if err := ch.Flush(); err != nil {
		t.Fatal(err)
	}
	re := NewChoices(ch.path)
	if b := re.ScopeBonuses("tools.text"); b["tools:upper"] <= 0 {
		t.Fatalf("picks must survive a restart: %v", b)
	}
	data, _ := os.ReadFile(ch.path)
	if len(data) > 400 {
		t.Fatalf("one pick should be a tiny file, got %d bytes", len(data))
	}
}

type toolsScopeProv struct{ fakeProvider }

func (toolsScopeProv) QueryScope(_ context.Context, _ Scope, _ string) []Item {
	return []Item{{ID: "tools:lower"}, {ID: "tools:upper"}, {ID: "tools:slug"}}
}

type fbProv struct {
	fakeProvider
	fb []Item
}

func (f *fbProv) Fallback(string) []Item { return f.fb }

func TestWithFallbacksCapsAndInterleavesProviders(t *testing.T) {
	cat := NewCatalog(nil)
	mk := func(name string, n int) *fbProv {
		p := &fbProv{fakeProvider: fakeProvider{name: name}}
		for i := 0; i < n; i++ {
			p.fb = append(p.fb, Item{ID: fmt.Sprintf("%s:%d", name, i), Title: "fb"})
		}
		return p
	}
	tools, web := mk("tools", 6), mk("web", 1)
	var out []Item
	for i := 0; i < 8; i++ {
		out = append(out, Item{ID: fmt.Sprintf("r:%d", i)})
	}
	got := cat.withFallbacks(out, []Provider{tools, web}, "q", false, 8)
	if len(got) != 8 {
		t.Fatalf("the list never grows past the limit: %d", len(got))
	}
	var fbIDs []string
	for _, it := range got[4:] {
		fbIDs = append(fbIDs, it.ID)
	}
	want := []string{"tools:0", "tools:1", "tools:2", "web:0"}
	for i := range want {
		if fbIDs[i] != want[i] {
			t.Fatalf("every provider gets its first row, then the cap applies: %v", fbIDs)
		}
	}
}
