package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/core"
)

func TestWebSearchFallbackAndComplete(t *testing.T) {
	cfg := core.DefaultConfig()
	w := newWebSearch(&cfg)
	fb := w.Fallback("some thing")
	if len(fb) != 1 || !strings.Contains(fb[0].Action.Target, "some+thing") || fb[0].Complete != "" || fb[0].Scope == nil {
		t.Fatalf("fallback: %+v", fb)
	}
	if w.Fallback("example.com") != nil {
		t.Fatal("URL queries should not get a web-search fallback")
	}
	kw := cfg.WebSearch[0].Keyword
	items := w.Query(context.Background(), kw+" ")
	if len(items) != 1 || items[0].Complete != kw+" " {
		t.Fatalf("keyword hint should complete to %q: %+v", kw+" ", items)
	}
}

func TestCatalogAppendsFallbackWithinLimit(t *testing.T) {
	cfg := core.DefaultConfig()
	cat := core.NewCatalog(core.NewHistory(""))
	cat.SetLimit(3)
	cat.Register(newWebSearch(&cfg))
	items := cat.Query(context.Background(), "zzzz qqqq")
	if len(items) == 0 || items[len(items)-1].ID != "websearch:fallback" {
		t.Fatalf("expected trailing fallback, got %+v", items)
	}
	if len(items) > 3 {
		t.Fatalf("limit exceeded: %d", len(items))
	}
}

func TestWebSearchScopeListsEngines(t *testing.T) {
	cfg := core.DefaultConfig()
	w := newWebSearch(&cfg)
	fb := w.Fallback("rust lang")
	if len(fb) != 1 || fb[0].Scope == nil || fb[0].Scope.Provider != "websearch" || fb[0].Scope.Query != "rust lang" {
		t.Fatalf("fallback must carry a scope: %+v", fb)
	}
	items := w.QueryScope(context.Background(), *fb[0].Scope, "rust lang")
	if len(items) != len(cfg.WebSearch) {
		t.Fatalf("one item per engine, got %d", len(items))
	}
	for _, it := range items {
		if it.Action.Kind != "open" || !strings.Contains(it.Action.Target, "rust+lang") {
			t.Fatalf("engine item must open a search for the text: %+v", it)
		}
	}
	empty := w.QueryScope(context.Background(), *fb[0].Scope, "")
	if len(empty) != len(cfg.WebSearch) || empty[0].Action.Kind != "none" {
		t.Fatalf("empty query lists engines without actions: %+v", empty)
	}
}
