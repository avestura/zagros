package core

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakeProv struct {
	name  string
	tier  int
	delay time.Duration
	items []Item
	rank  bool
}

func (f *fakeProv) Name() string { return f.name }
func (f *fakeProv) Ranked() bool { return f.rank }
func (f *fakeProv) Tier() int    { return f.tier }
func (f *fakeProv) Query(ctx context.Context, q string) []Item {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil
		}
	}
	return f.items
}

func mk(prefix string, n int) []Item {
	var out []Item
	for i := 0; i < n; i++ {
		out = append(out, Item{ID: fmt.Sprintf("%s:%d", prefix, i), Title: fmt.Sprintf("alpha %s %d", prefix, i)})
	}
	return out
}

func TestStreamFastResultsNotDelayedBySlowTier(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProv{name: "fast", items: mk("f", 3)})
	c.Register(&fakeProv{name: "slow", tier: 1, delay: 300 * time.Millisecond, rank: true, items: mk("s", 2)})
	start := time.Now()
	var firstWithItems time.Duration
	seen := false
	var snaps []Snapshot
	c.QueryStream(context.Background(), "alpha", func(s Snapshot) {
		if !seen && len(s.Items) > 0 {
			seen = true
			firstWithItems = time.Since(start)
		}
		snaps = append(snaps, s)
	})
	if firstWithItems > 100*time.Millisecond {
		t.Fatalf("fast results took %v", firstWithItems)
	}
	last := snaps[len(snaps)-1]
	if !last.Done || len(last.Items) != 5 {
		t.Fatalf("last snapshot done=%v items=%d", last.Done, len(last.Items))
	}
	if len(snaps) < 2 || !snaps[0].Waiting {
		t.Fatalf("expected a waiting snapshot first, got %d snapshots, first waiting=%v", len(snaps), snaps[0].Waiting)
	}
}

func TestStreamRankStableRegardlessOfArrival(t *testing.T) {
	run := func(slowDelay, fastDelay time.Duration) []string {
		c := NewCatalog(nil)
		c.Register(&fakeProv{name: "a", delay: slowDelay, rank: true, items: mk("a", 3)})
		c.Register(&fakeProv{name: "b", delay: fastDelay, rank: true, items: mk("b", 3)})
		var last Snapshot
		c.QueryStream(context.Background(), "alpha", func(s Snapshot) { last = s })
		var ids []string
		for _, it := range last.Items {
			ids = append(ids, it.ID)
		}
		return ids
	}
	x := run(30*time.Millisecond, 0)
	y := run(0, 30*time.Millisecond)
	if fmt.Sprint(x) != fmt.Sprint(y) {
		t.Fatalf("order depends on arrival:\n%v\n%v", x, y)
	}
}

func TestStreamEmptyTier0StillEmits(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProv{name: "fast"})
	c.Register(&fakeProv{name: "slow", tier: 1, delay: 100 * time.Millisecond, rank: true, items: mk("s", 1)})
	var snaps []Snapshot
	c.QueryStream(context.Background(), "zzz", func(s Snapshot) { snaps = append(snaps, s) })
	if len(snaps) < 2 || len(snaps[0].Items) != 0 || snaps[0].Done {
		t.Fatalf("want empty interim snapshot then final, got %+v", snaps)
	}
}

func TestStreamCancelStopsEmitting(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProv{name: "slow", tier: 1, delay: time.Second, rank: true, items: mk("s", 1)})
	ctx, cancel := context.WithCancel(context.Background())
	var mu sync.Mutex
	n := 0
	done := make(chan struct{})
	go func() {
		c.QueryStream(ctx, "x", func(Snapshot) { mu.Lock(); n++; mu.Unlock() })
		close(done)
	}()
	time.Sleep(80 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("QueryStream did not return after cancel")
	}
	mu.Lock()
	defer mu.Unlock()
	if n != 1 { // only the interim "waiting" snapshot, no Done
		t.Fatalf("emitted %d snapshots", n)
	}
}

func TestQueryWrapperWaitsForSlowTier(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProv{name: "fast", items: mk("f", 1)})
	c.Register(&fakeProv{name: "slow", tier: 1, delay: 60 * time.Millisecond, rank: true, items: mk("s", 1)})
	if got := len(c.Query(context.Background(), "alpha")); got != 2 {
		t.Fatalf("Query returned %d items, want 2", got)
	}
}

func TestTraceStats(t *testing.T) {
	tr := &Tracer{}
	tr.Event("x", time.Second, "ignored while off")
	if tr.Stat("x").Count != 0 {
		t.Fatal("recorded while off")
	}
	tr.EnableMemory()
	tr.Event("x", 2*time.Millisecond, "a")
	tr.Event("x", 4*time.Millisecond, "b")
	s := tr.Stat("x")
	if s.Count != 2 || s.Max != 4*time.Millisecond || s.Total != 6*time.Millisecond {
		t.Fatalf("stat %+v", s)
	}
}

type failProv struct {
	fakeProv
	msg string
}

func (f *failProv) Failure() string { return f.msg }

func TestStreamNamesSlowAndFailedProviders(t *testing.T) {
	c := NewCatalog(nil)
	c.Register(&fakeProv{name: "slowpoke", tier: 1, delay: 700 * time.Millisecond, items: mk("s", 1)})
	c.Register(&failProv{fakeProv: fakeProv{name: "broken", tier: 1}, msg: "no network"})
	var snaps []Snapshot
	c.QueryStream(context.Background(), "alpha", func(s Snapshot) { snaps = append(snaps, s) })
	slow := false
	for _, s := range snaps {
		if len(s.Slow) == 1 && s.Slow[0] == "slowpoke" {
			slow = true
		}
	}
	if !slow {
		t.Fatalf("no snapshot named the provider still running: %+v", snaps)
	}
	last := snaps[len(snaps)-1]
	if !last.Done || len(last.Problems) != 1 || last.Problems[0] != "broken: no network" {
		t.Fatalf("problems = %v", last.Problems)
	}
}

func TestStreamReportsTimedOutProvider(t *testing.T) {
	c := NewCatalog(nil)
	c.slowTimeout = 150 * time.Millisecond
	c.Register(&fakeProv{name: "hang", tier: 1, delay: 5 * time.Second})
	var last Snapshot
	c.QueryStream(context.Background(), "alpha", func(s Snapshot) { last = s })
	if len(last.Problems) != 1 || last.Problems[0] != "hang did not answer in time" {
		t.Fatalf("problems = %v", last.Problems)
	}
}
