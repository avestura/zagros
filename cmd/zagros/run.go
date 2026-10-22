//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/assets"
	"github.com/avestura/zagros/internal/builtin"
	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/theme"
	"github.com/avestura/zagros/internal/ui"
	"golang.org/x/sys/windows/registry"
)

// runtimeState is everything the running launcher shares between goroutines.
type runtimeState struct {
	mu        sync.Mutex
	cfg       core.Config
	themeC    *theme.Theme
	cat       *core.Catalog
	hist      *core.History
	choices   *core.Choices
	mgr       *ext.Manager
	app       *ui.App
	mcpCancel context.CancelFunc // non-nil while the HTTP MCP server runs
	mcpURL    string
	builtin   []string       // provider names registered from builtin.All
	extProv   []string       // provider names registered from extensions
	logs      *ext.LogBuffer // every log line, in memory only
	logWin    *os.Process    // the open log console window, if any
	upd       updateState
}

func (s *runtimeState) config() core.Config { s.mu.Lock(); defer s.mu.Unlock(); return s.cfg }
func (s *runtimeState) getApp() *ui.App     { s.mu.Lock(); defer s.mu.Unlock(); return s.app }

func systemIsDark() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("AppsUseLightTheme")
	return err != nil || v == 0
}

func (s *runtimeState) theme() theme.Theme {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.themeC == nil {
		t := theme.Resolve(s.cfg.Theme, core.ThemeDir(), systemIsDark())
		s.themeC = &t
	} else if s.cfg.Theme == "" {
		// following Windows: re-evaluate cheaply
		if want := systemIsDark(); s.themeC.Dark != want {
			t := theme.Default(want)
			s.themeC = &t
		}
	}
	return *s.themeC
}

func (s *runtimeState) syncProviders() {
	cfg := s.config()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.builtin {
		s.cat.Unregister(n)
	}
	s.builtin = nil
	for _, p := range builtin.All(&cfg) {
		s.cat.Register(p)
		s.builtin = append(s.builtin, p.Name())
	}
	for _, n := range s.extProv {
		s.cat.Unregister(n)
	}
	s.extProv = nil
	if s.mgr != nil {
		for _, p := range s.mgr.Providers() {
			s.cat.Register(p)
			s.extProv = append(s.extProv, p.Name())
		}
	}
	s.cat.ApplyConfig(cfg)
	s.cat.Warm() // load indexes in the background, off the UI thread
}

func runApp() int { return runMode(false) }

// runBench starts a second, private instance (no single-instance lock) and
// types into it; see ui.RunInputBench.
func runBench() int { return runMode(true) }

func runMode(bench bool) int {
	if !bench {
		release, ok := ui.AcquireSingleInstance()
		if !ok {
			return 0 // the running instance was asked to show itself
		}
		defer release()
	}

	st := &runtimeState{}
	st.logs = ext.NewLogBuffer()
	log.SetFlags(0)
	log.SetOutput(logWriter{st.logs})
	if !bench {
		stopLogs := serveLogs(st.logs)
		defer stopLogs()
	}
	cfg, cfgErr := core.LoadConfig(core.ConfigPath())
	st.cfg = cfg
	st.hist = core.NewHistory(core.DefaultHistoryPath())
	st.cat = core.NewCatalog(st.hist)
	st.choices = core.NewChoices(core.DefaultChoicesPath())
	st.cat.SetChoices(st.choices)

	base := executor{}
	mgr, err := ext.NewManager(ext.Options{
		ExtDir:      core.ExtDir(),
		DataDir:     core.DataDir(),
		Prompter:    guiPrompter{app: st.getApp, mgr: func() *ext.Manager { return st.mgr }},
		Platform:    platform{app: st.getApp},
		Log:         func(string, string, string) {}, // kept in memory by Logs, never written to disk
		Logs:        st.logs,
		AppVersion:  strings.TrimPrefix(version, "v"),
		AI:          func() core.AIConfig { return st.config().AI },
		ReindexApps: builtin.ReindexApps,
	})
	var exec core.Executor = base
	if err == nil {
		st.mgr = mgr
		exec = mgr.WrapExecutor(base)
	}

	var admin ui.ExtAdmin
	var mk ui.Market
	if st.mgr != nil {
		admin = extAdmin{m: st.mgr, upd: st.extUpdateVersions, apply: st.updateExtension}
		mk = marketAdapter{c: market.New(core.DataDir(), cfg.Marketplace.RegistryURL), mgr: st.mgr}
	}

	reload := func() error {
		newCfg, err := core.LoadConfig(core.ConfigPath())
		st.mu.Lock()
		st.cfg = newCfg
		st.themeC = nil
		st.mu.Unlock()
		if st.mgr != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if _, derr := st.mgr.Discover(ctx); derr != nil && err == nil {
				err = derr
			}
		}
		st.syncProviders()
		if a := st.getApp(); a != nil {
			a.SetConfig(newCfg)
			a.SetTrayState(extSummary(st.mgr))
		}
		return err
	}

	reloadTheme := func() error {
		newCfg, err := core.LoadConfig(core.ConfigPath())
		st.mu.Lock()
		st.cfg = newCfg
		st.themeC = nil
		st.mu.Unlock()
		return err
	}

	app := ui.New(ui.Options{
		Config:  cfg,
		Theme:   st.theme,
		Catalog: st.cat,
		Exec:    exec,
		Record: func(it core.Item, q string) {
			if it.Action.Extra["nohistory"] == "1" {
				return
			}
			st.hist.Record(it.ID, q)
		},
		Choose: func(it core.Item, q string, sc *core.Scope) {
			switch {
			case sc != nil: // a scope's picks are remembered per scope, never with its text
				st.choices.RecordScope(sc.ID, it.ID)
			case it.Action.Extra["nohistory"] != "1": // the typed text of such rows may be a secret
				st.choices.RecordQuery(q, it.ID)
			}
		},
		Failed: func(it core.Item) { st.choices.Forget(it.ID) },
		Highlight: func(q, text string) []int {
			_, idx, ok := core.FuzzyMatch(q, text)
			if !ok {
				return nil
			}
			return idx
		},
		Reload:      reload,
		ReloadTheme: reloadTheme,
		CheckUpdate: updateMessage,
		MCPOn:       st.mcpOn,
		MCPToggle:   st.toggleMCP,
		LogsOn:      st.logsOn,
		LogsToggle:  st.toggleLogs,
		AgentPrompt: func() string { return agentPrompt(st.config(), st.mcpOn()) },
		ConfigPath:  core.ConfigPath(),
		ConfigDir:   core.ConfigDir(),
		Ext:         admin,
		Market:      mk,
		Themes:      themeAdmin{},
		Version:     version,
		FadeMS:      benchFade(bench), // no fade: a layered window makes every repaint slower
		IconICO:     assets.Icon,
		OnAutostart: func(on bool) {
			c := st.config()
			c.StartWithWindows = on
			_ = core.SaveConfig(core.ConfigPath(), c)
			st.mu.Lock()
			st.cfg = c
			st.mu.Unlock()
		},
		OnQuit: func() { st.stopMCP(); _ = st.hist.Flush(); _ = st.choices.Flush() },
		OnReady: func() {
			if cfgErr != nil {
				st.getApp().Notify("Zagros", "Config problem: "+cfgErr.Error())
			}
			if err != nil {
				st.getApp().Notify("Zagros", "Extensions disabled: "+err.Error())
			}
			// keep the Start with Windows entry in sync with the config file
			if c := st.config(); c.StartWithWindows != ui.AutostartEnabled() {
				_ = ui.SetAutostart(c.StartWithWindows)
			}
			if err := reload(); err != nil {
				st.getApp().Notify("Zagros", "Startup: "+err.Error())
			}
			st.wireMCP()
			if !bench {
				st.wireUpdates()
			}
			if bench {
				go st.getApp().RunInputBench(os.Stdout, st.waitReady)
			}
		},
	})
	st.mu.Lock()
	st.app = app
	st.mu.Unlock()
	if st.mgr != nil {
		st.mgr.OnChange(func() {
			st.syncProviders()
			app.SetTrayState(extSummary(st.mgr))
		})
	}
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "zagros:", err)
		return 1
	}
	return 0
}

func extSummary(m *ext.Manager) string {
	if m == nil {
		return ""
	}
	n := 0
	for _, e := range m.List() {
		if e.Enabled() {
			n++
		}
	}
	return fmt.Sprintf("%d extension(s) enabled", n)
}

// selfTest renders frames offscreen without showing any window.
func selfTest(shots string) int {
	app := ui.New(ui.Options{
		Config:  core.DefaultConfig(),
		IconICO: assets.Icon,
		Theme:   func() theme.Theme { return theme.Default(true) },
	})
	app.ShotsDir = shots
	if err := app.SelfTest(); err != nil {
		fmt.Fprintln(os.Stderr, "selftest FAILED:", err)
		return 1
	}
	fmt.Println("selftest ok")
	return 0
}

// waitReady blocks until every provider that loads an index reports ready.
func (s *runtimeState) waitReady() {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		ready := true
		for _, p := range s.cat.Providers() {
			if w, ok := p.(core.Warmer); ok && !w.Ready() {
				ready = false
			}
		}
		if ready {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// benchFade lets --bench-input compare against the old layered window
// (ZAGROS_BENCH_LAYERED=1); the shipped launcher never fades.
func benchFade(bench bool) int {
	if bench && os.Getenv("ZAGROS_BENCH_LAYERED") == "1" {
		return 70
	}
	return 0
}
