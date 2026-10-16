package ui

import "unicode"

// HighlightFunc returns rune indices of text matched by query.
type HighlightFunc func(query, text string) []int

// SubsequenceMatch is the fallback highlighter: it prefers a contiguous
// case-insensitive substring, else the leftmost subsequence. Returns nil when
// the query does not match.
func SubsequenceMatch(query, text string) []int {
	q := []rune(query)
	t := []rune(text)
	if len(q) == 0 {
		return nil
	}
	lower := func(r rune) rune { return unicode.ToLower(r) }
	for i := 0; i+len(q) <= len(t); i++ {
		ok := true
		for j := range q {
			if lower(t[i+j]) != lower(q[j]) {
				ok = false
				break
			}
		}
		if ok {
			idx := make([]int, len(q))
			for j := range idx {
				idx[j] = i + j
			}
			return idx
		}
	}
	var idx []int
	j := 0
	for i := 0; i < len(t) && j < len(q); i++ {
		if lower(t[i]) == lower(q[j]) {
			idx = append(idx, i)
			j++
		}
	}
	if j < len(q) {
		return nil
	}
	return idx
}

// Run is a span of text with a highlight flag.
type Run struct {
	Text string
	Hit  bool
}

// SplitRuns splits text into runs according to matched rune indices
// (ascending).
func SplitRuns(text string, idx []int) []Run {
	rs := []rune(text)
	if len(idx) == 0 {
		return []Run{{Text: text}}
	}
	hit := make([]bool, len(rs))
	for _, i := range idx {
		if i >= 0 && i < len(rs) {
			hit[i] = true
		}
	}
	var out []Run
	start := 0
	for i := 1; i <= len(rs); i++ {
		if i == len(rs) || hit[i] != hit[start] {
			out = append(out, Run{Text: string(rs[start:i]), Hit: hit[start]})
			start = i
		}
	}
	return out
}
