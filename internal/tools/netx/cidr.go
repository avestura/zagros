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

func accNet(s string) bool { return okNet(s) }

func splitItems(in []string) []string {
	var out []string
	for _, s := range in {
		for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '\n' || r == ';' }) {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func maskString(bits int, is4 bool) string {
	if is4 {
		var m uint32
		if bits > 0 {
			m = ^uint32(0) << uint(32-bits)
		}
		return netip.AddrFrom4([4]byte{byte(m >> 24), byte(m >> 16), byte(m >> 8), byte(m)}).String()
	}
	n := new(big.Int).Sub(pow2(128), pow2(128-bits))
	a, _ := addrOf(n, false)
	return a.String()
}

func wildcardString(bits int, is4 bool) string {
	if is4 {
		var m uint32
		if bits > 0 {
			m = ^uint32(0) << uint(32-bits)
		}
		m = ^m
		return netip.AddrFrom4([4]byte{byte(m >> 24), byte(m >> 16), byte(m >> 8), byte(m)}).String()
	}
	a, _ := addrOf(new(big.Int).Sub(pow2(128-bits), big.NewInt(1)), false)
	return a.String()
}

func binaryMask(bits int, is4 bool) string {
	if is4 {
		s := strings.Repeat("1", bits) + strings.Repeat("0", 32-bits)
		return s[0:8] + "." + s[8:16] + "." + s[16:24] + "." + s[24:32]
	}
	s := strings.Repeat("1", bits) + strings.Repeat("0", 128-bits)
	parts := make([]string, 8)
	for i := range parts {
		parts[i] = s[i*16 : i*16+16]
	}
	return strings.Join(parts, ":")
}

func hexMask(bits int, is4 bool) string {
	if is4 {
		var m uint32
		if bits > 0 {
			m = ^uint32(0) << uint(32-bits)
		}
		return fmt.Sprintf("0x%08x", m)
	}
	n := new(big.Int).Sub(pow2(128), pow2(128-bits))
	return fmt.Sprintf("0x%032x", n)
}

func ipClass(a netip.Addr) string {
	if !a.Is4() {
		return ""
	}
	switch o := a.As4()[0]; {
	case o < 128:
		return "A"
	case o < 192:
		return "B"
	case o < 224:
		return "C"
	case o < 240:
		return "D"
	default:
		return "E"
	}
}

func usable(p netip.Prefix) *big.Int {
	t := total(p)
	if p.Addr().Is4() {
		switch {
		case p.Bits() == 32:
			return big.NewInt(1)
		case p.Bits() == 31:
			return big.NewInt(2)
		}
		return t.Sub(t, big.NewInt(2))
	}
	return t
}

// hostRange returns the first and last usable host of p.
func hostRange(p netip.Prefix) (netip.Addr, netip.Addr) {
	f, l := firstLast(p)
	if p.Addr().Is4() && p.Bits() <= 30 {
		f = f.Next()
		l = l.Prev()
	}
	return f, l
}

func ptrName(a netip.Addr) string {
	if a.Is4() {
		b := a.As4()
		return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa", b[3], b[2], b[1], b[0])
	}
	b := a.As16()
	var sb strings.Builder
	for i := 15; i >= 0; i-- {
		fmt.Fprintf(&sb, "%x.%x.", b[i]&0xf, b[i]>>4)
	}
	return sb.String() + "ip6.arpa"
}

func reverseZone(p netip.Prefix) string {
	a := p.Masked().Addr()
	if a.Is4() {
		b := a.As4()
		n := p.Bits() / 8
		var labels []string
		for i := n - 1; i >= 0; i-- {
			labels = append(labels, strconv.Itoa(int(b[i])))
		}
		zone := strings.Join(labels, ".")
		if n == 0 {
			zone = "in-addr.arpa"
		} else {
			zone += ".in-addr.arpa"
		}
		if p.Bits()%8 != 0 && p.Bits() > 24 {
			return fmt.Sprintf("%d/%d.%s", b[3], p.Bits(), zone) // RFC 2317 classless delegation
		}
		return zone
	}
	full := ptrName(a)
	nib := p.Bits() / 4
	labels := strings.Split(full, ".")
	rest := labels[32-nib:]
	if nib == 0 {
		return "ip6.arpa"
	}
	return strings.Join(rest, ".")
}

func cidrInfo(s string) (map[string]string, error) {
	addr, p, err := parseNet(s)
	if err != nil {
		return nil, err
	}
	is4 := addr.Is4()
	first, last := firstLast(p)
	hf, hl := hostRange(p)
	m := map[string]string{
		"address":         addr.String(),
		"cidr":            p.String(),
		"network":         first.String(),
		"prefix":          strconv.Itoa(p.Bits()),
		"version":         map[bool]string{true: "4", false: "6"}[is4],
		"netmask":         maskString(p.Bits(), is4),
		"wildcard":        wildcardString(p.Bits(), is4),
		"first_host":      hf.String(),
		"last_host":       hl.String(),
		"last_address":    last.String(),
		"usable_hosts":    usable(p).String(),
		"total_addresses": total(p).String(),
		"type":            classifyAddr(first),
		"binary_netmask":  binaryMask(p.Bits(), is4),
		"hex_netmask":     hexMask(p.Bits(), is4),
		"reverse_zone":    reverseZone(p),
	}
	if is4 {
		m["broadcast"] = last.String()
		m["ip_class"] = ipClass(addr)
		if p.Bits() == 31 {
			m["note"] = "point-to-point link (RFC 3021): both addresses are usable"
		}
	}
	return m, nil
}

func containsStr(cidr, other string) (bool, error) {
	p, err := mustNet(cidr)
	if err != nil {
		return false, err
	}
	if strings.ContainsAny(other, "/ ") {
		_, q, err := parseNet(other)
		if err != nil {
			return false, err
		}
		return p.Addr().Is4() == q.Addr().Is4() && p.Bits() <= q.Bits() && p.Contains(q.Addr()), nil
	}
	a, err := parseAddr(other)
	if err != nil {
		return false, err
	}
	return p.Contains(a), nil
}

func subnets(p netip.Prefix, newBits int, limit int) ([]netip.Prefix, *big.Int, error) {
	mb := maxBits(p.Addr())
	if newBits < p.Bits() || newBits > mb {
		return nil, nil, fmt.Errorf("new prefix must be between /%d and /%d", p.Bits(), mb)
	}
	count := pow2(newBits - p.Bits())
	first, _ := firstLast(p)
	cur := bigOf(first)
	step := pow2(mb - newBits)
	var out []netip.Prefix
	for i := 0; i < limit && big.NewInt(int64(i)).Cmp(count) < 0; i++ {
		a, _ := addrOf(cur, first.Is4())
		q, _ := a.Prefix(newBits)
		out = append(out, q)
		cur.Add(cur, step)
	}
	return out, count, nil
}

func capLimit(n int64, def int) int {
	if n <= 0 {
		return def
	}
	if n > maxEnum {
		return maxEnum
	}
	return int(n)
}

func subtract(p netip.Prefix, ex netip.Prefix) []netip.Prefix {
	f, l := firstLast(p)
	ef, el := firstLast(ex)
	if f.Is4() != ef.Is4() {
		return []netip.Prefix{p}
	}
	lo, hi := bigOf(f), bigOf(l)
	elo, ehi := bigOf(ef), bigOf(el)
	if ehi.Cmp(lo) < 0 || elo.Cmp(hi) > 0 {
		return []netip.Prefix{p.Masked()}
	}
	var out []netip.Prefix
	if elo.Cmp(lo) > 0 {
		out = append(out, rangeToPrefixes(lo, new(big.Int).Sub(elo, big.NewInt(1)), f.Is4())...)
	}
	if ehi.Cmp(hi) < 0 {
		out = append(out, rangeToPrefixes(new(big.Int).Add(ehi, big.NewInt(1)), hi, f.Is4())...)
	}
	return out
}

func log2ceil(n int64) int {
	b := 0
	for (int64(1) << uint(b)) < n {
		b++
	}
	return b
}

type vlsmReq struct {
	name  string
	hosts int64
}

func parseDecimalInt(s string) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid number %q", s)
	}
	return n, nil
}

func vlsm(base string, hosts []string) ([]string, error) {
	_, p, err := parseNet(base)
	if err != nil {
		return nil, err
	}
	if !p.Addr().Is4() {
		return nil, fmt.Errorf("vlsm supports IPv4 base networks only")
	}
	hosts = splitItems(hosts)
	if len(hosts) == 0 {
		return nil, fmt.Errorf("give at least one host count, e.g. [\"LAN-A:100\", \"50\"]")
	}
	var reqs []vlsmReq
	for i, h := range hosts {
		name := fmt.Sprintf("net%d", i+1)
		num := h
		if j := strings.LastIndexAny(h, ":="); j >= 0 {
			name, num = strings.TrimSpace(h[:j]), h[j+1:]
		}
		n, err := parseDecimalInt(num)
		if err != nil {
			return nil, err
		}
		reqs = append(reqs, vlsmReq{name, n})
	}
	sort.SliceStable(reqs, func(i, j int) bool { return reqs[i].hosts > reqs[j].hosts })
	first, last := firstLast(p)
	cur := bigOf(first)
	end := bigOf(last)
	out := []string{"name | hosts needed | block | network | netmask | usable range | broadcast"}
	for _, r := range reqs {
		hb := log2ceil(r.hosts + 2)
		if hb < 2 {
			hb = 2
		}
		if hb > 32 {
			return nil, fmt.Errorf("%s: too many hosts", r.name)
		}
		size := pow2(hb)
		// align cur up to a multiple of size
		rem := new(big.Int).Mod(cur, size)
		if rem.Sign() != 0 {
			cur.Add(cur, new(big.Int).Sub(size, rem))
		}
		blockEnd := new(big.Int).Add(cur, size)
		blockEnd.Sub(blockEnd, big.NewInt(1))
		if blockEnd.Cmp(end) > 0 {
			return nil, fmt.Errorf("%s needs a /%d (%d addresses) but %s has no room left", r.name, 32-hb, size.Int64(), p)
		}
		a, _ := addrOf(cur, true)
		q, _ := a.Prefix(32 - hb)
		f, l := hostRange(q)
		bc, _ := addrOf(blockEnd, true)
		out = append(out, fmt.Sprintf("%s | %d | /%d | %s | %s | %s - %s | %s", r.name, r.hosts, 32-hb, q, maskString(32-hb, true), f, l, bc))
		cur.Add(cur, size)
	}
	return out, nil
}

func init() {
	reg(kwCIDR, tools.Func{
		Name: "cidr_info", Title: "CIDR / subnet info", Order: 1,
		Doc: "Full subnet calculator for IPv4 and IPv6: network, broadcast (IPv4), netmask, wildcard, first/last host, usable hosts, total addresses (string, can be huge for IPv6), class, type (private, public, loopback, CGNAT...), binary and hex netmask and the reverse DNS zone. " +
			"Accepts " + cidrForms + " /31 and /32 follow RFC 3021.",
		Params:  []tools.Param{pCIDR("CIDR or address with mask: " + cidrForms)},
		Returns: "dict[str, str]", Example: `net.cidr_info("192.168.1.130/26")`,
		Tags:    []string{"subnet calculator", "netmask", "wildcard", "broadcast", "network address", "hosts", "ipcalc", "cidr notation", "prefix", "usable hosts"},
		Accepts: accNet,
		Fn:      func(a tools.Args) (any, error) { return cidrInfo(a.Str("cidr")) },
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_contains", Title: "CIDR contains IP", Order: 20,
		Doc:     "Reports whether an IP address (or a whole CIDR block) lies inside the CIDR. Mixing IPv4 and IPv6 gives false.",
		Params:  []tools.Param{pCIDR("The containing network."), req("ip", "str", "IP address or CIDR to test.")},
		Returns: "bool", Example: `net.cidr_contains("10.0.0.0/8", "10.20.30.40")`,
		Tags: []string{"in subnet", "belongs", "member", "within", "is ip in network"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) { return containsStr(a.Str("cidr"), a.Str("ip")) },
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_overlap", Title: "CIDRs overlap", Order: 21,
		Doc:     "Reports whether two CIDR blocks share any address (one contains the other, or they are equal).",
		Params:  []tools.Param{req("a", "str", "First CIDR."), req("b", "str", "Second CIDR.")},
		Returns: "bool", Example: `net.cidr_overlap("10.0.0.0/16", "10.0.5.0/24")`, Input: "a",
		Tags: []string{"intersect", "conflict", "collision", "overlapping subnets"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("a"))
			if err != nil {
				return nil, err
			}
			q, err := mustNet(a.Str("b"))
			if err != nil {
				return nil, err
			}
			return p.Addr().Is4() == q.Addr().Is4() && p.Overlaps(q), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_subnets", Title: "Divide into subnets", Order: 22,
		Doc:     "Splits a CIDR into all subnets of the longer prefix new_prefix (for example /16 into /24s). At most limit subnets are returned (default 256, max 65536); a larger split is truncated.",
		Params:  []tools.Param{pCIDR("Network to divide."), req("new_prefix", "int", "Target prefix length, longer than the current one."), opt("limit", "int", "Maximum subnets to list (default 256, max 65536).", int64(256))},
		Returns: "list[str]", Example: `net.cidr_subnets("192.168.0.0/24", 26)`,
		Tags: []string{"split", "divide", "subnetting", "carve", "child networks"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			out, _, err := subnets(p, int(a.Int("new_prefix")), capLimit(a.Int("limit"), 256))
			return prefixStrings(out), err
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_split", Title: "Split CIDR into equal parts", Order: 23,
		Doc:     "Splits a CIDR into parts equal subnets. parts must be a power of two (2, 4, 8...) and at most 65536.",
		Params:  []tools.Param{pCIDR("Network to split."), req("parts", "int", "Number of equal parts (power of two).")},
		Returns: "list[str]", Example: `net.cidr_split("10.0.0.0/24", 4)`,
		Tags: []string{"equal subnets", "divide", "halve", "partition"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			n := a.Int("parts")
			if n < 1 || n > maxEnum || n&(n-1) != 0 {
				return nil, fmt.Errorf("parts must be a power of two between 1 and %d", maxEnum)
			}
			out, _, err := subnets(p, p.Bits()+log2ceil(n), maxEnum)
			return prefixStrings(out), err
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_supernet", Title: "Supernet", Order: 24,
		Doc:     "Returns the larger network with the shorter prefix new_prefix that contains the given CIDR (for example 192.168.5.0/24 with /16 gives 192.168.0.0/16).",
		Params:  []tools.Param{pCIDR("Network to widen."), req("new_prefix", "int", "Shorter target prefix length.")},
		Returns: "str", Example: `net.cidr_supernet("192.168.5.0/24", 16)`,
		Tags: []string{"parent network", "aggregate", "widen", "summarize one"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			addr, p, err := parseNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			n := int(a.Int("new_prefix"))
			if n < 0 || n > p.Bits() {
				return nil, fmt.Errorf("new prefix must be between /0 and /%d", p.Bits())
			}
			q, err := addr.Prefix(n)
			return q.String(), err
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_summarize", Title: "Summarize / merge CIDRs", Order: 25,
		Doc:     "Merges a list of CIDRs, IP addresses and ranges (\"10.0.0.1-10.0.0.9\") into the smallest set of CIDR blocks covering exactly the same addresses (route aggregation). Entries may also be comma or newline separated. IPv4 results come first, then IPv6.",
		Params:  []tools.Param{req("items", "list[str]", "CIDRs, IPs or ranges to merge.")},
		Returns: "list[str]", Example: `net.cidr_summarize(["10.0.0.0/25", "10.0.0.128/25", "10.0.1.0/24"])`,
		Tags: []string{"aggregate", "merge", "supernet", "route summarization", "collapse", "minimal"},
		Fn: func(a tools.Args) (any, error) {
			var v4, v6 []ival
			for _, it := range splitItems(a.List("items")) {
				iv, is4, err := parseRangeItem(it)
				if err != nil {
					return nil, err
				}
				if is4 {
					v4 = append(v4, iv)
				} else {
					v6 = append(v6, iv)
				}
			}
			out := []string{}
			for _, g := range []struct {
				l   []ival
				is4 bool
			}{{v4, true}, {v6, false}} {
				for _, iv := range mergeIvals(g.l) {
					out = append(out, prefixStrings(rangeToPrefixes(iv.lo, iv.hi, g.is4))...)
				}
			}
			if len(out) > maxItems {
				out = out[:maxItems]
			}
			return out, nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_exclude", Title: "Exclude subnet from CIDR", Order: 26,
		Doc:     "Removes the exclude block from the CIDR and returns the remaining space as the minimal list of CIDRs. When they do not overlap the original network is returned; when exclude covers everything the list is empty.",
		Params:  []tools.Param{pCIDR("Network to cut from."), req("exclude", "str", "Block to remove.")},
		Returns: "list[str]", Example: `net.cidr_exclude("10.0.0.0/24", "10.0.0.64/26")`,
		Tags: []string{"subtract", "remove", "carve out", "difference", "minus"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			ex, err := mustNet(a.Str("exclude"))
			if err != nil {
				return nil, err
			}
			return prefixStrings(subtract(p, ex)), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "range_to_cidrs", Title: "IP range to CIDRs", Order: 27, Input: "start",
		Doc:     "Converts an inclusive IP range into the minimal list of CIDR blocks covering it exactly. start and end must be the same IP version; swapped bounds are corrected.",
		Params:  []tools.Param{req("start", "str", "First IP of the range."), req("end", "str", "Last IP of the range.")},
		Returns: "list[str]", Example: `net.range_to_cidrs("10.0.0.5", "10.0.0.20")`,
		Tags: []string{"range", "ip range", "from to", "cidr list", "cover"}, Accepts: okIP,
		Fn: func(a tools.Args) (any, error) {
			s, err := parseAddr(a.Str("start"))
			if err != nil {
				return nil, err
			}
			e, err := parseAddr(a.Str("end"))
			if err != nil {
				return nil, err
			}
			if s.Is4() != e.Is4() {
				return nil, fmt.Errorf("start and end must be the same IP version")
			}
			lo, hi := bigOf(s), bigOf(e)
			if lo.Cmp(hi) > 0 {
				lo, hi = hi, lo
			}
			return prefixStrings(rangeToPrefixes(lo, hi, s.Is4())), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_to_range", Title: "CIDR to IP range", Order: 28,
		Doc:     `Returns the first and last address of the CIDR as "first - last" (network and broadcast included).`,
		Params:  []tools.Param{pCIDR("CIDR block.")},
		Returns: "str", Example: `net.cidr_to_range("192.168.1.0/24")`,
		Tags: []string{"range", "first last", "from to", "start end"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			f, l := firstLast(p)
			return f.String() + " - " + l.String(), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_hosts", Title: "List usable hosts", Order: 29,
		Doc:     "Lists usable host addresses of the CIDR (network and broadcast excluded for IPv4 up to /30). At most limit addresses are returned (default 256, max 65536).",
		Params:  []tools.Param{pCIDR("CIDR block."), opt("limit", "int", "Maximum addresses to list (default 256, max 65536).", int64(256))},
		Returns: "list[str]", Example: `net.cidr_hosts("192.168.1.0/29")`,
		Tags: []string{"addresses", "enumerate", "expand", "all ips", "list ips"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			f, l := hostRange(p)
			lim := capLimit(a.Int("limit"), 256)
			out := []string{}
			for cur := f; len(out) < lim; cur = cur.Next() {
				out = append(out, cur.String())
				if cur == l || !cur.Next().IsValid() {
					break
				}
			}
			return out, nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_next", Title: "Next adjacent CIDR", Order: 30,
		Doc:     "Returns the next network of the same size directly after the given CIDR (10.0.0.0/24 gives 10.0.1.0/24). Fails at the end of the address space.",
		Params:  []tools.Param{pCIDR("CIDR block.")},
		Returns: "str", Example: `net.cidr_next("10.0.0.0/24")`,
		Tags: []string{"following", "adjacent", "next subnet", "successor"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			f, _ := firstLast(p)
			n := bigOf(f)
			n.Add(n, total(p))
			na, ok := addrOf(n, f.Is4())
			if !ok {
				return nil, fmt.Errorf("no network after %s", p)
			}
			q, _ := na.Prefix(p.Bits())
			return q.String(), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "cidr_count", Title: "Count addresses", Order: 31,
		Doc:     "Returns the number of addresses in the CIDR as a decimal string (exact, even for huge IPv6 blocks). With usable true it returns the usable host count instead (IPv4 subtracts network and broadcast up to /30).",
		Params:  []tools.Param{pCIDR("CIDR block."), opt("usable", "bool", "Count usable hosts instead of all addresses.", false)},
		Returns: "str", Example: `net.cidr_count("10.0.0.0/22")`,
		Tags: []string{"size", "how many", "number of ips", "hosts count"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			if a.Bool("usable") {
				return usable(p).String(), nil
			}
			return total(p).String(), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "mask_to_prefix", Title: "Netmask to prefix length", Order: 32, Input: "mask",
		Doc:     "Converts a dotted IPv4 netmask (255.255.255.0) or wildcard mask (0.0.0.255) to its prefix length (24). Fails for non-contiguous masks.",
		Params:  []tools.Param{req("mask", "str", "Dotted netmask or wildcard mask.")},
		Returns: "int", Example: `net.mask_to_prefix("255.255.240.0")`,
		Tags:    []string{"netmask to cidr", "slash", "prefix length", "mask bits", "wildcard_to_prefix"},
		Accepts: func(s string) bool { a, err := parseAddr(s); _, ok := maskBits(a); return err == nil && ok },
		Fn: func(a tools.Args) (any, error) {
			m, err := parseAddr(a.Str("mask"))
			if err != nil {
				return nil, err
			}
			n, ok := maskBits(m)
			if !ok {
				return nil, fmt.Errorf("%q is not a valid contiguous netmask", a.Str("mask"))
			}
			return int64(n), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "wildcard_to_prefix", Title: "Wildcard mask to prefix length", Order: 33, Input: "wildcard",
		Doc:     "Converts a wildcard (inverse) mask such as 0.0.0.255 to its prefix length (24), as used in Cisco ACLs and OSPF network statements. Fails for non-contiguous masks.",
		Params:  []tools.Param{req("wildcard", "str", "Wildcard mask, e.g. 0.0.15.255.")},
		Returns: "int", Example: `net.wildcard_to_prefix("0.0.15.255")`,
		Tags:    []string{"inverse mask", "acl", "cisco", "ospf"},
		Accepts: func(s string) bool { return okIP(s) && strings.HasPrefix(strings.TrimSpace(s), "0.") },
		Fn: func(a tools.Args) (any, error) {
			m, err := parseAddr(a.Str("wildcard"))
			if err != nil || !m.Is4() {
				return nil, fmt.Errorf("invalid wildcard mask %q", a.Str("wildcard"))
			}
			b := m.As4()
			x := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
			n, ok := contiguous(^x)
			if !ok {
				return nil, fmt.Errorf("%q is not a contiguous wildcard mask", a.Str("wildcard"))
			}
			return int64(n), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "prefix_to_mask", Title: "Prefix length to netmask", Order: 34, Input: "prefix",
		Doc:     "Converts a prefix length (24 or /24, given as a string) to a netmask. IPv4 by default (255.255.255.0); version 6 gives the IPv6 mask. With wildcard true it returns the inverse wildcard mask instead.",
		Params:  []tools.Param{req("prefix", "str", `Prefix length such as "24" or "/24".`), opt("version", "int", "4 (default) or 6.", int64(4)), opt("wildcard", "bool", "Return the wildcard mask instead.", false)},
		Returns: "str", Example: `net.prefix_to_mask("20")`,
		Tags: []string{"cidr to netmask", "slash to mask", "subnet mask", "wildcard"},
		Accepts: func(s string) bool {
			n, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(s), "/"))
			return err == nil && n >= 0 && n <= 128
		},
		Fn: func(a tools.Args) (any, error) {
			n, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(a.Str("prefix")), "/"))
			if err != nil {
				return nil, fmt.Errorf("invalid prefix length %q", a.Str("prefix"))
			}
			is4 := a.Int("version") != 6
			if n < 0 || (is4 && n > 32) || n > 128 {
				return nil, fmt.Errorf("prefix length %d out of range", n)
			}
			if a.Bool("wildcard") {
				return wildcardString(n, is4), nil
			}
			return maskString(n, is4), nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "subnet_for_hosts", Title: "Subnet size for host count", Order: 35, Input: "hosts",
		Doc:     "Finds the smallest IPv4 subnet that offers at least the needed number of usable hosts (network and broadcast excluded) and returns prefix, netmask and capacity. With base set it also gives the first such subnet inside the base network and how many fit.",
		Params:  []tools.Param{req("hosts", "str", `Needed usable hosts, as a string such as "50".`), opt("base", "str", "Optional base network to allocate from, e.g. 10.0.0.0/16.", "")},
		Returns: "dict[str, str]", Example: `net.subnet_for_hosts("50")`,
		Tags:    []string{"how big", "vlsm", "host count", "size subnet", "smallest prefix"},
		Accepts: func(s string) bool { n, err := parseDecimalInt(s); return err == nil && n < 1<<31 },
		Fn: func(a tools.Args) (any, error) {
			n, err := parseDecimalInt(a.Str("hosts"))
			if err != nil || n > (1<<32)-2 {
				return nil, fmt.Errorf("invalid host count %q", a.Str("hosts"))
			}
			hb := log2ceil(n + 2)
			if hb < 2 {
				hb = 2
			}
			bits := 32 - hb
			p := netip.PrefixFrom(netip.AddrFrom4([4]byte{}), bits)
			m := map[string]string{
				"hosts_needed": strconv.FormatInt(n, 10), "prefix": strconv.Itoa(bits), "netmask": maskString(bits, true),
				"wildcard": wildcardString(bits, true), "usable_hosts": usable(p).String(), "total_addresses": total(p).String(),
			}
			if b := strings.TrimSpace(a.Str("base")); b != "" {
				bp, err := mustNet(b)
				if err != nil {
					return nil, err
				}
				if !bp.Addr().Is4() || bits < bp.Bits() {
					return nil, fmt.Errorf("a /%d does not fit inside %s", bits, bp)
				}
				subs, count, _ := subnets(bp, bits, 1)
				m["first_subnet"] = subs[0].String()
				m["subnets_available"] = count.String()
			}
			return m, nil
		},
	})
	reg(kwCIDR, tools.Func{
		Name: "vlsm", Title: "VLSM allocation", Order: 36, Input: "base",
		Doc:     "Variable-length subnet masking: allocates subnets from an IPv4 base network for a list of host requirements, largest first, each aligned on its block size. Each requirement is a number or \"name:number\". Returns a table (header row first) with block, network, mask, usable range and broadcast; fails when the base is too small.",
		Params:  []tools.Param{req("base", "str", "IPv4 network to allocate from, e.g. 192.168.0.0/24."), req("hosts", "list[str]", `Host requirements such as ["LAN-A:100", "LAN-B:50", "20"].`)},
		Returns: "list[str]", Example: `net.vlsm("192.168.0.0/24", ["LAN-A:100", "LAN-B:50", "WAN:2"])`,
		Tags: []string{"variable length subnet", "allocation", "plan", "address plan", "subnetting exercise"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) { return vlsm(a.Str("base"), a.List("hosts")) },
	})
	reg(kwCIDR, tools.Func{
		Name: "subnet_table", Title: "Subnet quick reference", Order: 37,
		Doc:     "Quick-reference table for dividing a CIDR: for each longer prefix up to 8 bits deeper (or the maximum), lists the netmask, number of subnets, addresses per subnet and usable hosts per subnet.",
		Params:  []tools.Param{pCIDR("Network to tabulate, e.g. 10.0.0.0/16.")},
		Returns: "list[str]", Example: `net.subnet_table("192.168.0.0/24")`,
		Tags: []string{"cheat sheet", "reference", "subnet chart", "cidr chart", "table"}, Accepts: accNet,
		Fn: func(a tools.Args) (any, error) {
			p, err := mustNet(a.Str("cidr"))
			if err != nil {
				return nil, err
			}
			is4 := p.Addr().Is4()
			mb := maxBits(p.Addr())
			out := []string{"prefix | netmask | subnets | addresses each | usable each"}
			for b := p.Bits(); b <= mb && b <= p.Bits()+8; b++ {
				q := netip.PrefixFrom(p.Addr(), b)
				out = append(out, fmt.Sprintf("/%d | %s | %s | %s | %s", b, maskString(b, is4), pow2(b-p.Bits()), total(q), usable(q)))
			}
			return out, nil
		},
	})
}
