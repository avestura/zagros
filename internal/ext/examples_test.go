package ext

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The shipped examples must validate, pack and install.
func TestShippedExamples(t *testing.T) {
	for _, name := range []string{"hello", "github-repos"} {
		dir := filepath.Join("..", "..", "examples", "extensions", name)
		if is := Validate(dir); len(is) > 0 {
			t.Errorf("%s: %v", name, is)
		}
		z, err := Pack(dir)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, err := zip.NewReader(bytes.NewReader(z), int64(len(z))); err != nil {
			t.Fatal(err)
		}
		m, _, _ := newMgr(t, AutoAllow{})
		e, err := m.InstallBytes(context.Background(), z, "example")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if name == "hello" {
			items, err := NewProvider(e).Extension().Query(context.Background(), "Ada")
			if err != nil || len(items) != 1 || items[0].Title != "Hello, Ada!" {
				t.Fatalf("hello: %v %v", err, items)
			}
		}
	}
}

type fakeWeb string

func (f fakeWeb) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Header: http.Header{}, Request: req, Body: io.NopCloser(strings.NewReader(string(f)))}, nil
}

// The example shows a loading row while its search runs in a task, then rows.
func TestGithubReposExampleLoadsAsync(t *testing.T) {
	dir := filepath.Join("..", "..", "examples", "extensions", "github-repos")
	e, err := OpenExtension(dir, Options{DataDir: t.TempDir(), HTTPClient: &http.Client{Transport: fakeWeb(
		`{"items":[{"id":1,"full_name":"a/b","description":"d","stargazers_count":3,"html_url":"https://github.com/a/b","clone_url":"https://github.com/a/b.git"}]}`)}})
	if err != nil {
		t.Fatal(err)
	}
	e.SetGranted(e.Manifest.Permissions)
	ctx := context.Background()
	if err := e.Start(ctx); err != nil {
		t.Fatal(err)
	}
	items, err := e.Query(ctx, "starlark")
	if err != nil || len(items) != 1 || !strings.HasSuffix(items[0].ID, ":loading") {
		t.Fatalf("first answer must be the loading row: %v %+v", err, items)
	}
	for i := 0; i < 400; i++ {
		if items, err = e.Query(ctx, "starlark"); err != nil || !strings.HasSuffix(items[0].ID, ":loading") {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err != nil || len(items) != 1 || items[0].Title != "a/b" {
		t.Fatalf("after the task: %v %+v", err, items)
	}
}
