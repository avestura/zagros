package pathx

import (
	"math/rand"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "path") }

func str(t *testing.T, name string, kv ...any) string {
	t.Helper()
	v := toolstest.Must(t, "path", name, kv...)
	s, ok := v.(string)
	if !ok {
		t.Fatalf("%s returned %T", name, v)
	}
	return s
}

func TestClean(t *testing.T) {
	cases := []struct{ flavor, in, want string }{
		{"posix", "/a/./b/../c//d/", "/a/c/d"},
		{"posix", "", "."},
		{"posix", "../../a", "../../a"},
		{"posix", "/..", "/"},
		{"posix", "a/..", "."},
		{"windows", `C:\a\.\b\..\c//d\`, `C:\a\c\d`},
		{"windows", `C:`, `C:.`},
		{"windows", `c:..\x`, `c:..\x`},
		{"windows", `C:\..\x`, `C:\x`},
		{"windows", `\\host\share\a\..\b`, `\\host\share\b`},
		{"windows", `\\host\share`, `\\host\share\`},
		{"windows", `/a/b`, `\a\b`},
	}
	for _, c := range cases {
		if got := str(t, "clean", "path", c.in, "flavor", c.flavor); got != c.want {
			t.Errorf("clean(%q,%s) = %q want %q", c.in, c.flavor, got, c.want)
		}
	}
}

func TestJoin(t *testing.T) {
	got := toolstest.Must(t, "path", "join", "parts", []string{"/a", "", "b/", "../c", "d.txt"})
	if got != "/a/c/d.txt" {
		t.Errorf("got %v", got)
	}
	got = toolstest.Must(t, "path", "join", "parts", []string{`C:\Users`, "me", "..", "docs"}, "flavor", "windows")
	if got != `C:\Users\docs` {
		t.Errorf("got %v", got)
	}
	got = toolstest.Must(t, "path", "join", "parts", []string{"", ""})
	if got != "" {
		t.Errorf("got %v", got)
	}
	got = toolstest.Must(t, "path", "join", "parts", []string{`\\host`, "share", "x"}, "flavor", "windows")
	if got != `\\host\share\x` {
		t.Errorf("got %v", got)
	}
	got = toolstest.Must(t, "path", "join", "parts", []string{`C:\a`, "b"}, "flavor", "auto")
	if got != `C:\a\b` {
		t.Errorf("auto: got %v", got)
	}
}

func TestPieces(t *testing.T) {
	type c struct{ fn, flavor, in, want string }
	for _, x := range []c{
		{"base", "posix", "/usr/local/bin/tool.exe", "tool.exe"},
		{"base", "posix", "/a/b/", "b"},
		{"base", "posix", "/", "/"},
		{"base", "posix", "", "."},
		{"base", "windows", `C:\Users\me\`, "me"},
		{"base", "windows", `C:\`, `\`},
		{"dir", "posix", "/a/b/c.txt", "/a/b"},
		{"dir", "posix", "c.txt", "."},
		{"dir", "posix", "/a", "/"},
		{"dir", "windows", `C:\Users\me\a.txt`, `C:\Users\me`},
		{"dir", "windows", `C:a.txt`, `C:.`},
		{"dir", "windows", `\\host\share\a.txt`, `\\host\share\`},
		{"ext", "posix", "archive.tar.gz", ".gz"},
		{"ext", "posix", ".bashrc", ""},
		{"ext", "posix", "name.", ""},
		{"ext", "posix", "/a.b/c", ""},
		{"ext", "windows", `C:\a.b\file.TXT`, ".TXT"},
		{"stem", "posix", "/tmp/report.final.pdf", "report.final"},
		{"stem", "posix", ".env", ".env"},
	} {
		if got := str(t, x.fn, "path", x.in, "flavor", x.flavor); got != x.want {
			t.Errorf("%s(%q,%s) = %q want %q", x.fn, x.in, x.flavor, got, x.want)
		}
	}
	sp := toolstest.Must(t, "path", "split", "path", "/a/b/c.txt").([]string)
	if len(sp) != 2 || sp[0] != "/a/b/" || sp[1] != "c.txt" {
		t.Errorf("split: %v", sp)
	}
	sp = toolstest.Must(t, "path", "split", "path", `C:\a\b.txt`, "flavor", "windows").([]string)
	if sp[0] != `C:\a\` || sp[1] != "b.txt" {
		t.Errorf("split: %v", sp)
	}
	pt := toolstest.Must(t, "path", "parts", "path", `C:\a\.\b\c.txt`, "flavor", "windows").([]string)
	if len(pt) != 4 || pt[0] != `C:\` || pt[3] != "c.txt" {
		t.Errorf("parts: %v", pt)
	}
	pt = toolstest.Must(t, "path", "parts", "path", "/a/b/").([]string)
	if len(pt) != 3 || pt[0] != "/" {
		t.Errorf("parts: %v", pt)
	}
}

func TestIsAbs(t *testing.T) {
	for in, want := range map[string]bool{"/a": true, "a": false, "": false, `C:\a`: false} {
		if got := toolstest.Must(t, "path", "is_abs", "path", in); got != want {
			t.Errorf("posix is_abs(%q) = %v", in, got)
		}
	}
	for in, want := range map[string]bool{`C:\a`: true, `C:/a`: true, `\a`: false, `C:a`: false, `\\h\s\x`: true, `a\b`: false} {
		if got := toolstest.Must(t, "path", "is_abs", "path", in, "flavor", "windows"); got != want {
			t.Errorf("windows is_abs(%q) = %v", in, got)
		}
	}
}

func TestRel(t *testing.T) {
	cases := []struct{ flavor, base, target, want string }{
		{"posix", "/data/app", "/data/lib/x.so", "../lib/x.so"},
		{"posix", "/a/b", "/a/b", "."},
		{"posix", "a", "a/b/c", "b/c"},
		{"posix", "/a/b/c", "/a", "../.."},
		{"windows", `C:\Data\App`, `c:\data\lib\x.dll`, `..\lib\x.dll`},
	}
	for _, c := range cases {
		if got := str(t, "rel", "base", c.base, "target", c.target, "flavor", c.flavor); got != c.want {
			t.Errorf("rel(%q,%q) = %q want %q", c.base, c.target, got, c.want)
		}
	}
	toolstest.Fails(t, "path", "rel", "base", "/a", "target", "b")
	toolstest.Fails(t, "path", "rel", "base", `C:\a`, "target", `D:\a`, "flavor", "windows")
	toolstest.Fails(t, "path", "rel", "base", "../a", "target", "b")
}

func TestIsWithin(t *testing.T) {
	cases := []struct {
		parent, child string
		want          bool
	}{
		{"/srv/data", "/srv/data/a/b", true},
		{"/srv/data", "/srv/data", true},
		{"/srv/data", "/srv/data/../etc/passwd", false},
		{"/srv/data", "/srv/database", false},
		{"/srv/data", "relative", false},
	}
	for _, c := range cases {
		if got := toolstest.Must(t, "path", "is_within", "parent", c.parent, "child", c.child); got != c.want {
			t.Errorf("is_within(%q,%q) = %v", c.parent, c.child, got)
		}
	}
	if got := toolstest.Must(t, "path", "is_within", "parent", `C:\Data`, "child", `c:\data\x`, "flavor", "windows"); got != true {
		t.Error("windows comparison must ignore case")
	}
}

func TestWithExtAndConvert(t *testing.T) {
	for _, c := range [][3]string{{"/tmp/a.txt", "md", "/tmp/a.md"}, {"/tmp/a.tar.gz", ".zip", "/tmp/a.tar.zip"}, {"/tmp/a.txt", "", "/tmp/a"}, {"/tmp/noext", "x", "/tmp/noext.x"}, {"/tmp/dir/", "x", "/tmp/dir.x"}} {
		if got := str(t, "with_ext", "path", c[0], "ext", c[1]); got != c[2] {
			t.Errorf("with_ext(%q,%q) = %q want %q", c[0], c[1], got, c[2])
		}
	}
	toolstest.Fails(t, "path", "with_ext", "path", "/a/b", "ext", "x/y")
	toolstest.Fails(t, "path", "with_ext", "path", "/", "ext", "x")
	toolstest.Fails(t, "path", "with_ext", "path", "a", "ext", ".")
	if got := str(t, "convert", "path", `C:\Users\me\a.txt`, "to", "posix"); got != "C:/Users/me/a.txt" {
		t.Errorf("got %q", got)
	}
	if got := str(t, "convert", "path", "/a/b", "to", "windows"); got != `\a\b` {
		t.Errorf("got %q", got)
	}
	toolstest.Fails(t, "path", "convert", "path", "/a", "to", "dos")
	toolstest.Fails(t, "path", "clean", "path", "/a", "flavor", "amiga")
}

func TestNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(6))
	alphabet := []byte("ab./\\:C \x00é")
	names := []string{"clean", "base", "dir", "ext", "stem", "split", "parts", "is_abs"}
	for i := 0; i < 5000; i++ {
		b := make([]byte, rng.Intn(14))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		s := string(b)
		for _, fl := range []string{"posix", "windows", "auto"} {
			for _, n := range names {
				toolstest.Call("path", n, "path", s, "flavor", fl)
			}
			toolstest.Call("path", "rel", "base", s, "target", s[len(s)/2:], "flavor", fl)
			toolstest.Call("path", "is_within", "parent", s[len(s)/2:], "child", s, "flavor", fl)
			toolstest.Call("path", "with_ext", "path", s, "ext", "x", "flavor", fl)
			toolstest.Call("path", "join", "parts", []string{s, s[len(s)/2:]}, "flavor", fl)
		}
	}
}
