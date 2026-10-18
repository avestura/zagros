package cipherx

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }

// mapLetters applies f(letterIndex, value 0..25) to each ASCII letter,
// keeping case and passing everything else through unchanged.
func mapLetters(text string, f func(i, v int) int) string {
	var sb strings.Builder
	i := 0
	for _, r := range text {
		switch {
		case isUpper(r):
			sb.WriteRune('A' + rune(mod(f(i, int(r-'A')), 26)))
			i++
		case isLower(r):
			sb.WriteRune('a' + rune(mod(f(i, int(r-'a')), 26)))
			i++
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func mod(a, m int) int { return ((a % m) + m) % m }

func lettersOf(s string) []int {
	var out []int
	for _, r := range s {
		switch {
		case isUpper(r):
			out = append(out, int(r-'A'))
		case isLower(r):
			out = append(out, int(r-'a'))
		}
	}
	return out
}

func caesar(text string, shift int64) string {
	s := int(mod(int(shift%26), 26))
	return mapLetters(text, func(_, v int) int { return v + s })
}

func atbash(text string) string { return mapLetters(text, func(_, v int) int { return 25 - v }) }

func rot5(text string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '0' + (r-'0'+5)%10
		}
		return r
	}, text)
}

func rot47(text string) string {
	return strings.Map(func(r rune) rune {
		if r >= 33 && r <= 126 {
			return 33 + (r-33+47)%94
		}
		return r
	}, text)
}

func vigenere(text, key string, dec bool) (string, error) {
	k := lettersOf(key)
	if len(k) == 0 {
		return "", errors.New("key must contain at least one letter")
	}
	return mapLetters(text, func(i, v int) int {
		if dec {
			return v - k[i%len(k)]
		}
		return v + k[i%len(k)]
	}), nil
}

func autokey(text, key string, dec bool) (string, error) {
	stream := lettersOf(key)
	if len(stream) == 0 {
		return "", errors.New("key must contain at least one letter")
	}
	return mapLetters(text, func(i, v int) int {
		if dec {
			p := mod(v-stream[i], 26)
			stream = append(stream, p)
			return p
		}
		stream = append(stream, v)
		return v + stream[i]
	}), nil
}

func beaufort(text, key string) (string, error) {
	k := lettersOf(key)
	if len(k) == 0 {
		return "", errors.New("key must contain at least one letter")
	}
	return mapLetters(text, func(i, v int) int { return k[i%len(k)] - v }), nil
}

func modInv26(a int) (int, bool) {
	for x := 1; x < 26; x += 2 {
		if (a*x)%26 == 1 {
			return x, true
		}
	}
	return 0, false
}

func affine(text string, a, b int64, dec bool) (string, error) {
	aa, bb := mod(int(a%26), 26), mod(int(b%26), 26)
	inv, ok := modInv26(aa)
	if !ok {
		return "", fmt.Errorf("a must be coprime to 26 (odd and not 13); got %d", a)
	}
	return mapLetters(text, func(_, v int) int {
		if dec {
			return inv * (v - bb)
		}
		return aa*v + bb
	}), nil
}

func substitution(text, key string, dec bool) (string, error) {
	k := strings.ToUpper(strings.TrimSpace(key))
	if len(k) != 26 {
		return "", errors.New("key must be 26 distinct letters (the cipher alphabet for A..Z)")
	}
	var fwd, back [26]int
	var seen [26]bool
	for i, r := range k {
		if !isUpper(r) || seen[r-'A'] {
			return "", errors.New("key must be 26 distinct letters (the cipher alphabet for A..Z)")
		}
		seen[r-'A'] = true
		fwd[i] = int(r - 'A')
		back[r-'A'] = i
	}
	return mapLetters(text, func(_, v int) int {
		if dec {
			return back[v]
		}
		return fwd[v]
	}), nil
}

// permute: cipher[j] = plain[order[j]].
func permute(text string, order []int) string {
	r := []rune(text)
	out := make([]rune, len(r))
	for j, o := range order {
		out[j] = r[o]
	}
	return string(out)
}

func unpermute(text string, order []int) string {
	r := []rune(text)
	out := make([]rune, len(r))
	for j, o := range order {
		out[o] = r[j]
	}
	return string(out)
}

func railOrder(n, rails int) []int {
	if rails < 2 || rails >= n {
		o := make([]int, n)
		for i := range o {
			o[i] = i
		}
		return o
	}
	rail := make([]int, n)
	r, d := 0, 1
	for i := 0; i < n; i++ {
		rail[i] = r
		if r == 0 {
			d = 1
		} else if r == rails-1 {
			d = -1
		}
		r += d
	}
	o := make([]int, 0, n)
	for k := 0; k < rails; k++ {
		for i := 0; i < n; i++ {
			if rail[i] == k {
				o = append(o, i)
			}
		}
	}
	return o
}

func columnOrder(n int, key []rune) []int {
	k := len(key)
	idx := make([]int, k)
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return key[idx[a]] < key[idx[b]] })
	o := make([]int, 0, n)
	for _, c := range idx {
		for p := c; p < n; p += k {
			o = append(o, p)
		}
	}
	return o
}

func scytaleOrder(n, d int) []int {
	o := make([]int, 0, n)
	for c := 0; c < d; c++ {
		for p := c; p < n; p += d {
			o = append(o, p)
		}
	}
	return o
}

// ---- Playfair ----

func playfairSquare(key string) [25]byte {
	var sq [25]byte
	var used [26]bool
	n := 0
	add := func(r rune) {
		if r == 'J' {
			r = 'I'
		}
		if isUpper(r) && !used[r-'A'] {
			used[r-'A'] = true
			sq[n] = byte(r)
			n++
		}
	}
	for _, r := range strings.ToUpper(key) {
		add(r)
	}
	for r := 'A'; r <= 'Z'; r++ {
		add(r)
	}
	return sq
}

func playfair(text, key string, dec bool) (string, error) {
	if len(lettersOf(key)) == 0 {
		return "", errors.New("key must contain at least one letter")
	}
	sq := playfairSquare(key)
	pos := map[byte]int{}
	for i, c := range sq {
		pos[c] = i
	}
	var in []byte
	for _, r := range strings.ToUpper(text) {
		if isUpper(r) {
			if r == 'J' {
				r = 'I'
			}
			in = append(in, byte(r))
		}
	}
	var pairs [][2]byte
	if dec {
		if len(in)%2 != 0 {
			return "", errors.New("playfair ciphertext must have an even number of letters")
		}
		for i := 0; i < len(in); i += 2 {
			pairs = append(pairs, [2]byte{in[i], in[i+1]})
		}
	} else {
		for i := 0; i < len(in); {
			a := in[i]
			b := byte('X')
			if i+1 < len(in) && in[i+1] != a {
				b = in[i+1]
				i += 2
			} else {
				if a == 'X' && i+1 < len(in) {
					b = 'Q'
				}
				i++
			}
			pairs = append(pairs, [2]byte{a, b})
		}
	}
	step := 1
	if dec {
		step = 4
	}
	var sb strings.Builder
	for _, p := range pairs {
		ra, ca := pos[p[0]]/5, pos[p[0]]%5
		rb, cb := pos[p[1]]/5, pos[p[1]]%5
		switch {
		case ra == rb:
			sb.WriteByte(sq[ra*5+(ca+step)%5])
			sb.WriteByte(sq[rb*5+(cb+step)%5])
		case ca == cb:
			sb.WriteByte(sq[((ra+step)%5)*5+ca])
			sb.WriteByte(sq[((rb+step)%5)*5+cb])
		default:
			sb.WriteByte(sq[ra*5+cb])
			sb.WriteByte(sq[rb*5+ca])
		}
	}
	return sb.String(), nil
}

// ---- Polybius ----

func polybius(text, key string, dec bool) (string, error) {
	sq := playfairSquare(key)
	if dec {
		var digits []int
		for _, r := range text {
			if r >= '1' && r <= '5' {
				digits = append(digits, int(r-'1'))
			}
		}
		if len(digits)%2 != 0 {
			return "", errors.New("polybius ciphertext needs an even number of digits 1-5")
		}
		var sb strings.Builder
		for i := 0; i < len(digits); i += 2 {
			sb.WriteByte(sq[digits[i]*5+digits[i+1]])
		}
		return sb.String(), nil
	}
	pos := map[byte]int{}
	for i, c := range sq {
		pos[c] = i
	}
	var parts []string
	for _, r := range strings.ToUpper(text) {
		if r == 'J' {
			r = 'I'
		}
		if isUpper(r) {
			p := pos[byte(r)]
			parts = append(parts, fmt.Sprintf("%d%d", p/5+1, p%5+1))
		}
	}
	return strings.Join(parts, " "), nil
}

// ---- Bacon ----

func baconCode(v int) string {
	b := make([]byte, 5)
	for i := 4; i >= 0; i-- {
		if v&1 == 1 {
			b[i] = 'B'
		} else {
			b[i] = 'A'
		}
		v >>= 1
	}
	return string(b)
}

func baconIndex(r rune, classic bool) int {
	v := int(r - 'A')
	if classic {
		if r == 'J' {
			v = int('I' - 'A')
		}
		if r == 'V' {
			v = int('U' - 'A')
		}
		// squeeze out J and V
		n := 0
		for c := 0; c < v; c++ {
			if rune('A'+c) != 'J' && rune('A'+c) != 'V' {
				n++
			}
		}
		return n
	}
	return v
}

func baconLetters(classic bool) []byte {
	var l []byte
	for c := 'A'; c <= 'Z'; c++ {
		if classic && (c == 'J' || c == 'V') {
			continue
		}
		l = append(l, byte(c))
	}
	return l
}

func bacon(text string, classic, dec bool) (string, error) {
	if dec {
		var bits []int
		for _, r := range strings.ToUpper(text) {
			switch r {
			case 'A', '0':
				bits = append(bits, 0)
			case 'B', '1':
				bits = append(bits, 1)
			}
		}
		if len(bits)%5 != 0 {
			return "", errors.New("bacon ciphertext needs groups of 5 A/B symbols")
		}
		l := baconLetters(classic)
		var sb strings.Builder
		for i := 0; i < len(bits); i += 5 {
			v := 0
			for _, b := range bits[i : i+5] {
				v = v<<1 | b
			}
			if v >= len(l) {
				return "", fmt.Errorf("invalid bacon group %s", baconCode(v))
			}
			sb.WriteByte(l[v])
		}
		return sb.String(), nil
	}
	var parts []string
	for _, r := range strings.ToUpper(text) {
		if isUpper(r) {
			parts = append(parts, baconCode(baconIndex(r, classic)))
		}
	}
	return strings.Join(parts, " "), nil
}

// ---- registration ----

func classicFn(name, title, doc, example string, params []tools.Param, fn func(a tools.Args) (string, error), order int, tags ...string) {
	reg(tools.Func{
		Name: name, Title: title, Doc: doc,
		Params:  params,
		Returns: "str", Input: "text",
		Fn:      func(a tools.Args) (any, error) { return fn(a) },
		Example: example,
		Tags:    append([]string{"classic cipher", "historical cipher", "puzzle", "ctf", "cryptogram", name}, tags...),
		Family:  "cipher", Order: order,
		Keywords: []string{name, "classic", "cipher"},
	})
}

func txt() tools.Param {
	return reqStr("text", "Text to process.")
}

func keyP(doc string) tools.Param { return reqStr("key", doc) }

func init() {
	const letterNote = " Letters keep their case; digits, spaces and punctuation pass through unchanged."
	classicFn("caesar", "Caesar cipher", "Shifts each letter forward by shift places (default 3, any integer; negative shifts go backwards, the shift wraps modulo 26)."+letterNote+" Use caesar_decrypt to undo or caesar_brute to try all shifts.",
		`cipher.caesar("Hello, World!", 3)`, []tools.Param{txt(), {Name: "shift", Type: "int", Default: int64(3), Doc: "Places to shift (default 3); negative shifts backwards."}},
		func(a tools.Args) (string, error) { return caesar(a.Str("text"), a.Int("shift")), nil }, 5, "caesar", "shift cipher", "rot", "julius caesar")
	classicFn("caesar_decrypt", "Caesar decrypt", "Undoes a Caesar cipher with a known shift (default 3) by shifting letters backwards."+letterNote,
		`cipher.caesar_decrypt("Khoor, Zruog!", 3)`, []tools.Param{txt(), {Name: "shift", Type: "int", Default: int64(3), Doc: "The shift that was used to encrypt (default 3)."}},
		func(a tools.Args) (string, error) { return caesar(a.Str("text"), -a.Int("shift")), nil }, 104, "caesar", "decrypt", "shift")
	reg(tools.Func{
		Name: "caesar_brute", Title: "Caesar brute force",
		Doc:     "Tries every Caesar shift from 1 to 25 on the text and returns 25 lines of the form 'N: candidate', where N is the forward shift applied to the text (so if the text was encrypted with shift S, the line with N = 26-S is the plaintext). Letters keep their case; other characters are unchanged.",
		Params:  []tools.Param{txt()},
		Returns: "list[str]", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			out := make([]string, 0, 25)
			for s := int64(1); s <= 25; s++ {
				out = append(out, fmt.Sprintf("%d: %s", s, caesar(a.Str("text"), s)))
			}
			return out, nil
		},
		Example: `cipher.caesar_brute("Khoor")`,
		Tags:    []string{"caesar", "brute force", "crack", "all shifts", "ctf", "break", "rot"},
		Family:  "cipher", Order: 105, Keywords: []string{"caesar", "brute"},
	})
	same := func(name, title, example string, f func(string) string, doc string, order int, tags ...string) {
		classicFn(name, title, doc+" It is its own inverse: applying it twice returns the original text.", example, []tools.Param{txt()},
			func(a tools.Args) (string, error) { return f(a.Str("text")), nil }, order, tags...)
	}
	same("rot13", "ROT13", `cipher.rot13("Hello")`, func(s string) string { return caesar(s, 13) },
		"Rotates every letter by 13 places (A to N)."+letterNote, 6, "rot13", "rot-13", "rotate 13", "usenet", "spoiler")
	same("rot5", "ROT5", `cipher.rot5("2024")`, rot5, "Rotates every digit by 5 places (0 to 5); letters and other characters are unchanged.", 106, "rot5", "digits")
	same("rot18", "ROT18", `cipher.rot18("Hello 2024")`, func(s string) string { return rot5(caesar(s, 13)) },
		"ROT13 on letters combined with ROT5 on digits.", 107, "rot18")
	same("rot47", "ROT47", `cipher.rot47("Hello")`, rot47,
		"Rotates every printable ASCII character from '!' to '~' by 47 places, so letters, digits and punctuation all change; space and non-ASCII are unchanged.", 108, "rot47", "ascii")
	same("atbash", "Atbash", `cipher.atbash("Hello")`, atbash, "Mirrors the alphabet (A to Z, B to Y, ...)."+letterNote, 109, "atbash", "mirror alphabet", "hebrew", "reverse alphabet")

	pair := func(name, title, what string, f func(text, key string, dec bool) (string, error), order int, tags ...string) {
		for i, dec := range []bool{false, true} {
			dec := dec
			suffix, verb := "_encrypt", "Encrypts"
			ex := fmt.Sprintf(`cipher.%s_encrypt("Attack at dawn", "LEMON")`, name)
			if dec {
				suffix, verb = "_decrypt", "Decrypts"
				ex = fmt.Sprintf(`cipher.%s_decrypt("Lxfopv ef rnhr", "LEMON")`, name)
			}
			classicFn(name+suffix, title+map[bool]string{false: " encrypt", true: " decrypt"}[dec],
				verb+" text with the "+title+": "+what+" The key must contain at least one letter; non-letters in the key are ignored.",
				ex, []tools.Param{txt(), keyP("Key word (letters).")},
				func(a tools.Args) (string, error) { return f(a.Str("text"), a.Str("key"), dec) }, order+i, tags...)
		}
	}
	pair("vigenere", "Vigenere cipher", "each letter is shifted by the matching letter of the repeating key (A=0). Letters keep their case, other characters pass through and do not advance the key.", vigenere, 110, "vigenere", "polyalphabetic", "keyword")
	pair("autokey", "Autokey cipher", "like Vigenere, but after the key the plaintext itself becomes the key stream. Letters keep their case; other characters pass through.", autokey, 112, "autokey", "polyalphabetic", "running key")
	classicFn("beaufort", "Beaufort cipher", "Beaufort cipher: each letter becomes (key letter - letter) mod 26 with a repeating key. It is its own inverse, so the same call decrypts. Letters keep their case; other characters pass through.",
		`cipher.beaufort("Attack at dawn", "LEMON")`, []tools.Param{txt(), keyP("Key word (letters).")},
		func(a tools.Args) (string, error) { return beaufort(a.Str("text"), a.Str("key")) }, 114, "beaufort", "polyalphabetic", "reciprocal")
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "affine_encrypt", "Affine encrypt", `cipher.affine_encrypt("AFFINE CIPHER", 5, 8)`
		if dec {
			name, title, ex = "affine_decrypt", "Affine decrypt", `cipher.affine_decrypt("IHHWVC SWFRCP", 5, 8)`
		}
		classicFn(name, title, "Affine cipher: maps letter x to (a*x+b) mod 26 (decrypt reverses it). a must be coprime to 26 (1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23, 25, or equivalents), otherwise it is an error."+letterNote,
			ex, []tools.Param{txt(), {Name: "a", Type: "int", Default: int64(5), Doc: "Multiplier, coprime to 26 (default 5)."}, {Name: "b", Type: "int", Default: int64(8), Doc: "Shift (default 8)."}},
			func(a tools.Args) (string, error) { return affine(a.Str("text"), a.Int("a"), a.Int("b"), dec) }, 116+i, "affine", "linear cipher")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "substitution_encrypt", "Substitution encrypt", `cipher.substitution_encrypt("Hello", "QWERTYUIOPASDFGHJKLZXCVBNM")`
		if dec {
			name, title, ex = "substitution_decrypt", "Substitution decrypt", `cipher.substitution_decrypt("Itssg", "QWERTYUIOPASDFGHJKLZXCVBNM")`
		}
		classicFn(name, title, "Monoalphabetic substitution: the 26-letter key is the cipher alphabet for A..Z (each letter exactly once; anything else is an error)."+letterNote,
			ex, []tools.Param{txt(), keyP("26 distinct letters, the cipher alphabet in order A..Z.")},
			func(a tools.Args) (string, error) { return substitution(a.Str("text"), a.Str("key"), dec) }, 118+i, "substitution", "monoalphabetic", "cryptogram", "keyed alphabet")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "rail_fence_encrypt", "Rail fence encrypt", `cipher.rail_fence_encrypt("WEAREDISCOVEREDFLEEATONCE", 3)`
		if dec {
			name, title, ex = "rail_fence_decrypt", "Rail fence decrypt", `cipher.rail_fence_decrypt("WECRLTEERDSOEEFEAOCAIVDEN", 3)`
		}
		classicFn(name, title, "Rail fence transposition: writes the text in a zigzag over the given number of rails (2 or more, up to 1000) and reads the rails in order. Every character, including spaces, takes part. Rails at or above the text length leave it unchanged.",
			ex, []tools.Param{txt(), {Name: "rails", Type: "int", Default: int64(3), Doc: "Number of rails, 2 to 1000 (default 3)."}},
			func(a tools.Args) (string, error) {
				if err := clampInt(a.Int("rails"), 2, 1000, "rails"); err != nil {
					return "", err
				}
				o := railOrder(len([]rune(a.Str("text"))), int(a.Int("rails")))
				if dec {
					return unpermute(a.Str("text"), o), nil
				}
				return permute(a.Str("text"), o), nil
			}, 120+i, "rail fence", "zigzag", "transposition")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "columnar_encrypt", "Columnar transposition encrypt", `cipher.columnar_encrypt("WEAREDISCOVERED", "ZEBRAS")`
		if dec {
			name, title, ex = "columnar_decrypt", "Columnar transposition decrypt", `cipher.columnar_decrypt("EEWDRAIDOSRCEVE", "ZEBRAS")`
		}
		classicFn(name, title, "Columnar transposition: writes the text in rows under the key word and reads the columns in alphabetical order of the key letters (ties left to right). Irregular last row, no padding; every character, including spaces, takes part. The key must not be empty.",
			ex, []tools.Param{txt(), keyP("Key word; its letters order the columns.")},
			func(a tools.Args) (string, error) {
				k := []rune(strings.ToUpper(a.Str("key")))
				if len(k) == 0 {
					return "", errors.New("key must not be empty")
				}
				o := columnOrder(len([]rune(a.Str("text"))), k)
				if dec {
					return unpermute(a.Str("text"), o), nil
				}
				return permute(a.Str("text"), o), nil
			}, 122+i, "columnar", "columnar transposition", "transposition", "route")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "scytale_encrypt", "Scytale encrypt", `cipher.scytale_encrypt("HELPMEIAMTRAPPED", 4)`
		if dec {
			name, title, ex = "scytale_decrypt", "Scytale decrypt", `cipher.scytale_decrypt("HMAEEPLITPAPMRED", 4)`
		}
		classicFn(name, title, "Scytale transposition: text wrapped around a rod of the given diameter (letters visible per turn); reading off the strip takes every diameter-th character in turn. Every character takes part; no padding.",
			ex, []tools.Param{txt(), {Name: "diameter", Type: "int", Default: int64(4), Doc: "Rod diameter, 1 to 1000 (default 4)."}},
			func(a tools.Args) (string, error) {
				if err := clampInt(a.Int("diameter"), 1, 1000, "diameter"); err != nil {
					return "", err
				}
				o := scytaleOrder(len([]rune(a.Str("text"))), int(a.Int("diameter")))
				if dec {
					return unpermute(a.Str("text"), o), nil
				}
				return permute(a.Str("text"), o), nil
			}, 124+i, "scytale", "skytale", "sparta", "transposition")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "playfair_encrypt", "Playfair encrypt", `cipher.playfair_encrypt("Hide the gold in the tree stump", "Playfair example")`
		if dec {
			name, title, ex = "playfair_decrypt", "Playfair decrypt", `cipher.playfair_decrypt("BMODZBXDNABEKUDMUIXMMOUVIF", "Playfair example")`
		}
		classicFn(name, title, "Playfair digraph cipher on a 5x5 key square (J merged into I). Only letters are used: output is uppercase without spaces or punctuation; encryption inserts X between doubled letters and pads an odd end, and decryption leaves those padding letters in place. Decrypt needs an even letter count.",
			ex, []tools.Param{txt(), keyP("Key word or phrase for the square.")},
			func(a tools.Args) (string, error) { return playfair(a.Str("text"), a.Str("key"), dec) }, 126+i, "playfair", "digraph", "wheatstone")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "polybius_encrypt", "Polybius square encrypt", `cipher.polybius_encrypt("Hello")`
		if dec {
			name, title, ex = "polybius_decrypt", "Polybius square decrypt", `cipher.polybius_decrypt("23 15 31 31 34")`
		}
		classicFn(name, title, "Polybius square on a 5x5 grid (J merged into I): each letter becomes row and column digits 1-5, groups separated by spaces. Only letters are encoded (other characters are dropped); decryption reads digits 1-5 in pairs and returns uppercase letters. An optional key word scrambles the square.",
			ex, []tools.Param{txt(), strParam("key", "Optional key word for the square (empty = plain alphabet).", "")},
			func(a tools.Args) (string, error) { return polybius(a.Str("text"), a.Str("key"), dec) }, 128+i, "polybius", "checkerboard", "tap code", "grid")
	}
	for i, dec := range []bool{false, true} {
		dec := dec
		name, title, ex := "bacon_encrypt", "Bacon cipher encrypt", `cipher.bacon_encrypt("Hello")`
		if dec {
			name, title, ex = "bacon_decrypt", "Bacon cipher decrypt", `cipher.bacon_decrypt("AABBB AABAA ABABB ABABB ABBAB")`
		}
		classicFn(name, title, "Bacon's cipher: each letter becomes five A/B symbols, groups separated by spaces (variant 'distinct' uses 26 codes, 'classic' merges I/J and U/V into 24). Only letters are encoded; decryption reads A/B (or 0/1) in groups of five and returns uppercase letters.",
			ex, []tools.Param{txt(), {Name: "variant", Type: "str", Default: "distinct", Choices: []string{"distinct", "classic"}, Doc: "distinct (26 letters, default) or classic (24 letters)."}},
			func(a tools.Args) (string, error) {
				v := strings.ToLower(a.Str("variant"))
				if v != "distinct" && v != "classic" {
					return "", errors.New("variant must be distinct or classic")
				}
				return bacon(a.Str("text"), v == "classic", dec)
			}, 130+i, "bacon", "baconian", "binary letters", "steganography")
	}
	reg(tools.Func{
		Name: "classic_all", Title: "Classic ciphers overview",
		Doc:     "Applies the parameter-free classic ciphers to the text at once and returns a dict: rot13, rot5, rot18, rot47, atbash, caesar_1, caesar_2, caesar_3 (shifts forward), reverse (characters reversed). Handy for quickly eyeballing a CTF string; use caesar_brute for all shifts.",
		Params:  []tools.Param{txt()},
		Returns: "dict[str, str]", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			t := a.Str("text")
			r := []rune(t)
			for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
				r[i], r[j] = r[j], r[i]
			}
			return map[string]string{
				"rot13": caesar(t, 13), "rot5": rot5(t), "rot18": rot5(caesar(t, 13)), "rot47": rot47(t), "atbash": atbash(t),
				"caesar_1": caesar(t, 1), "caesar_2": caesar(t, 2), "caesar_3": caesar(t, 3), "reverse": string(r),
			}, nil
		},
		Example: `cipher.classic_all("Hello, World!")`,
		Tags:    []string{"classic", "rot13", "atbash", "caesar", "rot47", "all", "overview", "ctf", "decode"},
		Family:  "cipher", Order: 7, Keywords: []string{"classic", "rot13", "atbash", "caesar"},
	})
}
