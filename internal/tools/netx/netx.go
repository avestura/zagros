// Package netx registers the "net" tool group: CIDR and subnet math, IP, MAC,
// port, protocol, HTTP, MIME, URL, domain and e-mail helpers. Every function
// is pure: nothing here resolves DNS or touches the network.
package netx

import (
	"fmt"
	"math/big"
	"net/netip"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxEnum  = 65536  // subnet / host enumeration cap
	maxItems = 100000 // result cap
)

// Launcher keyword sets per category.
var (
	kwCIDR   = []string{"net", "cidr", "subnet", "ip"}
	kwIP     = []string{"net", "ip"}
	kwMAC    = []string{"net", "mac"}
	kwPort   = []string{"net", "port"}
	kwStatus = []string{"net", "status", "http"}
	kwMime   = []string{"net", "mime"}
	kwURL    = []string{"net", "url"}
	kwDomain = []string{"net", "domain"}
	kwBand   = []string{"net", "bandwidth"}
)

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

func opt(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

// reg registers f in the "net" group and launcher family with the keywords kw.
func reg(kw []string, f tools.Func) {
	f.Group = "net"
	f.Family = "net"
	if f.Order == 0 {
		f.Order = 100
	}
	f.Keywords = kw
	hasStr := false
	for _, p := range f.Params {
		if p.Type == "str" {
			hasStr = true
		}
	}
	if !hasStr {
		f.Input = ""
		// the launcher needs a text input; list-only functions stay API-only
		f.Family = ""
	}
	tools.Register(f)
}

func pCIDR(doc string) tools.Param { return req("cidr", "str", doc) }
func pIP(doc string) tools.Param   { return req("ip", "str", doc) }

const cidrForms = `"10.0.0.5/24", "10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255" (wildcard mask), IPv6 like "2001:db8::/32", or a bare IP (/32 or /128).`

// ---------------------------------------------------------------- parsing

func bigOf(a netip.Addr) *big.Int { return new(big.Int).SetBytes(a.AsSlice()) }

func addrOf(n *big.Int, is4 bool) (netip.Addr, bool) {
	size := 16
	if is4 {
		size = 4
	}
	if n.Sign() < 0 || n.BitLen() > size*8 {
		return netip.Addr{}, false
	}
	buf := make([]byte, size)
	n.FillBytes(buf)
	a, _ := netip.AddrFromSlice(buf)
	return a, true
}

func maxBits(a netip.Addr) int {
	if a.Is4() {
		return 32
	}
	return 128
}

func pow2(n int) *big.Int { return new(big.Int).Lsh(big.NewInt(1), uint(n)) }

var dottedRe = regexp.MustCompile(`^\d{1,3}(?:\.\d{1,3}){3}$`)

// parseAddr parses an IPv4 or IPv6 address. It tolerates surrounding space,
// [brackets], a zone suffix and leading zeros in IPv4 octets.
func parseAddr(s string) (netip.Addr, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.TrimPrefix(s, "["), "]")
	if s == "" {
		return netip.Addr{}, fmt.Errorf("empty IP address")
	}
	a, err := netip.ParseAddr(s)
	if err != nil && dottedRe.MatchString(s) {
		parts := strings.Split(s, ".")
		for i, p := range parts {
			n, _ := strconv.Atoi(p)
			parts[i] = strconv.Itoa(n)
		}
		a, err = netip.ParseAddr(strings.Join(parts, "."))
	}
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid IP address %q", s)
	}
	return a.WithZone(""), nil
}

// maskBits converts a dotted IPv4 netmask or wildcard mask to a prefix length.
func maskBits(a netip.Addr) (int, bool) {
	if !a.Is4() {
		return 0, false
	}
	b := a.As4()
	x := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	if n, ok := contiguous(x); ok {
		return n, true
	}
	if n, ok := contiguous(^x); ok {
		return n, true
	}
	return 0, false
}

// contiguous reports the number of leading one bits when x is a valid netmask.
func contiguous(x uint32) (int, bool) {
	n := 0
	for n < 32 && x&(1<<(31-uint(n))) != 0 {
		n++
	}
	if n == 32 {
		return 32, true
	}
	if x<<uint(n) != 0 {
		return 0, false
	}
	return n, true
}

// parseNet parses every CIDR spelling Zagros accepts; see cidrForms. The
// returned prefix is the masked network; addr is the address as written.
func parseNet(s string) (netip.Addr, netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}, netip.Prefix{}, fmt.Errorf("empty CIDR")
	}
	var left, right string
	switch {
	case strings.Contains(s, "/"):
		i := strings.Index(s, "/")
		left, right = s[:i], strings.TrimSpace(s[i+1:])
	case len(strings.Fields(s)) == 2:
		f := strings.Fields(s)
		left, right = f[0], f[1]
	default:
		left = s
	}
	a, err := parseAddr(left)
	if err != nil {
		return a, netip.Prefix{}, err
	}
	bits := maxBits(a)
	if right != "" {
		right = strings.TrimPrefix(right, "/")
		if n, err := strconv.Atoi(right); err == nil && !strings.Contains(right, ".") {
			if n < 0 || n > maxBits(a) {
				return a, netip.Prefix{}, fmt.Errorf("prefix length %d out of range 0-%d", n, maxBits(a))
			}
			bits = n
		} else if m, err := parseAddr(right); err == nil && a.Is4() && m.Is4() {
			n, ok := maskBits(m)
			if !ok {
				return a, netip.Prefix{}, fmt.Errorf("%q is not a valid netmask or wildcard mask", right)
			}
			bits = n
		} else {
			return a, netip.Prefix{}, fmt.Errorf("invalid prefix or mask %q", right)
		}
	}
	p, err := a.Prefix(bits)
	if err != nil {
		return a, p, err
	}
	return a, p, nil
}

func mustNet(s string) (netip.Prefix, error) {
	_, p, err := parseNet(s)
	return p, err
}

// firstLast returns the first and last address of p.
func firstLast(p netip.Prefix) (netip.Addr, netip.Addr) {
	p = p.Masked()
	first := p.Addr()
	n := bigOf(first)
	n.Add(n, pow2(maxBits(first)-p.Bits()))
	n.Sub(n, big.NewInt(1))
	last, _ := addrOf(n, first.Is4())
	return first, last
}

func total(p netip.Prefix) *big.Int { return pow2(maxBits(p.Addr()) - p.Bits()) }

func okNet(s string) bool {
	_, _, err := parseNet(s)
	return err == nil
}

func okIP(s string) bool {
	_, err := parseAddr(s)
	return err == nil
}

// ----------------------------------------------------------- range helpers

type ival struct{ lo, hi *big.Int }

// rangeToPrefixes returns the minimal list of CIDRs covering [lo, hi].
func rangeToPrefixes(lo, hi *big.Int, is4 bool) []netip.Prefix {
	bitsMax := 128
	if is4 {
		bitsMax = 32
	}
	var out []netip.Prefix
	cur := new(big.Int).Set(lo)
	one := big.NewInt(1)
	for cur.Cmp(hi) <= 0 && len(out) < maxItems {
		k := bitsMax
		if cur.Sign() != 0 {
			k = int(cur.TrailingZeroBits())
			if k > bitsMax {
				k = bitsMax
			}
		}
		span := new(big.Int).Sub(hi, cur)
		span.Add(span, one)
		if fit := span.BitLen() - 1; fit < k {
			k = fit
		}
		a, _ := addrOf(cur, is4)
		p, _ := a.Prefix(bitsMax - k)
		out = append(out, p)
		cur.Add(cur, pow2(k))
	}
	return out
}

func prefixStrings(ps []netip.Prefix) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.String()
	}
	return out
}

// parseRangeItem parses "cidr", "ip", "a-b" or "a - b" into an interval.
func parseRangeItem(s string) (ival, bool, error) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "-"); i > 0 && !strings.Contains(s, "/") {
		a, err := parseAddr(s[:i])
		if err != nil {
			return ival{}, false, err
		}
		b, err := parseAddr(s[i+1:])
		if err != nil {
			return ival{}, false, err
		}
		if a.Is4() != b.Is4() {
			return ival{}, false, fmt.Errorf("range %q mixes IPv4 and IPv6", s)
		}
		lo, hi := bigOf(a), bigOf(b)
		if lo.Cmp(hi) > 0 {
			lo, hi = hi, lo
		}
		return ival{lo, hi}, a.Is4(), nil
	}
	_, p, err := parseNet(s)
	if err != nil {
		return ival{}, false, err
	}
	f, l := firstLast(p)
	return ival{bigOf(f), bigOf(l)}, f.Is4(), nil
}

func mergeIvals(in []ival) []ival {
	if len(in) == 0 {
		return nil
	}
	s := append([]ival(nil), in...)
	sort.Slice(s, func(i, j int) bool { return s[i].lo.Cmp(s[j].lo) < 0 })
	out := []ival{{new(big.Int).Set(s[0].lo), new(big.Int).Set(s[0].hi)}}
	for _, v := range s[1:] {
		last := &out[len(out)-1]
		next := new(big.Int).Add(last.hi, big.NewInt(1))
		if v.lo.Cmp(next) <= 0 {
			if v.hi.Cmp(last.hi) > 0 {
				last.hi = new(big.Int).Set(v.hi)
			}
			continue
		}
		out = append(out, ival{new(big.Int).Set(v.lo), new(big.Int).Set(v.hi)})
	}
	return out
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
