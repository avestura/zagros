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

// logKeep is how many entries are kept per extension.
const logKeep = 100

// LogBuffer keeps the most recent log entries of each extension in memory so
// the manager can show them without reading zagros.log.
type LogBuffer struct {
	mu   sync.Mutex
	byID map[string][]LogEntry
}

// NewLogBuffer returns an empty buffer.
func NewLogBuffer() *LogBuffer { return &LogBuffer{byID: map[string][]LogEntry{}} }

// Add records an entry, dropping the oldest of that extension beyond logKeep.
func (b *LogBuffer) Add(id, level, msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	l := append(b.byID[id], LogEntry{Time: time.Now(), ExtID: id, Level: level, Msg: msg})
	if len(l) > logKeep {
		l = append([]LogEntry(nil), l[len(l)-logKeep:]...)
	}
	b.byID[id] = l
}

// Entries returns the entries of one extension, oldest first.
func (b *LogBuffer) Entries(id string) []LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]LogEntry(nil), b.byID[id]...)
}

// Clear forgets the entries of one extension.
func (b *LogBuffer) Clear(id string) {
	b.mu.Lock()
	delete(b.byID, id)
	b.mu.Unlock()
}

// Logs returns the recent log entries of an extension, oldest first.
func (m *Manager) Logs(id string) []LogEntry { return m.opts.Logs.Entries(id) }

// ClearLogs forgets the in-memory log of an extension (zagros.log is kept).
func (m *Manager) ClearLogs(id string) { m.opts.Logs.Clear(id) }

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
