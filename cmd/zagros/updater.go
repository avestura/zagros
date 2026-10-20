//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avestura/zagros/internal/builtin"
	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/ext"
	"github.com/avestura/zagros/internal/market"
	"github.com/avestura/zagros/internal/ui"
	"github.com/avestura/zagros/internal/update"
	"golang.org/x/sys/windows"
)

// updateState is what the background updater knows. Everything here is in
// memory except the staged download, which lives in <data dir>/update.
type updateState struct {
	mu        sync.Mutex
	extUps    []market.Update
	extAt     time.Time
	staging   bool
	announced string // the version the user was told about this session
	extTold   int    // how many extension updates the user was told about
}

const (
	updateFirstDelay = 45 * time.Second
	updateInterval   = 6 * time.Hour
	updateAnnounce   = "✔ Update installed · Restart to update"
)

// isRelease reports whether this build has a release version (a "dev" build
// is never updated).
func isRelease() bool {
	_, ok := update.CompareVersions(version, "0.0.0")
	return ok
}

// wireUpdates connects the launcher's "update" keyword to the updater and
// starts the background loop when it is enabled in the config.
func (s *runtimeState) wireUpdates() {
	builtin.UpdateHooks.Staged = func() (string, bool) {
		st, ok := update.LoadStaged(core.DataDir(), version)
		return st.Version, ok
	}
	builtin.UpdateHooks.Download = s.downloadLatest
	builtin.UpdateHooks.Restart = s.restartForUpdate
	builtin.UpdateHooks.ExtUpdates = func() ([]builtin.ExtUpdate, time.Time) {
		s.upd.mu.Lock()
		defer s.upd.mu.Unlock()
		var out []builtin.ExtUpdate
		for _, u := range s.upd.extUps {
			out = append(out, builtin.ExtUpdate{ID: u.ID, Name: u.Entry.Name, Installed: u.Installed, Latest: u.Latest})
		}
		return out, s.upd.extAt
	}
	builtin.UpdateHooks.CheckExts = func(ctx context.Context) error { _, err := s.checkExtUpdates(ctx); return err }
	builtin.UpdateHooks.UpdateExt = s.updateExtension
	if !isRelease() {
		builtin.UpdateHooks.Download, builtin.UpdateHooks.Restart, builtin.UpdateHooks.Staged = nil, nil, nil
	}
	go s.updateLoop()
}

func (s *runtimeState) updateLoop() {
	if s.config().Updates.Background {
		// an update downloaded by an earlier session is still waiting
		if st, ok := update.LoadStaged(core.DataDir(), version); ok && isRelease() {
			s.announceUpdate(st.Version)
		}
	}
	time.Sleep(updateFirstDelay)
	for {
		if s.config().Updates.Background {
			s.updateTick()
		}
		time.Sleep(updateInterval)
	}
}

// updateTick checks for a newer Zagros (downloading it) and for newer
// extensions. Errors are logged; a failed check must never bother the user.
func (s *runtimeState) updateTick() {
	defer func() { _ = recover() }()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	if isRelease() {
		res, err := update.CheckCached(ctx, core.DataDir(), version, 5*time.Hour, update.Options{})
		switch {
		case err != nil:
			s.logs.Add("zagros", "warn", "background update check: "+err.Error())
		case res.Available:
			if _, ok := update.LoadStaged(core.DataDir(), version); !ok {
				if err := s.downloadLatest(ctx); err != nil {
					s.logs.Add("zagros", "warn", "background update download: "+err.Error())
				}
			}
			if st, ok := update.LoadStaged(core.DataDir(), version); ok {
				s.announceUpdate(st.Version)
			}
		}
	}
	if ups, err := s.checkExtUpdates(ctx); err != nil {
		s.logs.Add("zagros", "warn", "background extension check: "+err.Error())
	} else {
		s.announceExtUpdates(len(ups))
	}
}

// announceUpdate tells the user once per version, briefly: a tray balloon and
// a line on the empty launcher that disappears after a few seconds.
func (s *runtimeState) announceUpdate(v string) {
	s.upd.mu.Lock()
	again := s.upd.announced == v
	s.upd.announced = v
	s.upd.mu.Unlock()
	if again {
		return
	}
	if a := s.getApp(); a != nil {
		a.Flash(updateAnnounce)
		a.Notify("Zagros "+v+" is ready", updateAnnounce+". Type \"update\" in the launcher.")
	}
}

func (s *runtimeState) announceExtUpdates(n int) {
	s.upd.mu.Lock()
	told := s.upd.extTold
	s.upd.extTold = n
	s.upd.mu.Unlock()
	if n == 0 || n == told {
		return
	}
	if a := s.getApp(); a != nil {
		a.Notify("Extension updates", fmt.Sprintf("%d extension update(s) available. Type \"update\" in the launcher.", n))
	}
}

// downloadLatest stages the newest release (does nothing when one is staged).
func (s *runtimeState) downloadLatest(ctx context.Context) error {
	if !isRelease() {
		return errors.New("this is a development build")
	}
	s.upd.mu.Lock()
	if s.upd.staging {
		s.upd.mu.Unlock()
		return nil
	}
	s.upd.staging = true
	s.upd.mu.Unlock()
	defer func() { s.upd.mu.Lock(); s.upd.staging = false; s.upd.mu.Unlock() }()

	res, err := update.CheckCached(ctx, core.DataDir(), version, 0, update.Options{})
	if err != nil {
		return err
	}
	if !res.Available {
		return nil
	}
	if st, ok := update.LoadStaged(core.DataDir(), version); ok && st.Version == res.Latest.Version {
		return nil
	}
	if _, err := update.Stage(ctx, res.Latest, runtime.GOARCH, core.DataDir(), update.Options{}); err != nil {
		return err
	}
	s.logs.Add("zagros", "info", "downloaded Zagros "+res.Latest.Version+", ready to install")
	return nil
}

// restartForUpdate swaps in the staged executable and starts it. The new
// process waits for this one to exit (--wait-pid) before taking over, so the
// single-instance lock is free when it starts.
func (s *runtimeState) restartForUpdate() error {
	st, ok := update.LoadStaged(core.DataDir(), version)
	if !ok {
		return errors.New("no update is ready")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil {
		exe = r
	}
	if err := update.Apply(st, exe); err != nil {
		return err
	}
	cmd := exec.Command(exe, "--wait-pid", strconv.Itoa(os.Getpid()))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("the update is installed but Zagros could not start: %w", err)
	}
	if a := s.getApp(); a != nil {
		a.Quit()
	}
	return nil
}

// waitForPID blocks until the process exits (at most 20 seconds), so a
// restarted Zagros starts after the old one released its lock.
func waitForPID(pid uint32) {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return // already gone
	}
	defer windows.CloseHandle(h)
	_, _ = windows.WaitForSingleObject(h, 20000)
}

// ---- extensions ----

func (s *runtimeState) marketClient() *market.Client {
	return market.New(core.DataDir(), s.config().Marketplace.RegistryURL)
}

// checkExtUpdates asks the registry which installed extensions have a newer
// version and remembers the answer.
func (s *runtimeState) checkExtUpdates(ctx context.Context) ([]market.Update, error) {
	if s.mgr == nil {
		return nil, nil
	}
	ups, err := s.marketClient().CheckUpdates(ctx, installedFromRegistry(s.mgr))
	if err != nil {
		return nil, err
	}
	s.upd.mu.Lock()
	s.upd.extUps, s.upd.extAt = ups, time.Now()
	s.upd.mu.Unlock()
	if m := s.mgr; m != nil {
		m.Changed() // the manager shows the new "update" tags
	}
	return ups, nil
}

// installedFromRegistry maps id to version for the extensions that came from
// a registry or a GitHub release: a folder installed with "ext install
// <path>" is the author's own and is never offered a "newer" registry copy.
func installedFromRegistry(m *ext.Manager) map[string]string {
	out := map[string]string{}
	for _, e := range m.List() {
		if m.FromFolder(e.ID()) {
			continue
		}
		out[e.ID()] = e.Manifest.Version
	}
	return out
}

// updateExtension installs the registry's newest version of one extension
// over the installed one. Permissions it newly asks for are prompted for.
func (s *runtimeState) updateExtension(ctx context.Context, id string) error {
	if s.mgr == nil {
		return errors.New("extensions are not available")
	}
	pkg, err := s.marketClient().Resolve(ctx, id)
	if err != nil {
		return err
	}
	if pkg.Kind == market.KindTheme {
		return fmt.Errorf("%s is a theme", id)
	}
	if _, err := s.mgr.InstallBytes(ctx, pkg.Zip, pkg.Source); err != nil {
		return err
	}
	s.logs.Add("zagros", "info", "updated extension "+id+" to "+pkg.Manifest.Version)
	if _, err := s.checkExtUpdates(ctx); err != nil {
		return nil // updated; the list refreshes on the next check
	}
	return nil
}

// extUpdateVersions returns id -> newest version for the manager's tags.
func (s *runtimeState) extUpdateVersions() map[string]string {
	s.upd.mu.Lock()
	defer s.upd.mu.Unlock()
	out := map[string]string{}
	for _, u := range s.upd.extUps {
		out[u.ID] = u.Latest
	}
	return out
}

// ---- CLI ----

// cmdUpdateExts implements `zagros ext outdated` and `zagros ext update`.
func extOutdated(ctx context.Context) int {
	m, err := newCLIManager(false)
	if err != nil {
		return fail(err)
	}
	if _, err := m.Discover(ctx); err != nil {
		return fail(err)
	}
	c := market.New(core.DataDir(), mustConfig().Marketplace.RegistryURL)
	ups, err := c.CheckUpdates(ctx, installedFromRegistry(m))
	if err != nil {
		return fail(err)
	}
	if len(ups) == 0 {
		fmt.Println("All extensions are up to date.")
		return 0
	}
	for _, u := range ups {
		fmt.Printf("%-32s %s -> %s\n", u.ID, u.Installed, u.Latest)
	}
	fmt.Println("\nUpdate with: zagros ext update <id>   or   zagros ext update --all")
	return 0
}

func extUpdate(ctx context.Context, args []string) int {
	all := len(args) == 1 && args[0] == "--all"
	if len(args) != 1 {
		return fail(errors.New("usage: zagros ext update <id> | --all"))
	}
	m, err := newCLIManager(true)
	if err != nil {
		return fail(err)
	}
	if _, err := m.Discover(ctx); err != nil {
		return fail(err)
	}
	c := market.New(core.DataDir(), mustConfig().Marketplace.RegistryURL)
	ups, err := c.CheckUpdates(ctx, installedFromRegistry(m))
	if err != nil {
		return fail(err)
	}
	var todo []market.Update
	for _, u := range ups {
		if all || strings.EqualFold(u.ID, args[0]) {
			todo = append(todo, u)
		}
	}
	if len(todo) == 0 {
		if all {
			fmt.Println("All extensions are up to date.")
			return 0
		}
		return fail(fmt.Errorf("%s has no update (see: zagros ext outdated)", args[0]))
	}
	rc := 0
	for _, u := range todo {
		pkg, err := c.Resolve(ctx, u.ID)
		if err == nil {
			_, err = m.InstallBytes(ctx, pkg.Zip, pkg.Source)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", u.ID, err)
			rc = 1
			continue
		}
		fmt.Printf("updated %s %s -> %s\n", u.ID, u.Installed, u.Latest)
	}
	if ui.SignalReload() {
		fmt.Println("asked the running Zagros to reload")
	}
	return rc
}
