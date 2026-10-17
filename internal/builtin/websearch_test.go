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
	if len(fb) != 1 || !strings.Contains(fb[0].Action.Target, "some+thing") || fb[0].Complete != "" || len(fb[0].Alt) != len(cfg.WebSearch)-1 {
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
