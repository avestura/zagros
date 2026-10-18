package text

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "text") }

type M = map[string]any

var tested = map[string]bool{}

func call(t *testing.T, name string, args M) (string, error) {
	t.Helper()
	tested[name] = true
	v, err := tools.Call("text", name, args)
	if err != nil {
		return "", err
	}
	return tools.Format(v), nil
}

func TestBehaviour(t *testing.T) {
	long := strings.Repeat("ab", 10)
	cases := []struct {
		fn   string
		args M
		want string
	}{
		{"lower", M{"s": "HÉLLO Σ"}, "héllo σ"},
		{"upper", M{"s": "héllo"}, "HÉLLO"},
		{"title", M{"s": "the qUICK don't stop"}, "The Quick Don't Stop"},
		{"title", M{"s": ""}, ""},
		{"sentence_case", M{"s": "HELLO there. HOW are you? fine"}, "Hello there. How are you? Fine"},
		{"swapcase", M{"s": "Hello 1"}, "hELLO 1"},
		{"alternating_case", M{"s": "hello world"}, "hElLo WoRlD"},
		{"camel", M{"s": "Hello big_world-foo"}, "helloBigWorldFoo"},
		{"camel", M{"s": "HTMLParser"}, "htmlParser"},
		{"pascal", M{"s": "hello big world"}, "HelloBigWorld"},
		{"snake", M{"s": "helloWorld Foo"}, "hello_world_foo"},
		{"kebab", M{"s": "Hello World"}, "hello-world"},
		{"constant", M{"s": "hello world"}, "HELLO_WORLD"},
		{"dot_case", M{"s": "Hello World"}, "hello.world"},
		{"path_case", M{"s": "Hello World"}, "hello/world"},
		{"train_case", M{"s": "hello world"}, "Hello-World"},
		{"snake", M{"s": "  --  "}, ""},
		{"slug", M{"s": "Héllo, Wörld! Straße 2024"}, "hello-world-strasse-2024"},
		{"slug", M{"s": "A  B", "separator": "_"}, "a_b"},
		{"slug", M{"s": "Привет мир", "ascii": false}, "привет-мир"},
		{"slug", M{"s": "日本"}, ""},
		{"capitalize_words", M{"s": "hello wORLD"}, "Hello WORLD"},
		{"uncapitalize", M{"s": "Hello"}, "hello"},
		{"uncapitalize", M{"s": ""}, ""},
		{"length", M{"s": "héllo😀"}, "6"},
		{"byte_length", M{"s": "é"}, "2"},
		{"grapheme_length", M{"s": "e\u0301"}, "1"},
		{"grapheme_length", M{"s": "👨‍👩‍👧x🇮🇷👍🏽"}, "4"},
		{"grapheme_length", M{"s": "a\r\nb"}, "3"},
		{"word_count", M{"s": "Hello, don't-stop well-known  ... 42"}, "4"},
		{"word_count", M{"s": ""}, "0"},
		{"line_count", M{"s": "a\r\nb\n"}, "2"},
		{"line_count", M{"s": ""}, "0"},
		{"sentence_count", M{"s": "Hi. How are you? Fine"}, "3"},
		{"sentence_count", M{"s": "Pi is 3.14 okay!"}, "1"},
		{"sentence_count", M{"s": ""}, "0"},
		{"char_count", M{"s": "banana", "ch": "an"}, "2"},
		{"char_count", M{"s": "aAa", "ch": "a", "ignore_case": true}, "3"},
		{"is_palindrome", M{"s": "A man, a plan, a canal: Panama"}, "true"},
		{"is_palindrome", M{"s": "ab", "ignore_non_alnum": false}, "false"},
		{"is_palindrome", M{"s": ""}, "true"},
		{"longest_palindromic_substring", M{"s": "babad"}, "bab"},
		{"longest_palindromic_substring", M{"s": "cbbd"}, "bb"},
		{"longest_palindromic_substring", M{"s": ""}, ""},
		{"is_anagram", M{"a": "Dormitory", "b": "dirty room"}, "true"},
		{"is_anagram", M{"a": "abc", "b": "abd"}, "false"},
		{"anagram_key", M{"s": "Listen"}, "eilnst"},
		{"is_pangram", M{"s": "The quick brown fox jumps over the lazy dog"}, "true"},
		{"is_pangram", M{"s": "hello"}, "false"},
		{"is_ascii", M{"s": "abc"}, "true"},
		{"is_ascii", M{"s": "é"}, "false"},
		{"is_numeric", M{"s": "123"}, "true"},
		{"is_numeric", M{"s": ""}, "false"},
		{"is_numeric", M{"s": "1.5"}, "false"},
		{"is_alpha", M{"s": "héllo"}, "true"},
		{"is_alpha", M{"s": "a1"}, "false"},
		{"is_blank", M{"s": " \t\n"}, "true"},
		{"is_blank", M{"s": "a"}, "false"},
		{"vowel_count", M{"s": "Education é"}, "6"},
		{"consonant_count", M{"s": "Education 123"}, "4"},
		{"char_frequency", M{"s": "aab"}, "a: 2\nb: 1"},
		{"char_frequency_sorted", M{"s": "abba c"}, "a: 2\nb: 2\nc: 1"},
		{"unique_chars", M{"s": "hello"}, "helo"},
		{"duplicate_chars", M{"s": "hello"}, "l"},
		{"most_common_char", M{"s": "hello"}, "l"},
		{"most_common_char", M{"s": ""}, ""},
		{"reverse", M{"s": "abc"}, "cba"},
		{"reverse", M{"s": "ae\u0301👍🏽"}, "👍🏽e\u0301a"},
		{"reverse_words", M{"s": "one  two three"}, "three two one"},
		{"sort_chars", M{"s": "hello"}, "ehllo"},
		{"sort_chars", M{"s": "b a", "desc": true, "remove_space": true}, "ba"},
		{"sort_words", M{"s": "pear Apple fig"}, "Apple fig pear"},
		{"sort_words", M{"s": "a c b", "desc": true}, "c b a"},
		{"trim", M{"s": "  hi \n"}, "hi"},
		{"trim", M{"s": "  hi  ", "mode": "left"}, "hi  "},
		{"trim", M{"s": "xxhixx", "mode": "right", "cutset": "x"}, "xxhi"},
		{"collapse_whitespace", M{"s": " a   b \n c "}, "a b c"},
		{"remove_whitespace", M{"s": "a b\tc\n"}, "abc"},
		{"remove_punctuation", M{"s": "Hi, there! $5"}, "Hi there $5"},
		{"remove_punctuation", M{"s": "Hi, $5", "symbols": true}, "Hi 5"},
		{"remove_digits", M{"s": "a1b2"}, "ab"},
		{"only_digits", M{"s": "tel: 555-12"}, "55512"},
		{"only_letters", M{"s": "a1 é!"}, "aé"},
		{"only_alnum", M{"s": "a-1 é_"}, "a1é"},
		{"remove_non_ascii", M{"s": "héllo ✓"}, "hllo "},
		{"strip_accents", M{"s": "Crème brûlée ø"}, "Creme brulee ø"},
		{"normalize", M{"s": "e\u0301", "form": "NFC"}, "\u00e9"},
		{"normalize", M{"s": "\u00e9", "form": "nfd"}, "e\u0301"},
		{"normalize", M{"s": "ﬁ", "form": "NFKC"}, "fi"},
		{"repeat", M{"s": "ab", "n": int64(3), "sep": "-"}, "ab-ab-ab"},
		{"repeat", M{"s": "ab", "n": int64(0)}, ""},
		{"pad_left", M{"s": "7", "width": int64(3), "fill": "0"}, "007"},
		{"pad_right", M{"s": "é", "width": int64(3), "fill": "ab"}, "éab"},
		{"center", M{"s": "hi", "width": int64(7), "fill": "*"}, "**hi***"},
		{"pad_left", M{"s": "long", "width": int64(2)}, "long"},
		{"truncate", M{"s": "Hello world", "max": int64(8)}, "Hello w…"},
		{"truncate", M{"s": "short", "max": int64(8)}, "short"},
		{"truncate", M{"s": "Hello", "max": int64(3), "ellipsis": "...."}, "Hel"},
		{"truncate_words", M{"s": "one two three four", "max_words": int64(2)}, "one two…"},
		{"wrap", M{"s": "the quick brown fox", "width": int64(10)}, "the quick\nbrown fox"},
		{"wrap", M{"s": "a\n\nsupercalifragilistic", "width": int64(5)}, "a\n\nsupercalifragilistic"},
		{"wrap", M{"s": "abcdefgh", "width": int64(3), "break_long": true}, "abc\ndef\ngh"},
		{"indent", M{"s": "a\n\nb", "prefix": "  "}, "  a\n\n  b"},
		{"unindent", M{"s": "  a\n    b\n\n  c"}, "a\n  b\n\nc"},
		{"replace", M{"s": "a-b-c", "old": "-", "new": "+"}, "a+b+c"},
		{"replace", M{"s": "a-b-c", "old": "-", "new": "+", "count": int64(1)}, "a+b-c"},
		{"replace", M{"s": "abc", "old": ""}, "abc"},
		{"replace_ci", M{"s": "Hello HELLO", "old": "hello", "new": "bye"}, "bye bye"},
		{"replace_ci", M{"s": "İstanbul İ", "old": "İ", "new": "I"}, "Istanbul I"},
		{"remove", M{"s": "banana", "sub": "an"}, "ba"},
		{"insert", M{"s": "held", "index": int64(3), "text": "l"}, "helld"},
		{"insert", M{"s": "ab", "index": int64(-1), "text": "-"}, "a-b"},
		{"insert", M{"s": "ab", "index": int64(99), "text": "!"}, "ab!"},
		{"delete_range", M{"s": "hello", "start": int64(1), "end": int64(3)}, "hlo"},
		{"delete_range", M{"s": "hello", "start": int64(3), "end": int64(1)}, "hello"},
		{"substring", M{"s": "héllo wörld", "start": int64(6), "length": int64(5)}, "wörld"},
		{"substring", M{"s": "hello", "start": int64(-3)}, "llo"},
		{"substring", M{"s": "hello", "start": int64(10)}, ""},
		{"split_at", M{"s": "héllo", "index": int64(2)}, "hé\nllo"},
		{"left", M{"s": "héllo", "n": int64(2)}, "hé"},
		{"left", M{"s": "ab", "n": int64(-2)}, ""},
		{"right", M{"s": "héllo", "n": int64(3)}, "llo"},
		{"right", M{"s": "ab", "n": int64(9)}, "ab"},
		{"remove_prefix", M{"s": "foobar", "prefix": "foo"}, "bar"},
		{"remove_suffix", M{"s": "foobar", "suffix": "bar"}, "foo"},
		{"leetspeak", M{"s": "Elite Test"}, "3l173 7357"},
		{"mirror", M{"s": "ab"}, "qɐ"},
		{"remove_zero_width", M{"s": "a\u200bb\ufeffc"}, "abc"},
		{"show_invisibles", M{"s": "a\tb\u200b\n c"}, `a\tb<U+200B>\n c`},
		{"show_invisibles", M{"s": "a b", "show_spaces": true}, "a·b"},
		{"smart_quotes_to_straight", M{"s": "“hi” it’s"}, `"hi" it's`},
		{"smart_quotes_to_straight", M{"s": "a—b…", "punctuation": true}, "a-b..."},
		{"straight_to_smart", M{"s": `"hi" it's 'x'`}, "“hi” it’s ‘x’"},
		{"remove_emoji", M{"s": "hi 😀 there 👍🏽 👨‍👩‍👧 ❤️"}, "hi  there   "},
		{"to_fullwidth", M{"s": "Ab 1"}, "Ａｂ　１"},
		{"to_halfwidth", M{"s": "Ａｂ　１"}, "Ab 1"},
		{"levenshtein", M{"a": "kitten", "b": "sitting"}, "3"},
		{"levenshtein", M{"a": "", "b": "é"}, "1"},
		{"damerau_levenshtein", M{"a": "ca", "b": "ac"}, "1"},
		{"damerau_levenshtein", M{"a": "abcd", "b": "acbd"}, "1"},
		{"damerau_levenshtein", M{"a": "kitten", "b": "sitting"}, "3"},
		{"hamming", M{"a": "karolin", "b": "kathrin"}, "3"},
		{"jaro_winkler", M{"a": "martha", "b": "marhta"}, "0.961111111111"},
		{"jaro_winkler", M{"a": "", "b": ""}, "1"},
		{"jaro_winkler", M{"a": "abc", "b": "xyz"}, "0"},
		{"similarity", M{"a": "abc", "b": "abc"}, "1"},
		{"similarity", M{"a": "abcd", "b": "abce"}, "0.75"},
		{"similarity", M{"a": "", "b": ""}, "1"},
		{"common_prefix", M{"a": "interview", "b": "internet"}, "inter"},
		{"common_suffix", M{"a": "testing", "b": "running"}, "ing"},
		{"longest_common_substring", M{"a": "abcdef", "b": "zcdemn"}, "cde"},
		{"diff_summary", M{"a": "kitten", "b": "sitting"}, "deleted: 2\ndistance: 5\ninserted: 3\nsimilarity: 0.6154\nunchanged: 4"},
		{"compare_ci", M{"a": "Apple", "b": "apple"}, "0"},
		{"compare_ci", M{"a": "a", "b": "B"}, "-1"},
		{"equals_ci", M{"a": "Hello", "b": "hELLO"}, "true"},
		{"codepoints", M{"s": "aé😀"}, "U+0061\nU+00E9\nU+1F600"},
		{"from_codepoints", M{"points": []string{"U+0048", "0x69", "\\u0021", "&#33;"}}, "Hi!!"},
		{"to_codepoint", M{"ch": "😀"}, "128512"},
		{"from_codepoint", M{"n": int64(233)}, "é"},
		{"unicode_blocks", M{"s": "aб😀a"}, "Basic Latin\nCyrillic\nEmoticons"},
		{"counter_string", M{"length": int64(10)}, "*3*5*7*10*"},
		{"counter_string", M{"length": int64(1)}, "*"},
		{"counter_string", M{"length": int64(0)}, ""},
		{"lorem", M{"words": int64(8)}, "Lorem ipsum dolor sit amet, consectetur adipiscing elit."},
		{"repeat", M{"s": long, "n": int64(1)}, long},
	}
	for _, c := range cases {
		got, err := call(t, c.fn, c.args)
		if err != nil {
			t.Errorf("%s(%v): unexpected error %v", c.fn, c.args, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s(%v) = %q, want %q", c.fn, c.args, got, c.want)
		}
	}
}

func TestDicts(t *testing.T) {
	v, _ := tools.Call("text", "case_all", M{"s": "hello World"})
	m := v.(map[string]string)
	want := map[string]string{"lower": "hello world", "upper": "HELLO WORLD", "title": "Hello World", "sentence": "Hello world",
		"swapcase": "HELLO wORLD", "alternating": "hElLo WoRlD", "camel": "helloWorld", "pascal": "HelloWorld", "snake": "hello_world",
		"kebab": "hello-world", "constant": "HELLO_WORLD", "dot": "hello.world", "path": "hello/world", "train": "Hello-World", "slug": "hello-world"}
	for k, w := range want {
		if m[k] != w {
			t.Errorf("case_all[%s] = %q, want %q", k, m[k], w)
		}
	}
	v, _ = tools.Call("text", "stats", M{"s": "Hello world. Bye!"})
	m = v.(map[string]string)
	for k, w := range map[string]string{"chars": "17", "words": "3", "lines": "1", "sentences": "2", "longest_word": "Hello", "reading_time_seconds": "1", "avg_word_length": "4.33"} {
		if m[k] != w {
			t.Errorf("stats[%s] = %q, want %q", k, m[k], w)
		}
	}
	v, _ = tools.Call("text", "palindrome_report", M{"s": "Never odd or even"})
	m = v.(map[string]string)
	if m["is_palindrome"] != "true" || m["cleaned"] != "neveroddoreven" || m["reversed"] != "neveroddoreven" {
		t.Errorf("palindrome_report = %v", m)
	}
	v, _ = tools.Call("text", "char_info", M{"ch": "é"})
	m = v.(map[string]string)
	if m["codepoint"] != "U+00E9" || m["utf8"] != "C3 A9" || m["utf16"] != "00E9" || m["category"] != "Ll" || m["block"] != "Latin-1 Supplement" {
		t.Errorf("char_info = %v", m)
	}
	v, _ = tools.Call("text", "char_info", M{"ch": "😀"})
	m = v.(map[string]string)
	if m["utf16"] != "D83D DE00" || m["utf8_length"] != "4" || m["block"] != "Emoticons" || m["category"] != "So" {
		t.Errorf("char_info emoji = %v", m)
	}
}

func TestLoremDeterministic(t *testing.T) {
	a, _ := call(t, "lorem", M{"words": int64(100), "seed": int64(7)})
	b, _ := call(t, "lorem", M{"words": int64(100), "seed": int64(7)})
	c, _ := call(t, "lorem", M{"words": int64(100), "seed": int64(8)})
	if a != b || a == c {
		t.Error("lorem is not deterministic per seed")
	}
	if n := len(strings.Fields(a)); n != 100 {
		t.Errorf("lorem word count = %d", n)
	}
	p, err := call(t, "lorem_paragraphs", M{"paragraphs": int64(2), "sentences": int64(3)})
	if err != nil || strings.Count(p, "\n\n") != 1 || strings.Count(p, ".") != 6 {
		t.Errorf("lorem_paragraphs = %q, %v", p, err)
	}
}

func TestErrors(t *testing.T) {
	bad := []struct {
		fn   string
		args M
	}{
		{"char_count", M{"s": "a", "ch": ""}},
		{"hamming", M{"a": "a", "b": "ab"}},
		{"repeat", M{"s": "abc", "n": int64(-1)}},
		{"repeat", M{"s": "abcdefghij", "n": int64(1 << 40)}},
		{"pad_left", M{"s": "a", "width": int64(1 << 40)}},
		{"pad_left", M{"s": "a", "width": int64(5), "fill": ""}},
		{"wrap", M{"s": "a", "width": int64(0)}},
		{"normalize", M{"s": "a", "form": "NFX"}},
		{"trim", M{"s": "a", "mode": "middle"}},
		{"to_codepoint", M{"ch": ""}},
		{"from_codepoint", M{"n": int64(0xD800)}},
		{"from_codepoints", M{"points": []string{"zzz"}}},
		{"char_info", M{"ch": ""}},
		{"lorem", M{"words": int64(0)}},
		{"counter_string", M{"length": int64(-1)}},
		{"counter_string", M{"length": int64(5), "marker": "ab"}},
		{"levenshtein", M{"a": strings.Repeat("a", 8000), "b": strings.Repeat("b", 8000)}},
		{"truncate", M{"s": "a", "max": int64(-1)}},
		{"is_pangram", M{"s": "a", "alphabet": ""}},
	}
	for _, c := range bad {
		if _, err := call(t, c.fn, c.args); err == nil {
			t.Errorf("%s(%v): expected an error", c.fn, c.args)
		}
	}
}

// TestNoPanicOnGarbage calls every function with empty and invalid-UTF-8 text.
func TestNoPanicOnGarbage(t *testing.T) {
	for _, f := range tools.All() {
		if f.Group != "text" {
			continue
		}
		for _, s := range []string{"", "\xff\xfe", "a\u0301\u200d", "😀\x00"} {
			named := M{}
			for _, p := range f.Params {
				if p.Required {
					switch p.Type {
					case "str":
						named[p.Name] = s
					case "int":
						named[p.Name] = int64(1)
					case "list[str]":
						named[p.Name] = []string{s}
					}
				}
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s panicked on %q: %v", f.Key(), s, r)
					}
				}()
				_, _ = f.Invoke(named)
			}()
		}
	}
}

func TestEveryFunctionCovered(t *testing.T) {
	for _, n := range []string{"case_all", "stats", "palindrome_report", "char_info"} {
		tested[n] = true
	}
	for _, f := range tools.All() {
		if f.Group == "text" && !tested[f.Name] {
			t.Errorf("%s has no behaviour test", f.Key())
		}
	}
}
