package builtin

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
)

// runner executes shell commands: "> dir" via cmd.exe, "ps Get-Process" via PowerShell.
type runner struct{ base }

func newRunner() *runner { return &runner{base{"runner"}} }

func (r *runner) Ranked() bool       { return true }
func (r *runner) Keywords() []string { return []string{">", "ps"} }

func (r *runner) Query(ctx context.Context, q string) []core.Item {
	q = strings.TrimLeft(q, " ")
	if strings.HasPrefix(q, ">") {
		cmd := strings.TrimSpace(q[1:])
		if cmd == "" {
			return []core.Item{hint("runner", "cmd", "> command", "run in cmd.exe")}
		}
		return []core.Item{{
			ID: "runner:cmd", Title: "Run: " + cmd, Subtitle: "cmd.exe (window stays open)", Category: "Command",
			Action: core.Action{Kind: "run", Target: "cmd.exe", Args: []string{"/k", cmd}},
			Alt:    []core.NamedAction{{Name: "Run and close", Action: core.Action{Kind: "run", Target: "cmd.exe", Args: []string{"/c", cmd}}}},
		}}
	}
	if cmd, _, ok := trigger(q, "ps"); ok {
		if cmd == "" {
			return []core.Item{hint("runner", "ps", "ps command", "run in PowerShell")}
		}
		return []core.Item{{
			ID: "runner:ps", Title: "PowerShell: " + cmd, Subtitle: "powershell.exe -NoExit", Category: "Command",
			Action: core.Action{Kind: "run", Target: "powershell.exe", Args: []string{"-NoExit", "-Command", cmd}},
		}}
	}
	return nil
}

// ---- recent files ----

type recent struct {
	base
	mu      sync.Mutex
	at      time.Time
	entries []recentEntry
}

type recentEntry struct {
	name, path string
	mod        time.Time
}

func newRecent() *recent { return &recent{base: base{"recent"}} }

func (r *recent) Ranked() bool       { return true }
func (r *recent) Tier() int          { return 1 }
func (r *recent) Keywords() []string { return []string{"recent"} }

func (r *recent) load() []recentEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Since(r.at) < 10*time.Second && r.entries != nil {
		return r.entries
	}
	r.at = time.Now()
	r.entries = nil
	dir := filepath.Join(os.Getenv("APPDATA"), `Microsoft\Windows\Recent`)
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, d := range des {
		if d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".lnk") {
			continue
		}
		fi, err := d.Info()
		if err != nil {
			continue
		}
		r.entries = append(r.entries, recentEntry{strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())), filepath.Join(dir, d.Name()), fi.ModTime()})
	}
	sort.Slice(r.entries, func(i, j int) bool { return r.entries[i].mod.After(r.entries[j].mod) })
	return r.entries
}

func (r *recent) Query(ctx context.Context, q string) []core.Item {
	rest, _, ok := trigger(q, "recent")
	if !ok {
		return nil
	}
	var out []core.Item
	for _, e := range r.load() {
		if rest != "" && !containsFold(e.name, rest) {
			continue
		}
		out = append(out, core.Item{
			ID: "recent:" + strings.ToLower(e.name), Title: e.name, Subtitle: e.mod.Format("2006-01-02 15:04"),
			Icon: e.path, Category: "Recent", Action: core.Action{Kind: "open", Target: e.path},
		})
		if len(out) >= 20 {
			break
		}
	}
	return out
}
