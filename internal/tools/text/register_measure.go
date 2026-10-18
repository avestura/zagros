package text

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/text/unicode/norm"
)

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	s = normNL(s)
	n := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		n++
	}
	return n
}

func isTerm(r rune) bool {
	return r == '.' || r == '!' || r == '?' || r == '…' || r == '。' || r == '！' || r == '？'
}

func sentenceCount(s string) int {
	rs := []rune(s)
	n := 0
	content := false
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		if isTerm(r) {
			j := i
			for j+1 < len(rs) && (isTerm(rs[j+1]) || rs[j+1] == '"' || rs[j+1] == '\'' || rs[j+1] == ')' || rs[j+1] == '”' || rs[j+1] == '’') {
				j++
			}
			if content && (j+1 >= len(rs) || unicode.IsSpace(rs[j+1]) || r == '。' || r == '！' || r == '？') {
				n++
				content = false
			}
			i = j
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			content = true
		}
	}
	if content {
		n++
	}
	return n
}

func itoa(n int) string { return strconv.Itoa(n) }

func stats(s string) map[string]string {
	ws := words(s)
	longest := ""
	total := 0
	for _, w := range ws {
		l := utf8.RuneCountInString(w)
		total += l
		if l > utf8.RuneCountInString(longest) {
			longest = w
		}
	}
	avg := 0.0
	if len(ws) > 0 {
		avg = float64(total) / float64(len(ws))
	}
	uniq := map[rune]bool{}
	for _, r := range s {
		uniq[r] = true
	}
	secs := int(math.Round(float64(len(ws)) * 60 / 200))
	rt := fmt.Sprintf("%d s", secs)
	if secs >= 60 {
		rt = fmt.Sprintf("%d min %d s", secs/60, secs%60)
	}
	return map[string]string{
		"chars": itoa(utf8.RuneCountInString(s)), "bytes": itoa(len(s)), "graphemes": itoa(len(graphemes(s))),
		"words": itoa(len(ws)), "lines": itoa(lineCount(s)), "sentences": itoa(sentenceCount(s)),
		"unique_chars": itoa(len(uniq)), "avg_word_length": strconv.FormatFloat(avg, 'f', 2, 64),
		"longest_word": longest, "reading_time_seconds": itoa(secs), "reading_time": rt,
	}
}

func cleanForPalin(s string, ignoreCase, alnum bool) []rune {
	var out []rune
	for _, r := range norm.NFC.String(s) {
		if alnum && !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			continue
		}
		if ignoreCase {
			r = unicode.ToLower(r)
		}
		out = append(out, r)
	}
	return out
}

func isPalin(rs []rune) bool {
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		if rs[i] != rs[j] {
			return false
		}
	}
	return true
}

func reverseGraphemes(s string) string {
	g := graphemes(s)
	for i, j := 0, len(g)-1; i < j; i, j = i+1, j-1 {
		g[i], g[j] = g[j], g[i]
	}
	return strings.Join(g, "")
}

// longestPalin uses Manacher's algorithm (O(n)).
func longestPalin(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return ""
	}
	t := make([]rune, 0, 2*len(rs)+1)
	t = append(t, -1)
	for _, r := range rs {
		t = append(t, r, -1)
	}
	p := make([]int, len(t))
	c, rgt := 0, 0
	best, bc := 0, 0
	for i := range t {
		if i < rgt {
			p[i] = min(rgt-i, p[2*c-i])
		}
		for i-p[i]-1 >= 0 && i+p[i]+1 < len(t) && t[i-p[i]-1] == t[i+p[i]+1] {
			p[i]++
		}
		if i+p[i] > rgt {
			c, rgt = i, i+p[i]
		}
		if p[i] > best {
			best, bc = p[i], i
		}
	}
	start := (bc - best) / 2
	return string(rs[start : start+best])
}

func sortedRunes(rs []rune) string {
	sort.Slice(rs, func(i, j int) bool { return rs[i] < rs[j] })
	return string(rs)
}

func letterKey(s string, ignoreCase bool) string {
	var rs []rune
	for _, r := range norm.NFC.String(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if ignoreCase {
				r = unicode.ToLower(r)
			}
			rs = append(rs, r)
		}
	}
	return sortedRunes(rs)
}

func allRunes(s string, f func(rune) bool) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !f(r) {
			return false
		}
	}
	return true
}

func isVowel(r rune) bool {
	r = unicode.ToLower(r)
	if strings.ContainsRune("aeiou", r) {
		return true
	}
	if r > 127 && unicode.IsLetter(r) {
		for _, b := range norm.NFD.String(string(r)) {
			return strings.ContainsRune("aeiou", b)
		}
	}
	return false
}

func freq(s string, space bool) (map[string]int, []string) {
	m := map[string]int{}
	var order []string
	for _, g := range graphemes(s) {
		if !space && strings.TrimSpace(g) == "" {
			continue
		}
		if _, ok := m[g]; !ok {
			order = append(order, g)
		}
		m[g]++
	}
	return m, order
}

func init() {
	reg("length", "Character count", "Number of characters (Unicode code points / runes) in the text. Emoji with modifiers or combining marks count each code point; see grapheme_length.", "int", `text.length("héllo")`, 20, "length len count characters chars size strlen",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(utf8.RuneCountInString(a.Str("s"))), nil })
	reg("byte_length", "Byte count (UTF-8)", "Number of bytes of the text encoded as UTF-8.", "int", `text.byte_length("héllo")`, 21, "bytes size utf8 byte length",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(len(a.Str("s"))), nil })
	reg("grapheme_length", "Grapheme count", "Number of user-perceived characters. Approximation of extended grapheme clusters: a base rune plus following combining marks, variation selectors, skin-tone modifiers, tag characters and ZWJ-joined runes forms one grapheme; regional-indicator pairs (flags) and CRLF are one grapheme.", "int", `text.grapheme_length("é👨‍👩‍👧")`, 22, "grapheme cluster visible characters emoji length perceived",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(len(graphemes(a.Str("s")))), nil })
	reg("word_count", "Word count", "Counts words: runs of letters, digits and combining marks, which may contain apostrophes or hyphens between them (don't and well-known are one word each). Empty or punctuation-only text gives 0.", "int", `text.word_count("Hello, big world!")`, 8, "words count wc number of words",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(len(words(a.Str("s")))), nil })
	reg("line_count", "Line count", "Number of lines. CRLF, LF and lone CR all separate lines; a single trailing line break does not add an empty line; empty text has 0 lines.", "int", `text.line_count("a\nb\nc")`, 23, "lines count wc -l rows",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(lineCount(a.Str("s"))), nil })
	reg("sentence_count", "Sentence count", "Counts sentences ended by . ! ? or ... followed by whitespace or the end; trailing text without a terminator counts as one sentence. Decimals like 3.14 do not split.", "int", `text.sentence_count("Hi. How are you? Fine")`, 24, "sentences count periods",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return int64(sentenceCount(a.Str("s"))), nil })
	reg("char_count", "Count occurrences", "Counts non-overlapping occurrences of a character or substring. An empty needle is an error.", "int", `text.char_count("banana", "a")`, 25, "count occurrences substring character frequency how many",
		[]P{ps("s", "Text to search."), ps("ch", "Character or substring to count."), pb("ignore_case", "Compare case-insensitively.", false)},
		func(a tools.Args) (any, error) {
			s, ch := a.Str("s"), a.Str("ch")
			if ch == "" {
				return nil, errors.New("ch must not be empty")
			}
			if a.Bool("ignore_case") {
				s, ch = strings.ToLower(s), strings.ToLower(ch)
			}
			return int64(strings.Count(s, ch)), nil
		})
	reg("stats", "Text statistics", "Returns a dict with chars, bytes, graphemes, words, lines, sentences, unique_chars, avg_word_length, longest_word, reading_time_seconds and reading_time (at 200 words per minute).", "dict[str, str]", `text.stats("Hello world. Bye.")`, 0, "statistics stats analyze summary counts reading time wc info",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return stats(a.Str("s")), nil })

	// predicates and analysis
	reg("is_palindrome", "Is palindrome", "True when the text reads the same forwards and backwards. By default case and every non-letter/non-digit are ignored; empty text counts as a palindrome.", "bool", `text.is_palindrome("A man, a plan, a canal: Panama")`, 30, "palindrome check mirror same backwards",
		[]P{ps("s", "Input text."), pb("ignore_case", "Ignore letter case.", true), pb("ignore_non_alnum", "Ignore spaces, punctuation and symbols.", true)},
		func(a tools.Args) (any, error) {
			return isPalin(cleanForPalin(a.Str("s"), a.Bool("ignore_case"), a.Bool("ignore_non_alnum"))), nil
		})
	reg("palindrome_report", "Palindrome report", "Returns a dict with is_palindrome, cleaned (the compared text), reversed and longest_substring (longest palindromic substring of the cleaned text).", "dict[str, str]", `text.palindrome_report("Never odd or even")`, 1, "palindrome report check reversed cleaned",
		[]P{ps("s", "Input text."), pb("ignore_case", "Ignore letter case.", true), pb("ignore_non_alnum", "Ignore spaces, punctuation and symbols.", true)},
		func(a tools.Args) (any, error) {
			rs := cleanForPalin(a.Str("s"), a.Bool("ignore_case"), a.Bool("ignore_non_alnum"))
			c := string(rs)
			return map[string]string{"is_palindrome": bstr(isPalin(rs)), "cleaned": c, "reversed": reverseGraphemes(c), "longest_substring": longestPalin(c)}, nil
		})
	reg("longest_palindromic_substring", "Longest palindromic substring", "Returns the longest substring (case-sensitive, rune-based, first one on ties) that reads the same in both directions; empty text returns empty.", "str", `text.longest_palindromic_substring("babad")`, 31, "palindrome longest substring manacher",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return longestPalin(a.Str("s")), nil })
	reg("is_anagram", "Is anagram", "True when both texts use the same letters and digits the same number of times. Spaces and punctuation are ignored; case is ignored by default.", "bool", `text.is_anagram("Listen", "Silent")`, 32, "anagram check same letters permutation",
		[]P{ps("a", "First text."), ps("b", "Second text."), pb("ignore_case", "Ignore letter case.", true)},
		func(a tools.Args) (any, error) {
			ic := a.Bool("ignore_case")
			return letterKey(a.Str("a"), ic) == letterKey(a.Str("b"), ic), nil
		})
	reg("anagram_key", "Anagram key", "Returns the letters and digits of the text lower-cased and sorted; two texts are anagrams when their keys match.", "str", `text.anagram_key("Listen")`, 33, "anagram sorted letters key signature canonical",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return letterKey(a.Str("s"), true), nil })
	reg("is_pangram", "Is pangram", "True when the text contains every letter of the alphabet (default a-z, case-insensitive). Pass another alphabet to test other languages.", "bool", `text.is_pangram("The quick brown fox jumps over the lazy dog")`, 34, "pangram all letters alphabet check",
		[]P{ps("s", "Input text."), pso("alphabet", "Letters that must all appear.", "abcdefghijklmnopqrstuvwxyz")},
		func(a tools.Args) (any, error) {
			have := map[rune]bool{}
			for _, r := range strings.ToLower(a.Str("s")) {
				have[r] = true
			}
			al := strings.ToLower(a.Str("alphabet"))
			if al == "" {
				return nil, errors.New("alphabet must not be empty")
			}
			for _, r := range al {
				if !have[r] {
					return false, nil
				}
			}
			return true, nil
		})
	pred := func(name, title, doc, ex string, tags string, f func(string) bool) {
		reg(name, title, doc, "bool", ex, 40, tags, []P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return f(a.Str("s")), nil })
	}
	pred("is_ascii", "Is ASCII", "True when every character is ASCII (code point below 128). Empty text is ASCII.", `text.is_ascii("hello")`, "ascii check 7-bit plain", func(s string) bool {
		for i := 0; i < len(s); i++ {
			if s[i] >= 0x80 {
				return false
			}
		}
		return true
	})
	pred("is_numeric", "Is numeric", "True when the text is non-empty and every character is a Unicode decimal digit (Nd). Signs, dots and spaces make it false.", `text.is_numeric("12345")`, "numeric digits only number check", func(s string) bool { return allRunes(s, unicode.IsDigit) })
	pred("is_alpha", "Is alphabetic", "True when the text is non-empty and every character is a letter (combining marks allowed).", `text.is_alpha("héllo")`, "alpha letters only alphabetic check", func(s string) bool {
		return allRunes(s, func(r rune) bool { return unicode.IsLetter(r) || isMark(r) })
	})
	pred("is_blank", "Is blank", "True when the text is empty or contains only whitespace (including Unicode spaces).", `text.is_blank("  \t")`, "blank empty whitespace check", func(s string) bool { return strings.TrimSpace(s) == "" })
	cnt := func(name, title, doc, ex string, want bool) {
		reg(name, title, doc, "int", ex, 41, "count letters "+name+" vowels consonants",
			[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
				n := 0
				for _, r := range a.Str("s") {
					if unicode.IsLetter(r) && isVowel(r) == want {
						n++
					}
				}
				return int64(n), nil
			})
	}
	cnt("vowel_count", "Vowel count", "Counts the vowels a e i o u, including their accented forms (é, ü ...), case-insensitive.", `text.vowel_count("Education")`, true)
	cnt("consonant_count", "Consonant count", "Counts letters that are not vowels (y counts as a consonant); digits and punctuation are ignored. Non-Latin letters count as consonants.", `text.consonant_count("Education")`, false)
	freqDoc := "Grapheme-level frequency; whitespace is skipped unless include_space is true."
	reg("char_frequency", "Character frequency", "Returns a dict mapping each character to how many times it occurs (values are counts as text). "+freqDoc+" Use char_frequency_sorted for ordered output.", "dict[str, str]", `text.char_frequency("hello")`, 42, "frequency histogram count characters letters occurrences",
		[]P{ps("s", "Input text."), pb("include_space", "Count whitespace too.", false)},
		func(a tools.Args) (any, error) {
			m, _ := freq(a.Str("s"), a.Bool("include_space"))
			out := make(map[string]string, len(m))
			for k, v := range m {
				out[k] = itoa(v)
			}
			return out, nil
		})
	reg("char_frequency_sorted", "Character frequency (sorted)", "Returns lines like 'l: 2' ordered by count descending, then by character ascending. "+freqDoc, "list[str]", `text.char_frequency_sorted("hello")`, 43, "frequency histogram sorted ranked count characters",
		[]P{ps("s", "Input text."), pb("include_space", "Count whitespace too.", false)},
		func(a tools.Args) (any, error) {
			m, _ := freq(a.Str("s"), a.Bool("include_space"))
			keys := sortedKeys(m)
			sort.SliceStable(keys, func(i, j int) bool { return m[keys[i]] > m[keys[j]] })
			out := make([]string, len(keys))
			for i, k := range keys {
				out[i] = k + ": " + itoa(m[k])
			}
			return out, nil
		})
	reg("unique_chars", "Unique characters", "Returns each distinct character once, in order of first appearance (whitespace skipped).", "str", `text.unique_chars("hello")`, 44, "unique distinct characters dedupe letters",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
			_, order := freq(a.Str("s"), false)
			return strings.Join(order, ""), nil
		})
	reg("duplicate_chars", "Duplicate characters", "Returns the characters that occur more than once, each listed once in order of first appearance (whitespace skipped).", "str", `text.duplicate_chars("hello")`, 45, "duplicate repeated characters letters twice",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
			m, order := freq(a.Str("s"), false)
			var b strings.Builder
			for _, g := range order {
				if m[g] > 1 {
					b.WriteString(g)
				}
			}
			return b.String(), nil
		})
	reg("most_common_char", "Most common character", "Returns the most frequent character (whitespace skipped); ties go to the smallest character; empty text returns empty.", "str", `text.most_common_char("hello")`, 46, "most common frequent character mode top",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
			m, _ := freq(a.Str("s"), false)
			best := ""
			for _, k := range sortedKeys(m) {
				if best == "" || m[k] > m[best] {
					best = k
				}
			}
			return best, nil
		})
}
