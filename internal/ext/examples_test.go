package ext

import (
	"archive/zip"
	"bytes"
	"context"
	"path/filepath"
	"testing"
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
