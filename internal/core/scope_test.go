package core

import (
	"context"
	"testing"
)

type scopeProv struct{ fakeProvider }

func (*scopeProv) QueryScope(ctx context.Context, s Scope, q string) []Item {
	return []Item{{ID: "s:1", Title: s.Label + ":" + q}}
}

func TestQueryScopeStaticFiltersAndProviderAnswers(t *testing.T) {
	c := NewCatalog(NewHistory(""))
	static := Scope{ID: "actions", Label: "Actions", Items: []Item{{Title: "Open"}, {Title: "Copy path"}, {Title: "Run as admin"}}}
	if got := c.QueryScope(context.Background(), static, ""); len(got) != 3 {
		t.Fatalf("empty filter should list everything: %v", got)
	}
	if got := c.QueryScope(context.Background(), static, "copy"); len(got) != 1 || got[0].Title != "Copy path" {
		t.Fatalf("filter: %v", got)
	}
	c.Register(&scopeProv{fakeProvider{name: "web"}})
	got := c.QueryScope(context.Background(), Scope{Label: "Web", Provider: "web"}, "go")
	if len(got) != 1 || got[0].Title != "Web:go" || got[0].Category != "web" {
		t.Fatalf("provider scope: %+v", got)
	}
	if got := c.QueryScope(context.Background(), Scope{Provider: "nope"}, "x"); len(got) != 0 {
		t.Fatalf("unknown provider: %v", got)
	}
}
