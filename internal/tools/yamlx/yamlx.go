// Package yamlx registers the "yaml" tool group: parse and write YAML. Pure:
// it works on strings only.
package yamlx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

func init() {
	tools.Register(tools.Func{
		Group: "yaml", Name: "parse", Title: "Parse YAML",
		Doc:     "Parse the first YAML document in text into dicts, lists, strings, ints, floats, bools and None. Supports block and flow collections, quoted and block scalars, comments, anchors with aliases, merge keys and the core scalar types (true/false are bools; yes/no stay strings). Empty text gives None. Raises with a line number on malformed YAML or a duplicate key.",
		Terms:   []string{"yaml to dict", "yaml to json", "read yaml", "load yaml", "config file", "yml", "deserialize yaml"},
		Params:  []tools.Param{tools.Req("text", "str", "The YAML text.")},
		Returns: "any",
		Example: `yaml.parse("name: zagros\ntags: [a, b]\nport: 8080")`,
		Tags:    []string{"yaml", "yml", "parse", "load", "config"},
		Fn: func(a tools.Args) (any, error) {
			docs, err := splitDocs(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return parseDocument(docs[0])
		},
	})
	tools.Register(tools.Func{
		Group: "yaml", Name: "parse_all", Title: "Parse all YAML documents",
		Doc:     "Parse every document of a multi-document YAML stream (separated by ---) and return them as a list. Raises with a line number on malformed YAML.",
		Terms:   []string{"multi document yaml", "yaml stream", "kubernetes manifests", "split yaml documents"},
		Params:  []tools.Param{tools.Req("text", "str", "The YAML text.")},
		Returns: "any",
		Example: `yaml.parse_all("a: 1\n---\nb: 2")`,
		Tags:    []string{"yaml", "documents", "stream", "parse"},
		Fn: func(a tools.Args) (any, error) {
			docs, err := splitDocs(a.Str("text"))
			if err != nil {
				return nil, err
			}
			out := []any{}
			for _, d := range docs {
				v, err := parseDocument(d)
				if err != nil {
					return nil, err
				}
				out = append(out, v)
			}
			return out, nil
		},
	})
	tools.Register(tools.Func{
		Group: "yaml", Name: "dump", Title: "Write YAML",
		Doc:   "Write a value (dict, list, string, number, bool, None) as YAML text with the given indentation. Dict keys keep their order. Strings that would read back as another type, or that contain special characters, are quoted; multi-line strings use the literal | style. Raises when the value is nested deeper than 100 levels.",
		Terms: []string{"dict to yaml", "json to yaml", "write yaml", "save yaml", "serialize yaml", "yml output"},
		Params: []tools.Param{
			tools.Req("value", "any", "The value to write."),
			tools.Opt("indent", "int", "Spaces per level, 1 to 8.", int64(2)),
		},
		Returns: "str",
		Example: `yaml.dump({"name": "zagros", "tags": ["a", "b"]})`,
		Tags:    []string{"yaml", "yml", "dump", "write", "serialize"},
		Fn: func(a tools.Args) (any, error) {
			n := int(a.Int("indent"))
			if n < 1 || n > 8 {
				return nil, fmt.Errorf("indent must be between 1 and 8, got %d", n)
			}
			w := &writer{step: n}
			if err := w.block(a["value"], 0, 0); err != nil {
				return nil, err
			}
			return w.b.String(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "yaml", Name: "is_valid", Title: "Is valid YAML",
		Doc:     "Report whether text parses as YAML (every document). Use yaml.parse to get the error message.",
		Terms:   []string{"validate yaml", "check yaml", "lint yaml"},
		Params:  []tools.Param{tools.Req("text", "str", "The YAML text.")},
		Returns: "bool",
		Example: `yaml.is_valid("a: [1, 2")`,
		Tags:    []string{"yaml", "validate", "check"},
		Fn: func(a tools.Args) (any, error) {
			docs, err := splitDocs(a.Str("text"))
			if err != nil {
				return false, nil
			}
			for _, d := range docs {
				if _, err := parseDocument(d); err != nil {
					return false, nil
				}
			}
			return true, nil
		},
	})
}

type writer struct {
	b    strings.Builder
	step int
}

// block writes a collection (or a lone scalar) with every line indented by
// at least col spaces.
func (w *writer) block(v any, col, depth int) error {
	if depth > maxDepth {
		return fmt.Errorf("value is nested deeper than %d levels", maxDepth)
	}
	pad := strings.Repeat(" ", col)
	switch x := v.(type) {
	case *tools.Dict:
		for _, k := range x.Keys() {
			e, _ := x.Get(k)
			w.b.WriteString(pad + quoteString(k) + ":")
			if err := w.after(e, col, depth, false); err != nil {
				return err
			}
		}
	case []any:
		for _, e := range x {
			w.b.WriteString(pad + "-")
			if err := w.after(e, col, depth, true); err != nil {
				return err
			}
		}
	case []string:
		items := make([]any, len(x))
		for i, s := range x {
			items[i] = s
		}
		return w.block(items, col, depth)
	default:
		w.b.WriteString(pad)
		return w.scalar(v, col)
	}
	return nil
}

// after writes what follows "key:" or "-" on a line at column col.
func (w *writer) after(e any, col, depth int, seq bool) error {
	switch y := e.(type) {
	case *tools.Dict:
		if y.Len() > 0 {
			return w.nested(y, col, depth, seq)
		}
		w.b.WriteString(" {}\n")
		return nil
	case []any:
		if len(y) > 0 {
			return w.nested(y, col, depth, seq)
		}
		w.b.WriteString(" []\n")
		return nil
	case []string:
		items := make([]any, len(y))
		for i, s := range y {
			items[i] = s
		}
		return w.after(items, col, depth, seq)
	}
	w.b.WriteString(" ")
	return w.scalar(e, col+w.step)
}

func (w *writer) nested(v any, col, depth int, seq bool) error {
	if seq {
		// sequence entry: the first line of the child shares the dash line
		var sub writer
		sub.step = w.step
		if err := sub.block(v, col+2, depth+1); err != nil {
			return err
		}
		w.b.WriteString(" " + strings.TrimLeft(sub.b.String(), " "))
		return nil
	}
	w.b.WriteString("\n")
	return w.block(v, col+w.step, depth+1)
}

func (w *writer) scalar(v any, col int) error {
	switch x := v.(type) {
	case nil:
		w.b.WriteString("null\n")
	case bool:
		w.b.WriteString(strconv.FormatBool(x) + "\n")
	case int64:
		w.b.WriteString(strconv.FormatInt(x, 10) + "\n")
	case int:
		w.b.WriteString(strconv.Itoa(x) + "\n")
	case float64:
		switch {
		case math.IsNaN(x):
			w.b.WriteString(".nan\n")
		case math.IsInf(x, 1):
			w.b.WriteString(".inf\n")
		case math.IsInf(x, -1):
			w.b.WriteString("-.inf\n")
		default:
			s := strconv.FormatFloat(x, 'g', -1, 64)
			if !strings.ContainsAny(s, ".eE") {
				s += ".0"
			}
			w.b.WriteString(s + "\n")
		}
	case string:
		w.str(x, col)
	default:
		return fmt.Errorf("cannot write a value of type %T as YAML", v)
	}
	return nil
}

func (w *writer) str(s string, col int) {
	if strings.Contains(s, "\n") && printable(s) && !strings.HasPrefix(s, " ") && !strings.HasPrefix(s, "\n") && !strings.Contains(s, "\n\n\n") && !strings.HasSuffix(s, "\n\n") {
		hdr := "|"
		body := s
		if strings.HasSuffix(body, "\n") {
			body = body[:len(body)-1]
		} else {
			hdr += "-"
		}
		w.b.WriteString(hdr + "\n")
		pad := strings.Repeat(" ", col)
		for _, l := range strings.Split(body, "\n") {
			if l != "" {
				w.b.WriteString(pad + l)
			}
			w.b.WriteString("\n")
		}
		return
	}
	w.b.WriteString(quoteString(s) + "\n")
}

func printable(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f || r == utf8.RuneError || (r >= 0x80 && r < 0xa0) || r == 0x2028 || r == 0x2029 || r == 0xfeff {
			return false
		}
	}
	return true
}

func needsQuote(s string) bool {
	if s == "" || s != strings.TrimSpace(s) {
		return true
	}
	if !printable(s) || strings.ContainsAny(s, "\n\t") {
		return true
	}
	if _, isStr := resolvePlain(s, "").(string); !isStr {
		return true
	}
	switch s[0] {
	case '-', '?', ':', ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
		if len(s) == 1 || s[0] != '-' && s[0] != '?' && s[0] != ':' || s[1] == ' ' {
			return true
		}
	}
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":") {
		return true
	}
	switch strings.ToLower(s) {
	case "yes", "no", "on", "off", "y", "n", "---", "...":
		return true
	}
	return false
}

func quoteString(s string) string {
	if !needsQuote(s) {
		return s
	}
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			switch {
			case r < 0x20 || r == 0x7f:
				fmt.Fprintf(&b, `\x%02x`, r)
			case r == utf8.RuneError || (r >= 0x80 && r < 0xa0) || r == 0x2028 || r == 0x2029 || r == 0xfeff:
				fmt.Fprintf(&b, `\u%04x`, r)
			default:
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
