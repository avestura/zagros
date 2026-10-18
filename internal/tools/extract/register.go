package extract

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

var phoneRe = regexp.MustCompile(`(?:\+|00)?\(?\d[\d \t().\-\x{a0}]{5,}\d`)
var (
	phoneDate = regexp.MustCompile(`^\d{4}-\d{1,2}-\d{1,2}$|^\d{1,2}[/.\-]\d{1,2}[/.\-]\d{2,4}$`)
	phoneSSN  = regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)
	phoneDots = regexp.MustCompile(`^\d+(?:\.\d+)+$`)
)

func phoneHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range phoneRe.FindAllStringIndex(text, -1) {
		s, e := m[0], m[1]
		v := text[s:e]
		p, n := prevRune(text, s), nextRune(text, e)
		if isAlnum(p) || p == '.' || p == '/' || p == ':' || p == '-' || p == '_' || p == '@' || isAlnum(n) || n == '-' || n == '_' {
			continue
		}
		d := onlyDigits(v)
		if len(d) < 7 || len(d) > 15 {
			continue
		}
		if phoneDate.MatchString(v) || phoneSSN.MatchString(v) || phoneDots.MatchString(v) && !strings.HasPrefix(v, "+") {
			continue
		}
		hasPlus := strings.HasPrefix(v, "+") || strings.HasPrefix(v, "00")
		ok := hasPlus || strings.ContainsAny(v, "()-") || strings.Count(v, " ")+strings.Count(v, " ") >= 2 && len(d) >= 9
		if !ok {
			continue
		}
		out = append(out, hit{s, e, strings.TrimSpace(v)})
	}
	return out
}

func phoneSpecs() []spec {
	return []spec{{name: "phone_numbers", title: "Extract phone numbers", order: 7, keywords: []string{"phones"},
		doc:     "Finds international and national phone numbers: optional + or 00, digits grouped by spaces, dashes, dots or parentheses, 7 to 15 digits in total. Heuristic: plain digit runs without + or separators, ISO dates, SSNs, IPs and dotted version numbers are skipped, while long dash-separated ids may still match.",
		example: `extract.phone_numbers("Call +1 (555) 123-4567 or 020 7946 0958")`,
		tags:    []string{"phone", "telephone", "mobile", "cell", "tel", "contact numbers", "e164", "international"}, find: phoneHits}}
}

func init() {
	specs = append(specs, netSpecs()...)
	specs = append(specs, textSpecs()...)
	specs = append(specs, idSpecs()...)
	specs = append(specs, phoneSpecs()...)
	for i := range specs {
		registerSpec(specs[i])
	}
	registerExtras()
}

func counts(text string, unique bool) map[string]string {
	out := map[string]string{}
	for _, sp := range specs {
		if sp.noAll {
			continue
		}
		if n := countOf(sp, text, unique); n > 0 {
			out[sp.name] = strconv.Itoa(n)
		}
	}
	return out
}

// regAPI registers an API-only function (not listed in the launcher).
func regAPI(f tools.Func) {
	f.Group = "extract"
	tools.Register(f)
}

func rowReg(f tools.Func) {
	f.Group = "extract"
	f.Family = "extract"
	if f.Input == "" {
		f.Input = "text"
	}
	f.Keywords = append([]string{"extract", "ex"}, f.Keywords...)
	tools.Register(f)
}

func joinVals(sp spec, text string, limit int) string {
	return strings.Join(finish(sp.find(text, defaults(sp)), true, false, limit), ", ")
}

type span struct {
	s, e int
	kind string
}

func redactText(text string, kinds []string, mask string) (string, error) {
	if len(kinds) == 0 {
		kinds = []string{"emails", "ips", "phone_numbers", "credit_cards"}
	}
	var spans []span
	for _, k := range kinds {
		sp, ok := lookup(k)
		if !ok {
			return "", fmt.Errorf("unknown kind %q; known: %s", k, strings.Join(kindNames(), ", "))
		}
		for _, h := range sp.find(text, defaults(sp)) {
			spans = append(spans, span{h.s, h.e, sp.name})
		}
	}
	sort.SliceStable(spans, func(i, j int) bool {
		if spans[i].s != spans[j].s {
			return spans[i].s < spans[j].s
		}
		return spans[i].e > spans[j].e
	})
	var b strings.Builder
	last := 0
	for _, sp := range spans {
		if sp.s < last {
			continue
		}
		b.WriteString(text[last:sp.s])
		switch {
		case utf8.RuneCountInString(mask) == 1:
			b.WriteString(strings.Repeat(mask, utf8.RuneCountInString(text[sp.s:sp.e])))
		default:
			b.WriteString(strings.ReplaceAll(mask, "{kind}", sp.kind))
		}
		last = sp.e
	}
	b.WriteString(text[last:])
	return b.String(), nil
}

func maskCards(text, mask string, keep int) string {
	if mask == "" {
		mask = "*"
	}
	hs := cardHits(text, nil)
	var b strings.Builder
	last := 0
	for _, h := range hs {
		b.WriteString(text[last:h.s])
		digits := len(onlyDigits(text[h.s:h.e]))
		seen := 0
		for i := h.s; i < h.e; i++ {
			c := text[i]
			if c >= '0' && c <= '9' {
				seen++
				if seen <= digits-keep {
					b.WriteString(mask)
					continue
				}
			}
			b.WriteByte(c)
		}
		last = h.e
	}
	b.WriteString(text[last:])
	return b.String()
}

func registerExtras() {
	rowReg(tools.Func{
		Name: "all", Title: "Extract everything (counts)", Order: 1,
		Doc:     "Headline summary of what a text contains: a dict mapping each extractor name (emails, urls, ips, numbers, dates, hashes...) to the number of matches found, listing only non-zero kinds. Structural kinds (words, sentences, paragraphs, integers, decimals, tlds, html tags...) are left out; use extract.summary for those.",
		Params:  []tools.Param{pText},
		Returns: "dict[str, str]", Example: `extract.all("Mail bob@example.com or call +1 555 123 4567, see https://example.com")`,
		Tags: []string{"everything", "overview", "find all", "entities", "analyze text", "counts", "inventory"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkAggregate(a.Str("text")); err != nil {
				return nil, err
			}
			return counts(a.Str("text"), false), nil
		},
	})
	rowReg(tools.Func{
		Name: "entities", Title: "Extract key entities", Order: 8,
		Doc:     "The most useful entities of a text as a dict of kind to a comma-separated list of unique values: emails, urls, domains, ips, phone_numbers, hashtags, mentions and dates. Kinds with no match are omitted; each list is capped at 50 values.",
		Params:  []tools.Param{pText},
		Returns: "dict[str, str]", Example: `extract.entities("Contact bob@example.com #news @alice 2026-10-17")`,
		Tags: []string{"named entities", "contacts", "ner", "key info", "overview"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			out := map[string]string{}
			for _, k := range []string{"emails", "urls", "domains", "ips", "phone_numbers", "hashtags", "mentions", "dates"} {
				sp, _ := lookup(k)
				if v := joinVals(sp, a.Str("text"), 50); v != "" {
					out[k] = v
				}
			}
			return out, nil
		},
	})
	rowReg(tools.Func{
		Name: "summary", Title: "Text summary", Order: 9,
		Doc:     "Text statistics (characters, bytes, lines, words, sentences, paragraphs) plus the number of matches of every extractor with at least one match, as a dict of strings.",
		Params:  []tools.Param{pText},
		Returns: "dict[str, str]", Example: `extract.summary("Hello world. Mail me at a@b.co")`,
		Tags: []string{"stats", "statistics", "word count", "character count", "line count", "analyze", "profile"},
		Fn: func(a tools.Args) (any, error) {
			t := a.Str("text")
			if err := checkAggregate(t); err != nil {
				return nil, err
			}
			out := counts(t, false)
			lines := 0
			if t != "" {
				lines = strings.Count(t, "\n") + 1
				if strings.HasSuffix(t, "\n") {
					lines--
				}
			}
			out["characters"] = strconv.Itoa(utf8.RuneCountInString(t))
			out["bytes"] = strconv.Itoa(len(t))
			out["lines"] = strconv.Itoa(lines)
			out["words"] = strconv.Itoa(len(wordRe.FindAllStringIndex(t, -1)))
			out["sentences"] = strconv.Itoa(len(sentenceHits(t, nil)))
			out["paragraphs"] = strconv.Itoa(len(paragraphHits(t, nil)))
			return out, nil
		},
	})
	rowReg(tools.Func{
		Name: "redact", Title: "Redact sensitive data", Order: 10,
		Doc:     "Replaces matches of the chosen kinds with a mask. kinds defaults to emails, ips, phone_numbers and credit_cards; any extractor name works (see extract.kinds). A one-character mask is repeated to the match length (\"*\"), otherwise the mask is inserted as is and {kind} is replaced by the kind name. Overlapping matches are replaced once.",
		Params:  []tools.Param{pText, opt("kinds", "list[str]", "Extractor names to redact; empty uses the default four.", []string(nil)), opt("mask", "str", `Replacement text, default "[REDACTED]"; "*" masks character by character; "{kind}" is substituted.`, "[REDACTED]")},
		Returns: "str", Example: `extract.redact("Mail bob@example.com from 10.0.0.1", ["emails", "ips"], "<{kind}>")`,
		Tags: []string{"mask", "anonymize", "censor", "pii", "scrub", "sanitize", "privacy", "hide"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			return redactText(a.Str("text"), a.List("kinds"), a.Str("mask"))
		},
	})
	rowReg(tools.Func{
		Name: "mask_credit_cards", Title: "Mask credit card numbers", Order: 11,
		Doc:     "Masks Luhn-valid 13-19 digit card numbers found in the text, keeping separators and the last keep_last digits (default 4): 4111 1111 1111 1111 becomes **** **** **** 1111.",
		Params:  []tools.Param{pText, opt("mask", "str", "Mask character(s) for each hidden digit.", "*"), opt("keep_last", "int", "Digits left visible at the end (0-8).", int64(4))},
		Returns: "str", Example: `extract.mask_credit_cards("card 4111 1111 1111 1111")`,
		Tags: []string{"pan", "pci", "hide card", "redact card", "obfuscate", "payment"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			k := int(a.Int("keep_last"))
			if k < 0 || k > 8 {
				return nil, fmt.Errorf("keep_last must be between 0 and 8")
			}
			return maskCards(a.Str("text"), a.Str("mask"), k), nil
		},
	})
	rowReg(tools.Func{
		Name: "count", Title: "Count matches of a kind", Order: 12,
		Doc:     "Counts how many items of one kind (see extract.kinds, e.g. emails, urls, ips) appear in the text. By default every occurrence counts; with unique true only distinct values.",
		Params:  []tools.Param{pText, req("kind", "str", "Extractor name such as emails, urls or ips."), opt("unique", "bool", "Count distinct values only.", false)},
		Returns: "int", Example: `extract.count("a@b.co a@b.co c@d.org", "emails")`,
		Tags: []string{"how many", "number of", "occurrences", "tally"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			sp, ok := lookup(a.Str("kind"))
			if !ok {
				return nil, fmt.Errorf("unknown kind %q; known: %s", a.Str("kind"), strings.Join(kindNames(), ", "))
			}
			return int64(countOf(sp, a.Str("text"), a.Bool("unique"))), nil
		},
	})
	rowReg(tools.Func{
		Name: "extract", Title: "Extract by kind", Order: 13,
		Doc:     "Generic dispatcher: runs the extractor named kind (see extract.kinds) with default options and returns its list. Equivalent to calling extract.<kind>(text); unknown kinds fail with the list of valid names.",
		Params:  []tools.Param{req("kind", "str", "Extractor name such as emails, urls, ips, dates."), pText, pUnique, pSort},
		Returns: "list[str]", Example: `extract.extract("emails", "write to bob@example.com")`,
		Tags: []string{"dispatch", "generic", "by name", "any kind"},
		Fn: func(a tools.Args) (any, error) {
			sp, ok := lookup(a.Str("kind"))
			if !ok {
				return nil, fmt.Errorf("unknown kind %q; known: %s", a.Str("kind"), strings.Join(kindNames(), ", "))
			}
			hs, err := run(sp, a.Str("text"), defaults(sp))
			if err != nil {
				return nil, err
			}
			return finish(hs, a.Bool("unique"), a.Bool("sort"), defaultLimit), nil
		},
	})
	regAPI(tools.Func{
		Name: "kinds", Title: "List extractor kinds", Order: 14, Input: "filter",
		Doc:     "Lists the names of all extractors usable with extract.extract, extract.count and extract.redact, optionally filtered by a substring.",
		Params:  []tools.Param{opt("filter", "str", "Only kinds containing this text.", "")},
		Returns: "list[str]", Example: `extract.kinds("ip")`,
		Tags: []string{"available", "list extractors", "help", "names"},
		Fn: func(a tools.Args) (any, error) {
			f := strings.ToLower(strings.TrimSpace(a.Str("filter")))
			out := []string{}
			for _, k := range kindNames() {
				if strings.Contains(k, f) {
					out = append(out, k)
				}
			}
			return out, nil
		},
	})
	rowReg(tools.Func{
		Name: "lines_containing", Title: "Lines containing text", Order: 15,
		Doc:     "Returns the lines of the text that contain a literal substring (grep -F). ignore_case compares case-insensitively and invert returns the lines that do not contain it. An empty substring matches every line.",
		Params:  []tools.Param{pText, req("substr", "str", "Literal text to look for."), opt("ignore_case", "bool", "Case-insensitive match.", false), opt("invert", "bool", "Return the non-matching lines instead.", false), pLimit},
		Returns: "list[str]", Example: `extract.lines_containing("ok\nERROR x\nok", "error", True)`,
		Tags: []string{"grep", "find lines", "filter lines", "search", "contains", "findstr"},
		Fn: func(a tools.Args) (any, error) {
			t := a.Str("text")
			if err := checkText(t); err != nil {
				return nil, err
			}
			sub := a.Str("substr")
			if a.Bool("ignore_case") {
				sub = strings.ToLower(sub)
			}
			lim := clampLimit(a.Int("limit"))
			out := []string{}
			for _, ln := range strings.Split(strings.ReplaceAll(t, "\r\n", "\n"), "\n") {
				cmp := ln
				if a.Bool("ignore_case") {
					cmp = strings.ToLower(ln)
				}
				if strings.Contains(cmp, sub) != a.Bool("invert") {
					out = append(out, ln)
					if len(out) >= lim {
						break
					}
				}
			}
			return out, nil
		},
	})
}
