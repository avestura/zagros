package builtin

import (
	"context"
	"testing"
)

func TestPasswordKeywordPrefixes(t *testing.T) {
	p := newPassword()
	for _, q := range []string{"pass", "passw", "passwo", "password", "password 24", "pw"} {
		items := p.Query(context.Background(), q)
		if len(items) == 0 {
			t.Fatalf("%q should generate passwords", q)
		}
		if !items[0].Refresh {
			t.Fatalf("%q: items must be refreshable with Tab", q)
		}
	}
	if len(p.Query(context.Background(), "pas")) != 0 || len(p.Query(context.Background(), "passwords")) != 0 {
		t.Fatal("unrelated words must not trigger the generator")
	}
	if items := p.Query(context.Background(), "passphrase"); len(items) == 0 {
		t.Fatal("passphrase should generate passphrases")
	}
	a, b := p.Query(context.Background(), "password")[0].Title, p.Query(context.Background(), "password")[0].Title
	if a == b {
		t.Fatal("each query must produce new passwords")
	}
}
