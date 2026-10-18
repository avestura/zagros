package market

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func mkZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for n, c := range files {
		w, _ := zw.Create(n)
		w.Write([]byte(c))
	}
	zw.Close()
	return buf.Bytes()
}

const manifestV1 = "id = \"hello\"\nname = \"Hello\"\nversion = \"1.2.0\"\npermissions = [\"clipboard\", \"net\"]\n"

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	return &Client{
		RegistryURL: srv.URL + "/registry.json",
		APIBase:     srv.URL,
		CacheDir:    t.TempDir(),
		Insecure:    true,
	}
}

func TestParseRef(t *testing.T) {
	cases := []struct {
		in   string
		want Ref
		bad  bool
	}{
		{"hello-world", Ref{Kind: RefRegistry, ID: "hello-world"}, false},
		{"github:a/b", Ref{Kind: RefGitHub, Owner: "a", Repo: "b"}, false},
		{"github:a/b@v1.2.0", Ref{Kind: RefGitHub, Owner: "a", Repo: "b", Tag: "v1.2.0"}, false},
		{"a/b", Ref{Kind: RefGitHub, Owner: "a", Repo: "b"}, false},
		{"https://github.com/a/b", Ref{Kind: RefGitHub, Owner: "a", Repo: "b"}, false},
		{"https://github.com/a/b.git", Ref{Kind: RefGitHub, Owner: "a", Repo: "b"}, false},
		{"https://github.com/a/b/releases/tag/v2", Ref{Kind: RefGitHub, Owner: "a", Repo: "b", Tag: "v2"}, false},
		{"https://github.com/a/b/tree/dev", Ref{Kind: RefGitHub, Owner: "a", Repo: "b", Tag: "dev"}, false},
		{"https://github.com/a/b/releases/download/v1/x.zagros", Ref{Kind: RefURL, URL: "https://github.com/a/b/releases/download/v1/x.zagros"}, false},
		{"https://example.com/x.zagros", Ref{Kind: RefURL, URL: "https://example.com/x.zagros"}, false},
		{"", Ref{}, true},
		{"Bad Id", Ref{}, true},
	}
	for _, c := range cases {
		got, err := ParseRef(c.in)
		if c.bad {
			if err == nil {
				t.Errorf("%q: expected error", c.in)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("%q: got %+v, %v want %+v", c.in, got, err, c.want)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0}, {"v1.2.0", "1.10.0", -1}, {"2.0.0", "1.9.9", 1},
		{"1.0.0-rc.1", "1.0.0", -1}, {"1.0.0-rc.2", "1.0.0-rc.10", -1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1}, {"1.0+build", "1.0.0", 0}, {"junk", "0.0.1", -1},
	}
	for _, c := range cases {
		if got := CompareVersions(c.a, c.b); got != c.want {
			t.Errorf("%s vs %s = %d want %d", c.a, c.b, got, c.want)
		}
	}
}

const regJSON = `{"schema":1,"generated":"x","extensions":[
{"id":"hello","name":"Hello World","description":"says hi","author":"me","repo":"me/hello","tags":["demo"],"latest":"1.2.0","permissions":["net"],"verified":true},
{"id":"calc2","name":"Better Calc","description":"math","author":"you","repo":"you/calc2","tags":["math"],"latest":"0.9.0","permissions":[],"verified":false}]}`

func TestListCachingAndOffline(t *testing.T) {
	var hits, notMod int32
	online := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !online {
			http.Error(w, "down", 500)
			return
		}
		atomic.AddInt32(&hits, 1)
		if r.Header.Get("If-None-Match") == `"v1"` {
			atomic.AddInt32(&notMod, 1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		fmt.Fprint(w, regJSON)
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	ctx := context.Background()

	reg, err := c.List(ctx)
	if err != nil || len(reg.Extensions) != 2 || reg.Stale {
		t.Fatalf("first list: %+v %v", reg, err)
	}
	reg, err = c.List(ctx)
	if err != nil || len(reg.Extensions) != 2 || notMod != 1 {
		t.Fatalf("second list should use 304: %+v %v notMod=%d", reg, err, notMod)
	}
	online = false
	reg, err = c.List(ctx)
	if err != nil || !reg.Stale || len(reg.Extensions) != 2 {
		t.Fatalf("offline fallback: %+v %v", reg, err)
	}
	// No cache and no network -> error.
	c2 := newTestClient(t, srv)
	if _, err := c2.List(ctx); err == nil {
		t.Fatal("expected error without cache")
	}
}

func TestSearchAndUpdates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, regJSON) }))
	defer srv.Close()
	c := newTestClient(t, srv)
	if c.Search("x") != nil {
		t.Fatal("search before list should be nil")
	}
	if _, err := c.List(context.Background()); err != nil {
		t.Fatal(err)
	}
	if r := c.Search("hel"); len(r) != 1 || r[0].ID != "hello" {
		t.Errorf("search hel: %+v", r)
	}
	if r := c.Search("math"); len(r) != 1 || r[0].ID != "calc2" {
		t.Errorf("search tag: %+v", r)
	}
	if r := c.Search(""); len(r) != 2 {
		t.Errorf("empty search: %+v", r)
	}
	ups, err := c.CheckUpdates(context.Background(), map[string]string{"hello": "1.1.9", "calc2": "0.9.0", "other": "1.0.0"})
	if err != nil || len(ups) != 1 || ups[0].ID != "hello" || ups[0].Latest != "1.2.0" {
		t.Errorf("updates: %+v %v", ups, err)
	}
}

// ghServer emulates the slice of the GitHub API the client uses.
func ghServer(t *testing.T, asset []byte, zipball []byte) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/registry.json":
			fmt.Fprint(w, regJSON)
		case strings.HasSuffix(r.URL.Path, "/releases/latest") && asset != nil, strings.Contains(r.URL.Path, "/releases/tags/v1.2.0"):
			fmt.Fprintf(w, `{"tag_name":"v1.2.0","assets":[{"name":"src.zip","browser_download_url":"%s/dl/src.zip"},{"name":"hello.zagros","browser_download_url":"%s/dl/hello.zagros"}]}`, srv.URL, srv.URL)
		case strings.Contains(r.URL.Path, "/releases/"):
			http.NotFound(w, r)
		case r.URL.Path == "/dl/hello.zagros":
			w.Write(asset)
		case r.URL.Path == "/dl/src.zip":
			w.Write(mkZip(t, map[string]string{"README": "x"}))
		case strings.Contains(r.URL.Path, "/zipball"):
			if r.Header.Get("Accept") == "application/octet-stream" {
				http.Error(w, "unsupported", http.StatusUnsupportedMediaType)
				return
			}
			w.Write(zipball)
		case r.URL.Path == "/redir":
			http.Redirect(w, r, "http://evil.example.com/x.zagros", http.StatusFound)
		case r.URL.Path == "/big":
			w.Write(make([]byte, 2048))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv
}

func TestResolveReleaseAsset(t *testing.T) {
	asset := mkZip(t, map[string]string{"manifest.toml": manifestV1, "main.star": "x"})
	srv := ghServer(t, asset, nil)
	defer srv.Close()
	c := newTestClient(t, srv)
	for _, ref := range []string{"github:me/hello", "github:me/hello@v1.2.0", "hello"} {
		pkg, err := c.Resolve(context.Background(), ref)
		if err != nil {
			t.Fatalf("%s: %v", ref, err)
		}
		if pkg.ID != "hello" || pkg.Version != "1.2.0" || pkg.SHA256 != sum256(asset) || pkg.Source != "github:me/hello@v1.2.0" {
			t.Errorf("%s: %+v", ref, pkg)
		}
		if len(pkg.Manifest.Permissions) != 2 {
			t.Errorf("permissions: %v", pkg.Manifest.Permissions)
		}
	}
}

func TestResolveZipballFallbackSubdir(t *testing.T) {
	zb := mkZip(t, map[string]string{
		"me-hello-abc123/README.md":           "r",
		"me-hello-abc123/ext/manifest.toml":   manifestV1,
		"me-hello-abc123/ext/main.star":       "print(1)",
		"me-hello-abc123/other/manifest.toml": "id = \"other\"\n",
	})
	srv := ghServer(t, nil, zb)
	defer srv.Close()
	c := newTestClient(t, srv)

	if _, err := c.Resolve(context.Background(), "github:me/hello"); err == nil {
		t.Error("manifest is not at root; expected error")
	}
	// Subdir comes from the registry entry.
	reg := `{"schema":1,"extensions":[{"id":"hello","name":"H","repo":"me/hello","subdir":"ext"}]}`
	c.RegistryURL = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, reg) })).URL + "/r.json"
	pkg, err := c.Resolve(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.ID != "hello" || pkg.SHA256 != sum256(zb) {
		t.Errorf("%+v", pkg)
	}
	m, err := inspectZip(pkg.Zip)
	if err != nil || m.Version != "1.2.0" {
		t.Errorf("repacked zip: %+v %v", m, err)
	}
}

func TestRepackRootManifest(t *testing.T) {
	zb := mkZip(t, map[string]string{"x-1/manifest.toml": manifestV1, "x-1/a/b.star": "1"})
	out, mf, err := repack(zb, "")
	if err != nil || mf.ID != "hello" {
		t.Fatal(err)
	}
	zr, _ := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["manifest.toml"] || !names["a/b.star"] {
		t.Errorf("names: %v", names)
	}
}

func TestPinnedChecksum(t *testing.T) {
	asset := mkZip(t, map[string]string{"manifest.toml": manifestV1})
	srv := ghServer(t, asset, nil)
	defer srv.Close()
	c := newTestClient(t, srv)
	mk := func(pin string) {
		reg := fmt.Sprintf(`{"schema":1,"extensions":[{"id":"hello","name":"H","repo":"me/hello","sha256":"%s"}]}`, pin)
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, reg) }))
		t.Cleanup(s.Close)
		c.RegistryURL = s.URL
		c.CacheDir = t.TempDir()
	}
	mk(strings.Repeat("0", 64))
	if _, err := c.Resolve(context.Background(), "hello"); !errors.Is(err, ErrChecksum) {
		t.Errorf("want ErrChecksum, got %v", err)
	}
	mk(sum256(asset))
	if _, err := c.Resolve(context.Background(), "hello"); err != nil {
		t.Errorf("good pin: %v", err)
	}
}

func TestDownloadLimitsAndAllowlist(t *testing.T) {
	srv := ghServer(t, nil, nil)
	defer srv.Close()
	c := newTestClient(t, srv)
	c.MaxBytes = 1024
	if _, err := c.download(context.Background(), srv.URL+"/big"); !errors.Is(err, ErrTooLarge) {
		t.Errorf("want ErrTooLarge, got %v", err)
	}
	// Redirect to a non-allowed host is refused.
	if _, err := c.download(context.Background(), srv.URL+"/redir"); err == nil || !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("redirect: %v", err)
	}
	// Direct URL on a non-allowed host.
	if _, err := c.Resolve(context.Background(), "https://evil.example.com/x.zagros"); !errors.Is(err, ErrHostNotAllowed) {
		t.Errorf("host: %v", err)
	}
	// http is refused when not Insecure.
	c.Insecure = false
	if _, err := c.download(context.Background(), srv.URL+"/big"); !errors.Is(err, ErrHostNotAllowed) {
		t.Errorf("http scheme: %v", err)
	}
}

func TestResolveDirectURLAndInstall(t *testing.T) {
	asset := mkZip(t, map[string]string{"manifest.toml": manifestV1})
	srv := ghServer(t, asset, nil)
	defer srv.Close()
	c := newTestClient(t, srv)
	inst := &fakeInstaller{}
	pkg, err := c.Install(context.Background(), inst, srv.URL+"/dl/hello.zagros")
	if err != nil || pkg.ID != "hello" || inst.src != srv.URL+"/dl/hello.zagros" || !bytes.Equal(inst.zip, asset) {
		t.Fatalf("%+v %v %+v", pkg, err, inst)
	}
}

type fakeInstaller struct {
	zip []byte
	src string
}

func (f *fakeInstaller) Install(_ context.Context, z []byte, source string) (string, error) {
	f.zip, f.src = z, source
	return "hello", nil
}
func (f *fakeInstaller) Installed() map[string]string { return map[string]string{} }

func TestParseSpecAndManifest(t *testing.T) {
	s, err := ParseSpec([]byte("repo = \"a/b\" # c\nsubdir = \"/ext/\"\ntags = [\n \"x\", # note\n \"y\",\n]\nverified = true\n"), "hello")
	if err != nil || s.Repo != "a/b" || s.Subdir != "ext" || len(s.Tags) != 2 || !s.Verified {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := ParseSpec([]byte("repo = \"nope\""), "x"); err == nil {
		t.Error("bad repo accepted")
	}
	if _, err := ParseSpec([]byte("id = \"other\"\nrepo = \"a/b\""), "x"); err == nil {
		t.Error("id/file mismatch accepted")
	}
	m, err := ParseManifest([]byte("id = \"p\"\nversion = \"v2.0.0\"\n[permissions]\nnet = [\"a.com\"]\nclipboard = true\nfs = false\n"))
	if err != nil || m.Version != "2.0.0" || strings.Join(m.Permissions, ",") != "clipboard,net" {
		t.Errorf("%+v %v", m, err)
	}
}

func TestBuildEntryAndOfflineSpecs(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			fmt.Fprint(w, `{"tag_name":"v1.3.0","assets":[{"name":"a.zagros","download_count":7},{"name":"b.zip","download_count":3}]}`)
		case strings.Contains(r.URL.Path, "/contents/manifest.toml"):
			if r.URL.Query().Get("ref") != "v1.3.0" {
				http.NotFound(w, r)
				return
			}
			fmt.Fprint(w, "id = \"hello\"\nname = \"Hello\"\nversion = \"1.3.0\"\nauthor = \"me\"\npermissions = [\"net\"]\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := newTestClient(t, srv)
	e, err := c.BuildEntry(context.Background(), Spec{ID: "hello", Repo: "me/hello", Tags: []string{"x"}, Verified: true})
	if err != nil || e.Latest != "1.3.0" || e.Downloads != 10 || e.Name != "Hello" || len(e.Permissions) != 1 || !e.Verified {
		t.Fatalf("%+v %v", e, err)
	}
	if _, err := c.BuildEntry(context.Background(), Spec{ID: "other", Repo: "me/hello"}); err == nil {
		t.Error("id mismatch accepted")
	}
	specs, err := LoadSpecs("../../registry/extensions")
	if err != nil || len(specs) == 0 {
		t.Fatalf("%v %v", specs, err)
	}
	for _, s := range specs {
		if _, err := EntryFromSpec(s); err != nil {
			t.Errorf("%s: %v", s.ID, err)
		}
	}
}

func TestRegistryURLMayRedirectAndUpgradesHTTP(t *testing.T) {
	c := New("", "https://old.example.com/registry.json")
	root, _ := http.NewRequest("GET", c.RegistryURL, nil)
	hop, _ := http.NewRequest("GET", "http://new.example.com/zagros/registry.json", nil)
	if !c.registryRedirect(hop, []*http.Request{root}) || hop.URL.String() != "https://new.example.com/zagros/registry.json" {
		t.Fatalf("a registry redirect to http must be upgraded to https and followed: %v", hop.URL)
	}
	other, _ := http.NewRequest("GET", "https://old.example.com/other.zagros", nil)
	hop2, _ := http.NewRequest("GET", "https://evil.example.net/x", nil)
	if c.registryRedirect(hop2, []*http.Request{other}) {
		t.Fatal("redirects of other downloads stay on the allowlist")
	}
}
