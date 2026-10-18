package randx

import (
	crand "crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	pwLower   = "abcdefghijklmnopqrstuvwxyz"
	pwUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pwDigits  = "0123456789"
	pwSymbols = "!@#$%^&*()-_=+[]{};:,.?"
	pwAmbig   = "0O1lI|`'\"" // removed with exclude_ambiguous
)

var charsets = map[string]string{
	"alnum":     pwLower + pwUpper + pwDigits,
	"alpha":     pwLower + pwUpper,
	"digits":    pwDigits,
	"numeric":   pwDigits,
	"hex":       "0123456789abcdef",
	"upperhex":  "0123456789ABCDEF",
	"lower":     pwLower,
	"upper":     pwUpper,
	"symbols":   pwSymbols,
	"base32":    "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
	"base58":    "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz",
	"base64url": pwLower + pwUpper + pwDigits + "-_",
	"binary":    "01",
	"all":       pwLower + pwUpper + pwDigits + pwSymbols,
}

type pwPolicy struct {
	length                        int
	lower, upper, digits, symbols bool
	noAmbiguous                   bool
}

func (p pwPolicy) classes() []string {
	var cs []string
	add := func(on bool, set string) {
		if !on {
			return
		}
		if p.noAmbiguous {
			set = strings.Map(func(r rune) rune {
				if strings.ContainsRune(pwAmbig, r) {
					return -1
				}
				return r
			}, set)
		}
		cs = append(cs, set)
	}
	add(p.lower, pwLower)
	add(p.upper, pwUpper)
	add(p.digits, pwDigits)
	add(p.symbols, pwSymbols)
	return cs
}

func (p pwPolicy) generate(r *rand.Rand) string {
	cs := p.classes()
	pool := strings.Join(cs, "")
	out := make([]byte, 0, p.length)
	for _, c := range cs {
		out = append(out, c[r.IntN(len(c))])
	}
	for len(out) < p.length {
		out = append(out, pool[r.IntN(len(pool))])
	}
	out = out[:p.length]
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return string(out)
}

// policyFrom builds the policy the same way internal/builtin/password.go does:
// length is clamped to 4..128 and an empty class selection falls back to lowercase.
func policyFrom(a tools.Args) pwPolicy {
	p := pwPolicy{length: int(a.Int("length")), lower: a.Bool("lower"), upper: a.Bool("upper"),
		digits: a.Bool("digits"), symbols: a.Bool("symbols"), noAmbiguous: a.Bool("exclude_ambiguous")}
	for _, f := range strings.Fields(strings.ToLower(a.Str("args"))) {
		if n, err := strconv.Atoi(f); err == nil {
			p.length = n
			continue
		}
		switch f {
		case "nosym", "nosymbols", "alnum":
			p.symbols = false
		case "nonum", "nodigit", "nodigits":
			p.digits = false
		case "noupper":
			p.upper = false
		case "nolower":
			p.lower = false
		case "noamb", "noambiguous":
			p.noAmbiguous = true
		}
	}
	p.length = min(max(p.length, 4), 128)
	if !p.lower && !p.upper && !p.digits && !p.symbols {
		p.lower = true
	}
	return p
}

func pwParams(extra ...tools.Param) []tools.Param {
	ps := append([]tools.Param(nil), extra...)
	return append(ps,
		p("length", "int", "Characters per password, clamped to 4..128.", int64(20)),
		p("lower", "bool", "Include lowercase letters.", true),
		p("upper", "bool", "Include uppercase letters.", true),
		p("digits", "bool", "Include digits.", true),
		p("symbols", "bool", "Include symbols !@#$%^&*()-_=+[]{};:,.?", true),
		p("exclude_ambiguous", "bool", "Drop look-alike characters 0 O 1 l I | ` ' \".", false),
		argsParam("'24 nosym nonum noupper nolower noamb'."),
	)
}

func init() {
	reg(tools.Func{
		Name: "string", Title: "Random string", Order: 4, Input: "charset", Keywords: []string{"str", "chars"},
		Doc:     "Returns a random string of length characters drawn from a named charset (alnum, alpha, digits, hex, upperhex, lower, upper, symbols, base32, base58, base64url, binary, all) or from any custom characters you pass. Not for secrets unless seed is 0.",
		Params:  []tools.Param{p("length", "int", "Characters (0..100000).", int64(16)), p("charset", "str", "Charset name or the literal characters to draw from.", "alnum"), seedParam()},
		Returns: "str", Example: `rand.string(12, "hex")`,
		Tags: []string{"random text", "random characters", "alphanumeric", "generator", "id"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkLength(a.Int("length"), "length"); err != nil {
				return nil, err
			}
			cs := a.Str("charset")
			if cs == "" {
				cs = "alnum"
			}
			if named, ok := charsets[strings.ToLower(cs)]; ok {
				cs = named
			}
			rs := []rune(cs)
			r := newRng(a.Int("seed"))
			var b strings.Builder
			for i := int64(0); i < a.Int("length"); i++ {
				b.WriteRune(rs[r.IntN(len(rs))])
			}
			return b.String(), nil
		},
	})
	reg(tools.Func{
		Name: "password", Title: "Password", Order: 5, Input: "args", Keywords: []string{"pw", "pass", "password", "secret"},
		Doc:     "Generates a password with crypto/rand, guaranteeing at least one character of every enabled class (lowercase, uppercase, digits, symbols). Length is clamped to 4..128 and if every class is disabled lowercase is used.",
		Params:  pwParams(),
		Returns: "str", Example: `rand.password(24, symbols=False)`,
		Tags: []string{"secure password", "strong password", "random password", "generator", "credentials"},
		Fn: func(a tools.Args) (any, error) {
			return policyFrom(a).generate(cryptoRng()), nil
		},
	})
	reg(tools.Func{
		Name: "passwords", Title: "Several passwords",
		Doc:     "Generates count passwords (1..1000) with crypto/rand using the same rules as password().",
		Params:  pwParams(p("count", "int", "How many passwords.", int64(5))),
		Returns: "list[str]", Example: `rand.passwords(5, 16)`,
		Tags: []string{"bulk passwords", "many passwords", "batch"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("count")
			if n < 0 || n > 1000 {
				return nil, fmt.Errorf("count must be 0..1000")
			}
			pol := policyFrom(a)
			r := cryptoRng()
			out := make([]string, n)
			for i := range out {
				out[i] = pol.generate(r)
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "passphrase", Title: "Passphrase", Order: 6, Input: "args", Keywords: []string{"pp", "phrase", "diceware"},
		Doc:     "Generates a passphrase of random common words using crypto/rand, e.g. \"gold-tree-moon-hand-wave\". Optionally capitalizes words and appends a two-digit number.",
		Params:  []tools.Param{p("words", "int", "Number of words (1..64).", int64(5)), p("separator", "str", "Text between words.", "-"), p("capitalize", "bool", "Capitalize each word.", false), p("number", "bool", "Append a number 0..99.", false), argsParam("a word count such as '6'.")},
		Returns: "str", Example: `rand.passphrase(4, " ", True)`,
		Tags: []string{"diceware", "xkcd", "memorable password", "words"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("words")
			if v := numArgs(a.Str("args")); len(v) > 0 {
				n = v[0]
			}
			if n < 1 || n > 64 {
				return nil, fmt.Errorf("words must be 1..64")
			}
			r := cryptoRng()
			w := make([]string, n)
			for i := range w {
				w[i] = passWords[r.IntN(len(passWords))]
				if a.Bool("capitalize") {
					w[i] = strings.ToUpper(w[i][:1]) + w[i][1:]
				}
			}
			s := strings.Join(w, a.Str("separator"))
			if a.Bool("number") {
				s += a.Str("separator") + strconv.Itoa(r.IntN(100))
			}
			return s, nil
		},
	})
	reg(tools.Func{
		Name: "pin", Title: "Random PIN", Order: 7, Input: "args", Keywords: []string{"pin", "code", "otp"},
		Doc:     "Generates a numeric PIN of the given number of digits (1..64) using crypto/rand; leading zeros are kept.",
		Params:  []tools.Param{p("digits", "int", "Number of digits.", int64(6)), argsParam("a digit count such as '8'.")},
		Returns: "str", Example: `rand.pin(6)`,
		Tags: []string{"numeric code", "otp", "verification code", "passcode"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("digits")
			if v := numArgs(a.Str("args")); len(v) > 0 {
				n = v[0]
			}
			if n < 1 || n > 64 {
				return nil, fmt.Errorf("digits must be 1..64")
			}
			r := cryptoRng()
			b := make([]byte, n)
			for i := range b {
				b[i] = '0' + byte(r.IntN(10))
			}
			return string(b), nil
		},
	})

	tokenParams := []tools.Param{p("bytes", "int", "Number of random bytes (1..50000).", int64(32))}
	token := func(name, title, doc, ex string, enc func([]byte) string) {
		reg(tools.Func{
			Name: name, Title: title, Doc: doc, Params: tokenParams, Returns: "str", Example: ex,
			Tags: []string{"token", "secret", "api key", "random bytes", "crypto", "session id", "csprng"},
			Fn: func(a tools.Args) (any, error) {
				n := a.Int("bytes")
				if n < 1 || n > maxLength/2 {
					return nil, fmt.Errorf("bytes must be 1..%d", maxLength/2)
				}
				b := make([]byte, n)
				_, _ = crand.Read(b)
				return enc(b), nil
			},
		})
	}
	token("token_hex", "Random token (hex)", "Returns bytes cryptographically random bytes as lowercase hex (2 characters per byte).", `rand.token_hex(16)`, hex.EncodeToString)
	token("token_base64url", "Random token (base64url)", "Returns bytes cryptographically random bytes as unpadded URL-safe base64.", `rand.token_base64url(32)`, base64.RawURLEncoding.EncodeToString)
	token("token_base32", "Random token (base32)", "Returns bytes cryptographically random bytes as unpadded uppercase RFC 4648 base32, handy for TOTP secrets.", `rand.token_base32(20)`,
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString)
	reg(tools.Func{
		Name: "bytes_hex", Title: "Random bytes (hex)",
		Doc:     "Returns count random bytes as hex. Uses crypto/rand when seed is 0, otherwise a reproducible pseudo-random stream (not for secrets).",
		Params:  []tools.Param{p("count", "int", "Number of bytes (1..50000).", int64(16)), seedParam()},
		Returns: "str", Example: `rand.bytes_hex(8)`,
		Tags: []string{"random bytes", "hex dump", "nonce", "iv", "salt"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("count")
			if n < 1 || n > maxLength/2 {
				return nil, fmt.Errorf("count must be 1..%d", maxLength/2)
			}
			r := newRng(a.Int("seed"))
			b := make([]byte, n)
			for i := range b {
				b[i] = byte(r.Uint32())
			}
			return hex.EncodeToString(b), nil
		},
	})
}
