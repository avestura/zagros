// Package gen renders the tools registry as .pyi stubs and Markdown docs.
// cmd/toolsgen writes them to stubs/zagros and docs/extensions; a test keeps
// the committed files equal to this output.
package gen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// GroupTitle is the human name of each group.
var GroupTitle = map[string]string{
	"text":    "Text",
	"lines":   "Lines",
	"num":     "Numbers and bases",
	"math":    "Math and statistics",
	"encode":  "Encoding and escaping",
	"hash":    "Hashes",
	"cipher":  "Encryption",
	"rand":    "Random generation",
	"uuid":    "UUIDs and IDs",
	"regex":   "Regular expressions",
	"net":     "Networking",
	"dt":      "Date and time",
	"extract": "Extractors",
}

// GroupBlurb is one sentence describing each group.
var GroupBlurb = map[string]string{
	"text":    "Change, measure and analyse strings: case styles, palindromes, lengths, normalisation, similarity, Unicode.",
	"lines":   "Work on multi-line text: sort, unique, filter, join, columns, CSV and list conversions.",
	"num":     "Base conversion, the hex and binary calculator, bit operations, IEEE 754, roman numerals, number words, primes and checksums.",
	"math":    "An expression evaluator, percentages, statistics, algebra, geometry and finance helpers.",
	"encode":  "Base16/32/58/62/64/85 and more, charsets, URL, HTML, punycode, escape and unescape, JWT decode.",
	"hash":    "Digests (MD5, SHA-1/2/3, BLAKE, CRC, Adler, FNV, xxHash), HMAC, PBKDF2, bcrypt, scrypt, Argon2, TOTP.",
	"cipher":  "AES, ChaCha20, RSA, Ed25519, XOR and classic ciphers, each with a decrypt counterpart.",
	"rand":    "Random numbers, dice, strings, passwords, passphrases, colors and picks, with optional seeds.",
	"uuid":    "UUID v1-v7, ULID, KSUID, Snowflake, ObjectID, nanoid: generate, parse, convert.",
	"regex":   "Test, match, replace and explain regular expressions (Go RE2 syntax).",
	"net":     "CIDR and subnet calculators, IP and MAC tools, ports, HTTP status, MIME types, URLs and domains.",
	"dt":      "Parse, convert and calculate dates, times, time zones, durations and cron schedules.",
	"extract": "Pull emails, URLs, domains, IPs, numbers, phone numbers, dates and more out of any text.",
}

func pyType(t string) string {
	switch t {
	case "str", "int", "float", "bool", "None":
		return t
	case "list[str]":
		return "List[str]"
	case "list[float]":
		return "List[float]"
	case "dict[str, str]":
		return "Dict[str, str]"
	}
	return "Any"
}

func pyDefault(p tools.Param) string {
	if p.Required {
		return ""
	}
	switch d := p.Default.(type) {
	case nil:
		switch p.Type {
		case "str":
			return `""`
		case "int":
			return "0"
		case "float":
			return "0.0"
		case "bool":
			return "False"
		case "list[str]", "list[float]":
			return "[]"
		case "dict[str, str]":
			return "{}"
		}
		return "None"
	case string:
		return strconv.Quote(d)
	case int64:
		return strconv.FormatInt(d, 10)
	case float64:
		s := strconv.FormatFloat(d, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s
	case bool:
		if d {
			return "True"
		}
		return "False"
	case []string, []float64:
		return "[]"
	case map[string]string:
		return "{}"
	}
	return "..."
}

// Signature renders `name(a: str, b: int = 3) -> str` for docs.
func Signature(f tools.Func) string {
	var ps []string
	for _, p := range f.Params {
		s := p.Name + ": " + pyType(p.Type)
		if d := pyDefault(p); d != "" {
			s += " = " + d
		}
		ps = append(ps, s)
	}
	return fmt.Sprintf("%s(%s) -> %s", f.Name, strings.Join(ps, ", "), pyType(f.Returns))
}

func docstring(f tools.Func) string {
	var b strings.Builder
	b.WriteString(f.Doc)
	var rows []string
	for _, p := range f.Params {
		line := "    " + p.Name + ": " + p.Doc
		if len(p.Choices) > 0 {
			line += " One of: " + strings.Join(p.Choices, ", ") + "."
		}
		rows = append(rows, line)
	}
	if len(rows) > 0 {
		b.WriteString("\n\n")
		b.WriteString(strings.Join(rows, "\n"))
	}
	b.WriteString("\n\nExample: " + f.Example)
	b.WriteString("\nNeeds: none (pure function).")
	return b.String()
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = pad + l
		}
	}
	return strings.Join(lines, "\n")
}

// Stub renders stubs/zagros/<group>.pyi.
func Stub(group string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\"\"\"%s module (generated from the tools registry; do not edit by hand).\n\n%s\nNeeds: none. These are pure functions: no file, network, clipboard or environment access.\n\"\"\"\n\nfrom typing import Dict, List\n", GroupTitle[group], GroupBlurb[group])
	for _, f := range tools.All() {
		if f.Group != group {
			continue
		}
		var ps []string
		for _, p := range f.Params {
			s := p.Name + ": " + pyType(p.Type)
			if d := pyDefault(p); d != "" {
				s += " = " + d
			}
			ps = append(ps, s)
		}
		fmt.Fprintf(&b, "\ndef %s(%s) -> %s:\n    \"\"\"%s\n    \"\"\"\n", f.Name, strings.Join(ps, ", "), pyType(f.Returns), strings.ReplaceAll(indent(docstring(f), "    ")[4:], `"""`, `'''`))
	}
	return b.String()
}

// Groups lists the groups that have functions.
func Groups() []string { return tools.Groups() }

// DocsGroup renders docs/extensions/tools-<group>.md.
func DocsGroup(group string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s (`%s`)\n\n%s\n\nNo permission is needed: these are pure functions. Generated from the tools registry.\n\n", GroupTitle[group], group, GroupBlurb[group])
	n := 0
	for _, f := range tools.All() {
		if f.Group == group {
			n++
		}
	}
	fmt.Fprintf(&b, "%d functions. Use them as `%s.<name>(...)` in any extension.\n\n", n, group)
	for _, f := range tools.All() {
		if f.Group != group {
			continue
		}
		fmt.Fprintf(&b, "## %s.%s\n\n`%s.%s`\n\n%s\n\n", group, f.Name, group, Signature(f), mdText(f.Doc))
		if len(f.Params) > 0 {
			b.WriteString("| Parameter | Type | Description |\n|---|---|---|\n")
			for _, p := range f.Params {
				desc := strings.ReplaceAll(p.Doc, "|", "\\|")
				if len(p.Choices) > 0 {
					desc += " One of: " + strings.Join(p.Choices, ", ") + "."
				}
				def := ""
				if d := pyDefault(p); d != "" {
					def = " (default `" + d + "`)"
				}
				fmt.Fprintf(&b, "| `%s` | `%s`%s | %s |\n", p.Name, pyType(p.Type), def, desc)
			}
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "Returns `%s`. Example: `%s`\n\n", pyType(f.Returns), f.Example)
	}
	return b.String()
}

// DocsIndex renders docs/extensions/tools.md.
func DocsIndex() string {
	var b strings.Builder
	b.WriteString("# Built-in tools API\n\nZagros ships a large library of pure utility functions. The same functions power the launcher's tool families (Tab opens the whole family for what you typed) and are available to extensions as modules, with no permission needed.\n\n")
	b.WriteString("| Module | What it covers | Functions |\n|---|---|---|\n")
	count := map[string]int{}
	for _, f := range tools.All() {
		count[f.Group]++
	}
	groups := tools.Groups()
	sort.Strings(groups)
	for _, g := range groups {
		fmt.Fprintf(&b, "| [`%s`](tools-%s.md) | %s | %d |\n", g, g, GroupBlurb[g], count[g])
	}
	b.WriteString("\n```python\ndef on_query(ctx, query):\n    return [item(id = \"h\", title = hash.sha256(query), action = action_copy(hash.sha256(query)))]\n```\n\n")
	b.WriteString("Compose them with the host modules: read text with `clipboard.get()` (needs `clipboard.read`), transform it with these functions, and write it back with `clipboard.set(...)` (needs `clipboard.write`).\n\nThe exact signatures are in the API stubs (`zagros-stubs/zagros/<module>.pyi`); coding agents can also search them through the Zagros MCP server.\n")
	return b.String()
}

// mdText escapes the characters that would turn documentation text into
// Markdown links or HTML (doc strings talk about "[text](url)" and "<tag>").
func mdText(s string) string {
	return strings.NewReplacer("[", "\\[", "]", "\\]", "<", "&lt;", ">", "&gt;").Replace(s)
}
