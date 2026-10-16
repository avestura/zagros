package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Tracer records timing events. When off, Event is a single atomic load.
// It writes lines to a file (--trace / ZAGROS_TRACE=1) and/or keeps
// per-event statistics in memory (used by --selftest).
type Tracer struct {
	on    atomic.Bool
	mu    sync.Mutex
	w     io.Writer
	f     *os.File
	start time.Time
	stats map[string]*TraceStat
}

// TraceStat summarises all events with one name.
type TraceStat struct {
	Count int
	Total time.Duration
	Max   time.Duration
}

// Trace is the process-wide tracer.
var Trace = &Tracer{}

// TraceRequested reports whether tracing was asked for by environment.
func TraceRequested() bool {
	v := os.Getenv("ZAGROS_TRACE")
	return v != "" && v != "0"
}

// On reports whether events are recorded.
func (t *Tracer) On() bool { return t.on.Load() }

// EnableFile starts tracing into DataDir/trace.log (appending).
func (t *Tracer) EnableFile() error {
	_ = os.MkdirAll(DataDir(), 0o755)
	f, err := os.OpenFile(filepath.Join(DataDir(), "trace.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.f, t.w = f, f
	t.start = time.Now()
	if t.stats == nil {
		t.stats = map[string]*TraceStat{}
	}
	t.mu.Unlock()
	t.on.Store(true)
	return nil
}

// EnableMemory records statistics only (no file).
func (t *Tracer) EnableMemory() {
	t.mu.Lock()
	t.stats = map[string]*TraceStat{}
	t.start = time.Now()
	t.mu.Unlock()
	t.on.Store(true)
}

// Disable stops recording and closes the file.
func (t *Tracer) Disable() {
	t.on.Store(false)
	t.mu.Lock()
	if t.f != nil {
		_ = t.f.Close()
	}
	t.f, t.w = nil, nil
	t.mu.Unlock()
}

// Event records one measurement. detail is only formatted when tracing is on.
func (t *Tracer) Event(name string, d time.Duration, detail string, args ...any) {
	if !t.on.Load() {
		return
	}
	if len(args) > 0 {
		detail = fmt.Sprintf(detail, args...)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stats == nil {
		t.stats = map[string]*TraceStat{}
	}
	s := t.stats[name]
	if s == nil {
		s = &TraceStat{}
		t.stats[name] = s
	}
	s.Count++
	s.Total += d
	if d > s.Max {
		s.Max = d
	}
	if t.w != nil {
		fmt.Fprintf(t.w, "%s %-18s %8.2fms %s\n", time.Now().Format("15:04:05.000"), name, float64(d.Microseconds())/1000, detail)
	}
}

// Stat returns the statistics for one event name.
func (t *Tracer) Stat(name string) TraceStat {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s := t.stats[name]; s != nil {
		return *s
	}
	return TraceStat{}
}

// Summary formats the statistics, one line per event name.
func (t *Tracer) Summary() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	names := make([]string, 0, len(t.stats))
	for n := range t.stats {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, n := range names {
		s := t.stats[n]
		avg := time.Duration(0)
		if s.Count > 0 {
			avg = s.Total / time.Duration(s.Count)
		}
		fmt.Fprintf(&b, "  %-22s n=%-4d avg=%7.2fms max=%7.2fms\n", n, s.Count,
			float64(avg.Microseconds())/1000, float64(s.Max.Microseconds())/1000)
	}
	return b.String()
}
