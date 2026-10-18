package xmlx

import (
	"fmt"
	"strconv"
	"strings"
)

type pred struct {
	kind  string // pos, last, attr, child, text
	n     int
	name  string
	val   string
	hasV  bool
	negEq bool
}

type step struct {
	deep  bool   // reached through //
	name  string // element name, "*", "@attr" or "text()"
	preds []pred
}

func parsePath(p string) ([]step, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return nil, fmt.Errorf("empty path")
	}
	var steps []step
	i := 0
	for i < len(p) {
		deep := false
		if strings.HasPrefix(p[i:], "//") {
			deep = true
			i += 2
		} else if p[i] == '/' {
			i++
		} else if i > 0 {
			return nil, fmt.Errorf("expected / at position %d of %q", i, p)
		}
		j := i
		for j < len(p) && p[j] != '/' && p[j] != '[' {
			j++
		}
		name := strings.TrimSpace(p[i:j])
		if name == "" {
			return nil, fmt.Errorf("missing name at position %d of %q", i, p)
		}
		st := step{deep: deep, name: name}
		i = j
		for i < len(p) && p[i] == '[' {
			end, err := closeBracket(p, i)
			if err != nil {
				return nil, err
			}
			pr, err := parsePred(p[i+1 : end])
			if err != nil {
				return nil, err
			}
			st.preds = append(st.preds, pr)
			i = end + 1
		}
		steps = append(steps, st)
	}
	for k, s := range steps {
		if (strings.HasPrefix(s.name, "@") || s.name == "text()") && k != len(steps)-1 {
			return nil, fmt.Errorf("%s can only be the last step", s.name)
		}
		if strings.HasPrefix(s.name, "@") && len(s.name) == 1 {
			return nil, fmt.Errorf("missing attribute name after @")
		}
	}
	return steps, nil
}

func closeBracket(p string, open int) (int, error) {
	var q byte
	for i := open + 1; i < len(p); i++ {
		c := p[i]
		switch {
		case q != 0:
			if c == q {
				q = 0
			}
		case c == '\'' || c == '"':
			q = c
		case c == ']':
			return i, nil
		}
	}
	return 0, fmt.Errorf("missing ] in %q", p)
}

func parsePred(s string) (pred, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pred{}, fmt.Errorf("empty [] predicate")
	}
	if n, err := strconv.Atoi(s); err == nil {
		if n < 1 {
			return pred{}, fmt.Errorf("positions start at 1, got [%d]", n)
		}
		return pred{kind: "pos", n: n}, nil
	}
	if s == "last()" {
		return pred{kind: "last"}, nil
	}
	lhs, op, rhs := s, "", ""
	if i := strings.Index(s, "!="); i >= 0 {
		lhs, op, rhs = s[:i], "!=", s[i+2:]
	} else if i := strings.IndexByte(s, '='); i >= 0 {
		lhs, op, rhs = s[:i], "=", s[i+1:]
	}
	lhs = strings.TrimSpace(lhs)
	p := pred{hasV: op != "", negEq: op == "!="}
	if op != "" {
		rhs = strings.TrimSpace(rhs)
		if len(rhs) < 2 || (rhs[0] != '\'' && rhs[0] != '"') || rhs[len(rhs)-1] != rhs[0] {
			return pred{}, fmt.Errorf("a predicate value must be quoted: [%s]", s)
		}
		p.val = rhs[1 : len(rhs)-1]
	}
	switch {
	case strings.HasPrefix(lhs, "@") && len(lhs) > 1:
		p.kind, p.name = "attr", lhs[1:]
	case lhs == "text()" || lhs == ".":
		p.kind = "text"
	case lhs != "" && !strings.ContainsAny(lhs, "()@/ "):
		p.kind, p.name = "child", lhs
	default:
		return pred{}, fmt.Errorf("unsupported predicate [%s]", s)
	}
	return p, nil
}

func (p pred) match(n *node) bool {
	test := func(got string) bool {
		if !p.hasV {
			return true
		}
		return (got == p.val) != p.negEq
	}
	switch p.kind {
	case "attr":
		v, ok := n.attr(p.name)
		if !ok {
			return p.negEq && p.hasV
		}
		return test(v)
	case "text":
		return test(n.text)
	case "child":
		for _, c := range n.children {
			if c.tag == p.name || p.name == "*" {
				if !p.hasV || (c.text == p.val) != p.negEq {
					return true
				}
			}
		}
		return false
	}
	return true
}

func descendants(n *node, out *[]*node) {
	for _, c := range n.children {
		*out = append(*out, c)
		descendants(c, out)
	}
}

// query returns *node and string values.
func query(root *node, path string) ([]any, error) {
	steps, err := parsePath(path)
	if err != nil {
		return nil, fmt.Errorf("bad path %q: %v", path, err)
	}
	doc := &node{children: []*node{root}}
	ctx := []*node{doc}
	for k, st := range steps {
		last := k == len(steps)-1
		if last && (strings.HasPrefix(st.name, "@") || st.name == "text()") {
			holders := ctx
			if st.deep {
				holders = nil
				for _, c := range ctx {
					holders = append(holders, c)
					descendants(c, &holders)
				}
			}
			var out []any
			for _, h := range holders {
				if h == doc {
					continue
				}
				if st.name == "text()" {
					if h.text != "" {
						out = append(out, h.text)
					}
					continue
				}
				name := st.name[1:]
				for _, a := range h.attrs {
					if name == "*" || a.k == name {
						out = append(out, a.v)
					}
				}
			}
			return out, nil
		}
		var parents []*node
		for _, c := range ctx {
			parents = append(parents, c)
			if st.deep {
				descendants(c, &parents)
			}
		}
		var next []*node
		seen := map[*node]bool{}
		for _, par := range parents {
			var sel []*node
			for _, c := range par.children {
				if st.name == "*" || c.tag == st.name {
					sel = append(sel, c)
				}
			}
			for _, pr := range st.preds {
				var kept []*node
				for i, c := range sel {
					switch pr.kind {
					case "pos":
						if i+1 == pr.n {
							kept = append(kept, c)
						}
					case "last":
						if i == len(sel)-1 {
							kept = append(kept, c)
						}
					default:
						if pr.match(c) {
							kept = append(kept, c)
						}
					}
				}
				sel = kept
			}
			for _, c := range sel {
				if !seen[c] {
					seen[c] = true
					next = append(next, c)
				}
			}
		}
		ctx = next
	}
	out := make([]any, len(ctx))
	for i, c := range ctx {
		out[i] = c
	}
	return out, nil
}
