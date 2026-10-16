package builtin

import (
	"strings"
	"testing"
)

func TestParsePwArgs(t *testing.T) {
	p := parsePwArgs("24 nosym noamb")
	if p.Length != 24 || p.Symbols || !p.NoAmbiguous || !p.Lower || !p.Upper || !p.Digits {
		t.Fatalf("unexpected policy %+v", p)
	}
	if p := parsePwArgs("1"); p.Length != 4 {
		t.Errorf("length not clamped up: %d", p.Length)
	}
	if p := parsePwArgs("9999"); p.Length != 128 {
		t.Errorf("length not clamped down: %d", p.Length)
	}
	if p := parsePwArgs("nosym nonum noupper nolower"); !p.Lower {
		t.Error("policy with no classes must fall back to lower")
	}
}

func TestGeneratePassword(t *testing.T) {
	p := parsePwArgs("16 noamb")
	for i := 0; i < 200; i++ {
		s := generatePassword(p)
		if len(s) != 16 {
			t.Fatalf("len %d", len(s))
		}
		if !strings.ContainsAny(s, pwLower) || !strings.ContainsAny(s, pwUpper) ||
			!strings.ContainsAny(s, pwDigits) || !strings.ContainsAny(s, pwSymbols) {
			t.Fatalf("missing a class: %q", s)
		}
		if strings.ContainsAny(s, pwAmbig) {
			t.Fatalf("ambiguous char in %q", s)
		}
	}
	if s := generatePassword(parsePwArgs("30 nosym")); strings.ContainsAny(s, pwSymbols) {
		t.Errorf("nosym violated: %q", s)
	}
}

func TestPinAndPassphrase(t *testing.T) {
	if s := generatePIN(8); len(s) != 8 || strings.Trim(s, pwDigits) != "" {
		t.Errorf("bad pin %q", s)
	}
	if s := generatePassphrase(4, "-", true, true); strings.Count(s, "-") != 4 {
		t.Errorf("bad passphrase %q", s)
	}
	if len(passWords) < 256 {
		t.Errorf("word list too short: %d", len(passWords))
	}
}

func TestPasswordItemsNotInHistory(t *testing.T) {
	items := newPassword().Query(nil, "pw 12")
	if len(items) == 0 {
		t.Fatal("no items")
	}
	for _, it := range items {
		if it.Action.Extra["nohistory"] != "1" {
			t.Error("item must carry nohistory")
		}
	}
}
