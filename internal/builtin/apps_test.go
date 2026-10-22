package builtin

import (
	"context"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/core"
)

func TestParseStartAppsKeepsPlainAppIDs(t *testing.T) {
	b := []byte(`[{"Name":"Docker Desktop","AppID":"Docker.DockerForWindows.Settings"},{"Name":"Calc","AppID":"Microsoft.WindowsCalculator_8wekyb3d8bbwe!App"},{"Name":"","AppID":"x"}]`)
	got := parseStartApps(b)
	if len(got) != 2 || got[0].P != `shell:AppsFolder\Docker.DockerForWindows.Settings` {
		t.Fatalf("got %+v", got)
	}
}

func TestReindexProviderRescansNow(t *testing.T) {
	t.Setenv("ZAGROS_HOME", t.TempDir())
	a := newApps(&core.Config{})
	a.once.Do(func() {})
	a.cache.Updated = time.Now()
	theApps.Store(a)
	defer theApps.Store(nil)

	items := newReindex().Query(context.Background(), "reindex")
	if len(items) != 1 || items[0].Action.Func == nil || items[0].Action.Kind != "func" {
		t.Fatalf("items: %+v", items)
	}
	if got := newReindex().Query(context.Background(), "chrome"); len(got) != 0 {
		t.Fatalf("other queries must not match: %+v", got)
	}
	if !ReindexApps() {
		t.Fatal("a fresh cache must still be rescanned on request")
	}
	for i := 0; a.scanning.Load() && i < 600; i++ {
		time.Sleep(50 * time.Millisecond)
	}
	if a.cache.Updated.IsZero() || time.Since(a.cache.Updated) > time.Minute || a.cache.Version != appsCacheVersion {
		t.Fatalf("cache not rewritten: %+v", a.cache.Updated)
	}
}
