package builtin

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

// ---- windows: switch to an open window ----

type windows struct{ base }

func newWindows() *windows { return &windows{base{"windows"}} }

func (w *windows) Ranked() bool       { return false }
func (w *windows) Keywords() []string { return []string{"w"} }

func (w *windows) Query(ctx context.Context, q string) []core.Item {
	rest, _, ok := trigger(q, "w")
	if !ok {
		return nil
	}
	var out []core.Item
	for _, wi := range listWindows() {
		if rest != "" && !containsFold(wi.Title, rest) {
			continue
		}
		h := wi.HWND
		out = append(out, core.Item{
			ID: fmt.Sprintf("windows:%d", h), Title: wi.Title, Subtitle: "Switch to window", Category: "Window",
			Action: core.Action{Kind: "func", Func: func() error { focusWindow(h); return nil }},
		})
	}
	return out
}

// ---- processes: kill <name> ----

type processes struct{ base }

func newProcesses() *processes { return &processes{base{"processes"}} }

func (p *processes) Ranked() bool       { return true }
func (p *processes) Keywords() []string { return []string{"kill"} }

func (p *processes) Query(ctx context.Context, q string) []core.Item {
	rest, _, ok := trigger(q, "kill")
	if !ok || rest == "" {
		return nil
	}
	self := uint32(os.Getpid())
	procs := listProcs()
	sort.Slice(procs, func(i, j int) bool { return strings.ToLower(procs[i].Name) < strings.ToLower(procs[j].Name) })
	var out []core.Item
	for _, pr := range procs {
		if pr.PID <= 4 || pr.PID == self || !containsFold(pr.Name, rest) {
			continue
		}
		pid := pr.PID
		out = append(out, core.Item{
			ID: fmt.Sprintf("processes:%s", strings.ToLower(pr.Name)), Title: "Kill " + pr.Name,
			Subtitle: fmt.Sprintf("PID %d", pr.PID), Category: "Process",
			Action: core.Action{Kind: "func", Func: func() error { return killPID(pid) }, Extra: map[string]string{"confirm": "1"}},
		})
		if len(out) >= 20 {
			break
		}
	}
	return out
}
