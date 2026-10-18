// Package extract registers the "extract" tool group: pull emails, URLs, IPs,
// numbers, dates, hashes, identifiers and other entities out of free text.
// Every extractor is a heuristic over plain text and never touches the network.
package extract

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxText      = 4 << 20
	defaultLimit = 10000
	maxLimit     = 100000
)

// hit is one extracted item: its byte span in the text and the value reported.
type hit struct {
	s, e int
	v    string
}

type finder func(text string, a tools.Args) []hit

// spec describes one extractor; init registers each as extract.<name>.
type spec struct {
	name     string
	title    string
	doc      string
	example  string
	tags     []string
	keywords []string
	order    int
	extra    []tools.Param
	find     finder
	noAll    bool // excluded from all() and summary() (structural or noisy kinds)
}

var specs []spec

func opt(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

var (
	pText   = req("text", "str", "Text to search.")
	pUnique = opt("unique", "bool", "Drop duplicates, keeping first-seen order.", true)
	pSort   = opt("sort", "bool", "Sort the result alphabetically.", false)
	pLimit  = opt("limit", "int", "Maximum number of results (default 10000, max 100000).", int64(defaultLimit))
	pLower  = opt("lower", "bool", "Convert results to lower case.", false)
)

func commonParams(extra []tools.Param) []tools.Param {
	return append([]tools.Param{pText, pUnique, pSort, pLimit}, extra...)
}

func checkText(t string) error {
	if len(t) > maxText {
		return fmt.Errorf("text too large (max 4 MiB)")
	}
	return nil
}

func clampLimit(n int64) int {
	if n <= 0 {
		return defaultLimit
	}
	if n > maxLimit {
		return maxLimit
	}
	return int(n)
}

// finish applies unique, sort and limit to hits in text order.
func finish(hs []hit, unique, doSort bool, limit int) []string {
	sort.SliceStable(hs, func(i, j int) bool { return hs[i].s < hs[j].s })
	out := make([]string, 0, len(hs))
	seen := map[string]bool{}
	for _, h := range hs {
		if unique {
			if seen[h.v] {
				continue
			}
			seen[h.v] = true
		}
		out = append(out, h.v)
	}
	if doSort {
		sort.Strings(out)
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func run(sp spec, text string, a tools.Args) ([]hit, error) {
	if err := checkText(text); err != nil {
		return nil, err
	}
	return sp.find(text, a), nil
}

func lookup(name string) (spec, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.NewReplacer(" ", "_", "-", "_").Replace(name)
	alias := map[string]string{
		"email": "emails", "url": "urls", "link": "urls", "links": "urls", "domain": "domains", "ip": "ips", "ip_address": "ips",
		"ip_addresses": "ips", "phone": "phone_numbers", "phones": "phone_numbers", "phone_number": "phone_numbers", "mac": "mac_addresses",
		"number": "numbers", "integer": "integers", "decimal": "decimals", "percent": "percentages", "percentage": "percentages",
		"currency": "currency_amounts", "money": "currency_amounts", "hashtag": "hashtags", "mention": "mentions", "uuid": "uuids",
		"hash": "hashes", "date": "dates", "time": "times", "datetime": "datetimes", "path": "file_paths", "paths": "file_paths",
		"word": "words", "sentence": "sentences", "paragraph": "paragraphs", "version": "versions", "semver": "versions",
		"emoji": "emojis", "acronym": "acronyms", "credit_card": "credit_cards", "iban": "ibans", "crypto": "crypto_addresses",
		"env_var": "env_vars", "placeholder": "placeholders", "json": "json_snippets", "color": "hex_colors", "colors": "hex_colors",
		"hex_color": "hex_colors", "rgb_color": "rgb_colors", "base64": "base64_strings", "jwt": "jwt_tokens", "port": "ports",
		"cidr": "cidrs", "tld": "tlds", "commit": "git_commit_hashes", "commits": "git_commit_hashes", "filename": "file_names",
		"filenames": "file_names", "extension": "file_extensions", "extensions": "file_extensions", "quote": "quoted_strings",
		"quotes": "quoted_strings", "ssns": "ssn", "tag": "html_tags", "tags": "html_tags", "image": "images_html", "images": "images_html",
	}
	if a, ok := alias[name]; ok {
		name = a
	}
	for _, s := range specs {
		if s.name == name {
			return s, true
		}
	}
	return spec{}, false
}

func kindNames() []string {
	out := make([]string, 0, len(specs))
	for _, s := range specs {
		out = append(out, s.name)
	}
	sort.Strings(out)
	return out
}

// defaults builds an Args with every extra parameter at its default value.
func defaults(sp spec) tools.Args {
	a := tools.Args{}
	for _, p := range sp.extra {
		if p.Default != nil {
			a[p.Name] = p.Default
		}
	}
	return a
}

// ---------------------------------------------------------------- text helpers

func prevRune(t string, i int) rune {
	if i <= 0 {
		return 0
	}
	r, _ := utf8.DecodeLastRuneInString(t[:i])
	return r
}

func nextRune(t string, i int) rune {
	if i >= len(t) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t[i:])
	return r
}

func isWordRune(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

func isAlnum(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// trimTrailing removes sentence punctuation (and unbalanced closing brackets)
// from the end of s.
func trimTrailing(s string) string {
	for len(s) > 0 {
		c := s[len(s)-1]
		switch {
		case strings.IndexByte(".,;:!?'\"", c) >= 0:
			s = s[:len(s)-1]
		case c == ')' && strings.Count(s, ")") > strings.Count(s, "("):
			s = s[:len(s)-1]
		case c == ']' && strings.Count(s, "]") > strings.Count(s, "["):
			s = s[:len(s)-1]
		case c == '}' && strings.Count(s, "}") > strings.Count(s, "{"):
			s = s[:len(s)-1]
		default:
			return s
		}
	}
	return s
}

func registerSpec(sp spec) {
	kw := append([]string{"extract", "ex"}, sp.keywords...)
	tools.Register(tools.Func{
		Group: "extract", Name: sp.name, Title: sp.title, Doc: sp.doc, Params: commonParams(sp.extra), Returns: "list[str]",
		Example: sp.example, Tags: sp.tags, Family: "extract", Order: sp.order, Input: "text", Keywords: kw,
		Fn: func(a tools.Args) (any, error) {
			hs, err := run(sp, a.Str("text"), a)
			if err != nil {
				return nil, err
			}
			return finish(hs, a.Bool("unique"), a.Bool("sort"), clampLimit(a.Int("limit"))), nil
		},
	})
}

// countOf counts the hits of sp in text; unique counts distinct values only.
func countOf(sp spec, text string, unique bool) int {
	hs := sp.find(text, defaults(sp))
	if !unique {
		return len(hs)
	}
	seen := map[string]bool{}
	for _, h := range hs {
		seen[h.v] = true
	}
	return len(seen)
}

// checkAggregate limits the multi-extractor functions, which run every extractor.
func checkAggregate(t string) error {
	if len(t) > 2<<20 {
		return fmt.Errorf("text too large for this function (max 2 MiB)")
	}
	return nil
}
