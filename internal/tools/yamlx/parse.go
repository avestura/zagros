package yamlx

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

// The parser reads the YAML that configuration files use: block and flow
// collections, plain/quoted/block scalars, comments, anchors with aliases and
// merge keys, tags for the core scalar types, and multiple documents. It
// does not support explicit "? key" entries, complex keys or custom tags.

const (
	maxDepth = 100
	maxLines = 1_000_000
)

var errEOF = errors.New("unexpected end of input")

type line struct {
	indent int
	text   string // content after the indentation (comments are stripped on demand)
	num    int
}

type parser struct {
	lines   []line
	pos     int
	anchors map[string]any
	depth   int
}

type yamlError struct {
	line int
	msg  string
}

func (e *yamlError) Error() string {
	if e.line > 0 {
		return fmt.Sprintf("yaml: line %d: %s", e.line, e.msg)
	}
	return "yaml: " + e.msg
}

func (p *parser) fail(format string, a ...any) error {
	n := 0
	if p.pos < len(p.lines) {
		n = p.lines[p.pos].num
	} else if len(p.lines) > 0 {
		n = p.lines[len(p.lines)-1].num
	}
	return &yamlError{n, fmt.Sprintf(format, a...)}
}

// splitDocs splits text into documents of lines.
func splitDocs(text string) ([][]line, error) {
	text = strings.TrimPrefix(text, "\uFEFF")
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	raw := strings.Split(text, "\n")
	if len(raw) > maxLines {
		return nil, &yamlError{0, "too many lines"}
	}
	var docs [][]line
	var cur []line
	started := false
	flush := func() {
		if started || len(cur) > 0 {
			docs = append(docs, cur)
		}
		cur = nil
	}
	for i, r := range raw {
		if strings.HasPrefix(r, "%") && !started && len(cur) == 0 {
			continue // directive
		}
		if r == "---" || strings.HasPrefix(r, "--- ") || strings.HasPrefix(r, "---\t") {
			if started || hasContent(cur) {
				flush()
			}
			started = true
			if rest := strings.TrimSpace(r[3:]); rest != "" && !strings.HasPrefix(rest, "#") {
				cur = append(cur, line{0, rest, i + 1})
			}
			continue
		}
		if r == "..." || strings.HasPrefix(r, "... ") {
			flush()
			started = false
			continue
		}
		if strings.ContainsRune(r, '\t') {
			t := strings.TrimLeft(r, " ")
			if strings.HasPrefix(t, "\t") && strings.TrimSpace(t) != "" && !strings.HasPrefix(strings.TrimSpace(t), "#") {
				return nil, &yamlError{i + 1, "tabs cannot be used for indentation"}
			}
		}
		t := strings.TrimLeft(r, " ")
		cur = append(cur, line{len(r) - len(t), strings.TrimRight(t, " \t"), i + 1})
	}
	if started || hasContent(cur) {
		flush()
	}
	if len(docs) == 0 {
		docs = append(docs, nil)
	}
	return docs, nil
}

func hasContent(ls []line) bool {
	for _, l := range ls {
		if t := l.text; t != "" && !strings.HasPrefix(t, "#") {
			return true
		}
	}
	return false
}

func parseDocument(lines []line) (any, error) {
	p := &parser{lines: lines, anchors: map[string]any{}}
	p.skipBlank()
	if p.pos >= len(p.lines) {
		return nil, nil
	}
	v, err := p.node(-1)
	if err != nil {
		return nil, err
	}
	p.skipBlank()
	if p.pos < len(p.lines) {
		return nil, p.fail("unexpected content %q", clip(p.lines[p.pos].text))
	}
	return v, nil
}

func clip(s string) string {
	if utf8.RuneCountInString(s) > 40 {
		r := []rune(s)
		return string(r[:40]) + "..."
	}
	return s
}

func (p *parser) skipBlank() {
	for p.pos < len(p.lines) {
		t := p.lines[p.pos].text
		if t == "" || strings.HasPrefix(t, "#") {
			p.pos++
			continue
		}
		return
	}
}

func isSeqItem(t string) bool { return t == "-" || strings.HasPrefix(t, "- ") }

// node parses the value that starts at the current line, which must be
// indented deeper than parent.
func (p *parser) node(parent int) (any, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxDepth {
		return nil, p.fail("nesting is deeper than %d levels", maxDepth)
	}
	p.skipBlank()
	if p.pos >= len(p.lines) || p.lines[p.pos].indent <= parent {
		return nil, nil
	}
	l := p.lines[p.pos]
	if strings.HasPrefix(l.text, "? ") || l.text == "?" {
		return nil, p.fail("explicit keys (?) are not supported")
	}
	if isSeqItem(l.text) {
		return p.seq(l.indent)
	}
	if _, _, ok, err := splitKey(stripComment(l.text)); err != nil {
		return nil, p.fail("%v", err)
	} else if ok {
		return p.mapping(l.indent)
	}
	return p.inline(stripComment(l.text), parent)
}

func (p *parser) seq(indent int) (any, error) {
	out := []any{}
	for {
		p.skipBlank()
		if p.pos >= len(p.lines) {
			break
		}
		l := p.lines[p.pos]
		if l.indent != indent || !isSeqItem(l.text) {
			if l.indent > indent {
				return nil, p.fail("bad indentation of a sequence entry")
			}
			break
		}
		rest := strings.TrimLeft(l.text[1:], " ")
		offset := len(l.text) - len(rest)
		code := stripComment(rest)
		switch {
		case code == "":
			p.pos++
			v, err := p.node(indent)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		case isSeqItem(code), keyLike(code):
			p.lines[p.pos] = line{indent + offset, rest, l.num}
			v, err := p.node(indent)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		default:
			v, err := p.inline(code, indent)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
	}
	return out, nil
}

func keyLike(code string) bool {
	_, _, ok, err := splitKey(code)
	return ok && err == nil
}

func (p *parser) mapping(indent int) (any, error) {
	d := tools.NewDict()
	explicit := map[string]bool{}
	for {
		p.skipBlank()
		if p.pos >= len(p.lines) {
			break
		}
		l := p.lines[p.pos]
		if strings.HasPrefix(l.text, "? ") || l.text == "?" {
			return nil, p.fail("explicit keys (?) are not supported")
		}
		if l.indent != indent {
			if l.indent > indent {
				return nil, p.fail("bad indentation of a mapping entry")
			}
			break
		}
		key, rest, ok, err := splitKey(stripComment(l.text))
		if err != nil {
			return nil, p.fail("%v", err)
		}
		if !ok {
			if isSeqItem(l.text) {
				return nil, p.fail("a sequence entry is not allowed here")
			}
			return nil, p.fail("expected \"key: value\", got %q", clip(l.text))
		}
		var val any
		meta, rest2 := props(rest)
		if rest2 == "" {
			p.pos++
			p.skipBlank()
			switch {
			case p.pos < len(p.lines) && p.lines[p.pos].indent > indent:
				val, err = p.node(indent)
			case p.pos < len(p.lines) && p.lines[p.pos].indent == indent && isSeqItem(p.lines[p.pos].text):
				val, err = p.seq(indent)
			}
			if err != nil {
				return nil, err
			}
			if val == nil {
				val = resolveTag(meta.tag, "", true)
			}
			val = p.applyMeta(meta, val)
		} else {
			val, err = p.inline(rest, indent)
			if err != nil {
				return nil, err
			}
		}
		if key == "<<" {
			if err := mergeInto(d, val); err != nil {
				return nil, p.fail("%v", err)
			}
			continue
		}
		if explicit[key] {
			return nil, &yamlError{l.num, fmt.Sprintf("duplicate key %q", key)}
		}
		explicit[key] = true
		d.Set(key, val)
	}
	return d, nil
}

func mergeInto(d *tools.Dict, v any) error {
	switch x := v.(type) {
	case *tools.Dict:
		for _, k := range x.Keys() {
			if _, has := d.Get(k); !has {
				e, _ := x.Get(k)
				d.Set(k, e)
			}
		}
	case []any:
		for _, e := range x {
			if err := mergeInto(d, e); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("a merge key (<<) needs a mapping or a list of mappings")
	}
	return nil
}

type meta struct {
	anchor string
	tag    string
}

// props strips leading "&anchor" and "!tag" properties from a value.
func props(s string) (meta, string) {
	var m meta
	for {
		s = strings.TrimLeft(s, " ")
		if strings.HasPrefix(s, "&") {
			i := strings.IndexAny(s, " ")
			if i < 0 {
				i = len(s)
			}
			m.anchor = s[1:i]
			s = s[i:]
			continue
		}
		if strings.HasPrefix(s, "!") {
			i := strings.IndexAny(s, " ")
			if i < 0 {
				i = len(s)
			}
			m.tag = s[:i]
			s = s[i:]
			continue
		}
		return m, s
	}
}

func (p *parser) applyMeta(m meta, v any) any {
	if m.anchor != "" {
		p.anchors[m.anchor] = v
	}
	return v
}

// inline parses a value that begins on the current line (text is the part
// after "key:" or "- ") and advances past every line it uses.
func (p *parser) inline(text string, parent int) (any, error) {
	m, s := props(text)
	if s == "" {
		// properties only; the value is on the following lines
		p.pos++
		v, err := p.node(parent)
		if err != nil {
			return nil, err
		}
		return p.applyMeta(m, v), nil
	}
	if strings.HasPrefix(s, "*") {
		name := strings.TrimSpace(s[1:])
		v, ok := p.anchors[name]
		if !ok {
			return nil, p.fail("unknown alias *%s", name)
		}
		p.pos++
		return v, nil
	}
	if s[0] == '|' || s[0] == '>' {
		v, err := p.block(s, parent)
		if err != nil {
			return nil, err
		}
		return p.applyMeta(m, v), nil
	}
	var v any
	switch s[0] {
	case '[', '{', '"', '\'':
		first := p.pos
		whole := s
		for {
			val, n, err := parseFlowValue(whole, 0)
			if err == nil {
				if tail := stripComment(whole[n:]); strings.TrimSpace(tail) != "" {
					return nil, &yamlError{p.lines[first].num, fmt.Sprintf("unexpected %q after value", clip(strings.TrimSpace(tail)))}
				}
				v = val
				p.pos++
				break
			}
			if err != errEOF {
				return nil, &yamlError{p.lines[first].num, err.Error()}
			}
			p.pos++
			if p.pos >= len(p.lines) {
				return nil, &yamlError{p.lines[first].num, "unterminated quote or bracket"}
			}
			next := p.lines[p.pos]
			whole += "\n" + strings.Repeat(" ", next.indent) + next.text
		}
		if m.tag != "" {
			if str, ok := v.(string); ok {
				v = resolveTag(m.tag, str, false)
			}
		}
	default:
		// plain scalar, possibly continued on deeper lines
		parts := []string{s}
		p.pos++
		for p.pos < len(p.lines) {
			n := p.lines[p.pos]
			if n.text == "" {
				// a blank line inside a multi-line plain scalar
				j := p.pos
				for j < len(p.lines) && p.lines[j].text == "" {
					j++
				}
				if j < len(p.lines) && p.lines[j].indent > parent && !strings.HasPrefix(p.lines[j].text, "#") && !isStructural(p.lines[j].text) {
					parts = append(parts, "\n")
					p.pos = j
					continue
				}
				break
			}
			if n.indent <= parent || strings.HasPrefix(n.text, "#") || isStructural(n.text) {
				break
			}
			parts = append(parts, stripComment(n.text))
			p.pos++
		}
		v = resolvePlain(foldPlain(parts), m.tag)
	}
	return p.applyMeta(m, v), nil
}

func isStructural(t string) bool {
	if isSeqItem(t) {
		return true
	}
	return keyLike(stripComment(t))
}

func foldPlain(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	var b strings.Builder
	for i, s := range parts {
		switch {
		case i == 0:
			b.WriteString(s)
		case s == "\n":
			b.WriteString("\n")
		case parts[i-1] == "\n":
			b.WriteString(s)
		default:
			b.WriteString(" " + s)
		}
	}
	return b.String()
}

// block parses a | or > scalar whose header is s and whose content lines
// are indented deeper than parent.
func (p *parser) block(s string, parent int) (any, error) {
	style := s[0]
	chomp := byte(0)
	explicit := 0
	for _, c := range []byte(strings.TrimSpace(stripComment(s[1:]))) {
		switch {
		case c == '+' || c == '-':
			chomp = c
		case c >= '1' && c <= '9':
			explicit = int(c - '0')
		default:
			return nil, p.fail("bad block scalar header %q", clip(s))
		}
	}
	p.pos++
	var rows []string
	blockIndent := 0
	if explicit > 0 {
		blockIndent = max(parent, 0) + explicit
		if parent < 0 {
			blockIndent = explicit
		}
	}
	for p.pos < len(p.lines) {
		l := p.lines[p.pos]
		if l.text == "" {
			rows = append(rows, "")
			p.pos++
			continue
		}
		if blockIndent == 0 {
			if l.indent <= parent {
				break
			}
			blockIndent = l.indent
		}
		if l.indent < blockIndent {
			break
		}
		rows = append(rows, strings.Repeat(" ", l.indent-blockIndent)+l.text)
		p.pos++
	}
	// trailing blank rows belong to the chomping rule
	trail := 0
	for len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
		trail++
	}
	var body string
	if style == '|' {
		body = strings.Join(rows, "\n")
	} else {
		var b strings.Builder
		for i, r := range rows {
			if i > 0 {
				prev := rows[i-1]
				switch {
				case r == "" || prev == "":
					b.WriteString("\n")
				case strings.HasPrefix(r, " ") || strings.HasPrefix(prev, " "):
					b.WriteString("\n")
				default:
					b.WriteString(" ")
				}
			}
			b.WriteString(r)
		}
		body = b.String()
		// a blank row already produced one newline; collapse the fold rule
		body = strings.ReplaceAll(body, "\n\n", "\n")
	}
	if len(rows) == 0 {
		if chomp == '+' {
			return strings.Repeat("\n", trail), nil
		}
		return "", nil
	}
	switch chomp {
	case '-':
		return body, nil
	case '+':
		return body + "\n" + strings.Repeat("\n", trail), nil
	}
	return body + "\n", nil
}

// splitKey reports whether code is a "key: rest" line and returns both parts.
func splitKey(code string) (key, rest string, ok bool, err error) {
	if code == "" {
		return
	}
	switch code[0] {
	case '"', '\'':
		v, n, perr := parseFlowValue(code, 0)
		if perr != nil {
			if perr == errEOF {
				return "", "", false, nil
			}
			return "", "", false, perr
		}
		t := strings.TrimLeft(code[n:], " ")
		if t == ":" || strings.HasPrefix(t, ": ") {
			return v.(string), strings.TrimSpace(t[1:]), true, nil
		}
		return "", "", false, nil
	case '[', '{', '|', '>', '*', '#':
		return
	case '-':
		if isSeqItem(code) {
			return
		}
	}
	for i := 0; i < len(code); i++ {
		if code[i] == ':' && (i+1 == len(code) || code[i+1] == ' ') {
			k := strings.TrimSpace(code[:i])
			if k == "" {
				return "", "", false, nil
			}
			_, k2 := props(k)
			return scalarKey(resolvePlain(k2, "")), strings.TrimSpace(code[i+1:]), true, nil
		}
	}
	return
}

func scalarKey(v any) string {
	if s, ok := tools.Scalar(v); ok {
		if v == nil {
			return "null"
		}
		return s
	}
	return fmt.Sprint(v)
}

// stripComment removes a trailing " # comment" that is outside quotes.
func stripComment(s string) string {
	inS, inD := false, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\' && inD:
			i++
		case c == '"' && !inS:
			inD = !inD
		case c == '\'' && !inD:
			if inS && i+1 < len(s) && s[i+1] == '\'' {
				i++
			} else if inS || i == 0 || strings.ContainsRune(" [{,:", rune(s[i-1])) {
				inS = !inS
			}
		case c == '#' && !inS && !inD && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t'):
			return strings.TrimRight(s[:i], " \t")
		}
	}
	return s
}

// ---- flow collections and quoted scalars ----

func skipWS(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case ' ', '\t', '\n':
			i++
		case '#':
			if i == 0 || s[i-1] == ' ' || s[i-1] == '\n' {
				for i < len(s) && s[i] != '\n' {
					i++
				}
				continue
			}
			return i
		default:
			return i
		}
	}
	return i
}

// parseFlowValue parses one flow node from s[i:] and returns the index after it.
func parseFlowValue(s string, i int) (any, int, error) {
	return parseFlow(s, i, 0)
}

func parseFlow(s string, i, depth int) (any, int, error) {
	if depth > maxDepth {
		return nil, 0, fmt.Errorf("flow collection is nested deeper than %d levels", maxDepth)
	}
	i = skipWS(s, i)
	if i >= len(s) {
		return nil, 0, errEOF
	}
	switch s[i] {
	case '[':
		out := []any{}
		i++
		for {
			i = skipWS(s, i)
			if i >= len(s) {
				return nil, 0, errEOF
			}
			if s[i] == ']' {
				return out, i + 1, nil
			}
			v, n, err := parseFlow(s, i, depth+1)
			if err != nil {
				return nil, 0, err
			}
			i = skipWS(s, n)
			if i >= len(s) {
				return nil, 0, errEOF
			}
			if s[i] == ':' && (i+1 >= len(s) || s[i+1] == ' ' || s[i+1] == '\n') { // single-pair map
				val, n2, err := parseFlow(s, i+1, depth+1)
				if err != nil {
					return nil, 0, err
				}
				d := tools.NewDict()
				d.Set(scalarKey(v), val)
				v, i = d, skipWS(s, n2)
				if i >= len(s) {
					return nil, 0, errEOF
				}
			}
			out = append(out, v)
			switch s[i] {
			case ',':
				i++
			case ']':
			default:
				return nil, 0, fmt.Errorf("expected , or ] in a flow list, got %q", string(s[i]))
			}
		}
	case '{':
		d := tools.NewDict()
		i++
		for {
			i = skipWS(s, i)
			if i >= len(s) {
				return nil, 0, errEOF
			}
			if s[i] == '}' {
				return d, i + 1, nil
			}
			k, n, err := parseFlow(s, i, depth+1)
			if err != nil {
				return nil, 0, err
			}
			i = skipWS(s, n)
			if i >= len(s) {
				return nil, 0, errEOF
			}
			var v any
			if s[i] == ':' {
				v, n, err = parseFlow(s, i+1, depth+1)
				if err == errEOF {
					return nil, 0, err
				}
				if err != nil {
					return nil, 0, err
				}
				i = skipWS(s, n)
				if i >= len(s) {
					return nil, 0, errEOF
				}
			}
			d.Set(scalarKey(k), v)
			switch s[i] {
			case ',':
				i++
			case '}':
			default:
				return nil, 0, fmt.Errorf("expected , or } in a flow mapping, got %q", string(s[i]))
			}
		}
	case '"':
		return parseDouble(s, i)
	case '\'':
		var b strings.Builder
		for j := i + 1; j < len(s); j++ {
			if s[j] == '\'' {
				if j+1 < len(s) && s[j+1] == '\'' {
					b.WriteByte('\'')
					j++
					continue
				}
				return foldQuoted(b.String()), j + 1, nil
			}
			b.WriteByte(s[j])
		}
		return nil, 0, errEOF
	case ']', '}', ',':
		return nil, 0, fmt.Errorf("unexpected %q", string(s[i]))
	case '*':
		return nil, 0, fmt.Errorf("aliases are not supported inside flow collections")
	}
	// plain scalar inside a flow context
	j := i
	for j < len(s) {
		c := s[j]
		if c == ',' || c == ']' || c == '}' || c == '\n' {
			break
		}
		if c == ':' && (j+1 >= len(s) || s[j+1] == ' ' || s[j+1] == '\n' || s[j+1] == ',' || s[j+1] == ']' || s[j+1] == '}') {
			break
		}
		if c == '#' && j > 0 && s[j-1] == ' ' {
			break
		}
		j++
	}
	return resolvePlain(strings.TrimSpace(s[i:j]), ""), j, nil
}

func foldQuoted(s string) string {
	if !strings.Contains(s, "\n") {
		return s
	}
	ls := strings.Split(s, "\n")
	var b strings.Builder
	for i, l := range ls {
		l = strings.TrimSpace(l)
		if i == 0 {
			b.WriteString(strings.TrimRight(ls[0], " \t"))
			continue
		}
		if l == "" {
			b.WriteString("\n")
			continue
		}
		if prev := b.String(); prev != "" && !strings.HasSuffix(prev, "\n") {
			b.WriteString(" ")
		}
		b.WriteString(l)
	}
	return b.String()
}

func parseDouble(s string, i int) (any, int, error) {
	var b strings.Builder
	for j := i + 1; j < len(s); j++ {
		c := s[j]
		if c == '"' {
			return foldQuoted(b.String()), j + 1, nil
		}
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		j++
		if j >= len(s) {
			return nil, 0, errEOF
		}
		switch s[j] {
		case '0':
			b.WriteByte(0)
		case 'a':
			b.WriteByte(7)
		case 'b':
			b.WriteByte(8)
		case 't', '\t':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'v':
			b.WriteByte(11)
		case 'f':
			b.WriteByte(12)
		case 'r':
			b.WriteByte('\r')
		case 'e':
			b.WriteByte(27)
		case ' ', '"', '/', '\\':
			b.WriteByte(s[j])
		case 'N':
			b.WriteString("\u0085")
		case '_':
			b.WriteString(" ")
		case 'L':
			b.WriteString(" ")
		case 'P':
			b.WriteString(" ")
		case 'x', 'u', 'U':
			n := map[byte]int{'x': 2, 'u': 4, 'U': 8}[s[j]]
			if j+n >= len(s) {
				return nil, 0, errEOF
			}
			cp, err := strconv.ParseUint(s[j+1:j+1+n], 16, 32)
			if err != nil || !utf8.ValidRune(rune(cp)) {
				return nil, 0, fmt.Errorf("bad escape \\%s", s[j:j+1+n])
			}
			b.WriteRune(rune(cp))
			j += n
		case '\n':
			for j+1 < len(s) && (s[j+1] == ' ' || s[j+1] == '\t') {
				j++
			}
		default:
			return nil, 0, fmt.Errorf("unknown escape \\%c", s[j])
		}
	}
	return nil, 0, errEOF
}

// ---- scalar typing ----

func resolveTag(tag, s string, empty bool) any {
	switch tag {
	case "!!str", "!str":
		return s
	case "":
		if empty {
			return nil
		}
		return resolvePlain(s, "")
	}
	return resolvePlain(s, tag)
}

func resolvePlain(s, tag string) any {
	switch tag {
	case "!!str", "!str":
		return s
	case "!!int":
		if v, ok := parseInt(s); ok {
			return v
		}
	case "!!float":
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	case "!!bool":
		if b, ok := parseBool(s); ok {
			return b
		}
	case "!!null":
		return nil
	}
	switch s {
	case "", "~", "null", "Null", "NULL":
		return nil
	}
	if b, ok := parseBool(s); ok {
		return b
	}
	if v, ok := parseInt(s); ok {
		return v
	}
	switch s {
	case ".inf", ".Inf", ".INF", "+.inf", "+.Inf", "+.INF":
		return math.Inf(1)
	case "-.inf", "-.Inf", "-.INF":
		return math.Inf(-1)
	case ".nan", ".NaN", ".NAN":
		return math.NaN()
	}
	if looksFloat(s) {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return s
}

func parseBool(s string) (bool, bool) {
	switch s {
	case "true", "True", "TRUE":
		return true, true
	case "false", "False", "FALSE":
		return false, true
	}
	return false, false
}

func parseInt(s string) (int64, bool) {
	t := s
	neg := false
	if strings.HasPrefix(t, "-") || strings.HasPrefix(t, "+") {
		neg = t[0] == '-'
		t = t[1:]
	}
	if t == "" {
		return 0, false
	}
	base := 10
	switch {
	case strings.HasPrefix(t, "0x"):
		base, t = 16, t[2:]
	case strings.HasPrefix(t, "0o"):
		base, t = 8, t[2:]
	case strings.HasPrefix(t, "0b"):
		base, t = 2, t[2:]
	}
	if t == "" {
		return 0, false
	}
	for i := 0; i < len(t); i++ {
		if t[i] == '_' || t[i] == '+' || t[i] == '-' {
			return 0, false
		}
	}
	u, err := strconv.ParseUint(t, base, 64)
	if err != nil {
		return 0, false
	}
	if neg {
		if u > 1<<63 {
			return 0, false
		}
		return -int64(u), true
	}
	if u > math.MaxInt64 {
		return 0, false
	}
	return int64(u), true
}

func looksFloat(s string) bool {
	if s == "" {
		return false
	}
	digit, dot, exp := false, false, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
			digit = true
		case c == '.' && !dot && !exp:
			dot = true
		case (c == 'e' || c == 'E') && digit && !exp:
			exp = true
		case (c == '+' || c == '-') && (i == 0 || s[i-1] == 'e' || s[i-1] == 'E'):
		default:
			return false
		}
	}
	return digit
}
