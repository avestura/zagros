package builtin

import (
	"reflect"
	"testing"
)

func TestEvCacheExtendAndBackspace(t *testing.T) {
	var c evCache
	c.put("rep", 20, []string{`C:\a\report.txt`, `C:\a\repo\x.go`, `C:\b\Reply.doc`})
	got, ok := c.reuse("repo", 20)
	if !ok || !reflect.DeepEqual(got, []string{`C:\a\report.txt`}) {
		t.Fatalf("extend: %v %v", got, ok)
	}
	// "x.go" has a dot and matches the file name only
	got, ok = c.reuse("rep go", 20)
	if !ok || len(got) != 0 {
		t.Fatalf("multi-term: %v %v", got, ok)
	}
	c.put("repo", 20, []string{`C:eport.txt`})
	if g, ok := c.reuse("rep", 20); !ok || len(g) != 3 {
		t.Fatalf("backspace should hit the cache: %v %v", g, ok)
	}
}

func TestEvCacheIncompleteAndOperators(t *testing.T) {
	var c evCache
	c.put("a", 2, []string{"1a", "2a"}) // limit hit: not the full set
	if _, ok := c.reuse("ab", 2); ok {
		t.Fatal("must not filter an incomplete result set")
	}
	c.put("b", 5, []string{"b1"})
	if _, ok := c.reuse("b*", 5); ok {
		t.Fatal("wildcards must go to Everything")
	}
	if _, ok := c.reuse(`b\x`, 5); ok {
		t.Fatal("path terms must go to Everything")
	}
	if g, ok := c.reuse("B", 5); !ok || len(g) != 1 {
		t.Fatal("exact repeat (case-insensitive) must hit")
	}
}
