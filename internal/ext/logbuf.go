package ext

import (
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogEntry is one line an extension, or the host on its behalf, logged.
type LogEntry struct {
	Time  time.Time
	ExtID string
	Level string // "info", "warn" or "error"
	Msg   string
}

const (
	logKeep   = 100  // entries kept per extension
	logAll    = 2000 // entries kept across all extensions, for the console
	logMaxMsg = 4000 // longer messages are cut
)

// LogBuffer keeps the most recent log entries in memory, per extension and
// overall, and streams new ones to subscribers. Nothing is written to disk.
type LogBuffer struct {
	mu   sync.Mutex
	byID map[string][]LogEntry
	all  []LogEntry
	subs map[int]func(LogEntry)
	next int
}

// NewLogBuffer returns an empty buffer.
func NewLogBuffer() *LogBuffer {
	return &LogBuffer{byID: map[string][]LogEntry{}, subs: map[int]func(LogEntry){}}
}

// Add records an entry, dropping the oldest beyond the limits, and passes it
// to every subscriber.
func (b *LogBuffer) Add(id, level, msg string) {
	if len(msg) > logMaxMsg {
		msg = msg[:logMaxMsg] + "..."
	}
	e := LogEntry{Time: time.Now(), ExtID: id, Level: level, Msg: msg}
	b.mu.Lock()
	b.byID[id] = keepLast(append(b.byID[id], e), logKeep)
	b.all = keepLast(append(b.all, e), logAll)
	subs := make([]func(LogEntry), 0, len(b.subs))
	for _, f := range b.subs {
		subs = append(subs, f)
	}
	b.mu.Unlock()
	for _, f := range subs {
		f(e)
	}
}

// keepLast trims l to its last n entries, copying so old ones can be freed.
func keepLast(l []LogEntry, n int) []LogEntry {
	if len(l) <= n {
		return l
	}
	return append([]LogEntry(nil), l[len(l)-n:]...)
}

// Entries returns the entries of one extension, oldest first.
func (b *LogBuffer) Entries(id string) []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]LogEntry(nil), b.byID[id]...)
}

// All returns the recent entries of every extension, oldest first.
func (b *LogBuffer) All() []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]LogEntry(nil), b.all...)
}

// Subscribe returns what is buffered now and calls fn for every entry added
// after that, without gaps or repeats. fn runs on the logging goroutine and
// must not block. cancel stops the calls.
func (b *LogBuffer) Subscribe(fn func(LogEntry)) (backlog []LogEntry, cancel func()) {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	b.subs[id] = fn
	return append([]LogEntry(nil), b.all...), func() {
		b.mu.Lock()
		delete(b.subs, id)
		b.mu.Unlock()
	}
}

// Clear forgets the entries of one extension.
func (b *LogBuffer) Clear(id string) {
	b.mu.Lock()
	delete(b.byID, id)
	b.mu.Unlock()
}

// Logs returns the recent log entries of an extension, oldest first.
func (m *Manager) Logs(id string) []LogEntry { return m.opts.Logs.Entries(id) }

// ClearLogs forgets the in-memory log of an extension.
func (m *Manager) ClearLogs(id string) { m.opts.Logs.Clear(id) }

// LogBuffer is where every extension's log output is collected.
func (m *Manager) LogBuffer() *LogBuffer { return m.opts.Logs }

// BrokenPrefix starts the id the manager lists a folder under when its
// extension failed to load, so it can still be seen, read about and removed.
const BrokenPrefix = "broken:"

// BrokenDir returns the folder of a failed extension listed as BrokenPrefix+name.
func (m *Manager) BrokenDir(id string) (string, bool) {
	name, ok := strings.CutPrefix(id, BrokenPrefix)
	if !ok || name == "" || name != filepath.Base(name) {
		return "", false
	}
	if _, bad := m.BrokenDirs()[name]; !bad {
		return "", false
	}
	return filepath.Join(m.opts.ExtDir, name), true
}

// FromFolder reports whether id was installed from a project folder with
// "ext install <path>" (the author's own copy, never offered registry updates).
func (m *Manager) FromFolder(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.records[id].Source == "local"
}

// Changed tells the UI that what it lists may have changed.
func (m *Manager) Changed() { m.changed() }
