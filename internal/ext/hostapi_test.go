package ext

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run executes body inside on_query and returns the title of the single
// result row; body must assign `out`.
func runScript(t *testing.T, body string, granted []Permission, o Options) (string, error) {
	t.Helper()
	script := "def on_query(ctx, q):\n"
	for _, l := range strings.Split(strings.TrimSpace(body), "\n") {
		script += "    " + l + "\n"
	}
	script += "    return [item(\"1\", str(out))]\n"
	e := openStarted(t, baseManifest, script, granted, o)
	t.Cleanup(e.Stop)
	items, err := e.Query(context.Background(), "")
	if err != nil {
		return "", err
	}
	return items[0].Title, nil
}

func mustRun(t *testing.T, body string, granted []Permission, o Options) string {
	t.Helper()
	s, err := runScript(t, body, granted, o)
	if err != nil {
		t.Fatalf("%v\nscript:\n%s", err, body)
	}
	return s
}

func slash(p string) string { return filepath.ToSlash(p) }

func TestAttemptCatchesErrors(t *testing.T) {
	got := mustRun(t, `
v, err = attempt(fs.read_text, "/definitely/not/here.txt")
ok, e2 = attempt(lambda x: x + 1, 1)
_, e3 = attempt(fail, "boom")
out = [v, err != None, ok, e2, e3]`, []Permission{"fs.read"}, Options{})
	if got != `[None, True, 2, None, "fail: boom"]` {
		t.Fatal(got)
	}
	// Permission errors are catchable, the step limit is not.
	got = mustRun(t, `_, err = attempt(fs.read_text, "/x")
out = "permission denied" in err`, nil, Options{})
	if got != "True" {
		t.Fatal(got)
	}
	e := openStarted(t, baseManifest, "def on_query(ctx, q):\n    attempt(lambda: [1 for _ in range(1000000000)])\n    return []\n", nil, Options{Limits: Limits{MaxSteps: 10_000}})
	if _, err := e.Query(context.Background(), ""); err == nil {
		t.Fatal("step limit must not be swallowed by attempt")
	}
}

func TestHTTPMethodsAndRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redir":
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		case "/cookie":
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc"})
			http.SetCookie(w, &http.Cookie{Name: "two", Value: "2"})
			return
		case "/big":
			w.Write(bytes.Repeat([]byte("x"), 5000))
			return
		case "/bin":
			w.Write([]byte{0, 1, 2, 255})
			return
		}
		b, _ := io.ReadAll(r.Body)
		r.ParseForm()
		fmt.Fprintf(w, "%s|%s|%s|%s|%s", r.Method, r.Header.Get("Content-Type"), string(b), r.Header.Get("Authorization"), r.Header.Get("Cookie"))
	}))
	defer srv.Close()
	u := srv.URL
	perms := []Permission{"network:127.0.0.1"}
	got := mustRun(t, `
out = [http.put("`+u+`/x", body = "b").text.split("|")[0],
       http.patch("`+u+`/x", json = {"a": 1}).text,
       http.delete("`+u+`/x").text.split("|")[0],
       http.head("`+u+`/x").text,
       http.head("`+u+`/x").status]`, perms, Options{})
	if got != `["PUT", "PATCH|application/json|{\"a\":1}||", "DELETE", "", 200]` {
		t.Fatal(got)
	}
	got = mustRun(t, `
r1 = http.request("post", "`+u+`/x", form = {"a": "1 2", "b": "x"}, auth = ("me", "pw"))
r2 = http.request("GET", "`+u+`/x", auth = "tok")
r3 = http.request("GET", "`+u+`/redir", follow_redirects = False)
r4 = http.request("GET", "`+u+`/bin")
out = [r1.text, r2.text.split("|")[3], r3.status, r3.headers["location"], r4.bytes == b"\x00\x01\x02\xff", len(r4.bytes)]`, perms, Options{})
	if !strings.HasPrefix(got, `["POST|application/x-www-form-urlencoded|a=1+2&b=x|Basic bWU6cHc=|", "Bearer tok", 302, "/final", True, 4]`) {
		t.Fatal(got)
	}
	got = mustRun(t, `
http.request("GET", "`+u+`/cookie", cookies = True)
a = http.request("GET", "`+u+`/x", cookies = True).text.split("|")[4]
b = http.request("GET", "`+u+`/x").text.split("|")[4]
http.clear_cookies()
c = http.request("GET", "`+u+`/x", cookies = True).text.split("|")[4]
raw = http.get("`+u+`/cookie").raw_headers["set-cookie"]
out = [a, b == "", c == "", len(raw)]`, perms, Options{})
	if got != `["sid=abc; two=2", True, True, 2]` {
		t.Fatal(got)
	}
	got = mustRun(t, `
r = http.request("POST", "`+u+`/x", multipart = {"note": "hello", "f": {"filename": "a.txt", "content": "data", "content_type": "text/plain"}, "raw": b"\x01"})
out = [r.text.split("|")[1].startswith("multipart/form-data; boundary="), 'name="note"' in r.text, 'filename="a.txt"' in r.text, 'filename="raw"' in r.text, "hello" in r.text]`, perms, Options{})
	if got != `[True, True, True, True, True]` {
		t.Fatal(got)
	}
	// size caps and argument validation
	if _, err := runScript(t, `out = http.request("GET", "`+u+`/big", max_bytes = 100)`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "larger than 100") {
		t.Fatalf("max_bytes: %v", err)
	}
	if _, err := runScript(t, `out = http.get("`+u+`/big")`, perms, Options{Limits: Limits{MaxHTTPBytes: 1000}}); err == nil || !strings.Contains(err.Error(), "larger than 1000") {
		t.Fatalf("cap: %v", err)
	}
	if _, err := runScript(t, `out = http.request("POST", "`+u+`/x", body = "a", json = {})`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "only one of") {
		t.Fatalf("exclusive: %v", err)
	}
	// host scoping applies to every new function
	for _, call := range []string{`http.put("%s/x")`, `http.patch("%s/x")`, `http.delete("%s/x")`, `http.head("%s/x")`, `http.request("GET", "%s/x")`, `http.download("%s/x", "/tmp/x")`} {
		_, err := runScript(t, "out = "+fmt.Sprintf(call, u), []Permission{"network:example.com", "fs.write"}, Options{})
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s without the host: %v", call, err)
		}
	}
	// redirect to an ungranted host stays blocked with the new request()
	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://localhost:1/", 302) }))
	defer redir.Close()
	if _, err := runScript(t, `out = http.request("GET", "`+redir.URL+`")`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("redirect: %v", err)
	}
}

func TestHTTPDownload(t *testing.T) {
	payload := bytes.Repeat([]byte("0123456789"), 100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.NotFound(w, r)
			return
		}
		w.Write(payload)
	}))
	defer srv.Close()
	dir := t.TempDir()
	dst := slash(filepath.Join(dir, "sub", "f.bin"))
	perms := []Permission{"network:127.0.0.1", Permission("fs.write:" + dir), Permission("fs.read:" + dir)}
	got := mustRun(t, `
d = http.download("`+srv.URL+`/f", "`+dst+`")
m = http.download("`+srv.URL+`/missing", "`+slash(filepath.Join(dir, "nope.bin"))+`")
out = [d.status, d.size, d.path != None, m.status, m.path, m.ok, len(fs.read_bytes("`+dst+`"))]`, perms, Options{})
	if got != "[200, 1000, True, 404, None, False, 1000]" {
		t.Fatal(got)
	}
	if _, err := os.Stat(filepath.Join(dir, "nope.bin")); err == nil {
		t.Fatal("a 404 must not create the file")
	}
	// network alone is not enough: the target needs fs.write
	if _, err := runScript(t, `out = http.download("`+srv.URL+`/f", "`+dst+`")`, []Permission{"network:127.0.0.1"}, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("no fs.write: %v", err)
	}
	// outside the writable folder
	other := slash(filepath.Join(t.TempDir(), "x.bin"))
	if _, err := runScript(t, `out = http.download("`+srv.URL+`/f", "`+other+`")`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("outside scope: %v", err)
	}
	// size cap leaves nothing behind
	small := slash(filepath.Join(dir, "small.bin"))
	if _, err := runScript(t, `out = http.download("`+srv.URL+`/f", "`+small+`", max_bytes = 100)`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "larger than 100") {
		t.Fatalf("cap: %v", err)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 1 { // only sub/
		t.Fatalf("leftover files: %v", ents)
	}
	if _, err := runScript(t, `out = http.download("`+srv.URL+`/f", "`+dst+`", overwrite = False)`, perms, Options{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("overwrite=False: %v", err)
	}
}

func fsPerms(dir string) []Permission {
	return []Permission{Permission("fs.read:" + dir), Permission("fs.write:" + dir)}
}

func TestFSNewFunctions(t *testing.T) {
	dir := t.TempDir()
	d := slash(dir)
	got := mustRun(t, `
fs.mkdir("`+d+`/a/b")
fs.write_bytes("`+d+`/a/b/x.bin", b"\x00\x01\x02")
fs.write_text("`+d+`/a/t.txt", "one")
fs.append_text("`+d+`/a/t.txt", "\ntwo")
fs.append_text("`+d+`/new.txt", "made")
st = fs.stat("`+d+`/a/t.txt")
sd = fs.stat("`+d+`/a")
walk = [e.rel for e in fs.walk("`+d+`")]
shallow = [e.rel for e in fs.walk("`+d+`", max_depth = 1)]
out = [fs.read_bytes("`+d+`/a/b/x.bin") == b"\x00\x01\x02", fs.read_text("`+d+`/a/t.txt"), fs.read_text("`+d+`/new.txt"),
       st.size, st.is_file, sd.is_dir, st.mtime > 0, walk, shallow]`, fsPerms(dir), Options{})
	want := `[True, "one\ntwo", "made", 7, True, True, True, ["a", "a/b", "a/b/x.bin", "a/t.txt", "new.txt"], ["a", "new.txt"]]`
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}

	got = mustRun(t, `
fs.copy("`+d+`/a", "`+d+`/copy")
fs.rename("`+d+`/new.txt", "`+d+`/moved.txt")
g = fs.glob("`+d+`/**/*.txt")
g2 = fs.glob("`+d+`/*/*.bin")
g3 = fs.glob("`+d+`/a/**/*.bin")
r1 = fs.remove("`+d+`/copy", recursive = True)
r2 = fs.remove("`+d+`/copy", missing_ok = True)
out = [[p.removeprefix("`+d+`/") for p in g], g2, [p.removeprefix("`+d+`/") for p in g3], r1, r2, fs.exists("`+d+`/copy"), fs.exists("`+d+`/new.txt"), fs.exists("`+d+`/moved.txt")]`, fsPerms(dir), Options{})
	want = `[["a/t.txt", "copy/t.txt", "moved.txt"], [], ["a/b/x.bin"], True, False, False, False, True]`
	if got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}

	// errors raise; attempt turns them into values
	got = mustRun(t, `
_, e1 = attempt(fs.remove, "`+d+`/a")
_, e2 = attempt(fs.copy, "`+d+`/a/t.txt", "`+d+`/moved.txt")
_, e3 = attempt(fs.rename, "`+d+`/a/t.txt", "`+d+`/moved.txt")
_, e4 = attempt(fs.stat, "`+d+`/gone")
_, e5 = attempt(fs.remove, "`+d+`")
out = [e1 != None, "already exists" in e2, "already exists" in e3, e4 != None, "refusing" in e5]`, fsPerms(dir), Options{})
	if got != "[True, True, True, True, True]" {
		t.Fatal(got)
	}
}

func TestFSScopeForNewFunctions(t *testing.T) {
	allowed, other := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(other, "secret.txt"), []byte("s"), 0o644)
	a, o := slash(allowed), slash(other)
	calls := []string{
		`fs.read_bytes("%[1]s/secret.txt")`, `fs.stat("%[1]s/secret.txt")`, `fs.glob("%[1]s/*.txt")`, `fs.walk("%[1]s")`,
		`fs.write_bytes("%[1]s/n.bin", b"x")`, `fs.append_text("%[1]s/secret.txt", "x")`, `fs.mkdir("%[1]s/d")`,
		`fs.remove("%[1]s/secret.txt")`, `fs.rename("%[1]s/secret.txt", "%[2]s/s.txt")`, `fs.rename("%[2]s/ok.txt", "%[1]s/s.txt")`,
		`fs.copy("%[1]s/secret.txt", "%[2]s/s.txt")`, `fs.copy("%[2]s/ok.txt", "%[1]s/s.txt")`,
		`fs.remove("%[2]s/../" + "` + filepath.Base(other) + `/secret.txt")`,
	}
	os.WriteFile(filepath.Join(allowed, "ok.txt"), []byte("ok"), 0o644)
	for _, c := range calls {
		body := "out = " + fmt.Sprintf(c, o, a)
		_, err := runScript(t, body, fsPerms(allowed), Options{})
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s: want permission denied, got %v", body, err)
		}
	}
	// read-only grant cannot write
	if _, err := runScript(t, `out = fs.mkdir("`+a+`/z")`, []Permission{Permission("fs.read:" + allowed)}, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("write with read grant: %v", err)
	}
	// the granted root itself cannot be removed
	if _, err := runScript(t, `out = fs.remove("`+a+`", recursive = True)`, fsPerms(allowed), Options{}); err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("remove root: %v", err)
	}
}

func TestFSSymlinkEscape(t *testing.T) {
	allowed, other := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(other, "secret.txt"), []byte("s"), 0o644)
	os.MkdirAll(filepath.Join(other, "dir"), 0o755)
	os.WriteFile(filepath.Join(other, "dir", "f.txt"), []byte("f"), 0o644)
	if err := os.Symlink(other, filepath.Join(allowed, "link")); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}
	a := slash(allowed)
	for _, body := range []string{
		`out = fs.read_bytes("` + a + `/link/secret.txt")`,
		`out = fs.walk("` + a + `/link")`,
		`out = fs.glob("` + a + `/link/*.txt")`,
		`out = fs.write_bytes("` + a + `/link/new.bin", b"x")`,
		`out = fs.remove("` + a + `/link/secret.txt")`,
		`out = fs.copy("` + a + `/link/secret.txt", "` + a + `/c.txt")`,
		`out = fs.rename("` + a + `/link/secret.txt", "` + a + `/c.txt")`,
	} {
		if _, err := runScript(t, body, fsPerms(allowed), Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s: want permission denied, got %v", body, err)
		}
	}
	// links inside a walked/copied/archived tree are neither followed nor copied
	os.MkdirAll(filepath.Join(allowed, "tree"), 0o755)
	os.Symlink(other, filepath.Join(allowed, "tree", "l2"))
	os.WriteFile(filepath.Join(allowed, "tree", "real.txt"), []byte("r"), 0o644)
	got := mustRun(t, `
w = sorted([e.rel + ":" + str(e.is_symlink) for e in fs.walk("`+a+`/tree")])
fs.copy("`+a+`/tree", "`+a+`/treecopy")
fs.write_text("`+a+`/treecopy/x.txt", "x")
out = [w, sorted(fs.list("`+a+`/treecopy"))]`, fsPerms(allowed), Options{})
	if got != `[["l2:True", "real.txt:False"], ["real.txt", "x.txt"]]` {
		t.Fatal(got)
	}
	// removing the link itself is allowed and leaves the target alone
	mustRun(t, `out = fs.remove("`+a+`/link")`, fsPerms(allowed), Options{})
	if _, err := os.Stat(filepath.Join(other, "secret.txt")); err != nil {
		t.Fatal("target of a removed link must survive")
	}
}

func TestCompressBytes(t *testing.T) {
	got := mustRun(t, `
data = b"hello hello hello hello hello hello"
out = [compress.gunzip(compress.gzip(data)) == data, compress.unzlib(compress.zlib(data, 9)) == data,
       compress.inflate(compress.deflate(data)) == data, len(compress.gzip(data * 50)) < len(data) * 50,
       compress.gunzip(compress.gzip("text")) == b"text"]`, nil, Options{})
	if got != "[True, True, True, True, True]" {
		t.Fatal(got)
	}
	// decompression bomb: tiny input, huge output
	_, err := runScript(t, `out = compress.gunzip(compress.gzip(b"\x00" * 3000000), max_bytes = 1000)`, nil, Options{})
	if err == nil || !strings.Contains(err.Error(), "larger than 1000") {
		t.Fatalf("bomb: %v", err)
	}
	_, err = runScript(t, `out = compress.gunzip(compress.gzip(b"\x00" * 3000))`, nil, Options{Limits: Limits{MaxFileBytes: 1000}})
	if err == nil || !strings.Contains(err.Error(), "larger than 1000") {
		t.Fatalf("default cap: %v", err)
	}
	if _, err := runScript(t, `out = compress.gunzip(b"not gzip")`, nil, Options{}); err == nil {
		t.Fatal("garbage must fail")
	}
}

func TestCompressArchives(t *testing.T) {
	dir := t.TempDir()
	d := slash(dir)
	os.MkdirAll(filepath.Join(dir, "src", "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "a.txt"), []byte("alpha"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "sub", "b.txt"), []byte("beta"), 0o644)
	os.WriteFile(filepath.Join(dir, "single.txt"), []byte("one"), 0o644)
	for _, ext := range []string{"zip", "tar", "tar.gz", "tgz"} {
		got := mustRun(t, `
n = compress.create("`+d+`/out.`+ext+`", ["`+d+`/src", "`+d+`/single.txt"])
names = sorted([e.name for e in compress.list("`+d+`/out.`+ext+`")])
files = compress.extract("`+d+`/out.`+ext+`", "`+d+`/x_`+strings.ReplaceAll(ext, ".", "_")+`")
out = [n, names, sorted(files), fs.read_text("`+d+`/x_`+strings.ReplaceAll(ext, ".", "_")+`/src/sub/b.txt"), fs.read_text("`+d+`/x_`+strings.ReplaceAll(ext, ".", "_")+`/single.txt")]`, fsPerms(dir), Options{})
		want := `[5, ["single.txt", "src/", "src/a.txt", "src/sub/", "src/sub/b.txt"], ["single.txt", "src/a.txt", "src/sub/b.txt"], "beta", "one"]`
		if got != want {
			t.Errorf("%s: got %q", ext, got)
		}
	}
	// existing files are an error without overwrite
	if _, err := runScript(t, `out = compress.extract("`+d+`/out.zip", "`+d+`/x_zip")`, fsPerms(dir), Options{}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("overwrite: %v", err)
	}
	mustRun(t, `out = compress.extract("`+d+`/out.zip", "`+d+`/x_zip", overwrite = True)`, fsPerms(dir), Options{})
	// the archive does not contain itself when written inside a source
	mustRun(t, `out = compress.create("`+d+`/src/self.zip", ["`+d+`/src"])`, fsPerms(dir), Options{})
	got := mustRun(t, `out = sorted([e.name for e in compress.list("`+d+`/src/self.zip")])`, fsPerms(dir), Options{})
	if strings.Contains(got, "self.zip") {
		t.Fatalf("archive contains itself: %s", got)
	}
	// scope: sources need fs.read, dest needs fs.write
	other := t.TempDir()
	os.WriteFile(filepath.Join(other, "s.txt"), []byte("s"), 0o644)
	for _, body := range []string{
		`out = compress.create("` + d + `/bad.zip", ["` + slash(other) + `/s.txt"])`,
		`out = compress.create("` + slash(other) + `/bad.zip", ["` + d + `/single.txt"])`,
		`out = compress.extract("` + d + `/out.zip", "` + slash(other) + `/x")`,
		`out = compress.list("` + slash(other) + `/nope.zip")`,
	} {
		if _, err := runScript(t, body, fsPerms(dir), Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s: %v", body, err)
		}
	}
	if _, err := runScript(t, `out = compress.extract("`+d+`/out.zip", "`+d+`/ro")`, []Permission{Permission("fs.read:" + dir)}, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("extract needs fs.write: %v", err)
	}
	if _, err := runScript(t, `out = compress.create("`+d+`/q.data", ["`+d+`/single.txt"])`, fsPerms(dir), Options{}); err == nil || !strings.Contains(err.Error(), "format") {
		t.Fatalf("unknown format: %v", err)
	}
}

func TestCompressZipSlipAndBombs(t *testing.T) {
	dir := t.TempDir()
	d := slash(dir)
	writeZip := func(name string, entries map[string]string) {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for n, c := range entries {
			w, _ := zw.Create(n)
			w.Write([]byte(c))
		}
		zw.Close()
		os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644)
	}
	for i, evil := range []string{"../evil.txt", "a/../../evil.txt", "/abs/evil.txt", `..\evil.txt`, "C:/evil.txt", "C:evil.txt"} {
		name := fmt.Sprintf("slip%d.zip", i)
		writeZip(name, map[string]string{evil: "pwned"})
		_, err := runScript(t, `out = compress.extract("`+d+`/`+name+`", "`+d+`/out`+fmt.Sprint(i)+`")`, fsPerms(dir), Options{})
		if err == nil || !strings.Contains(err.Error(), "unsafe archive entry") {
			t.Errorf("%q: want rejection, got %v", evil, err)
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "evil.txt")); err == nil {
		t.Fatal("zip-slip wrote outside the destination")
	}
	// tar: traversal, plus a symlink entry that must be skipped
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	tw.WriteHeader(&tar.Header{Name: "../tar-evil.txt", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg})
	tw.Write([]byte("x"))
	tw.Close()
	os.WriteFile(filepath.Join(dir, "slip.tar"), tb.Bytes(), 0o644)
	if _, err := runScript(t, `out = compress.extract("`+d+`/slip.tar", "`+d+`/tout")`, fsPerms(dir), Options{}); err == nil || !strings.Contains(err.Error(), "unsafe archive entry") {
		t.Fatalf("tar slip: %v", err)
	}
	tb.Reset()
	tw = tar.NewWriter(&tb)
	tw.WriteHeader(&tar.Header{Name: "lnk", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink, Mode: 0o777})
	tw.WriteHeader(&tar.Header{Name: "ok.txt", Mode: 0o644, Size: 2, Typeflag: tar.TypeReg})
	tw.Write([]byte("ok"))
	tw.Close()
	os.WriteFile(filepath.Join(dir, "link.tar"), tb.Bytes(), 0o644)
	got := mustRun(t, `out = compress.extract("`+d+`/link.tar", "`+d+`/lout")`, fsPerms(dir), Options{})
	if got != `["ok.txt"]` {
		t.Fatalf("links must be skipped: %s", got)
	}
	// bomb: a small zip of a large file hits the extracted-size cap
	writeZip("bomb.zip", map[string]string{"big.bin": strings.Repeat("\x00", 200000)})
	_, err := runScript(t, `out = compress.extract("`+d+`/bomb.zip", "`+d+`/bomb")`, fsPerms(dir), Options{Limits: Limits{MaxBulkBytes: 10000}})
	if err == nil || !strings.Contains(err.Error(), "larger than 10000") {
		t.Fatalf("bomb: %v", err)
	}
	if ents, _ := os.ReadDir(filepath.Join(dir, "bomb")); len(ents) != 0 {
		t.Fatalf("partial file left behind: %v", ents)
	}
	// too many entries
	many := map[string]string{}
	for i := 0; i < 30; i++ {
		many[fmt.Sprintf("f%d", i)] = "x"
	}
	writeZip("many.zip", many)
	_, err = runScript(t, `out = compress.extract("`+d+`/many.zip", "`+d+`/many")`, fsPerms(dir), Options{Limits: Limits{MaxEntries: 10}})
	if err == nil || !strings.Contains(err.Error(), "more than 10 entries") {
		t.Fatalf("entry cap: %v", err)
	}
}

func TestSQLite(t *testing.T) {
	data := t.TempDir()
	o := Options{DataDir: data}
	script := `
db = sqlite.open("notes")
db.exec("CREATE TABLE IF NOT EXISTS n (id INTEGER PRIMARY KEY, title TEXT, score REAL, blob BLOB, done INTEGER)")
r = db.exec("INSERT INTO n(title, score, blob) VALUES (?, ?, ?)", ["a", 1.5, b"\x01\x02"])
db.exec("INSERT INTO n(title) VALUES (:t)", {"t": "b"})
rows = db.query("SELECT id, title, score, blob FROM n ORDER BY id")
named = db.query("SELECT title FROM n WHERE title = @t", {"t": "b"})
db.transaction(lambda tx: tx.exec("UPDATE n SET done = 1"))
_, rolled = attempt(db.transaction, rollback)
count = db.query("SELECT count(*) AS c FROM n WHERE done = 1")[0]["c"]
out = [r.rows_affected, r.last_insert_id, rows, named, count, rolled != None, db.query("SELECT title FROM n WHERE title = 'zz'")]
`
	body := `def rollback(tx):
    tx.exec("INSERT INTO n(title) VALUES ('ghost')")
    fail("nope")
`
	full := body + "\ndef on_query(ctx, q):\n"
	for _, l := range strings.Split(strings.TrimSpace(script), "\n") {
		full += "    " + l + "\n"
	}
	full += "    return [item(\"1\", str(out))]\n"
	e := openStarted(t, baseManifest, full, nil, o)
	t.Cleanup(e.Stop)
	items, err := e.Query(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	want := `[1, 1, [{"id": 1, "title": "a", "score": 1.5, "blob": b"\x01\x02"}, {"id": 2, "title": "b", "score": None, "blob": None}], [{"title": "b"}], 2, True, []]`
	if items[0].Title != want {
		t.Fatalf("got  %s\nwant %s", items[0].Title, want)
	}
	// the rollback left no ghost row
	got := mustRun(t, `db = sqlite.open("notes")
out = db.query("SELECT count(*) AS c FROM n WHERE title = 'ghost'")[0]["c"]`, nil, o)
	if got != "0" {
		t.Fatalf("rolled back row survived: %s", got)
	}
	e.Stop()
	if _, err := os.Stat(filepath.Join(data, "ext", baseDirName(), "db", "notes.sqlite3")); err != nil {
		t.Fatalf("database file not under the extension data dir: %v", err)
	}
	// persistence across a restart
	e2 := openStarted(t, baseManifest, "def on_query(ctx, q):\n    return [item(\"1\", str(sqlite.open(\"notes\").query(\"SELECT count(*) AS c FROM n\")[0][\"c\"]))]\n", nil, o)
	t.Cleanup(e2.Stop)
	items, err = e2.Query(context.Background(), "")
	if err != nil || items[0].Title != "2" {
		t.Fatalf("persisted rows: %v %v", items, err)
	}
}

func baseDirName() string {
	m, _ := ParseManifest([]byte(baseManifest))
	return m.DirName()
}

func TestSQLiteIsolation(t *testing.T) {
	data := t.TempDir()
	o := Options{DataDir: data}
	outside := filepath.Join(t.TempDir(), "other.db")
	for _, name := range []string{"../x", "a/b", `a\b`, "", "x.db", "C:x", strings.Repeat("a", 65), ".."} {
		if _, err := runScript(t, fmt.Sprintf(`out = sqlite.open(%q)`, name), nil, o); err == nil {
			t.Errorf("name %q must be rejected", name)
		}
	}
	for _, q := range []string{
		fmt.Sprintf("ATTACH DATABASE '%s' AS o", slash(outside)),
		"  attach '" + slash(outside) + "' as o",
		"SELECT 1; ATTACH '" + slash(outside) + "' AS o",
		"/* hi */ ATTACH '" + slash(outside) + "' AS o",
		fmt.Sprintf("VACUUM INTO '%s'", slash(outside)),
		"vacuum main into '" + slash(outside) + "'",
		"PRAGMA max_page_count = 999999999",
		"PRAGMA main.writable_schema = 1",
		"PRAGMA temp_store_directory = 'C:/'",
		"SELECT load_extension('x')",
	} {
		_, err := runScript(t, fmt.Sprintf(`db = sqlite.open()
out = db.exec(%q)`, q), nil, o)
		if err == nil || !strings.Contains(err.Error(), "not allowed") {
			t.Errorf("%q: want not allowed, got %v", q, err)
		}
	}
	if _, err := os.Stat(outside); err == nil {
		t.Fatal("a file appeared outside the extension storage")
	}
	// safe forms still work, including words that merely look dangerous inside literals
	got := mustRun(t, `db = sqlite.open()
db.exec("CREATE TABLE t (note TEXT)")
db.exec("INSERT INTO t VALUES ('ATTACH DATABASE x; VACUUM INTO y')")
db.exec("PRAGMA user_version = 3")
out = [db.query("SELECT note FROM t")[0]["note"], db.query("PRAGMA user_version")[0]["user_version"], len(db.query("PRAGMA table_info(t)"))]`, nil, o)
	if got != `["ATTACH DATABASE x; VACUUM INTO y", 3, 1]` {
		t.Fatal(got)
	}
	// ATTACH is also disabled in the engine, should the text guard ever miss a spelling
	ex := &Extension{opts: &Options{}}
	ex.opts.defaults()
	d, err := ex.openSQLite(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer d.close()
	if _, err := d.conn.ExecContext(context.Background(), "ATTACH DATABASE '"+slash(outside)+"' AS o"); err == nil {
		t.Fatal("engine-level ATTACH must fail")
	}
}

func TestSQLiteLimitsAndTransactionRules(t *testing.T) {
	o := Options{DataDir: t.TempDir(), Limits: Limits{MaxDBBytes: 64 * 1024, MaxEntries: 20}}
	_, err := runScript(t, `db = sqlite.open()
db.exec("CREATE TABLE t (b BLOB)")
for i in range(100):
    db.exec("INSERT INTO t VALUES (?)", [b"x" * 4000])
out = "unreachable"`, nil, o)
	if err == nil || !strings.Contains(err.Error(), "full") {
		t.Fatalf("db size cap: %v", err)
	}
	_, err = runScript(t, `db = sqlite.open("rows")
db.exec("CREATE TABLE t (n INTEGER)")
for i in range(30):
    db.exec("INSERT INTO t VALUES (?)", [i])
out = db.query("SELECT n FROM t")`, nil, o)
	if err == nil || !strings.Contains(err.Error(), "more than 20 rows") {
		t.Fatalf("row cap: %v", err)
	}
	_, err = runScript(t, `db = sqlite.open("tx")
def f(tx):
    db.query("SELECT 1")
out = db.transaction(f)`, nil, o)
	if err == nil || !strings.Contains(err.Error(), "use the tx argument") {
		t.Fatalf("db inside tx: %v", err)
	}
	if _, err := runScript(t, `db = sqlite.open("bad")
out = db.query("SELEC nonsense")`, nil, o); err == nil {
		t.Fatal("sql errors must raise")
	}
	if _, err := runScript(t, `db = sqlite.open("p")
out = db.query("SELECT ?", [{}])`, nil, o); err == nil || !strings.Contains(err.Error(), "cannot bind") {
		t.Fatalf("param type: %v", err)
	}
}

func TestSQLiteInMemoryWithoutDataDir(t *testing.T) {
	got := mustRun(t, `db = sqlite.open()
db.exec("CREATE TABLE t (x)")
db.exec("INSERT INTO t VALUES (1)")
out = db.query("SELECT x FROM t")`, nil, Options{})
	if got != `[{"x": 1}]` {
		t.Fatal(got)
	}
}

func TestBinaryHelpers(t *testing.T) {
	got := mustRun(t, `
rec = binary.pack(">HI4s?", 1, 4096, "ZAGR", True)
a, b, c, d = binary.unpack(">HI4s?", rec)
le = binary.pack("<hqf", -2, -3, 1.5)
out = [binary.hex(rec), binary.calcsize(">HI4s?"), a, b, c, d, binary.unpack("<hqf", le), binary.hex(le, sep = ":"),
       binary.unpack_from("<H", b"\xff\x01\x02\x00", 2), binary.pack("<2B3x", 1, 2) == b"\x01\x02\x00\x00\x00",
       binary.unpack("<Q", b"\xff" * 8)[0]]`, nil, Options{})
	if !strings.HasPrefix(got, `["0001000010005a41475201", 11, 1, 4096, b"ZAGR", True, (-2, -3, 1.5), "fe:ff:fd:ff:ff:ff:ff:ff:ff:ff:00:00:c0:3f"`) {
		t.Fatal(got)
	}
	if !strings.HasSuffix(got, `, (2,), True, 18446744073709551615]`) {
		t.Fatal(got)
	}
	got = mustRun(t, `
out = [binary.hex(binary.unhex("01 AB:ff-00")), binary.to_int(b"\x01\x00"), binary.to_int(b"\x01\x00", "little"), binary.to_int(b"\xff\xfe", signed = True),
       binary.hex(binary.from_int(258, 2)), binary.hex(binary.from_int(-2, 2, "little", True)), binary.xor(b"abc", b"\x01") == b"`+"`"+`cb",
       binary.chunks(b"abcdefg", 3), binary.reverse(b"abc"), binary.to_int(binary.from_int((1 << 100), 13))  == (1 << 100)]`, nil, Options{})
	if got != `["01abff00", 256, 1, -2, "0102", "feff", True, [b"abc", b"def", b"g"], b"cba", True]` {
		t.Fatal(got)
	}
	for _, body := range []string{
		`out = binary.pack("<B", 256)`, `out = binary.pack("<b", 128)`, `out = binary.pack("<H", -1)`, `out = binary.pack("<I")`,
		`out = binary.pack("<I", 1, 2)`, `out = binary.pack("<z", 1)`, `out = binary.unpack("<I", b"\x01")`,
		`out = binary.from_int(256, 1)`, `out = binary.from_int(-1, 1)`, `out = binary.unhex("zz")`, `out = binary.chunks(b"a", 0)`,
		`out = binary.xor(b"a", b"")`, `out = binary.to_int(b"a", "middle")`, `out = binary.pack("<c", b"ab")`,
	} {
		if _, err := runScript(t, body, nil, Options{}); err == nil {
			t.Errorf("%s should fail", body)
		}
	}
}

func TestNewAPIsCallableWithoutPermissionWhereNone(t *testing.T) {
	// compress bytes, binary and sqlite are always available; files and network are not.
	if _, err := runScript(t, `out = compress.create("/x.zip", ["/y"])`, nil, Options{}); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("create without permission: %v", err)
	}
}
