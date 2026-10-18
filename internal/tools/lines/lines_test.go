package lines

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "lines") }

type M = map[string]any

var tested = map[string]bool{}

func call(t *testing.T, name string, args M) (string, error) {
	t.Helper()
	tested[name] = true
	v, err := tools.Call("lines", name, args)
	if err != nil {
		return "", err
	}
	return tools.Format(v), nil
}

func TestBehaviour(t *testing.T) {
	cases := []struct {
		fn   string
		args M
		want string
	}{
		{"sort_lines", M{"s": "b\na\nc"}, "a\nb\nc"},
		{"sort_lines", M{"s": "b\na\nc\n", "order": "desc"}, "c\nb\na\n"},
		{"sort_lines", M{"s": "a\r\nB\r\nc\r\n", "mode": "case_insensitive"}, "a\r\nB\r\nc\r\n"},
		{"sort_lines", M{"s": "file10\nfile2\nFile1", "mode": "natural"}, "File1\nfile2\nfile10"},
		{"sort_lines", M{"s": "10\n9\nabc\n-1.5\n1e2", "mode": "numeric"}, "-1.5\n9\n10\n1e2\nabc"},
		{"sort_lines", M{"s": "ccc\na\nbb", "mode": "length"}, "a\nbb\nccc"},
		{"sort_lines", M{"s": "é\nz\na"}, "a\nz\né"},
		{"sort_lines", M{"s": "b\na", "reverse": true}, "b\na"},
		{"sort_lines", M{"s": ""}, ""},
		{"unique_lines", M{"s": "a\nb\na\nA\n"}, "a\nb\nA\n"},
		{"unique_lines", M{"s": "a\nb\na\nA", "case_insensitive": true}, "a\nb"},
		{"duplicate_lines", M{"s": "a\nb\na\nb\na\nc"}, "a\nb"},
		{"sort_unique", M{"s": "b\na\nb"}, "a\nb"},
		{"reverse_lines", M{"s": "a\nb\nc\n"}, "c\nb\na\n"},
		{"shuffle_lines", M{"s": "a\nb\nc\nd\ne\nf", "seed": int64(1)}, ""}, // checked below
		{"count_lines", M{"s": "a\n\nb\n"}, "3"},
		{"count_lines", M{"s": "a\r\n\r\nb", "skip_empty": true}, "2"},
		{"count_lines", M{"s": ""}, "0"},
		{"number_lines", M{"s": "a\nb"}, "1. a\n2. b"},
		{"number_lines", M{"s": "a\nb", "start": int64(9), "sep": ") ", "align": true}, " 9) a\n10) b"},
		{"remove_empty_lines", M{"s": "a\n\n  \nb\n"}, "a\nb\n"},
		{"remove_empty_lines", M{"s": "a\n\n  \nb", "whitespace_only": false}, "a\n  \nb"},
		{"trim_lines", M{"s": "  a \n b"}, "a\nb"},
		{"trim_lines", M{"s": "  a \n b", "mode": "left"}, "a \nb"},
		{"join_lines", M{"s": "a\r\nb\r\nc\r\n", "sep": ", "}, "a, b, c"},
		{"join_lines", M{"s": "a\n\nb", "skip_empty": true}, "a b"},
		{"split_to_lines", M{"s": "a, b,c", "trim": true}, "a\nb\nc"},
		{"split_to_lines", M{"s": "a;b", "sep": ";"}, "a\nb"},
		{"prefix_lines", M{"s": "a\n\nb", "prefix": "> ", "skip_empty": true}, "> a\n\n> b"},
		{"suffix_lines", M{"s": "a\nb\n"}, "a;\nb;\n"},
		{"surround_lines", M{"s": "a\nb"}, "\"a\"\n\"b\""},
		{"indent_lines", M{"s": "a\n\nb", "width": int64(2)}, "  a\n\n  b"},
		{"indent_lines", M{"s": "a", "width": int64(1), "use_tabs": true}, "\ta"},
		{"unindent_lines", M{"s": "    a\n      b"}, "a\n  b"},
		{"unindent_lines", M{"s": "    a\n  b", "width": int64(3)}, " a\nb"},
		{"filter_contains", M{"s": "apple\nbanana\ncherry", "substr": "an"}, "banana"},
		{"filter_contains", M{"s": "Apple\nbanana", "substr": "APPLE", "invert": true, "ignore_case": true}, "banana"},
		{"filter_prefix", M{"s": "#a\nb\n#c", "prefix": "#", "invert": true}, "b"},
		{"filter_suffix", M{"s": "a.go\nb.py", "suffix": ".go"}, "a.go"},
		{"head", M{"s": "a\nb\nc", "n": int64(2)}, "a\nb"},
		{"head", M{"s": "a\nb", "n": int64(0)}, ""},
		{"tail", M{"s": "a\nb\nc\n", "n": int64(2)}, "b\nc\n"},
		{"tail", M{"s": "a", "n": int64(5)}, "a"},
		{"nth_line", M{"s": "a\nb\nc", "n": int64(2)}, "b"},
		{"nth_line", M{"s": "a\nb\nc", "n": int64(-1)}, "c"},
		{"line_range", M{"s": "a\nb\nc\nd", "from": int64(2), "to": int64(3)}, "b\nc"},
		{"line_range", M{"s": "a\nb\nc\nd", "from": int64(3)}, "c\nd"},
		{"line_range", M{"s": "a\nb\nc\nd", "from": int64(-2), "to": int64(-1)}, "c\nd"},
		{"line_range", M{"s": "a\nb", "from": int64(5)}, ""},
		{"skip", M{"s": "a\nb\nc"}, "b\nc"},
		{"take_every", M{"s": "a\nb\nc\nd\ne", "n": int64(2)}, "a\nc\ne"},
		{"take_every", M{"s": "a\nb\nc\nd\ne", "n": int64(2), "offset": int64(1)}, "b\nd"},
		{"dedupe_consecutive", M{"s": "a\na\nb\na\nA", "case_insensitive": true}, "a\nb\na"},
		{"min_max_lines", M{"s": "bb\na\nccc"}, "count: 3\nmax: ccc\nmin: a"},
		{"min_max_lines", M{"s": "10\n9\nx", "mode": "numeric"}, "count: 2\nmax: 10\nmin: 9"},
		{"transpose", M{"s": "a,b,c\nd,e", "delimiter": ","}, "a,d\nb,e\nc,"},
		{"transpose", M{"s": "a b\nc d", "delimiter": ""}, "a c\nb d"},
		{"columns", M{"s": "a,b,c\nd,e", "delimiter": ",", "indexes": []string{"1", "3"}}, "a,c\nd"},
		{"columns", M{"s": "a b c d", "delimiter": "", "indexes": []string{"2-3", "-1"}}, "b c d"},
		{"columns", M{"s": "a b c d", "delimiter": "", "indexes": []string{"3-"}, "output_delimiter": "|"}, "c|d"},
		{"align_columns", M{"s": "a,bbb\ncc,d", "delimiter": ",", "separator": " | "}, "a  | bbb\ncc | d"},
		{"align_columns", M{"s": "é,1\nab,22", "align": "right", "separator": " "}, " é  1\nab 22"},
		{"swap_columns", M{"s": "a,b,c\nx", "col_a": int64(1), "col_b": int64(3)}, "c,b,a\nx"},
		{"sort_by_column", M{"s": "x 10\ny 9\nz 100", "column": int64(2), "mode": "numeric"}, "y 9\nx 10\nz 100"},
		{"sort_by_column", M{"s": "a,2\nb,1", "column": int64(-1), "delimiter": ",", "order": "desc"}, "a,2\nb,1"},
		{"sum_numbers", M{"s": "1\n2.5\n\nabc\n1,000"}, "1003.5"},
		{"sum_numbers", M{"s": ""}, "0"},
		{"average_numbers", M{"s": "1\n2\n6"}, "3"},
		{"list_to_csv_row", M{"s": "a\nb,c\nd\"e\n"}, `a,"b,c","d""e"`},
		{"csv_row_to_lines", M{"s": `a,"b,c","d""e",`}, "a\nb,c\nd\"e\n"},
		{"csv_to_tsv", M{"s": "a,\"b,c\"\n1,\"x\ty\"\n"}, "a\tb,c\n1\tx y\n"},
		{"tsv_to_csv", M{"s": "a\tb,c\n1\t\"q\""}, "a,\"b,c\"\n1,\"\"\"q\"\"\""},
		{"markdown_table_from_csv", M{"s": "name,age\nAnn,30\nB|b"}, "| name | age |\n| ---- | --- |\n| Ann  | 30  |\n| B\\|b |     |"},
		{"markdown_table_from_csv", M{"s": "a,b", "header": false}, "| Column 1 | Column 2 |\n| -------- | -------- |\n| a        | b        |"},
		{"markdown_bullets", M{"s": "a\n\nb"}, "- a\n\n- b"},
		{"numbered_list", M{"s": "a\n\nb"}, "1. a\n\n2. b"},
		{"remove_markdown_bullets", M{"s": "- a\n  * b\n2) c\n- [x] d\nplain"}, "a\n  b\nc\nd\nplain"},
		{"lines_diff", M{"a": "a\nb\nc", "b": "b\nc\nd"}, "common: b\nc\nonly_in_a: a\nonly_in_b: d"},
		{"intersect_lines", M{"a": "a\nb\nb\nc", "b": "B\nc", "ignore_case": true}, "b\nc"},
		{"union_lines", M{"a": "a\nb", "b": "b\nc\n"}, "a\nb\nc"},
		{"subtract_lines", M{"a": "a\r\nb\r\nc\r\n", "b": "b"}, "a\r\nc\r\n"},
		{"wrap_each", M{"s": "abcdefgh\n\nxy", "width": int64(3)}, "abc\ndef\ngh\n\nxy"},
		{"word_wrap_lines", M{"s": "the quick brown fox\nhi", "width": int64(10)}, "the quick\nbrown fox\nhi"},
		{"word_wrap_lines", M{"s": "abcdefgh", "width": int64(3), "break_long": true}, "abc\ndef\ngh"},
		{"to_json_array", M{"s": "a\n\"b\"\n<é>"}, `["a","\"b\"","<é>"]`},
		{"to_json_array", M{"s": "a\n\nb", "skip_empty": true, "pretty": true}, "[\n  \"a\",\n  \"b\"\n]"},
		{"from_json_array", M{"s": `["a", "b\nc", 3, null, {"k": 1}]`}, "a\nb\nc\n3\nnull\n{\"k\":1}"},
		{"to_sql_in_list", M{"s": "a\nit's\n"}, `'a','it''s'`},
		{"to_sql_in_list", M{"s": "1\n2.5", "numbers": true, "parens": true}, "(1,2.5)"},
		{"to_sql_in_list", M{"s": "1\nx", "numbers": true}, "'1','x'"},
		{"to_quoted_csv", M{"s": "a\nb\"c"}, `"a","b""c"`},
		{"to_lf", M{"s": "a\r\nb\rc\n"}, "a\nb\nc\n"},
		{"to_crlf", M{"s": "a\nb\r\nc\rd"}, "a\r\nb\r\nc\r\nd"},
		{"detect_line_endings", M{"s": "a\r\nb\r\n"}, "CRLF"},
		{"detect_line_endings", M{"s": "a\nb"}, "LF"},
		{"detect_line_endings", M{"s": "a\rb"}, "CR"},
		{"detect_line_endings", M{"s": "a\r\nb\nc"}, "mixed"},
		{"detect_line_endings", M{"s": "abc"}, "none"},
		{"strip_trailing_whitespace", M{"s": "a  \nb\t\n \nc"}, "a\nb\n\nc"},
		{"tabs_to_spaces", M{"s": "a\tb\n\tc", "width": int64(4)}, "a   b\n    c"},
		{"spaces_to_tabs", M{"s": "        x\n      y\n a b", "width": int64(4)}, "\t\tx\n\t  y\n a b"},
		{"remove_comments", M{"s": "# a\ncode # trailing\n  // b\n\nx"}, "code # trailing\n\nx"},
		{"remove_comments", M{"s": "; a\nb\n\n", "prefixes": []string{";"}, "drop_empty": true}, "b\n"},
	}
	for _, c := range cases {
		got, err := call(t, c.fn, c.args)
		if err != nil {
			t.Errorf("%s(%v): unexpected error %v", c.fn, c.args, err)
			continue
		}
		if c.fn == "shuffle_lines" {
			continue
		}
		if got != c.want {
			t.Errorf("%s(%q) = %q, want %q", c.fn, c.args, got, c.want)
		}
	}
}

func TestShuffle(t *testing.T) {
	in := "a\nb\nc\nd\ne\nf\ng\nh"
	a, _ := call(t, "shuffle_lines", M{"s": in, "seed": int64(5)})
	b, _ := call(t, "shuffle_lines", M{"s": in, "seed": int64(5)})
	c, _ := call(t, "shuffle_lines", M{"s": in, "seed": int64(6)})
	if a != b || a == c || a == in {
		t.Errorf("shuffle not deterministic per seed: %q %q %q", a, b, c)
	}
	sorted, _ := call(t, "sort_lines", M{"s": a})
	if sorted != in {
		t.Errorf("shuffle lost lines: %q", a)
	}
}

func TestErrors(t *testing.T) {
	bad := []struct {
		fn   string
		args M
	}{
		{"sort_lines", M{"s": "a", "mode": "bogus"}},
		{"sort_lines", M{"s": "a", "order": "up"}},
		{"trim_lines", M{"s": "a", "mode": "x"}},
		{"head", M{"s": "a", "n": int64(-1)}},
		{"nth_line", M{"s": "a", "n": int64(2)}},
		{"nth_line", M{"s": "", "n": int64(1)}},
		{"split_to_lines", M{"s": "a", "sep": ""}},
		{"take_every", M{"s": "a", "n": int64(0)}},
		{"columns", M{"s": "a", "indexes": []string{"x"}}},
		{"columns", M{"s": "a", "indexes": []string{"0"}}},
		{"swap_columns", M{"s": "a", "col_a": int64(0)}},
		{"align_columns", M{"s": "a", "align": "center"}},
		{"average_numbers", M{"s": "abc"}},
		{"sum_numbers", M{"s": "1\nabc", "strict": true}},
		{"csv_row_to_lines", M{"s": "a", "delimiter": "ab"}},
		{"csv_to_tsv", M{"s": "a", "delimiter": "\""}},
		{"from_json_array", M{"s": "{}"}},
		{"from_json_array", M{"s": ""}},
		{"wrap_each", M{"s": "a", "width": int64(0)}},
		{"word_wrap_lines", M{"s": "a", "width": int64(0)}},
		{"tabs_to_spaces", M{"s": "a", "width": int64(0)}},
		{"spaces_to_tabs", M{"s": "a", "width": int64(99)}},
		{"indent_lines", M{"s": "a", "width": int64(-1)}},
		{"to_quoted_csv", M{"s": "a", "quote": ""}},
		{"min_max_lines", M{"s": "a", "mode": "x"}},
		{"indent_lines", M{"s": strings.Repeat("a\n", 100000), "width": int64(1000)}},
		{"prefix_lines", M{"s": strings.Repeat("a\n", 100000), "prefix": strings.Repeat("p", 1000)}},
	}
	for _, c := range bad {
		if _, err := call(t, c.fn, c.args); err == nil {
			t.Errorf("%s(%.60v): expected an error", c.fn, c.args)
		}
	}
}

func TestLineEndingRules(t *testing.T) {
	// CRLF input keeps CRLF and its trailing break; text without a trailing break gets none.
	got, _ := call(t, "reverse_lines", M{"s": "a\r\nb\r\n"})
	if got != "b\r\na\r\n" {
		t.Errorf("CRLF not preserved: %q", got)
	}
	got, _ = call(t, "reverse_lines", M{"s": "a\nb"})
	if got != "b\na" {
		t.Errorf("no-trailing-newline not preserved: %q", got)
	}
	got, _ = call(t, "reverse_lines", M{"s": "\n"})
	if got != "\n" {
		t.Errorf("single newline: %q", got)
	}
	got, _ = call(t, "reverse_lines", M{"s": "a\rb"})
	if got != "b\ra" {
		t.Errorf("CR not preserved: %q", got)
	}
	got, _ = call(t, "sort_lines", M{"s": "ü\n😀\nz"})
	if got != "z\nü\n😀" {
		t.Errorf("unicode sort: %q", got)
	}
}

// TestNoPanicOnGarbage calls every function with empty and odd input.
func TestNoPanicOnGarbage(t *testing.T) {
	for _, f := range tools.All() {
		if f.Group != "lines" {
			continue
		}
		for _, s := range []string{"", "\n", "\xff\xfe", "\r\n\r\n", "á‍\n😀"} {
			named := M{}
			for _, p := range f.Params {
				if p.Required {
					named[p.Name] = s
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
	for _, f := range tools.All() {
		if f.Group == "lines" && !tested[f.Name] {
			t.Errorf("%s has no behaviour test", f.Key())
		}
	}
}
