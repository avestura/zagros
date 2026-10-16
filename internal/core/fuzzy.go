package core

import (
	"sync"
	"unicode"
)

// Scoring constants for FuzzyMatch.
const (
	scoreMatch       = 16
	bonusConsecutive = 12
	bonusWordStart   = 14
	bonusPrefix      = 20
	bonusExactCase   = 1
	penaltyGap       = 1
	penaltyLead      = 1
	maxLeadPenalty   = 8

	// Inputs up to these sizes use the optimal DP; larger ones fall back to a
	// greedy match.
	maxPattern = 32
	maxText    = 256

	negInf = int32(-1 << 30)
)

type dpScratch struct {
	p     [maxPattern]rune
	t     [maxText]rune
	score [maxPattern * maxText]int32
	from  [maxPattern * maxText]int16
}

var dpPool = sync.Pool{New: func() any { return new(dpScratch) }}

// FuzzyMatch reports whether pattern is a case-insensitive subsequence of
// text. It returns a score (higher is better) and the rune indices in text of
// the best-scoring match, usable for highlighting. An empty pattern matches
// with score 0 and no indices.
func FuzzyMatch(pattern, text string) (score int, idx []int, ok bool) {
	return fuzzy(pattern, text, true)
}

// FuzzyScore is like FuzzyMatch but skips building the index slice and does
// not allocate in the common case.
func FuzzyScore(pattern, text string) (int, bool) {
	s, _, ok := fuzzy(pattern, text, false)
	return s, ok
}

func fuzzy(pattern, text string, wantIdx bool) (int, []int, bool) {
	if pattern == "" {
		return 0, nil, true
	}
	if len(pattern) > len(text)*4 { // cheap reject: more runes than the text can hold
		return 0, nil, false
	}
	sc := dpPool.Get().(*dpScratch)
	defer dpPool.Put(sc)

	p, t := sc.p[:0], sc.t[:0]
	var pr, tr []rune
	small := true
	for _, r := range pattern {
		if len(p) == maxPattern {
			small = false
			break
		}
		p = append(p, r)
	}
	if small {
		for _, r := range text {
			if len(t) == maxText {
				small = false
				break
			}
			t = append(t, r)
		}
	}
	if !small {
		pr, tr = []rune(pattern), []rune(text)
	} else {
		pr, tr = p, t
	}
	if len(pr) > len(tr) || !isSubsequence(pr, tr) {
		return 0, nil, false
	}
	var idx []int
	if wantIdx {
		idx = make([]int, len(pr))
	}
	if small {
		s := fuzzyDP(sc, pr, tr, idx)
		return int(s), idx, true
	}
	return fuzzyGreedy(pr, tr, idx), idx, true
}

func isSubsequence(p, t []rune) bool {
	pi := 0
	for _, r := range t {
		if foldEq(r, p[pi]) {
			pi++
			if pi == len(p) {
				return true
			}
		}
	}
	return false
}

// posBonus is the bonus for matching t[j] given its neighbour.
func posBonus(t []rune, j int) int32 {
	switch {
	case j == 0:
		return bonusPrefix
	case isWordSep(t[j-1]) || (unicode.IsLower(t[j-1]) && unicode.IsUpper(t[j])):
		return bonusWordStart
	}
	return 0
}

// fuzzyDP finds the highest-scoring placement of p in t. score[i][j] is the
// best score with p[i] matched at t[j]; gaps cost a linear penalty, adjacent
// matches earn a bonus. Runs in O(len(p)*len(t)) with no allocation.
func fuzzyDP(sc *dpScratch, p, t []rune, idx []int) int32 {
	n, m := len(p), len(t)
	for i := 0; i < n; i++ {
		row := i * m
		run, runK := negInf, -1 // best previous-row cell at k <= j-2, gap-penalised
		for j := 0; j < m; j++ {
			if i > 0 && j >= 2 {
				run -= penaltyGap
				if c := sc.score[row-m+j-2] - penaltyGap; c > run {
					run, runK = c, j-2
				}
			}
			sc.score[row+j] = negInf
			sc.from[row+j] = -1
			if !foldEq(t[j], p[i]) {
				continue
			}
			s := int32(scoreMatch) + posBonus(t, j)
			if t[j] == p[i] {
				s += bonusExactCase
			}
			if i == 0 {
				sc.score[row+j] = s - int32(min(j, maxLeadPenalty)*penaltyLead)
				continue
			}
			best, from := negInf, -1
			if j >= 1 && sc.score[row-m+j-1] > negInf {
				best, from = sc.score[row-m+j-1]+bonusConsecutive, j-1
			}
			if run > best {
				best, from = run, runK
			}
			if best > negInf/2 {
				sc.score[row+j] = best + s
				sc.from[row+j] = int16(from)
			}
		}
	}
	last := (n - 1) * m
	bestJ := -1
	for j := 0; j < m; j++ {
		if sc.score[last+j] > negInf/2 && (bestJ < 0 || sc.score[last+j] > sc.score[last+bestJ]) {
			bestJ = j
		}
	}
	total := sc.score[last+bestJ]
	if idx != nil {
		j := bestJ
		for i := n - 1; i >= 0; i-- {
			idx[i] = j
			j = int(sc.from[i*m+j])
		}
	}
	return total
}

// fuzzyGreedy is the fallback for very long inputs: leftmost match end,
// tightened backwards, then scored.
func fuzzyGreedy(p, t []rune, idx []int) int {
	pi, end := 0, -1
	for i := 0; i < len(t) && pi < len(p); i++ {
		if foldEq(t[i], p[pi]) {
			pi++
			if pi == len(p) {
				end = i
			}
		}
	}
	pi = len(p) - 1
	start := end
	for i := end; i >= 0; i-- {
		if foldEq(t[i], p[pi]) {
			if pi == 0 {
				start = i
				break
			}
			pi--
		}
	}
	score, prev := 0, -2
	pi = 0
	for i := start; i <= end && pi < len(p); i++ {
		if !foldEq(t[i], p[pi]) {
			continue
		}
		score += scoreMatch + int(posBonus(t, i))
		if t[i] == p[pi] {
			score += bonusExactCase
		}
		if prev == i-1 {
			score += bonusConsecutive
		} else if prev >= 0 {
			score -= penaltyGap * (i - prev - 1)
		}
		if idx != nil {
			idx[pi] = i
		}
		prev = i
		pi++
	}
	return score - min(start, maxLeadPenalty)*penaltyLead
}

func foldEq(a, b rune) bool {
	if a == b {
		return true
	}
	if a < 0x80 && b < 0x80 {
		if 'A' <= a && a <= 'Z' {
			a += 32
		}
		if 'A' <= b && b <= 'Z' {
			b += 32
		}
		return a == b
	}
	return unicode.ToLower(a) == unicode.ToLower(b)
}

func isWordSep(r rune) bool {
	switch r {
	case ' ', '-', '_', '.', '/', '\\', ':', '(', ')', '[', ']':
		return true
	}
	return r >= 0x80 && (unicode.IsSpace(r) || unicode.IsPunct(r))
}
