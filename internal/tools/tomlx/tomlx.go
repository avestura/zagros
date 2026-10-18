// Package tomlx registers the "toml" tool group: parse and write TOML. Pure:
// it works on strings only.
package tomlx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/avestura/zagros/internal/tools"
)

const maxDepth = 100

func init() {
	tools.Register(tools.Func{
		Group: "toml", Name: "parse", Title: "Parse TOML",
		Doc:     "Parse TOML text into a dict of dicts, lists, strings, ints, floats and bools. Keys keep their document order. Dates and times become RFC 3339 strings (local dates and times keep their own form). Raises with the line and column on invalid TOML or duplicate keys.",
		Terms:   []string{"toml to dict", "toml to json", "read toml", "load toml", "config file", "cargo.toml", "pyproject", "deserialize toml"},
		Params:  []tools.Param{tools.Req("text", "str", "The TOML text.")},
		Returns: "any",
		Example: `toml.parse("name = \"zagros\"\nports = [80, 443]\n[db]\nhost = \"localhost\"")`,
		Tags:    []string{"toml", "parse", "load", "config"},
		Fn:      parse,
	})
	tools.Register(tools.Func{
		Group: "toml", Name: "dump", Title: "Write TOML",
		Doc:     "Write a dict as TOML text: plain values first, then [tables] and [[arrays of tables]], keeping key order. None has no TOML form and raises, as does a top-level value that is not a dict, a mixed list of tables and values, or nesting deeper than 100 levels. Strings stay strings, so a parsed date is written back as a quoted string.",
		Terms:   []string{"dict to toml", "json to toml", "write toml", "save toml", "serialize toml"},
		Params:  []tools.Param{tools.Req("value", "any", "A dict to write.")},
		Returns: "str",
		Example: `toml.dump({"name": "zagros", "db": {"host": "localhost", "port": 5432}})`,
		Tags:    []string{"toml", "dump", "write", "serialize"},
		Fn: func(a tools.Args) (any, error) {
			d, ok := a["value"].(*tools.Dict)
			if !ok {
				return nil, fmt.Errorf("a TOML document must be a dict")
			}
			var b strings.Builder
			if err := writeTable(&b, nil, d, 0); err != nil {
				return nil, err
			}
			return b.String(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "toml", Name: "is_valid", Title: "Is valid TOML",
		Doc:     "Report whether text is valid TOML. Use toml.parse to get the error message.",
		Terms:   []string{"validate toml", "check toml", "lint toml"},
		Params:  []tools.Param{tools.Req("text", "str", "The TOML text.")},
		Returns: "bool",
		Example: `toml.is_valid("a = 1")`,
		Tags:    []string{"toml", "validate", "check"},
		Fn: func(a tools.Args) (any, error) {
			var m map[string]any
			_, err := toml.Decode(a.Str("text"), &m)
			return err == nil, nil
		},
	})
}

func parse(a tools.Args) (res any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("toml: cannot parse (%v)", r)
		}
	}()
	var m map[string]any
	md, derr := toml.Decode(strings.TrimPrefix(a.Str("text"), "\uFEFF"), &m)
	if derr != nil {
		return nil, fmt.Errorf("toml: %s", cleanErr(derr))
	}
	order := map[string]int{}
	for i, k := range md.Keys() {
		p := strings.Join(k, "\x00")
		if _, seen := order[p]; !seen {
			order[p] = i
		}
	}
	return convert(m, nil, order, 0)
}

func cleanErr(err error) string {
	if pe, ok := err.(toml.ParseError); ok {
		msg := pe.Message
		if msg == "" {
			msg = pe.Error()
		}
		return fmt.Sprintf("line %d, column %d: %s", pe.Position.Line, pe.Position.Col, msg)
	}
	return strings.TrimPrefix(err.Error(), "toml: ")
}

func convert(v any, path []string, order map[string]int, depth int) (any, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("toml: nesting is deeper than %d levels", maxDepth)
	}
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		idx := func(k string) int {
			if i, ok := order[strings.Join(append(append([]string(nil), path...), k), "\x00")]; ok {
				return i
			}
			return 1 << 30
		}
		sortBy(keys, idx)
		d := tools.NewDict()
		for _, k := range keys {
			e, err := convert(x[k], append(append([]string(nil), path...), k), order, depth+1)
			if err != nil {
				return nil, err
			}
			d.Set(k, e)
		}
		return d, nil
	case []map[string]any:
		out := make([]any, len(x))
		for i, e := range x {
			c, err := convert(e, path, order, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			c, err := convert(e, path, order, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	case time.Time:
		return formatTime(x), nil
	case int64, float64, string, bool:
		return x, nil
	case int:
		return int64(x), nil
	}
	return fmt.Sprint(v), nil
}

func sortBy(keys []string, idx func(string) int) {
	// insertion sort by document position, then name; key counts are small
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0; j-- {
			a, b := keys[j-1], keys[j]
			ia, ib := idx(a), idx(b)
			if ia < ib || (ia == ib && a <= b) {
				break
			}
			keys[j-1], keys[j] = b, a
		}
	}
}

func formatTime(t time.Time) string {
	switch t.Location().String() {
	case "date-local":
		return t.Format("2006-01-02")
	case "time-local":
		return t.Format("15:04:05.999999999")
	case "datetime-local":
		return t.Format("2006-01-02T15:04:05.999999999")
	}
	return t.Format(time.RFC3339Nano)
}

// ---- writer ----

func bareKey(k string) bool {
	if k == "" {
		return false
	}
	for _, r := range k {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\u%04x`, r)
		case r == utf8.RuneError:
			b.WriteString(`�`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func keyText(k string) string {
	if bareKey(k) {
		return k
	}
	return quote(k)
}

func dotted(path []string) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = keyText(p)
	}
	return strings.Join(parts, ".")
}

func isTableList(v any) bool {
	l, ok := v.([]any)
	if !ok || len(l) == 0 {
		return false
	}
	for _, e := range l {
		if _, ok := e.(*tools.Dict); !ok {
			return false
		}
	}
	return true
}

func inline(v any, depth int) (string, error) {
	if depth > maxDepth {
		return "", fmt.Errorf("nesting is deeper than %d levels", maxDepth)
	}
	switch x := v.(type) {
	case nil:
		return "", fmt.Errorf("None cannot be written as TOML; leave the key out")
	case string:
		return quote(x), nil
	case bool:
		return strconv.FormatBool(x), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case int:
		return strconv.Itoa(x), nil
	case float64:
		switch {
		case math.IsNaN(x):
			return "nan", nil
		case math.IsInf(x, 1):
			return "inf", nil
		case math.IsInf(x, -1):
			return "-inf", nil
		}
		s := strconv.FormatFloat(x, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s, nil
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			s, err := inline(e, depth+1)
			if err != nil {
				return "", err
			}
			parts[i] = s
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	case *tools.Dict:
		parts := make([]string, 0, x.Len())
		for _, k := range x.Keys() {
			e, _ := x.Get(k)
			s, err := inline(e, depth+1)
			if err != nil {
				return "", fmt.Errorf("key %q: %w", k, err)
			}
			parts = append(parts, keyText(k)+" = "+s)
		}
		if len(parts) == 0 {
			return "{}", nil
		}
		return "{ " + strings.Join(parts, ", ") + " }", nil
	}
	return "", fmt.Errorf("cannot write a value of type %T as TOML", v)
}

func writeTable(b *strings.Builder, path []string, d *tools.Dict, depth int) error {
	if depth > maxDepth {
		return fmt.Errorf("nesting is deeper than %d levels", maxDepth)
	}
	var tables, arrays []string
	for _, k := range d.Keys() {
		v, _ := d.Get(k)
		switch {
		case isDict(v):
			tables = append(tables, k)
		case isTableList(v):
			arrays = append(arrays, k)
		default:
			s, err := inline(v, 0)
			if err != nil {
				return fmt.Errorf("key %q: %w", k, err)
			}
			b.WriteString(keyText(k) + " = " + s + "\n")
		}
	}
	for _, k := range tables {
		v, _ := d.Get(k)
		sub := append(append([]string(nil), path...), k)
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[" + dotted(sub) + "]\n")
		if err := writeTable(b, sub, v.(*tools.Dict), depth+1); err != nil {
			return err
		}
	}
	for _, k := range arrays {
		v, _ := d.Get(k)
		sub := append(append([]string(nil), path...), k)
		for _, e := range v.([]any) {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString("[[" + dotted(sub) + "]]\n")
			if err := writeTable(b, sub, e.(*tools.Dict), depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func isDict(v any) bool { _, ok := v.(*tools.Dict); return ok }
