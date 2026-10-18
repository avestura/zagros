package text

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/avestura/zagros/internal/tools"
)

var errTooBig = errors.New("inputs are too large for this comparison")

func fits(n, m int) error {
	if n > 0 && m > 0 && n > maxCells/m {
		return errTooBig
	}
	return nil
}

func levenshtein(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			c := 1
			if a[i-1] == b[j-1] {
				c = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+c)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// osa is the optimal string alignment distance (restricted Damerau-Levenshtein:
// adjacent transpositions count as one edit, no substring is edited twice).
func osa(a, b []rune) int {
	la, lb := len(a), len(b)
	d := make([][]int, 3)
	for i := range d {
		d[i] = make([]int, lb+1)
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	// rows: d[(i-2)%3], d[(i-1)%3], d[i%3]
	for i := 1; i <= la; i++ {
		cur, p1, p2 := d[i%3], d[(i-1)%3], d[(i+1)%3]
		cur[0] = i
		for j := 1; j <= lb; j++ {
			c := 1
			if a[i-1] == b[j-1] {
				c = 0
			}
			v := min(p1[j]+1, cur[j-1]+1, p1[j-1]+c)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				v = min(v, p2[j-2]+1)
			}
			cur[j] = v
		}
	}
	return d[la%3][lb]
}

func jaroWinkler(a, b []rune) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	window := max(max(len(a), len(b))/2-1, 0)
	am, bm := make([]bool, len(a)), make([]bool, len(b))
	matches := 0
	for i := range a {
		lo, hi := max(0, i-window), min(len(b), i+window+1)
		for j := lo; j < hi; j++ {
			if !bm[j] && a[i] == b[j] {
				am[i], bm[j] = true, true
				matches++
				break
			}
		}
	}
	if matches == 0 {
		return 0
	}
	trans, k := 0, 0
	for i := range a {
		if !am[i] {
			continue
		}
		for !bm[k] {
			k++
		}
		if a[i] != b[k] {
			trans++
		}
		k++
	}
	m := float64(matches)
	jaro := (m/float64(len(a)) + m/float64(len(b)) + (m-float64(trans)/2)/m) / 3
	if jaro < 0.7 {
		return jaro
	}
	l := 0
	for l < 4 && l < len(a) && l < len(b) && a[l] == b[l] {
		l++
	}
	return jaro + float64(l)*0.1*(1-jaro)
}

func lcsLen(a, b []rune) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				cur[j] = prev[j-1] + 1
			} else {
				cur[j] = max(prev[j], cur[j-1])
			}
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func longestCommonSubstring(a, b []rune) string {
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	best, end := 0, 0
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				cur[j] = prev[j-1] + 1
				if cur[j] > best {
					best, end = cur[j], i
				}
			} else {
				cur[j] = 0
			}
		}
		prev, cur = cur, prev
	}
	return string(a[end-best : end])
}

func foldStr(s string) string { return strings.ToLower(strings.Map(unicode.ToLower, s)) }

func init() {
	two := func(ex string) []P {
		return []P{ps("a", "First text."), ps("b", "Second text.")}
	}
	reg("levenshtein", "Levenshtein distance", "Edit distance between two texts: the minimum number of single-character insertions, deletions and substitutions (rune-based). Fails for very large inputs (over 50 million table cells).", "int", `text.levenshtein("kitten", "sitting")`, 70, "levenshtein edit distance difference fuzzy compare",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			if err := fits(len(x), len(y)); err != nil {
				return nil, err
			}
			return int64(levenshtein(x, y)), nil
		})
	reg("damerau_levenshtein", "Damerau-Levenshtein distance", "Edit distance that also counts a swap of two adjacent characters as one edit (optimal string alignment variant, rune-based). ca -> ac is 1.", "int", `text.damerau_levenshtein("ca", "ac")`, 71, "damerau levenshtein transposition edit distance typo fuzzy",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			if err := fits(len(x), len(y)); err != nil {
				return nil, err
			}
			return int64(osa(x, y)), nil
		})
	reg("hamming", "Hamming distance", "Number of positions at which two texts of equal length (in characters) differ. Different lengths are an error.", "int", `text.hamming("karolin", "kathrin")`, 72, "hamming distance difference positions bits",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			if len(x) != len(y) {
				return nil, errors.New("hamming distance needs texts of equal length")
			}
			n := 0
			for i := range x {
				if x[i] != y[i] {
					n++
				}
			}
			return int64(n), nil
		})
	reg("jaro_winkler", "Jaro-Winkler similarity", "Jaro-Winkler similarity from 0 (nothing in common) to 1 (identical), favouring a shared prefix of up to 4 characters; two empty texts give 1.", "float", `text.jaro_winkler("martha", "marhta")`, 73, "jaro winkler similarity fuzzy match name",
		two(""), func(a tools.Args) (any, error) { return jaroWinkler([]rune(a.Str("a")), []rune(a.Str("b"))), nil })
	reg("similarity", "Text similarity", "Similarity from 0 to 1 computed as 1 - levenshtein / longer length; two empty texts give 1.", "float", `text.similarity("kitten", "sitting")`, 74, "similarity ratio fuzzy percent match",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			l := max(len(x), len(y))
			if l == 0 {
				return float64(1), nil
			}
			if err := fits(len(x), len(y)); err != nil {
				return nil, err
			}
			return 1 - float64(levenshtein(x, y))/float64(l), nil
		})
	reg("common_prefix", "Common prefix", "Longest prefix shared by both texts (character-based); empty when they differ from the first character.", "str", `text.common_prefix("interview", "internet")`, 75, "common prefix shared start",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			i := 0
			for i < len(x) && i < len(y) && x[i] == y[i] {
				i++
			}
			return string(x[:i]), nil
		})
	reg("common_suffix", "Common suffix", "Longest suffix shared by both texts (character-based); empty when the last characters differ.", "str", `text.common_suffix("testing", "running")`, 76, "common suffix shared end",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			i := 0
			for i < len(x) && i < len(y) && x[len(x)-1-i] == y[len(y)-1-i] {
				i++
			}
			return string(x[len(x)-i:]), nil
		})
	reg("longest_common_substring", "Longest common substring", "Longest contiguous piece of text found in both inputs (case-sensitive, first one in the first text on ties). Fails for very large inputs.", "str", `text.longest_common_substring("abcdef", "zcdemn")`, 77, "longest common substring shared overlap",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			if err := fits(len(x), len(y)); err != nil {
				return nil, err
			}
			return longestCommonSubstring(x, y), nil
		})
	reg("diff_summary", "Diff summary", "Character-level summary of how a becomes b using the longest common subsequence: dict with inserted, deleted, unchanged, distance (inserted + deleted) and similarity (0..1). Fails for very large inputs.", "dict[str, str]", `text.diff_summary("kitten", "sitting")`, 78, "diff summary insert delete changes compare characters lcs",
		two(""), func(a tools.Args) (any, error) {
			x, y := []rune(a.Str("a")), []rune(a.Str("b"))
			if err := fits(len(x), len(y)); err != nil {
				return nil, err
			}
			l := lcsLen(x, y)
			sim := 1.0
			if t := len(x) + len(y); t > 0 {
				sim = 2 * float64(l) / float64(t)
			}
			return map[string]string{
				"inserted": itoa(len(y) - l), "deleted": itoa(len(x) - l), "unchanged": itoa(l),
				"distance": itoa(len(x) + len(y) - 2*l), "similarity": strconv.FormatFloat(sim, 'f', 4, 64),
			}, nil
		})
	reg("compare_ci", "Compare (ignore case)", "Compares two texts ignoring case: returns -1 if a sorts before b, 0 if equal, 1 if after.", "int", `text.compare_ci("Apple", "apple")`, 79, "compare case insensitive ignore case equals order sort",
		two(""), func(a tools.Args) (any, error) {
			return int64(strings.Compare(foldStr(a.Str("a")), foldStr(a.Str("b")))), nil
		})
	reg("equals_ci", "Equals (ignore case)", "True when both texts are equal ignoring case (Unicode simple folding).", "bool", `text.equals_ci("Hello", "hELLO")`, 80, "equals case insensitive ignore case equal fold",
		two(""), func(a tools.Args) (any, error) { return strings.EqualFold(a.Str("a"), a.Str("b")), nil })
}
