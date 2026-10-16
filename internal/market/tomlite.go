package market

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// parseTOML parses the small TOML subset used by registry specs and
// manifests: tables, strings (basic, literal, multi-line basic), booleans,
// integers and arrays of scalars. Keys inside tables are flattened with dots
// ("permissions.net"). Array-of-tables headers are skipped.
func parseTOML(src string) (map[string]any, error) {
	out := map[string]any{}
	prefix := ""
	skip := false
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(stripComment(lines[i]))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[[") {
			skip = true
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			prefix = strings.TrimSpace(line[1:len(line)-1]) + "."
			skip = false
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			return nil, fmt.Errorf("toml line %d: expected key = value", i+1)
		}
		key := strings.Trim(strings.TrimSpace(line[:eq]), `"'`)
		val := strings.TrimSpace(line[eq+1:])
		// Multi-line basic string (use the raw line: comments are not stripped inside).
		if raw := strings.TrimSpace(lines[i]); strings.Contains(raw, `"""`) && strings.HasPrefix(strings.TrimSpace(raw[strings.IndexByte(raw, '=')+1:]), `"""`) {
			rest := strings.TrimSpace(raw[strings.IndexByte(raw, '=')+1:])[3:]
			var sb strings.Builder
			for {
				if j := strings.Index(rest, `"""`); j >= 0 {
					sb.WriteString(rest[:j])
					break
				}
				sb.WriteString(rest)
				sb.WriteByte('\n')
				i++
				if i >= len(lines) {
					return nil, fmt.Errorf("toml: unterminated multi-line string for %q", key)
				}
				rest = lines[i]
			}
			if !skip {
				out[prefix+key] = strings.TrimPrefix(sb.String(), "\n")
			}
			continue
		}
		// Multi-line arrays: keep appending until brackets balance.
		for strings.HasPrefix(val, "[") && bracketDepth(val) > 0 {
			i++
			if i >= len(lines) {
				return nil, fmt.Errorf("toml: unterminated array for %q", key)
			}
			val += " " + strings.TrimSpace(stripComment(lines[i]))
		}
		v, err := parseValue(val)
		if err != nil {
			return nil, fmt.Errorf("toml line %d: %w", i+1, err)
		}
		if !skip {
			out[prefix+key] = v
		}
	}
	return out, nil
}

func stripComment(s string) string {
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case q != 0:
			if c == '\\' && q == '"' {
				i++
			} else if c == q {
				q = 0
			}
		case c == '"' || c == '\'':
			q = c
		case c == '#':
			return s[:i]
		}
	}
	return s
}

func bracketDepth(s string) int {
	d := 0
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case q != 0:
			if c == '\\' && q == '"' {
				i++
			} else if c == q {
				q = 0
			}
		case c == '"' || c == '\'':
			q = c
		case c == '[':
			d++
		case c == ']':
			d--
		}
	}
	return d
}

func parseValue(s string) (any, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return nil, fmt.Errorf("empty value")
	case s[0] == '"':
		u, err := strconv.Unquote(s)
		if err != nil {
			return nil, fmt.Errorf("bad string %s", s)
		}
		return u, nil
	case s[0] == '\'':
		if len(s) < 2 || s[len(s)-1] != '\'' {
			return nil, fmt.Errorf("bad string %s", s)
		}
		return s[1 : len(s)-1], nil
	case s == "true":
		return true, nil
	case s == "false":
		return false, nil
	case s[0] == '[':
		if s[len(s)-1] != ']' {
			return nil, fmt.Errorf("bad array %s", s)
		}
		var arr []any
		for _, part := range splitTop(s[1 : len(s)-1]) {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			v, err := parseValue(part)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	case s[0] == '{':
		return s, nil // inline tables are kept raw; not needed here
	}
	if n, err := strconv.ParseInt(strings.ReplaceAll(s, "_", ""), 10, 64); err == nil {
		return n, nil
	}
	return nil, fmt.Errorf("unsupported value %q", s)
}

func splitTop(s string) []string {
	var parts []string
	depth, start := 0, 0
	var q byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case q != 0:
			if c == '\\' && q == '"' {
				i++
			} else if c == q {
				q = 0
			}
		case c == '"' || c == '\'':
			q = c
		case c == '[' || c == '{':
			depth++
		case c == ']' || c == '}':
			depth--
		case c == ',' && depth == 0:
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return append(parts, s[start:])
}

func tomlString(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

func tomlStrings(m map[string]any, k string) []string {
	arr, _ := m[k].([]any)
	var out []string
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// ParseManifest extracts the marketplace-relevant fields from manifest.toml
// text. Permissions come from a top-level `permissions = [...]` array or from
// the keys of a `[permissions]` table.
func ParseManifest(data []byte) (Manifest, error) {
	m, err := parseTOML(string(data))
	if err != nil {
		return Manifest{}, err
	}
	mf := Manifest{
		ID:          tomlString(m, "id"),
		Name:        tomlString(m, "name"),
		Version:     strings.TrimPrefix(tomlString(m, "version"), "v"),
		Description: tomlString(m, "description"),
		Author:      tomlString(m, "author"),
		Permissions: tomlStrings(m, "permissions"),
	}
	if len(mf.Permissions) == 0 {
		seen := map[string]bool{}
		for k, v := range m {
			if !strings.HasPrefix(k, "permissions.") {
				continue
			}
			name := strings.TrimPrefix(k, "permissions.")
			if i := strings.IndexByte(name, '.'); i >= 0 {
				name = name[:i]
			}
			if b, ok := v.(bool); ok && !b {
				continue
			}
			if !seen[name] {
				seen[name] = true
				mf.Permissions = append(mf.Permissions, name)
			}
		}
		sort.Strings(mf.Permissions)
	}
	if mf.ID == "" {
		return mf, fmt.Errorf("manifest.toml has no id")
	}
	return mf, nil
}
