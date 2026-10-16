package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/avestura/zagros/internal/core"
)

const appsCacheTTL = 12 * time.Hour

type appEntry struct {
	N string `json:"n"`           // display name
	P string `json:"p"`           // what to open (lnk/url/exe path, or shell:AppsFolder\AUMID)
	T string `json:"t,omitempty"` // resolved target, used for the icon
	U bool   `json:"u,omitempty"` // UWP/Store app
}

type appCache struct {
	Updated time.Time  `json:"updated"`
	Apps    []appEntry `json:"apps"`
	UWP     []appEntry `json:"uwp"`
}

type apps struct {
	base
	cfg      *core.Config
	once     sync.Once
	idx      atomic.Pointer[[]appEntry]
	cache    appCache
	scanning atomic.Bool
	warming  atomic.Bool
	cbMu     sync.Mutex
	onReady  func()
}

func newApps(cfg *core.Config) *apps { return &apps{base: base{"apps"}, cfg: cfg} }

func (a *apps) Ranked() bool { return false }

// Warm loads the cached index (and rescans when it is stale). Call it once at
// startup, off the UI thread; queries never block on it afterwards.
func (a *apps) Warm() {
	a.warming.Store(true)
	a.once.Do(a.init)
}

// Ready reports whether an index is available to search.
func (a *apps) Ready() bool { return a.idx.Load() != nil }

// OnReady sets fn, called whenever the index was (re)published.
func (a *apps) OnReady(fn func()) {
	a.cbMu.Lock()
	a.onReady = fn
	a.cbMu.Unlock()
}

func (a *apps) cachePath() string { return filepath.Join(core.DataDir(), "apps-index.json") }

func (a *apps) Query(ctx context.Context, q string) []core.Item {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	p := a.idx.Load()
	if p == nil && !a.warming.Load() { // CLI and tests: load on demand
		a.once.Do(a.init)
		p = a.idx.Load()
	}
	if p == nil {
		return nil
	}
	var out []core.Item
	for _, e := range *p {
		if ctx.Err() != nil {
			return out
		}
		if !subseq(q, e.N) {
			continue
		}
		it := core.Item{
			ID: "apps:" + strings.ToLower(e.N), Title: e.N, Category: "App",
			Action: core.Action{Kind: "open", Target: e.P},
		}
		if e.U {
			it.Subtitle = "Store app"
		} else {
			it.Subtitle = e.T
			if it.Subtitle == "" {
				it.Subtitle = e.P
			}
			it.Icon = e.T
			if it.Icon == "" {
				it.Icon = e.P
			}
		}
		out = append(out, it)
	}
	return out
}

func (a *apps) init() {
	if b, err := os.ReadFile(a.cachePath()); err == nil {
		if json.Unmarshal(b, &a.cache) == nil {
			a.publish()
		}
	}
	if time.Since(a.cache.Updated) > appsCacheTTL || len(a.cache.Apps) == 0 {
		a.Refresh()
	}
}

// Refresh rescans in the background. Safe to call at any time.
func (a *apps) Refresh() {
	if !a.scanning.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer a.scanning.Store(false)
		c := appCache{UWP: a.cache.UWP}
		c.Apps = a.scanShortcuts()
		a.setCache(c)
		if u := scanUWP(); u != nil {
			c.UWP = u
		}
		c.Updated = time.Now()
		a.setCache(c)
		if b, err := json.Marshal(&c); err == nil {
			_ = os.MkdirAll(filepath.Dir(a.cachePath()), 0o755)
			_ = os.WriteFile(a.cachePath(), b, 0o644)
		}
	}()
}

func (a *apps) setCache(c appCache) {
	a.cache = c
	a.publish()
}

func (a *apps) publish() {
	seen := make(map[string]bool, len(a.cache.Apps)+len(a.cache.UWP))
	all := make([]appEntry, 0, len(a.cache.Apps)+len(a.cache.UWP))
	for _, list := range [][]appEntry{a.cache.Apps, a.cache.UWP} {
		for _, e := range list {
			k := strings.ToLower(e.N)
			if seen[k] {
				continue
			}
			seen[k] = true
			all = append(all, e)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].N < all[j].N })
	a.idx.Store(&all)
	a.cbMu.Lock()
	fn := a.onReady
	a.cbMu.Unlock()
	if fn != nil {
		fn()
	}
}

func startMenuDirs() []string {
	var d []string
	if pd := os.Getenv("ProgramData"); pd != "" {
		d = append(d, filepath.Join(pd, `Microsoft\Windows\Start Menu\Programs`))
	}
	if ad := os.Getenv("APPDATA"); ad != "" {
		d = append(d, filepath.Join(ad, `Microsoft\Windows\Start Menu\Programs`))
	}
	return d
}

var skipTargetExt = map[string]bool{".chm": true, ".txt": true, ".html": true, ".htm": true, ".pdf": true, ".rtf": true, ".md": true, ".log": true, ".ini": true}

func (a *apps) scanShortcuts() []appEntry {
	res, _ := newLnkResolver() // nil when unavailable: fall back to the shortcut path
	if res != nil {
		defer res.Close()
	}
	var out []appEntry
	walk := func(root string, maxDepth int, exes bool) {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if maxDepth > 0 && strings.Count(strings.TrimPrefix(p, root), string(filepath.Separator)) >= maxDepth {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			name := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
			if strings.Contains(strings.ToLower(name), "uninstall") {
				return nil
			}
			switch ext {
			case ".lnk":
				e := appEntry{N: name, P: p}
				if res != nil {
					if t, _, err := res.Resolve(p); err == nil && t != "" {
						if skipTargetExt[strings.ToLower(filepath.Ext(t))] {
							return nil
						}
						e.T = t
					}
				}
				out = append(out, e)
			case ".url":
				out = append(out, appEntry{N: name, P: p})
			case ".exe":
				if exes {
					out = append(out, appEntry{N: name, P: p, T: p})
				}
			}
			return nil
		})
	}
	for _, d := range startMenuDirs() {
		walk(d, 0, false)
	}
	for _, d := range a.cfg.ScanDirs {
		walk(os.ExpandEnv(d), 3, true)
	}
	return out
}

// scanUWP lists Store apps through Get-StartApps (slow, so only run in the background).
func scanUWP() []appEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
		"Get-StartApps | Select-Object Name,AppID | ConvertTo-Json -Compress")
	hideCmd(cmd)
	b, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseStartApps(b)
}

func parseStartApps(b []byte) []appEntry {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	var list []struct{ Name, AppID string }
	if err := json.Unmarshal(b, &list); err != nil {
		var one struct{ Name, AppID string }
		if json.Unmarshal(b, &one) != nil {
			return nil
		}
		list = append(list, one)
	}
	var out []appEntry
	for _, x := range list {
		if x.Name == "" || !strings.Contains(x.AppID, "!") { // only AUMIDs: Package!App
			continue
		}
		out = append(out, appEntry{N: x.Name, P: `shell:AppsFolder\` + x.AppID, U: true})
	}
	return out
}
