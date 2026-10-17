package builtin

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

type password struct{ base }

func newPassword() *password { return &password{base{"password"}} }

func (p *password) Ranked() bool       { return true }
func (p *password) Keywords() []string { return []string{"pw", "pass", "pp", "pin"} }

type pwPolicy struct {
	Length                        int
	Lower, Upper, Digits, Symbols bool
	NoAmbiguous                   bool
}

const (
	pwLower   = "abcdefghijklmnopqrstuvwxyz"
	pwUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	pwDigits  = "0123456789"
	pwSymbols = "!@#$%^&*()-_=+[]{};:,.?"
	pwAmbig   = "0O1lI|`'\"" // removed with "noamb"
)

// parsePwArgs reads "24 nosym nonum noupper nolower noamb".
func parsePwArgs(args string) pwPolicy {
	p := pwPolicy{Length: 20, Lower: true, Upper: true, Digits: true, Symbols: true}
	for _, f := range strings.Fields(strings.ToLower(args)) {
		if n, err := strconv.Atoi(f); err == nil {
			p.Length = n
			continue
		}
		switch f {
		case "nosym", "nosymbols", "alnum":
			p.Symbols = false
		case "nonum", "nodigit", "nodigits":
			p.Digits = false
		case "noupper":
			p.Upper = false
		case "nolower":
			p.Lower = false
		case "noamb", "noambiguous":
			p.NoAmbiguous = true
		}
	}
	if p.Length < 4 {
		p.Length = 4
	}
	if p.Length > 128 {
		p.Length = 128
	}
	if !p.Lower && !p.Upper && !p.Digits && !p.Symbols {
		p.Lower = true
	}
	return p
}

func (p pwPolicy) classes() []string {
	var cs []string
	add := func(on bool, set string) {
		if !on {
			return
		}
		if p.NoAmbiguous {
			set = strings.Map(func(r rune) rune {
				if strings.ContainsRune(pwAmbig, r) {
					return -1
				}
				return r
			}, set)
		}
		cs = append(cs, set)
	}
	add(p.Lower, pwLower)
	add(p.Upper, pwUpper)
	add(p.Digits, pwDigits)
	add(p.Symbols, pwSymbols)
	return cs
}

// entropy returns the entropy in bits of a password drawn from the policy's pool.
func (p pwPolicy) entropy() float64 {
	n := 0
	for _, c := range p.classes() {
		n += len(c)
	}
	return float64(p.Length) * math.Log2(float64(n))
}

func randInt(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return int(v.Int64())
}

// generatePassword draws every character with crypto/rand and guarantees one
// character from each enabled class.
func generatePassword(p pwPolicy) string {
	cs := p.classes()
	pool := strings.Join(cs, "")
	out := make([]byte, 0, p.Length)
	for _, c := range cs {
		out = append(out, c[randInt(len(c))])
	}
	for len(out) < p.Length {
		out = append(out, pool[randInt(len(pool))])
	}
	out = out[:p.Length]
	for i := len(out) - 1; i > 0; i-- { // Fisher-Yates
		j := randInt(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func generatePIN(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '0' + byte(randInt(10))
	}
	return string(b)
}

// generatePassphrase joins n random words.
func generatePassphrase(n int, sep string, capitalize, number bool) string {
	words := make([]string, n)
	for i := range words {
		w := passWords[randInt(len(passWords))]
		if capitalize {
			w = strings.ToUpper(w[:1]) + w[1:]
		}
		words[i] = w
	}
	s := strings.Join(words, sep)
	if number {
		s += sep + strconv.Itoa(randInt(100))
	}
	return s
}

func (p *password) Query(ctx context.Context, q string) []core.Item {
	rest, kw, ok := trigger(q, p.Keywords()...)
	if !ok {
		rest, kw, ok = pwPrefix(q)
	}
	if !ok {
		return nil
	}
	mk := func(i int, secret, sub string) core.Item {
		return core.Item{
			ID: fmt.Sprintf("password:ephemeral:%d", i), Title: secret, Subtitle: sub, Category: "Password",
			Action: noHistoryAction(copyAction(secret)), Refresh: true,
		}
	}
	var out []core.Item
	switch strings.ToLower(kw) {
	case "pw", "pass":
		pol := parsePwArgs(rest)
		sub := fmt.Sprintf("%d chars, ~%.0f bits - Enter copies", pol.Length, pol.entropy())
		for i := 0; i < 5; i++ {
			out = append(out, mk(i, generatePassword(pol), sub))
		}
	case "pp":
		n, cap_, num := 5, false, false
		for _, f := range strings.Fields(strings.ToLower(rest)) {
			if v, err := strconv.Atoi(f); err == nil && v >= 2 && v <= 20 {
				n = v
			}
			cap_ = cap_ || f == "cap"
			num = num || f == "num"
		}
		sub := fmt.Sprintf("%d words, ~%.0f bits - Enter copies", n, float64(n)*math.Log2(float64(len(passWords))))
		for i := 0; i < 4; i++ {
			out = append(out, mk(i, generatePassphrase(n, "-", cap_, num), sub))
		}
	case "pin":
		n := 6
		if v, err := strconv.Atoi(strings.TrimSpace(rest)); err == nil && v >= 3 && v <= 32 {
			n = v
		}
		for i := 0; i < 4; i++ {
			out = append(out, mk(i, generatePIN(n), fmt.Sprintf("%d digits - Enter copies", n)))
		}
	}
	return out
}

// passWords is the passphrase word list (short, common, easy to type).
var passWords = strings.Fields(`able acid aged also area army away baby back ball band bank base bath bear beat
been beer bell belt best bill bird blow blue boat body bold bone book boom born boss both bowl bulk burn bush busy
cafe cake call calm came camp card care case cash cast cave cell chat chef chip city clay club coal coat code coin
cold come cook cool cope copy core corn cost crew crop dark data date dawn days dead deal dean dear debt deep deer
desk dial diet dirt dish disk dive door dose down draw drew drop drum dual duck duke dust duty each earn ease east
easy edge else even ever evil exit face fact fail fair fall farm fast fate fear feed feel feet fell felt file fill
film find fine fire firm fish five flat flow food foot ford form fort four free from fuel full fund gain game gate
gave gear gene gift girl give glad goal goes gold golf gone good gray grew grid grow gulf hair half hall hand hang
hard harm hate have head hear heat held hell help here hero high hill hint hold hole holy home hope horn host hour
huge hunt hurt idea inch into iron item jack jade jazz join jump jury just keen keep kept kick kind king knee knew
know lack lady laid lake lamp land lane last late lawn lead leaf lean left lend less life lift like lime line link
lion list live load loan lock logo long look loop lord lose loss lost loud love luck made mail main make male mall
many mark mask mass matt meal mean meat meet melt menu mere mild mile milk mill mind mine mint miss mode moon more
most move much must myth name navy near neat neck need news next nice nine node none noon norm nose note noun odds
okay once only onto open oral oven pace pack page paid pain pair palm park part pass past path peak pick pier pile
pine pink pipe plan play plot plus poem poet pole poll pond pool poor port post pour pray pull pure push quit race
rack rail rain rank rare rate read real rely rent rest rice rich ride ring rise risk road rock role roll roof room
root rope rose rule runs rush safe sail salt same sand save seat seed seek seem sell send sent ship shop shot show
shut sick side sign silk sing site size skin slow snow soft soil sold sole some song soon sort soul soup span spot
star stay stem step stop such suit sure swim tail take tale talk tall tank tape task team tear tell tent term test
text than that them then they thin this thus tide tidy tile time tiny tire toad told tone took tool tops tour town
tree trim trip true tube tune turn twin type ugly unit upon urge used user vary vast very view vote wage wait wake
walk wall want warm wash wave ways weak wear week well went were west what when whom wide wife wild will wind wine
wing wire wise wish with wolf wood wool word wore work worn wrap yard yarn year yoga zero zone zoom`)

// pwPrefix accepts "passw", "passwo", ... "password" as the password keyword and
// "passp" ... "passphrase" as the passphrase keyword, so the generators answer
// while the full word is being typed.
func pwPrefix(q string) (rest, kw string, ok bool) {
	q = strings.TrimLeft(q, " 	")
	first, rest, _ := strings.Cut(q, " ")
	rest = strings.TrimSpace(rest)
	f := strings.ToLower(first)
	switch {
	case len(f) >= 4 && strings.HasPrefix("password", f):
		return rest, "pw", true
	case len(f) >= 5 && strings.HasPrefix("passphrase", f):
		return rest, "pp", true
	}
	return "", "", false
}
