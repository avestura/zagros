// Package csvx registers the "csv" tool group: parse and write delimited
// text. Pure: it works on strings only.
package csvx

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const maxRecords = 1_000_000

func init() {
	tools.Register(tools.Func{
		Group: "csv", Name: "parse", Title: "Parse CSV",
		Doc:   "Parse CSV or TSV text. Returns a list of rows (lists of strings), or with header=True a list of dicts keyed by the first row. An empty delimiter detects comma, semicolon, tab or pipe from the first lines. Quotes may be loose unless strict=True. A UTF-8 BOM is dropped. Raises on malformed quoting in strict mode.",
		Terms: []string{"csv to list", "tsv", "spreadsheet", "read table", "split columns", "comma separated", "parse delimited text", "rows to dicts"},
		Params: []tools.Param{
			tools.Req("text", "str", "The CSV text."),
			tools.Opt("delimiter", "str", "Single character separator (use \\t for tab); empty detects it.", ""),
			tools.Opt("header", "bool", "Use the first row as keys and return dicts.", false),
			tools.Opt("trim", "bool", "Trim spaces around each cell.", false),
			tools.Opt("comment", "str", "A character that starts a comment line to skip; empty keeps all lines.", ""),
			tools.Opt("strict", "bool", "Reject bare quotes inside unquoted cells.", false),
		},
		Returns: "any",
		Example: `csv.parse("name,age\nada,36\nbob,41", header=True)`,
		Tags:    []string{"csv", "tsv", "parse", "table", "rows", "delimiter"},
		Fn:      parse,
	})
	tools.Register(tools.Func{
		Group: "csv", Name: "dump", Title: "Write CSV",
		Doc:   "Write rows as CSV text. Rows are lists of values or dicts; for dicts the columns are the keys in first-seen order unless columns is given. Cells that need it are quoted. None becomes an empty cell. Raises when a cell is a list or dict.",
		Terms: []string{"list to csv", "table to csv", "export csv", "tsv output", "rows to text", "dicts to csv"},
		Params: []tools.Param{
			tools.Req("rows", "any", "A list of lists, or a list of dicts."),
			tools.Opt("delimiter", "str", "Single character separator (use \\t for tab).", ","),
			tools.Opt("columns", "list[str]", "Column order for dict rows; empty derives it from the rows.", nil),
			tools.Opt("header", "bool", "Write a header line when rows are dicts or columns is set.", true),
			tools.Opt("crlf", "bool", "End lines with CRLF instead of LF.", false),
		},
		Returns: "str",
		Example: `csv.dump([{"name": "ada", "age": 36}, {"name": "bob", "age": 41}])`,
		Tags:    []string{"csv", "tsv", "write", "export", "table", "rows"},
		Fn:      dump,
	})
	tools.Register(tools.Func{
		Group: "csv", Name: "detect_delimiter", Title: "Detect CSV delimiter",
		Doc:     "Guess the separator of delimited text: comma, semicolon, tab or pipe, chosen by how consistently each appears across the first lines. Returns a comma when nothing fits.",
		Terms:   []string{"which delimiter", "separator", "sniff csv"},
		Params:  []tools.Param{tools.Req("text", "str", "The delimited text.")},
		Returns: "str",
		Example: `csv.detect_delimiter("a;b;c\n1;2;3")`,
		Tags:    []string{"csv", "delimiter", "separator", "sniff"},
		Fn: func(a tools.Args) (any, error) {
			return string(detect(a.Str("text"))), nil
		},
	})
}

func oneRune(name, s string, def rune) (rune, error) {
	if s == `\t` {
		s = "\t"
	}
	if s == "" {
		return def, nil
	}
	r := []rune(s)
	if len(r) != 1 || r[0] == '"' || r[0] == '\r' || r[0] == '\n' || r[0] == 0xFFFD {
		return 0, fmt.Errorf("%s must be a single character other than a quote or newline, got %q", name, s)
	}
	return r[0], nil
}

func detect(text string) rune {
	text = strings.TrimPrefix(text, "\uFEFF")
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) > 10 {
		lines = lines[:10]
	}
	best, bestScore := ',', 0
	for _, d := range []rune{',', ';', '\t', '|'} {
		var counts []int
		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			counts = append(counts, countOutsideQuotes(l, d))
		}
		if len(counts) == 0 || counts[0] == 0 {
			continue
		}
		same := 0
		for _, c := range counts {
			if c == counts[0] {
				same++
			}
		}
		score := same*1000 + counts[0]
		if same == len(counts) {
			score += 100000
		}
		if score > bestScore {
			best, bestScore = d, score
		}
	}
	return best
}

func countOutsideQuotes(l string, d rune) int {
	n, in := 0, false
	for _, r := range l {
		switch {
		case r == '"':
			in = !in
		case r == d && !in:
			n++
		}
	}
	return n
}

func parse(a tools.Args) (any, error) {
	text := strings.TrimPrefix(a.Str("text"), "\uFEFF")
	delim := detect(text)
	if s := a.Str("delimiter"); s != "" {
		var err error
		if delim, err = oneRune("delimiter", s, ','); err != nil {
			return nil, err
		}
	}
	cm, err := oneRune("comment", a.Str("comment"), 0)
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = delim
	r.Comment = cm
	r.FieldsPerRecord = -1
	r.LazyQuotes = !a.Bool("strict")
	trim := a.Bool("trim")
	r.TrimLeadingSpace = trim
	var rows [][]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid CSV: %v", err)
		}
		if len(rows) >= maxRecords {
			return nil, fmt.Errorf("more than %d rows", maxRecords)
		}
		if trim {
			for i := range rec {
				rec[i] = strings.TrimSpace(rec[i])
			}
		}
		rows = append(rows, rec)
	}
	out := make([]any, 0, len(rows))
	if !a.Bool("header") {
		for _, rec := range rows {
			row := make([]any, len(rec))
			for i, c := range rec {
				row[i] = c
			}
			out = append(out, row)
		}
		return out, nil
	}
	if len(rows) == 0 {
		return out, nil
	}
	head := rows[0]
	for _, rec := range rows[1:] {
		d := tools.NewDict()
		for i, c := range rec {
			k := fmt.Sprintf("col%d", i+1)
			if i < len(head) && head[i] != "" {
				k = head[i]
			}
			d.Set(k, c)
		}
		for i := len(rec); i < len(head); i++ {
			if head[i] != "" {
				d.Set(head[i], "")
			}
		}
		out = append(out, d)
	}
	return out, nil
}

func dump(a tools.Args) (any, error) {
	list, ok := a["rows"].([]any)
	if !ok {
		return nil, fmt.Errorf("rows must be a list of lists or a list of dicts")
	}
	delim, err := oneRune("delimiter", a.Str("delimiter"), ',')
	if err != nil {
		return nil, err
	}
	cols := a.List("columns")
	dicts := false
	for _, r := range list {
		if _, ok := r.(*tools.Dict); ok {
			dicts = true
		}
	}
	if dicts && len(cols) == 0 {
		seen := map[string]bool{}
		for _, r := range list {
			if d, ok := r.(*tools.Dict); ok {
				for _, k := range d.Keys() {
					if !seen[k] {
						seen[k] = true
						cols = append(cols, k)
					}
				}
			}
		}
	}
	var table [][]string
	if len(cols) > 0 && a.Bool("header") {
		table = append(table, cols)
	}
	for i, r := range list {
		var rec []string
		switch row := r.(type) {
		case *tools.Dict:
			for _, c := range cols {
				v, _ := row.Get(c)
				s, ok := tools.Scalar(v)
				if !ok {
					return nil, fmt.Errorf("row %d column %q: a list or dict cannot be a CSV cell", i, c)
				}
				rec = append(rec, s)
			}
		case []any:
			for j, v := range row {
				s, ok := tools.Scalar(v)
				if !ok {
					return nil, fmt.Errorf("row %d cell %d: a list or dict cannot be a CSV cell", i, j)
				}
				rec = append(rec, s)
			}
		default:
			return nil, fmt.Errorf("row %d must be a list or a dict", i)
		}
		table = append(table, rec)
	}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.Comma = delim
	w.UseCRLF = a.Bool("crlf")
	if err := w.WriteAll(table); err != nil {
		return nil, err
	}
	return b.String(), nil
}
