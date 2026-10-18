package ext

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	fetchAllMax         = 100 // requests per call
	fetchAllConcurrency = 4   // default
	fetchAllCap         = 8   // largest max_concurrency
)

// biHTTPFetchAll runs many http.request calls at once. The Starlark thread
// parses every entry first, then blocks while goroutines send them, so the
// call's time limit covers the whole batch. Each entry goes through the same
// network check, redirect check and size caps as http.request.
func (e *Extension) biHTTPFetchAll(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var list starlark.Value
	conc := fetchAllConcurrency
	if err := starlark.UnpackArgs("http.fetch_all", args, kw, "requests", &list, "max_concurrency?", &conc); err != nil {
		return nil, err
	}
	if conc < 1 || conc > fetchAllCap {
		return nil, fmt.Errorf("http.fetch_all: max_concurrency must be 1 to %d", fetchAllCap)
	}
	seq, ok := list.(starlark.Indexable)
	if !ok {
		return nil, fmt.Errorf("http.fetch_all: requests must be a list, got %s", list.Type())
	}
	n := seq.Len()
	if n > fetchAllMax {
		return nil, fmt.Errorf("http.fetch_all: at most %d requests per call", fetchAllMax)
	}
	reqs := make([]httpReq, n)
	caps := make([]int64, n)
	for i := 0; i < n; i++ {
		q, mb, err := e.fetchEntry(th, i, seq.Index(i))
		if err != nil {
			return nil, err
		}
		reqs[i], caps[i] = q, mb
	}

	ctx := callCtx(th)
	results := make([]starlark.Value, n)
	var total atomic.Int64
	budget := 4 * e.opts.Limits.MaxHTTPBytes
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	for i := range reqs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results[i] = fetchError(reqs[i].url, fmt.Errorf("%v", r))
				}
			}()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[i] = fetchError(reqs[i].url, ctx.Err())
				return
			}
			r, err := e.sendCtx(ctx, reqs[i], caps[i])
			if err == nil && total.Add(int64(len(r.raw))) > budget {
				err = fmt.Errorf("responses of this call are larger than %d bytes together", budget)
			}
			if err != nil {
				results[i] = fetchError(reqs[i].url, err)
				return
			}
			results[i] = r.value()
		}()
	}
	wg.Wait()
	return starlark.NewList(results), nil
}

// fetchEntry turns one list entry (a URL string or a dict of http.request
// arguments with method defaulting to GET) into a request.
func (e *Extension) fetchEntry(th *starlark.Thread, i int, v starlark.Value) (httpReq, int64, error) {
	name := "http.fetch_all[" + strconv.Itoa(i) + "]"
	var kw []starlark.Tuple
	switch x := v.(type) {
	case starlark.String:
		kw = []starlark.Tuple{{starlark.String("method"), starlark.String("GET")}, {starlark.String("url"), x}}
	case *starlark.Dict:
		has := false
		for _, kv := range x.Items() {
			k, ok := starlark.AsString(kv[0])
			if !ok {
				return httpReq{}, 0, fmt.Errorf("%s: keys must be strings", name)
			}
			if k == "method" {
				has = true
			}
			kw = append(kw, starlark.Tuple{starlark.String(k), kv[1]})
		}
		if !has {
			kw = append(kw, starlark.Tuple{starlark.String("method"), starlark.String("GET")})
		}
	default:
		return httpReq{}, 0, fmt.Errorf("%s: want a URL string or a dict of http.request arguments, got %s", name, v.Type())
	}
	return e.parseRequest(th, name, nil, kw)
}

// fetchError is the entry for a request that failed. It has ok = False and
// status = 0 like a failed Response would, plus the message.
func fetchError(url string, err error) starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String("fetch_error"), starlark.StringDict{
		"ok": starlark.False, "status": starlark.MakeInt(0), "url": starlark.String(url), "error": starlark.String(err.Error()),
	})
}
