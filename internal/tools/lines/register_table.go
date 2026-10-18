package lines

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

// colSel selects columns: "2", "-1", "2-4" or "3-" (1-based, inclusive).
type colSel struct {
	from, to     int
	open, single bool
}

var selRe = regexp.MustCompile(`^\s*(-?\d+)\s*(?:(-)\s*(\d*))?\s*$`)

func parseSels(specs []string) ([]colSel, error) {
	var out []colSel
	for _, sp := range specs {
		for _, part := range strings.Split(sp, ",") {
			if strings.TrimSpace(part) == "" {
				continue
			}
			m := selRe.FindStringSubmatch(part)
			if m == nil {
				return nil, fmt.Errorf("bad column spec %q (use 2, -1, 2-4 or 3-)", part)
			}
			from, _ := strconv.Atoi(m[1])
			if from == 0 {
				return nil, errors.New("columns are numbered from 1")
			}
			c := colSel{from: from, single: m[2] == ""}
			if m[2] != "" {
				if m[3] == "" {
					c.open = true
				} else {
					c.to, _ = strconv.Atoi(m[3])
				}
			}
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("indexes must not be empty")
	}
	return out, nil
}

func resolve(c colSel, n int) (lo, hi int) {
	lo = c.from
	if lo < 0 {
		lo = n + lo + 1
	}
	switch {
	case c.single:
		hi = lo
	case c.open:
		hi = n
	default:
		hi = c.to
	}
	return max(lo, 1), min(hi, n)
}

func pickColumns(row []string, sels []colSel) []string {
	var out []string
	for _, s := range sels {
		lo, hi := resolve(s, len(row))
		for i := lo; i <= hi; i++ {
			out = append(out, row[i-1])
		}
	}
	return out
}

func csvQuote(f, delim string) string {
	if strings.ContainsAny(f, "\"\r\n") || (delim != "" && strings.Contains(f, delim)) {
		return `"` + strings.ReplaceAll(f, `"`, `""`) + `"`
	}
	return f
}

func csvDelim(d string) (rune, error) {
	r, n := utf8.DecodeRuneInString(d)
	if n == 0 || n != len(d) {
		return 0, errors.New("delimiter must be exactly one character")
	}
	if r == '"' || r == '\r' || r == '\n' || r == utf8.RuneError {
		return 0, errors.New("invalid CSV delimiter")
	}
	return r, nil
}

func parseCSV(s, delim string) ([][]string, error) {
	d, err := csvDelim(delim)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = d
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	return r.ReadAll()
}

func toNumbers(ls []string, strict bool) ([]float64, error) {
	var out []float64
	for i, l := range ls {
		if strings.TrimSpace(l) == "" {
			continue
		}
		f, ok := parseNumber(l)
		if !ok {
			if strict {
				return nil, fmt.Errorf("line %d is not a number: %q", i+1, l)
			}
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

func setOp(name, title, doc, ex string, order int, tags string, f func(a, b []string, ci bool) []string) {
	regRaw(name, title, doc+" Uses the line ending style and trailing newline of a.", "str", ex, order, tags,
		[]P{ps("a", "First text, one item per line."), ps("b", "Second text, one item per line."), pb("ignore_case", "Compare lines ignoring case.", false)},
		func(a tools.Args) (any, error) {
			ta, tb := split(a.Str("a")), split(a.Str("b"))
			return ta.join(f(ta.lines, tb.lines, a.Bool("ignore_case")))
		})
}

func set(ls []string, ci bool) map[string]bool {
	m := make(map[string]bool, len(ls))
	for _, l := range ls {
		m[fold(l, ci)] = true
	}
	return m
}

func wordWrap(line string, width int, breakLong bool) []string {
	ws := strings.Fields(line)
	if len(ws) == 0 {
		return []string{""}
	}
	var out []string
	cur, curN := "", 0
	for _, w := range ws {
		wn := utf8.RuneCountInString(w)
		for breakLong && wn > width {
			if curN > 0 {
				out = append(out, cur)
				cur, curN = "", 0
			}
			rs := []rune(w)
			out = append(out, string(rs[:width]))
			w = string(rs[width:])
			wn -= width
		}
		switch {
		case curN == 0:
			cur, curN = w, wn
		case curN+1+wn <= width:
			cur += " " + w
			curN += 1 + wn
		default:
			out = append(out, cur)
			cur, curN = w, wn
		}
	}
	if curN > 0 {
		out = append(out, cur)
	}
	return out
}

func init() {
	delimP := func(def string) P {
		return pso("delimiter", "Column delimiter; empty means runs of whitespace.", def)
	}
	reg("transpose", "Transpose rows and columns", "Turns rows into columns: each line is split by the delimiter and the table is flipped. Short rows are padded with empty cells.",
		"lines.transpose(\"a,b\\nc,d\", \",\")", 40, "transpose rotate matrix rows columns flip table swap axes",
		[]P{sIn, delimP(",")},
		func(ls []string, a tools.Args) ([]string, error) {
			d := a.Str("delimiter")
			rows := make([][]string, len(ls))
			w := 0
			for i, l := range ls {
				rows[i] = fields(l, d)
				w = max(w, len(rows[i]))
			}
			if int64(w)*int64(len(rows)) > 2_000_000 {
				return nil, errors.New("table too large")
			}
			out := make([]string, w)
			for j := 0; j < w; j++ {
				cells := make([]string, len(rows))
				for i := range rows {
					if j < len(rows[i]) {
						cells[i] = rows[i][j]
					}
				}
				out[j] = joinDelim(cells, d)
			}
			return out, nil
		})
	reg("columns", "Pick columns", "Keeps selected columns of each line. indexes are 1-based: '2', '-1' (last), '2-4' or '3-' (to the end); comma-separated entries also work. Rows lacking a column skip it. Output is joined with output_delimiter (default: the input delimiter, or a space for whitespace splitting).",
		"lines.columns(\"a,b,c\", \",\", [\"1\", \"3\"])", 41, "columns pick cut select fields awk extract column csv",
		[]P{sIn, delimP(","), pls("indexes", "Columns to keep.", []string{"1"}), pso("output_delimiter", "Joiner for the kept columns; empty reuses the delimiter.", "")},
		func(ls []string, a tools.Args) ([]string, error) {
			sels, err := parseSels(a.List("indexes"))
			if err != nil {
				return nil, err
			}
			od := a.Str("output_delimiter")
			if od == "" {
				od = a.Str("delimiter")
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				out[i] = joinDelim(pickColumns(fields(l, a.Str("delimiter")), sels), od)
			}
			return out, nil
		})
	reg("align_columns", "Align columns", "Pads the cells of each column to equal width so the lines form a table. Widths count characters; trailing padding is removed. align is left or right.",
		"lines.align_columns(\"a,bbb\\ncc,d\", \",\", \" | \")", 42, "align columns table pad format pretty print column",
		[]P{sIn, delimP(","), pso("separator", "Text between aligned columns.", "  "), pso("align", "left or right.", "left", "left", "right")},
		func(ls []string, a tools.Args) ([]string, error) {
			rows := make([][]string, len(ls))
			var widths []int
			for i, l := range ls {
				rows[i] = fields(l, a.Str("delimiter"))
				for j, c := range rows[i] {
					if j >= len(widths) {
						widths = append(widths, 0)
					}
					widths[j] = max(widths[j], utf8.RuneCountInString(c))
				}
			}
			right := strings.EqualFold(a.Str("align"), "right")
			if !right && !strings.EqualFold(a.Str("align"), "left") {
				return nil, errors.New("align must be left or right")
			}
			out := make([]string, len(ls))
			for i, r := range rows {
				cells := make([]string, len(r))
				for j, c := range r {
					pad := strings.Repeat(" ", widths[j]-utf8.RuneCountInString(c))
					if right {
						cells[j] = pad + c
					} else {
						cells[j] = c + pad
					}
				}
				out[i] = strings.TrimRight(strings.Join(cells, a.Str("separator")), " ")
			}
			return out, nil
		})
	reg("swap_columns", "Swap columns", "Swaps two 1-based columns (negative counts from the end) on every line. Lines that lack either column are left unchanged.",
		"lines.swap_columns(\"a,b,c\", 1, 3, \",\")", 43, "swap columns exchange reorder fields switch",
		[]P{sIn, pi("col_a", "First column.", 1), pi("col_b", "Second column.", 2), delimP(",")},
		func(ls []string, a tools.Args) ([]string, error) {
			if a.Int("col_a") == 0 || a.Int("col_b") == 0 {
				return nil, errors.New("columns are numbered from 1")
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				f := fields(l, a.Str("delimiter"))
				x, y := int(a.Int("col_a")), int(a.Int("col_b"))
				if x < 0 {
					x += len(f) + 1
				}
				if y < 0 {
					y += len(f) + 1
				}
				if x < 1 || y < 1 || x > len(f) || y > len(f) {
					out[i] = l
					continue
				}
				f[x-1], f[y-1] = f[y-1], f[x-1]
				out[i] = joinDelim(f, a.Str("delimiter"))
			}
			return out, nil
		})
	reg("sort_by_column", "Sort by column", "Sorts lines by one 1-based column (negative from the end); lines lacking the column sort as empty. mode and order as in sort_lines. The sort is stable.",
		"lines.sort_by_column(\"b,2\\na,1\", 2, \",\", \"numeric\")", 44, "sort by column field key csv table order",
		[]P{sIn, pi("column", "Column to sort by.", 1), delimP(""), pso("mode", "How keys are compared.", "text", sortModes...), pso("order", "asc or desc.", "asc", "asc", "desc")},
		func(ls []string, a tools.Args) ([]string, error) {
			desc, err := orderIsDesc(a.Str("order"))
			if err != nil {
				return nil, err
			}
			cmp, err := compareFor(a.Str("mode"), desc)
			if err != nil {
				return nil, err
			}
			col := int(a.Int("column"))
			if col == 0 {
				return nil, errors.New("columns are numbered from 1")
			}
			key := func(l string) string {
				f := fields(l, a.Str("delimiter"))
				c := col
				if c < 0 {
					c += len(f) + 1
				}
				if c < 1 || c > len(f) {
					return ""
				}
				return f[c-1]
			}
			out := append([]string(nil), ls...)
			keys := make(map[string]string, len(out))
			for _, l := range out {
				keys[l] = key(l)
			}
			sort.SliceStable(out, func(i, j int) bool { return cmp(keys[out[i]], keys[out[j]]) < 0 })
			return out, nil
		})

	// numbers
	regStr("sum_numbers", "Sum of numbers", "Adds up the numbers written one per line (thousands separators like 1,234.5 are accepted). Blank and non-numeric lines are ignored unless strict is true, in which case they are an error. No numbers gives 0.", "float",
		"lines.sum_numbers(\"1\\n2.5\\n3\")", 45, "sum total add numbers lines calculate",
		[]P{sIn, pb("strict", "Fail on lines that are not numbers.", false)},
		func(t text, a tools.Args) (any, error) {
			ns, err := toNumbers(t.lines, a.Bool("strict"))
			if err != nil {
				return nil, err
			}
			sum := 0.0
			for _, n := range ns {
				sum += n
			}
			return sum, nil
		})
	regStr("average_numbers", "Average of numbers", "Arithmetic mean of the numbers written one per line. Blank and non-numeric lines are ignored unless strict is true; with no numbers it is an error.", "float",
		"lines.average_numbers(\"1\\n2\\n6\")", 46, "average mean avg numbers lines calculate",
		[]P{sIn, pb("strict", "Fail on lines that are not numbers.", false)},
		func(t text, a tools.Args) (any, error) {
			ns, err := toNumbers(t.lines, a.Bool("strict"))
			if err != nil {
				return nil, err
			}
			if len(ns) == 0 {
				return nil, errors.New("no numbers found")
			}
			sum := 0.0
			for _, n := range ns {
				sum += n
			}
			return sum / float64(len(ns)), nil
		})

	// CSV / TSV / markdown
	regStr("list_to_csv_row", "Lines to CSV row", "Joins the lines into one RFC 4180 CSV row, quoting fields that contain the delimiter, a quote or a line break (quotes are doubled). skip_empty drops blank lines.", "str",
		"lines.list_to_csv_row(\"a\\nb,c\")", 47, "csv row list join comma separated rfc4180 quote",
		[]P{sIn, pso("delimiter", "Field delimiter.", ","), pb("skip_empty", "Drop blank lines.", false)},
		func(t text, a tools.Args) (any, error) {
			var f []string
			for _, l := range t.lines {
				if a.Bool("skip_empty") && isBlank(l) {
					continue
				}
				f = append(f, csvQuote(l, a.Str("delimiter")))
			}
			return strings.Join(f, a.Str("delimiter")), nil
		})
	regStr("csv_row_to_lines", "CSV row to lines", "Parses CSV text (normally one row) and returns each field on its own line. Handles quoted fields with embedded delimiters, quotes and line breaks; with several rows, all fields are listed in order.", "str",
		"lines.csv_row_to_lines(\"a,\\\"b,c\\\",d\")", 48, "csv row split fields lines parse explode",
		[]P{sIn, pso("delimiter", "Field delimiter (one character).", ",")},
		func(t text, a tools.Args) (any, error) {
			recs, err := parseCSV(strings.Join(t.lines, "\n"), a.Str("delimiter"))
			if err != nil {
				return nil, err
			}
			var out []string
			for _, r := range recs {
				out = append(out, r...)
			}
			return strings.Join(out, t.eol), nil
		})
	regStr("csv_to_tsv", "CSV to TSV", "Converts CSV to tab-separated text. Tabs and line breaks inside fields are replaced by single spaces because TSV has no quoting. delimiter is the CSV delimiter (one character).", "str",
		"lines.csv_to_tsv(\"a,\\\"b,c\\\"\\n1,2\")", 49, "csv tsv convert tab separated spreadsheet",
		[]P{sIn, pso("delimiter", "CSV delimiter.", ",")},
		func(t text, a tools.Args) (any, error) {
			recs, err := parseCSV(strings.Join(t.lines, "\n"), a.Str("delimiter"))
			if err != nil {
				return nil, err
			}
			rep := strings.NewReplacer("\t", " ", "\r\n", " ", "\n", " ", "\r", " ")
			ls := make([]string, len(recs))
			for i, r := range recs {
				for j := range r {
					r[j] = rep.Replace(r[j])
				}
				ls[i] = strings.Join(r, "\t")
			}
			return t.join(ls)
		})
	regStr("tsv_to_csv", "TSV to CSV", "Converts tab-separated text to RFC 4180 CSV: fields containing the delimiter, quotes or CR are quoted with doubled quotes. delimiter is the output delimiter (one character).", "str",
		"lines.tsv_to_csv(\"a\\tb,c\\n1\\t2\")", 50, "tsv csv convert tab separated spreadsheet",
		[]P{sIn, pso("delimiter", "CSV delimiter to produce.", ",")},
		func(t text, a tools.Args) (any, error) {
			if _, err := csvDelim(a.Str("delimiter")); err != nil {
				return nil, err
			}
			ls := make([]string, len(t.lines))
			for i, l := range t.lines {
				f := strings.Split(l, "\t")
				for j := range f {
					f[j] = csvQuote(f[j], a.Str("delimiter"))
				}
				ls[i] = strings.Join(f, a.Str("delimiter"))
			}
			return t.join(ls)
		})
	regStr("markdown_table_from_csv", "Markdown table from CSV", "Builds a GitHub-style Markdown table from CSV (or any single-character delimiter such as a tab). With header true the first row is the header; otherwise the header is Column 1, Column 2 ... Short rows are padded and | in cells is escaped.", "str",
		"lines.markdown_table_from_csv(\"name,age\\nAnn,30\")", 51, "markdown table csv convert github gfm md",
		[]P{sIn, pso("delimiter", "CSV delimiter.", ","), pb("header", "First row is the header.", true)},
		func(t text, a tools.Args) (any, error) {
			recs, err := parseCSV(strings.Join(t.lines, "\n"), a.Str("delimiter"))
			if err != nil {
				return nil, err
			}
			w := 0
			for _, r := range recs {
				w = max(w, len(r))
			}
			if w == 0 {
				return "", nil
			}
			if !a.Bool("header") {
				h := make([]string, w)
				for i := range h {
					h[i] = "Column " + itoa(i+1)
				}
				recs = append([][]string{h}, recs...)
			}
			esc := strings.NewReplacer("|", `\|`, "\r\n", "<br>", "\n", "<br>", "\r", "<br>")
			widths := make([]int, w)
			for i := range widths {
				widths[i] = 3
			}
			for _, r := range recs {
				for j := 0; j < w; j++ {
					c := ""
					if j < len(r) {
						c = esc.Replace(r[j])
					}
					widths[j] = max(widths[j], utf8.RuneCountInString(c))
				}
			}
			row := func(r []string) string {
				cells := make([]string, w)
				for j := range cells {
					c := ""
					if j < len(r) {
						c = esc.Replace(r[j])
					}
					cells[j] = c + strings.Repeat(" ", widths[j]-utf8.RuneCountInString(c))
				}
				return "| " + strings.Join(cells, " | ") + " |"
			}
			out := []string{row(recs[0])}
			sep := make([]string, w)
			for j := range sep {
				sep[j] = strings.Repeat("-", widths[j])
			}
			out = append(out, "| "+strings.Join(sep, " | ")+" |")
			for _, r := range recs[1:] {
				out = append(out, row(r))
			}
			return strings.Join(out, "\n"), nil
		})
	reg("markdown_bullets", "Markdown bullets", "Turns each non-empty line into a Markdown bullet by prefixing the marker and a space (marker - by default; * and + also work). Blank lines are kept as they are.",
		"lines.markdown_bullets(\"a\\nb\")", 52, "markdown bullets list unordered dash asterisk md",
		[]P{sIn, pso("marker", "Bullet marker.", "-", "-", "*", "+")},
		func(ls []string, a tools.Args) ([]string, error) { return surround(ls, a.Str("marker")+" ", "", true) })
	reg("numbered_list", "Numbered list", "Turns each non-empty line into a numbered item ('1. text'); blank lines are kept and do not consume a number.",
		"lines.numbered_list(\"a\\nb\")", 53, "numbered list ordered markdown enumerate md 1.",
		[]P{sIn, pi("start", "First number.", 1), pso("sep", "Text between number and line.", ". ")},
		func(ls []string, a tools.Args) ([]string, error) {
			out := make([]string, len(ls))
			n := a.Int("start")
			for i, l := range ls {
				if isBlank(l) {
					out[i] = l
					continue
				}
				out[i] = strconv.FormatInt(n, 10) + a.Str("sep") + l
				n++
			}
			return out, nil
		})
	bulletRe := regexp.MustCompile(`^(\s*)(?:[-*+\x{2022}]|\d+[.)])\s+(?:\[[ xX]\]\s+)?`)
	reg("remove_markdown_bullets", "Remove list markers", "Strips list markers from the start of each line: - * + and bullet characters, numbered markers like 1. or 2), and task boxes [ ] / [x]. Indentation is kept; lines without a marker are unchanged.",
		"lines.remove_markdown_bullets(\"- a\\n2. b\")", 54, "remove bullets markdown list markers strip numbers unordered",
		[]P{sIn},
		func(ls []string, a tools.Args) ([]string, error) {
			out := make([]string, len(ls))
			for i, l := range ls {
				out[i] = bulletRe.ReplaceAllString(l, "$1")
			}
			return out, nil
		})

	// set operations
	regRaw("lines_diff", "Compare two line lists", "Compares two texts as sets of lines and returns a dict with only_in_a, only_in_b and common (each a newline-joined list in order of first appearance, without duplicates).", "dict[str, str]",
		"lines.lines_diff(\"a\\nb\", \"b\\nc\")", 55, "diff compare lines difference set only in common intersect",
		[]P{ps("a", "First text, one item per line."), ps("b", "Second text, one item per line."), pb("ignore_case", "Compare lines ignoring case.", false)},
		func(a tools.Args) (any, error) {
			ci := a.Bool("ignore_case")
			la, lb := uniq(split(a.Str("a")).lines, ci), uniq(split(a.Str("b")).lines, ci)
			sa, sb := set(la, ci), set(lb, ci)
			var oa, ob, com []string
			for _, l := range la {
				if sb[fold(l, ci)] {
					com = append(com, l)
				} else {
					oa = append(oa, l)
				}
			}
			for _, l := range lb {
				if !sa[fold(l, ci)] {
					ob = append(ob, l)
				}
			}
			j := func(l []string) string { return strings.Join(l, "\n") }
			return map[string]string{"only_in_a": j(oa), "only_in_b": j(ob), "common": j(com)}, nil
		})
	setOp("intersect_lines", "Intersect lines", "Lines of a that also appear in b (duplicates removed, order of a).", "lines.intersect_lines(\"a\\nb\", \"b\\nc\")", 56, "intersect common both lines set and",
		func(a, b []string, ci bool) []string {
			sb := set(b, ci)
			var out []string
			for _, l := range uniq(a, ci) {
				if sb[fold(l, ci)] {
					out = append(out, l)
				}
			}
			return out
		})
	setOp("union_lines", "Union of lines", "All distinct lines of a followed by the new lines of b.", "lines.union_lines(\"a\\nb\", \"b\\nc\")", 57, "union merge combine lines set or distinct",
		func(a, b []string, ci bool) []string { return uniq(append(append([]string(nil), a...), b...), ci) })
	setOp("subtract_lines", "Subtract lines", "Distinct lines of a that do not appear in b.", "lines.subtract_lines(\"a\\nb\", \"b\\nc\")", 58, "subtract difference except minus remove lines set not in",
		func(a, b []string, ci bool) []string {
			sb := set(b, ci)
			var out []string
			for _, l := range uniq(a, ci) {
				if !sb[fold(l, ci)] {
					out = append(out, l)
				}
			}
			return out
		})

	// wrapping
	wrapP := func() []P {
		return []P{sIn, pi("width", "Maximum line width in characters.", 72)}
	}
	reg("wrap_each", "Hard-wrap each line", "Splits every line into chunks of exactly width characters (the last chunk may be shorter), ignoring word boundaries. Blank lines are kept.",
		"lines.wrap_each(\"abcdefgh\", 3)", 59, "wrap hard wrap chunk split width columns fixed each line",
		wrapP(), func(ls []string, a tools.Args) ([]string, error) {
			if a.Int("width") < 1 {
				return nil, errors.New("width must be at least 1")
			}
			w := int(a.Int("width"))
			var out []string
			for _, l := range ls {
				rs := []rune(l)
				if len(rs) == 0 {
					out = append(out, "")
				}
				for len(rs) > 0 {
					n := min(w, len(rs))
					out = append(out, string(rs[:n]))
					rs = rs[n:]
				}
			}
			return out, nil
		})
	reg("word_wrap_lines", "Word-wrap each line", "Word-wraps every line separately to width characters, breaking only between words (runs of whitespace collapse to one space); blank lines are kept, over-long words stay whole unless break_long is true.",
		"lines.word_wrap_lines(\"the quick brown fox\", 10)", 60, "word wrap reflow width paragraph fill lines soft",
		append(wrapP(), pb("break_long", "Split words longer than width.", false)),
		func(ls []string, a tools.Args) ([]string, error) {
			if a.Int("width") < 1 {
				return nil, errors.New("width must be at least 1")
			}
			var out []string
			for _, l := range ls {
				out = append(out, wordWrap(l, int(a.Int("width")), a.Bool("break_long"))...)
			}
			return out, nil
		})

	// structured output
	regStr("to_json_array", "Lines to JSON array", "Converts the lines to a JSON array of strings (valid UTF-8, no HTML escaping). skip_empty drops blank lines; pretty indents one item per line.", "str",
		"lines.to_json_array(\"a\\nb\")", 61, "json array lines convert list strings stringify",
		[]P{sIn, pb("skip_empty", "Drop blank lines.", false), pb("pretty", "Indent the output.", false)},
		func(t text, a tools.Args) (any, error) {
			arr := make([]string, 0, len(t.lines))
			for _, l := range t.lines {
				if !(a.Bool("skip_empty") && isBlank(l)) {
					arr = append(arr, l)
				}
			}
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			if a.Bool("pretty") {
				enc.SetIndent("", "  ")
			}
			if err := enc.Encode(arr); err != nil {
				return nil, err
			}
			return strings.TrimRight(buf.String(), "\n"), nil
		})
	regStr("from_json_array", "JSON array to lines", "Reads a JSON array and puts each element on its own line: strings unquoted, other values (numbers, booleans, null, objects) as compact JSON. Anything that is not a JSON array is an error.", "str",
		"lines.from_json_array('[\"a\", \"b\", 3]')", 62, "json array to lines parse list explode strings",
		[]P{sIn},
		func(t text, a tools.Args) (any, error) {
			var raw []json.RawMessage
			if err := json.Unmarshal([]byte(a.Str("s")), &raw); err != nil {
				return nil, fmt.Errorf("not a JSON array: %v", err)
			}
			out := make([]string, len(raw))
			for i, r := range raw {
				if len(r) > 0 && r[0] == '"' {
					var s string
					if err := json.Unmarshal(r, &s); err != nil {
						return nil, err
					}
					out[i] = s
					continue
				}
				var b bytes.Buffer
				if err := json.Compact(&b, r); err != nil {
					return nil, err
				}
				out[i] = b.String()
			}
			return strings.Join(out, "\n"), nil
		})
	regStr("to_sql_in_list", "Lines to SQL IN list", "Turns the lines into a SQL list such as 'a','b' with single quotes doubled for escaping. numbers=true leaves values unquoted when every item is a number; parens wraps the list in (...); blank lines are dropped unless skip_empty is false.", "str",
		"lines.to_sql_in_list(\"a\\nb\")", 63, "sql in list quote values where clause database query strings",
		[]P{sIn, pso("sep", "Separator between items.", ","), pb("parens", "Wrap the list in parentheses.", false), pb("numbers", "Do not quote numeric items.", false), pb("skip_empty", "Drop blank lines.", true)},
		func(t text, a tools.Args) (any, error) {
			var items []string
			allNum := true
			for _, l := range t.lines {
				if a.Bool("skip_empty") && isBlank(l) {
					continue
				}
				items = append(items, l)
				if _, ok := parseNumber(l); !ok || strings.Contains(l, ",") {
					allNum = false
				}
			}
			for i, it := range items {
				if a.Bool("numbers") && allNum {
					items[i] = strings.TrimSpace(it)
				} else {
					items[i] = "'" + strings.ReplaceAll(it, "'", "''") + "'"
				}
			}
			out := strings.Join(items, a.Str("sep"))
			if a.Bool("parens") {
				out = "(" + out + ")"
			}
			return out, nil
		})
	regStr("to_quoted_csv", "Lines to quoted CSV", "Joins the lines into one CSV row in which every field is quoted (quotes inside are doubled), e.g. \"a\",\"b\". skip_empty drops blank lines.", "str",
		"lines.to_quoted_csv(\"a\\nb\")", 64, "quoted csv row join comma quote all fields list",
		[]P{sIn, pso("delimiter", "Field delimiter.", ","), pso("quote", "Quote character.", "\""), pb("skip_empty", "Drop blank lines.", false)},
		func(t text, a tools.Args) (any, error) {
			q := a.Str("quote")
			var f []string
			for _, l := range t.lines {
				if a.Bool("skip_empty") && isBlank(l) {
					continue
				}
				f = append(f, q+strings.ReplaceAll(l, q, q+q)+q)
			}
			if q == "" {
				return nil, errors.New("quote must not be empty")
			}
			return strings.Join(f, a.Str("delimiter")), nil
		})

	// line endings and whitespace
	regRaw("to_lf", "Convert to LF", "Converts every line ending (CRLF and lone CR) to LF (Unix); all other text is untouched.", "str",
		"lines.to_lf(\"a\\r\\nb\")", 65, "lf unix line endings newline convert dos2unix crlf normalize",
		[]P{sIn}, func(a tools.Args) (any, error) {
			return strings.ReplaceAll(strings.ReplaceAll(a.Str("s"), "\r\n", "\n"), "\r", "\n"), nil
		})
	regRaw("to_crlf", "Convert to CRLF", "Converts every line ending (LF and lone CR) to CRLF (Windows); existing CRLF is kept as one break.", "str",
		"lines.to_crlf(\"a\\nb\")", 66, "crlf windows line endings newline convert unix2dos normalize",
		[]P{sIn}, func(a tools.Args) (any, error) {
			s := strings.ReplaceAll(strings.ReplaceAll(a.Str("s"), "\r\n", "\n"), "\r", "\n")
			out := strings.ReplaceAll(s, "\n", "\r\n")
			if len(out) > maxOut {
				return nil, errors.New("result too large")
			}
			return out, nil
		})
	regRaw("detect_line_endings", "Detect line endings", "Reports the line ending style: LF, CRLF, CR, mixed (more than one style) or none (no line break at all).", "str",
		"lines.detect_line_endings(\"a\\r\\nb\")", 67, "detect line endings crlf lf cr newline type windows unix mixed",
		[]P{sIn}, func(a tools.Args) (any, error) {
			s := a.Str("s")
			crlf := strings.Count(s, "\r\n")
			lf, cr := strings.Count(s, "\n")-crlf, strings.Count(s, "\r")-crlf
			kinds := 0
			res := "none"
			for _, k := range []struct {
				n    int
				name string
			}{{lf, "LF"}, {crlf, "CRLF"}, {cr, "CR"}} {
				if k.n > 0 {
					kinds++
					res = k.name
				}
			}
			if kinds > 1 {
				res = "mixed"
			}
			return res, nil
		})
	reg("strip_trailing_whitespace", "Strip trailing whitespace", "Removes spaces, tabs and other Unicode whitespace from the end of every line.",
		"lines.strip_trailing_whitespace(\"a  \\nb\\t\")", 68, "trailing whitespace strip trim right end spaces clean rstrip",
		[]P{sIn}, func(ls []string, a tools.Args) ([]string, error) {
			out := make([]string, len(ls))
			for i, l := range ls {
				out[i] = strings.TrimRightFunc(l, unicode.IsSpace)
			}
			return out, nil
		})
	reg("tabs_to_spaces", "Tabs to spaces", "Replaces tab characters with spaces, aligning to tab stops every width columns (counted in characters). width is 1 to 16.",
		"lines.tabs_to_spaces(\"a\\tb\", 4)", 69, "tabs spaces convert expandtabs indent untabify",
		[]P{sIn, pi("width", "Tab stop width.", 4)}, func(ls []string, a tools.Args) ([]string, error) {
			w := int(a.Int("width"))
			if w < 1 || w > 16 {
				return nil, errors.New("width must be between 1 and 16")
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				var b strings.Builder
				col := 0
				for _, r := range l {
					if r == '\t' {
						n := w - col%w
						b.WriteString(strings.Repeat(" ", n))
						col += n
						continue
					}
					b.WriteRune(r)
					col++
				}
				out[i] = b.String()
			}
			return out, nil
		})
	reg("spaces_to_tabs", "Spaces to tabs", "Converts leading indentation to tabs: every width spaces at the start of a line become one tab; leftover spaces and spaces elsewhere are kept. width is 1 to 16.",
		"lines.spaces_to_tabs(\"        x\", 4)", 70, "spaces tabs convert indent tabify unexpand",
		[]P{sIn, pi("width", "Spaces per tab.", 4)}, func(ls []string, a tools.Args) ([]string, error) {
			w := int(a.Int("width"))
			if w < 1 || w > 16 {
				return nil, errors.New("width must be between 1 and 16")
			}
			out := make([]string, len(ls))
			for i, l := range ls {
				n := len(l) - len(strings.TrimLeft(l, " "))
				out[i] = strings.Repeat("\t", n/w) + strings.Repeat(" ", n%w) + l[n:]
			}
			return out, nil
		})
	reg("remove_comments", "Remove comment lines", "Deletes lines whose first non-blank characters match one of the prefixes (default # and //). Only whole-line comments are removed, not trailing ones; drop_empty also removes blank lines.",
		"lines.remove_comments(\"# note\\ncode\\n// x\")", 71, "remove comments strip hash slash config code lines delete",
		[]P{sIn, pls("prefixes", "Comment markers.", []string{"#", "//"}), pb("drop_empty", "Also remove blank lines.", false)},
		func(ls []string, a tools.Args) ([]string, error) {
			var out []string
			for _, l := range ls {
				t := strings.TrimLeftFunc(l, unicode.IsSpace)
				drop := a.Bool("drop_empty") && t == ""
				for _, p := range a.List("prefixes") {
					if p != "" && strings.HasPrefix(t, p) {
						drop = true
					}
				}
				if !drop {
					out = append(out, l)
				}
			}
			return out, nil
		})
}
