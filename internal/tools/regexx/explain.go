package regexx

import (
	"fmt"
	"strings"
)

// explain describes a pattern token by token. It is a lexer over the pattern
// text, not a full parser, so it stays useful for slightly invalid patterns.
func explain(p string) []string {
	rs := []rune(p)
	var out []string
	depth := 0
	emit := func(tok, desc string) {
		out = append(out, strings.Repeat("  ", depth)+tok+"  -  "+desc)
	}
	if len(rs) == 0 {
		return []string{"(empty pattern)  -  matches the empty string at every position"}
	}
	i := 0
	for i < len(rs) {
		c := rs[i]
		switch c {
		case '^':
			emit("^", "start of text (start of a line with the m flag)")
			i++
		case '$':
			emit("$", "end of text (end of a line with the m flag)")
			i++
		case '.':
			emit(".", "any character except a newline (any character with the s flag)")
			i++
		case '|':
			emit("|", "OR - either what comes before or what comes after")
			i++
		case '(':
			j := i + 1
			switch {
			case hasAt(rs, j, "?:"):
				emit("(?:", "start non-capturing group")
				i = j + 2
			case hasAt(rs, j, "?="), hasAt(rs, j, "?!"), hasAt(rs, j, "?<="), hasAt(rs, j, "?<!"):
				n := 2
				if hasAt(rs, j, "?<=") || hasAt(rs, j, "?<!") {
					n = 3
				}
				emit("("+string(rs[j:j+n]), "lookaround - not supported by RE2")
				i = j + n
			case hasAt(rs, j, "?P<"), hasAt(rs, j, "?<"):
				k := j + 2
				if rs[k] == '<' {
					k++
				}
				e := k
				for e < len(rs) && rs[e] != '>' {
					e++
				}
				emit(string(rs[i:min(e+1, len(rs))]), fmt.Sprintf("start capture group named %q", string(rs[k:min(e, len(rs))])))
				i = min(e+1, len(rs))
			case hasAt(rs, j, "?"):
				e := j + 1
				for e < len(rs) && rs[e] != ')' && rs[e] != ':' {
					e++
				}
				flags := string(rs[j+1 : min(e, len(rs))])
				if e < len(rs) && rs[e] == ':' {
					emit("(?"+flags+":", "start non-capturing group with flags "+flagWords(flags))
					i = e + 1
				} else {
					emit("(?"+flags+")", "set flags "+flagWords(flags)+" for the rest of the group")
					i = min(e+1, len(rs))
					continue
				}
			default:
				emit("(", "start capture group")
				i++
			}
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
			emit(")", "end group")
			i++
		case '[':
			end, desc := parseClass(rs, i)
			emit(string(rs[i:end]), desc)
			i = end
		case '\\':
			end, d := parseEscape(rs, i)
			emit(string(rs[i:end]), d.text)
			i = end
		case '*', '+', '?', '{':
			end, desc, ok := parseQuant(rs, i)
			if !ok {
				emit(string(c), "literal "+string(c))
				i++
				break
			}
			emit(string(rs[i:end]), desc)
			i = end
		default:
			j := i
			for j < len(rs) && !strings.ContainsRune(`^$.|()[\*+?{`, rs[j]) {
				j++
			}
			if j < len(rs) && strings.ContainsRune("*+?{", rs[j]) && j-i > 1 {
				if _, _, ok := parseQuant(rs, j); ok {
					j--
				}
			}
			lit := string(rs[i:j])
			if j-i == 1 {
				emit(lit, "literal "+quote(lit))
			} else {
				emit(lit, "literal text "+quote(lit))
			}
			i = j
		}
	}
	if depth > 0 {
		out = append(out, "(warning: unclosed group)")
	}
	return out
}

func quote(s string) string { return `"` + s + `"` }

func hasAt(rs []rune, i int, s string) bool {
	r := []rune(s)
	if i+len(r) > len(rs) {
		return false
	}
	for k := range r {
		if rs[i+k] != r[k] {
			return false
		}
	}
	return true
}

func flagWords(f string) string {
	var parts []string
	neg := false
	for _, c := range f {
		name := ""
		switch c {
		case 'i':
			name = "case-insensitive"
		case 'm':
			name = "multiline"
		case 's':
			name = "dot matches newline"
		case 'U':
			name = "ungreedy"
		case '-':
			neg = true
			continue
		default:
			name = string(c)
		}
		if neg {
			name = "not " + name
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

type escInfo struct {
	tok  string
	text string
}

// parseEscape reads an escape sequence starting at rs[i] == '\\'.
func parseEscape(rs []rune, i int) (int, escInfo) {
	if i+1 >= len(rs) {
		return i + 1, escInfo{`\`, "dangling backslash (invalid)"}
	}
	n := rs[i+1]
	tok := `\` + string(n)
	simple := map[rune]string{
		'd': "a digit 0-9", 'D': "any character that is not a digit",
		'w': "a word character (letter, digit or underscore)", 'W': "any character that is not a word character",
		's': "whitespace (space, tab, newline...)", 'S': "any character that is not whitespace",
		'b': "word boundary", 'B': "not a word boundary",
		'A': "start of text", 'z': "end of text", 'Z': "end of text",
		'n': "newline", 't': "tab", 'r': "carriage return", 'f': "form feed", 'v': "vertical tab", 'a': "bell",
	}
	if d, ok := simple[n]; ok {
		return i + 2, escInfo{tok, d}
	}
	switch {
	case n == 'p' || n == 'P':
		neg := ""
		if n == 'P' {
			neg = "not "
		}
		if i+2 < len(rs) && rs[i+2] == '{' {
			e := i + 3
			for e < len(rs) && rs[e] != '}' {
				e++
			}
			if e < len(rs) {
				name := string(rs[i+3 : e])
				return e + 1, escInfo{string(rs[i : e+1]), "a character " + neg + "in Unicode class " + name}
			}
		}
		if i+2 < len(rs) {
			return i + 3, escInfo{string(rs[i : i+3]), "a character " + neg + "in Unicode class " + string(rs[i+2])}
		}
	case n == 'x':
		if i+2 < len(rs) && rs[i+2] == '{' {
			e := i + 3
			for e < len(rs) && rs[e] != '}' {
				e++
			}
			if e < len(rs) {
				return e + 1, escInfo{string(rs[i : e+1]), "character with hex code " + string(rs[i+3:e])}
			}
		}
		if i+3 < len(rs) {
			return i + 4, escInfo{string(rs[i : i+4]), "character with hex code " + string(rs[i+2:i+4])}
		}
	case n == 'Q':
		e := i + 2
		for e < len(rs) && !(rs[e] == '\\' && e+1 < len(rs) && rs[e+1] == 'E') {
			e++
		}
		lit := string(rs[i+2 : min(e, len(rs))])
		end := min(e+2, len(rs))
		return end, escInfo{string(rs[i:end]), "literal text " + quote(lit)}
	case n >= '1' && n <= '9':
		return i + 2, escInfo{tok, "backreference - not supported by RE2"}
	}
	return i + 2, escInfo{tok, "literal " + quote(string(n))}
}

func parseClass(rs []rune, i int) (int, string) {
	j := i + 1
	neg := false
	if j < len(rs) && rs[j] == '^' {
		neg = true
		j++
	}
	start := j
	if j < len(rs) && rs[j] == ']' {
		j++
	}
	for j < len(rs) && rs[j] != ']' {
		if rs[j] == '[' && j+1 < len(rs) && rs[j+1] == ':' {
			if e := strings.Index(string(rs[j:]), ":]"); e >= 0 {
				j += len([]rune(string(rs[j:])[:e+2]))
				continue
			}
		}
		if rs[j] == '\\' {
			j++
		}
		j++
	}
	if j >= len(rs) {
		return len(rs), "unterminated character class (invalid)"
	}
	body := rs[start:j]
	parts := describeClass(body)
	what := "any one of: "
	if neg {
		what = "any one character except: "
	}
	return j + 1, what + strings.Join(parts, ", ")
}

func describeClass(b []rune) []string {
	var parts []string
	for k := 0; k < len(b); k++ {
		c := b[k]
		switch {
		case c == '[' && k+1 < len(b) && b[k+1] == ':':
			e := strings.Index(string(b[k:]), ":]")
			if e > 0 {
				name := string(b[k+2:])[:e-2]
				parts = append(parts, "POSIX class "+name)
				k += len([]rune(string(b[k:])[:e+2])) - 1
				continue
			}
			parts = append(parts, "[")
		case c == '\\' && k+1 < len(b):
			e, d := parseEscape(b, k)
			parts = append(parts, d.text)
			k = e - 1
		case k+2 < len(b) && b[k+1] == '-':
			parts = append(parts, fmt.Sprintf("%c to %c", c, b[k+2]))
			k += 2
		default:
			parts = append(parts, quote(string(c)))
		}
	}
	return parts
}

func parseQuant(rs []rune, i int) (int, string, bool) {
	var desc string
	end := i + 1
	switch rs[i] {
	case '*':
		desc = "zero or more times"
	case '+':
		desc = "one or more times"
	case '?':
		desc = "optional (zero or one time)"
	case '{':
		e := i + 1
		for e < len(rs) && rs[e] != '}' {
			if !(rs[e] >= '0' && rs[e] <= '9') && rs[e] != ',' {
				return 0, "", false
			}
			e++
		}
		if e >= len(rs) || e == i+1 {
			return 0, "", false
		}
		body := string(rs[i+1 : e])
		end = e + 1
		switch {
		case !strings.Contains(body, ","):
			desc = "exactly " + body + " times"
		case strings.HasSuffix(body, ","):
			desc = "at least " + strings.TrimSuffix(body, ",") + " times"
		default:
			lh := strings.SplitN(body, ",", 2)
			desc = "between " + lh[0] + " and " + lh[1] + " times"
		}
	}
	if end < len(rs) && rs[end] == '?' {
		end++
		desc += ", as few as possible (lazy)"
	} else {
		desc += ", as many as possible (greedy)"
	}
	return end, "repeat the previous item " + desc, true
}
