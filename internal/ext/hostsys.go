package ext

import (
	"errors"
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// secretBackend stores secrets by target name; the OS file picks the
// implementation (Credential Manager on Windows).
var secretBackend interface {
	Set(target, value string) error
	Get(target string) (string, bool, error)
	Delete(target string) error
}

var errSecretsUnsupported = errors.New("secrets are only available on Windows")

const (
	maxSecretKey   = 256
	maxSecretValue = 2560 // Credential Manager blob limit
)

type sysProc struct {
	Name string
	PID  int
}

type sysWindow struct {
	Title   string
	Process string
	PID     int
}

type sysService struct {
	Name, Display, State, StartType string
	PID                             int
}

// secretTarget namespaces a key by extension id. Ids never contain "|", so
// the first two separators are unambiguous and one extension cannot name
// another's entry.
func secretTarget(extID, key string) string { return "zagros|" + extID + "|" + key }

func secretKey(key string) error {
	if key == "" || len(key) > maxSecretKey {
		return fmt.Errorf("secret key must be 1 to %d bytes", maxSecretKey)
	}
	return nil
}

func (e *Extension) biSecretSet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key, value string
	if err := starlark.UnpackArgs("secrets.set", args, kw, "key", &key, "value", &value); err != nil {
		return nil, err
	}
	if err := e.require(PermSecrets, ""); err != nil {
		return nil, err
	}
	if err := secretKey(key); err != nil {
		return nil, err
	}
	if len(value) > maxSecretValue {
		return nil, fmt.Errorf("secret value is larger than %d bytes", maxSecretValue)
	}
	if err := secretBackend.Set(secretTarget(e.Manifest.ID, key), value); err != nil {
		return nil, fmt.Errorf("secrets.set %q: %w", key, err)
	}
	return starlark.None, nil
}

func (e *Extension) biSecretGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	var def starlark.Value = starlark.None
	if err := starlark.UnpackArgs("secrets.get", args, kw, "key", &key, "default?", &def); err != nil {
		return nil, err
	}
	if err := e.require(PermSecrets, ""); err != nil {
		return nil, err
	}
	if err := secretKey(key); err != nil {
		return nil, err
	}
	v, ok, err := secretBackend.Get(secretTarget(e.Manifest.ID, key))
	if err != nil {
		return nil, fmt.Errorf("secrets.get %q: %w", key, err)
	}
	if !ok {
		return def, nil
	}
	return starlark.String(v), nil
}

func (e *Extension) biSecretDelete(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	if err := starlark.UnpackArgs("secrets.delete", args, kw, "key", &key); err != nil {
		return nil, err
	}
	if err := e.require(PermSecrets, ""); err != nil {
		return nil, err
	}
	if err := secretKey(key); err != nil {
		return nil, err
	}
	if err := secretBackend.Delete(secretTarget(e.Manifest.ID, key)); err != nil {
		return nil, fmt.Errorf("secrets.delete %q: %w", key, err)
	}
	return starlark.None, nil
}

// ---- registry ----

var regRoots = map[string]string{
	"HKCU": "HKCU", "HKEY_CURRENT_USER": "HKCU",
	"HKLM": "HKLM", "HKEY_LOCAL_MACHINE": "HKLM",
}

// parseRegKey splits "HKCU\Software\X" into a canonical hive and path.
func parseRegKey(s string) (hive, path string, err error) {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '\\' || r == '/' })
	if len(parts) == 0 {
		return "", "", fmt.Errorf("registry key %q needs a root such as HKCU or HKLM", s)
	}
	h, ok := regRoots[strings.ToUpper(parts[0])]
	if !ok {
		return "", "", fmt.Errorf("registry root %q is not supported (use HKCU or HKLM)", parts[0])
	}
	return h, strings.Join(parts[1:], `\`), nil
}

// registryWithin reports whether target is the scope key or below it.
func registryWithin(scope, target string) bool {
	sh, sp, err := parseRegKey(scope)
	if err != nil {
		return false
	}
	th, tp, err := parseRegKey(target)
	if err != nil || sh != th {
		return false
	}
	sp, tp = strings.ToLower(sp), strings.ToLower(tp)
	return sp == "" || tp == sp || strings.HasPrefix(tp, sp+`\`)
}

func (e *Extension) regKey(fn, key string) (hive, path string, err error) {
	hive, path, err = parseRegKey(key)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", fn, err)
	}
	target := hive
	if path != "" {
		target += `\` + path
	}
	if err := e.require(PermRegistryRead, target); err != nil {
		return "", "", err
	}
	return hive, path, nil
}

func sysValue(v any) starlark.Value {
	switch x := v.(type) {
	case nil:
		return starlark.None
	case string:
		return starlark.String(x)
	case int64:
		return starlark.MakeInt64(x)
	case []byte:
		return starlark.Bytes(x)
	case []string:
		return goToStarlark(x)
	}
	return starlark.None
}

func (e *Extension) biRegistryGet(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key, name string
	if err := starlark.UnpackArgs("windows.registry_get", args, kw, "key", &key, "name?", &name); err != nil {
		return nil, err
	}
	h, p, err := e.regKey("windows.registry_get", key)
	if err != nil {
		return nil, err
	}
	v, err := osRegistryGet(h, p, name)
	if err != nil {
		return nil, fmt.Errorf("windows.registry_get %s: %w", key, err)
	}
	return sysValue(v), nil
}

func (e *Extension) biRegistryKeys(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	if err := starlark.UnpackArgs("windows.registry_keys", args, kw, "key", &key); err != nil {
		return nil, err
	}
	h, p, err := e.regKey("windows.registry_keys", key)
	if err != nil {
		return nil, err
	}
	names, err := osRegistryKeys(h, p)
	if err != nil {
		return nil, fmt.Errorf("windows.registry_keys %s: %w", key, err)
	}
	return goToStarlark(names), nil
}

func (e *Extension) biRegistryValues(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var key string
	if err := starlark.UnpackArgs("windows.registry_values", args, kw, "key", &key); err != nil {
		return nil, err
	}
	h, p, err := e.regKey("windows.registry_values", key)
	if err != nil {
		return nil, err
	}
	vals, err := osRegistryValues(h, p)
	if err != nil {
		return nil, fmt.Errorf("windows.registry_values %s: %w", key, err)
	}
	d := starlark.NewDict(len(vals))
	for k, v := range vals {
		d.SetKey(starlark.String(k), sysValue(v))
	}
	return d, nil
}

// ---- system ----

func dict(kv ...any) *starlark.Dict {
	d := starlark.NewDict(len(kv) / 2)
	for i := 0; i+1 < len(kv); i += 2 {
		d.SetKey(starlark.String(kv[i].(string)), goToStarlark(kv[i+1]))
	}
	return d
}

func serviceDict(s sysService) *starlark.Dict {
	return dict("name", s.Name, "display_name", s.Display, "state", s.State, "start_type", s.StartType, "pid", s.PID)
}

func (e *Extension) biServices(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("windows.services", args, kw); err != nil {
		return nil, err
	}
	if err := e.require(PermSystemRead, ""); err != nil {
		return nil, err
	}
	svcs, err := osServices()
	if err != nil {
		return nil, fmt.Errorf("windows.services: %w", err)
	}
	out := make([]starlark.Value, len(svcs))
	for i, s := range svcs {
		out[i] = serviceDict(s)
	}
	return starlark.NewList(out), nil
}

func (e *Extension) biServiceStatus(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs("windows.service_status", args, kw, "name", &name); err != nil {
		return nil, err
	}
	if err := e.require(PermSystemRead, ""); err != nil {
		return nil, err
	}
	s, err := osServiceStatus(name)
	if err != nil {
		return nil, fmt.Errorf("windows.service_status: %w", err)
	}
	return serviceDict(s), nil
}

func (e *Extension) biListWindows(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("windows.list_windows", args, kw); err != nil {
		return nil, err
	}
	if err := e.require(PermSystemRead, ""); err != nil {
		return nil, err
	}
	ws, err := osWindows()
	if err != nil {
		return nil, fmt.Errorf("windows.list_windows: %w", err)
	}
	out := make([]starlark.Value, len(ws))
	for i, w := range ws {
		out[i] = dict("title", w.Title, "process", w.Process, "pid", w.PID)
	}
	return starlark.NewList(out), nil
}

func (e *Extension) biProcesses(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("windows.processes", args, kw); err != nil {
		return nil, err
	}
	if err := e.require(PermSystemRead, ""); err != nil {
		return nil, err
	}
	ps, err := osProcesses()
	if err != nil {
		return nil, fmt.Errorf("windows.processes: %w", err)
	}
	out := make([]starlark.Value, len(ps))
	for i, p := range ps {
		out[i] = dict("name", p.Name, "pid", p.PID)
	}
	return starlark.NewList(out), nil
}

func (e *Extension) biKnownFolder(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs("windows.known_folder", args, kw, "name", &name); err != nil {
		return nil, err
	}
	if err := e.require(PermSystemRead, ""); err != nil {
		return nil, err
	}
	p, err := osKnownFolder(name)
	if err != nil {
		return nil, fmt.Errorf("windows.known_folder: %w", err)
	}
	return starlark.String(p), nil
}
