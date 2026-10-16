package core

// Tiered is optionally implemented by providers to say how they are
// scheduled by Catalog.QueryStream.
//
// Tier 0 (the default) is in-memory work, answered inline within the
// keystroke. Tier 1 is slow I/O (disk, IPC, processes, HTTP): it starts after
// a short debounce and has its own, longer budget, so it never delays the
// results of tier 0.
type Tiered interface {
	Tier() int
}

// Warmer is optionally implemented by providers with an index that should be
// loaded after startup, off the UI thread. Warm must not block for long and
// must be safe to call more than once.
type Warmer interface {
	Warm()
	Ready() bool
}

// ReadyNotifier is optionally implemented by providers that can tell when
// their data changed (an index landed) so the UI can re-run the query.
type ReadyNotifier interface {
	OnReady(fn func())
}

// TierOf returns the provider's tier (0 when it does not implement Tiered).
func TierOf(p Provider) int {
	if t, ok := p.(Tiered); ok && t.Tier() > 0 {
		return 1
	}
	return 0
}

// Snapshot is the merged, ranked state of one query at a point in time.
type Snapshot struct {
	Items []Item
	// Done is true on the last snapshot of a query.
	Done bool
	// Waiting is true while tier-1 providers are still running.
	Waiting bool
	// NotReady lists providers that were skipped because their index is
	// still loading.
	NotReady []string
}
