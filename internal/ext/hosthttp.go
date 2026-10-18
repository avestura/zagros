package ext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// hostState holds per-extension runtime state of the host modules.
type hostState struct {
	mu  sync.Mutex
	jar *cookiejar.Jar
	dbs map[string]*sqliteDB
	ai  aiUsage
	net netState

	sqls  map[*sqliteDB]bool // sql.open connections
	ldaps map[*ldapConn]bool // ldap.connect connections
}

func (e *Extension) cookieJar() *cookiejar.Jar {
	e.host.mu.Lock()
	defer e.host.mu.Unlock()
	if e.host.jar == nil {
		e.host.jar, _ = cookiejar.New(nil)
	}
	return e.host.jar
}

func (e *Extension) httpModule() *starlarkstruct.Module {
	m := starlark.StringDict{
		"get":           e.fn("http.get", e.biHTTPGet),
		"post":          e.fn("http.post", e.biHTTPPost),
		"put":           e.fn("http.put", e.bodyMethod("PUT")),
		"patch":         e.fn("http.patch", e.bodyMethod("PATCH")),
		"delete":        e.fn("http.delete", e.bodyMethod("DELETE")),
		"head":          e.fn("http.head", e.biHTTPHead),
		"request":       e.fn("http.request", e.biHTTPRequest),
		"download":      e.fn("http.download", e.biHTTPDownload),
		"clear_cookies": e.fn("http.clear_cookies", e.biHTTPClearCookies),
		"fetch_all":     e.fn("http.fetch_all", e.biHTTPFetchAll),
		"json":          e.fn("http.json", e.biHTTPJSON),
		"quote":         e.fn("http.quote", biQuote),
	}
	return module("http", m)
}

func (e *Extension) client() *http.Client {
	c := *e.opts.HTTPClient
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("too many redirects")
		}
		return e.require(PermNetwork, req.URL.Hostname())
	}
	return &c
}

type httpResult struct {
	status int
	raw    []byte
	hdr    http.Header
	url    string
}

func (r *httpResult) value() starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String("response"), starlark.StringDict{
		"status":      starlark.MakeInt(r.status),
		"ok":          starlark.Bool(r.status >= 200 && r.status < 300),
		"url":         starlark.String(r.url),
		"headers":     headerDict(r.hdr, false),
		"raw_headers": headerDict(r.hdr, true),
		"text":        starlark.String(r.raw),
		"bytes":       starlark.Bytes(r.raw),
		"json": starlark.NewBuiltin("response.json", func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
			return jsonCall(th, "decode", starlark.String(r.raw))
		}),
	})
}

func headerDict(h http.Header, all bool) *starlark.Dict {
	d := starlark.NewDict(len(h))
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := h[k]
		if len(v) == 0 {
			continue
		}
		var val starlark.Value = starlark.String(v[0])
		if all {
			l := make([]starlark.Value, len(v))
			for i, s := range v {
				l[i] = starlark.String(s)
			}
			val = starlark.NewList(l)
		}
		_ = d.SetKey(starlark.String(strings.ToLower(k)), val)
	}
	return d
}

type httpReq struct {
	method     string
	url        string
	headers    map[string]string
	body       io.Reader
	timeout    float64
	noRedirect bool
	cookies    bool
	auth       starlark.Value
}

// roundTrip sends the request after the network check. The caller must call
// done once it has finished with the response body.
func (e *Extension) roundTrip(th *starlark.Thread, q httpReq) (*http.Response, func(), error) {
	return e.roundTripCtx(callCtx(th), q)
}

func (e *Extension) roundTripCtx(parent context.Context, q httpReq) (resp *http.Response, done func(), err error) {
	u, err := url.Parse(q.url)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, nil, fmt.Errorf("invalid url %q (need http:// or https://)", q.url)
	}
	if err := e.require(PermNetwork, u.Hostname()); err != nil {
		return nil, nil, err
	}
	if q.timeout <= 0 {
		return nil, nil, errors.New("timeout must be positive")
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(q.timeout*float64(time.Second)))
	req, err := http.NewRequestWithContext(ctx, q.method, q.url, q.body)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Zagros/"+e.Manifest.ID+"/"+e.Manifest.Version)
	for k, v := range q.headers {
		req.Header.Set(k, v)
	}
	if err := setAuth(req, q.auth); err != nil {
		cancel()
		return nil, nil, err
	}
	c := e.client()
	if q.cookies {
		c.Jar = e.cookieJar()
	}
	if q.noRedirect {
		c.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	}
	resp, err = c.Do(req)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return resp, func() { resp.Body.Close(); cancel() }, nil
}

func setAuth(req *http.Request, auth starlark.Value) error {
	if auth == nil || auth == starlark.None || req.Header.Get("Authorization") != "" {
		return nil
	}
	switch a := auth.(type) {
	case starlark.String:
		req.Header.Set("Authorization", "Bearer "+string(a))
	case starlark.Tuple:
		if len(a) != 2 {
			return errors.New("auth tuple must be (user, password)")
		}
		u, ok1 := starlark.AsString(a[0])
		p, ok2 := starlark.AsString(a[1])
		if !ok1 || !ok2 {
			return errors.New("auth tuple must hold strings")
		}
		req.SetBasicAuth(u, p)
	default:
		return fmt.Errorf("auth must be a token string or a (user, password) tuple, got %s", auth.Type())
	}
	return nil
}

func (e *Extension) readCapped(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("response larger than %d bytes", limit)
	}
	return b, nil
}

func (e *Extension) send(th *starlark.Thread, q httpReq, maxBytes int64) (*httpResult, error) {
	return e.sendCtx(callCtx(th), q, maxBytes)
}

func (e *Extension) sendCtx(ctx context.Context, q httpReq, maxBytes int64) (*httpResult, error) {
	resp, done, err := e.roundTripCtx(ctx, q)
	if err != nil {
		return nil, err
	}
	defer done()
	limit := e.opts.Limits.MaxHTTPBytes
	if maxBytes > 0 && maxBytes < limit {
		limit = maxBytes
	}
	b, err := e.readCapped(resp.Body, limit)
	if err != nil {
		return nil, err
	}
	return &httpResult{status: resp.StatusCode, raw: b, hdr: resp.Header, url: resp.Request.URL.String()}, nil
}

func (e *Extension) doHTTP(th *starlark.Thread, method, rawurl string, body io.Reader, headers map[string]string, timeout float64) (*httpResult, error) {
	return e.send(th, httpReq{method: method, url: rawurl, headers: headers, body: body, timeout: timeout}, 0)
}

func (e *Extension) biHTTPGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var u string
	var hd *starlark.Dict
	timeout := secs(10)
	if err := starlark.UnpackArgs("http.get", args, kw, "url", &u, "headers?", &hd, "timeout?", &timeout); err != nil {
		return nil, err
	}
	h, err := stringMap(hd)
	if err != nil {
		return nil, err
	}
	r, err := e.doHTTP(th, "GET", u, nil, h, float64(timeout))
	if err != nil {
		return nil, err
	}
	return r.value(), nil
}

func (e *Extension) biHTTPHead(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var u string
	var hd *starlark.Dict
	timeout := secs(10)
	if err := starlark.UnpackArgs("http.head", args, kw, "url", &u, "headers?", &hd, "timeout?", &timeout); err != nil {
		return nil, err
	}
	h, err := stringMap(hd)
	if err != nil {
		return nil, err
	}
	r, err := e.doHTTP(th, "HEAD", u, nil, h, float64(timeout))
	if err != nil {
		return nil, err
	}
	return r.value(), nil
}

func (e *Extension) biHTTPJSON(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var u string
	var hd *starlark.Dict
	timeout := secs(10)
	if err := starlark.UnpackArgs("http.json", args, kw, "url", &u, "headers?", &hd, "timeout?", &timeout); err != nil {
		return nil, err
	}
	h, err := stringMap(hd)
	if err != nil {
		return nil, err
	}
	if _, ok := h["Accept"]; !ok {
		h["Accept"] = "application/json"
	}
	r, err := e.doHTTP(th, "GET", u, nil, h, float64(timeout))
	if err != nil {
		return nil, err
	}
	if r.status < 200 || r.status > 299 {
		return nil, fmt.Errorf("http.json %s: status %d", u, r.status)
	}
	return jsonCall(th, "decode", starlark.String(r.raw))
}

func (e *Extension) biHTTPPost(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	return e.bodyMethod("POST")(th, args, kw)
}

// bodyMethod builds post, put, patch and delete: url, body, json, headers, timeout.
func (e *Extension) bodyMethod(method string) bfn {
	name := "http." + strings.ToLower(method)
	return func(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		var u string
		var body starlark.Value = starlark.String("")
		var hd *starlark.Dict
		var js starlark.Value = starlark.None
		timeout := secs(10)
		if err := starlark.UnpackArgs(name, args, kw, "url", &u, "body?", &body, "json?", &js, "headers?", &hd, "timeout?", &timeout); err != nil {
			return nil, err
		}
		h, err := stringMap(hd)
		if err != nil {
			return nil, err
		}
		rd, err := e.bodyReader(th, h, body, js, nil, nil)
		if err != nil {
			return nil, err
		}
		r, err := e.doHTTP(th, method, u, rd, h, float64(timeout))
		if err != nil {
			return nil, err
		}
		return r.value(), nil
	}
}

func hasHeader(h map[string]string, name string) bool {
	for k := range h {
		if strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}

// bodyReader turns the body, json, form and multipart arguments (at most one)
// into a request body and sets the matching Content-Type.
func (e *Extension) bodyReader(th *starlark.Thread, h map[string]string, body, js starlark.Value, form, mp *starlark.Dict) (io.Reader, error) {
	n := 0
	if body != nil && body != starlark.None && body != starlark.String("") {
		n++
	}
	if js != nil && js != starlark.None {
		n++
	}
	if form != nil {
		n++
	}
	if mp != nil {
		n++
	}
	if n > 1 {
		return nil, errors.New("give only one of body, json, form and multipart")
	}
	setType := func(t string) {
		if !hasHeader(h, "Content-Type") {
			h["Content-Type"] = t
		}
	}
	limit := e.opts.Limits.MaxHTTPBytes
	switch {
	case js != nil && js != starlark.None:
		enc, err := jsonCall(th, "encode", js)
		if err != nil {
			return nil, err
		}
		setType("application/json")
		return strings.NewReader(string(enc.(starlark.String))), nil
	case form != nil:
		vals := url.Values{}
		for _, kv := range form.Items() {
			k, ok1 := starlark.AsString(kv[0])
			v, ok2 := starlark.AsString(kv[1])
			if !ok1 || !ok2 {
				return nil, errors.New("form must be a dict of string to string")
			}
			vals.Set(k, v)
		}
		setType("application/x-www-form-urlencoded")
		return strings.NewReader(vals.Encode()), nil
	case mp != nil:
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		items := mp.Items()
		sort.Slice(items, func(i, j int) bool { return items[i][0].String() < items[j][0].String() })
		for _, kv := range items {
			name, ok := starlark.AsString(kv[0])
			if !ok {
				return nil, errors.New("multipart keys must be strings")
			}
			if err := e.addPart(w, name, kv[1]); err != nil {
				return nil, err
			}
			if int64(buf.Len()) > limit {
				return nil, fmt.Errorf("request body larger than %d bytes", limit)
			}
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		setType(w.FormDataContentType())
		return &buf, nil
	}
	if body == nil || body == starlark.None {
		return nil, nil
	}
	b, err := asBytes(body)
	if err != nil {
		return nil, fmt.Errorf("body: %w", err)
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("request body larger than %d bytes", limit)
	}
	if len(b) == 0 {
		return nil, nil
	}
	return bytes.NewReader(b), nil
}

var quoteEsc = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// addPart adds a multipart field: a string is a text field, bytes or a dict
// {"filename", "content" | "path", "content_type"} is a file part.
func (e *Extension) addPart(w *multipart.Writer, name string, v starlark.Value) error {
	if s, ok := v.(starlark.String); ok {
		return w.WriteField(name, string(s))
	}
	var filename, ctype string
	var content []byte
	switch x := v.(type) {
	case starlark.Bytes:
		filename, content = name, []byte(x)
	case *starlark.Dict:
		get := func(k string) starlark.Value {
			val, ok, _ := x.Get(starlark.String(k))
			if !ok {
				return nil
			}
			return val
		}
		if s, ok := starlark.AsString(get("filename")); ok {
			filename = s
		}
		if s, ok := starlark.AsString(get("content_type")); ok {
			ctype = s
		}
		if p, ok := starlark.AsString(get("path")); ok {
			r, err := e.checkFS(PermFSRead, p)
			if err != nil {
				return err
			}
			b, err := e.readFileCapped(r)
			if err != nil {
				return err
			}
			content = b
			if filename == "" {
				filename = filepath.Base(r)
			}
		} else if c := get("content"); c != nil {
			b, err := asBytes(c)
			if err != nil {
				return fmt.Errorf("multipart %q content: %w", name, err)
			}
			content = b
		} else {
			return fmt.Errorf("multipart %q: file dict needs content or path", name)
		}
		if filename == "" {
			filename = name
		}
	default:
		return fmt.Errorf("multipart %q: want str, bytes or a file dict, got %s", name, v.Type())
	}
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, quoteEsc.Replace(name), quoteEsc.Replace(filename)))
	h.Set("Content-Type", ctype)
	pw, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = pw.Write(content)
	return err
}

func (e *Extension) readFileCapped(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, e.opts.Limits.MaxFileBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > e.opts.Limits.MaxFileBytes {
		return nil, fmt.Errorf("file larger than %d bytes", e.opts.Limits.MaxFileBytes)
	}
	return b, nil
}

func (e *Extension) biHTTPRequest(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	q, maxBytes, err := e.parseRequest(th, "http.request", args, kw)
	if err != nil {
		return nil, err
	}
	r, err := e.send(th, q, maxBytes)
	if err != nil {
		return nil, err
	}
	return r.value(), nil
}

// parseRequest reads the http.request arguments into a request. It runs on
// the Starlark thread: bodies are built here, before any goroutine starts.
func (e *Extension) parseRequest(th *starlark.Thread, fname string, args starlark.Tuple, kw []starlark.Tuple) (httpReq, int64, error) {
	var method, u string
	var body, js, auth starlark.Value = starlark.None, starlark.None, starlark.None
	var hd, form, mp *starlark.Dict
	timeout := secs(10)
	follow, cookies := true, false
	var maxBytes int64
	if err := starlark.UnpackArgs(fname, args, kw, "method", &method, "url", &u, "headers?", &hd, "body?", &body, "json?", &js,
		"form?", &form, "multipart?", &mp, "timeout?", &timeout, "follow_redirects?", &follow, "auth?", &auth, "cookies?", &cookies, "max_bytes?", &maxBytes); err != nil {
		return httpReq{}, 0, err
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" || strings.ContainsAny(method, " \t\r\n") {
		return httpReq{}, 0, fmt.Errorf("invalid method %q", method)
	}
	h, err := stringMap(hd)
	if err != nil {
		return httpReq{}, 0, err
	}
	rd, err := e.bodyReader(th, h, body, js, form, mp)
	if err != nil {
		return httpReq{}, 0, err
	}
	return httpReq{method: method, url: u, headers: h, body: rd, timeout: float64(timeout), noRedirect: !follow, cookies: cookies, auth: auth}, maxBytes, nil
}

func (e *Extension) biHTTPClearCookies(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("http.clear_cookies", args, kw); err != nil {
		return nil, err
	}
	e.host.mu.Lock()
	e.host.jar = nil
	e.host.mu.Unlock()
	return starlark.None, nil
}

func (e *Extension) biHTTPDownload(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var u, p string
	var hd *starlark.Dict
	var auth starlark.Value = starlark.None
	timeout := secs(60)
	cookies, overwrite := false, true
	var maxBytes int64
	if err := starlark.UnpackArgs("http.download", args, kw, "url", &u, "path", &p, "headers?", &hd, "timeout?", &timeout,
		"auth?", &auth, "cookies?", &cookies, "overwrite?", &overwrite, "max_bytes?", &maxBytes); err != nil {
		return nil, err
	}
	h, err := stringMap(hd)
	if err != nil {
		return nil, err
	}
	dst, err := e.checkFS(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	if !overwrite {
		if _, err := os.Lstat(dst); err == nil {
			return nil, fmt.Errorf("%s already exists", p)
		}
	}
	resp, done, err := e.roundTrip(th, httpReq{method: "GET", url: u, headers: h, timeout: float64(timeout), cookies: cookies, auth: auth})
	if err != nil {
		return nil, err
	}
	defer done()
	res := starlark.StringDict{
		"status":  starlark.MakeInt(resp.StatusCode),
		"ok":      starlark.Bool(resp.StatusCode >= 200 && resp.StatusCode < 300),
		"url":     starlark.String(resp.Request.URL.String()),
		"headers": headerDict(resp.Header, false),
		"path":    starlark.None,
		"size":    starlark.MakeInt(0),
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return starlarkstruct.FromStringDict(starlark.String("download"), res), nil
	}
	limit := e.bulkLimit(th)
	if maxBytes > 0 && maxBytes < limit {
		limit = maxBytes
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, err
	}
	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(f, io.LimitReader(resp.Body, limit+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > limit {
		err = fmt.Errorf("download larger than %d bytes", limit)
	}
	if err == nil {
		err = os.Rename(tmp, dst)
	}
	if err != nil {
		os.Remove(tmp)
		return nil, err
	}
	res["path"] = starlark.String(filepath.ToSlash(dst))
	res["size"] = starlark.MakeInt64(n)
	return starlarkstruct.FromStringDict(starlark.String("download"), res), nil
}
