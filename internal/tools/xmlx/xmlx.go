// Package xmlx registers the "xml" tool group: parse XML into a tree of
// dicts, query it with simple paths and write it back. Pure: it works on
// strings and values only.
package xmlx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxDepth = 256
	maxNodes = 1_000_000
)

func init() {
	tools.Register(tools.Func{
		Group: "xml", Name: "parse", Title: "Parse XML",
		Doc:     "Parse XML text into a tree. Every element is a dict with tag (the name as written, prefix included), attrs (dict of strings), text (the element's own text, trimmed) and children (list of elements). Comments, processing instructions and the DOCTYPE are dropped; CDATA becomes text; the five standard entities and numeric references are decoded. Raises on malformed XML, more than one root, or nesting beyond 256 levels.",
		Terms:   []string{"xml to dict", "xml to json", "read xml", "load xml", "rss", "atom", "svg", "sitemap", "deserialize xml"},
		Params:  []tools.Param{tools.Req("text", "str", "The XML text.")},
		Returns: "any",
		Example: `xml.parse("<a id='1'><b>hi</b></a>")`,
		Tags:    []string{"xml", "parse", "tree", "dom"},
		Fn: func(a tools.Args) (any, error) {
			n, err := parse(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return n.dict(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "xml", Name: "query", Title: "Query XML",
		Doc:   "Select parts of an XML document with a simple path. source is XML text or a tree from xml.parse. The path starts at the document, so the first step names the root: /library/book/title. Steps: a name or *, // for any depth, [n] position (1-based), [last()], [@attr], [@attr='v'], [child='v']. A final @attr returns attribute values and text() returns text, both as strings; otherwise the result is a list of elements. Raises on a bad path or XML.",
		Terms: []string{"xpath", "find elements", "select nodes", "xml search", "get attribute", "extract from xml", "rss items"},
		Params: []tools.Param{
			tools.Req("source", "any", "XML text, or a tree returned by xml.parse."),
			tools.Req("path", "str", "The path, for example //item[@id='2']/title or /feed/entry/@href."),
		},
		Returns: "any",
		Example: `xml.query("<l><b id='1'>x</b><b id='2'>y</b></l>", "/l/b[@id='2']/text()")`,
		Tags:    []string{"xml", "query", "xpath", "select", "find"},
		Fn: func(a tools.Args) (any, error) {
			root, err := source(a["source"])
			if err != nil {
				return nil, err
			}
			res, err := query(root, a.Str("path"))
			if err != nil {
				return nil, err
			}
			out := make([]any, len(res))
			for i, r := range res {
				if n, ok := r.(*node); ok {
					r = n.dict()
				}
				out[i] = r
			}
			return out, nil
		},
	})
	tools.Register(tools.Func{
		Group: "xml", Name: "text", Title: "XML text content",
		Doc:   "Return all the text inside an element and its descendants, in document order, joined by the separator. source is XML text or a tree from xml.parse; path is optional and selects the first matching element.",
		Terms: []string{"strip tags", "inner text", "all text", "flatten xml"},
		Params: []tools.Param{
			tools.Req("source", "any", "XML text, or a tree returned by xml.parse."),
			tools.Opt("path", "str", "Optional path of the element to read; empty uses the root.", ""),
			tools.Opt("separator", "str", "Text between pieces.", " "),
		},
		Returns: "str",
		Example: `xml.text("<a>one <b>two</b></a>")`,
		Tags:    []string{"xml", "text", "strip", "inner"},
		Fn: func(a tools.Args) (any, error) {
			root, err := source(a["source"])
			if err != nil {
				return nil, err
			}
			n := root
			if p := a.Str("path"); p != "" {
				res, err := query(root, p)
				if err != nil {
					return nil, err
				}
				var found *node
				for _, r := range res {
					if e, ok := r.(*node); ok {
						found = e
						break
					}
				}
				if found == nil {
					return "", nil
				}
				n = found
			}
			var parts []string
			n.collect(&parts)
			return strings.Join(parts, a.Str("separator")), nil
		},
	})
	tools.Register(tools.Func{
		Group: "xml", Name: "dump", Title: "Write XML",
		Doc:   "Write a tree (the shape xml.parse returns) as XML text. Only tag is required; attrs, text and children are optional. Text and attribute values are escaped. Raises on an invalid tag or attribute name or a malformed tree. With indent empty the output is on one line.",
		Terms: []string{"dict to xml", "json to xml", "write xml", "serialize xml", "build xml", "pretty print xml"},
		Params: []tools.Param{
			tools.Req("node", "any", "A dict with tag, and optionally attrs, text and children."),
			tools.Opt("indent", "str", "Indentation per level; empty writes a single line.", "  "),
			tools.Opt("declaration", "bool", "Start with the <?xml ...?> declaration.", false),
		},
		Returns: "str",
		Example: `xml.dump({"tag": "a", "attrs": {"id": "1"}, "children": [{"tag": "b", "text": "hi"}]})`,
		Tags:    []string{"xml", "dump", "write", "serialize", "pretty"},
		Fn: func(a tools.Args) (any, error) {
			n, err := fromValue(a["node"], 0)
			if err != nil {
				return nil, err
			}
			var b strings.Builder
			if a.Bool("declaration") {
				b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
				if a.Str("indent") != "" {
					b.WriteString("\n")
				}
			}
			n.write(&b, a.Str("indent"), 0)
			if a.Str("indent") != "" {
				b.WriteString("\n")
			}
			return b.String(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "xml", Name: "is_valid", Title: "Is valid XML",
		Doc:     "Report whether text is well-formed XML with a single root element.",
		Terms:   []string{"validate xml", "check xml", "well formed"},
		Params:  []tools.Param{tools.Req("text", "str", "The XML text.")},
		Returns: "bool",
		Example: `xml.is_valid("<a><b></a>")`,
		Tags:    []string{"xml", "validate", "check"},
		Fn: func(a tools.Args) (any, error) {
			_, err := parse(a.Str("text"))
			return err == nil, nil
		},
	})
}

type node struct {
	tag      string
	attrs    []attr
	text     string
	children []*node
}

type attr struct{ k, v string }

func (n *node) dict() *tools.Dict {
	d := tools.NewDict()
	d.Set("tag", n.tag)
	at := tools.NewDict()
	for _, a := range n.attrs {
		at.Set(a.k, a.v)
	}
	d.Set("attrs", at)
	d.Set("text", n.text)
	ch := make([]any, len(n.children))
	for i, c := range n.children {
		ch[i] = c.dict()
	}
	d.Set("children", ch)
	return d
}

func (n *node) attr(name string) (string, bool) {
	for _, a := range n.attrs {
		if a.k == name {
			return a.v, true
		}
	}
	return "", false
}

func (n *node) collect(parts *[]string) {
	if n.text != "" {
		*parts = append(*parts, n.text)
	}
	for _, c := range n.children {
		c.collect(parts)
	}
}

func qname(n xml.Name) string {
	if n.Space != "" {
		return n.Space + ":" + n.Local
	}
	return n.Local
}

func parse(text string) (*node, error) {
	dec := xml.NewDecoder(strings.NewReader(strings.TrimPrefix(text, "\uFEFF")))
	dec.Strict = true
	var root *node
	var stack []*node
	var buf []*strings.Builder
	count := 0
	for {
		tok, err := dec.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml: %v", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if len(stack) == 0 && root != nil {
				return nil, fmt.Errorf("xml: more than one root element")
			}
			if len(stack) >= maxDepth {
				return nil, fmt.Errorf("xml: nesting is deeper than %d levels", maxDepth)
			}
			if count++; count > maxNodes {
				return nil, fmt.Errorf("xml: more than %d elements", maxNodes)
			}
			n := &node{tag: qname(t.Name)}
			for _, a := range t.Attr {
				n.attrs = append(n.attrs, attr{qname(a.Name), a.Value})
			}
			if len(stack) > 0 {
				p := stack[len(stack)-1]
				p.children = append(p.children, n)
			} else {
				root = n
			}
			stack = append(stack, n)
			buf = append(buf, &strings.Builder{})
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("xml: unexpected closing tag </%s>", qname(t.Name))
			}
			n := stack[len(stack)-1]
			if n.tag != qname(t.Name) {
				return nil, fmt.Errorf("xml: closing tag </%s> does not match <%s>", qname(t.Name), n.tag)
			}
			n.text = strings.TrimSpace(buf[len(buf)-1].String())
			stack, buf = stack[:len(stack)-1], buf[:len(buf)-1]
		case xml.CharData:
			if len(stack) > 0 {
				buf[len(buf)-1].Write(t)
			} else if len(bytes.TrimSpace(t)) > 0 {
				return nil, fmt.Errorf("xml: text outside the root element")
			}
		}
	}
	if len(stack) > 0 {
		return nil, fmt.Errorf("xml: element <%s> is not closed", stack[len(stack)-1].tag)
	}
	if root == nil {
		return nil, fmt.Errorf("xml: no root element")
	}
	return root, nil
}

// source accepts XML text or a tree produced by parse.
func source(v any) (*node, error) {
	switch x := v.(type) {
	case string:
		return parse(x)
	case *tools.Dict:
		return fromValue(x, 0)
	}
	return nil, fmt.Errorf("source must be XML text or a tree from xml.parse")
}

func fromValue(v any, depth int) (*node, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("xml: tree is nested deeper than %d levels", maxDepth)
	}
	d, ok := v.(*tools.Dict)
	if !ok {
		return nil, fmt.Errorf("an XML node must be a dict with a tag")
	}
	t, _ := d.Get("tag")
	tag, ok := t.(string)
	if !ok || !validName(tag) {
		return nil, fmt.Errorf("invalid or missing tag %q", fmt.Sprint(t))
	}
	n := &node{tag: tag}
	if at, ok := d.Get("attrs"); ok && at != nil {
		ad, ok := at.(*tools.Dict)
		if !ok {
			return nil, fmt.Errorf("<%s>: attrs must be a dict", tag)
		}
		for _, k := range ad.Keys() {
			if !validName(k) {
				return nil, fmt.Errorf("<%s>: invalid attribute name %q", tag, k)
			}
			e, _ := ad.Get(k)
			s, ok := tools.Scalar(e)
			if !ok {
				return nil, fmt.Errorf("<%s>: attribute %q must be a scalar", tag, k)
			}
			n.attrs = append(n.attrs, attr{k, s})
		}
	}
	if tx, ok := d.Get("text"); ok {
		s, ok := tools.Scalar(tx)
		if !ok {
			return nil, fmt.Errorf("<%s>: text must be a scalar", tag)
		}
		n.text = s
	}
	if ch, ok := d.Get("children"); ok && ch != nil {
		l, ok := ch.([]any)
		if !ok {
			return nil, fmt.Errorf("<%s>: children must be a list", tag)
		}
		for _, c := range l {
			cn, err := fromValue(c, depth+1)
			if err != nil {
				return nil, err
			}
			n.children = append(n.children, cn)
		}
	}
	return n, nil
}

func validName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		ok := unicode.IsLetter(r) || r == '_' || r == ':'
		if i > 0 {
			ok = ok || unicode.IsDigit(r) || r == '-' || r == '.'
		}
		if !ok {
			return false
		}
	}
	return true
}

func esc(s string, attr bool) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			if attr {
				b.WriteString("&quot;")
			} else {
				b.WriteRune(r)
			}
		case '\n', '\t', '\r':
			if attr {
				fmt.Fprintf(&b, "&#%d;", r)
			} else {
				b.WriteRune(r)
			}
		default:
			if r < 0x20 || r == 0xFFFE || r == 0xFFFF {
				continue // not allowed in XML 1.0
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (n *node) write(b *strings.Builder, indent string, level int) {
	pad := ""
	if indent != "" {
		pad = strings.Repeat(indent, level)
	}
	b.WriteString(pad + "<" + n.tag)
	for _, a := range n.attrs {
		b.WriteString(" " + a.k + `="` + esc(a.v, true) + `"`)
	}
	if n.text == "" && len(n.children) == 0 {
		b.WriteString("/>")
		return
	}
	b.WriteString(">")
	if len(n.children) == 0 {
		b.WriteString(esc(n.text, false) + "</" + n.tag + ">")
		return
	}
	nl := ""
	if indent != "" {
		nl = "\n"
	}
	if n.text != "" {
		b.WriteString(nl + pad + indent + esc(n.text, false))
	}
	for _, c := range n.children {
		b.WriteString(nl)
		c.write(b, indent, level+1)
	}
	b.WriteString(nl + pad + "</" + n.tag + ">")
}
