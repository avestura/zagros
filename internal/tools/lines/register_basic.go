package lines

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

func sortParams() []P {
	return []P{sIn,
		pso("order", "asc or desc.", "asc", "asc", "desc"),
		pso("mode", "How lines are compared.", "text", sortModes...),
		pb("reverse", "Reverse the sorted result afterwards.", false)}
}

func sortLines(ls []string, a tools.Args) ([]string, error) {
	desc, err := orderIsDesc(a.Str("order"))
	if err != nil {
		return nil, err
	}
	cmp, err := compareFor(a.Str("mode"), desc)
	if err != nil {
		return nil, err
	}
	out := append([]string(nil), ls...)
	sort.SliceStable(out, func(i, j int) bool { return cmp(out[i], out[j]) < 0 })
	if a.Bool("reverse") {
		reverse(out)
	}
	return out, nil
}

func reverse(l []string) {
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
}

func uniq(ls []string, ci bool) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		k := fold(l, ci)
		if !seen[k] {
			seen[k] = true
			out = append(out, l)
		}
	}
	return out
}

func trimLine(l, mode string) (string, error) {
	switch strings.ToLower(mode) {
	case "both", "":
		return strings.TrimFunc(l, unicode.IsSpace), nil
	case "left":
		return strings.TrimLeftFunc(l, unicode.IsSpace), nil
	case "right":
		return strings.TrimRightFunc(l, unicode.IsSpace), nil
	}
	return "", fmt.Errorf("unknown mode %q (use both, left or right)", mode)
}

func isBlank(l string) bool { return strings.TrimSpace(l) == "" }

func checkN(n int64, name string) error {
	if n < 0 {
		return fmt.Errorf("%s must not be negative", name)
	}
	return nil
}

func clampTo(n int64, l int) int {
	if n > int64(l) {
		return l
	}
	return int(n)
}

func init() {
	reg("sort_lines", "Sort lines", "Sorts lines. mode is text (code point order), natural (file2 before file10), numeric (by leading number; lines without one go last), length (in characters) or case_insensitive; order is asc or desc; reverse flips the result. The sort is stable.",
		"lines.sort_lines(\"b\\na\\nc\")", 1, "sort order alphabetical natural numeric length lines arrange a-z", sortParams(), sortLines)
	reg("unique_lines", "Unique lines", "Removes duplicate lines, keeping the first occurrence and the original order.",
		"lines.unique_lines(\"a\\nb\\na\")", 2, "unique dedupe duplicates remove distinct uniq lines",
		[]P{sIn, pb("case_insensitive", "Treat lines differing only in case as equal.", false)},
		func(ls []string, a tools.Args) ([]string, error) { return uniq(ls, a.Bool("case_insensitive")), nil })
	reg("duplicate_lines", "Duplicate lines", "Keeps one copy of each line that appears more than once, in order of first appearance.",
		"lines.duplicate_lines(\"a\\nb\\na\")", 20, "duplicates repeated find lines uniq -d",
		[]P{sIn, pb("case_insensitive", "Treat lines differing only in case as equal.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			ci := a.Bool("case_insensitive")
			cnt := map[string]int{}
			for _, l := range ls {
				cnt[fold(l, ci)]++
			}
			var out []string
			seen := map[string]bool{}
			for _, l := range ls {
				k := fold(l, ci)
				if cnt[k] > 1 && !seen[k] {
					seen[k] = true
					out = append(out, l)
				}
			}
			return out, nil
		})
	reg("sort_unique", "Sort and deduplicate lines", "Sorts the lines (same options as sort_lines) and removes duplicates, like sort -u.",
		"lines.sort_unique(\"b\\na\\nb\")", 21, "sort unique dedupe sort -u distinct", append(sortParams(), pb("case_insensitive", "Dedupe ignoring case.", false)),
		func(ls []string, a tools.Args) ([]string, error) {
			out, err := sortLines(ls, a)
			if err != nil {
				return nil, err
			}
			return uniq(out, a.Bool("case_insensitive")), nil
		})
	reg("reverse_lines", "Reverse line order", "Reverses the order of the lines (last line first).",
		"lines.reverse_lines(\"a\\nb\\nc\")", 3, "reverse order flip lines tac backwards", []P{sIn},
		func(ls []string, a tools.Args) ([]string, error) {
			out := append([]string(nil), ls...)
			reverse(out)
			return out, nil
		})
	reg("shuffle_lines", "Shuffle lines", "Shuffles the lines pseudo-randomly. The function is pure: the same seed always gives the same order (seed 0 is a valid fixed seed); change the seed for a different shuffle.",
		"lines.shuffle_lines(\"a\\nb\\nc\\nd\", 42)", 22, "shuffle random order lines permute", []P{sIn, pi("seed", "Seed value.", 0)},
		func(ls []string, a tools.Args) ([]string, error) {
			out := append([]string(nil), ls...)
			r := rand.New(rand.NewPCG(uint64(a.Int("seed")), 0x7a61))
			r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
			return out, nil
		})
	regStr("count_lines", "Count lines", "Counts lines. A single trailing line break does not add a line; empty text has 0 lines. skip_empty ignores blank lines.", "int",
		"lines.count_lines(\"a\\n\\nb\")", 4, "count lines number wc -l how many",
		[]P{sIn, pb("skip_empty", "Do not count blank lines.", false)},
		func(t text, a tools.Args) (any, error) {
			n := 0
			for _, l := range t.lines {
				if !a.Bool("skip_empty") || !isBlank(l) {
					n++
				}
			}
			return int64(n), nil
		})
	reg("number_lines", "Number lines", "Prefixes each line with its number followed by sep (default '. '). With align true the numbers are right-aligned to the same width.",
		"lines.number_lines(\"a\\nb\", 1, \". \")", 5, "number lines index enumerate line numbers nl cat -n",
		[]P{sIn, pi("start", "First number.", 1), pso("sep", "Text between number and line.", ". "), pb("align", "Right-align numbers to equal width.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			w := 0
			if a.Bool("align") {
				w = len(fmt.Sprint(a.Int("start") + int64(len(ls)) - 1))
				if w2 := len(fmt.Sprint(a.Int("start"))); w2 > w {
					w = w2
				}
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				out[i] = fmt.Sprintf("%*d", w, a.Int("start")+int64(i)) + a.Str("sep") + l
			}
			return out, nil
		})
	reg("remove_empty_lines", "Remove empty lines", "Deletes blank lines. By default lines containing only whitespace count as blank; set whitespace_only to false to delete only truly empty lines.",
		"lines.remove_empty_lines(\"a\\n\\nb\")", 6, "remove empty blank lines delete clean squeeze",
		[]P{sIn, pb("whitespace_only", "Also remove lines containing only whitespace.", true)},
		func(ls []string, a tools.Args) ([]string, error) {
			out := make([]string, 0, len(ls))
			for _, l := range ls {
				if l == "" || (a.Bool("whitespace_only") && isBlank(l)) {
					continue
				}
				out = append(out, l)
			}
			return out, nil
		})
	reg("trim_lines", "Trim lines", "Trims whitespace from each line (mode both, left or right).",
		"lines.trim_lines(\"  a \\n b\")", 7, "trim strip whitespace lines each leading trailing",
		[]P{sIn, pso("mode", "Which ends to trim.", "both", "both", "left", "right")},
		func(ls []string, a tools.Args) ([]string, error) {
			out := make([]string, len(ls))
			for i, l := range ls {
				v, err := trimLine(l, a.Str("mode"))
				if err != nil {
					return nil, err
				}
				out[i] = v
			}
			return out, nil
		})
	regStr("join_lines", "Join lines", "Joins all lines into one line, replacing line breaks with sep (default a space). skip_empty drops blank lines first. The result has no trailing line break.", "str",
		"lines.join_lines(\"a\\nb\\nc\", \", \")", 8, "join lines merge combine concatenate implode comma separate single line",
		[]P{sIn, pso("sep", "Text placed between lines.", " "), pb("skip_empty", "Drop blank lines.", false)},
		func(t text, a tools.Args) (any, error) {
			ls := t.lines
			if a.Bool("skip_empty") {
				ls = nil
				for _, l := range t.lines {
					if !isBlank(l) {
						ls = append(ls, l)
					}
				}
			}
			out := strings.Join(ls, a.Str("sep"))
			if len(out) > maxOut {
				return nil, errors.New("result too large")
			}
			return out, nil
		})
	regStr("split_to_lines", "Split into lines", "Replaces every occurrence of sep with a line break (default comma); trim removes surrounding whitespace from each piece. An empty sep is an error.", "str",
		"lines.split_to_lines(\"a, b, c\", \",\", True)", 9, "split lines separator explode comma break newline delimiter",
		[]P{sIn, pso("sep", "Separator to split on.", ","), pb("trim", "Trim each piece.", false)},
		func(t text, a tools.Args) (any, error) {
			if a.Str("sep") == "" {
				return nil, errors.New("sep must not be empty")
			}
			s := strings.Join(t.lines, "\n")
			parts := strings.Split(s, a.Str("sep"))
			if a.Bool("trim") {
				for i, p := range parts {
					parts[i] = strings.TrimSpace(p)
				}
			}
			out := strings.Join(parts, t.eol)
			if len(out) > maxOut {
				return nil, errors.New("result too large")
			}
			return out, nil
		})
	reg("prefix_lines", "Prefix lines", "Adds text to the start of every line. skip_empty leaves blank lines untouched.",
		"lines.prefix_lines(\"a\\nb\", \"> \")", 10, "prefix add start each line prepend quote", []P{sIn, pso("prefix", "Text to add.", "> "), pb("skip_empty", "Skip blank lines.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			return surround(ls, a.Str("prefix"), "", a.Bool("skip_empty"))
		})
	reg("suffix_lines", "Suffix lines", "Adds text to the end of every line. skip_empty leaves blank lines untouched.",
		"lines.suffix_lines(\"a\\nb\", \";\")", 11, "suffix add end each line append semicolon", []P{sIn, pso("suffix", "Text to add.", ";"), pb("skip_empty", "Skip blank lines.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			return surround(ls, "", a.Str("suffix"), a.Bool("skip_empty"))
		})
	reg("surround_lines", "Surround lines", "Wraps every line between a prefix and a suffix (default double quotes). skip_empty leaves blank lines untouched.",
		"lines.surround_lines(\"a\\nb\", \"'\", \"'\")", 12, "surround wrap quote each line double quotes single brackets",
		[]P{sIn, pso("prefix", "Text before each line.", "\""), pso("suffix", "Text after each line.", "\""), pb("skip_empty", "Skip blank lines.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			return surround(ls, a.Str("prefix"), a.Str("suffix"), a.Bool("skip_empty"))
		})
	reg("indent_lines", "Indent lines", "Indents non-empty lines by width spaces (or tabs when use_tabs is true).",
		"lines.indent_lines(\"a\\nb\", 2)", 13, "indent tab spaces lines shift right", []P{sIn, pi("width", "Spaces (or tabs) to add.", 4), pb("use_tabs", "Indent with tab characters.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			if a.Int("width") < 0 || a.Int("width") > 1000 {
				return nil, errors.New("width must be between 0 and 1000")
			}
			unit := " "
			if a.Bool("use_tabs") {
				unit = "\t"
			}
			return surround(ls, strings.Repeat(unit, int(a.Int("width"))), "", true)
		})
	reg("unindent_lines", "Unindent lines", "Removes up to width leading spaces (or one tab per step) from each line; with width 0 removes the indentation common to all non-blank lines.",
		"lines.unindent_lines(\"    a\\n    b\", 4)", 14, "unindent dedent outdent remove indentation lines",
		[]P{sIn, pi("width", "Columns to remove; 0 removes the common indentation.", 0)},
		func(ls []string, a tools.Args) ([]string, error) {
			w := int(a.Int("width"))
			if a.Int("width") < 0 {
				return nil, errors.New("width must not be negative")
			}
			if w == 0 {
				w = 1 << 30
				for _, l := range ls {
					if !isBlank(l) {
						w = min(w, len(l)-len(strings.TrimLeft(l, " \t")))
					}
				}
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				n := 0
				for n < w && n < len(l) && (l[n] == ' ' || l[n] == '\t') {
					n++
				}
				out[i] = l[n:]
			}
			return out, nil
		})
	filter := func(name, title, doc, ex, tags string, param string, test func(l, v string) bool) {
		reg(name, title, doc, ex, 15, tags,
			[]P{sIn, ps(param, "Text to look for."), pb("invert", "Keep the lines that do NOT match.", false), pb("ignore_case", "Ignore letter case.", false)},
			func(ls []string, a tools.Args) ([]string, error) {
				v := fold(a.Str(param), a.Bool("ignore_case"))
				var out []string
				for _, l := range ls {
					if test(fold(l, a.Bool("ignore_case")), v) != a.Bool("invert") {
						out = append(out, l)
					}
				}
				return out, nil
			})
	}
	filter("filter_contains", "Keep lines containing", "Keeps the lines that contain substr (or, with invert, the lines that do not).", "lines.filter_contains(\"apple\\nbanana\", \"an\")", "filter grep contains match search keep lines", "substr", strings.Contains)
	filter("filter_prefix", "Keep lines starting with", "Keeps the lines that start with prefix (or, with invert, the lines that do not).", "lines.filter_prefix(\"#a\\nb\", \"#\")", "filter starts with prefix begin keep lines", "prefix", strings.HasPrefix)
	filter("filter_suffix", "Keep lines ending with", "Keeps the lines that end with suffix (or, with invert, the lines that do not).", "lines.filter_suffix(\"a.go\\nb.py\", \".go\")", "filter ends with suffix extension keep lines", "suffix", strings.HasSuffix)
	reg("head", "First lines", "Keeps the first n lines (all of them if fewer). n of 0 gives empty text.",
		"lines.head(\"a\\nb\\nc\", 2)", 16, "head first lines top take n", []P{sIn, pi("n", "Number of lines.", 10)},
		func(ls []string, a tools.Args) ([]string, error) {
			if err := checkN(a.Int("n"), "n"); err != nil {
				return nil, err
			}
			return ls[:clampTo(a.Int("n"), len(ls))], nil
		})
	reg("tail", "Last lines", "Keeps the last n lines (all of them if fewer). n of 0 gives empty text.",
		"lines.tail(\"a\\nb\\nc\", 2)", 17, "tail last lines bottom end take n", []P{sIn, pi("n", "Number of lines.", 10)},
		func(ls []string, a tools.Args) ([]string, error) {
			if err := checkN(a.Int("n"), "n"); err != nil {
				return nil, err
			}
			return ls[len(ls)-clampTo(a.Int("n"), len(ls)):], nil
		})
	regStr("nth_line", "Nth line", "Returns line n, counting from 1; a negative n counts from the end (-1 is the last line). Out of range is an error.", "str",
		"lines.nth_line(\"a\\nb\\nc\", 2)", 18, "nth line get line number pick index row", []P{sIn, pi("n", "1-based line number, negative from the end.", 1)},
		func(t text, a tools.Args) (any, error) {
			n, l := a.Int("n"), int64(len(t.lines))
			if n < 0 {
				n += l + 1
			}
			if n < 1 || n > l {
				return nil, fmt.Errorf("line %d is out of range (text has %d lines)", a.Int("n"), l)
			}
			return t.lines[n-1], nil
		})
	reg("line_range", "Line range", "Keeps lines from..to, 1-based and inclusive; to of 0 means up to the last line; negative numbers count from the end. Out-of-range bounds are clamped.",
		"lines.line_range(\"a\\nb\\nc\\nd\", 2, 3)", 19, "line range slice lines from to between sed", []P{sIn, pi("from", "First line (1-based).", 1), pi("to", "Last line (inclusive); 0 for the end.", 0)},
		func(ls []string, a tools.Args) ([]string, error) {
			l := int64(len(ls))
			from, to := a.Int("from"), a.Int("to")
			if from < 0 {
				from += l + 1
			}
			if to == 0 {
				to = l
			} else if to < 0 {
				to += l + 1
			}
			from, to = max(from, 1), min(to, l)
			if from > to {
				return nil, nil
			}
			return ls[from-1 : to], nil
		})
	reg("skip", "Skip lines", "Drops the first n lines and keeps the rest.",
		"lines.skip(\"a\\nb\\nc\", 1)", 23, "skip drop first lines offset tail from", []P{sIn, pi("n", "Lines to skip.", 1)},
		func(ls []string, a tools.Args) ([]string, error) {
			if err := checkN(a.Int("n"), "n"); err != nil {
				return nil, err
			}
			return ls[clampTo(a.Int("n"), len(ls)):], nil
		})
	reg("take_every", "Every nth line", "Keeps every nth line starting with line offset+1 (offset defaults to 0, so n=2 keeps lines 1, 3, 5...). n must be at least 1.",
		"lines.take_every(\"a\\nb\\nc\\nd\", 2)", 24, "every nth line step sample stride pick alternate", []P{sIn, pi("n", "Step.", 2), pi("offset", "Lines to skip before the first kept one.", 0)},
		func(ls []string, a tools.Args) ([]string, error) {
			if a.Int("n") < 1 || a.Int("offset") < 0 {
				return nil, errors.New("n must be at least 1 and offset must not be negative")
			}
			var out []string
			for i := a.Int("offset"); i < int64(len(ls)); i += a.Int("n") {
				out = append(out, ls[i])
			}
			return out, nil
		})
	reg("dedupe_consecutive", "Remove consecutive duplicates", "Collapses runs of identical adjacent lines into one (like uniq without -d); non-adjacent duplicates stay.",
		"lines.dedupe_consecutive(\"a\\na\\nb\\na\")", 25, "dedupe consecutive adjacent duplicates uniq repeated runs",
		[]P{sIn, pb("case_insensitive", "Compare ignoring case.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			var out []string
			for i, l := range ls {
				if i == 0 || fold(l, a.Bool("case_insensitive")) != fold(ls[i-1], a.Bool("case_insensitive")) {
					out = append(out, l)
				}
			}
			return out, nil
		})
	regStr("min_max_lines", "Min and max lines", "Finds the extreme lines by mode: length (shortest and longest, in characters), text (first and last in code point order) or numeric (smallest and largest leading number). Returns a dict with count, min and max (empty strings for empty input; lines without a number are ignored in numeric mode).", "dict[str, str]",
		"lines.min_max_lines(\"bb\\na\\nccc\")", 26, "min max shortest longest smallest largest lines extreme",
		[]P{sIn, pso("mode", "How lines are compared.", "length", "length", "text", "numeric")},
		func(t text, a tools.Args) (any, error) {
			mode := strings.ToLower(a.Str("mode"))
			ls := t.lines
			if mode == "numeric" {
				ls = nil
				for _, l := range t.lines {
					if _, ok := leadingNumber(l); ok {
						ls = append(ls, l)
					}
				}
			} else if mode != "length" && mode != "text" {
				return nil, fmt.Errorf("unknown mode %q (use length, text or numeric)", a.Str("mode"))
			}
			res := map[string]string{"count": itoa(len(ls)), "min": "", "max": ""}
			if len(ls) == 0 {
				return res, nil
			}
			less := func(x, y string) bool {
				switch mode {
				case "length":
					return utf8.RuneCountInString(x) < utf8.RuneCountInString(y)
				case "numeric":
					fx, _ := leadingNumber(x)
					fy, _ := leadingNumber(y)
					return fx < fy
				}
				return x < y
			}
			mn, mx := ls[0], ls[0]
			for _, l := range ls[1:] {
				if less(l, mn) {
					mn = l
				}
				if less(mx, l) {
					mx = l
				}
			}
			res["min"], res["max"] = mn, mx
			return res, nil
		})
}

func surround(ls []string, pre, suf string, skipEmpty bool) ([]string, error) {
	if int64(len(ls))*int64(len(pre)+len(suf)) > maxOut {
		return nil, errors.New("result would exceed 4 MiB")
	}
	out := make([]string, len(ls))
	for i, l := range ls {
		if skipEmpty && isBlank(l) {
			out[i] = l
			continue
		}
		out[i] = pre + l + suf
	}
	return out, nil
}
