//go:build windows

package ui

import (
	"strings"
	"testing"
)

func TestDefaultExplanationDatabaseAndDirectory(t *testing.T) {
	for name, want := range map[string]string{
		"sql":        "database servers",
		"ldap":       "Search the listed directory",
		"ldap.write": "Add, change or delete entries",
	} {
		got := DefaultExplanation(name)
		if !strings.Contains(got, want) {
			t.Errorf("%s: %q lacks %q", name, got, want)
		}
		if strings.Contains(got, "Use the") || strings.Contains(got, "files") {
			t.Errorf("%s fell through to a generic text: %q", name, got)
		}
	}
}
