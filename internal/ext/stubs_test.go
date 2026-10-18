package ext

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	_ "github.com/avestura/zagros/internal/tools/all" // the tool modules must be stubbed too
	"github.com/avestura/zagros/stubs"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var defRe = regexp.MustCompile(`(?m)^def (\w+)\(`)

func stubDefs(t *testing.T, file string) map[string]bool {
	t.Helper()
	b, err := stubs.FS().Open("zagros/" + file)
	if err != nil {
		t.Fatalf("missing stub %s: %v", file, err)
	}
	defer b.Close()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := b.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	out := map[string]bool{}
	for _, m := range defRe.FindAllStringSubmatch(sb.String(), -1) {
		out[m[1]] = true
	}
	return out
}

// TestStubsMatchHost keeps the .pyi stubs in lockstep with the real runtime:
// every function the host predeclares must be stubbed, and nothing may be
// stubbed that the host does not have. A new host feature without its stub
// fails here.
func TestStubsMatchHost(t *testing.T) {
	e := &Extension{Manifest: &Manifest{ID: "t.t", Name: "t", Version: "1.0.0"}, opts: &Options{}, kv: newKV("", 0)}
	e.opts.defaults()
	pre := e.predeclared()

	globals := stubDefs(t, "__init__.pyi")
	hostGlobals := map[string]bool{}
	var problems []string
	for name, v := range pre {
		switch x := v.(type) {
		case starlark.Callable:
			hostGlobals[name] = true
			if !globals[name] {
				problems = append(problems, "global function not stubbed in __init__.pyi: "+name)
			}
		case *starlarkstruct.Module:
			members := stubDefs(t, name+".pyi")
			host := map[string]bool{}
			for _, a := range x.AttrNames() {
				av, err := x.Attr(a)
				if err != nil {
					continue
				}
				if _, ok := av.(starlark.Callable); ok {
					host[a] = true
					if !members[a] {
						problems = append(problems, "module function not stubbed: "+name+"."+a)
					}
				}
			}
			for m := range members {
				if !host[m] {
					problems = append(problems, "stub has no host function: "+name+"."+m)
				}
			}
		}
	}
	for g := range globals {
		if !hostGlobals[g] {
			problems = append(problems, "stub has no host global: "+g)
		}
	}
	sort.Strings(problems)
	if len(problems) > 0 {
		t.Fatalf("stubs out of sync with the host (update stubs/zagros/*.pyi, docs and the site):\n  %s", strings.Join(problems, "\n  "))
	}

	// hooks
	hooks := stubDefs(t, "hooks.pyi")
	for _, h := range []string{"on_load", "on_query", "on_scope", "on_execute", "on_settings_changed"} {
		if !hooks[h] {
			t.Errorf("hook %s is not in hooks.pyi", h)
		}
	}
	// permissions
	perm, _ := stubs.FS().Open("zagros/permissions.pyi")
	pb := make([]byte, 16384)
	n, _ := perm.Read(pb)
	for k := range allKinds {
		if !strings.Contains(string(pb[:n]), `"`+k+`"`) {
			t.Errorf("permission kind %q missing from permissions.pyi", k)
		}
	}
}

func TestScaffoldValidatesAndPacksWithoutDevFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-tool")
	if err := Scaffold(dir, "", ""); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"manifest.toml", "main.star", "AGENTS.md", "CLAUDE.md", "zagros-stubs/zagros/__init__.pyi", "zagros-stubs/zagros/hooks.pyi"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("scaffold missing %s", f)
		}
	}
	for _, is := range Validate(dir) {
		t.Errorf("generated project has an issue: %v", is)
	}
	if err := Scaffold(dir, "", ""); err == nil {
		t.Error("scaffold must not overwrite an existing project")
	}
	data, err := Pack(dir)
	if err != nil {
		t.Fatal(err)
	}
	zr := string(data)
	for _, bad := range []string{"zagros-stubs", "AGENTS.md", "CLAUDE.md"} {
		if strings.Contains(zr, bad) {
			t.Errorf("packed archive must not contain %s", bad)
		}
	}
	if !strings.Contains(zr, "main.star") || !strings.Contains(zr, "manifest.toml") {
		t.Error("packed archive lacks main files")
	}
}
