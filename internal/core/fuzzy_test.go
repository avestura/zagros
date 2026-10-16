package core

import (
	"reflect"
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pat, text string
		ok        bool
		idx       []int
	}{
		{"", "anything", true, nil},
		{"abc", "ab", false, nil},
		{"xyz", "visual studio", false, nil},
		{"vs", "Visual Studio", true, []int{0, 7}},
		{"code", "Visual Studio Code", true, []int{14, 15, 16, 17}},
		{"CHR", "google chrome", true, []int{7, 8, 9}},
		{"é", "Café", true, []int{3}},
	}
	for _, c := range cases {
		_, idx, ok := FuzzyMatch(c.pat, c.text)
		if ok != c.ok {
			t.Errorf("%q in %q: ok=%v want %v", c.pat, c.text, ok, c.ok)
			continue
		}
		if ok && c.idx != nil && !reflect.DeepEqual(idx, c.idx) {
			t.Errorf("%q in %q: idx=%v want %v", c.pat, c.text, idx, c.idx)
		}
	}
}

func TestFuzzyRanking(t *testing.T) {
	score := func(p, s string) int {
		v, ok := FuzzyScore(p, s)
		if !ok {
			t.Fatalf("%q should match %q", p, s)
		}
		return v
	}
	if score("chr", "Chrome") <= score("chr", "Archive Reader") {
		t.Error("prefix should beat scattered match")
	}
	if score("vs", "Visual Studio") <= score("vs", "Canvas") {
		t.Error("word starts should beat inner matches")
	}
	if score("note", "Notepad") <= score("note", "Another Text Editor") {
		t.Error("consecutive prefix should win")
	}
}

func TestFuzzyScoreAgreesWithMatch(t *testing.T) {
	texts := []string{"Visual Studio Code", "notepad++", "Windows Terminal", "Task Manager"}
	pats := []string{"vsc", "np", "wt", "tm", "term", "zz"}
	for _, x := range texts {
		for _, p := range pats {
			a, ok1 := FuzzyScore(p, x)
			b, _, ok2 := FuzzyMatch(p, x)
			if ok1 != ok2 || (ok1 && a != b) {
				t.Errorf("%q/%q: score %d,%v vs match %d,%v", p, x, a, ok1, b, ok2)
			}
		}
	}
}

func BenchmarkFuzzyScoreASCII(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FuzzyScore("vsc", "Visual Studio Code")
	}
}

func BenchmarkFuzzyMatch(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		FuzzyMatch("vsc", "Visual Studio Code")
	}
}
