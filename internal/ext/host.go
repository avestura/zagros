package ext

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/update"
	starjson "go.starlark.net/lib/json"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type bfn = func(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error)

func (e *Extension) fn(name string, f bfn) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		return f(th, args, kw)
	})
}

func module(name string, members starlark.StringDict) *starlarkstruct.Module {
	return &starlarkstruct.Module{Name: name, Members: members}
}

func (e *Extension) ctxValue() starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String("ctx"), starlark.StringDict{
		"id":      starlark.String(e.Manifest.ID),
		"name":    starlark.String(e.Manifest.Name),
		"version": starlark.String(e.Manifest.Version),
		"dir":     starlark.String(e.Dir),
	})
}

// predeclared builds the host API visible to the extension's scripts.
func (e *Extension) predeclared() starlark.StringDict {
	settings := goToStarlark(e.Settings())
	pre := starlark.StringDict{
		"item":        e.fn("item", biItem),
		"scope":       e.fn("scope", biScope),
		"action_open": e.fn("action_open", biActionOpen),
		"action_copy": e.fn("action_copy", biActionCopy),
		"action_run":  e.fn("action_run", biActionRun),
		"action_call": e.fn("action_call", biActionCall),
		"notify":      e.fn("notify", e.biNotify),
		"settings":    settings,
		"json":        starjson.Module,
		"time":        startime.Module,
		"http":        e.httpModule(),
		"fs":          e.fsModule(),
		"clipboard":   module("clipboard", starlark.StringDict{"get": e.fn("clipboard.get", e.biClipGet), "set": e.fn("clipboard.set", e.biClipSet)}),
		"proc":        module("proc", starlark.StringDict{"run": e.fn("proc.run", e.biProcRun)}),
		"host":        module("host", starlark.StringDict{"version": e.fn("host.version", e.biHostVersion), "compare_versions": e.fn("host.compare_versions", biCompareVersions)}),
		"env":         module("env", starlark.StringDict{"get": e.fn("env.get", e.biEnvGet)}),
		"storage":     module("storage", starlark.StringDict{"get": e.fn("storage.get", e.biStorageGet), "set": e.fn("storage.set", e.biStorageSet), "delete": e.fn("storage.delete", e.biStorageDelete), "keys": e.fn("storage.keys", e.biStorageKeys)}),
		"secrets":     module("secrets", starlark.StringDict{"set": e.fn("secrets.set", e.biSecretSet), "get": e.fn("secrets.get", e.biSecretGet), "delete": e.fn("secrets.delete", e.biSecretDelete)}),
		"windows":     module("windows", starlark.StringDict{"registry_get": e.fn("windows.registry_get", e.biRegistryGet), "registry_keys": e.fn("windows.registry_keys", e.biRegistryKeys), "registry_values": e.fn("windows.registry_values", e.biRegistryValues), "services": e.fn("windows.services", e.biServices), "service_status": e.fn("windows.service_status", e.biServiceStatus), "list_windows": e.fn("windows.list_windows", e.biListWindows), "processes": e.fn("windows.processes", e.biProcesses), "known_folder": e.fn("windows.known_folder", e.biKnownFolder)}),
		"ai":          e.aiModule(),
		"compress":    e.compressModule(),
		"binary":      e.binaryModule(),
		"sqlite":      e.sqliteModule(),
		"tcp":         e.tcpModule(),
		"udp":         e.udpModule(),
		"tls":         e.tlsModule(),
		"dns":         e.dnsModule(),
		"ssh":         e.sshModule(),
		"sql":         e.sqlModule(),
		"ldap":        e.ldapModule(),
		"attempt":     starlark.NewBuiltin("attempt", biAttempt),
		"task":        e.fn("task", e.biTask),
		"crypto":      module("crypto", starlark.StringDict{"sha256": hashFn("crypto.sha256", sha256.New), "md5": hashFn("crypto.md5", md5.New), "hmac": starlark.NewBuiltin("crypto.hmac", biHMAC)}),
		"log":         module("log", starlark.StringDict{"info": e.logFn("info"), "warn": e.logFn("warn"), "error": e.logFn("error")}),
	}
	for k, v := range toolModules() { // text, lines, num, math, encode, hash, cipher, rand, uuid, regex, net, dt, extract
		pre[k] = v
	}
	pre.Freeze()
	return pre
}

// ---- conversions ----

func goToStarlark(v any) starlark.Value {
	switch x := v.(type) {
	case nil:
		return starlark.None
	case starlark.Value:
		return x
	case string:
		return starlark.String(x)
	case bool:
		return starlark.Bool(x)
	case int:
		return starlark.MakeInt(x)
	case int64:
		return starlark.MakeInt64(x)
	case float64:
		return starlark.Float(x)
	case []string:
		l := make([]starlark.Value, len(x))
		for i, s := range x {
			l[i] = starlark.String(s)
		}
		return starlark.NewList(l)
	case []any:
		l := make([]starlark.Value, len(x))
		for i, s := range x {
			l[i] = goToStarlark(s)
		}
		return starlark.NewList(l)
	case map[string]any:
		d := starlark.NewDict(len(x))
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			_ = d.SetKey(starlark.String(k), goToStarlark(x[k]))
		}
		return d
	}
	return starlark.String(fmt.Sprint(v))
}

func stringMap(d *starlark.Dict) (map[string]string, error) {
	out := map[string]string{}
	if d == nil {
		return out, nil
	}
	for _, kv := range d.Items() {
		k, ok1 := starlark.AsString(kv[0])
		v, ok2 := starlark.AsString(kv[1])
		if !ok1 || !ok2 {
			return nil, errors.New("expected a dict of string to string")
		}
		out[k] = v
	}
	return out, nil
}

func stringList(l starlark.Value) ([]string, error) {
	if l == nil || l == starlark.None {
		return nil, nil
	}
	it, ok := l.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("expected a list of strings, got %s", l.Type())
	}
	iter := it.Iterate()
	defer iter.Done()
	var out []string
	var v starlark.Value
	for iter.Next(&v) {
		s, ok := starlark.AsString(v)
		if !ok {
			return nil, fmt.Errorf("expected string in list, got %s", v.Type())
		}
		out = append(out, s)
	}
	return out, nil
}

func jsonCall(th *starlark.Thread, name string, arg starlark.Value) (starlark.Value, error) {
	return starlark.Call(th, starjson.Module.Members[name], starlark.Tuple{arg}, nil)
}

// ---- log / notify ----

func (e *Extension) logFn(level string) *starlark.Builtin {
	return starlark.NewBuiltin("log."+level, func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		parts := make([]string, len(args))
		for i, a := range args {
			if s, ok := starlark.AsString(a); ok {
				parts[i] = s
			} else {
				parts[i] = a.String()
			}
		}
		e.opts.Log(e.Manifest.ID, level, strings.Join(parts, " "))
		return starlark.None, nil
	})
}

func (e *Extension) biNotify(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var title, body string
	if err := starlark.UnpackArgs("notify", args, kw, "title", &title, "body?", &body); err != nil {
		return nil, err
	}
	if err := e.require(PermNotifications, ""); err != nil {
		return nil, err
	}
	return starlark.None, e.opts.Platform.Notify(title, body)
}

// ---- clipboard / env ----

func (e *Extension) biClipGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("clipboard.get", args, kw); err != nil {
		return nil, err
	}
	if err := e.require(PermClipboardRead, ""); err != nil {
		return nil, err
	}
	s, err := e.opts.Platform.ClipboardGet()
	return starlark.String(s), err
}

func (e *Extension) biClipSet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs("clipboard.set", args, kw, "text", &s); err != nil {
		return nil, err
	}
	if err := e.require(PermClipboardWrite, ""); err != nil {
		return nil, err
	}
	return starlark.None, e.opts.Platform.ClipboardSet(s)
}

func (e *Extension) biEnvGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name string
	var def starlark.Value = starlark.None
	if err := starlark.UnpackArgs("env.get", args, kw, "name", &name, "default?", &def); err != nil {
		return nil, err
	}
	if err := e.require(PermEnvRead, name); err != nil {
		return nil, err
	}
	if v, ok := os.LookupEnv(name); ok {
		return starlark.String(v), nil
	}
	return def, nil
}

// ---- storage ----

func (e *Extension) biStorageGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	var def starlark.Value = starlark.None
	if err := starlark.UnpackArgs("storage.get", args, kw, "key", &key, "default?", &def); err != nil {
		return nil, err
	}
	raw, ok := e.kv.get(key)
	if !ok {
		return def, nil
	}
	return jsonCall(th, "decode", starlark.String(raw))
}

func (e *Extension) biStorageSet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	var val starlark.Value
	if err := starlark.UnpackArgs("storage.set", args, kw, "key", &key, "value", &val); err != nil {
		return nil, err
	}
	enc, err := jsonCall(th, "encode", val)
	if err != nil {
		return nil, err
	}
	return starlark.None, e.kv.set(key, []byte(string(enc.(starlark.String))))
}

func (e *Extension) biStorageDelete(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	if err := starlark.UnpackArgs("storage.delete", args, kw, "key", &key); err != nil {
		return nil, err
	}
	return starlark.None, e.kv.delete(key)
}

func (e *Extension) biStorageKeys(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	k := e.kv.keys()
	sort.Strings(k)
	return goToStarlark(k), nil
}

// ---- crypto ----

func asBytes(v starlark.Value) ([]byte, error) {
	switch x := v.(type) {
	case starlark.String:
		return []byte(x), nil
	case starlark.Bytes:
		return []byte(x), nil
	}
	return nil, fmt.Errorf("expected string or bytes, got %s", v.Type())
}

func hashFn(name string, h func() hash.Hash) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		var v starlark.Value
		if err := starlark.UnpackArgs(name, args, kw, "data", &v); err != nil {
			return nil, err
		}
		data, err := asBytes(v)
		if err != nil {
			return nil, err
		}
		x := h()
		x.Write(data)
		return starlark.String(hex.EncodeToString(x.Sum(nil))), nil
	})
}

func biHMAC(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var k, d starlark.Value
	algo := "sha256"
	if err := starlark.UnpackArgs("crypto.hmac", args, kw, "key", &k, "data", &d, "algo?", &algo); err != nil {
		return nil, err
	}
	key, err := asBytes(k)
	if err != nil {
		return nil, err
	}
	data, err := asBytes(d)
	if err != nil {
		return nil, err
	}
	var h func() hash.Hash
	switch algo {
	case "sha256":
		h = sha256.New
	case "sha1":
		h = sha1.New
	case "sha512":
		h = sha512.New
	case "md5":
		h = md5.New
	default:
		return nil, fmt.Errorf("unsupported algo %q (sha256, sha1, sha512, md5)", algo)
	}
	m := hmac.New(h, key)
	m.Write(data)
	return starlark.String(hex.EncodeToString(m.Sum(nil))), nil
}

// ---- filesystem ----

// resolvePath expands ~, makes the path absolute and resolves symlinks of the
// longest existing prefix so scope checks cannot be bypassed with links.
func resolvePath(p string) (string, error) {
	p = expandHome(p)
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	cur, rest := abs, ""
	for {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(r, rest), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs, nil
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}

func (e *Extension) checkFS(kind, p string) (string, error) {
	r, err := resolvePath(p)
	if err != nil {
		return "", err
	}
	if kind == PermFSRead {
		own, _ := resolvePath(e.Dir)
		if pathWithin(own, r) {
			return r, nil
		}
	}
	return r, e.require(kind, r)
}

func (e *Extension) biFSRead(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("fs.read_text", args, kw, "path", &p); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(r)
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
	return starlark.String(b), nil
}

func (e *Extension) biFSWrite(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p, text string
	if err := starlark.UnpackArgs("fs.write_text", args, kw, "path", &p, "text", &text); err != nil {
		return nil, err
	}
	if err := e.capWrite(len(text)); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	return starlark.None, os.WriteFile(r, []byte(text), 0o644)
}

func (e *Extension) biFSList(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("fs.list", args, kw, "path", &p); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	ents, err := os.ReadDir(r)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(ents))
	for _, d := range ents {
		n := d.Name()
		if d.IsDir() {
			n += "/"
		}
		out = append(out, n)
	}
	return goToStarlark(out), nil
}

func (e *Extension) biFSExists(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("fs.exists", args, kw, "path", &p); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(r)
	return starlark.Bool(err == nil), nil
}

// ---- process ----

type capBuf struct {
	bytes.Buffer
	max int
}

func (c *capBuf) Write(p []byte) (int, error) {
	if c.Len()+len(p) > c.max {
		p = p[:max(0, c.max-c.Len())]
	}
	c.Buffer.Write(p)
	return len(p), nil
}

func (e *Extension) biProcRun(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var cmd, stdin, cwd string
	var argv starlark.Value = starlark.None
	timeout := secs(10)
	if err := starlark.UnpackArgs("proc.run", args, kw, "cmd", &cmd, "args?", &argv, "timeout?", &timeout, "stdin?", &stdin, "cwd?", &cwd); err != nil {
		return nil, err
	}
	list, err := stringList(argv)
	if err != nil {
		return nil, err
	}
	target := cmd
	if full, err := exec.LookPath(cmd); err == nil {
		target = full
	}
	if err := e.require(PermProcessRun, target); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(callCtx(th), time.Duration(float64(timeout)*float64(time.Second)))
	defer cancel()
	c := exec.CommandContext(ctx, target, list...)
	c.Dir = cwd
	hideWindow(c)
	if stdin != "" {
		c.Stdin = strings.NewReader(stdin)
	}
	out, errb := &capBuf{max: 1 << 20}, &capBuf{max: 1 << 20}
	c.Stdout, c.Stderr = out, errb
	runErr := c.Run()
	code := 0
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			code = ee.ExitCode()
		} else {
			return nil, runErr
		}
	}
	return starlarkstruct.FromStringDict(starlark.String("result"), starlark.StringDict{
		"code": starlark.MakeInt(code), "ok": starlark.Bool(code == 0),
		"stdout": starlark.String(out.String()), "stderr": starlark.String(errb.String()),
	}), nil
}

// ---- http: see hosthttp.go ----

func biQuote(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs("http.quote", args, kw, "text", &s); err != nil {
		return nil, err
	}
	return starlark.String(url.QueryEscape(s)), nil
}

// host.version() returns the running Zagros version ("dev" for unversioned builds).
func (e *Extension) biHostVersion(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("host.version", args, kw); err != nil {
		return nil, err
	}
	if e.opts.AppVersion == "" {
		return starlark.String("dev"), nil
	}
	return starlark.String(e.opts.AppVersion), nil
}

// host.compare_versions(a, b) compares two semantic versions: -1, 0 or 1.
func biCompareVersions(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var a, b string
	if err := starlark.UnpackArgs("host.compare_versions", args, kw, "a", &a, "b", &b); err != nil {
		return nil, err
	}
	c, ok := update.CompareVersions(a, b)
	if !ok {
		return nil, fmt.Errorf("host.compare_versions: %q and %q must be semantic versions like 1.2.3 or v1.2.3-rc.1", a, b)
	}
	return starlark.MakeInt(c), nil
}
