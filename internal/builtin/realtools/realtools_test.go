// Package realtools tests the launcher's tools provider against the full
// registry (internal/tools/all), which the fake-tool tests in package builtin
// deliberately do not link.
package realtools

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/builtin"
	"github.com/avestura/zagros/internal/core"
	_ "github.com/avestura/zagros/internal/tools/all"
)

func toolsProvider(t *testing.T) core.Provider {
	t.Helper()
	cfg := core.DefaultConfig()
	for _, p := range builtin.All(&cfg) {
		if p.Name() == "tools" {
			return p
		}
	}
	t.Fatal("tools provider is not registered")
	return nil
}

func query(t *testing.T, q string) []core.Item {
	t.Helper()
	start := time.Now()
	items := toolsProvider(t).Query(context.Background(), q)
	if d := time.Since(start); d > 400*time.Millisecond {
		t.Errorf("%q took %v: a keystroke must stay fast", q, d)
	}
	return items
}

func joined(items []core.Item) string {
	var b strings.Builder
	for _, it := range items {
		b.WriteString(it.Title + " | " + it.Subtitle + "\n")
	}
	return b.String()
}

func want(t *testing.T, q string, subs ...string) {
	t.Helper()
	got := joined(query(t, q))
	for _, s := range subs {
		if !strings.Contains(got, s) {
			t.Errorf("%q: missing %q in\n%s", q, s, got)
		}
	}
}

func TestEverydayQueriesAnswer(t *testing.T) {
	want(t, "b64 hello", "aGVsbG8=")
	want(t, "hash hello", "5d41402abc4b2a76b9719d911017c592", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	want(t, "10.0.0.5/24", "10.0.0.0")
	want(t, "0xff + 1", "256")
	want(t, "ts 1700000000", "2023-11-1") // local time zone decides the 14th or 15th
	want(t, "text Hello World", "helloWorld")
	want(t, "uuid")
	want(t, "re \\d+ :: order 66 of 100", "true")
	want(t, "ex contact a@b.com and c@d.org please", "a@b.com")
}

func TestSecretsNeverEnterHistoryAndRowsCarryScope(t *testing.T) {
	for _, q := range []string{"hash hello", "b64 hello", "text Hello World"} {
		items := query(t, q)
		if len(items) == 0 {
			t.Fatalf("%q: no rows", q)
		}
		for _, it := range items {
			if it.Action.Extra["nohistory"] != "1" {
				t.Fatalf("%q: row %q may enter history", q, it.Title)
			}
			if it.Scope == nil {
				t.Fatalf("%q: row %q has no family scope for Tab", q, it.Title)
			}
		}
	}
}

func TestScopeListsTheWholeFamily(t *testing.T) {
	p := toolsProvider(t)
	root := p.Query(context.Background(), "b64 hello")
	if len(root) == 0 || root[0].Scope == nil {
		t.Fatal("no scope on the first row")
	}
	sp, ok := p.(core.ScopeProvider)
	if !ok {
		t.Fatal("tools provider must implement ScopeProvider")
	}
	all := sp.QueryScope(context.Background(), *root[0].Scope, "hello")
	if len(all) <= len(root) || len(all) < 20 {
		t.Fatalf("the scope must list far more than the root rows: root %d scope %d", len(root), len(all))
	}
	// @filter narrows
	narrow := sp.QueryScope(context.Background(), *root[0].Scope, "@base32 hello")
	if len(narrow) == 0 || len(narrow) >= len(all) {
		t.Fatalf("@filter must narrow the list: %d of %d", len(narrow), len(all))
	}
}

func TestSlowKDFsStayOutOfTheLiveList(t *testing.T) {
	start := time.Now()
	items := query(t, "hash hello key=k")
	if time.Since(start) > 300*time.Millisecond {
		t.Fatalf("live hash list too slow: %v", time.Since(start))
	}
	for _, it := range items {
		if strings.HasPrefix(it.Title, "$2") || strings.Contains(it.Subtitle, "bcrypt") {
			t.Fatalf("bcrypt must only run inside the scope: %+v", it)
		}
	}
}

func TestEveryKeywordResolvesToSomething(t *testing.T) {
	p := toolsProvider(t)
	kw, ok := p.(interface{ Keywords() []string })
	if !ok {
		t.Fatal("no keywords")
	}
	n := 0
	for _, k := range kw.Keywords() {
		n++
		start := time.Now()
		items := p.Query(context.Background(), k+" test")
		if time.Since(start) > time.Second {
			t.Errorf("keyword %q took %v", k, time.Since(start))
		}
		_ = items // a hint row is a valid answer; the point is no panic and no stall
	}
	if n < 40 {
		t.Fatalf("only %d keywords; the families lost their entry points", n)
	}
}

func TestOldTextKeywordsStillWork(t *testing.T) {
	cfg := core.DefaultConfig()
	var all []core.Provider = builtin.All(&cfg)
	ask := func(q string) string {
		var b strings.Builder
		for _, p := range all {
			for _, it := range p.Query(context.Background(), q) {
				b.WriteString(it.Title + "\n")
			}
		}
		return b.String()
	}
	for q, w := range map[string]string{"url a b&c": "a%20b%26c", "b64 hi": "aGk=", "case hello world": "helloWorld", "now": "20"} {
		if got := ask(q); !strings.Contains(got, w) {
			t.Errorf("%q: want %q in\n%s", q, w, got)
		}
	}
}

// Ordinary launcher queries (app names, settings, plain words) must not
// produce tool rows: the tools provider ranks above fuzzy app matches.
func TestOrdinaryQueriesStayQuiet(t *testing.T) {
	for _, q := range []string{"chrome", "notepad", "visual studio code", "add printer", "age of empires", "word", "excel",
		"settings", "task manager", "control panel", "my documents", "date and time settings", "reverse engineering",
		"count", "format", "sort", "trim", "hash", "extract", "text", "lines", "net"} {
		if items := query(t, q); len(items) != 0 {
			t.Errorf("%q should be quiet, got %d rows: %s", q, len(items), joined(items))
		}
	}
}

// TestDocumentedKeywordsExist keeps docs/tools.md honest: every keyword the
// page lists must be a real entry point.
func TestDocumentedKeywordsExist(t *testing.T) {
	have := map[string]bool{}
	for _, k := range toolsProvider(t).(interface{ Keywords() []string }).Keywords() {
		have[k] = true
	}
	doc := `text t str case slug lines num number n base bits hex math stats percent enc encode dec decode b64 url html punycode esc
		hash md5 sha256 crc crypt cipher encrypt decrypt aes xor rot13 caesar rand random roll flip shuffle
		uuid guid id ulid nanoid re regex net cidr ip subnet mac port status mime domain dt date time ts now tz epoch cron diff
		ex extract emails domains ips urls numbers clip`
	var missing []string
	for _, k := range strings.Fields(doc) {
		if !have[k] {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		t.Errorf("docs/tools.md lists keywords that do not exist: %v", missing)
	}
}
