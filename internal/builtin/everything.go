package builtin

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/core"
)

type everything struct {
	base
	cfg    *core.Config
	once   sync.Once
	api    *evAPI
	es     string
	mu     sync.Mutex
	status string
	cache  evCache
}

func newEverything(cfg *core.Config) *everything {
	return &everything{base: base{"everything"}, cfg: cfg, status: "not initialised"}
}

func (e *everything) Ranked() bool       { return true }
func (e *everything) Tier() int          { return 1 }
func (e *everything) Keywords() []string { return []string{"f"} }

// Status describes the integration state for the UI ("ready (DLL)", "Everything is not running", ...).
func (e *everything) Status() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

func (e *everything) setStatus(s string) {
	e.mu.Lock()
	e.status = s
	e.mu.Unlock()
}

func (e *everything) init() {
	api, path, err := loadEverything(e.cfg.Everything.DLLPath)
	if err == nil {
		e.api = api
		e.setStatus("ready (" + path + ")")
		return
	}
	if es := findES(); es != "" {
		e.es = es
		e.setStatus("ready (es.exe: " + es + ")")
		return
	}
	e.setStatus("Everything SDK DLL and es.exe not found")
}

func findES() string {
	if p, err := exec.LookPath("es.exe"); err == nil {
		return p
	}
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if pf := os.Getenv(env); pf != "" {
			p := filepath.Join(pf, "Everything", "es.exe")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func (e *everything) Query(ctx context.Context, q string) []core.Item {
	limit := e.cfg.Everything.Limit
	if limit <= 0 {
		limit = 20
	}
	search, kw := "", false
	if rest, _, ok := trigger(q, "f"); ok {
		search, kw = rest, true
	} else {
		search = strings.TrimSpace(q)
		if len([]rune(search)) < 3 {
			return nil
		}
		if limit > 5 {
			limit = 5 // plain queries only append a few results
		}
	}
	if search == "" || ctx.Err() != nil {
		return nil
	}
	e.once.Do(e.init)
	_ = kw
	paths, hit := e.cache.reuse(search, limit)
	if !hit {
		switch {
		case e.api != nil:
			p, code, err := e.api.search(ctx, search, limit)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				if code == 2 {
					e.setStatus("Everything is not running")
				} else {
					e.setStatus("query failed: error " + strconv.Itoa(int(code)))
				}
				return nil
			}
			e.setStatus("ready")
			paths = p
		case e.es != "":
			// es.exe costs a process spawn: wait out fast typing first.
			t := time.NewTimer(esDebounce)
			select {
			case <-t.C:
			case <-ctx.Done():
				t.Stop()
				return nil
			}
			var ok bool
			if paths, ok = runES(ctx, e.es, search, limit); !ok {
				return nil
			}
		default:
			return nil
		}
		e.cache.put(search, limit, paths)
	}
	out := make([]core.Item, 0, len(paths))
	for _, p := range paths {
		out = append(out, fileItem(p))
	}
	return out
}

func fileItem(p string) core.Item {
	return core.Item{
		ID: "everything:" + strings.ToLower(p), Title: filepath.Base(p), Subtitle: filepath.Dir(p),
		Icon: p, Category: "File",
		Action: core.Action{Kind: "open", Target: p},
		Alt: []core.NamedAction{
			{Name: "Reveal in Explorer", Action: core.Action{Kind: "run", Target: "explorer.exe", Args: []string{"/select," + p}}},
			{Name: "Copy path", Action: copyAction(p)},
		},
	}
}

const esDebounce = 60 * time.Millisecond

// runES runs es.exe; cancelling ctx kills the process. ok is false on failure.
func runES(ctx context.Context, es, q string, limit int) (out []string, ok bool) {
	cmd := exec.CommandContext(ctx, es, "-n", strconv.Itoa(limit), q)
	hideCmd(cmd)
	b, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		if l := strings.TrimSpace(sc.Text()); l != "" {
			out = append(out, l)
		}
	}
	return out, true
}
