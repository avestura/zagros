// Package htmlx registers the "html" tool group: parse HTML, select elements
// with CSS selectors and pull out text and attributes. Escaping and
// unescaping live in the encode group. Pure: it works on strings only.
package htmlx

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	maxInput = 16 << 20
	maxDepth = 512
)

func init() {
	tools.Register(tools.Func{
		Group: "html", Name: "parse", Title: "Parse HTML",
		Doc:   "Parse HTML the way a browser does (missing tags are filled in, bad markup is repaired) into a tree. Every element is a dict with tag (lower case), attrs (dict of strings), text (the element's own text with whitespace collapsed) and children (list of elements). With fragment=True the markup is treated as the inside of a body and the result is a dict with tag \"#fragment\"; otherwise it is the html element. Comments and doctype are dropped. Never raises on bad markup; raises on input over 16 MB.",
		Terms: []string{"html to dict", "html to json", "dom", "scrape", "web page", "read html", "parse web page", "html tree"},
		Params: []tools.Param{
			tools.Req("text", "str", "The HTML text."),
			tools.Opt("fragment", "bool", "Parse as a fragment rather than a whole page.", false),
		},
		Returns: "any",
		Example: `html.parse("<ul><li>a</li><li>b</li></ul>", fragment=True)`,
		Tags:    []string{"html", "parse", "dom", "tree", "scrape"},
		Fn: func(a tools.Args) (any, error) {
			root, err := parseHTML(a.Str("text"), a.Bool("fragment"))
			if err != nil {
				return nil, err
			}
			if a.Bool("fragment") {
				d := tools.NewDict()
				d.Set("tag", "#fragment")
				d.Set("attrs", tools.NewDict())
				d.Set("text", "")
				var kids []any
				for _, c := range elementChildren(root) {
					kids = append(kids, toDict(c, 0))
				}
				if kids == nil {
					kids = []any{}
				}
				d.Set("children", kids)
				return d, nil
			}
			for _, c := range elementChildren(root) {
				return toDict(c, 0), nil
			}
			return nil, fmt.Errorf("html: no content")
		},
	})
	tools.Register(tools.Func{
		Group: "html", Name: "select", Title: "Select HTML elements",
		Doc:   "Find the elements of an HTML page that match a CSS selector and return them as element dicts (see html.parse). Supported: type, *, #id, .class, [attr] with = ~= |= ^= $= *= (and the i flag), the combinators space > + ~, comma lists, and :first-child, :last-child, :only-child, :nth-child(an+b|odd|even), :nth-last-child, :first-of-type, :last-of-type, :nth-of-type, :empty, :root, :not(), :is(), :checked, :disabled. Results are in document order; limit caps their number (0 means all). Raises on an invalid selector.",
		Terms: []string{"css selector", "querySelectorAll", "jquery", "find elements", "scrape", "cheerio", "beautifulsoup", "extract from html"},
		Params: []tools.Param{
			tools.Req("text", "str", "The HTML text."),
			tools.Req("selector", "str", "A CSS selector such as a.external[href^=\"https\"]."),
			tools.Opt("limit", "int", "Maximum number of results; 0 returns all.", int64(0)),
		},
		Returns: "any",
		Example: `html.select("<p class='x'>a</p><p>b</p>", "p.x")`,
		Tags:    []string{"html", "select", "css", "selector", "query", "find"},
		Fn: func(a tools.Args) (any, error) {
			nodes, err := run(a)
			if err != nil {
				return nil, err
			}
			out := make([]any, len(nodes))
			for i, n := range nodes {
				out[i] = toDict(n, 0)
			}
			return out, nil
		},
	})
	tools.Register(tools.Func{
		Group: "html", Name: "text", Title: "HTML text",
		Doc:   "Extract the readable text of an HTML page, or of the elements matching a CSS selector. Scripts, styles, templates and comments are skipped, block elements and <br> break lines, and runs of spaces collapse. With a selector, the text of each match is returned on its own line.",
		Terms: []string{"strip html", "html to text", "visible text", "innerText", "textContent", "plain text from html", "readable text"},
		Params: []tools.Param{
			tools.Req("text", "str", "The HTML text."),
			tools.Opt("selector", "str", "Optional CSS selector; empty reads the whole page.", ""),
		},
		Returns: "str",
		Example: `html.text("<h1>Title</h1><p>Hello <b>world</b></p>")`,
		Tags:    []string{"html", "text", "strip", "innertext", "readable"},
		Fn: func(a tools.Args) (any, error) {
			root, err := parseHTML(a.Str("text"), false)
			if err != nil {
				return nil, err
			}
			targets := []*html.Node{root}
			if sel := a.Str("selector"); sel != "" {
				list, err := parseSelector(sel)
				if err != nil {
					return nil, err
				}
				targets = selectAll(root, list, 0)
			}
			var parts []string
			for _, t := range targets {
				var b strings.Builder
				visibleText(t, &b)
				if s := tidy(b.String()); s != "" {
					parts = append(parts, s)
				}
			}
			return strings.Join(parts, "\n"), nil
		},
	})
	tools.Register(tools.Func{
		Group: "html", Name: "attr", Title: "HTML attribute values",
		Doc:   "Return the value of an attribute for every element matching a CSS selector, in document order; elements without the attribute are skipped. Typical uses are href of links and src of images.",
		Terms: []string{"get attribute", "extract links", "href", "src", "image urls", "get href", "collect links"},
		Params: []tools.Param{
			tools.Req("text", "str", "The HTML text."),
			tools.Req("selector", "str", "A CSS selector."),
			tools.Req("name", "str", "The attribute name, for example href."),
		},
		Returns: "list[str]",
		Example: `html.attr("<a href='/a'>x</a><a>y</a>", "a", "href")`,
		Tags:    []string{"html", "attribute", "href", "src", "links"},
		Fn: func(a tools.Args) (any, error) {
			nodes, err := run(a)
			if err != nil {
				return nil, err
			}
			out := []string{}
			for _, n := range nodes {
				if v, ok := attrOf(n, a.Str("name")); ok {
					out = append(out, v)
				}
			}
			return out, nil
		},
	})
	tools.Register(tools.Func{
		Group: "html", Name: "outer_html", Title: "HTML of elements",
		Doc:   "Return the HTML source of every element matching a CSS selector (the element itself, re-serialized by the parser).",
		Terms: []string{"outerHTML", "element source", "get markup", "html snippet"},
		Params: []tools.Param{
			tools.Req("text", "str", "The HTML text."),
			tools.Req("selector", "str", "A CSS selector."),
		},
		Returns: "list[str]",
		Example: `html.outer_html("<div><p>a</p></div>", "p")`,
		Tags:    []string{"html", "outerhtml", "markup", "source"},
		Fn: func(a tools.Args) (any, error) {
			nodes, err := run(a)
			if err != nil {
				return nil, err
			}
			out := []string{}
			for _, n := range nodes {
				var b bytes.Buffer
				if err := html.Render(&b, n); err != nil {
					return nil, err
				}
				out = append(out, b.String())
			}
			return out, nil
		},
	})
}

func run(a tools.Args) ([]*html.Node, error) {
	list, err := parseSelector(a.Str("selector"))
	if err != nil {
		return nil, err
	}
	root, err := parseHTML(a.Str("text"), false)
	if err != nil {
		return nil, err
	}
	return selectAll(root, list, int(a.Int("limit"))), nil
}

func parseHTML(text string, fragment bool) (*html.Node, error) {
	if len(text) > maxInput {
		return nil, fmt.Errorf("html: input is larger than %d MB", maxInput>>20)
	}
	if !fragment {
		return html.Parse(strings.NewReader(text))
	}
	ctx := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(text), ctx)
	if err != nil {
		return nil, err
	}
	root := &html.Node{Type: html.DocumentNode}
	for _, n := range nodes {
		root.AppendChild(n)
	}
	return root, nil
}

func toDict(n *html.Node, depth int) *tools.Dict {
	d := tools.NewDict()
	d.Set("tag", n.Data)
	at := tools.NewDict()
	for _, a := range n.Attr {
		k := a.Key
		if a.Namespace != "" {
			k = a.Namespace + ":" + k
		}
		if _, dup := at.Get(k); !dup {
			at.Set(k, a.Val)
		}
	}
	d.Set("attrs", at)
	var own []string
	kids := []any{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			own = append(own, c.Data)
		case html.ElementNode:
			if depth < maxDepth {
				kids = append(kids, toDict(c, depth+1))
			}
		}
	}
	d.Set("text", collapse(strings.Join(own, " ")))
	d.Set("children", kids)
	return d
}

var blockTags = map[string]bool{
	"p": true, "div": true, "br": true, "li": true, "ul": true, "ol": true, "tr": true, "table": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "section": true, "article": true,
	"header": true, "footer": true, "nav": true, "aside": true, "main": true, "pre": true, "blockquote": true,
	"hr": true, "form": true, "dl": true, "dt": true, "dd": true, "figure": true, "figcaption": true, "address": true,
	"details": true, "summary": true, "fieldset": true, "title": true,
}

func visibleText(n *html.Node, b *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		b.WriteString(n.Data)
		return
	case html.ElementNode:
		switch n.Data {
		case "script", "style", "template", "noscript", "head":
			return
		}
		if n.Data == "br" || n.Data == "hr" {
			b.WriteString("\n")
			return
		}
		if blockTags[n.Data] {
			b.WriteString("\n")
			defer b.WriteString("\n")
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visibleText(c, b)
	}
}

// tidy collapses spaces inside lines and removes blank lines.
func tidy(s string) string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = collapse(l); l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}
