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
