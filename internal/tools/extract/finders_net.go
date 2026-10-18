package extract

import (
	"html"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"

	"github.com/avestura/zagros/internal/tools"
)

var (
	emailRe = regexp.MustCompile(`[\p{L}\p{N}._%+\-]+@(?:[\p{L}\p{N}](?:[\p{L}\p{N}\-]*[\p{L}\p{N}])?\.)+\p{L}{2,}`)
	urlRe   = regexp.MustCompile("(?i)(?:\\b(?:https?|ftp|wss?)://|\\bwww\\.)[^\\s<>\"`“”‘’]+")
	mdLink  = regexp.MustCompile(`(!?)\[((?:[^\[\]]|\[[^\]]*\])*)\]\(\s*(<[^>]*>|[^()\s]*(?:\([^()\s]*\)[^()\s]*)*)(?:\s+(?:"[^"]*"|'[^']*'))?\s*\)`)
	mdRef   = regexp.MustCompile(`(?m)^[ ]{0,3}\[[^\]]+\]:[ \t]*<?([^\s>]+)>?(?:[ \t]+.*)?$`)
	mdAuto  = regexp.MustCompile(`<((?:https?|ftp|mailto):[^\s<>]+)>`)
	tagRe   = regexp.MustCompile(`(?s)<([A-Za-z][A-Za-z0-9:\-]*)((?:\s+[^\s"'<>/=]+(?:\s*=\s*(?:"[^"]*"|'[^']*'|[^\s"'=<>` + "`" + `]+))?)*)\s*/?>`)
	attrRe  = regexp.MustCompile(`([^\s"'<>/=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+)))?`)
	xmlTag  = regexp.MustCompile(`<([A-Za-z_][\w.\-]*(?::[\w.\-]+)?)(?:\s[^<>]*)?/?>`)

	ipv4Re   = regexp.MustCompile(`\d{1,3}(?:\.\d{1,3}){3}`)
	v6CandRe = regexp.MustCompile(`[0-9A-Fa-f:.]{3,}`)
	cidr4Re  = regexp.MustCompile(`\d{1,3}(?:\.\d{1,3}){3}/\d{1,2}`)
	cidr6Re  = regexp.MustCompile(`[0-9A-Fa-f:]{2,}/\d{1,3}`)
	portRe   = regexp.MustCompile(`(?i)(?:\d{1,3}(?:\.\d{1,3}){3}|\[[0-9a-f:.]+\]|localhost|(?:[a-z0-9](?:[a-z0-9\-]*[a-z0-9])?\.)+[a-z]{2,}):(\d{1,5})`)
	macRe    = regexp.MustCompile(`[0-9A-Fa-f]{2}(?:[:\-][0-9A-Fa-f]{2}){5}|[0-9A-Fa-f]{4}(?:\.[0-9A-Fa-f]{4}){2}`)
	domainRe = regexp.MustCompile(`(?i)(?:\b[a-z][a-z0-9+.\-]*://(?:[^\s/?#@<>"']*@)?(\[[0-9a-f:.]+\]|[^\s/?#:<>"'\[\]]+))|((?:[\p{L}\p{N}](?:[\p{L}\p{N}\-]{0,61}[\p{L}\p{N}])?\.)+\p{L}{2,63})`)
)

func emailHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range emailRe.FindAllStringIndex(text, -1) {
		s, e := m[0], m[1]
		for s < e && strings.IndexByte(".-_%+", text[s]) >= 0 {
			s++
		}
		at := strings.IndexByte(text[s:e], '@')
		if at <= 0 || strings.HasSuffix(text[s:s+at], ".") || strings.Contains(text[s:s+at], "..") {
			continue
		}
		if strings.Contains(text[s+at:e], "..") {
			continue
		}
		v := text[s:e]
		if a.Bool("lower") {
			v = strings.ToLower(v)
		}
		out = append(out, hit{s, e, v})
	}
	return out
}

func urlHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range urlRe.FindAllStringIndex(text, -1) {
		v := trimTrailing(text[m[0]:m[1]])
		low := strings.ToLower(v)
		switch {
		case strings.HasPrefix(low, "www."):
			if len(v) <= 4 || !strings.Contains(v[4:], ".") {
				continue
			}
		default:
			i := strings.Index(v, "://")
			if i < 0 || len(v) <= i+3 || !isAlnum(rune(v[i+3])) && v[i+3] != '[' {
				continue
			}
		}
		out = append(out, hit{m[0], m[0] + len(v), v})
	}
	return out
}

func mdLinkHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range mdLink.FindAllStringSubmatchIndex(text, -1) {
		if m[3] > m[2] { // image
			continue
		}
		u := strings.Trim(text[m[6]:m[7]], "<>")
		if u == "" {
			continue
		}
		v := u
		if a.Bool("with_text") {
			v = text[m[4]:m[5]] + " -> " + u
		}
		out = append(out, hit{m[0], m[1], v})
	}
	for _, m := range mdRef.FindAllStringSubmatchIndex(text, -1) {
		out = append(out, hit{m[0], m[1], text[m[2]:m[3]]})
	}
	for _, m := range mdAuto.FindAllStringSubmatchIndex(text, -1) {
		out = append(out, hit{m[0], m[1], text[m[2]:m[3]]})
	}
	return out
}

// htmlAttrHits returns attribute values of tags: tags limits tag names (nil
// means all), attr is the attribute name (lower case).
func htmlAttrHits(text string, tags map[string]bool, attr string) []hit {
	var out []hit
	for _, m := range tagRe.FindAllStringSubmatchIndex(text, -1) {
		name := strings.ToLower(text[m[2]:m[3]])
		if tags != nil && !tags[name] {
			continue
		}
		if m[4] < 0 {
			continue
		}
		base := m[4]
		attrs := text[m[4]:m[5]]
		for _, am := range attrRe.FindAllStringSubmatchIndex(attrs, -1) {
			if !strings.EqualFold(attrs[am[2]:am[3]], attr) {
				continue
			}
			vs, ve := -1, -1
			for g := 2; g <= 4; g++ {
				if am[2*g] >= 0 {
					vs, ve = am[2*g], am[2*g+1]
				}
			}
			if vs < 0 {
				continue
			}
			v := strings.TrimSpace(html.UnescapeString(attrs[vs:ve]))
			if v == "" {
				continue
			}
			out = append(out, hit{base + vs, base + ve, v})
		}
	}
	return out
}

func tagNameHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range tagRe.FindAllStringSubmatchIndex(text, -1) {
		v := strings.ToLower(text[m[2]:m[3]])
		if a.Bool("full") {
			v = text[m[0]:m[1]]
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func validOctets(s string) bool {
	for _, p := range strings.Split(s, ".") {
		if len(p) > 1 && p[0] == '0' {
			return false
		}
		n, err := strconv.Atoi(p)
		if err != nil || n > 255 {
			return false
		}
	}
	return true
}

func isPrivateAddr(a netip.Addr) bool {
	if a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() || a.IsUnspecified() || a.IsMulticast() {
		return true
	}
	if a.Is4() {
		b := a.As4()
		return b[0] == 100 && b[1]&0xc0 == 64 || b[0] >= 240 || b[0] == 192 && b[1] == 0 && b[2] == 2 || b[0] == 198 && b[1] == 51 && b[2] == 100 || b[0] == 203 && b[1] == 0 && b[2] == 113
	}
	return false
}

func ipv4Hits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range ipv4Re.FindAllStringIndex(text, -1) {
		s, e := m[0], m[1]
		if !validOctets(text[s:e]) {
			continue
		}
		// part of a longer dotted number such as 1.2.3.4.5 or a version
		if p := prevRune(text, s); p == '.' && s >= 2 && text[s-2] >= '0' && text[s-2] <= '9' {
			continue
		}
		if nr := nextRune(text, e); nr == '.' && e+1 < len(text) && text[e+1] >= '0' && text[e+1] <= '9' {
			continue
		}
		if unicodeLetter(prevRune(text, s)) || unicodeLetter(nextRune(text, e)) {
			continue
		}
		ad, err := netip.ParseAddr(text[s:e])
		if err != nil {
			continue
		}
		if a.Bool("exclude_private") && isPrivateAddr(ad) {
			continue
		}
		out = append(out, hit{s, e, text[s:e]})
	}
	return out
}

func unicodeLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_' || r > 127 && isAlnum(r)
}

func ipv6Hits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range v6CandRe.FindAllStringIndex(text, -1) {
		s, e := m[0], m[1]
		if strings.Count(text[s:e], ":") < 2 {
			continue
		}
		if p := prevRune(text, s); isAlnum(p) || p == '_' || p == ':' {
			continue
		}
		// accept an IPv6 zone like %eth0 by ignoring it; trim until valid
		cand := text[s:e]
		for len(cand) >= 3 {
			if ad, err := netip.ParseAddr(cand); err == nil && ad.Is6() {
				break
			}
			last := cand[len(cand)-1]
			if last != '.' && last != ':' {
				cand = ""
				break
			}
			cand = cand[:len(cand)-1]
		}
		if len(cand) < 3 {
			continue
		}
		end := s + len(cand)
		if nr := nextRune(text, end); isAlnum(nr) || nr == '_' || (nr == ':' && cand[len(cand)-1] != ':') {
			continue
		}
		if cand == "::" {
			continue
		}
		ad, _ := netip.ParseAddr(cand)
		if a.Bool("exclude_private") && isPrivateAddr(ad) {
			continue
		}
		out = append(out, hit{s, end, cand})
	}
	return out
}

func cidrHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range cidr4Re.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isAlnum(p) || p == '.' {
			continue
		}
		if n := nextRune(text, m[1]); isAlnum(n) {
			continue
		}
		if pf, err := netip.ParsePrefix(text[m[0]:m[1]]); err == nil && pf.Addr().Is4() && validOctets(text[m[0]:m[0]+strings.IndexByte(text[m[0]:m[1]], '/')]) {
			out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
		}
	}
	for _, m := range cidr6Re.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isAlnum(p) || p == ':' {
			continue
		}
		if n := nextRune(text, m[1]); isAlnum(n) {
			continue
		}
		if pf, err := netip.ParsePrefix(text[m[0]:m[1]]); err == nil && pf.Addr().Is6() {
			out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
		}
	}
	return out
}

func portHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range portRe.FindAllStringSubmatchIndex(text, -1) {
		s, e := m[0], m[1]
		if p := prevRune(text, s); isAlnum(p) || p == '.' || p == '-' || p == '_' {
			continue
		}
		if n := nextRune(text, e); isAlnum(n) {
			continue
		}
		port, _ := strconv.Atoi(text[m[2]:m[3]])
		if port < 1 || port > 65535 {
			continue
		}
		host := text[s : m[2]-1]
		if strings.Count(host, ".") == 3 && host[0] >= '0' && host[0] <= '9' && !validOctets(host) {
			continue
		}
		out = append(out, hit{s, e, text[s:e]})
	}
	return out
}

func macHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range macRe.FindAllStringIndex(text, -1) {
		p, n := prevRune(text, m[0]), nextRune(text, m[1])
		if isAlnum(p) || p == ':' || p == '-' || p == '.' || isAlnum(n) || n == ':' || n == '-' {
			continue
		}
		out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
	}
	return out
}

type domHit struct {
	s, e int
	host string
}

// domainCore finds host names in URLs and bare domains with a public suffix.
func domainCore(text string) []domHit {
	var out []domHit
	for _, m := range domainRe.FindAllStringSubmatchIndex(text, -1) {
		s, e := m[0], m[1]
		var host string
		if m[2] >= 0 { // URL authority
			host = strings.ToLower(text[m[2]:m[3]])
			if strings.HasPrefix(host, "[") {
				continue
			}
			host = strings.TrimSuffix(host, ".")
			if host == "" || strings.Count(host, ".") == 3 && ipv4Re.MatchString(host) && validOctets(host) {
				continue
			}
			if !hostLooksValid(host) {
				continue
			}
		} else {
			if p := prevRune(text, s); isWordRune(p) || p == '-' || p == '.' {
				continue
			}
			if n := nextRune(text, e); isWordRune(n) || n == '-' {
				continue
			}
			host = strings.ToLower(text[m[4]:m[5]])
			ascii, err := idna.Lookup.ToASCII(host)
			if err != nil {
				continue
			}
			suffix, icann := publicsuffix.PublicSuffix(ascii)
			if !icann && !strings.Contains(suffix, ".") {
				continue
			}
			if suffix == ascii {
				continue
			}
		}
		out = append(out, domHit{s, e, host})
	}
	return out
}

func hostLooksValid(h string) bool {
	if len(h) > 253 {
		return false
	}
	for _, l := range strings.Split(h, ".") {
		if l == "" || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
			return false
		}
	}
	return true
}

func registrable(h string) string {
	ascii, err := idna.Lookup.ToASCII(h)
	if err != nil {
		return h
	}
	r, err := publicsuffix.EffectiveTLDPlusOne(ascii)
	if err != nil {
		return h
	}
	return r
}

func domainHits(text string, a tools.Args) []hit {
	var out []hit
	for _, d := range domainCore(text) {
		h := d.host
		if a.Bool("registrable_only") {
			h = registrable(h)
		} else if !a.Bool("include_subdomains") && registrable(h) != h {
			continue
		}
		h = strings.ToLower(h)
		out = append(out, hit{d.s, d.e, h})
	}
	return out
}

func tldHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, d := range domainCore(text) {
		if i := strings.LastIndexByte(d.host, '.'); i >= 0 {
			out = append(out, hit{d.s, d.e, d.host[i+1:]})
		}
	}
	return out
}

func netSpecs() []spec {
	return []spec{
		{name: "emails", title: "Extract emails", order: 2, keywords: []string{"emails"}, extra: []tools.Param{pLower},
			doc:     "Finds email addresses (pragmatic RFC 5322 subset, Unicode letters allowed). Trailing punctuation such as a final dot or comma is not included and \"mailto:\" prefixes are dropped. Addresses with consecutive dots or a dot before @ are skipped.",
			example: `extract.emails("Mail bob@example.com, or mailto:alice.smith@mail.example.org.")`,
			tags:    []string{"e-mail", "email addresses", "mailto", "contacts", "harvest"}, find: emailHits},
		{name: "urls", title: "Extract URLs", order: 3, keywords: []string{"urls"},
			doc:     "Finds http, https, ftp, ws and wss URLs and bare www. links. Trailing sentence punctuation and unbalanced closing brackets are trimmed, so (see https://a.com/x) yields https://a.com/x.",
			example: `extract.urls("see https://example.com/a?b=1, and www.test.org.")`,
			tags:    []string{"links", "hyperlinks", "http", "https", "web addresses", "ftp", "href"}, find: urlHits},
		{name: "links_markdown", title: "Extract Markdown links", order: 30,
			doc:     "Returns the targets of Markdown links [text](url), reference definitions ([id]: url) and <autolinks>. Images ![alt](src) are skipped. With with_text each result is \"text -> url\".",
			example: `extract.links_markdown("[Go](https://go.dev) and ![img](a.png)")`, extra: []tools.Param{opt("with_text", "bool", `Return "text -> url" instead of only the url.`, false)},
			tags: []string{"markdown", "md links", "anchor text", "reference links"}, find: mdLinkHits, noAll: true},
		{name: "links_html", title: "Extract HTML links (href)", order: 31,
			doc:     "Returns href values from <a>, <area>, <link> and <base> tags in HTML, with HTML entities decoded (&amp; becomes &). Single, double and unquoted attribute values are supported.",
			example: `extract.links_html("<a href='/x?a=1&amp;b=2'>x</a>")`,
			tags:    []string{"href", "anchors", "html links", "a tag", "scrape links"},
			find: func(t string, _ tools.Args) []hit {
				return htmlAttrHits(t, map[string]bool{"a": true, "area": true, "link": true, "base": true}, "href")
			}},
		{name: "images_html", title: "Extract HTML image sources", order: 32,
			doc:     "Returns the src values of <img> tags in HTML (entities decoded). srcset and CSS background images are not parsed.",
			example: `extract.images_html("<img src=\"a.png\"> <img src='b.jpg'>")`,
			tags:    []string{"img", "src", "pictures", "image urls", "scrape images"},
			find: func(t string, _ tools.Args) []hit {
				return htmlAttrHits(t, map[string]bool{"img": true}, "src")
			}},
		{name: "html_tags", title: "Extract HTML tags", order: 33, extra: []tools.Param{opt("full", "bool", "Return whole opening tags instead of names.", false)},
			doc:     "Lists the HTML opening tags found, as lower-case tag names (div, a, img...) or, with full, the complete opening tags including attributes. Comments, doctype and closing tags are ignored.",
			example: `extract.html_tags("<div class='x'><p>hi</p></div>")`,
			tags:    []string{"tags", "elements", "markup", "html elements"}, find: tagNameHits, noAll: true},
		{name: "html_attributes", title: "Extract HTML attribute values", order: 34, extra: []tools.Param{opt("name", "str", "Attribute name such as class, id, alt, data-id.", "class")},
			doc:     "Returns the values of one attribute (name, default class) from every HTML tag, e.g. all id or alt values. Entities are decoded; attributes without a value are skipped.",
			example: `extract.html_attributes("<p id='a' class='x y'>", "class")`,
			tags:    []string{"attribute", "class", "id", "alt", "title", "data attributes"},
			find: func(t string, a tools.Args) []hit {
				return htmlAttrHits(t, nil, strings.ToLower(a.Str("name")))
			}, noAll: true},
		{name: "xml_tags", title: "Extract XML element names", order: 35,
			doc:     "Lists XML element names (with namespace prefix, case preserved) from opening and self-closing tags. Comments, processing instructions, CDATA and closing tags are ignored.",
			example: `extract.xml_tags("<root><ns:item id='1'/></root>")`,
			tags:    []string{"xml", "elements", "svg", "tag names", "soap"},
			find: func(t string, _ tools.Args) []hit {
				var out []hit
				for _, m := range xmlTag.FindAllStringSubmatchIndex(t, -1) {
					out = append(out, hit{m[0], m[1], t[m[2]:m[3]]})
				}
				return out
			}, noAll: true},
		{name: "domains", title: "Extract domains", order: 4, keywords: []string{"domains"},
			extra:   []tools.Param{opt("include_subdomains", "bool", "Keep hosts that have subdomains; false keeps only apex (registrable) domains.", true), opt("registrable_only", "bool", "Reduce every host to its registrable domain (eTLD+1), e.g. a.b.example.co.uk -> example.co.uk.", false)},
			doc:     "Finds host names in URLs and bare domains whose suffix is in the public suffix list (so file names like notes.txt are ignored, but readme.md matches because .md is Moldova's TLD). IP addresses are skipped. Results are lower-case; Unicode names are kept as written.",
			example: `extract.domains("visit https://blog.example.co.uk/x or docs.python.org", registrable_only=True)`,
			tags:    []string{"hostnames", "fqdn", "websites", "sites", "tld", "registrable domain", "etld+1"}, find: domainHits},
		{name: "tlds", title: "Extract TLDs", order: 36,
			doc:     "Returns the top-level domain (last label, e.g. com, uk) of every domain or URL host found, using the same detection as extract.domains.",
			example: `extract.tlds("a.example.com b.example.co.uk")`,
			tags:    []string{"top level domain", "extensions", "country codes", "suffix"}, find: tldHits, noAll: true},
		{name: "ipv4", title: "Extract IPv4 addresses", order: 8, extra: []tools.Param{opt("exclude_private", "bool", "Skip private, loopback, link-local, CGNAT, documentation and multicast addresses.", false)},
			doc:     "Finds dotted-quad IPv4 addresses. Octets must be 0-255 without leading zeros; parts of longer dotted numbers (1.2.3.4.5, version strings) are ignored. exclude_private keeps only publicly routable addresses.",
			example: `extract.ipv4("from 10.0.0.1 to 8.8.8.8", exclude_private=True)`,
			tags:    []string{"ip address", "ipv4", "dotted quad", "log analysis", "public ip"}, find: ipv4Hits},
		{name: "ipv6", title: "Extract IPv6 addresses", order: 9, extra: []tools.Param{opt("exclude_private", "bool", "Skip loopback, link-local, ULA and multicast addresses.", false)},
			doc:     "Finds valid IPv6 addresses including :: compression and IPv4-mapped forms. Candidates are verified by a real parser; MAC addresses, times and C++ scope operators like std::vector are not reported.",
			example: `extract.ipv6("host 2001:db8::1 and ::1")`,
			tags:    []string{"ip address", "ipv6", "colon hex", "log analysis"}, find: ipv6Hits},
		{name: "ips", title: "Extract IP addresses", order: 5, keywords: []string{"ips"}, extra: []tools.Param{opt("exclude_private", "bool", "Skip private, loopback, link-local and similar non-public addresses.", false)},
			doc:     "Finds both IPv4 and IPv6 addresses in text order. See extract.ipv4 and extract.ipv6 for the validation rules; exclude_private keeps only public addresses.",
			example: `extract.ips("a 1.2.3.4 b 2001:db8::1")`,
			tags:    []string{"ip addresses", "ipv4", "ipv6", "addresses", "log analysis", "public ip"},
			find:    func(t string, a tools.Args) []hit { return append(ipv4Hits(t, a), ipv6Hits(t, a)...) }},
		{name: "cidrs", title: "Extract CIDR blocks", order: 10,
			doc:     "Finds IPv4 and IPv6 CIDR blocks such as 10.0.0.0/8 or 2001:db8::/32. The address and the prefix length (0-32 / 0-128) are validated.",
			example: `extract.cidrs("allow 10.0.0.0/8 and 2001:db8::/32")`,
			tags:    []string{"subnets", "networks", "ip ranges", "prefix", "acl"}, find: cidrHits},
		{name: "ports", title: "Extract host:port pairs", order: 11,
			doc:     "Finds host:port endpoints where the host is an IPv4 address, [IPv6], localhost or a domain name and the port is 1-65535 (example.com:8080, 10.0.0.1:22, [::1]:443). Times like 12:30 are not matched.",
			example: `extract.ports("connect to 10.0.0.5:8080 and localhost:3000")`,
			tags:    []string{"endpoints", "sockets", "host port", "address:port", "netstat"}, find: portHits},
		{name: "mac_addresses", title: "Extract MAC addresses", order: 12,
			doc:     "Finds MAC addresses written as aa:bb:cc:dd:ee:ff, aa-bb-cc-dd-ee-ff or Cisco aabb.ccdd.eeff. Longer hex/colon sequences (IPv6, EUI-64) are not reported.",
			example: `extract.mac_addresses("eth0 00:1A:2B:3C:4D:5E")`,
			tags:    []string{"mac", "hardware address", "ethernet", "arp", "physical address"}, find: macHits},
	}
}
