package toolstest

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
)

func namedArgs(kv []any) map[string]any {
	m := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

// Call runs group.name with alternating name, value pairs.
func Call(group, name string, kv ...any) (any, error) {
	return tools.Call(group, name, namedArgs(kv))
}

// Must runs group.name and fails the test on error.
func Must(t *testing.T, group, name string, kv ...any) any {
	t.Helper()
	v, err := Call(group, name, kv...)
	if err != nil {
		t.Fatalf("%s.%s: %v", group, name, err)
	}
	return v
}

// Compact renders a result as JSON (or plain text) with all whitespace
// outside strings kept as is except newlines and indentation, which are
// removed, so expectations stay on one line.
func Compact(v any) string {
	s := tools.Format(v)
	var b strings.Builder
	in, esc := false, false
	for _, r := range s {
		switch {
		case esc:
			esc = false
		case in && r == '\\':
			esc = true
		case r == '"':
			in = !in
		case !in && (r == '\n' || r == ' ' || r == '\t' || r == '\r'):
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Fails asserts that group.name returns an error.
func Fails(t *testing.T, group, name string, kv ...any) {
	t.Helper()
	if _, err := Call(group, name, kv...); err == nil {
		t.Errorf("%s.%s %v: expected an error", group, name, kv)
	}
}

// JSONValue decodes JSON into the structured values tools functions take and
// return (ordered *tools.Dict, []any, string, int64, float64, bool, nil).
func JSONValue(t *testing.T, s string) any {
	t.Helper()
	v, err := tools.ParseJSON(s)
	if err != nil {
		t.Fatalf("bad test JSON %q: %v", s, err)
	}
	return v
}
