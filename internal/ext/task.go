package ext

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// task(fn, args, ttl) runs a top-level function of the extension in its own
// Starlark thread with the longer task limits and returns a handle at once.
// The hook that asked keeps its short limit. When the run ends the host asks
// the launcher to run the query again (Provider.OnReady), and the same
// task(...) call then returns the finished handle.

const (
	taskThreadKey  = "zagros.task" // thread-local: this thread is a task run
	taskKeep       = 64            // entries kept, running or finished
	taskDefaultTTL = 30.0
	taskMinTTL     = 2.0
	taskMaxTTL     = 600.0
	taskNotifyWait = 30 * time.Millisecond
)

type genKey struct{}

type taskState struct {
	mu     sync.Mutex
	gen    uint64
	seq    int
	byKey  map[string]*task
	notify func()
	timer  *time.Timer
}

type task struct {
	id       string
	key      string
	gen      uint64 // newest query that asked for it
	pinned   bool   // asked outside a query hook: never swept
	cancel   context.CancelFunc
	done     bool
	value    starlark.Value
	err      string
	ttl      time.Duration
	finished time.Time
}

func (e *Extension) setTaskNotify(fn func()) {
	e.tasks.mu.Lock()
	e.tasks.notify = fn
	e.tasks.mu.Unlock()
}

// beginQuery numbers a query hook call; sweepTasks(gen) cancels the running
// tasks that no query at or after gen asked for.
func (e *Extension) beginQuery() uint64 {
	e.tasks.mu.Lock()
	defer e.tasks.mu.Unlock()
	e.tasks.gen++
	return e.tasks.gen
}

func (e *Extension) sweepTasks(gen uint64) {
	ts := &e.tasks
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for k, t := range ts.byKey {
		if !t.done && !t.pinned && t.gen < gen {
			t.cancel()
			delete(ts.byKey, k)
		}
	}
}

// cancelTasks stops every task and forgets the results (Stop, new settings).
func (e *Extension) cancelTasks() {
	ts := &e.tasks
	ts.mu.Lock()
	defer ts.mu.Unlock()
	for _, t := range ts.byKey {
		if !t.done {
			t.cancel()
		}
	}
	ts.byKey = nil
}

func (ts *taskState) ping() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.notify == nil {
		return
	}
	if ts.timer == nil {
		ts.timer = time.AfterFunc(taskNotifyWait, func() {
			ts.mu.Lock()
			fn := ts.notify
			ts.mu.Unlock()
			if fn != nil {
				fn()
			}
		})
		return
	}
	ts.timer.Reset(taskNotifyWait)
}

// bulkLimit is the byte cap for downloads and archives: larger in a task.
func (e *Extension) bulkLimit(th *starlark.Thread) int64 {
	l := e.opts.Limits
	if th.Local(taskThreadKey) != nil && l.TaskBulkBytes > l.MaxBulkBytes {
		return l.TaskBulkBytes
	}
	return l.MaxBulkBytes
}

func (e *Extension) biTask(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name string
	var argv starlark.Value = starlark.None
	ttl := taskDefaultTTL
	if err := starlark.UnpackArgs("task", args, kw, "fn", &name, "args?", &argv, "ttl?", &ttl); err != nil {
		return nil, err
	}
	if th.Local(taskThreadKey) != nil {
		return nil, errors.New("task: a task cannot start another task")
	}
	if ttl < taskMinTTL || ttl > taskMaxTTL {
		return nil, fmt.Errorf("task: ttl must be %g to %g seconds", taskMinTTL, taskMaxTTL)
	}
	rt := e.runtime()
	if rt == nil {
		return nil, errors.New("task: extension is not running")
	}
	fn, ok := rt.globals[name].(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("task: no top-level function %q (name it as a string)", name)
	}
	enc := "null"
	if argv != starlark.None {
		v, err := jsonCall(th, "encode", argv)
		if err != nil {
			return nil, fmt.Errorf("task: args must be JSON values: %w", err)
		}
		enc = string(v.(starlark.String))
	}
	key := name + "\x00" + enc
	gen, inQuery := callCtx(th).Value(genKey{}).(uint64)

	ts := &e.tasks
	ts.mu.Lock()
	t := ts.byKey[key]
	if t != nil && t.done && time.Since(t.finished) > t.ttl {
		delete(ts.byKey, key)
		t = nil
	}
	if t == nil {
		if err := e.admitTaskLocked(); err != nil {
			ts.mu.Unlock()
			return nil, err
		}
		if ts.byKey == nil {
			ts.byKey = map[string]*task{}
		}
		ts.seq++
		ctx, cancel := context.WithCancel(context.Background())
		t = &task{id: fmt.Sprintf("%s#%d", name, ts.seq), key: key, cancel: cancel, ttl: time.Duration(ttl * float64(time.Second))}
		ts.byKey[key] = t
		go e.runTask(ctx, t, rt, name, fn, enc)
	}
	if !t.done {
		if !inQuery {
			t.pinned = true
		} else if gen > t.gen {
			t.gen = gen
		}
	}
	h := taskHandle(t)
	ts.mu.Unlock()
	return h, nil
}

// admitTaskLocked makes room for one more run: it drops expired or the
// oldest finished entries, and refuses when too many are still running.
func (e *Extension) admitTaskLocked() error {
	ts := &e.tasks
	running := 0
	var fin []*task
	for k, t := range ts.byKey {
		if !t.done {
			running++
		} else if time.Since(t.finished) > t.ttl {
			delete(ts.byKey, k)
		} else {
			fin = append(fin, t)
		}
	}
	if running >= e.opts.Limits.MaxTasks {
		return fmt.Errorf("task: %d tasks are already running", running)
	}
	if over := len(ts.byKey) + 1 - taskKeep; over > 0 {
		sort.Slice(fin, func(i, j int) bool { return fin[i].finished.Before(fin[j].finished) })
		for i := 0; i < over && i < len(fin); i++ {
			delete(ts.byKey, fin[i].key)
		}
	}
	return nil
}

func taskHandle(t *task) starlark.Value {
	var val starlark.Value = starlark.None
	if t.done && t.value != nil {
		val = t.value
	}
	var errv starlark.Value = starlark.None
	if t.err != "" {
		errv = starlark.String(t.err)
	}
	return starlarkstruct.FromStringDict(starlark.String("task"), starlark.StringDict{
		"id": starlark.String(t.id), "done": starlark.Bool(t.done),
		"ok": starlark.Bool(t.done && t.err == ""), "value": val, "error": errv,
	})
}

func (e *Extension) runTask(ctx context.Context, t *task, rt *vm, name string, fn starlark.Callable, enc string) {
	var res starlark.Value
	var runErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				runErr = fmt.Errorf("task %s panicked: %v", name, r)
			}
		}()
		l := e.opts.Limits
		th := e.newThread(ctx, "task:"+name, nil)
		th.SetMaxExecutionSteps(l.TaskMaxSteps)
		th.SetLocal(taskThreadKey, true)
		runErr = e.withTimeout(ctx, th, l.TaskTimeout, func() error {
			var a starlark.Value = starlark.None
			if enc != "null" {
				v, err := jsonCall(th, "decode", starlark.String(enc))
				if err != nil {
					return err
				}
				a = v
			}
			var err error
			res, err = starlark.Call(th, fn, starlark.Tuple{rt.ctxVal, a}, nil)
			return err
		})
		if runErr == nil && res != nil {
			res.Freeze()
		}
	}()

	ts := &e.tasks
	ts.mu.Lock()
	if ctx.Err() != nil || ts.byKey[t.key] != t { // cancelled or swept: nobody waits for it
		ts.mu.Unlock()
		return
	}
	t.done, t.finished = true, time.Now()
	if runErr != nil {
		t.err = formatErr(runErr).Error()
		e.opts.Log(e.Manifest.ID, "warn", "task "+name+": "+t.err)
	} else {
		t.value = res
	}
	t.cancel()
	ts.mu.Unlock()
	ts.ping()
}
