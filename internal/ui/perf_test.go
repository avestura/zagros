package ui

import (
	"testing"
	"time"

	"github.com/avestura/zagros/internal/core"
)

func ids(ss ...string) []core.Item {
	var out []core.Item
	for _, s := range ss {
		out = append(out, core.Item{ID: s, Title: s})
	}
	return out
}

func TestSelectionAfterUpdate(t *testing.T) {
	prev := ids("a", "b", "c")
	if got := SelectionAfterUpdate(prev, 2, true, ids("x", "c", "a")); got != 1 {
		t.Fatalf("moved selection must follow its ID, got %d", got)
	}
	if got := SelectionAfterUpdate(prev, 2, false, ids("x", "c", "a")); got != 0 {
		t.Fatalf("untouched selection stays on top, got %d", got)
	}
	if got := SelectionAfterUpdate(prev, 2, true, ids("x", "y")); got != 0 {
		t.Fatalf("vanished item falls back to the top, got %d", got)
	}
	if got := SelectionAfterUpdate(nil, 0, true, ids("x")); got != 0 {
		t.Fatalf("empty previous list: %d", got)
	}
}

func TestSameItems(t *testing.T) {
	if !SameItems(ids("a", "b"), ids("a", "b")) || SameItems(ids("a"), ids("a", "b")) || SameItems(ids("a"), ids("b")) {
		t.Fatal("SameItems wrong")
	}
}

func TestRateGate(t *testing.T) {
	g := rateGate{min: 16 * time.Millisecond}
	t0 := time.Unix(100, 0)
	if g.Wait(t0) != 0 {
		t.Fatal("first update must pass")
	}
	if w := g.Wait(t0.Add(5 * time.Millisecond)); w != 11*time.Millisecond {
		t.Fatalf("wait %v", w)
	}
	if g.Wait(t0.Add(16*time.Millisecond)) != 0 {
		t.Fatal("update after the interval must pass")
	}
}

func TestLatestSlotCoalesces(t *testing.T) {
	var s latestSlot[int]
	wakes := 0
	for i := 1; i <= 50; i++ {
		if s.Put(i) {
			wakes++
		}
	}
	if wakes != 1 {
		t.Fatalf("50 puts caused %d wake-ups", wakes)
	}
	if v, ok := s.Take(); !ok || v != 50 {
		t.Fatalf("take = %d %v, want the latest", v, ok)
	}
	if _, ok := s.Take(); ok {
		t.Fatal("slot must be empty after take")
	}
	if !s.Put(1) {
		t.Fatal("put after take must wake again")
	}
}

func TestIconQueueDedupeAndDrop(t *testing.T) {
	q := newIconQueue(3)
	a, b, c, d := iconKey{"a", 20}, iconKey{"b", 20}, iconKey{"c", 20}, iconKey{"d", 20}
	if !q.Push(a) || q.Push(a) {
		t.Fatal("duplicate push must be rejected")
	}
	q.Push(b)
	q.Push(c)
	q.Push(d) // over capacity: a (oldest) is dropped
	if q.Len() != 3 {
		t.Fatalf("len %d", q.Len())
	}
	q.SetVisible([]iconKey{b, d})
	if q.Len() != 2 {
		t.Fatalf("invisible rows must be dropped, len %d", q.Len())
	}
	k, ok := q.Pop()
	if !ok || k != d {
		t.Fatalf("newest first, got %v", k)
	}
	if q.Push(d) {
		t.Fatal("in-flight request must not be queued again")
	}
	q.Done(d)
	if !q.Push(d) {
		t.Fatal("finished request may be queued again")
	}
}

func TestIconLRUCapAndDestroy(t *testing.T) {
	destroyed := map[uintptr]bool{}
	c := newIconLRU(3, func(h uintptr) { destroyed[h] = true })
	c.Put(iconKey{"a", 1}, 1)
	c.Put(iconKey{"b", 1}, 2)
	c.Put(iconKey{"c", 1}, 3)
	c.Get(iconKey{"a", 1}) // touch the oldest
	c.Put(iconKey{"z", 1}, 9)
	if c.Len() != 3 || !destroyed[2] || destroyed[1] {
		t.Fatalf("len %d destroyed %v", c.Len(), destroyed)
	}
	c.Put(iconKey{"missing", 1}, 0) // negative entries cost nothing to destroy
	c.Trim(0)
	if c.Len() != 0 || len(destroyed) != 4 {
		t.Fatalf("trim: len %d destroyed %v", c.Len(), destroyed)
	}
}
