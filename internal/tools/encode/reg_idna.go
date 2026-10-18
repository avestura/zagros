package encode

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/net/idna"
)

// ---- raw punycode, RFC 3492 ----

const (
	pcBase        = 36
	pcTMin        = 1
	pcTMax        = 26
	pcSkew        = 38
	pcDamp        = 700
	pcInitialBias = 72
	pcInitialN    = 128
	pcMaxInt      = 1<<31 - 1
)

func pcAdapt(delta, numPoints int, first bool) int {
	if first {
		delta /= pcDamp
	} else {
		delta /= 2
	}
	delta += delta / numPoints
	k := 0
	for delta > ((pcBase-pcTMin)*pcTMax)/2 {
		delta /= pcBase - pcTMin
		k += pcBase
	}
	return k + (pcBase-pcTMin+1)*delta/(delta+pcSkew)
}

func pcDigit(d int) byte {
	if d < 26 {
		return byte('a' + d)
	}
	return byte('0' + d - 26)
}

func pcThreshold(k, bias int) int {
	switch {
	case k <= bias:
		return pcTMin
	case k >= bias+pcTMax:
		return pcTMax
	}
	return k - bias
}

func punyEncode(s string) (string, error) {
	rs := []rune(s)
	for _, r := range rs {
		if r == utf8.RuneError || r >= 0xD800 && r <= 0xDFFF {
			return "", fmt.Errorf("invalid character in punycode input")
		}
	}
	n, delta, bias := pcInitialN, 0, pcInitialBias
	var out []byte
	for _, r := range rs {
		if r < 0x80 {
			out = append(out, byte(r))
		}
	}
	b := len(out)
	h := b
	if b > 0 {
		out = append(out, '-')
	}
	for h < len(rs) {
		m := pcMaxInt
		for _, r := range rs {
			if int(r) >= n && int(r) < m {
				m = int(r)
			}
		}
		if m-n > (pcMaxInt-delta)/(h+1) {
			return "", fmt.Errorf("punycode overflow")
		}
		delta += (m - n) * (h + 1)
		n = m
		for _, r := range rs {
			if int(r) < n {
				delta++
				if delta > pcMaxInt {
					return "", fmt.Errorf("punycode overflow")
				}
			}
			if int(r) == n {
				q := delta
				for k := pcBase; ; k += pcBase {
					t := pcThreshold(k, bias)
					if q < t {
						break
					}
					out = append(out, pcDigit(t+(q-t)%(pcBase-t)))
					q = (q - t) / (pcBase - t)
				}
				out = append(out, pcDigit(q))
				bias = pcAdapt(delta, h+1, h == b)
				delta = 0
				h++
			}
		}
		delta++
		n++
	}
	return string(out), nil
}

func punyDecode(s string) (string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return "", fmt.Errorf("punycode input must be ASCII")
		}
	}
	var out []rune
	b := strings.LastIndexByte(s, '-')
	if b > 0 {
		for _, c := range []byte(s[:b]) {
			out = append(out, rune(c))
		}
		s = s[b+1:]
	} else if b == 0 {
		s = s[1:]
	}
	n, i, bias := pcInitialN, 0, pcInitialBias
	for p := 0; p < len(s); {
		oldi, w := i, 1
		for k := pcBase; ; k += pcBase {
			if p >= len(s) {
				return "", fmt.Errorf("truncated punycode")
			}
			c := s[p]
			p++
			var d int
			switch {
			case c >= '0' && c <= '9':
				d = int(c-'0') + 26
			case c >= 'a' && c <= 'z':
				d = int(c - 'a')
			case c >= 'A' && c <= 'Z':
				d = int(c - 'A')
			default:
				return "", fmt.Errorf("invalid punycode digit %q", string(rune(c)))
			}
			if d > (pcMaxInt-i)/w {
				return "", fmt.Errorf("punycode overflow")
			}
			i += d * w
			t := pcThreshold(k, bias)
			if d < t {
				break
			}
			if w > pcMaxInt/(pcBase-t) {
				return "", fmt.Errorf("punycode overflow")
			}
			w *= pcBase - t
		}
		bias = pcAdapt(i-oldi, len(out)+1, oldi == 0)
		if i/(len(out)+1) > pcMaxInt-n {
			return "", fmt.Errorf("punycode overflow")
		}
		n += i / (len(out) + 1)
		i %= len(out) + 1
		if n > 0x10FFFF || n >= 0xD800 && n <= 0xDFFF {
			return "", fmt.Errorf("punycode decodes to an invalid code point")
		}
		out = append(out, 0)
		copy(out[i+1:], out[i:])
		out[i] = rune(n)
		i++
	}
	return string(out), nil
}

func idnaProfile(name string) (*idna.Profile, error) {
	switch normName(name) {
	case "", "lookup":
		return idna.Lookup, nil
	case "registration":
		return idna.Registration, nil
	case "display":
		return idna.Display, nil
	case "punycode", "raw":
		return idna.Punycode, nil
	}
	return nil, fmt.Errorf("unknown profile %q (use lookup, registration, display or punycode)", name)
}

func hasXN(s string) bool {
	return strings.Contains(strings.ToLower(s), "xn--")
}

func toASCIIDomain(d, profile string) (string, error) {
	p, err := idnaProfile(profile)
	if err != nil {
		return "", err
	}
	d = strings.TrimSpace(d)
	return p.ToASCII(d)
}

func toUnicodeDomain(d, profile string) (string, error) {
	p, err := idnaProfile(profile)
	if err != nil {
		return "", err
	}
	d = strings.TrimSpace(d)
	r, err := p.ToUnicode(d)
	if err != nil {
		if r2, err2 := idna.Punycode.ToUnicode(d); err2 == nil {
			return r2, nil
		}
		return "", err
	}
	return r, nil
}

func splitEmail(e string) (local, domain string, err error) {
	e = strings.TrimSpace(e)
	i := strings.LastIndexByte(e, '@')
	if i < 0 {
		return "", "", fmt.Errorf("not an email address (no @)")
	}
	return e[:i], e[i+1:], nil
}

var idnaProfiles = []string{"lookup", "registration", "display", "punycode"}

func init() {
	reg(tools.Func{
		Name: "punycode_encode", Title: "Punycode encode (RFC 3492)",
		Doc:     "Raw RFC 3492 Punycode of one string (a domain label): \"bücher\" gives \"bcher-kva\". No case folding or normalization; use idna_to_ascii for whole domains. prefix=true prepends \"xn--\". Pure-ASCII input yields the text plus a trailing hyphen.",
		Params:  []tools.Param{pReq("text", "Label to encode."), pBool("prefix", "Prepend the ACE prefix xn--.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s, err := punyEncode(a.Str("text"))
			if err != nil {
				return nil, err
			}
			if a.Bool("prefix") {
				s = "xn--" + s
			}
			return s, nil
		},
		Example:  `encode.punycode_encode("bücher")`,
		Tags:     tagsWith("punycode", "rfc3492", "idn", "xn--", "domain", "unicode to ascii", "label"),
		Keywords: []string{"punycode", "idn"},
	})
	reg(tools.Func{
		Name: "punycode_decode", Title: "Punycode decode (RFC 3492)",
		Doc:     "Raw RFC 3492 Punycode decode of one label; a leading \"xn--\" is removed first. \"bcher-kva\" gives \"bücher\". Fails on invalid digits, overflow or truncated input.",
		Params:  []tools.Param{pReq("text", "Punycode label, with or without xn--.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("text"))
			if len(s) >= 4 && strings.EqualFold(s[:4], "xn--") {
				s = s[4:]
			}
			return punyDecode(s)
		},
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			if len(s) < 5 || strings.ContainsAny(s, " \t\n@/") || !strings.EqualFold(s[:4], "xn--") {
				return false
			}
			r, err := punyDecode(s[4:])
			return err == nil && r != "" && !isASCII(r)
		},
		Example:  `encode.punycode_decode("xn--bcher-kva")`,
		Tags:     tagsWith("punycode", "rfc3492", "idn", "xn--", "domain", "ascii to unicode", "label", "decode"),
		Keywords: []string{"punycode", "idn"},
	})
	reg(tools.Func{
		Name: "idna_to_ascii", Title: "Domain to ASCII (IDNA / punycode)",
		Doc:     "Converts an internationalized domain name to ASCII: each non-ASCII label becomes xn-- punycode, with IDNA mapping (lower-casing, normalization) and validation according to the profile (lookup by default). \"münchen.de\" gives \"xn--mnchen-3ya.de\".",
		Params:  []tools.Param{pReq("domain", "Domain name such as \"bücher.example\"."), pStr("profile", "IDNA profile; punycode skips mapping and validation.", "lookup", idnaProfiles...)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return toASCIIDomain(a.Str("domain"), a.Str("profile")) },
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			if s == "" || isASCII(s) || strings.ContainsAny(s, " \n") {
				return false
			}
			_, err := toASCIIDomain(s, "lookup")
			return err == nil
		},
		Example:  `encode.idna_to_ascii("münchen.de")`,
		Tags:     tagsWith("idna", "idn", "punycode", "domain", "xn--", "internationalized domain", "unicode to ascii", "toascii"),
		Keywords: []string{"punycode", "idn", "idna"},
	})
	reg(tools.Func{
		Name: "idna_to_unicode", Title: "Domain to Unicode (IDNA / punycode)",
		Doc:     "Converts a domain with xn-- labels to its Unicode form: \"xn--mnchen-3ya.de\" gives \"münchen.de\". Labels without xn-- are kept. If strict validation fails the labels are still decoded with raw punycode when possible.",
		Params:  []tools.Param{pReq("domain", "Domain name such as \"xn--bcher-kva.example\"."), pStr("profile", "IDNA profile; punycode skips validation.", "lookup", idnaProfiles...)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return toUnicodeDomain(a.Str("domain"), a.Str("profile")) },
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			if !hasXN(s) || strings.ContainsAny(s, " \n") {
				return false
			}
			r, err := toUnicodeDomain(s, "lookup")
			return err == nil && r != s
		},
		Example:  `encode.idna_to_unicode("xn--mnchen-3ya.de")`,
		Tags:     tagsWith("idna", "idn", "punycode", "domain", "xn--", "internationalized domain", "ascii to unicode", "tounicode", "decode"),
		Keywords: []string{"punycode", "idn", "idna"},
	})
	reg(tools.Func{
		Name: "email_to_ascii", Title: "Email to ASCII (IDN domain)",
		Doc:     "Converts the domain part of an email address to ASCII punycode (user@bücher.de gives user@xn--bcher-kva.de). The local part before the last @ is left unchanged, since internationalized local parts need SMTPUTF8.",
		Params:  []tools.Param{pReq("email", "Email address."), pStr("profile", "IDNA profile.", "lookup", idnaProfiles...)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			l, d, err := splitEmail(a.Str("email"))
			if err != nil {
				return nil, err
			}
			ad, err := toASCIIDomain(d, a.Str("profile"))
			if err != nil {
				return nil, err
			}
			return l + "@" + ad, nil
		},
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			_, d, err := splitEmail(s)
			return err == nil && !isASCII(d) && !strings.ContainsAny(s, " \n")
		},
		Example:  `encode.email_to_ascii("info@bücher.de")`,
		Tags:     tagsWith("email", "idn", "idna", "punycode", "domain", "internationalized", "smtputf8"),
		Keywords: []string{"punycode", "idn", "email"},
	})
	reg(tools.Func{
		Name: "email_to_unicode", Title: "Email to Unicode (IDN domain)",
		Doc:     "Converts the xn-- domain part of an email address to Unicode (user@xn--bcher-kva.de gives user@bücher.de). The local part before the last @ is left unchanged.",
		Params:  []tools.Param{pReq("email", "Email address."), pStr("profile", "IDNA profile.", "lookup", idnaProfiles...)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			l, d, err := splitEmail(a.Str("email"))
			if err != nil {
				return nil, err
			}
			ud, err := toUnicodeDomain(d, a.Str("profile"))
			if err != nil {
				return nil, err
			}
			return l + "@" + ud, nil
		},
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			_, d, err := splitEmail(s)
			return err == nil && hasXN(d) && !strings.ContainsAny(s, " \n")
		},
		Example:  `encode.email_to_unicode("info@xn--bcher-kva.de")`,
		Tags:     tagsWith("email", "idn", "idna", "punycode", "domain", "internationalized", "decode"),
		Keywords: []string{"punycode", "idn", "email"},
	})
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
