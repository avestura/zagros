package builtin

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/update"
)

func TestUpdateProviderReportsAvailableAndRefreshes(t *testing.T) {
	p := newUpdate()
	ready := make(chan struct{}, 4)
	p.OnReady(func() { ready <- struct{}{} })
	p.check = func(ctx context.Context, force bool) (update.Result, error) {
		return update.Result{Current: "0.2.0", Available: true, Checked: time.Now(), Latest: update.Release{
			Version: "0.3.0", URL: "https://example/r", Assets: []update.Asset{
				{Name: "zagros_0.3.0_windows_amd64.zip", URL: "https://example/a64", Size: 5 << 20},
				{Name: "zagros_0.3.0_windows_arm64.zip", URL: "https://example/aarm", Size: 5 << 20}}}}, nil
	}
	items := p.Query(context.Background(), "update")
	if len(items) != 1 || !strings.Contains(items[0].Title, "Checking") {
		t.Fatalf("first answer must say checking: %+v", items)
	}
	<-ready // the provider tells the launcher to re-run the query
	items = p.Query(context.Background(), "update")
	if len(items) != 1 || items[0].Title != "Zagros 0.3.0 is available" || items[0].Action.Target != "https://example/r" {
		t.Fatalf("available: %+v", items)
	}
	if len(items[0].Alt) < 3 {
		t.Fatalf("expected download, copy and recheck alternatives: %+v", items[0].Alt)
	}
	if items[0].Action.Extra["nohistory"] != "1" {
		t.Fatal("update rows must not enter history")
	}
	if p.Query(context.Background(), "hello") != nil {
		t.Fatal("other queries must not trigger the update provider")
	}
}

func TestUpdateProviderError(t *testing.T) {
	p := newUpdate()
	done := make(chan struct{}, 2)
	p.OnReady(func() { done <- struct{}{} })
	p.check = func(context.Context, bool) (update.Result, error) { return update.Result{}, errors.New("offline") }
	p.Query(context.Background(), "update")
	<-done
	items := p.Query(context.Background(), "update")
	if len(items) != 1 || !strings.Contains(items[0].Title, "Could not") || items[0].Subtitle != "offline" {
		t.Fatalf("error row: %+v", items)
	}
}

func TestUpdateProviderUpToDate(t *testing.T) {
	p := newUpdate()
	p.res = &update.Result{Current: "0.3.0", Checked: time.Now(), Latest: update.Release{Version: "0.3.0"}}
	items := p.Query(context.Background(), "update")
	if len(items) != 1 || items[0].Title != "Zagros is up to date" {
		t.Fatalf("up to date: %+v", items)
	}
}

func withHooks(t *testing.T) {
	t.Helper()
	old := UpdateHooks
	t.Cleanup(func() { UpdateHooks = old })
}

func TestUpdateProviderStagedOffersRestart(t *testing.T) {
	withHooks(t)
	restarted := false
	UpdateHooks.Staged = func() (string, bool) { return "0.9.0", true }
	UpdateHooks.Restart = func() error { restarted = true; return nil }
	p := newUpdate()
	p.res = &update.Result{Current: "0.8.1", Available: true, Checked: time.Now(), Latest: update.Release{Version: "0.9.0"}}
	items := p.Query(context.Background(), "update")
	if len(items) != 1 || !strings.Contains(items[0].Title, "Restart to update to Zagros 0.9.0") || !strings.Contains(items[0].Subtitle, "Update installed") {
		t.Fatalf("staged: %+v", items)
	}
	if items[0].Action.Kind != "func" || items[0].Action.Func() != nil || !restarted {
		t.Fatalf("Enter must restart: %+v restarted=%v", items[0].Action, restarted)
	}
}

func TestUpdateProviderAvailableDownloadsInBackground(t *testing.T) {
	withHooks(t)
	done := make(chan struct{})
	UpdateHooks.Download = func(ctx context.Context) error { <-done; return nil }
	p := newUpdate()
	ready := make(chan struct{}, 8)
	p.OnReady(func() { ready <- struct{}{} })
	p.res = &update.Result{Current: "0.8.1", Available: true, Checked: time.Now(), Latest: update.Release{Version: "0.9.0"}}
	items := p.Query(context.Background(), "update")
	if len(items) != 1 || items[0].Action.Kind != "func" || !strings.Contains(items[0].Subtitle, "downloads and installs") {
		t.Fatalf("available with a Download hook: %+v", items)
	}
	_ = items[0].Action.Func()
	if got := p.Query(context.Background(), "update"); len(got) != 1 || !strings.Contains(got[0].Title, "Downloading") {
		t.Fatalf("while downloading: %+v", got)
	}
	close(done)
	<-ready
	staged := false
	UpdateHooks.Staged = func() (string, bool) { staged = true; return "0.9.0", true }
	if got := p.Query(context.Background(), "update"); !staged || !strings.Contains(got[0].Title, "Restart to update") {
		t.Fatalf("after the download: %+v", got)
	}
}

func TestUpdateProviderDownloadFailureCanRetry(t *testing.T) {
	withHooks(t)
	UpdateHooks.Download = func(ctx context.Context) error { return errors.New("checksum mismatch") }
	p := newUpdate()
	ready := make(chan struct{}, 8)
	p.OnReady(func() { ready <- struct{}{} })
	p.res = &update.Result{Current: "0.8.1", Available: true, Checked: time.Now(), Latest: update.Release{Version: "0.9.0"}}
	p.download()
	<-ready
	got := p.Query(context.Background(), "update")
	if len(got) != 1 || got[0].Title != "Could not install the update" || !strings.Contains(got[0].Subtitle, "checksum mismatch") {
		t.Fatalf("failure row: %+v", got)
	}
}

func TestUpdateProviderListsExtensionUpdates(t *testing.T) {
	withHooks(t)
	var updated []string
	UpdateHooks.ExtUpdates = func() ([]ExtUpdate, time.Time) {
		return []ExtUpdate{{ID: "weather", Name: "Weather", Installed: "1.0.0", Latest: "1.1.0"}, {ID: "wiki", Name: "Wikipedia", Installed: "1.0.0", Latest: "1.2.0"}}, time.Now()
	}
	UpdateHooks.UpdateExt = func(ctx context.Context, id string) error { updated = append(updated, id); return nil }
	p := newUpdate()
	p.res = &update.Result{Current: "0.8.1", Checked: time.Now(), Latest: update.Release{Version: "0.8.1"}}
	ready := make(chan struct{}, 16)
	p.OnReady(func() { ready <- struct{}{} })
	items := p.Query(context.Background(), "update")
	if len(items) != 2 || items[1].Title != "2 extension updates available" || len(items[1].Alt) != 2 {
		t.Fatalf("extension row: %+v", items)
	}
	if !strings.Contains(items[1].Subtitle, "Weather 1.0.0 to 1.1.0") {
		t.Fatalf("subtitle: %q", items[1].Subtitle)
	}
	_ = items[1].Alt[1].Action.Func() // only the second
	for i := 0; i < 200 && len(updated) < 1; i++ {
		time.Sleep(5 * time.Millisecond)
	}
	if len(updated) != 1 || updated[0] != "wiki" {
		t.Fatalf("updated %v", updated)
	}
	UpdateHooks.ExtUpdates = func() ([]ExtUpdate, time.Time) { return nil, time.Now() }
	for i := 0; i < 200; i++ { // wait until the update goroutine is idle again
		p.mu.Lock()
		busy := p.exBusy
		p.mu.Unlock()
		if busy == "" {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := p.Query(context.Background(), "update"); len(got) != 2 || got[1].Title != "Extensions are up to date" {
		t.Fatalf("up to date: %+v", got)
	}
}
