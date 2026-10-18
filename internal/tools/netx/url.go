package netx

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

var defaultPorts = map[string]int{
	"http": 80, "https": 443, "ftp": 21, "ftps": 990, "ws": 80, "wss": 443, "ssh": 22, "sftp": 22, "telnet": 23,
	"smtp": 25, "smtps": 465, "imap": 143, "imaps": 993, "pop3": 110, "pop3s": 995, "ldap": 389, "ldaps": 636,
	"mysql": 3306, "postgres": 5432, "postgresql": 5432, "redis": 6379, "rediss": 6379, "mongodb": 27017,
	"rtsp": 554, "rdp": 3389, "vnc": 5900, "git": 9418, "amqp": 5672, "amqps": 5671, "mqtt": 1883, "mqtts": 8883,
	"nntp": 119, "irc": 194, "ircs": 6697, "sip": 5060, "sips": 5061, "xmpp": 5222, "svn": 3690, "smb": 445,
	"nfs": 2049, "socks5": 1080, "socks": 1080, "gopher": 70, "memcached": 11211, "kafka": 9092, "clickhouse": 9000,
}

var (
	schemeRe    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.\-]*:`)
	hostStartRe = regexp.MustCompile(`^(?:\[[0-9A-Fa-f:.]+\]|[^\s/?#:@]+)(?::\d*)?(?:[/?#]|$)`)
)

// parseURL parses a URL; input without a scheme such as "example.com/a" is
// treated as a network location.
func parseURL(s string) (*url.URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty URL")
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return nil, fmt.Errorf("URL %q contains whitespace", s)
	}
	if !schemeRe.MatchString(s) || hostPortLike(s) {
		if !strings.HasPrefix(s, "//") && !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, "?") && !strings.HasPrefix(s, "#") && hostStartRe.MatchString(s) {
			s = "//" + s
		}
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", strings.TrimPrefix(err.Error(), "parse "))
	}
	return u, nil
}

// hostPortLike detects "host:8080/x" which looks like a scheme to the regexp.
func hostPortLike(s string) bool {
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return false
	}
	rest := s[i+1:]
	j := 0
	for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
		j++
	}
	return j > 0 && (j == len(rest) || rest[j] == '/' || rest[j] == '?' || rest[j] == '#') && strings.Contains(s[:i], ".") || s[:i] == "localhost" && j > 0
}

func hostWithPort(u *url.URL) string { return u.Host }

func originOf(u *url.URL) string {
	if u.Scheme == "" || u.Host == "" {
		return ""
	}
	h := strings.ToLower(u.Hostname())
	if strings.Contains(h, ":") {
		h = "[" + h + "]"
	}
	p := u.Port()
	if p != "" {
		if n, ok := defaultPorts[strings.ToLower(u.Scheme)]; ok && strconv.Itoa(n) == p {
			p = ""
		}
	}
	o := strings.ToLower(u.Scheme) + "://" + h
	if p != "" {
		o += ":" + p
	}
	return o
}

func urlParts(s string) (map[string]string, error) {
	u, err := parseURL(s)
	if err != nil {
		return nil, err
	}
	m := map[string]string{
		"scheme": u.Scheme, "host": u.Host, "hostname": u.Hostname(), "port": u.Port(),
		"path": u.EscapedPath(), "query": u.RawQuery, "fragment": u.EscapedFragment(), "origin": originOf(u),
	}
	if u.User != nil {
		m["user"] = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			m["password"] = pw
		}
	}
	if u.Opaque != "" {
		m["opaque"] = u.Opaque
	}
	scheme := strings.ToLower(u.Scheme)
	if n, ok := defaultPorts[scheme]; ok {
		m["default_port"] = strconv.Itoa(n)
		if u.Port() == "" {
			m["effective_port"] = strconv.Itoa(n)
		} else {
			m["effective_port"] = u.Port()
		}
	} else if u.Port() != "" {
		m["effective_port"] = u.Port()
	}
	if h := u.Hostname(); h != "" {
		if a, err := parseAddr(h); err == nil {
			m["host_type"] = "ip"
			if a.Is6() {
				m["host_type"] = "ipv6"
			} else {
				m["host_type"] = "ipv4"
			}
		} else {
			m["host_type"] = "domain"
		}
	}
	m["file_name"] = ""
	if p := u.Path; p != "" && !strings.HasSuffix(p, "/") {
		m["file_name"] = p[strings.LastIndex(p, "/")+1:]
	}
	return m, nil
}

func buildURL(p map[string]string) (string, error) {
	u := &url.URL{Scheme: p["scheme"]}
	host := p["host"]
	if host == "" {
		host = p["hostname"]
	}
	if port := p["port"]; port != "" {
		if n, err := strconv.Atoi(port); err != nil || n < 0 || n > 65535 {
			return "", fmt.Errorf("invalid port %q", port)
		}
		if h, _, ok := splitHostPortLoose(host); ok && h != "" {
			host = h
		}
		if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
			host = "[" + host + "]"
		}
		host += ":" + port
	} else if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		if _, err := parseAddr(host); err == nil {
			host = "[" + host + "]"
		}
	}
	u.Host = host
	if user := p["user"]; user != "" {
		if pw, ok := p["password"]; ok && pw != "" {
			u.User = url.UserPassword(user, pw)
		} else {
			u.User = url.User(user)
		}
	}
	path := p["path"]
	if path != "" && !strings.HasPrefix(path, "/") && host != "" {
		path = "/" + path
	}
	if dec, err := url.PathUnescape(path); err == nil {
		u.Path = dec
		u.RawPath = ""
		if url.PathEscape(dec) != "" && u.EscapedPath() != path && strings.Contains(path, "%") {
			u.RawPath = path
		}
	} else {
		u.Path = path
	}
	u.RawQuery = strings.TrimPrefix(p["query"], "?")
	if f := strings.TrimPrefix(p["fragment"], "#"); f != "" {
		u.Fragment = f
	}
	if u.Scheme == "" && u.Host == "" && u.Path == "" {
		return "", fmt.Errorf("need at least a scheme, host or path")
	}
	return u.String(), nil
}

func splitHostPortLoose(s string) (host, port string, ok bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") {
		i := strings.Index(s, "]")
		if i < 0 {
			return "", "", false
		}
		host = s[1:i]
		rest := s[i+1:]
		if rest == "" {
			return host, "", true
		}
		if !strings.HasPrefix(rest, ":") {
			return "", "", false
		}
		return host, rest[1:], true
	}
	if strings.Count(s, ":") > 1 { // bare IPv6
		return s, "", true
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[:i], s[i+1:], true
	}
	return s, "", true
}

// removeDotSegments implements RFC 3986 section 5.2.4.
func removeDotSegments(p string) string {
	var out []string
	in := p
	for in != "" {
		switch {
		case strings.HasPrefix(in, "../"):
			in = in[3:]
		case strings.HasPrefix(in, "./"):
			in = in[2:]
		case strings.HasPrefix(in, "/./"):
			in = in[2:]
		case in == "/.":
			in = "/"
		case strings.HasPrefix(in, "/../"):
			in = in[3:]
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		case in == "/..":
			in = "/"
			if len(out) > 0 {
				out = out[:len(out)-1]
			}
		case in == "." || in == "..":
			in = ""
		default:
			i := strings.Index(in[1:], "/")
			if i < 0 {
				out = append(out, in)
				in = ""
			} else {
				out = append(out, in[:i+1])
				in = in[i+1:]
			}
		}
	}
	return strings.Join(out, "")
}

func isUnreserved(c byte) bool {
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~'
}

// normalizePct upper-cases percent escapes and decodes unreserved characters.
func normalizePct(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			v, err := strconv.ParseUint(s[i+1:i+3], 16, 8)
			if err == nil {
				if isUnreserved(byte(v)) {
					b.WriteByte(byte(v))
				} else {
					b.WriteString("%" + strings.ToUpper(s[i+1:i+3]))
				}
				i += 2
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

type qpair struct{ key, raw string }

func parseQueryPairs(q string) []qpair {
	q = strings.TrimPrefix(q, "?")
	if q == "" {
		return nil
	}
	var out []qpair
	for _, seg := range strings.Split(q, "&") {
		if seg == "" {
			continue
		}
		k, _, _ := strings.Cut(seg, "=")
		if dk, err := url.QueryUnescape(k); err == nil {
			k = dk
		}
		out = append(out, qpair{k, seg})
	}
	return out
}

func joinPairs(ps []qpair) string {
	segs := make([]string, len(ps))
	for i, p := range ps {
		segs[i] = p.raw
	}
	return strings.Join(segs, "&")
}

func encPair(k, v string) qpair {
	return qpair{k, url.QueryEscape(k) + "=" + url.QueryEscape(v)}
}

func normalizeURL(s string, sortQuery, stripFragment bool) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Opaque != "" {
		return u.String(), nil
	}
	h := strings.ToLower(u.Hostname())
	h = strings.TrimSuffix(h, ".")
	port := u.Port()
	if n, ok := defaultPorts[u.Scheme]; ok && port == strconv.Itoa(n) {
		port = ""
	}
	if strings.Contains(h, ":") {
		h = "[" + h + "]"
	}
	if h != "" {
		u.Host = h
		if port != "" {
			u.Host += ":" + port
		}
	}
	p := normalizePct(u.EscapedPath())
	p = removeDotSegments(p)
	if p == "" && u.Host != "" {
		p = "/"
	}
	if dec, err := url.PathUnescape(p); err == nil {
		u.Path = dec
		u.RawPath = p
	}
	q := parseQueryPairs(u.RawQuery)
	for i := range q {
		q[i].raw = normalizePct(q[i].raw)
	}
	if sortQuery {
		sort.SliceStable(q, func(i, j int) bool { return q[i].raw < q[j].raw })
	}
	u.RawQuery = joinPairs(q)
	u.ForceQuery = false
	if stripFragment || u.Fragment == "" {
		u.Fragment, u.RawFragment = "", ""
	}
	return u.String(), nil
}

var trackingExact = map[string]bool{
	"fbclid": true, "gclid": true, "gclsrc": true, "dclid": true, "gbraid": true, "wbraid": true, "msclkid": true,
	"yclid": true, "twclid": true, "ttclid": true, "li_fat_id": true, "igshid": true, "igsh": true, "mc_cid": true, "mc_eid": true,
	"_hsenc": true, "_hsmi": true, "__hssc": true, "__hstc": true, "__hsfp": true, "hsctatracking": true,
	"mkt_tok": true, "vero_id": true, "vero_conv": true, "oly_enc_id": true, "oly_anon_id": true, "wickedid": true,
	"_ga": true, "_gl": true, "ref_src": true, "ref_url": true, "s_cid": true, "spm": true, "trk": true, "trkcampaign": true,
	"sr_share": true, "fb_action_ids": true, "fb_action_types": true, "fb_ref": true, "fb_source": true, "ocid": true,
	"cmpid": true, "campaignid": true, "adgroupid": true, "pk_campaign": true, "pk_kwd": true, "piwik_campaign": true,
	"piwik_kwd": true, "ncid": true, "soc_src": true, "soc_trk": true, "mc_tc": true, "icid": true, "_openstat": true,
	"yclid_": true, "rb_clickid": true, "s_kwcid": true, "ef_id": true, "epik": true,
}

func isTrackingParam(k string) bool {
	l := strings.ToLower(k)
	return trackingExact[l] || strings.HasPrefix(l, "utm_") || strings.HasPrefix(l, "hsa_") || strings.HasPrefix(l, "mtm_")
}

func rewriteQuery(s string, f func(q []qpair) []qpair) (string, error) {
	u, err := parseURL(s)
	if err != nil {
		return "", err
	}
	u.RawQuery = joinPairs(f(parseQueryPairs(u.RawQuery)))
	u.ForceQuery = false
	out := u.String()
	// parseURL may have added "//" to a scheme-less input; undo it.
	if !schemeRe.MatchString(strings.TrimSpace(s)) || hostPortLike(strings.TrimSpace(s)) {
		if !strings.HasPrefix(strings.TrimSpace(s), "//") {
			out = strings.TrimPrefix(out, "//")
		}
	}
	return out, nil
}

func keyMatches(pattern, key string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(strings.ToLower(key), strings.ToLower(strings.TrimSuffix(pattern, "*")))
	}
	return key == pattern
}

func accURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	if strings.Contains(s, "://") || strings.HasPrefix(s, "//") || strings.HasPrefix(strings.ToLower(s), "mailto:") {
		return true
	}
	return looksHostLike(s)
}

func looksHostLike(s string) bool {
	h := s
	if i := strings.IndexAny(h, "/?#"); i >= 0 {
		h = h[:i]
	}
	if i := strings.LastIndex(h, "@"); i >= 0 {
		h = h[i+1:]
	}
	if hh, _, ok := splitHostPortLoose(h); ok {
		h = hh
	}
	if strings.EqualFold(h, "localhost") {
		return true
	}
	return strings.Contains(h, ".") && validHostname(h, false) && !strings.HasPrefix(h, ".")
}

func init() {
	pURL := req("url", "str", "URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path.")
	reg(kwURL, tools.Func{
		Name: "url_parse", Title: "Parse URL", Order: 5,
		Doc:     "Splits a URL into scheme, user, password, host, hostname, port, effective_port (the default port of the scheme when none is given), path, raw query, fragment, origin, host_type (ipv4/ipv6/domain) and file_name. Components keep their percent-encoding; absent parts are empty or missing.",
		Params:  []tools.Param{pURL},
		Returns: "dict[str, str]", Example: `net.url_parse("https://user:pw@example.com:8443/a/b.html?x=1&y=2#top")`,
		Tags:    []string{"uri", "split url", "components", "query string", "hostname", "port", "origin", "scheme", "fragment", "dissect"},
		Accepts: accURL,
		Fn:      func(a tools.Args) (any, error) { return urlParts(a.Str("url")) },
	})
	reg(kwURL, tools.Func{
		Name: "url_build", Title: "Build URL", Order: 90, Input: "scheme",
		Doc:     "Builds a URL from parts. Recognized keys of parts: scheme, user, password, host (or hostname), port, path, query (raw, without ?) and fragment. IPv6 hosts are bracketed automatically and the path is percent-encoded as needed.",
		Params:  []tools.Param{req("parts", "dict[str, str]", `Dict with keys scheme, host, port, path, query, fragment, user, password.`), opt("scheme", "str", "Default scheme used when parts has none.", "https")},
		Returns: "str", Example: `net.url_build({"scheme": "https", "host": "example.com", "path": "/a b", "query": "x=1"})`,
		Tags: []string{"compose url", "assemble", "construct", "uri builder", "join parts"},
		Fn: func(a tools.Args) (any, error) {
			parts := map[string]string{}
			for k, v := range a["parts"].(map[string]string) {
				parts[strings.ToLower(k)] = v
			}
			if parts["scheme"] == "" {
				parts["scheme"] = a.Str("scheme")
			}
			return buildURL(parts)
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_join", Title: "Resolve relative URL", Order: 91, Input: "base",
		Doc:     "Resolves a relative reference against a base URL following RFC 3986, like a browser does for links (../img/a.png against https://x.com/a/b/c.html gives https://x.com/a/img/a.png). An absolute ref wins over the base.",
		Params:  []tools.Param{req("base", "str", "Base URL."), req("ref", "str", "Relative or absolute reference.")},
		Returns: "str", Example: `net.url_join("https://example.com/a/b/c.html", "../img/x.png")`,
		Tags: []string{"resolve", "relative link", "combine urls", "absolute url", "href"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			b, err := url.Parse(strings.TrimSpace(a.Str("base")))
			if err != nil {
				return nil, fmt.Errorf("invalid base URL: %v", err)
			}
			r, err := url.Parse(strings.TrimSpace(a.Str("ref")))
			if err != nil {
				return nil, fmt.Errorf("invalid reference: %v", err)
			}
			return b.ResolveReference(r).String(), nil
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_normalize", Title: "Normalize URL", Order: 92,
		Doc:     "Normalizes a URL for comparison: lower-cases scheme and host, removes the default port and a trailing host dot, resolves ./ and ../ segments, upper-cases percent escapes, decodes unreserved characters, adds / for an empty path and drops an empty fragment. sort_query orders query parameters; strip_fragment removes #fragment.",
		Params:  []tools.Param{pURL, opt("sort_query", "bool", "Sort query parameters alphabetically.", false), opt("strip_fragment", "bool", "Remove the #fragment.", false)},
		Returns: "str", Example: `net.url_normalize("HTTP://Example.COM:80/a/./b/../c?b=2&a=1", True)`,
		Tags:    []string{"canonical url", "canonicalize", "dedupe urls", "clean url", "dot segments", "compare urls"},
		Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			return normalizeURL(a.Str("url"), a.Bool("sort_query"), a.Bool("strip_fragment"))
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_query", Title: "URL query parameters", Order: 93,
		Doc:     "Returns the query parameters of a URL (or of a bare query string such as a=1&b=2) as a dict with decoded names and values. For repeated keys the values are joined with commas; a key without = has an empty value.",
		Params:  []tools.Param{req("url", "str", "URL or bare query string.")},
		Returns: "dict[str, str]", Example: `net.url_query("https://example.com/s?q=go+lang&tag=a&tag=b")`,
		Tags:    []string{"query string", "get parameters", "params", "args", "parse query", "querystring"},
		Accepts: func(s string) bool { return strings.Contains(s, "=") || accURL(s) },
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("url"))
			q := s
			if strings.Contains(s, "?") || strings.Contains(s, "://") || strings.HasPrefix(s, "/") {
				if i := strings.Index(s, "?"); i >= 0 {
					q = s[i+1:]
				} else {
					q = ""
				}
				if i := strings.Index(q, "#"); i >= 0 {
					q = q[:i]
				}
			}
			out := map[string]string{}
			for _, seg := range strings.Split(strings.TrimPrefix(q, "?"), "&") {
				if seg == "" {
					continue
				}
				k, v, _ := strings.Cut(seg, "=")
				if dk, err := url.QueryUnescape(k); err == nil {
					k = dk
				}
				if dv, err := url.QueryUnescape(v); err == nil {
					v = dv
				}
				if old, ok := out[k]; ok {
					out[k] = old + "," + v
				} else {
					out[k] = v
				}
			}
			return out, nil
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_set_query", Title: "Replace URL query", Order: 94,
		Doc:     "Replaces the whole query string of a URL with the given parameters (sorted by name, percent-encoded). An empty dict removes the query.",
		Params:  []tools.Param{req("url", "str", "URL to modify."), req("params", "dict[str, str]", "New query parameters.")},
		Returns: "str", Example: `net.url_set_query("https://example.com/s?old=1", {"q": "go lang", "page": "2"})`,
		Tags: []string{"change query", "set params", "rewrite query"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			ps := a["params"].(map[string]string)
			keys := make([]string, 0, len(ps))
			for k := range ps {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return rewriteQuery(a.Str("url"), func([]qpair) []qpair {
				var out []qpair
				for _, k := range keys {
					out = append(out, encPair(k, ps[k]))
				}
				return out
			})
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_add_param", Title: "Add query parameter", Order: 95,
		Doc:     "Appends a query parameter to a URL, keeping existing parameters (a repeated key is added again). Use url_set_param to replace instead.",
		Params:  []tools.Param{req("url", "str", "URL to modify."), req("key", "str", "Parameter name."), req("value", "str", "Parameter value (encoded for you).")},
		Returns: "str", Example: `net.url_add_param("https://example.com/s?q=1", "page", "2")`,
		Tags: []string{"append param", "add query", "append query"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			return rewriteQuery(a.Str("url"), func(q []qpair) []qpair { return append(q, encPair(a.Str("key"), a.Str("value"))) })
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_set_param", Title: "Set query parameter", Order: 96,
		Doc:     "Sets a query parameter: replaces the first occurrence in place, removes any further occurrences, or appends it when absent.",
		Params:  []tools.Param{req("url", "str", "URL to modify."), req("key", "str", "Parameter name."), req("value", "str", "Parameter value (encoded for you).")},
		Returns: "str", Example: `net.url_set_param("https://example.com/s?q=1&page=1", "page", "5")`,
		Tags: []string{"replace param", "update query", "overwrite"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			key := a.Str("key")
			return rewriteQuery(a.Str("url"), func(q []qpair) []qpair {
				var out []qpair
				done := false
				for _, p := range q {
					if p.key != key {
						out = append(out, p)
					} else if !done {
						out = append(out, encPair(key, a.Str("value")))
						done = true
					}
				}
				if !done {
					out = append(out, encPair(key, a.Str("value")))
				}
				return out
			})
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_remove_param", Title: "Remove query parameters", Order: 97,
		Doc:     "Removes the named query parameters from a URL. A name ending in * is a prefix match (utm_* removes utm_source, utm_medium...). Other parameters keep their original encoding and order.",
		Params:  []tools.Param{req("url", "str", "URL to modify."), req("keys", "list[str]", `Parameter names to remove, e.g. ["utm_*", "ref"].`)},
		Returns: "str", Example: `net.url_remove_param("https://example.com/?a=1&b=2", ["b"])`,
		Tags: []string{"delete param", "strip query", "drop parameter"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			keys := a.List("keys")
			return rewriteQuery(a.Str("url"), func(q []qpair) []qpair {
				var out []qpair
			next:
				for _, p := range q {
					for _, k := range keys {
						if keyMatches(k, p.key) {
							continue next
						}
					}
					out = append(out, p)
				}
				return out
			})
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_strip_tracking", Title: "Strip tracking parameters", Order: 98,
		Doc:     "Removes well-known tracking parameters from a URL: utm_*, fbclid, gclid, dclid, msclkid, yclid, mc_cid, mc_eid, igshid, twclid, ttclid, _hsenc, _ga, ref_src, spm, mkt_tok and more. Other parameters and the fragment are kept.",
		Params:  []tools.Param{req("url", "str", "URL to clean.")},
		Returns: "str", Example: `net.url_strip_tracking("https://example.com/p?id=7&utm_source=x&fbclid=abc")`,
		Tags:    []string{"clean link", "remove utm", "privacy", "tracking", "utm", "fbclid", "gclid", "sanitize url"},
		Accepts: func(s string) bool { return accURL(s) && strings.Contains(s, "?") },
		Fn: func(a tools.Args) (any, error) {
			return rewriteQuery(a.Str("url"), func(q []qpair) []qpair {
				var out []qpair
				for _, p := range q {
					if !isTrackingParam(p.key) {
						out = append(out, p)
					}
				}
				return out
			})
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_origin", Title: "URL origin", Order: 99,
		Doc:     "Returns the web origin (scheme://host[:port]) of a URL, lower-cased, with the port omitted when it is the scheme's default. Fails for URLs without scheme or host.",
		Params:  []tools.Param{pURL},
		Returns: "str", Example: `net.url_origin("https://Example.com:443/a?b=1")`,
		Tags: []string{"same origin", "cors", "scheme host port", "base url"}, Accepts: accURL,
		Fn: func(a tools.Args) (any, error) {
			u, err := parseURL(a.Str("url"))
			if err != nil {
				return nil, err
			}
			o := originOf(u)
			if o == "" {
				return nil, fmt.Errorf("URL needs a scheme and a host to have an origin")
			}
			return o, nil
		},
	})
	reg(kwURL, tools.Func{
		Name: "url_is_valid", Title: "Is valid URL", Order: 100,
		Doc:     "Reports whether the text is an absolute URL: a scheme, then for http/https/ftp/ws/wss a valid host (domain, IPv4 or [IPv6]) and an optional valid port. Other schemes (mailto:, file:, data:) only need to parse. Whitespace makes it invalid.",
		Params:  []tools.Param{req("url", "str", "Text to check.")},
		Returns: "bool", Example: `net.url_is_valid("https://example.com/path")`,
		Tags: []string{"validate url", "check link", "well-formed"},
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("url"))
			if s != a.Str("url") || !schemeRe.MatchString(s) || hostPortLike(s) {
				return false, nil
			}
			u, err := parseURL(s)
			if err != nil {
				return false, nil
			}
			switch strings.ToLower(u.Scheme) {
			case "http", "https", "ftp", "ftps", "ws", "wss", "ssh", "sftp":
				h := u.Hostname()
				if h == "" {
					return false, nil
				}
				if p := u.Port(); p != "" {
					if n, err := strconv.Atoi(p); err != nil || n > 65535 {
						return false, nil
					}
				}
				if _, err := parseAddr(h); err == nil {
					return true, nil
				}
				return validHostname(h, true), nil
			}
			return true, nil
		},
	})
}
