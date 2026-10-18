// Package urlx registers the "url" tool group: parse, build, join and
// normalize URLs and query strings as structured values. The net group has
// string-in string-out URL helpers for the launcher; this group returns dicts
// and takes dicts. Pure: no DNS, no network.
package urlx

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const maxLen = 1 << 20

var defaultPorts = map[string]string{
	"http": "80", "https": "443", "ftp": "21", "ws": "80", "wss": "443", "ssh": "22", "sftp": "22",
	"smtp": "25", "imap": "143", "imaps": "993", "pop3": "110", "pop3s": "995", "ldap": "389", "ldaps": "636",
	"postgres": "5432", "postgresql": "5432", "mysql": "3306", "redis": "6379", "mongodb": "27017",
}

func init() {
	tools.Register(tools.Func{
		Group: "url", Name: "parse", Title: "Parse URL into parts",
		Doc:   "Split a URL into a dict with scheme, user, password, host (with port), hostname, port, path, query (dict of key to list of values), query_string, fragment and origin. Missing parts are empty strings. With default_scheme set, a URL without a scheme such as example.com/a gets that scheme first. Raises on whitespace, control characters or malformed escapes.",
		Terms: []string{"split url", "url parts", "url to dict", "query string", "hostname from url", "get domain", "parse link", "url components"},
		Params: []tools.Param{
			tools.Req("url", "str", "The URL."),
			tools.Opt("default_scheme", "str", "Scheme to assume when the URL has none, for example https.", ""),
		},
		Returns: "any",
		Example: `url.parse("https://u:p@example.com:8080/a/b?x=1&x=2#top")`,
		Tags:    []string{"url", "parse", "split", "components", "hostname", "query"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseURL(a.Str("url"), a.Str("default_scheme"))
			if err != nil {
				return nil, err
			}
			return toDict(u), nil
		},
	})
	tools.Register(tools.Func{
		Group: "url", Name: "build", Title: "Build URL from parts",
		Doc:     "Assemble a URL from a dict with any of scheme, user, password, host or hostname plus port, path, query and fragment. query is a dict (values may be strings, numbers, bools or lists of them; None is skipped) or a ready query string. Parts are escaped as needed; a path gets a leading slash when there is a host. Raises when a part has the wrong type or the port is not a number.",
		Terms:   []string{"make url", "construct url", "dict to url", "assemble url", "add query params", "compose link"},
		Params:  []tools.Param{tools.Req("parts", "any", "A dict of URL parts.")},
		Returns: "str",
		Example: `url.build({"scheme": "https", "host": "example.com", "path": "/search", "query": {"q": "go lang", "page": 2}})`,
		Tags:    []string{"url", "build", "construct", "compose"},
		Fn: func(a tools.Args) (any, error) {
			d, ok := a["parts"].(*tools.Dict)
			if !ok {
				return nil, fmt.Errorf("parts must be a dict")
			}
			return build(d)
		},
	})
	tools.Register(tools.Func{
		Group: "url", Name: "join", Title: "Resolve a relative URL",
		Doc:   "Resolve ref against base the way a browser resolves a link: ../x, /root, ?q=1, #frag, //other.host/x and absolute URLs all work, and dot segments are removed. Raises when either side is malformed.",
		Terms: []string{"relative url", "absolute url", "resolve link", "combine url", "urljoin", "href to absolute"},
		Params: []tools.Param{
			tools.Req("base", "str", "The base URL."),
			tools.Req("ref", "str", "The relative or absolute reference."),
		},
		Returns: "str",
		Example: `url.join("https://example.com/a/b/c", "../d?x=1")`,
		Tags:    []string{"url", "join", "resolve", "relative", "absolute"},
		Fn: func(a tools.Args) (any, error) {
			b, err := parseURL(a.Str("base"), "")
			if err != nil {
				return nil, fmt.Errorf("base: %w", err)
			}
			r, err := parseURL(a.Str("ref"), "")
			if err != nil {
				return nil, fmt.Errorf("ref: %w", err)
			}
			return b.ResolveReference(r).String(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "url", Name: "query_parse", Title: "Parse a query string",
		Doc:   "Parse a query string (a leading ? is ignored) into a dict in order of first appearance. Values are percent-decoded and + becomes a space. By default a repeated key keeps its last value as a string; with multi=True every value is a list. A key without = has an empty value. Raises on a malformed percent escape.",
		Terms: []string{"get params", "query params to dict", "parse query", "querystring", "url parameters", "search params"},
		Params: []tools.Param{
			tools.Req("query", "str", "The query string, with or without the leading ?."),
			tools.Opt("multi", "bool", "Return a list of values for every key.", false),
		},
		Returns: "any",
		Example: `url.query_parse("?a=1&b=x%20y&a=2", multi=True)`,
		Tags:    []string{"url", "query", "params", "parse"},
		Fn: func(a tools.Args) (any, error) {
			keys, vals, err := parseQuery(a.Str("query"))
			if err != nil {
				return nil, err
			}
			return queryDict(keys, vals, a.Bool("multi")), nil
		},
	})
	tools.Register(tools.Func{
		Group: "url", Name: "query_encode", Title: "Encode a query string",
		Doc:   "Encode a dict as a query string without the leading ?. Values may be strings, numbers, bools or lists (one key=value pair per item); None values are skipped. Keys keep their order unless sort=True. Spaces become + by default, or %20 with plus=False. Raises on nested dicts.",
		Terms: []string{"dict to query string", "urlencode", "form encode", "build query", "params to string"},
		Params: []tools.Param{
			tools.Req("params", "any", "A dict of parameters."),
			tools.Opt("sort", "bool", "Sort keys alphabetically.", false),
			tools.Opt("plus", "bool", "Write spaces as + instead of %20.", true),
		},
		Returns: "str",
		Example: `url.query_encode({"q": "go lang", "tags": ["a", "b"], "n": 2})`,
		Tags:    []string{"url", "query", "encode", "urlencode", "form"},
		Fn: func(a tools.Args) (any, error) {
			d, ok := a["params"].(*tools.Dict)
			if !ok {
				return nil, fmt.Errorf("params must be a dict")
			}
			return encodeQuery(d, a.Bool("sort"), a.Bool("plus"))
		},
	})
	tools.Register(tools.Func{
		Group: "url", Name: "normalize", Title: "Normalize URL",
		Doc:   "Rewrite a URL into a canonical form so equal URLs compare equal: lower-case scheme and host, default port removed, dot segments resolved, percent escapes upper-cased and unnecessary ones decoded, an empty path made /, and the query sorted (sort_query). Options also drop the fragment and a trailing slash. Raises on malformed URLs.",
		Terms: []string{"canonical url", "dedupe urls", "clean url", "compare urls", "same url", "canonicalize"},
		Params: []tools.Param{
			tools.Req("url", "str", "The URL."),
			tools.Opt("sort_query", "bool", "Sort query parameters by key.", true),
			tools.Opt("strip_fragment", "bool", "Remove the #fragment.", false),
			tools.Opt("strip_trailing_slash", "bool", "Remove a trailing slash from the path (except for /).", false),
		},
		Returns: "str",
		Example: `url.normalize("HTTP://Example.com:80/a/./b/../c?b=2&a=1")`,
		Tags:    []string{"url", "normalize", "canonical", "dedupe"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseURL(a.Str("url"), "")
			if err != nil {
				return nil, err
			}
			return normalize(u, a.Bool("sort_query"), a.Bool("strip_fragment"), a.Bool("strip_trailing_slash")), nil
		},
	})
}

func parseURL(s, defScheme string) (*url.URL, error) {
	if len(s) > maxLen {
		return nil, fmt.Errorf("url is longer than %d bytes", maxLen)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty URL")
	}
	for _, r := range s {
		if r <= ' ' || r == 0x7f {
			return nil, fmt.Errorf("URL contains whitespace or a control character")
		}
	}
	if defScheme != "" && !hasScheme(s) && !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "?") && !strings.HasPrefix(s, "#") {
		s = strings.TrimSuffix(strings.TrimSuffix(defScheme, "/"), ":") + "://" + strings.TrimPrefix(s, "//")
	} else if defScheme != "" && strings.HasPrefix(s, "//") {
		s = strings.TrimSuffix(defScheme, ":") + ":" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", strings.TrimPrefix(err.Error(), "parse "))
	}
	return u, nil
}

func hasScheme(s string) bool {
	i := strings.IndexByte(s, ':')
	if i <= 0 {
		return false
	}
	for j, c := range s[:i] {
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || j > 0 && (c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.')
		if !ok {
			return false
		}
	}
	// "host:8080" is not a scheme
	rest := s[i+1:]
	if rest != "" && rest[0] >= '0' && rest[0] <= '9' {
		allDigits := true
		for _, c := range rest {
			if c == '/' || c == '?' || c == '#' {
				break
			}
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		return !allDigits
	}
	return true
}

func toDict(u *url.URL) *tools.Dict {
	d := tools.NewDict()
	d.Set("scheme", u.Scheme)
	pw, _ := u.User.Password()
	d.Set("user", u.User.Username())
	d.Set("password", pw)
	d.Set("host", u.Host)
	d.Set("hostname", u.Hostname())
	d.Set("port", u.Port())
	d.Set("path", u.Path)
	keys, vals, err := parseQuery(u.RawQuery)
	if err != nil {
		keys, vals = nil, nil
	}
	d.Set("query", queryDict(keys, vals, true))
	d.Set("query_string", u.RawQuery)
	d.Set("fragment", u.Fragment)
	origin := ""
	if u.Scheme != "" && u.Host != "" {
		origin = strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
	}
	d.Set("origin", origin)
	return d
}

func parseQuery(q string) (keys []string, vals map[string][]string, err error) {
	q = strings.TrimPrefix(q, "?")
	vals = map[string][]string{}
	if q == "" {
		return nil, vals, nil
	}
	for _, pair := range strings.Split(q, "&") {
		if pair == "" {
			continue
		}
		k, v, _ := strings.Cut(pair, "=")
		dk, err := url.QueryUnescape(k)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid query key %q: %s", k, shortErr(err))
		}
		dv, err := url.QueryUnescape(v)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid query value %q: %s", v, shortErr(err))
		}
		if _, ok := vals[dk]; !ok {
			keys = append(keys, dk)
		}
		vals[dk] = append(vals[dk], dv)
	}
	return keys, vals, nil
}

func shortErr(err error) string {
	if e, ok := err.(url.EscapeError); ok {
		return "bad percent escape " + string(e)
	}
	return err.Error()
}

func queryDict(keys []string, vals map[string][]string, multi bool) *tools.Dict {
	d := tools.NewDict()
	for _, k := range keys {
		vs := vals[k]
		if multi {
			l := make([]any, len(vs))
			for i, v := range vs {
				l[i] = v
			}
			d.Set(k, l)
		} else {
			d.Set(k, vs[len(vs)-1])
		}
	}
	return d
}

func escape(s string, plus bool) string {
	e := url.QueryEscape(s)
	if !plus {
		e = strings.ReplaceAll(e, "+", "%20")
	}
	return e
}

func encodeQuery(d *tools.Dict, sorted, plus bool) (string, error) {
	keys := d.Keys()
	if sorted {
		sort.Strings(keys)
	}
	var parts []string
	for _, k := range keys {
		v, _ := d.Get(k)
		items := []any{v}
		if l, ok := v.([]any); ok {
			items = l
		}
		for _, it := range items {
			if it == nil {
				continue
			}
			s, ok := tools.Scalar(it)
			if !ok {
				return "", fmt.Errorf("parameter %q: values must be strings, numbers, bools or lists of them", k)
			}
			parts = append(parts, escape(k, plus)+"="+escape(s, plus))
		}
	}
	return strings.Join(parts, "&"), nil
}

func str(d *tools.Dict, k string) (string, error) {
	v, ok := d.Get(k)
	if !ok || v == nil {
		return "", nil
	}
	s, ok := tools.Scalar(v)
	if !ok {
		return "", fmt.Errorf("%s must be a string or number", k)
	}
	return s, nil
}

func build(d *tools.Dict) (string, error) {
	get := func(k string) (string, error) { return str(d, k) }
	scheme, err := get("scheme")
	if err != nil {
		return "", err
	}
	user, err := get("user")
	if err != nil {
		return "", err
	}
	pass, err := get("password")
	if err != nil {
		return "", err
	}
	host, err := get("host")
	if err != nil {
		return "", err
	}
	hostname, err := get("hostname")
	if err != nil {
		return "", err
	}
	port, err := get("port")
	if err != nil {
		return "", err
	}
	path, err := get("path")
	if err != nil {
		return "", err
	}
	frag, err := get("fragment")
	if err != nil {
		return "", err
	}
	if host == "" && hostname != "" {
		host = hostname
		if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
			host = "[" + host + "]"
		}
		if port != "" {
			host += ":" + port
		}
	}
	if port != "" {
		for _, c := range port {
			if c < '0' || c > '9' {
				return "", fmt.Errorf("port must be a number, got %q", port)
			}
		}
	}
	u := &url.URL{Scheme: scheme, Host: host, Fragment: frag}
	if user != "" || pass != "" {
		if pass != "" {
			u.User = url.UserPassword(user, pass)
		} else {
			u.User = url.User(user)
		}
	}
	if host != "" && path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = path
	if q, ok := d.Get("query"); ok && q != nil {
		switch x := q.(type) {
		case string:
			u.RawQuery = strings.TrimPrefix(x, "?")
		case *tools.Dict:
			if u.RawQuery, err = encodeQuery(x, false, true); err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("query must be a dict or a string")
		}
	}
	return u.String(), nil
}

func normalize(u *url.URL, sortQuery, stripFrag, stripSlash bool) string {
	c := *u
	c.Scheme = strings.ToLower(c.Scheme)
	if c.Host != "" {
		h := strings.ToLower(c.Hostname())
		p := c.Port()
		if p == defaultPorts[c.Scheme] {
			p = ""
		}
		if strings.Contains(h, ":") {
			h = "[" + h + "]"
		}
		c.Host = h
		if p != "" {
			c.Host = h + ":" + p
		}
	}
	p := c.EscapedPath()
	if p == "" && c.Host != "" {
		p = "/"
	}
	if p != "" {
		p = normPercent(cleanDots(p))
		if stripSlash && len(p) > 1 && strings.HasSuffix(p, "/") {
			p = strings.TrimRight(p, "/")
			if p == "" {
				p = "/"
			}
		}
		if raw, err := url.PathUnescape(p); err == nil {
			c.Path, c.RawPath = raw, p
		}
	}
	if sortQuery && c.RawQuery != "" {
		parts := strings.Split(c.RawQuery, "&")
		sort.SliceStable(parts, func(i, j int) bool {
			ki, _, _ := strings.Cut(parts[i], "=")
			kj, _, _ := strings.Cut(parts[j], "=")
			return ki < kj
		})
		c.RawQuery = strings.Join(parts, "&")
	}
	if stripFrag {
		c.Fragment, c.RawFragment = "", ""
	}
	return c.String()
}

// normPercent upper-cases the hex digits of %xx escapes and decodes the ones
// that stand for unreserved characters (letters, digits, - . _ ~).
func normPercent(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			v := unhex(s[i+1])<<4 | unhex(s[i+2])
			if v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' || v == '-' || v == '.' || v == '_' || v == '~' {
				b.WriteByte(v)
			} else {
				b.WriteString("%" + strings.ToUpper(s[i+1:i+3]))
			}
			i += 2
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	}
	return c - 'A' + 10
}

// cleanDots removes . and .. segments, keeping a trailing slash.
func cleanDots(p string) string {
	trailing := strings.HasSuffix(p, "/") && len(p) > 1
	abs := strings.HasPrefix(p, "/")
	var out []string
	for _, seg := range strings.Split(p, "/") {
		switch seg {
		case "", ".":
		case "..":
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		default:
			out = append(out, seg)
		}
	}
	r := strings.Join(out, "/")
	if abs {
		r = "/" + r
	}
	if trailing && !strings.HasSuffix(r, "/") {
		r += "/"
	}
	if r == "" {
		return "/"
	}
	return r
}
