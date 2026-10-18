//go:build windows

package ext

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCredentialManagerRoundtrip(t *testing.T) {
	c := credManager{}
	target := secretTarget("zagros.test.credmgr", "k")
	t.Cleanup(func() { c.Delete(target) })
	if _, ok, err := c.Get(target); ok || err != nil {
		t.Fatalf("fresh get: ok=%v err=%v", ok, err)
	}
	want := "pässword ☃ with spaces"
	if err := c.Set(target, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := c.Get(target)
	if err != nil || !ok || got != want {
		t.Fatalf("got %q ok=%v err=%v", got, ok, err)
	}
	if err := c.Set(target, ""); err != nil {
		t.Fatal(err)
	}
	if got, ok, _ := c.Get(target); !ok || got != "" {
		t.Fatalf("empty value: %q %v", got, ok)
	}
	if err := c.Delete(target); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := c.Get(target); ok {
		t.Fatal("still present after delete")
	}
	if err := c.Delete(target); err != nil {
		t.Fatalf("deleting a missing entry: %v", err)
	}
}

func TestSecretsStarlarkRealBackend(t *testing.T) {
	m := strings.Replace(baseManifest, `"test.ext"`, `"zagros.test.starlark"`, 1)
	script := `def on_query(ctx, q):
    secrets.set("t", "v1")
    return [item("1", secrets.get("t"))]
`
	e := openStarted(t, m, script, []Permission{"secrets"}, Options{})
	t.Cleanup(func() { credManager{}.Delete(secretTarget("zagros.test.starlark", "t")) })
	if got := queryTitle(t, e); got != "v1" {
		t.Fatalf("got %q", got)
	}
}

func TestWindowsModuleLive(t *testing.T) {
	script := `def on_query(ctx, q):
    procs = windows.processes()
    names = [p["name"].lower() for p in procs]
    svc = windows.service_status("RpcSs")
    wins = windows.list_windows()
    build = windows.registry_get("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "CurrentBuild")
    miss = windows.registry_get("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "NoSuchValueZagros")
    keys = windows.registry_keys("HKLM\\SOFTWARE\\Microsoft\\Windows NT")
    vals = windows.registry_values("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion")
    all = windows.services()
    return [item("1", "|".join([
        str(len(procs) > 5), str("system" in names), svc["state"], str(len(all) > 20), str(type(wins)),
        str(int(build) > 1000), str(miss), str("CurrentVersion" in keys), str("CurrentBuild" in vals),
        windows.known_folder("documents"), windows.known_folder("downloads"), windows.known_folder("appdata"),
    ]))]
`
	e := openStarted(t, baseManifest, script, []Permission{"system.read", `registry.read:HKLM\SOFTWARE\Microsoft\Windows NT`}, Options{})
	got := queryTitle(t, e)
	f := strings.Split(got, "|")
	if len(f) < 12 {
		t.Fatalf("got %q", got)
	}
	if f[0] != "True" || f[1] != "True" || f[2] != "running" || f[3] != "True" || f[5] != "True" || f[6] != "None" || f[7] != "True" || f[8] != "True" {
		t.Fatalf("got %q", got)
	}
	for _, p := range f[9:12] {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("known folder %s: %v", p, err)
		}
	}
}

func TestKnownFolderUnknown(t *testing.T) {
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    windows.known_folder("nope")
    return []
`, []Permission{"system.read"}, Options{})
	if _, err := e.Query(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "unknown folder") {
		t.Fatalf("got %v", err)
	}
}
