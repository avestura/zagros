package ext

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type fakeSecrets struct {
	mu sync.Mutex
	m  map[string]string
}

func (f *fakeSecrets) Set(t, v string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[t] = v
	return nil
}
func (f *fakeSecrets) Get(t string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.m[t]
	return v, ok, nil
}
func (f *fakeSecrets) Delete(t string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, t)
	return nil
}

func useFakeSecrets(t *testing.T) *fakeSecrets {
	t.Helper()
	f := &fakeSecrets{m: map[string]string{}}
	old := secretBackend
	secretBackend = f
	t.Cleanup(func() { secretBackend = old })
	return f
}

func queryTitle(t *testing.T, e *Extension) string {
	t.Helper()
	items, err := e.Query(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items: %v", items)
	}
	return items[0].Title
}

func TestPermissionKindsSecretsSystem(t *testing.T) {
	ok := []Permission{"secrets", "system.read", `registry.read:HKCU\Software\X`, "registry.read"}
	for _, p := range ok {
		if err := p.Validate(); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	bad := []Permission{"secrets:x", "system.read:all", "registry.read:"}
	for _, p := range bad {
		if p.Validate() == nil {
			t.Errorf("%s should be invalid", p)
		}
	}
}

func TestRegistryScope(t *testing.T) {
	ps := PermissionSet{`registry.read:HKLM\SOFTWARE\Microsoft\Windows NT`}
	cases := []struct {
		target string
		want   bool
	}{
		{`HKLM\SOFTWARE\Microsoft\Windows NT`, true},
		{`HKLM\software\microsoft\windows nt\CurrentVersion`, true},
		{`HKLM\SOFTWARE\Microsoft\Windows NT2`, false},
		{`HKLM\SOFTWARE\Microsoft`, false},
		{`HKCU\SOFTWARE\Microsoft\Windows NT`, false},
		{`HKLM`, false},
	}
	for _, c := range cases {
		if got := ps.Allows(PermRegistryRead, c.target); got != c.want {
			t.Errorf("Allows(%s)=%v want %v", c.target, got, c.want)
		}
	}
	if !(PermissionSet{"registry.read:HKCU"}).Allows(PermRegistryRead, `HKCU\Anything\Deep`) {
		t.Error("root scope should cover the whole hive")
	}
	if !(PermissionSet{"registry.read"}).Allows(PermRegistryRead, `HKLM\X`) {
		t.Error("unscoped grant should allow any key")
	}
	if _, _, err := parseRegKey(`HKCR\x`); err == nil {
		t.Error("HKCR should be rejected")
	}
}

func TestSecretsDeniedWithoutPermission(t *testing.T) {
	useFakeSecrets(t)
	for _, call := range []string{`secrets.get("k")`, `secrets.set("k", "v")`, `secrets.delete("k")`} {
		e := openStarted(t, baseManifest, "def on_query(ctx, q):\n    "+call+"\n    return []\n", nil, Options{})
		if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s: want permission denied, got %v", call, err)
		}
	}
}

func TestSecretsRoundtripAndNamespace(t *testing.T) {
	f := useFakeSecrets(t)
	write := `def on_query(ctx, q):
    secrets.set("token", "s3cret")
    return [item("1", secrets.get("token") + "|" + str(secrets.get("missing", "dflt")))]
`
	a := openStarted(t, baseManifest, write, []Permission{"secrets"}, Options{})
	if got := queryTitle(t, a); got != "s3cret|dflt" {
		t.Fatalf("got %q", got)
	}
	if v := f.m[secretTarget("test.ext", "token")]; v != "s3cret" {
		t.Fatalf("backend: %v", f.m)
	}

	other := strings.Replace(baseManifest, `"test.ext"`, `"other.ext"`, 1)
	read := `def on_query(ctx, q):
    return [item("1", str(secrets.get("token")))]
`
	b := openStarted(t, other, read, []Permission{"secrets"}, Options{})
	if got := queryTitle(t, b); got != "None" {
		t.Fatalf("other extension read %q", got)
	}
	del := `def on_query(ctx, q):
    secrets.delete("token")
    secrets.delete("token")
    return [item("1", str(secrets.get("token")))]
`
	b2 := openStarted(t, other, del, []Permission{"secrets"}, Options{})
	queryTitle(t, b2)
	if f.m[secretTarget("test.ext", "token")] != "s3cret" {
		t.Fatal("other extension deleted the first one's secret")
	}
	a2 := openStarted(t, baseManifest, del, []Permission{"secrets"}, Options{})
	if got := queryTitle(t, a2); got != "None" {
		t.Fatalf("after delete: %q", got)
	}
}

func TestSecretTargetsDoNotCollide(t *testing.T) {
	if secretTarget("a/b", "c") == secretTarget("a", "b/c") {
		t.Fatal("collision across slash ids")
	}
}

func TestSecretsLimits(t *testing.T) {
	useFakeSecrets(t)
	big := strings.Repeat("x", maxSecretValue+1)
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    secrets.set("k", "`+big+`")
    return []
`, []Permission{"secrets"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("got %v", err)
	}
	e = openStarted(t, baseManifest, `def on_query(ctx, q):
    secrets.get("")
    return []
`, []Permission{"secrets"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil {
		t.Fatal("empty key accepted")
	}
}

func TestSecretValueNeverLogged(t *testing.T) {
	useFakeSecrets(t)
	var sink logSink
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    secrets.set("k", "TOPSECRETVALUE")
    secrets.get("k")
    return []
`, []Permission{"secrets"}, Options{Log: sink.fn})
	if _, err := e.Query(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if sink.has("TOPSECRETVALUE") {
		t.Fatal("secret value appeared in logs")
	}
}

func TestSystemDeniedWithoutPermission(t *testing.T) {
	calls := map[string]string{
		"services":       `windows.services()`,
		"service_status": `windows.service_status("Spooler")`,
		"list_windows":   `windows.list_windows()`,
		"processes":      `windows.processes()`,
		"known_folder":   `windows.known_folder("documents")`,
		"reg":            `windows.registry_get("HKCU\\Software")`,
		"reg_keys":       `windows.registry_keys("HKCU\\Software")`,
		"reg_values":     `windows.registry_values("HKCU\\Software")`,
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			e := openStarted(t, baseManifest, "def on_query(ctx, q):\n    "+call+"\n    return []\n", nil, Options{})
			if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("want permission denied, got %v", err)
			}
		})
	}
}

func TestRegistryOutsideScopeDenied(t *testing.T) {
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    windows.registry_get("HKLM\\SOFTWARE\\Other", "x")
    return []
`, []Permission{`registry.read:HKLM\SOFTWARE\Microsoft`}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("want permission denied, got %v", err)
	}
}

func TestWindowsModuleNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-Windows behaviour")
	}
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    windows.processes()
    return []
`, []Permission{"system.read"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "Windows") {
		t.Fatalf("got %v", err)
	}
}
