package netx

import (
	"fmt"
	"math/big"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

type ipClassEntry struct {
	prefix netip.Prefix
	kind   string
	desc   string
	global bool
}

func ce(p, kind, desc string, global bool) ipClassEntry {
	return ipClassEntry{netip.MustParsePrefix(p), kind, desc, global}
}

var ipClasses = []ipClassEntry{
	ce("0.0.0.0/32", "unspecified", "unspecified address (RFC 1122)", false),
	ce("0.0.0.0/8", "this-network", `"this network" (RFC 1122)`, false),
	ce("10.0.0.0/8", "private", "private use (RFC 1918)", false),
	ce("100.64.0.0/10", "cgnat", "carrier-grade NAT shared address space (RFC 6598)", false),
	ce("127.0.0.0/8", "loopback", "loopback (RFC 1122)", false),
	ce("169.254.0.0/16", "link-local", "link-local, APIPA (RFC 3927)", false),
	ce("172.16.0.0/12", "private", "private use (RFC 1918)", false),
	ce("192.0.0.0/24", "ietf-protocol", "IETF protocol assignments (RFC 6890)", false),
	ce("192.0.2.0/24", "documentation", "documentation TEST-NET-1 (RFC 5737)", false),
	ce("192.88.99.0/24", "6to4-relay", "6to4 relay anycast, deprecated (RFC 7526)", true),
	ce("192.168.0.0/16", "private", "private use (RFC 1918)", false),
	ce("198.18.0.0/15", "benchmark", "network benchmark testing (RFC 2544)", false),
	ce("198.51.100.0/24", "documentation", "documentation TEST-NET-2 (RFC 5737)", false),
	ce("203.0.113.0/24", "documentation", "documentation TEST-NET-3 (RFC 5737)", false),
	ce("224.0.0.0/4", "multicast", "multicast (RFC 5771)", true),
	ce("240.0.0.0/4", "reserved", "reserved for future use, class E (RFC 1112)", false),
	ce("255.255.255.255/32", "broadcast", "limited broadcast (RFC 919)", false),

	ce("::/128", "unspecified", "unspecified address (RFC 4291)", false),
	ce("::1/128", "loopback", "loopback (RFC 4291)", false),
	ce("::ffff:0:0/96", "ipv4-mapped", "IPv4-mapped IPv6 address (RFC 4291)", false),
	ce("64:ff9b::/96", "nat64", "NAT64 well-known prefix (RFC 6052)", true),
	ce("64:ff9b:1::/48", "nat64", "local-use NAT64 prefix (RFC 8215)", false),
	ce("100::/64", "discard", "discard-only prefix (RFC 6666)", false),
	ce("2001::/32", "teredo", "Teredo tunneling (RFC 4380)", false),
	ce("2001:2::/48", "benchmark", "benchmarking (RFC 5180)", false),
	ce("2001:db8::/32", "documentation", "documentation (RFC 3849)", false),
	ce("2001:10::/28", "orchid", "deprecated ORCHID (RFC 4843)", false),
	ce("2002::/16", "6to4", "6to4 tunneling (RFC 3056)", true),
	ce("3fff::/20", "documentation", "documentation (RFC 9637)", false),
	ce("2000::/3", "public", "global unicast", true),
	ce("fc00::/7", "ula", "unique local address, the IPv6 private range (RFC 4193)", false),
	ce("fe80::/10", "link-local", "link-local unicast (RFC 4291)", false),
	ce("fec0::/10", "site-local", "deprecated site-local (RFC 3879)", false),
	ce("ff00::/8", "multicast", "multicast (RFC 4291)", true),
}

func lookupClass(a netip.Addr) (ipClassEntry, bool) {
	a = a.WithZone("")
	best := -1
	for i, e := range ipClasses {
		if e.prefix.Addr().Is4() != a.Is4() || !e.prefix.Contains(a) {
			continue
		}
		if best < 0 || e.prefix.Bits() > ipClasses[best].prefix.Bits() {
			best = i
		}
	}
	if best < 0 {
		return ipClassEntry{}, false
	}
	return ipClasses[best], true
}

func classifyAddr(a netip.Addr) string {
	if e, ok := lookupClass(a); ok {
		return e.kind
	}
	if a.Is4() {
		return "public"
	}
	return "reserved"
}

func classifyDetail(a netip.Addr) map[string]string {
	e, ok := lookupClass(a)
	if !ok {
		e = ipClassEntry{kind: "public", desc: "globally routable unicast", global: true}
		if !a.Is4() {
			e = ipClassEntry{kind: "reserved", desc: "reserved or unassigned IPv6 space", global: false}
		}
	}
	m := map[string]string{
		"ip": a.String(), "version": map[bool]string{true: "4", false: "6"}[a.Is4()],
		"type": e.kind, "description": e.desc,
		"is_global":    strconv.FormatBool(e.global),
		"is_private":   strconv.FormatBool(e.kind == "private" || e.kind == "ula" || e.kind == "cgnat"),
		"is_loopback":  strconv.FormatBool(e.kind == "loopback"),
		"is_multicast": strconv.FormatBool(e.kind == "multicast"),
		"is_link_local": strconv.FormatBool(e.kind == "link-local" ||
			(a.Is4() && a.As4()[0] == 224 && a.As4()[1] == 0 && a.As4()[2] == 0)),
		"is_unspecified": strconv.FormatBool(a.IsUnspecified()),
	}
	switch e.kind {
	case "multicast":
		if !a.Is4() {
			names := map[byte]string{1: "interface-local", 2: "link-local", 3: "realm-local", 4: "admin-local", 5: "site-local", 8: "organization-local", 0xe: "global"}
			sc := a.As16()[1] & 0xf
			if n, ok := names[sc]; ok {
				m["scope"] = n
			} else {
				m["scope"] = "reserved"
			}
		} else if b := a.As4(); b[0] == 224 && b[1] == 0 && b[2] == 0 {
			m["scope"] = "link-local"
		} else {
			m["scope"] = "global"
		}
	case "private", "ula", "cgnat", "site-local":
		m["scope"] = "private"
	case "loopback", "unspecified", "this-network":
		m["scope"] = "host"
	case "link-local":
		m["scope"] = "link"
	case "public":
		m["scope"] = "global"
	default:
		m["scope"] = "special"
	}
	if e.kind == "ipv4-mapped" {
		m["embedded_ipv4"] = a.Unmap().String()
	}
	return m
}

// parseLoose also accepts decimal / hex / octal integer forms and short
// inet_aton forms like "127.1" on top of everything parseAddr accepts.
func parseLoose(s string) (netip.Addr, error) {
	if a, err := parseAddr(s); err == nil {
		return a, nil
	}
	t := strings.TrimSpace(s)
	if t == "" || strings.ContainsAny(t, ":/ ") {
		return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
	}
	parts := strings.Split(t, ".")
	if len(parts) > 4 {
		return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
	}
	vals := make([]uint64, len(parts))
	for i, p := range parts {
		v, err := strconv.ParseUint(p, 0, 64)
		if err != nil {
			if q := strings.TrimLeft(p, "0"); q != p && q != "" { // 0-prefixed octal like 0300
				v, err = strconv.ParseUint("0"+q, 0, 64)
			}
			if err != nil {
				return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
			}
		}
		vals[i] = v
	}
	var n uint64
	for i, v := range vals[:len(vals)-1] {
		if v > 255 {
			return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
		}
		n |= v << uint(8*(3-i))
	}
	last := vals[len(vals)-1]
	if last >= 1<<uint(8*(4-len(vals)+1)) {
		return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
	}
	n |= last
	return netip.AddrFrom4([4]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}), nil
}

func expandV6(a netip.Addr) string {
	b := a.As16()
	parts := make([]string, 8)
	for i := range parts {
		parts[i] = fmt.Sprintf("%02x%02x", b[2*i], b[2*i+1])
	}
	return strings.Join(parts, ":")
}

func binaryIP(a netip.Addr) string {
	b := a.AsSlice()
	if a.Is4() {
		parts := make([]string, 4)
		for i, x := range b {
			parts[i] = fmt.Sprintf("%08b", x)
		}
		return strings.Join(parts, ".")
	}
	parts := make([]string, 8)
	for i := range parts {
		parts[i] = fmt.Sprintf("%08b%08b", b[2*i], b[2*i+1])
	}
	return strings.Join(parts, ":")
}

func hexIP(a netip.Addr) string {
	return "0x" + fmt.Sprintf("%x", a.AsSlice())
}

func octalIP(a netip.Addr) string {
	b := a.As4()
	return fmt.Sprintf("%04o.%04o.%04o.%04o", b[0], b[1], b[2], b[3])
}

func v4In6(a netip.Addr) netip.Addr {
	b := a.As4()
	var s [16]byte
	s[10], s[11] = 0xff, 0xff
	copy(s[12:], b[:])
	return netip.AddrFrom16(s)
}

func ipFormats(s string) (map[string]string, error) {
	a, err := parseLoose(s)
	if err != nil {
		return nil, err
	}
	m := map[string]string{"version": map[bool]string{true: "4", false: "6"}[a.Is4()], "decimal": bigOf(a).String(), "hex": hexIP(a), "binary": binaryIP(a), "ptr": ptrName(a)}
	if a.Is4() {
		m["dotted"] = a.String()
		m["octal"] = octalIP(a)
		m["ipv6_mapped"] = v4In6(a).String()
		m["ipv6_mapped_expanded"] = expandV6(v4In6(a))
		m["url_decimal"] = "http://" + m["decimal"] + "/"
		m["url_hex"] = "http://" + m["hex"] + "/"
		m["url_octal"] = "http://" + octalIP(a) + "/"
		m["url_dotted_hex"] = fmt.Sprintf("http://0x%02x.0x%02x.0x%02x.0x%02x/", a.As4()[0], a.As4()[1], a.As4()[2], a.As4()[3])
	} else {
		m["compressed"] = a.String()
		m["expanded"] = expandV6(a)
		m["url"] = "http://[" + a.String() + "]/"
		if a.Is4In6() {
			m["ipv4"] = a.Unmap().String()
		}
	}
	return m, nil
}

func macBytes(s string) ([6]byte, error) {
	var out [6]byte
	var hex strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			hex.WriteRune(r)
		case r == ':' || r == '-' || r == '.' || r == ' ' || r == '_':
		default:
			return out, fmt.Errorf("invalid MAC address %q", s)
		}
	}
	h := hex.String()
	if len(h) != 12 {
		return out, fmt.Errorf("invalid MAC address %q: need 12 hex digits, got %d", s, len(h))
	}
	for i := 0; i < 6; i++ {
		v, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
		out[i] = byte(v)
	}
	return out, nil
}

func eui64Addr(prefix [8]byte, mac [6]byte) netip.Addr {
	var b [16]byte
	copy(b[:8], prefix[:])
	b[8] = mac[0] ^ 0x02
	b[9], b[10], b[11], b[12] = mac[1], mac[2], 0xff, 0xfe
	b[13], b[14], b[15] = mac[3], mac[4], mac[5]
	return netip.AddrFrom16(b)
}

var nat64Idx = map[int][4]int{32: {4, 5, 6, 7}, 40: {5, 6, 7, 9}, 48: {6, 7, 9, 10}, 56: {7, 9, 10, 11}, 64: {9, 10, 11, 12}, 96: {12, 13, 14, 15}}

func nat64Prefix(s string) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "/") {
		s += "/96"
	}
	p, err := netip.ParsePrefix(s)
	if err != nil || p.Addr().Is4() {
		return p, fmt.Errorf("invalid NAT64 prefix %q", s)
	}
	if _, ok := nat64Idx[p.Bits()]; !ok {
		return p, fmt.Errorf("NAT64 prefix length must be 32, 40, 48, 56, 64 or 96")
	}
	return p.Masked(), nil
}

func parseInt(s string) (*big.Int, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, "_", ""))
	n, ok := new(big.Int).SetString(s, 0)
	return n, ok
}

func sortAddrs(addrs []netip.Addr, desc bool) {
	sort.SliceStable(addrs, func(i, j int) bool {
		c := addrs[i].Compare(addrs[j])
		if desc {
			return c > 0
		}
		return c < 0
	})
}

func init() {
	one := func(name string) tools.Param { return pIP(name) }
	reg(kwIP, tools.Func{
		Name: "ip_formats", Title: "IP address formats", Order: 2,
		Doc:     "Shows an IP address in every common notation: dotted, decimal integer, hex, octal, binary, IPv4-mapped IPv6, and the obfuscated URL forms (http://3232235777/, 0x..., octal) used in phishing and SSRF tests. IPv6 gives compressed, expanded, decimal, hex and binary forms. Also accepts decimal, hex (0x...) and short (127.1) inputs.",
		Params:  []tools.Param{one("IPv4 or IPv6 address (also decimal, 0x hex or short forms).")},
		Returns: "dict[str, str]", Example: `net.ip_formats("192.168.1.1")`,
		Tags:    []string{"convert ip", "decimal ip", "hex ip", "octal ip", "binary ip", "obfuscated url", "ip notation", "ssrf"},
		Accepts: func(s string) bool { _, err := parseLoose(s); return err == nil && !isPlainSmallInt(s) },
		Fn:      func(a tools.Args) (any, error) { return ipFormats(a.Str("ip")) },
	})
	reg(kwIP, tools.Func{
		Name: "ip_to_int", Title: "IP to integer", Order: 3,
		Doc:     "Converts an IPv4 or IPv6 address to its decimal integer value, returned as a string because IPv6 values exceed 64 bits (192.168.1.1 gives 3232235777).",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "str", Example: `net.ip_to_int("192.168.1.1")`,
		Tags: []string{"ip to decimal", "long", "integer", "numeric ip", "inet_aton", "bigint"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return bigOf(ad).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "int_to_ip", Title: "Integer to IP", Order: 40, Input: "number",
		Doc:     "Converts a non-negative integer (decimal string, or 0x hex) to an IP address. version 4 or 6 forces the family; the default 0 picks IPv4 when the number fits in 32 bits and IPv6 otherwise.",
		Params:  []tools.Param{req("number", "str", `Integer as a string, e.g. "3232235777" or "0xC0A80101".`), opt("version", "int", "0 (auto), 4 or 6.", int64(0))},
		Returns: "str", Example: `net.int_to_ip("3232235777")`,
		Tags:    []string{"decimal to ip", "long to ip", "inet_ntoa", "numeric to ip"},
		Accepts: func(s string) bool { n, ok := parseInt(s); return ok && n.Sign() >= 0 },
		Fn: func(a tools.Args) (any, error) {
			n, ok := parseInt(a.Str("number"))
			if !ok || n.Sign() < 0 {
				return nil, fmt.Errorf("invalid non-negative integer %q", a.Str("number"))
			}
			v := a.Int("version")
			if v != 0 && v != 4 && v != 6 {
				return nil, fmt.Errorf("version must be 0, 4 or 6")
			}
			is4 := v == 4 || (v == 0 && n.BitLen() <= 32)
			ad, ok := addrOf(n, is4)
			if !ok {
				return nil, fmt.Errorf("number does not fit in an IPv%d address", map[bool]int{true: 4, false: 6}[is4])
			}
			return ad.String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_to_hex", Title: "IP to hex", Order: 41,
		Doc:     "Converts an IP address to hexadecimal with a 0x prefix (192.168.1.1 gives 0xc0a80101; IPv6 gives 32 hex digits).",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "str", Example: `net.ip_to_hex("192.168.1.1")`,
		Tags: []string{"hexadecimal", "ip hex"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return hexIP(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_to_binary", Title: "IP to binary", Order: 42,
		Doc:     "Writes an IP address in binary: IPv4 as four dotted 8-bit groups, IPv6 as eight colon-separated 16-bit groups.",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "str", Example: `net.ip_to_binary("192.168.1.1")`,
		Tags: []string{"bits", "binary ip", "octets binary"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return binaryIP(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "binary_to_ip", Title: "Binary to IP", Order: 43, Input: "bits",
		Doc:     "Converts 32 binary digits (IPv4) or 128 binary digits (IPv6) to an address. Dots, colons and spaces between groups are ignored.",
		Params:  []tools.Param{req("bits", "str", "Binary digits, e.g. 11000000.10101000.00000001.00000001.")},
		Returns: "str", Example: `net.binary_to_ip("11000000.10101000.00000001.00000001")`,
		Tags:    []string{"from binary", "bits to ip"},
		Accepts: func(s string) bool { c := cleanBits(s); return (len(c) == 32 || len(c) == 128) && onlyBits(c) },
		Fn: func(a tools.Args) (any, error) {
			c := cleanBits(a.Str("bits"))
			if !onlyBits(c) || (len(c) != 32 && len(c) != 128) {
				return nil, fmt.Errorf("need exactly 32 or 128 binary digits, got %d", len(c))
			}
			n, _ := new(big.Int).SetString(c, 2)
			ad, _ := addrOf(n, len(c) == 32)
			return ad.String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "hex_to_ip", Title: "Hex to IP", Order: 44, Input: "hex",
		Doc:     "Converts 8 hex digits (IPv4) or 32 hex digits (IPv6) to an address. A 0x prefix and separators (. : space) are ignored.",
		Params:  []tools.Param{req("hex", "str", "Hex digits such as C0A80101 or 0xc0a80101.")},
		Returns: "str", Example: `net.hex_to_ip("0xC0A80101")`,
		Tags:    []string{"from hex", "hex address"},
		Accepts: func(s string) bool { _, err := hexToAddr(s); return err == nil },
		Fn: func(a tools.Args) (any, error) {
			ad, err := hexToAddr(a.Str("hex"))
			return ad.String(), err
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_version", Title: "IP version", Order: 45,
		Doc:     "Returns 4 for an IPv4 address and 6 for an IPv6 address; fails for anything else. IPv4-mapped addresses (::ffff:1.2.3.4) report 6.",
		Params:  []tools.Param{one("Address to inspect.")},
		Returns: "int", Example: `net.ip_version("2001:db8::1")`,
		Tags: []string{"ipv4 or ipv6", "family", "v4", "v6"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			if ad.Is4() {
				return int64(4), nil
			}
			return int64(6), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_type", Title: "IP address type", Order: 4,
		Doc:     "Classifies an address: private (RFC 1918), cgnat (100.64/10), loopback, link-local, multicast, unspecified, documentation, benchmark, reserved, broadcast, 6to4, teredo, nat64, ula (IPv6 private), ipv4-mapped or public.",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "str", Example: `net.ip_type("100.64.0.1")`,
		Tags: []string{"is private", "is public", "rfc1918", "cgnat", "loopback", "link-local", "multicast", "ula", "documentation", "benchmark", "classify ip", "internal ip"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return classifyAddr(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_classify", Title: "IP classification details", Order: 46,
		Doc:     "Detailed classification of an address as a dict: type, description with RFC, scope, and is_global / is_private / is_loopback / is_multicast / is_link_local / is_unspecified flags (as \"true\"/\"false\"). Multicast addresses include their scope; IPv4-mapped ones include the embedded IPv4.",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "dict[str, str]", Example: `net.ip_classify("fe80::1")`,
		Tags: []string{"classify", "ip info", "scope", "is global", "routable", "ip properties"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return classifyDetail(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_normalize", Title: "Normalize IP", Order: 47,
		Doc:     "Returns the canonical text form of an address: IPv6 is lower-case and compressed (RFC 5952), IPv4 octets lose leading zeros. Also accepts [brackets], zones, decimal/hex integers and short forms like 127.1.",
		Params:  []tools.Param{one("Address in any common spelling.")},
		Returns: "str", Example: `net.ip_normalize("2001:0DB8:0000:0000:0000:0000:0000:0001")`,
		Tags:    []string{"canonical", "clean", "format ip", "rfc 5952", "leading zeros", "shorten ipv6"},
		Accepts: func(s string) bool { _, err := parseLoose(s); return err == nil && !isPlainSmallInt(s) },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseLoose(a.Str("ip"))
			return ad.String(), err
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv6_expand", Title: "Expand IPv6", Order: 48,
		Doc:     "Expands an IPv6 address to the full eight 4-digit groups (2001:db8::1 becomes 2001:0db8:0000:0000:0000:0000:0000:0001). IPv4 input is returned as its IPv4-mapped expanded form.",
		Params:  []tools.Param{one("IPv6 address.")},
		Returns: "str", Example: `net.ipv6_expand("2001:db8::1")`,
		Tags:    []string{"full form", "uncompress", "long form", "zero fill"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is6() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			if ad.Is4() {
				ad = v4In6(ad)
			}
			return expandV6(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv6_compress", Title: "Compress IPv6", Order: 49,
		Doc:     "Compresses an IPv6 address to the shortest RFC 5952 form: leading zeros dropped and the longest run of zero groups replaced by ::. IPv4-mapped addresses keep their dotted tail.",
		Params:  []tools.Param{one("IPv6 address.")},
		Returns: "str", Example: `net.ipv6_compress("2001:0db8:0000:0000:0000:0000:0000:0001")`,
		Tags:    []string{"shorten", "abbreviate", "short form", "zero compression"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is6() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			if ad.Is4() {
				return nil, fmt.Errorf("%s is an IPv4 address", ad)
			}
			return ad.String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv4_to_ipv6_mapped", Title: "IPv4 to IPv4-mapped IPv6", Order: 50,
		Doc:     "Maps an IPv4 address into IPv6 as ::ffff:a.b.c.d (RFC 4291).",
		Params:  []tools.Param{one("IPv4 address.")},
		Returns: "str", Example: `net.ipv4_to_ipv6_mapped("192.0.2.1")`,
		Tags:    []string{"::ffff", "mapped address", "dual stack", "v4 in v6"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is4() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil || !ad.Is4() {
				return nil, fmt.Errorf("%q is not an IPv4 address", a.Str("ip"))
			}
			return v4In6(ad).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv6_to_ipv4_mapped", Title: "IPv4-mapped IPv6 to IPv4", Order: 51,
		Doc:     "Extracts the IPv4 address from an IPv4-mapped IPv6 address (::ffff:192.0.2.1 gives 192.0.2.1). Fails for IPv6 addresses that are not IPv4-mapped.",
		Params:  []tools.Param{one("IPv4-mapped IPv6 address.")},
		Returns: "str", Example: `net.ipv6_to_ipv4_mapped("::ffff:c000:201")`,
		Tags:    []string{"unmap", "extract ipv4", "::ffff"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is4In6() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			if !ad.Is4In6() {
				return nil, fmt.Errorf("%s is not an IPv4-mapped IPv6 address", ad)
			}
			return ad.Unmap().String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv4_to_6to4", Title: "IPv4 to 6to4 prefix", Order: 52,
		Doc:     "Derives the 6to4 IPv6 prefix (RFC 3056) of an IPv4 address: 192.0.2.4 gives 2002:c000:204:: (the /48 prefix).",
		Params:  []tools.Param{one("IPv4 address.")},
		Returns: "str", Example: `net.ipv4_to_6to4("192.0.2.4")`,
		Tags:    []string{"6to4", "2002", "tunnel", "transition"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is4() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil || !ad.Is4() {
				return nil, fmt.Errorf("%q is not an IPv4 address", a.Str("ip"))
			}
			b := ad.As4()
			var s [16]byte
			s[0], s[1] = 0x20, 0x02
			copy(s[2:6], b[:])
			return netip.AddrFrom16(s).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "sixtofour_to_ipv4", Title: "6to4 to IPv4", Order: 53,
		Doc:     "Extracts the embedded IPv4 address from a 6to4 address (2002:c000:204::1 gives 192.0.2.4). Fails when the address is not inside 2002::/16.",
		Params:  []tools.Param{one("6to4 IPv6 address.")},
		Returns: "str", Example: `net.sixtofour_to_ipv4("2002:c000:204::1")`,
		Tags: []string{"6to4", "2002", "extract ipv4"},
		Accepts: func(s string) bool {
			a, err := parseAddr(s)
			return err == nil && a.Is6() && !a.Is4In6() && a.As16()[0] == 0x20 && a.As16()[1] == 2
		},
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			b := ad.As16()
			if ad.Is4() || b[0] != 0x20 || b[1] != 0x02 {
				return nil, fmt.Errorf("%s is not a 6to4 address (2002::/16)", ad)
			}
			return netip.AddrFrom4([4]byte{b[2], b[3], b[4], b[5]}).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "nat64", Title: "IPv4 to NAT64 address", Order: 54,
		Doc:     "Synthesizes the NAT64 IPv6 address for an IPv4 address (RFC 6052). The default prefix is the well-known 64:ff9b::/96; /32, /40, /48, /56, /64 and /96 prefixes are supported.",
		Params:  []tools.Param{one("IPv4 address."), opt("prefix", "str", "NAT64 prefix with length, default 64:ff9b::/96.", "64:ff9b::/96")},
		Returns: "str", Example: `net.nat64("192.0.2.33")`,
		Tags:    []string{"dns64", "64:ff9b", "ipv6 translation", "rfc6052"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is4() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil || !ad.Is4() {
				return nil, fmt.Errorf("%q is not an IPv4 address", a.Str("ip"))
			}
			p, err := nat64Prefix(a.Str("prefix"))
			if err != nil {
				return nil, err
			}
			b := p.Addr().As16()
			v := ad.As4()
			for i, idx := range nat64Idx[p.Bits()] {
				b[idx] = v[i]
			}
			return netip.AddrFrom16(b).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "nat64_to_ipv4", Title: "NAT64 address to IPv4", Order: 55,
		Doc:     "Extracts the IPv4 address embedded in a NAT64 IPv6 address (RFC 6052). The default prefix is 64:ff9b::/96; the address must lie inside the prefix.",
		Params:  []tools.Param{one("NAT64 IPv6 address."), opt("prefix", "str", "NAT64 prefix with length, default 64:ff9b::/96.", "64:ff9b::/96")},
		Returns: "str", Example: `net.nat64_to_ipv4("64:ff9b::c000:221")`,
		Tags:    []string{"dns64", "64:ff9b", "extract ipv4"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is6() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil || ad.Is4() {
				return nil, fmt.Errorf("%q is not an IPv6 address", a.Str("ip"))
			}
			p, err := nat64Prefix(a.Str("prefix"))
			if err != nil {
				return nil, err
			}
			if !p.Contains(ad) {
				return nil, fmt.Errorf("%s is not inside %s", ad, p)
			}
			b := ad.As16()
			var v [4]byte
			for i, idx := range nat64Idx[p.Bits()] {
				v[i] = b[idx]
			}
			return netip.AddrFrom4(v).String(), nil
		},
	})
	reg(kwMAC, tools.Func{
		Name: "eui64_from_mac", Title: "EUI-64 IPv6 address from MAC", Order: 56, Input: "mac",
		Doc:     "Builds the modified EUI-64 IPv6 address (RFC 4291) for a MAC address: inserts ff:fe in the middle and flips the universal/local bit. prefix is the 64-bit network prefix, default fe80:: (link-local).",
		Params:  []tools.Param{req("mac", "str", "MAC address in any common format."), opt("prefix", "str", "IPv6 /64 prefix, e.g. 2001:db8:1:2::", "fe80::")},
		Returns: "str", Example: `net.eui64_from_mac("00:1a:2b:3c:4d:5e", "2001:db8::")`,
		Tags: []string{"slaac", "interface id", "modified eui-64", "ipv6 from mac"}, Accepts: okMAC,
		Fn: func(a tools.Args) (any, error) {
			mac, err := macBytes(a.Str("mac"))
			if err != nil {
				return nil, err
			}
			ps := strings.TrimSuffix(strings.TrimSpace(a.Str("prefix")), "/64")
			pa, err := netip.ParseAddr(ps)
			if err != nil || pa.Is4() {
				return nil, fmt.Errorf("invalid IPv6 prefix %q", a.Str("prefix"))
			}
			var pre [8]byte
			copy(pre[:], pa.AsSlice()[:8])
			return eui64Addr(pre, mac).String(), nil
		},
	})
	reg(kwMAC, tools.Func{
		Name: "link_local_from_mac", Title: "Link-local IPv6 from MAC", Order: 57, Input: "mac",
		Doc:     "Returns the SLAAC link-local IPv6 address (fe80::/64 plus modified EUI-64) for a MAC address.",
		Params:  []tools.Param{req("mac", "str", "MAC address in any common format.")},
		Returns: "str", Example: `net.link_local_from_mac("00:1a:2b:3c:4d:5e")`,
		Tags: []string{"fe80", "slaac", "ipv6 from mac", "eui-64"}, Accepts: okMAC,
		Fn: func(a tools.Args) (any, error) {
			mac, err := macBytes(a.Str("mac"))
			if err != nil {
				return nil, err
			}
			return eui64Addr([8]byte{0xfe, 0x80}, mac).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ptr_name", Title: "Reverse DNS (PTR) name", Order: 58,
		Doc:     "Returns the reverse-DNS name for an address: 1.2.3.4 gives 4.3.2.1.in-addr.arpa, IPv6 gives the 32-nibble ip6.arpa name. No lookup is performed.",
		Params:  []tools.Param{one("IPv4 or IPv6 address.")},
		Returns: "str", Example: `net.ptr_name("192.0.2.10")`,
		Tags: []string{"reverse dns", "in-addr.arpa", "ip6.arpa", "rdns", "ptr record", "nibble"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			return ptrName(ad), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_compare", Title: "Compare IPs", Order: 59, Input: "a",
		Doc:     "Compares two addresses numerically: -1 if a < b, 0 if equal, 1 if a > b. IPv4 sorts before IPv6.",
		Params:  []tools.Param{req("a", "str", "First address."), req("b", "str", "Second address.")},
		Returns: "int", Example: `net.ip_compare("10.0.0.9", "10.0.0.10")`,
		Tags: []string{"order", "greater", "less", "equal", "sort"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			x, err := parseAddr(a.Str("a"))
			if err != nil {
				return nil, err
			}
			y, err := parseAddr(a.Str("b"))
			if err != nil {
				return nil, err
			}
			return int64(x.Compare(y)), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "sort_ips", Title: "Sort IP addresses", Order: 60,
		Doc:     "Sorts addresses numerically (10.0.0.9 before 10.0.0.10, unlike text sorting), IPv4 first. Entries may also be comma or newline separated. descending reverses the order and unique drops duplicates. Invalid entries are an error.",
		Params:  []tools.Param{req("ips", "list[str]", "Addresses to sort."), opt("descending", "bool", "Sort high to low.", false), opt("unique", "bool", "Remove duplicates.", false)},
		Returns: "list[str]", Example: `net.sort_ips(["10.0.0.10", "10.0.0.9", "::1"])`,
		Tags: []string{"order ips", "natural sort", "dedupe ips", "numeric sort"},
		Fn: func(a tools.Args) (any, error) {
			var addrs []netip.Addr
			for _, s := range splitItems(a.List("ips")) {
				ad, err := parseAddr(s)
				if err != nil {
					return nil, err
				}
				addrs = append(addrs, ad)
			}
			sortAddrs(addrs, a.Bool("descending"))
			out := []string{}
			for i, ad := range addrs {
				if a.Bool("unique") && i > 0 && addrs[i-1] == ad {
					continue
				}
				out = append(out, ad.String())
			}
			return out, nil
		},
	})
	matching := func(ip string, ranges []string) ([]string, error) {
		ad, err := parseAddr(ip)
		if err != nil {
			return nil, err
		}
		n := bigOf(ad)
		out := []string{}
		for _, r := range splitItems(ranges) {
			iv, is4, err := parseRangeItem(r)
			if err != nil {
				return nil, err
			}
			if is4 == ad.Is4() && n.Cmp(iv.lo) >= 0 && n.Cmp(iv.hi) <= 0 {
				out = append(out, r)
			}
		}
		return out, nil
	}
	reg(kwIP, tools.Func{
		Name: "ip_in_ranges", Title: "IP in any range", Order: 61, Input: "ip",
		Doc:     `Reports whether the address lies in at least one of the given CIDRs, single IPs or ranges ("10.0.0.1-10.0.0.50").`,
		Params:  []tools.Param{one("Address to test."), req("ranges", "list[str]", "CIDRs, IPs or a-b ranges.")},
		Returns: "bool", Example: `net.ip_in_ranges("10.1.2.3", ["192.168.0.0/16", "10.0.0.0/8"])`,
		Tags: []string{"allowlist", "acl", "whitelist", "in any subnet", "match"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			m, err := matching(a.Str("ip"), a.List("ranges"))
			return len(m) > 0, err
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_matching_ranges", Title: "Ranges containing IP", Order: 62, Input: "ip",
		Doc:     "Returns the entries (CIDRs, IPs or ranges) from the list that contain the address; empty when none do.",
		Params:  []tools.Param{one("Address to test."), req("ranges", "list[str]", "CIDRs, IPs or a-b ranges.")},
		Returns: "list[str]", Example: `net.ip_matching_ranges("10.1.2.3", ["10.0.0.0/8", "10.1.0.0/16", "172.16.0.0/12"])`,
		Tags: []string{"which subnet", "lookup", "find network", "matching cidr"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) { return matching(a.Str("ip"), a.List("ranges")) },
	})
	reg(kwIP, tools.Func{
		Name: "ip_add", Title: "Add to IP", Order: 63,
		Doc:     "Adds a signed integer to an address (10.0.0.255 + 1 = 10.0.1.0). Fails when the result leaves the address family's range.",
		Params:  []tools.Param{one("Starting address."), req("n", "int", "Offset, may be negative.")},
		Returns: "str", Example: `net.ip_add("10.0.0.255", 1)`,
		Tags: []string{"increment", "offset", "next ip", "previous ip", "plus"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil {
				return nil, err
			}
			r := bigOf(ad)
			r.Add(r, big.NewInt(a.Int("n")))
			out, ok := addrOf(r, ad.Is4())
			if !ok {
				return nil, fmt.Errorf("result is outside the IPv%d address space", map[bool]int{true: 4, false: 6}[ad.Is4()])
			}
			return out.String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ip_diff", Title: "Distance between IPs", Order: 64, Input: "a",
		Doc:     "Returns b minus a as a signed decimal string: the number of addresses from a to b (positive when b is higher). Both must be the same IP version.",
		Params:  []tools.Param{req("a", "str", "Start address."), req("b", "str", "End address.")},
		Returns: "str", Example: `net.ip_diff("10.0.0.1", "10.0.1.1")`,
		Tags: []string{"difference", "distance", "how many addresses", "subtract"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			x, err := parseAddr(a.Str("a"))
			if err != nil {
				return nil, err
			}
			y, err := parseAddr(a.Str("b"))
			if err != nil {
				return nil, err
			}
			if x.Is4() != y.Is4() {
				return nil, fmt.Errorf("addresses must be the same IP version")
			}
			return new(big.Int).Sub(bigOf(y), bigOf(x)).String(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "is_valid_ip", Title: "Is valid IP", Order: 65,
		Doc:     "Reports whether the text is a valid IPv4 or IPv6 address (surrounding whitespace, brackets and leading zeros in IPv4 octets are tolerated).",
		Params:  []tools.Param{one("Text to check.")},
		Returns: "bool", Example: `net.is_valid_ip("256.1.1.1")`,
		Tags: []string{"validate ip", "check ip", "ip address valid"},
		Fn:   func(a tools.Args) (any, error) { return okIP(a.Str("ip")), nil },
	})
	reg(kwIP, tools.Func{
		Name: "is_valid_ipv4", Title: "Is valid IPv4", Order: 66,
		Doc:     "Reports whether the text is a valid dotted-quad IPv4 address.",
		Params:  []tools.Param{one("Text to check.")},
		Returns: "bool", Example: `net.is_valid_ipv4("192.168.0.1")`,
		Tags: []string{"validate ipv4", "dotted quad"},
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			return err == nil && ad.Is4(), nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "is_valid_ipv6", Title: "Is valid IPv6", Order: 67,
		Doc:     "Reports whether the text is a valid IPv6 address (including :: compression and IPv4-mapped forms).",
		Params:  []tools.Param{one("Text to check.")},
		Returns: "bool", Example: `net.is_valid_ipv6("2001:db8::1")`,
		Tags: []string{"validate ipv6"},
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			return err == nil && ad.Is6(), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "is_valid_cidr", Title: "Is valid CIDR", Order: 68,
		Doc:     "Reports whether the text is valid CIDR notation (address/prefix, IPv4 or IPv6). Host bits may be set. Unlike cidr_info it does not accept netmask spellings or a bare address.",
		Params:  []tools.Param{pCIDR("Text to check.")},
		Returns: "bool", Example: `net.is_valid_cidr("10.0.0.0/33")`,
		Tags: []string{"validate cidr", "check subnet", "prefix valid"},
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("cidr"))
			if !strings.Contains(s, "/") {
				return false, nil
			}
			_, err := netip.ParsePrefix(s)
			return err == nil, nil
		},
	})
	reg(kwIP, tools.Func{
		Name: "ipv4_octets", Title: "IPv4 octets", Order: 69,
		Doc:     "Splits an IPv4 address into its four octets as decimal strings.",
		Params:  []tools.Param{one("IPv4 address.")},
		Returns: "list[str]", Example: `net.ipv4_octets("192.168.1.1")`,
		Tags:    []string{"split ip", "octets", "bytes", "parts"},
		Accepts: func(s string) bool { a, err := parseAddr(s); return err == nil && a.Is4() },
		Fn: func(a tools.Args) (any, error) {
			ad, err := parseAddr(a.Str("ip"))
			if err != nil || !ad.Is4() {
				return nil, fmt.Errorf("%q is not an IPv4 address", a.Str("ip"))
			}
			b := ad.As4()
			return []string{strconv.Itoa(int(b[0])), strconv.Itoa(int(b[1])), strconv.Itoa(int(b[2])), strconv.Itoa(int(b[3]))}, nil
		},
	})
}

func isPlainSmallInt(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 5 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func cleanBits(s string) string {
	return strings.NewReplacer(".", "", ":", "", " ", "", "_", "", "-", "", "0b", "").Replace(strings.TrimSpace(s))
}

func onlyBits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '0' && r != '1' {
			return false
		}
	}
	return true
}

func hexToAddr(s string) (netip.Addr, error) {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(strings.TrimPrefix(t, "0x"), "0X")
	t = strings.NewReplacer(".", "", ":", "", " ", "", "_", "", "-", "").Replace(t)
	if len(t) != 8 && len(t) != 32 {
		return netip.Addr{}, fmt.Errorf("need 8 or 32 hex digits, got %d", len(t))
	}
	n, ok := new(big.Int).SetString(t, 16)
	if !ok {
		return netip.Addr{}, fmt.Errorf("invalid hex %q", s)
	}
	a, _ := addrOf(n, len(t) == 8)
	return a, nil
}

func okMAC(s string) bool {
	_, err := macBytes(s)
	return err == nil
}
