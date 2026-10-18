package netx

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"

	"github.com/avestura/zagros/internal/tools"
)

var labelRe = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)
var labelUnderscoreRe = regexp.MustCompile(`^[A-Za-z0-9_](?:[A-Za-z0-9_-]{0,61}[A-Za-z0-9_])?$`)

// validHostname checks RFC 1123 syntax (Unicode names are converted to
// punycode first). A single trailing dot is allowed.
func validHostname(h string, allowUnderscore bool) bool {
	h = strings.TrimSuffix(h, ".")
	if h == "" {
		return false
	}
	if !isASCII(h) {
		a, err := idna.Lookup.ToASCII(h)
		if err != nil {
			return false
		}
		h = a
	}
	if len(h) > 253 {
		return false
	}
	re := labelRe
	if allowUnderscore {
		re = labelUnderscoreRe
	}
	for _, l := range strings.Split(h, ".") {
		if !re.MatchString(l) {
			return false
		}
	}
	return true
}

func validDomain(h string) bool {
	h = strings.TrimSuffix(strings.TrimSpace(h), ".")
	if !validHostname(h, false) {
		return false
	}
	if _, err := parseAddr(h); err == nil {
		return false
	}
	a, _ := idna.Lookup.ToASCII(h)
	labels := strings.Split(a, ".")
	if len(labels) < 2 {
		return false
	}
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	if strings.HasPrefix(tld, "xn--") {
		return true
	}
	for _, r := range tld {
		if r < 'a' && r < 'A' || r > 'z' || (r > 'Z' && r < 'a') {
			return false
		}
	}
	return true
}

func hostFromInput(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "://") || strings.HasPrefix(s, "//") {
		if u, err := parseURL(s); err == nil && u.Hostname() != "" {
			return u.Hostname()
		}
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "@"); i >= 0 {
		s = s[i+1:]
	}
	if h, _, ok := splitHostPortLoose(s); ok {
		s = h
	}
	return strings.TrimSuffix(strings.ToLower(s), ".")
}

func domainParts(in string) (map[string]string, error) {
	h := hostFromInput(in)
	if h == "" {
		return nil, fmt.Errorf("no host in %q", in)
	}
	if _, err := parseAddr(h); err == nil {
		return nil, fmt.Errorf("%q is an IP address, not a domain name", h)
	}
	ascii, err := idna.Lookup.ToASCII(h)
	if err != nil {
		return nil, fmt.Errorf("invalid domain name %q", h)
	}
	suffix, icann := publicsuffix.PublicSuffix(ascii)
	kind := "private"
	switch {
	case icann:
		kind = "icann"
	case !strings.Contains(suffix, "."):
		kind = "unknown"
	}
	uni := func(s string) string {
		if u, err := idna.Lookup.ToUnicode(s); err == nil {
			return u
		}
		return s
	}
	labels := strings.Split(ascii, ".")
	m := map[string]string{
		"host": ascii, "unicode_host": uni(ascii), "tld": labels[len(labels)-1], "public_suffix": suffix,
		"suffix_type": kind, "is_icann": strconv.FormatBool(icann), "labels": strconv.Itoa(len(labels)),
	}
	reg, err := publicsuffix.EffectiveTLDPlusOne(ascii)
	if err != nil {
		m["registrable_domain"] = ""
		m["domain"] = ""
		m["subdomain"] = ""
		m["note"] = "the host is itself a public suffix"
		return m, nil
	}
	m["registrable_domain"] = reg
	m["domain"] = strings.TrimSuffix(strings.TrimSuffix(reg, suffix), ".")
	m["subdomain"] = strings.TrimSuffix(strings.TrimSuffix(ascii, reg), ".")
	return m, nil
}

var localAtomRe = regexp.MustCompile("^[A-Za-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[A-Za-z0-9!#$%&'*+/=?^_`{|}~-]+)*$")
var quotedLocalRe = regexp.MustCompile(`^"(?:[^"\\\r\n]|\\.)*"$`)

func validEmail(s string) bool {
	if s != strings.TrimSpace(s) || len(s) > 254 {
		return false
	}
	i := strings.LastIndex(s, "@")
	if i <= 0 || i == len(s)-1 {
		return false
	}
	local, dom := s[:i], s[i+1:]
	if len(local) > 64 || !(localAtomRe.MatchString(local) || quotedLocalRe.MatchString(local)) {
		return false
	}
	if strings.HasPrefix(dom, "[") && strings.HasSuffix(dom, "]") {
		inner := strings.TrimPrefix(strings.TrimSuffix(dom[1:len(dom)-1], ""), "IPv6:")
		_, err := parseAddr(inner)
		return err == nil
	}
	return strings.Contains(dom, ".") && validHostname(dom, false) && !strings.HasSuffix(dom, ".")
}

func emailParse(in string) (map[string]string, error) {
	s := strings.TrimSpace(in)
	m := map[string]string{}
	if strings.ContainsAny(s, "<>") {
		a, err := mail.ParseAddress(s)
		if err != nil {
			return nil, fmt.Errorf("invalid email %q", in)
		}
		m["display_name"] = a.Name
		s = a.Address
	}
	s = strings.TrimPrefix(s, "mailto:")
	if !validEmail(s) {
		return nil, fmt.Errorf("invalid email address %q", in)
	}
	i := strings.LastIndex(s, "@")
	local, dom := s[:i], strings.ToLower(s[i+1:])
	base, tag, _ := strings.Cut(local, "+")
	m["address"] = s
	m["local"] = local
	m["domain"] = dom
	m["base_local"] = base
	m["tag"] = tag
	norm := strings.ToLower(local)
	if dom == "gmail.com" || dom == "googlemail.com" {
		b, _, _ := strings.Cut(norm, "+")
		norm = strings.ReplaceAll(b, ".", "") + "@gmail.com"
		m["provider"] = "gmail"
	} else {
		norm = norm + "@" + dom
	}
	m["normalized"] = norm
	if dp, err := domainParts(dom); err == nil {
		m["registrable_domain"] = dp["registrable_domain"]
	}
	return m, nil
}

func parseHostPort(in string) (map[string]string, error) {
	s := strings.TrimSpace(in)
	host, port, ok := splitHostPortLoose(s)
	if !ok || host == "" {
		return nil, fmt.Errorf("invalid host:port %q", in)
	}
	m := map[string]string{"host": host, "port": port}
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 0 || n > 65535 {
			return nil, fmt.Errorf("invalid port %q", port)
		}
	}
	if a, err := parseAddr(host); err == nil {
		m["host"] = a.String()
		m["host_type"] = map[bool]string{true: "ipv4", false: "ipv6"}[a.Is4()]
	} else if validHostname(host, true) {
		m["host_type"] = "domain"
	} else {
		return nil, fmt.Errorf("invalid host %q", host)
	}
	if m["host_type"] == "ipv6" {
		m["host_bracketed"] = "[" + m["host"] + "]"
	}
	return m, nil
}

func init() {
	pHost := req("host", "str", "Host name, domain or URL (the host part is used).")
	reg(kwDomain, tools.Func{
		Name: "domain_parts", Title: "Domain parts", Order: 9, Input: "host",
		Doc:     "Splits a host (or the host of a URL) using the public suffix list: public_suffix (e.g. co.uk), tld, registrable_domain (eTLD+1), domain label, subdomain, and whether the suffix is ICANN, private (e.g. github.io) or unknown. Unicode names are converted to punycode; IP addresses are rejected.",
		Params:  []tools.Param{pHost},
		Returns: "dict[str, str]", Example: `net.domain_parts("https://blog.example.co.uk/post")`,
		Tags: []string{"tld", "etld+1", "registrable domain", "subdomain", "public suffix", "second level domain", "sld", "apex domain", "root domain"},
		Accepts: func(s string) bool {
			_, err := domainParts(s)
			return err == nil && looksHostLike(strings.TrimSpace(s)) || strings.Contains(s, "://")
		},
		Fn: func(a tools.Args) (any, error) { return domainParts(a.Str("host")) },
	})
	reg(kwDomain, tools.Func{
		Name: "is_valid_hostname", Title: "Is valid hostname", Order: 110, Input: "host",
		Doc:     "Checks RFC 1123 hostname syntax: labels of 1-63 letters, digits and hyphens (no leading or trailing hyphen), at most 253 characters in total. A single label such as localhost is fine. allow_underscore also accepts _ (service names, DKIM).",
		Params:  []tools.Param{req("host", "str", "Hostname to check."), opt("allow_underscore", "bool", "Accept underscores in labels.", false)},
		Returns: "bool", Example: `net.is_valid_hostname("my-server.local")`,
		Tags: []string{"validate hostname", "rfc 1123", "dns name"},
		Fn:   func(a tools.Args) (any, error) { return validHostname(a.Str("host"), a.Bool("allow_underscore")), nil },
	})
	reg(kwDomain, tools.Func{
		Name: "is_valid_domain", Title: "Is valid domain name", Order: 111, Input: "host",
		Doc:     "Checks that the text is a plausible fully qualified domain name: valid hostname syntax, at least two labels, an alphabetic (or xn--) top-level label and not an IP address. It does not check that the TLD exists.",
		Params:  []tools.Param{req("host", "str", "Domain name to check.")},
		Returns: "bool", Example: `net.is_valid_domain("example.com")`,
		Tags: []string{"validate domain", "fqdn", "check domain"},
		Fn:   func(a tools.Args) (any, error) { return validDomain(a.Str("host")), nil },
	})
	reg(kwDomain, tools.Func{
		Name: "email_parse", Title: "Parse email address", Order: 112, Input: "email",
		Doc:     "Splits an email address into local, domain, base_local, +tag, registrable_domain and a normalized form (lower-case; Gmail/Googlemail addresses also lose dots and the +tag, googlemail.com becomes gmail.com). Accepts \"Name <a@b.c>\" and mailto: prefixes. Fails for invalid addresses.",
		Params:  []tools.Param{req("email", "str", "Email address.")},
		Returns: "dict[str, str]", Example: `net.email_parse("John.Doe+news@Gmail.com")`,
		Tags:    []string{"split email", "local part", "plus addressing", "subaddress", "gmail normalize", "dedupe email"},
		Accepts: func(s string) bool { return strings.Contains(s, "@") },
		Fn:      func(a tools.Args) (any, error) { return emailParse(a.Str("email")) },
	})
	reg(kwDomain, tools.Func{
		Name: "is_valid_email", Title: "Is valid email", Order: 113, Input: "email",
		Doc:     "Pragmatic RFC 5322 check: a dot-atom or quoted local part (max 64 chars), an @, and a hostname domain with at least two labels or an [IP literal]; total length up to 254. No display names, comments or whitespace. It does not verify the mailbox exists.",
		Params:  []tools.Param{req("email", "str", "Email address.")},
		Returns: "bool", Example: `net.is_valid_email("user@example.com")`,
		Tags: []string{"validate email", "check email", "email syntax"},
		Fn:   func(a tools.Args) (any, error) { return validEmail(a.Str("email")), nil },
	})
	reg(kwDomain, tools.Func{
		Name: "reverse_domain_name", Title: "Reverse domain name", Order: 114, Input: "host",
		Doc:     "Reverses the labels of a domain: a.b.com becomes com.b.a (Java package / Android application id style). A URL is reduced to its host first.",
		Params:  []tools.Param{req("host", "str", "Domain name.")},
		Returns: "str", Example: `net.reverse_domain_name("mail.example.com")`,
		Tags: []string{"package name", "bundle id", "reverse dns", "com.example"},
		Fn: func(a tools.Args) (any, error) {
			h := hostFromInput(a.Str("host"))
			if h == "" {
				return nil, fmt.Errorf("empty domain")
			}
			l := strings.Split(h, ".")
			for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
				l[i], l[j] = l[j], l[i]
			}
			return strings.Join(l, "."), nil
		},
	})
	reg(kwURL, tools.Func{
		Name: "basic_auth_header", Title: "Basic auth header", Order: 115, Input: "user",
		Doc:     "Builds an HTTP Basic Authorization header value: \"Basic \" plus base64(user:password). The user name must not contain a colon.",
		Params:  []tools.Param{req("user", "str", "User name (no colon)."), req("password", "str", "Password.")},
		Returns: "str", Example: `net.basic_auth_header("alice", "s3cret")`,
		Tags: []string{"authorization", "http auth", "credentials", "base64", "curl -u"},
		Fn: func(a tools.Args) (any, error) {
			if strings.Contains(a.Str("user"), ":") {
				return nil, fmt.Errorf("user name must not contain a colon")
			}
			return "Basic " + base64.StdEncoding.EncodeToString([]byte(a.Str("user")+":"+a.Str("password"))), nil
		},
	})
	reg(kwURL, tools.Func{
		Name: "parse_basic_auth", Title: "Parse Basic auth header", Order: 116, Input: "header",
		Doc:     "Decodes an HTTP Basic Authorization header (with or without the \"Basic \" prefix and \"Authorization:\" name) into user and password. Fails for invalid base64 or a missing colon.",
		Params:  []tools.Param{req("header", "str", "Header value such as Basic YWxpY2U6czNjcmV0.")},
		Returns: "dict[str, str]", Example: `net.parse_basic_auth("Basic YWxpY2U6czNjcmV0")`,
		Tags:    []string{"decode authorization", "credentials", "base64 decode"},
		Accepts: func(s string) bool { return strings.HasPrefix(strings.ToLower(strings.TrimSpace(s)), "basic ") },
		Fn: func(a tools.Args) (any, error) {
			h := strings.TrimSpace(a.Str("header"))
			if i := strings.Index(strings.ToLower(h), "authorization:"); i == 0 {
				h = strings.TrimSpace(h[len("authorization:"):])
			}
			if len(h) >= 6 && strings.EqualFold(h[:6], "basic ") {
				h = strings.TrimSpace(h[6:])
			}
			raw, err := base64.StdEncoding.DecodeString(h)
			if err != nil {
				raw, err = base64.RawStdEncoding.DecodeString(h)
			}
			if err != nil {
				return nil, fmt.Errorf("invalid base64 in Basic auth header")
			}
			u, p, ok := strings.Cut(string(raw), ":")
			if !ok {
				return nil, fmt.Errorf("decoded credentials have no colon")
			}
			return map[string]string{"user": u, "password": p}, nil
		},
	})
	reg(kwDomain, tools.Func{
		Name: "parse_host_port", Title: "Parse host:port", Order: 117, Input: "address",
		Doc:     "Splits host:port, including [IPv6]:port, bare IPv6 (no port), IPv4:port and name:port. Returns host, port (empty when absent), host_type (ipv4, ipv6, domain) and the bracketed IPv6 form. Fails for invalid hosts or ports above 65535.",
		Params:  []tools.Param{req("address", "str", "Address such as example.com:8080 or [::1]:443.")},
		Returns: "dict[str, str]", Example: `net.parse_host_port("[::1]:8080")`,
		Tags:    []string{"split host port", "socket address", "endpoint", "ipv6 port"},
		Accepts: func(s string) bool { _, err := parseHostPort(s); return err == nil && strings.Contains(s, ":") },
		Fn:      func(a tools.Args) (any, error) { return parseHostPort(a.Str("address")) },
	})
	reg(kwIP, tools.Func{
		Name: "hosts_entry", Title: "hosts file entry", Order: 118, Input: "ip",
		Doc:     "Formats a hosts-file line (ip, tab, space-separated names) after validating the address and host names, e.g. \"127.0.0.1\\tapp.local www.app.local\".",
		Params:  []tools.Param{pIP("IP address."), req("names", "list[str]", "Host names to map to the address.")},
		Returns: "str", Example: `net.hosts_entry("127.0.0.1", ["app.local", "www.app.local"])`,
		Tags: []string{"/etc/hosts", "windows hosts", "host mapping", "local dns"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			names := splitItems(a.List("names"))
			if len(names) == 0 {
				return nil, fmt.Errorf("give at least one host name")
			}
			for _, n := range names {
				if !validHostname(n, true) {
					return nil, fmt.Errorf("invalid host name %q", n)
				}
			}
			return ad.String() + "\t" + strings.Join(names, " "), nil
		},
	})
}
