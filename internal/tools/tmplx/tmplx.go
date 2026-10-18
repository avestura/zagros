// Package tmplx registers the "template" tool group: render text from a safe
// subset of Go's text/template. Templates cannot call anything but the
// functions listed here, cannot define or include other templates, and run
// under output and time limits, so rendering untrusted templates is safe.
package tmplx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"text/template/parse"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxTemplate = 256 << 10
	maxOutput   = 4 << 20
	maxRange    = 100000
	timeout     = 3 * time.Second
)

func init() {
	tools.Register(tools.Func{
		Group: "template", Name: "render", Title: "Render a text template",
		Doc:   "Render text from a template and a dict using Go template syntax: {{ .name }}, {{ if .ok }}...{{ else }}...{{ end }}, {{ range .items }}{{ . }}{{ end }}, {{ with .x }}, variables ($x := ...), pipelines ({{ .name | upper }}) and comments. Besides the Go built-ins (len, index, printf, eq, lt, and, or, not, slice, html, js, urlquery) these functions exist: upper, lower, title, trim, trim_prefix, trim_suffix, replace, split, join, contains, has_prefix, has_suffix, repeat, default, coalesce, quote, json, add, sub, mul, div, mod, int, float, string, list, dict, keys. Nothing can reach outside the sandbox, and {{ define }}, {{ template }}, {{ block }} and call are rejected. A missing key renders as empty text, or raises with strict=True. Raises with the position on syntax or runtime errors, and when the output passes 4 MB or rendering takes over 3 seconds.",
		Terms: []string{"mustache", "handlebars", "jinja", "fill template", "string interpolation", "mail merge", "placeholders", "go template", "text template", "format string with dict", "generate text from data"},
		Params: []tools.Param{
			tools.Req("template", "str", "The template text."),
			tools.Opt("data", "any", "The values, usually a dict; refer to keys as .key.", nil),
			tools.Opt("left", "str", "Opening delimiter.", "{{"),
			tools.Opt("right", "str", "Closing delimiter.", "}}"),
			tools.Opt("strict", "bool", "Raise on a missing key instead of rendering it as empty.", false),
		},
		Returns: "str",
		Example: `template.render("Hello {{ .name }}! {{ range .tags }}[{{ . | upper }}]{{ end }}", {"name": "Ada", "tags": ["a", "b"]})`,
		Tags:    []string{"template", "render", "format", "interpolate", "placeholder", "mustache"},
		Fn:      render,
	})
	tools.Register(tools.Func{
		Group: "template", Name: "is_valid", Title: "Is valid template",
		Doc:   "Report whether a template parses and uses only the allowed features (see template.render). Use template.render to get the error message.",
		Terms: []string{"validate template", "check template syntax", "lint template"},
		Params: []tools.Param{
			tools.Req("template", "str", "The template text."),
			tools.Opt("left", "str", "Opening delimiter.", "{{"),
			tools.Opt("right", "str", "Closing delimiter.", "}}"),
		},
		Returns: "bool",
		Example: `template.is_valid("{{ if .a }}x{{ end }}")`,
		Tags:    []string{"template", "validate", "check", "syntax"},
		Fn: func(a tools.Args) (any, error) {
			_, err := compile(a.Str("template"), a.Str("left"), a.Str("right"), false)
			return err == nil, nil
		},
	})
}

func compile(text, left, right string, strict bool) (*template.Template, error) {
	if len(text) > maxTemplate {
		return nil, fmt.Errorf("template is larger than %d KB", maxTemplate>>10)
	}
	t := template.New("t").Funcs(funcs())
	if strict {
		t = t.Option("missingkey=error")
	}
	if left != "" || right != "" {
		if left == "" || right == "" {
			return nil, fmt.Errorf("set both left and right delimiters, or neither")
		}
		t = t.Delims(left, right)
	}
	t, err := t.Parse(text)
	if err != nil {
		return nil, fmt.Errorf("%s", trimErr(err))
	}
	for _, tr := range t.Templates() {
		if tr.Tree == nil {
			continue
		}
		if tr.Name() != "t" {
			return nil, fmt.Errorf("{{ define }} and {{ block }} are not allowed")
		}
		if err := check(tr.Tree.Root); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func trimErr(err error) string {
	return strings.TrimPrefix(err.Error(), "template: ")
}

// check rejects template calls, call, and very large range counts.
func check(n parse.Node) error {
	switch x := n.(type) {
	case nil:
		return nil
	case *parse.ListNode:
		if x == nil {
			return nil
		}
		for _, c := range x.Nodes {
			if err := check(c); err != nil {
				return err
			}
		}
	case *parse.TemplateNode:
		return fmt.Errorf("{{ template }} is not allowed")
	case *parse.ActionNode:
		return checkPipe(x.Pipe)
	case *parse.IfNode:
		return checkBranch(&x.BranchNode)
	case *parse.WithNode:
		return checkBranch(&x.BranchNode)
	case *parse.RangeNode:
		return checkBranch(&x.BranchNode)
	}
	return nil
}

func checkBranch(b *parse.BranchNode) error {
	if err := checkPipe(b.Pipe); err != nil {
		return err
	}
	if err := check(b.List); err != nil {
		return err
	}
	if b.ElseList != nil {
		return check(b.ElseList)
	}
	return nil
}

func checkPipe(p *parse.PipeNode) error {
	if p == nil {
		return nil
	}
	for _, c := range p.Cmds {
		for _, a := range c.Args {
			if err := checkArg(a); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkArg(a parse.Node) error {
	switch x := a.(type) {
	case *parse.IdentifierNode:
		if x.Ident == "call" {
			return fmt.Errorf("call is not allowed")
		}
	case *parse.NumberNode:
		if x.IsInt && (x.Int64 > maxRange || x.Int64 < -maxRange) {
			return fmt.Errorf("numbers above %d are not allowed in templates", maxRange)
		}
	case *parse.PipeNode:
		return checkPipe(x)
	case *parse.ChainNode:
		return checkArg(x.Node)
	}
	return nil
}

type limitWriter struct {
	b bytes.Buffer
}

var errTooLong = fmt.Errorf("output is larger than %d MB", maxOutput>>20)

func (w *limitWriter) Write(p []byte) (int, error) {
	if w.b.Len()+len(p) > maxOutput {
		return 0, errTooLong
	}
	return w.b.Write(p)
}

func render(a tools.Args) (res any, err error) {
	t, err := compile(a.Str("template"), a.Str("left"), a.Str("right"), a.Bool("strict"))
	if err != nil {
		return nil, err
	}
	data := tools.Plain(a["data"])
	done := make(chan struct{})
	var out string
	var rerr error
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				rerr = fmt.Errorf("template failed: %v", r)
			}
		}()
		var w limitWriter
		if err := t.Execute(&w, data); err != nil {
			if strings.Contains(err.Error(), errTooLong.Error()) {
				rerr = errTooLong
			} else {
				rerr = fmt.Errorf("%s", trimErr(err))
			}
			return
		}
		out = w.b.String()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		return nil, fmt.Errorf("rendering took longer than %s", timeout)
	}
	if rerr != nil {
		return nil, rerr
	}
	if !a.Bool("strict") {
		out = strings.ReplaceAll(out, "<no value>", "")
	}
	return out, nil
}

// ---- functions ----

func num(v any) (float64, bool, error) {
	switch x := v.(type) {
	case int64:
		return float64(x), true, nil
	case int:
		return float64(x), true, nil
	case float64:
		return x, false, nil
	case string:
		if i, err := strconv.ParseInt(x, 10, 64); err == nil {
			return float64(i), true, nil
		}
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f, false, nil
		}
	}
	return 0, false, fmt.Errorf("not a number: %v", v)
}

func arith(op func(a, b float64) (float64, error)) func(a, b any) (any, error) {
	return func(a, b any) (any, error) {
		x, xi, err := num(a)
		if err != nil {
			return nil, err
		}
		y, yi, err := num(b)
		if err != nil {
			return nil, err
		}
		r, err := op(x, y)
		if err != nil {
			return nil, err
		}
		if xi && yi && r == math.Trunc(r) && math.Abs(r) < 1<<53 {
			return int64(r), nil
		}
		return r, nil
	}
}

func str(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := tools.Scalar(v); ok {
		return s
	}
	return fmt.Sprint(v)
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"upper":       func(s any) string { return strings.ToUpper(str(s)) },
		"lower":       func(s any) string { return strings.ToLower(str(s)) },
		"title":       func(s any) string { return titleCase(str(s)) },
		"trim":        func(s any) string { return strings.TrimSpace(str(s)) },
		"trim_prefix": func(p string, s any) string { return strings.TrimPrefix(str(s), p) },
		"trim_suffix": func(p string, s any) string { return strings.TrimSuffix(str(s), p) },
		"replace":     func(old, new string, s any) string { return strings.ReplaceAll(str(s), old, new) },
		"split":       func(sep string, s any) []string { return strings.Split(str(s), sep) },
		"join": func(sep string, v any) (string, error) {
			switch x := v.(type) {
			case []string:
				return strings.Join(x, sep), nil
			case []any:
				parts := make([]string, len(x))
				for i, e := range x {
					parts[i] = str(e)
				}
				return strings.Join(parts, sep), nil
			}
			return "", fmt.Errorf("join needs a list")
		},
		"contains":   func(sub string, s any) bool { return strings.Contains(str(s), sub) },
		"has_prefix": func(p string, s any) bool { return strings.HasPrefix(str(s), p) },
		"has_suffix": func(p string, s any) bool { return strings.HasSuffix(str(s), p) },
		"repeat": func(n int, s any) (string, error) {
			if n < 0 || n > 10000 || len(str(s))*n > maxOutput {
				return "", fmt.Errorf("repeat count %d is out of range", n)
			}
			return strings.Repeat(str(s), n), nil
		},
		"default": func(def, v any) any {
			if v == nil || v == "" || v == false || v == int64(0) || v == float64(0) {
				return def
			}
			if l, ok := v.([]any); ok && len(l) == 0 {
				return def
			}
			if m, ok := v.(map[string]any); ok && len(m) == 0 {
				return def
			}
			return v
		},
		"coalesce": func(vs ...any) any {
			for _, v := range vs {
				if v != nil && v != "" {
					return v
				}
			}
			return nil
		},
		"quote": func(s any) string { return strconv.Quote(str(s)) },
		"json": func(v any) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
		},
		"add": arith(func(a, b float64) (float64, error) { return a + b, nil }),
		"sub": arith(func(a, b float64) (float64, error) { return a - b, nil }),
		"mul": arith(func(a, b float64) (float64, error) { return a * b, nil }),
		"div": arith(func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return a / b, nil
		}),
		"mod": arith(func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return math.Mod(a, b), nil
		}),
		"int": func(v any) (int64, error) {
			f, _, err := num(v)
			return int64(f), err
		},
		"float": func(v any) (float64, error) {
			f, _, err := num(v)
			return f, err
		},
		"string": func(v any) string { return str(v) },
		"list":   func(vs ...any) []any { return vs },
		"dict": func(kv ...any) (map[string]any, error) {
			if len(kv)%2 != 0 {
				return nil, fmt.Errorf("dict needs key/value pairs")
			}
			m := map[string]any{}
			for i := 0; i < len(kv); i += 2 {
				k, ok := kv[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[k] = kv[i+1]
			}
			return m, nil
		},
		"keys": func(v any) ([]string, error) {
			m, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("keys needs a dict")
			}
			out := make([]string, 0, len(m))
			for k := range m {
				out = append(out, k)
			}
			sort.Strings(out)
			return out, nil
		},
	}
}

func titleCase(s string) string {
	var b strings.Builder
	up := true
	for _, r := range s {
		if up && r >= 'a' && r <= 'z' {
			r -= 32
		}
		up = r == ' ' || r == '\t' || r == '\n' || r == '-' || r == '_'
		b.WriteRune(r)
	}
	return b.String()
}
