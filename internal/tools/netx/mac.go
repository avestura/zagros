package netx

import (
	"fmt"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func macFormat(m [6]byte, format string, upper bool) (string, error) {
	var s string
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "colon", "lower":
		s = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
	case "upper":
		upper = true
		s = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
	case "dash", "hyphen":
		s = fmt.Sprintf("%02x-%02x-%02x-%02x-%02x-%02x", m[0], m[1], m[2], m[3], m[4], m[5])
	case "dot", "cisco":
		s = fmt.Sprintf("%02x%02x.%02x%02x.%02x%02x", m[0], m[1], m[2], m[3], m[4], m[5])
	case "plain", "none", "bare":
		s = fmt.Sprintf("%02x%02x%02x%02x%02x%02x", m[0], m[1], m[2], m[3], m[4], m[5])
	default:
		return "", fmt.Errorf("unknown format %q (use colon, dash, dot, plain, upper, lower)", format)
	}
	if upper {
		s = strings.ToUpper(s)
	}
	return s, nil
}

func macInfo(s string) (map[string]string, error) {
	m, err := macBytes(s)
	if err != nil {
		return nil, err
	}
	colon, _ := macFormat(m, "colon", false)
	oui := fmt.Sprintf("%02X:%02X:%02X", m[0], m[1], m[2])
	info := map[string]string{
		"mac":            colon,
		"upper":          strings.ToUpper(colon),
		"dash":           strings.ToUpper(strings.ReplaceAll(colon, ":", "-")),
		"cisco":          fmt.Sprintf("%02x%02x.%02x%02x.%02x%02x", m[0], m[1], m[2], m[3], m[4], m[5]),
		"oui":            oui,
		"nic":            fmt.Sprintf("%02X:%02X:%02X", m[3], m[4], m[5]),
		"eui64":          fmt.Sprintf("%02x:%02x:%02x:ff:fe:%02x:%02x:%02x", m[0]^2, m[1], m[2], m[3], m[4], m[5]),
		"link_local":     eui64Addr([8]byte{0xfe, 0x80}, m).String(),
		"cast":           "unicast",
		"administration": "universally administered (UAA)",
	}
	if m[0]&1 == 1 {
		info["cast"] = "multicast"
	}
	if m[0]&2 == 2 {
		info["administration"] = "locally administered (LAA)"
	}
	if m == [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} {
		info["cast"] = "broadcast"
	}
	if v, ok := vendorOf(m); ok {
		info["vendor"] = v
	}
	return info, nil
}

func vendorOf(m [6]byte) (string, bool) {
	key := fmt.Sprintf("%02X%02X%02X", m[0], m[1], m[2])
	if v, ok := ouiTable[key]; ok {
		return v, true
	}
	switch {
	case m == [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}:
		return "Broadcast", true
	case m[0] == 0x01 && m[1] == 0x00 && m[2] == 0x5e && m[3]&0x80 == 0:
		return "IPv4 multicast (RFC 1112)", true
	case m[0] == 0x33 && m[1] == 0x33:
		return "IPv6 multicast (RFC 2464)", true
	case m[0] == 0x01 && m[1] == 0x80 && m[2] == 0xc2:
		return "IEEE 802.1 bridge / link-layer protocols (STP, LLDP, LACP)", true
	case m[0] == 0x01 && m[1] == 0x00 && m[2] == 0x0c:
		return "Cisco protocols (CDP, VTP)", true
	case m[0] == 0x02 && m[1] == 0x42:
		return "Docker container (locally administered)", true
	}
	return "", false
}

func init() {
	pMac := req("mac", "str", "MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff).")
	reg(kwMAC, tools.Func{
		Name: "mac_normalize", Title: "Normalize MAC address", Order: 8,
		Doc:     "Rewrites a MAC address in a chosen format: colon (aa:bb:cc:dd:ee:ff, default), dash, dot (Cisco aabb.ccdd.eeff), plain (no separators), upper (colon, upper-case) or lower. upper=true upper-cases any format. Accepts any separators.",
		Params:  []tools.Param{pMac, opt("format", "str", "colon (default), dash, dot, plain, upper or lower.", "colon"), opt("upper", "bool", "Upper-case the result.", false)},
		Returns: "str", Example: `net.mac_normalize("00-1A-2B-3C-4D-5E", "dot")`,
		Tags:    []string{"mac format", "convert mac", "cisco format", "hyphen", "colon", "hardware address", "ethernet address"},
		Accepts: okMAC,
		Fn: func(a tools.Args) (any, error) {
			m, err := macBytes(a.Str("mac"))
			if err != nil {
				return nil, err
			}
			return macFormat(m, a.Str("format"), a.Bool("upper"))
		},
	})
	reg(kwMAC, tools.Func{
		Name: "mac_is_valid", Title: "Is valid MAC", Order: 70,
		Doc:     "Reports whether the text is a valid 48-bit MAC address: exactly 12 hex digits with optional : - . or space separators.",
		Params:  []tools.Param{pMac},
		Returns: "bool", Example: `net.mac_is_valid("00:1A:2B:3C:4D")`,
		Tags: []string{"validate mac", "check mac"},
		Fn:   func(a tools.Args) (any, error) { return okMAC(a.Str("mac")), nil },
	})
	reg(kwMAC, tools.Func{
		Name: "mac_info", Title: "MAC address info", Order: 71,
		Doc:     "Describes a MAC address: unicast/multicast/broadcast, universally or locally administered bit, OUI and NIC parts, the EUI-64 form, the SLAAC link-local IPv6 address and, for well-known prefixes, the vendor (the OUI table is partial).",
		Params:  []tools.Param{pMac},
		Returns: "dict[str, str]", Example: `net.mac_info("00:50:56:aa:bb:cc")`,
		Tags:    []string{"oui", "vendor", "multicast", "local bit", "eui-64", "unicast", "mac details"},
		Accepts: okMAC,
		Fn:      func(a tools.Args) (any, error) { return macInfo(a.Str("mac")) },
	})
	reg(kwMAC, tools.Func{
		Name: "oui_vendor", Title: "MAC vendor (OUI lookup)", Order: 72,
		Doc:     "Looks up the manufacturer of a MAC address (or an OUI like 00:50:56) in a small embedded table of about 200 well-known prefixes. The table is intentionally partial: an empty string means unknown, not unassigned. Also recognizes broadcast, multicast and Docker addresses.",
		Params:  []tools.Param{req("mac", "str", "MAC address or its first three bytes.")},
		Returns: "str", Example: `net.oui_vendor("b8:27:eb:12:34:56")`,
		Tags:    []string{"manufacturer", "vendor", "oui", "who makes", "mac lookup", "ieee"},
		Accepts: func(s string) bool { _, ok := ouiOf(s); return ok },
		Fn: func(a tools.Args) (any, error) {
			m, ok := ouiOf(a.Str("mac"))
			if !ok {
				return nil, fmt.Errorf("invalid MAC address or OUI %q", a.Str("mac"))
			}
			v, _ := vendorOf(m)
			return v, nil
		},
	})
}

// ouiOf accepts a full MAC or just the first three bytes.
func ouiOf(s string) ([6]byte, bool) {
	if m, err := macBytes(s); err == nil {
		return m, true
	}
	var h strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
			h.WriteRune(r)
		case r == ':' || r == '-' || r == '.' || r == ' ':
		default:
			return [6]byte{}, false
		}
	}
	if h.Len() != 6 {
		return [6]byte{}, false
	}
	m, err := macBytes(h.String() + "000000")
	return m, err == nil
}
