package ext

import (
	"context"
	"fmt"
	"github.com/avestura/zagros/internal/core"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchAllOrderConcurrencyAndErrors(t *testing.T) {
	var cur, peak atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := cur.Add(1)
		defer cur.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(150 * time.Millisecond)
		fmt.Fprintf(w, "%s %s", r.Method, r.URL.Path)
	}))
	defer srv.Close()
	u := srv.URL
	perms := []Permission{"network:127.0.0.1"}

	start := time.Now()
	got := mustRun(t, `
rs = http.fetch_all(["`+u+`/a", {"url": "`+u+`/b", "method": "post"}, "`+u+`/c", "`+u+`/d", "`+u+`/e", "`+u+`/f"], max_concurrency = 3)
out = [[r.text for r in rs], len(rs)]`, perms, Options{})
	if got != `[["GET /a", "POST /b", "GET /c", "GET /d", "GET /e", "GET /f"], 6]` {
		t.Fatal(got)
	}
	if peak.Load() != 3 {
		t.Fatalf("peak concurrency %d, want 3", peak.Load())
	}
	if d := time.Since(start); d > 700*time.Millisecond {
		t.Fatalf("six 150ms requests at 3 at a time took %v", d)
	}

	// A host that is not granted is an error entry, the others still run.
	got = mustRun(t, `
rs = http.fetch_all(["`+u+`/a", "http://example.org/x", {"url": "`+u+`/big", "max_bytes": 1}])
out = [rs[0].ok, rs[1].ok, rs[1].status, "permission denied" in rs[1].error, rs[2].ok, "larger than 1" in rs[2].error]`, perms, Options{})
	if got != `[True, False, 0, True, False, True]` {
		t.Fatal(got)
	}

	for _, body := range []string{
		`out = http.fetch_all(5)`,
		`out = http.fetch_all([5])`,
		`out = http.fetch_all([{"method": "GET"}])`,
		`out = http.fetch_all([{"url": "` + u + `", "bogus": 1}])`,
		`out = http.fetch_all([], max_concurrency = 99)`,
		`out = http.fetch_all([ "` + u + `" ] * 101)`,
	} {
		if _, err := runScript(t, body, perms, Options{}); err == nil {
			t.Errorf("want an error for %s", body)
		}
	}
	if got := mustRun(t, `out = http.fetch_all([])`, perms, Options{}); got != "[]" {
		t.Fatal(got)
	}
}

func TestFetchAllTimeLimitCoversWholeCall(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	start := time.Now()
	_, err := runScript(t, `rs = http.fetch_all(["`+srv.URL+`/a", "`+srv.URL+`/b"], max_concurrency = 1)
out = len(rs)`, []Permission{"network:127.0.0.1"}, Options{Limits: Limits{CallTimeout: 300 * time.Millisecond}})
	if err == nil {
		t.Fatal("the hook limit must apply to the whole call")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("took %v", d)
	}
}

const taskManifest = baseManifest + "keywords = [\"t\"]\n"

// taskExt starts an extension whose notify callback counts finished tasks.
func taskExt(t *testing.T, script string, granted []Permission, o Options) (*Extension, chan struct{}) {
	t.Helper()
	e := openStarted(t, baseManifest, script, granted, o)
	t.Cleanup(e.Stop)
	ch := make(chan struct{}, 16)
	NewProvider(e).OnReady(func() { ch <- struct{}{} })
	return e, ch
}

func waitNotify(t *testing.T, ch chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("no completion notice")
	}
}

const slowTaskScript = `
def slow(ctx, path):
    r = http.get(SRV + path, timeout = 20.0)
    return {"body": r.text, "n": 3}

def on_query(ctx, q):
    t = task("slow", "/" + q)
    if not t.done:
        return [item(id = "w", title = "wait " + q, pending = True)]
    if not t.ok:
        return [item(id = "e", title = "err " + t.error)]
    return [item(id = "r", title = t.value["body"] + str(t.value["n"]))]
`

func TestTaskRunsInBackgroundAndFeedsNextQuery(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-release:
			fmt.Fprint(w, "done:"+r.URL.Path)
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	script := strings.Replace(slowTaskScript, "SRV", `"`+srv.URL+`"`, 1)
	// The hook limit is 300ms, far below what the task needs.
	e, ch := taskExt(t, script, []Permission{"network:127.0.0.1"}, Options{Limits: Limits{CallTimeout: 300 * time.Millisecond}})
	ctx := context.Background()

	start := time.Now()
	items, err := e.Query(ctx, "a")
	if err != nil || len(items) != 1 || !strings.HasPrefix(items[0].Title, "wait a") || items[0].Subtitle != "Loading..." || items[0].Action.Kind != "none" {
		t.Fatalf("first query: %v %+v", err, items)
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatal("on_query must return at once")
	}
	if _, err := e.Query(ctx, "a"); err != nil { // same call: shares the run
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond) // longer than the hook limit
	close(release)
	waitNotify(t, ch)
	items, err = e.Query(ctx, "a")
	if err != nil || len(items) != 1 || items[0].Title != "done:/a3" {
		t.Fatalf("second query: %v %+v", err, items)
	}
	if hits.Load() != 1 {
		t.Fatalf("server hit %d times, want 1", hits.Load())
	}
}

func TestTaskCancelledWhenQueryChanges(t *testing.T) {
	gone := make(chan struct{})
	var once sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		if r.URL.Path == "/a" {
			once.Do(func() { close(gone) })
		}
	}))
	defer srv.Close()
	script := strings.Replace(slowTaskScript, "SRV", `"`+srv.URL+`"`, 1)
	e, _ := taskExt(t, script, []Permission{"network:127.0.0.1"}, Options{})
	ctx := context.Background()
	if _, err := e.Query(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := e.Query(ctx, "b"); err != nil { // asks for another task: "a" is dropped
		t.Fatal(err)
	}
	select {
	case <-gone:
	case <-time.After(3 * time.Second):
		t.Fatal("the task for the old query kept running")
	}
	e.tasks.mu.Lock()
	n := len(e.tasks.byKey)
	e.tasks.mu.Unlock()
	if n != 1 {
		t.Fatalf("%d tasks kept, want 1", n)
	}
}

func TestTaskCancelledOnStop(t *testing.T) {
	gone := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(gone)
	}))
	defer srv.Close()
	script := strings.Replace(slowTaskScript, "SRV", `"`+srv.URL+`"`, 1)
	e, _ := taskExt(t, script, []Permission{"network:127.0.0.1"}, Options{})
	if _, err := e.Query(context.Background(), "a"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	e.Stop()
	select {
	case <-gone:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not cancel the task")
	}
}

func TestTaskErrorsPermissionsAndNesting(t *testing.T) {
	script := `
def boom(ctx, a):
    fail("bad " + a)

def readit(ctx, a):
    return fs.read_text("/nope.txt")

def nest(ctx, a):
    return task("boom", "x")

def good(ctx, a):
    return [a, ctx.id]

def on_query(ctx, q):
    t = task(q, "x")
    if not t.done:
        return [item(id = "w", title = "wait", pending = True)]
    return [item(id = "r", title = str(t.ok) + "|" + str(t.value) + "|" + str(t.error))]
`
	e, ch := taskExt(t, script, nil, Options{})
	ctx := context.Background()
	for _, tc := range []struct{ q, want string }{
		{"boom", "False|None|"},
		{"readit", "permission denied"},
		{"nest", "cannot start another task"},
		{"good", `True|["x", "test.ext"]|None`},
	} {
		if _, err := e.Query(ctx, tc.q); err != nil {
			t.Fatal(err)
		}
		waitNotify(t, ch)
		items, err := e.Query(ctx, tc.q)
		if err != nil || len(items) != 1 || !strings.Contains(items[0].Title, tc.want) {
			t.Errorf("%s: %v %+v, want %q", tc.q, err, items, tc.want)
		}
	}
	// bad arguments to task() itself raise in the caller
	for _, s := range []string{`task("missing")`, `task("good", ttl = 0)`, `task("good", args = {1: lambda: 1})`, `task(5)`} {
		e2 := openStarted(t, baseManifest, "def good(ctx, a):\n    return 1\ndef on_query(ctx, q):\n    "+s+"\n    return []\n", nil, Options{})
		if _, err := e2.Query(ctx, ""); err == nil {
			t.Errorf("%s: want an error", s)
		}
		e2.Stop()
	}
}

func TestTaskLimitsAreSeparateFromHooks(t *testing.T) {
	script := `
def spin(ctx, n):
    for i in range(n):
        pass
    return n

def on_query(ctx, q):
    if q == "hook":
        spin(ctx, 5000)
        return []
    t = task("spin", 5000)
    return [item(id = "r", title = str(t.done) + str(t.error))]
`
	o := Options{Limits: Limits{MaxSteps: 1000, TaskMaxSteps: 1_000_000}}
	e, ch := taskExt(t, script, nil, o)
	ctx := context.Background()
	if _, err := e.Query(ctx, "hook"); err == nil {
		t.Fatal("the hook step limit must still apply")
	}
	if _, err := e.Query(ctx, "x"); err != nil {
		t.Fatal(err)
	}
	waitNotify(t, ch)
	items, _ := e.Query(ctx, "x")
	if len(items) != 1 || items[0].Title != "TrueNone" {
		t.Fatalf("%+v", items)
	}

	// A task that outruns TaskTimeout fails with a limit error.
	src := "def spin(ctx, n):\n    for i in range(100000000):\n        pass\ndef on_query(ctx, q):\n    t = task(\"spin\", 1)\n    return [item(id = \"r\", title = str(t.done) + str(t.error))]\n"
	e4 := openStarted(t, baseManifest, src, nil, Options{Limits: Limits{TaskTimeout: 100 * time.Millisecond, TaskMaxSteps: 1 << 40}})
	defer e4.Stop()
	ch4 := make(chan struct{}, 4)
	NewProvider(e4).OnReady(func() { ch4 <- struct{}{} })
	e4.Query(ctx, "")
	waitNotify(t, ch4)
	items, _ = e4.Query(ctx, "")
	if len(items) != 1 || !strings.HasPrefix(items[0].Title, "True") || !strings.Contains(items[0].Title, "cancelled") {
		t.Fatalf("%+v", items)
	}
}

func TestTaskAllowsLargerDownloads(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", 500)))
	}))
	defer srv.Close()
	dir := t.TempDir()
	dst := filepath.ToSlash(filepath.Join(dir, "f.bin"))
	script := `
def dl(ctx, a):
    return http.download("` + srv.URL + `/f", "` + dst + `").size

def on_query(ctx, q):
    if q == "hook":
        return [item(id = "h", title = str(http.download("` + srv.URL + `/f", "` + dst + `").size))]
    t = task("dl")
    return [item(id = "r", title = str(t.done) + str(t.value) + str(t.error))]
`
	perms := []Permission{"network:127.0.0.1", Permission("fs.write:" + dir)}
	e, ch := taskExt(t, script, perms, Options{Limits: Limits{MaxBulkBytes: 100, TaskBulkBytes: 1000}})
	ctx := context.Background()
	if _, err := e.Query(ctx, "hook"); err == nil || !strings.Contains(err.Error(), "larger than 100") {
		t.Fatalf("hook download: %v", err)
	}
	e.Query(ctx, "x")
	waitNotify(t, ch)
	items, _ := e.Query(ctx, "x")
	if len(items) != 1 || items[0].Title != "True500None" {
		t.Fatalf("%+v", items)
	}
}

func TestTaskLimitOnConcurrentRuns(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	script := `
def slow(ctx, n):
    return http.get("` + srv.URL + `", timeout = 20.0).status

def on_execute(ctx, it):
    for i in range(3):
        task("slow", i)
`
	e, _ := taskExt(t, script, []Permission{"network:127.0.0.1"}, Options{Limits: Limits{MaxTasks: 2}})
	err := e.Execute(context.Background(), core.Action{Kind: KindExtCall, Extra: map[string]string{"name": "default"}})
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("want a too-many-tasks error, got %v", err)
	}
	// started from on_execute: not tied to a query, so a query sweep keeps them
	e.sweepTasks(e.beginQuery())
	e.tasks.mu.Lock()
	n := len(e.tasks.byKey)
	e.tasks.mu.Unlock()
	if n != 2 {
		t.Fatalf("%d pinned tasks, want 2", n)
	}
}

func BenchmarkQueryWithTaskBookkeeping(b *testing.B) {
	e := &Extension{}
	for i := 0; i < b.N; i++ {
		e.sweepTasks(e.beginQuery())
	}
}
