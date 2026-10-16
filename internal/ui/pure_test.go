package ui

import (
	"reflect"
	"testing"

	"github.com/avestura/zagros/internal/theme"
)

func TestParseHotkey(t *testing.T) {
	cases := []struct {
		in        string
		mods, vk  uint32
		wantError bool
	}{
		{"Alt+Space", HotAlt, 0x20, false},
		{"ctrl+shift+k", HotControl | HotShift, 'K', false},
		{"Win+`", HotWin, 0xC0, false},
		{"F9", 0, 0x78, false},
		{"Ctrl+Alt+5", HotControl | HotAlt, '5', false},
		{"Space", 0, 0, true},
		{"", 0, 0, true},
		{"Alt+Foo", 0, 0, true},
		{"Alt+", 0, 0, true},
	}
	for _, c := range cases {
		m, v, err := ParseHotkey(c.in)
		if (err != nil) != c.wantError {
			t.Fatalf("%q: err=%v", c.in, err)
		}
		if err == nil && (m != c.mods || v != c.vk) {
			t.Errorf("%q: got %x %x want %x %x", c.in, m, v, c.mods, c.vk)
		}
	}
}

func TestEditor(t *testing.T) {
	var e Editor
	e.Insert("hello world")
	if e.Text() != "hello world" || e.Caret() != 11 {
		t.Fatal(e.Text(), e.Caret())
	}
	e.CtrlBackspace()
	if e.Text() != "hello " {
		t.Fatalf("%q", e.Text())
	}
	e.Backspace()
	e.Move(-1, true, false)
	if e.Caret() != 0 {
		t.Fatal(e.Caret())
	}
	e.SelectAll()
	e.Insert("x\ny")
	if e.Text() != "x y" {
		t.Fatalf("%q", e.Text())
	}
	e.Home(false)
	e.Move(1, false, true)
	if e.SelectedText() != "x" {
		t.Fatal(e.SelectedText())
	}
	e.Delete()
	if e.Text() != " y" {
		t.Fatalf("%q", e.Text())
	}
	e.End(false)
	e.Insert("é😀")
	e.Backspace()
	if e.Text() != " yé" {
		t.Fatalf("%q", e.Text())
	}
	e.Move(-1, false, false)
	e.CtrlDelete()
	if e.Text() != " y" {
		t.Fatalf("%q", e.Text())
	}
}

func TestMetricsAndHit(t *testing.T) {
	th := NormalizeTheme(theme.Theme{})
	m := ComputeMetrics(th, 192, 8, 20)
	if m.Width != 1280 || m.RowH != 88 || m.InputH != 96 {
		t.Fatalf("%+v", m)
	}
	if m.Height(0) != 2*m.Border+m.InputH {
		t.Fatal("empty height")
	}
	if m.Height(100) != 2*m.Border+m.InputH+8*m.RowH+20 {
		t.Fatal("full height")
	}
	y := m.Border + m.InputH + 2*m.RowH + 1
	if got := m.HitRow(100, y, 5); got != 2 {
		t.Fatal(got)
	}
	if got := m.HitRow(100, y, 2); got != -1 {
		t.Fatal(got)
	}
	if got := m.HitRow(100, 3, 5); got != -1 {
		t.Fatal(got)
	}
	if !m.InputRect().Contains(100, 10) {
		t.Fatal("input rect")
	}
}

func TestScroll(t *testing.T) {
	var s Scroll
	s.Reset(20, 5)
	s.Move(-1, true)
	if s.Sel != 19 || s.Top != 15 {
		t.Fatalf("%+v", s)
	}
	s.Move(1, true)
	if s.Sel != 0 || s.Top != 0 {
		t.Fatalf("%+v", s)
	}
	s.PageDown()
	if s.Sel != 4 || s.Top != 0 {
		t.Fatalf("%+v", s)
	}
	s.PageDown()
	if s.Sel != 8 || s.Top != 4 {
		t.Fatalf("%+v", s)
	}
	s.End()
	s.Home()
	s.Wheel(100)
	if s.Top != 15 {
		t.Fatalf("%+v", s)
	}
	if s.Abs(0) != 15 || s.Abs(5) != -1 || s.Rel(14) != -1 || s.Rel(16) != 1 {
		t.Fatal("abs/rel")
	}
	s.Select(19)
	s.SetCount(3, 3)
	if s.Sel != 2 || s.Top != 0 {
		t.Fatalf("%+v", s)
	}
}

func TestCaretIndexAt(t *testing.T) {
	w := []int{0, 10, 20, 30}
	for x, want := range map[int]int{0: 0, 4: 0, 6: 1, 16: 2, 29: 3, 100: 3} {
		if got := CaretIndexAt(w, x); got != want {
			t.Errorf("x=%d got %d want %d", x, got, want)
		}
	}
	if HScroll(50, 40, 100, 7) != 0 {
		t.Fatal("fits")
	}
	if off := HScroll(300, 320, 100, 0); off != 200 {
		t.Fatal(off)
	}
}

func TestMatchAndRuns(t *testing.T) {
	if got := SubsequenceMatch("cod", "Visual Studio Code"); !reflect.DeepEqual(got, []int{14, 15, 16}) {
		t.Fatal(got)
	}
	if got := SubsequenceMatch("vsc", "Visual Studio Code"); !reflect.DeepEqual(got, []int{0, 2, 14}) {
		t.Fatal(got)
	}
	if SubsequenceMatch("zzz", "abc") != nil {
		t.Fatal("nomatch")
	}
	runs := SplitRuns("abcd", []int{1, 2})
	want := []Run{{"a", false}, {"bc", true}, {"d", false}}
	if !reflect.DeepEqual(runs, want) {
		t.Fatal(runs)
	}
}
