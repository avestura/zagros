package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
		ok   bool
	}{
		{"0.2.0", "0.3.0", -1, true},
		{"v0.2.0", "0.2.0", 0, true},
		{"1.0.0", "0.9.9", 1, true},
		{"0.2.0-rc.1", "0.2.0", -1, true},
		{"0.2.0-rc.2", "0.2.0-rc.10", -1, true},
		{"v0.2.0-3-gabc1234", "0.2.0", 1, true},        // 3 commits after the tag
		{"v0.2.0-3-gabc1234-dirty", "0.3.0", -1, true}, // still older than the next release
		{"dev", "0.2.0", 0, false},
		{"", "0.2.0", 0, false},
	}
	for _, c := range cases {
		got, ok := CompareVersions(c.a, c.b)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("CompareVersions(%q,%q) = %d,%v want %d,%v", c.a, c.b, got, ok, c.want, c.ok)
		}
	}
}

func server(t *testing.T, status int, body string, hits *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			*hits++
		}
		if !strings.HasSuffix(r.URL.Path, "/releases/latest") && !strings.Contains(r.URL.RawQuery, "per_page") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

const latestJSON = `{"tag_name":"v0.3.0","html_url":"https://github.com/avestura/zagros/releases/tag/v0.3.0","body":"notes",
"published_at":"2026-10-20T10:00:00Z","assets":[
{"name":"zagros_0.3.0_windows_amd64.zip","browser_download_url":"https://example/amd64.zip","size":10},
{"name":"zagros_0.3.0_windows_arm64.zip","browser_download_url":"https://example/arm64.zip","size":11},
{"name":"checksums.txt","browser_download_url":"https://example/sums","size":1}]}`

func TestCheckAvailableAndAssets(t *testing.T) {
	srv := server(t, 200, latestJSON, nil)
	defer srv.Close()
	res, err := Check(context.Background(), "0.2.0", Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Available || res.Latest.Version != "0.3.0" || res.Dev || res.Ahead {
		t.Fatalf("result: %+v", res)
	}
	if a, ok := AssetFor(res.Latest, "arm64"); !ok || a.URL != "https://example/arm64.zip" {
		t.Fatalf("asset: %+v %v", a, ok)
	}
	if _, ok := AssetFor(res.Latest, "386"); ok {
		t.Fatal("no 386 asset expected")
	}
	if !strings.Contains(res.Summary(), "0.3.0 is available") {
		t.Fatalf("summary: %s", res.Summary())
	}
}

func TestCheckUpToDateAheadAndDev(t *testing.T) {
	srv := server(t, 200, latestJSON, nil)
	defer srv.Close()
	for _, c := range []struct {
		cur               string
		avail, ahead, dev bool
	}{
		{"0.3.0", false, false, false},
		{"0.4.0", false, true, false},
		{"dev", false, false, true},
		{"", false, false, true},
	} {
		res, err := Check(context.Background(), c.cur, Options{BaseURL: srv.URL})
		if err != nil {
			t.Fatal(err)
		}
		if res.Available != c.avail || res.Ahead != c.ahead || res.Dev != c.dev {
			t.Errorf("current %q: %+v", c.cur, res)
		}
	}
}

func TestCheckErrors(t *testing.T) {
	for status, want := range map[int]string{404: "no published release", 403: "rate limit", 500: "500"} {
		srv := server(t, status, `{}`, nil)
		_, err := Check(context.Background(), "0.2.0", Options{BaseURL: srv.URL})
		srv.Close()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("status %d: err %v, want %q", status, err, want)
		}
	}
	srv := server(t, 200, `not json`, nil)
	defer srv.Close()
	if _, err := Check(context.Background(), "0.2.0", Options{BaseURL: srv.URL}); err == nil {
		t.Error("bad JSON must fail")
	}
}

func TestCheckCachedReusesAndReevaluates(t *testing.T) {
	hits := 0
	srv := server(t, 200, latestJSON, &hits)
	defer srv.Close()
	dir := t.TempDir()
	o := Options{BaseURL: srv.URL}
	r1, err := CheckCached(context.Background(), dir, "0.2.0", time.Hour, o)
	if err != nil || !r1.Available || hits != 1 {
		t.Fatalf("first: %+v err %v hits %d", r1, err, hits)
	}
	r2, _ := CheckCached(context.Background(), dir, "0.2.0", time.Hour, o)
	if hits != 1 || !r2.Available {
		t.Fatalf("second must come from the cache: hits %d %+v", hits, r2)
	}
	r3, _ := CheckCached(context.Background(), dir, "0.3.0", time.Hour, o) // upgraded since
	if hits != 1 || r3.Available {
		t.Fatalf("after upgrading the cached result must say up to date: %+v", r3)
	}
	if _, err := CheckCached(context.Background(), dir, "0.2.0", 0, o); err != nil || hits != 2 {
		t.Fatalf("expired cache must refetch: hits %d err %v", hits, err)
	}
}
