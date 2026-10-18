package htmlx

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// A selector is a comma separated list of complex selectors. Supported:
// type, *, #id, .class, [attr], [attr=v], [attr~=v], [attr|=v], [attr^=v],
// [attr$=v], [attr*=v] (with an optional i flag), the combinators space,
// >, + and ~, and the pseudo-classes :first-child, :last-child, :only-child,
// :nth-child(), :nth-last-child(), :first-of-type, :last-of-type,
// :nth-of-type(), :empty, :root, :not(), :is() and :checked/:disabled.

type attrSel struct {
	name, op, val string
	fold          bool
}

type pseudo struct {
	name  string
	a, b  int // an+b for the nth-* family
	inner []complexSel
	last  bool
}

type compound struct {
	tag     string // empty or * means any
	ids     []string
	classes []string
	attrs   []attrSel
	pseudos []pseudo
}

type link struct {
	comb byte // relation to the previous compound: ' ', '>', '+', '~' (0 for the first)
	c    compound
}

type complexSel []link

const (
	maxSelector = 2000
	maxLinks    = 8
)

func parseSelector(s string) ([]complexSel, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty selector")
	}
	if len(s) > maxSelector {
		return nil, fmt.Errorf("selector is longer than %d characters", maxSelector)
	}
	p := &selParser{s: s}
	list, err := p.list(0)
	if err != nil {
		return nil, fmt.Errorf("bad selector %q: %v", s, err)
	}
	if p.i < len(p.s) {
		return nil, fmt.Errorf("bad selector %q: unexpected %q at %d", s, string(p.s[p.i]), p.i)
	}
	return list, nil
}

type selParser struct {
	s string
	i int
}

func (p *selParser) ws() bool {
	start := p.i
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\t' || p.s[p.i] == '\n' || p.s[p.i] == '\r') {
		p.i++
	}
	return p.i > start
}

func (p *selParser) list(depth int) ([]complexSel, error) {
	if depth > 10 {
		return nil, fmt.Errorf("selector is nested too deeply")
	}
	var out []complexSel
	for {
		c, err := p.complex(depth)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
		p.ws()
		if p.i < len(p.s) && p.s[p.i] == ',' {
			p.i++
			continue
		}
		return out, nil
	}
}

func (p *selParser) complex(depth int) (complexSel, error) {
	var cs complexSel
	p.ws()
	comb := byte(0)
	for {
		c, err := p.compound(depth)
		if err != nil {
			return nil, err
		}
		if len(cs) >= maxLinks {
			return nil, fmt.Errorf("a selector may have at most %d parts", maxLinks)
		}
		cs = append(cs, link{comb, c})
		hadWS := p.ws()
		if p.i >= len(p.s) || p.s[p.i] == ',' || p.s[p.i] == ')' {
			return cs, nil
		}
		switch p.s[p.i] {
		case '>', '+', '~':
			comb = p.s[p.i]
			p.i++
			p.ws()
		default:
			if !hadWS {
				return nil, fmt.Errorf("unexpected %q at %d", string(p.s[p.i]), p.i)
			}
			comb = ' '
		}
	}
}

func isIdent(r byte) bool {
	return r == '-' || r == '_' || r >= 0x80 || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func (p *selParser) ident() string {
	start := p.i
	for p.i < len(p.s) && isIdent(p.s[p.i]) {
		p.i++
	}
	return p.s[start:p.i]
}

func (p *selParser) compound(depth int) (compound, error) {
	var c compound
	empty := true
	if p.i < len(p.s) && p.s[p.i] == '*' {
		c.tag = "*"
		p.i++
		empty = false
	} else if id := p.ident(); id != "" {
		c.tag = strings.ToLower(id)
		empty = false
	}
	for p.i < len(p.s) {
		switch p.s[p.i] {
		case '#':
			p.i++
			id := p.ident()
			if id == "" {
				return c, fmt.Errorf("missing id after # at %d", p.i)
			}
			c.ids = append(c.ids, id)
		case '.':
			p.i++
			cl := p.ident()
			if cl == "" {
				return c, fmt.Errorf("missing class after . at %d", p.i)
			}
			c.classes = append(c.classes, cl)
		case '[':
			a, err := p.attr()
			if err != nil {
				return c, err
			}
			c.attrs = append(c.attrs, a)
		case ':':
			ps, err := p.pseudo(depth)
			if err != nil {
				return c, err
			}
			c.pseudos = append(c.pseudos, ps)
		default:
			if empty {
				return c, fmt.Errorf("unexpected %q at %d", string(p.s[p.i]), p.i)
			}
			return c, nil
		}
		empty = false
	}
	if empty {
		return c, fmt.Errorf("unexpected end of selector")
	}
	return c, nil
}

func (p *selParser) attr() (attrSel, error) {
	p.i++ // [
	p.ws()
	var a attrSel
	a.name = strings.ToLower(p.ident())
	if a.name == "" {
		return a, fmt.Errorf("missing attribute name at %d", p.i)
	}
	p.ws()
	if p.i < len(p.s) && p.s[p.i] == ']' {
		p.i++
		return a, nil
	}
	for _, op := range []string{"~=", "|=", "^=", "$=", "*=", "="} {
		if strings.HasPrefix(p.s[p.i:], op) {
			a.op = op
			p.i += len(op)
			break
		}
	}
	if a.op == "" {
		return a, fmt.Errorf("bad attribute selector at %d", p.i)
	}
	p.ws()
	if p.i < len(p.s) && (p.s[p.i] == '"' || p.s[p.i] == '\'') {
		q := p.s[p.i]
		p.i++
		start := p.i
		for p.i < len(p.s) && p.s[p.i] != q {
			p.i++
		}
		if p.i >= len(p.s) {
			return a, fmt.Errorf("unterminated string in attribute selector")
		}
		a.val = p.s[start:p.i]
		p.i++
	} else {
		a.val = p.ident()
	}
	p.ws()
	if p.i < len(p.s) && (p.s[p.i] == 'i' || p.s[p.i] == 'I') {
		a.fold = true
		p.i++
		p.ws()
	}
	if p.i >= len(p.s) || p.s[p.i] != ']' {
		return a, fmt.Errorf("missing ] in attribute selector")
	}
	p.i++
	return a, nil
}

func (p *selParser) pseudo(depth int) (pseudo, error) {
	p.i++ // :
	if p.i < len(p.s) && p.s[p.i] == ':' {
		return pseudo{}, fmt.Errorf("pseudo-elements are not supported")
	}
	name := strings.ToLower(p.ident())
	ps := pseudo{name: name}
	switch name {
	case "first-child", "last-child", "only-child", "first-of-type", "last-of-type", "only-of-type", "empty", "root", "checked", "disabled", "enabled":
		return ps, nil
	case "nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
		arg, err := p.parenArg()
		if err != nil {
			return ps, err
		}
		ps.a, ps.b, err = parseNth(arg)
		ps.last = strings.Contains(name, "last")
		return ps, err
	case "not", "is", "where":
		if p.i >= len(p.s) || p.s[p.i] != '(' {
			return ps, fmt.Errorf(":%s needs (...)", name)
		}
		p.i++
		inner, err := p.list(depth + 1)
		if err != nil {
			return ps, err
		}
		p.ws()
		if p.i >= len(p.s) || p.s[p.i] != ')' {
			return ps, fmt.Errorf("missing ) after :%s(", name)
		}
		p.i++
		ps.inner = inner
		return ps, nil
	}
	return ps, fmt.Errorf("unsupported pseudo-class :%s", name)
}

func (p *selParser) parenArg() (string, error) {
	if p.i >= len(p.s) || p.s[p.i] != '(' {
		return "", fmt.Errorf("missing ( after pseudo-class")
	}
	end := strings.IndexByte(p.s[p.i:], ')')
	if end < 0 {
		return "", fmt.Errorf("missing )")
	}
	arg := strings.TrimSpace(p.s[p.i+1 : p.i+end])
	p.i += end + 1
	return arg, nil
}

func parseNth(s string) (a, b int, err error) {
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))
	switch s {
	case "odd":
		return 2, 1, nil
	case "even":
		return 2, 0, nil
	case "":
		return 0, 0, fmt.Errorf("empty :nth-child()")
	}
	i := strings.IndexByte(s, 'n')
	if i < 0 {
		b, err = strconv.Atoi(s)
		return 0, b, err
	}
	as := s[:i]
	switch as {
	case "", "+":
		a = 1
	case "-":
		a = -1
	default:
		if a, err = strconv.Atoi(as); err != nil {
			return 0, 0, fmt.Errorf("bad :nth-child(%s)", s)
		}
	}
	if rest := s[i+1:]; rest != "" {
		if b, err = strconv.Atoi(rest); err != nil {
			return 0, 0, fmt.Errorf("bad :nth-child(%s)", s)
		}
	}
	return a, b, nil
}

// ---- matching ----

func elementChildren(n *html.Node) []*html.Node {
	var out []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			out = append(out, c)
		}
	}
	return out
}

func prevElement(n *html.Node) *html.Node {
	for s := n.PrevSibling; s != nil; s = s.PrevSibling {
		if s.Type == html.ElementNode {
			return s
		}
	}
	return nil
}

func attrOf(n *html.Node, name string) (string, bool) {
	for _, a := range n.Attr {
		if a.Namespace == "" && strings.EqualFold(a.Key, name) {
			return a.Val, true
		}
	}
	return "", false
}

func (a attrSel) match(n *html.Node) bool {
	v, ok := attrOf(n, a.name)
	if !ok {
		return false
	}
	if a.op == "" {
		return true
	}
	want := a.val
	if a.fold {
		v, want = strings.ToLower(v), strings.ToLower(want)
	}
	switch a.op {
	case "=":
		return v == want
	case "~=":
		for _, f := range strings.Fields(v) {
			if f == want {
				return want != ""
			}
		}
		return false
	case "|=":
		return v == want || strings.HasPrefix(v, want+"-")
	case "^=":
		return want != "" && strings.HasPrefix(v, want)
	case "$=":
		return want != "" && strings.HasSuffix(v, want)
	case "*=":
		return want != "" && strings.Contains(v, want)
	}
	return false
}

func (c compound) match(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if c.tag != "" && c.tag != "*" && !strings.EqualFold(n.Data, c.tag) {
		return false
	}
	for _, id := range c.ids {
		if v, ok := attrOf(n, "id"); !ok || v != id {
			return false
		}
	}
	if len(c.classes) > 0 {
		v, _ := attrOf(n, "class")
		have := strings.Fields(v)
		for _, cl := range c.classes {
			found := false
			for _, h := range have {
				if h == cl {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	for _, a := range c.attrs {
		if !a.match(n) {
			return false
		}
	}
	for _, ps := range c.pseudos {
		if !ps.match(n) {
			return false
		}
	}
	return true
}

func nthMatch(a, b, pos int) bool {
	if a == 0 {
		return pos == b
	}
	d := pos - b
	return d%a == 0 && d/a >= 0
}

func (ps pseudo) match(n *html.Node) bool {
	parent := n.Parent
	sibs := []*html.Node{n}
	if parent != nil {
		sibs = elementChildren(parent)
	}
	pos, count := 0, 0
	same := func(x *html.Node) bool { return !isTypeName(ps.name) || strings.EqualFold(x.Data, n.Data) }
	for _, s := range sibs {
		if same(s) {
			count++
			if s == n {
				pos = count
			}
		}
	}
	switch ps.name {
	case "first-child", "first-of-type":
		return pos == 1
	case "last-child", "last-of-type":
		return pos == count
	case "only-child", "only-of-type":
		return count == 1
	case "nth-child", "nth-of-type":
		return nthMatch(ps.a, ps.b, pos)
	case "nth-last-child", "nth-last-of-type":
		return nthMatch(ps.a, ps.b, count-pos+1)
	case "empty":
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode || c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
				return false
			}
		}
		return true
	case "root":
		return parent != nil && parent.Type == html.DocumentNode
	case "checked":
		_, ok := attrOf(n, "checked")
		_, sel := attrOf(n, "selected")
		return ok || sel
	case "disabled":
		_, ok := attrOf(n, "disabled")
		return ok
	case "enabled":
		_, ok := attrOf(n, "disabled")
		return !ok
	case "not":
		for _, cs := range ps.inner {
			if cs.matches(n) {
				return false
			}
		}
		return true
	case "is", "where":
		for _, cs := range ps.inner {
			if cs.matches(n) {
				return true
			}
		}
		return false
	}
	return false
}

func isTypeName(name string) bool { return strings.HasSuffix(name, "of-type") }

// matches reports whether n matches the whole complex selector (the last
// compound is tested against n, earlier ones against its relatives).
func (cs complexSel) matches(n *html.Node) bool {
	return matchFrom(cs, len(cs)-1, n)
}

func matchFrom(cs complexSel, i int, n *html.Node) bool {
	if !cs[i].c.match(n) {
		return false
	}
	if i == 0 {
		return true
	}
	switch cs[i].comb {
	case '>':
		return n.Parent != nil && matchFrom(cs, i-1, n.Parent)
	case ' ':
		for p := n.Parent; p != nil; p = p.Parent {
			if matchFrom(cs, i-1, p) {
				return true
			}
		}
		return false
	case '+':
		pv := prevElement(n)
		return pv != nil && matchFrom(cs, i-1, pv)
	case '~':
		for pv := prevElement(n); pv != nil; pv = prevElement(pv) {
			if matchFrom(cs, i-1, pv) {
				return true
			}
		}
		return false
	}
	return false
}

// selectAll returns the elements under root (excluding root) that match any
// selector, in document order.
func selectAll(root *html.Node, list []complexSel, limit int) []*html.Node {
	var out []*html.Node
	var walk func(n *html.Node) bool
	walk = func(n *html.Node) bool {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				for _, cs := range list {
					if cs.matches(c) {
						out = append(out, c)
						if limit > 0 && len(out) >= limit {
							return true
						}
						break
					}
				}
			}
			if walk(c) {
				return true
			}
		}
		return false
	}
	walk(root)
	return out
}

func collapse(s string) string {
	return strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")
}
