package encode

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/text/unicode/norm"
)

// ---- morse ----

var morseTable = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.", 'G': "--.", 'H': "....", 'I': "..", 'J': ".---",
	'K': "-.-", 'L': ".-..", 'M': "--", 'N': "-.", 'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.", 'S': "...", 'T': "-",
	'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-", 'Y': "-.--", 'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-", '5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.", '!': "-.-.--", '/': "-..-.", '(': "-.--.", ')': "-.--.-", '&': ".-...",
	':': "---...", ';': "-.-.-.", '=': "-...-", '+': ".-.-.", '-': "-....-", '_': "..--.-", '"': ".-..-.", '$': "...-..-", '@': ".--.-.",
	'Ä': ".-.-", 'Ö': "---.", 'Ü': "..--", 'Ñ': "--.--", 'É': "..-..", 'Ç': "-.-..", 'Å': ".--.-",
}

var morseRev = func() map[string]rune {
	m := map[string]rune{}
	for r, c := range morseTable {
		m[c] = r
	}
	return m
}()

func morseFold(r rune) (string, bool) {
	r = unicode.ToUpper(r)
	if c, ok := morseTable[r]; ok {
		return c, true
	}
	d := norm.NFD.String(string(r))
	if rs := []rune(d); len(rs) > 0 {
		if c, ok := morseTable[unicode.ToUpper(rs[0])]; ok {
			return c, true
		}
	}
	if r == 'ß' || r == 'ẞ' {
		return morseTable['S'] + " " + morseTable['S'], true
	}
	return "", false
}

func morseEncode(s, dot, dash, sep, wordSep string) string {
	var words []string
	for _, w := range strings.Fields(s) {
		var letters []string
		for _, r := range w {
			if c, ok := morseFold(r); ok {
				c = strings.NewReplacer(".", "\x01", "-", "\x02").Replace(c)
				c = strings.NewReplacer("\x01", dot, "\x02", dash).Replace(c)
				letters = append(letters, c)
			}
		}
		if len(letters) > 0 {
			words = append(words, strings.Join(letters, sep))
		}
	}
	return strings.Join(words, wordSep)
}

func morseNormalize(s string) string {
	return strings.NewReplacer("•", ".", "·", ".", "∙", ".", "–", "-", "—", "-", "−", "-", "\u2012", "-", "\r", "\n").Replace(s)
}

func morseDecode(s string) (string, error) {
	s = morseNormalize(s)
	for strings.Contains(s, "   ") {
		s = strings.ReplaceAll(s, "   ", " / ")
	}
	s = strings.ReplaceAll(s, "\n", " / ")
	s = strings.ReplaceAll(s, "|", "/")
	var sb strings.Builder
	any := false
	for wi, w := range strings.Split(s, "/") {
		toks := strings.Fields(w)
		if len(toks) == 0 {
			continue
		}
		if wi > 0 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		for _, t := range toks {
			if r, ok := morseRev[t]; ok {
				sb.WriteRune(r)
				any = true
			} else if strings.Trim(t, ".-") == "" {
				sb.WriteByte('?')
			} else {
				return "", fmt.Errorf("invalid morse token %q", t)
			}
		}
	}
	if !any {
		return "", fmt.Errorf("no morse code found")
	}
	return sb.String(), nil
}

// ---- NATO ----

var natoWords = []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo", "Foxtrot", "Golf", "Hotel", "India", "Juliet", "Kilo", "Lima", "Mike", "November", "Oscar", "Papa", "Quebec", "Romeo", "Sierra", "Tango", "Uniform", "Victor", "Whiskey", "X-ray", "Yankee", "Zulu"}
var natoDigits = []string{"Zero", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine"}

var natoRev = func() map[string]string {
	m := map[string]string{}
	for i, w := range natoWords {
		m[strings.ToLower(w)] = string(rune('A' + i))
	}
	m["alfa"], m["juliett"], m["xray"], m["x_ray"] = "A", "J", "X", "X"
	for i, w := range natoDigits {
		m[strings.ToLower(w)] = strconv.Itoa(i)
	}
	m["niner"], m["fife"], m["fower"], m["tree"], m["nought"], m["zéro"] = "9", "5", "4", "3", "0", "0"
	return m
}()

func natoEncode(s, sep string, digits bool) string {
	var parts []string
	for _, r := range s {
		u := unicode.ToUpper(r)
		switch {
		case u >= 'A' && u <= 'Z':
			parts = append(parts, natoWords[u-'A'])
		case r >= '0' && r <= '9' && digits:
			parts = append(parts, natoDigits[r-'0'])
		case unicode.IsSpace(r):
			parts = append(parts, "/")
		default:
			parts = append(parts, string(r))
		}
	}
	return strings.Join(parts, sep)
}

func natoTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '\r' })
}

func natoDecode(s string, lower bool) (string, error) {
	var sb strings.Builder
	n := 0
	for _, t := range natoTokens(s) {
		if t == "/" {
			sb.WriteByte(' ')
			continue
		}
		v, ok := natoRev[strings.ToLower(t)]
		if !ok {
			if utf8.RuneCountInString(t) == 1 && !unicode.IsLetter([]rune(t)[0]) && !unicode.IsDigit([]rune(t)[0]) {
				sb.WriteString(t)
				continue
			}
			return "", fmt.Errorf("%q is not a NATO phonetic word", t)
		}
		sb.WriteString(v)
		n++
	}
	if n == 0 {
		return "", fmt.Errorf("no NATO phonetic words found")
	}
	r := sb.String()
	if lower {
		r = strings.ToLower(r)
	}
	return r, nil
}

// ---- braille ----

const brailleASCII = " A1B'K2L@CIF/MSP\"E3H9O6R^DJG>NTQ,*5<-U8V.%[$+X!&;:4\\0Z7(_?W]#Y)="

var brailleLit = map[rune]rune{',': '⠂', ';': '⠆', ':': '⠒', '.': '⠲', '!': '⠖', '?': '⠦', '\'': '⠄', '-': '⠤', '(': '⠶', ')': '⠶', '"': '⠶'}

func brailleCell(c byte) rune {
	i := strings.IndexByte(brailleASCII, c)
	if i < 0 {
		return -1
	}
	return rune(0x2800 + i)
}

func brailleEncode(s, variant string, capitals bool) (string, error) {
	var sb strings.Builder
	switch normName(variant) {
	case "ascii", "brf":
		for _, r := range strings.ToUpper(s) {
			if r < 0x80 {
				if c := brailleCell(byte(r)); c >= 0 {
					sb.WriteRune(c)
					continue
				}
			}
			sb.WriteRune(r)
		}
		return sb.String(), nil
	case "", "literary", "grade1", "grade-1":
	default:
		return "", fmt.Errorf("variant must be literary or ascii")
	}
	inNum := false
	for _, r := range s {
		u := unicode.ToUpper(r)
		switch {
		case u >= 'A' && u <= 'Z':
			inNum = false
			if r != unicode.ToLower(r) && capitals {
				sb.WriteRune('⠠')
			}
			sb.WriteRune(brailleCell(byte(u)))
		case r >= '0' && r <= '9':
			if !inNum {
				sb.WriteRune('⠼')
				inNum = true
			}
			d := byte('J')
			if r != '0' {
				d = byte('A' + r - '1')
			}
			sb.WriteRune(brailleCell(d))
		case r == ' ' || r == '\t':
			inNum = false
			sb.WriteRune('⠀')
		case r == '\n':
			inNum = false
			sb.WriteRune('\n')
		default:
			if c, ok := brailleLit[r]; ok {
				sb.WriteRune(c)
				inNum = false
			} else {
				sb.WriteRune(r)
			}
		}
	}
	return sb.String(), nil
}

func brailleDecode(s, variant string) (string, error) {
	var sb strings.Builder
	rev := map[rune]rune{}
	for _, c := range []struct {
		r rune
		o rune
	}{{'⠂', ','}, {'⠆', ';'}, {'⠒', ':'}, {'⠲', '.'}, {'⠖', '!'}, {'⠦', '?'}, {'⠄', '\''}, {'⠤', '-'}, {'⠶', '"'}} {
		rev[c.r] = c.o
	}
	ascii := normName(variant) == "ascii" || normName(variant) == "brf"
	num, cap, capWord := false, false, false
	n := 0
	for _, r := range s {
		if r < 0x2800 || r > 0x28FF {
			if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
				num, capWord = false, false
				sb.WriteRune(r)
				continue
			}
			sb.WriteRune(r)
			continue
		}
		n++
		if r == 0x2800 {
			num, capWord = false, false
			sb.WriteByte(' ')
			continue
		}
		idx := int(r - 0x2800)
		if ascii {
			sb.WriteByte(brailleASCII[idx])
			continue
		}
		switch r {
		case '⠼':
			num = true
			continue
		case '⠠':
			if cap {
				capWord = true
			}
			cap = true
			continue
		}
		ch := brailleASCII[idx]
		if num && ch >= 'A' && ch <= 'J' {
			d := ch - 'A' + '1'
			if ch == 'J' {
				d = '0'
			}
			sb.WriteByte(d)
			continue
		}
		num = false
		if ch >= 'A' && ch <= 'Z' {
			if cap || capWord {
				sb.WriteByte(ch)
				cap = false
			} else {
				sb.WriteByte(ch + 32)
			}
			continue
		}
		cap = false
		if p, ok := rev[r]; ok {
			sb.WriteRune(p)
		} else {
			sb.WriteByte('?')
		}
	}
	if n == 0 {
		return "", fmt.Errorf("no braille characters found")
	}
	return sb.String(), nil
}

// ---- tap code ----

const tapSquare = "ABCDEFGHIJLMNOPQRSTUVWXYZ"

func tapEncode(s, style string) (string, error) {
	dots := normName(style) != "numbers"
	if st := normName(style); st != "" && st != "dots" && st != "numbers" {
		return "", fmt.Errorf("style must be dots or numbers")
	}
	var words []string
	for _, w := range strings.Fields(s) {
		var letters []string
		for _, r := range strings.ToUpper(w) {
			if r == 'K' {
				r = 'C'
			}
			i := strings.IndexRune(tapSquare, r)
			if i < 0 {
				continue
			}
			row, col := i/5+1, i%5+1
			if dots {
				letters = append(letters, strings.Repeat(".", row)+" "+strings.Repeat(".", col))
			} else {
				letters = append(letters, fmt.Sprintf("%d%d", row, col))
			}
		}
		if len(letters) > 0 {
			lsep := "  "
			if !dots {
				lsep = " "
			}
			words = append(words, strings.Join(letters, lsep))
		}
	}
	sep := " / "
	return strings.Join(words, sep), nil
}

func tapDecode(s string) (string, error) {
	var sb strings.Builder
	n := 0
	for wi, w := range strings.Split(s, "/") {
		toks := strings.Fields(w)
		if len(toks) == 0 {
			continue
		}
		if wi > 0 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		var nums []int
		for _, t := range toks {
			switch {
			case strings.Trim(t, ".") == "" && len(t) >= 1 && len(t) <= 5:
				nums = append(nums, len(t))
			case len(t) == 2 && t[0] >= '1' && t[0] <= '5' && t[1] >= '1' && t[1] <= '5':
				nums = append(nums, int(t[0]-'0'), int(t[1]-'0'))
			default:
				return "", fmt.Errorf("invalid tap code token %q", t)
			}
		}
		if len(nums)%2 != 0 {
			return "", fmt.Errorf("tap code needs row/column pairs")
		}
		for i := 0; i < len(nums); i += 2 {
			sb.WriteByte(tapSquare[(nums[i]-1)*5+nums[i+1]-1])
			n++
		}
	}
	if n == 0 {
		return "", fmt.Errorf("no tap code found")
	}
	return sb.String(), nil
}

// ---- A1Z26 ----

func a1z26Encode(s, sep string) string {
	var words []string
	for _, w := range strings.Fields(s) {
		var nums []string
		for _, r := range strings.ToUpper(w) {
			if r >= 'A' && r <= 'Z' {
				nums = append(nums, strconv.Itoa(int(r-'A')+1))
			}
		}
		if len(nums) > 0 {
			words = append(words, strings.Join(nums, sep))
		}
	}
	return strings.Join(words, " ")
}

func a1z26Decode(s string) (string, error) {
	hasLink := strings.ContainsAny(s, "-,")
	s = strings.ReplaceAll(s, "\n", " / ")
	var sb strings.Builder
	n := 0
	for wi, w := range strings.Split(s, "/") {
		var toks []string
		if hasLink {
			// spaces separate words, - and , separate letters
			for _, ws := range strings.Fields(w) {
				toks = append(toks, "\x00")
				toks = append(toks, strings.FieldsFunc(ws, func(r rune) bool { return r == '-' || r == ',' })...)
			}
			if len(toks) > 0 {
				toks = toks[1:]
			}
		} else {
			toks = strings.Fields(strings.ReplaceAll(w, ",", " "))
		}
		if len(toks) == 0 {
			continue
		}
		if wi > 0 && sb.Len() > 0 {
			sb.WriteByte(' ')
		}
		for _, t := range toks {
			if t == "\x00" {
				sb.WriteByte(' ')
				continue
			}
			v, err := strconv.Atoi(t)
			if err != nil || v < 1 || v > 26 {
				return "", fmt.Errorf("%q is not a number from 1 to 26", t)
			}
			sb.WriteByte(byte('A' + v - 1))
			n++
		}
	}
	if n == 0 {
		return "", fmt.Errorf("no numbers found")
	}
	return sb.String(), nil
}

// ---- ASCII table ----

var ctrlNames = []string{"NUL", "SOH", "STX", "ETX", "EOT", "ENQ", "ACK", "BEL", "BS", "HT", "LF", "VT", "FF", "CR", "SO", "SI", "DLE", "DC1", "DC2", "DC3", "DC4", "NAK", "SYN", "ETB", "CAN", "EM", "SUB", "ESC", "FS", "GS", "RS", "US"}

func asciiLookup(s string) (map[string]string, error) {
	if s == "" {
		return nil, fmt.Errorf("char is empty")
	}
	if utf8.RuneCountInString(s) != 1 {
		return nil, fmt.Errorf("expected a single character, got %d", utf8.RuneCountInString(s))
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return nil, fmt.Errorf("invalid UTF-8 character")
	}
	m := map[string]string{
		"char": string(r), "codepoint": fmt.Sprintf("U+%04X", r), "decimal": strconv.Itoa(int(r)), "hex": fmt.Sprintf("%02X", r),
		"octal": fmt.Sprintf("%03o", r), "binary": fmt.Sprintf("%08b", r), "utf8_hex": hexJoin([]byte(string(r)), true, " ", ""),
	}
	if r > 0xFFFF {
		m["binary"] = fmt.Sprintf("%b", r)
	}
	switch {
	case r < 32:
		m["name"] = ctrlNames[r]
		m["class"] = "control"
		m["char"] = "^" + string(rune('@'+r))
	case r == 32:
		m["name"], m["class"] = "SP", "space"
	case r == 127:
		m["name"], m["class"], m["char"] = "DEL", "control", "^?"
	case r < 127:
		m["name"] = string(r)
		switch {
		case unicode.IsLetter(r):
			m["class"] = "letter"
		case unicode.IsDigit(r):
			m["class"] = "digit"
		default:
			m["class"] = "punctuation"
		}
	default:
		m["class"] = "non-ascii"
		if n, ok := entByRune[r]; ok {
			m["html_entity"] = "&" + n + ";"
		}
	}
	if r < 0x80 {
		m["ascii"] = "true"
	} else {
		m["ascii"] = "false"
	}
	return m, nil
}

func init() {
	reg(tools.Func{
		Name: "morse_encode", Title: "Morse code encode",
		Doc:     "Encodes text as ITU Morse code: letters separated by a space, words by \" / \". Letters, digits and common punctuation are supported; accented letters are folded to their base letter (é becomes e) except ä ö ü ñ é ç å which have Morse forms; other characters are dropped. dot/dash/sep/word_sep customize the symbols.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("dot", "Symbol for a dot.", "."), pStr("dash", "Symbol for a dash.", "-"), pStr("sep", "Between letters.", " "), pStr("word_sep", "Between words.", " / ")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return morseEncode(a.Str("text"), a.Str("dot"), a.Str("dash"), a.Str("sep"), a.Str("word_sep")), nil
		},
		Example:  `encode.morse_encode("SOS hello")`,
		Tags:     tagsWith("morse", "itu", "dots and dashes", "telegraph", "sos", "cw", "encode"),
		Keywords: []string{"morse"},
	})
	reg(tools.Func{
		Name: "morse_decode", Title: "Morse code decode",
		Doc:     "Decodes Morse code made of . and - (also • · – —): letters separated by spaces, words by / or | or three or more spaces or a line break. Unknown dot/dash groups of valid characters decode to \"?\"; other characters are an error. Result is upper case.",
		Params:  []tools.Param{pReq("text", "Morse code, e.g. \"... --- ...\".")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return morseDecode(a.Str("text")) },
		Accepts: func(s string) bool {
			t := morseNormalize(s)
			if strings.Trim(t, ".-/| \n\t") != "" || !strings.ContainsAny(t, ".-") {
				return false
			}
			_, err := morseDecode(s)
			return err == nil
		},
		Example:  `encode.morse_decode("... --- ... / .... .. ")`,
		Tags:     tagsWith("morse", "itu", "dots and dashes", "telegraph", "sos", "cw", "decode"),
		Keywords: []string{"morse"},
	})
	reg(tools.Func{
		Name: "nato_phonetic_encode", Title: "NATO phonetic alphabet",
		Doc:     "Spells text with the NATO/ICAO phonetic alphabet (Alpha Bravo Charlie ...). Digits become words (Zero, One, ...) unless digits=false, spaces become \"/\", and other characters are kept as they are.",
		Params:  []tools.Param{pReq("text", "Text to spell."), pStr("separator", "Between words.", " "), pBool("digits", "Spell digits as words.", true)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return natoEncode(a.Str("text"), a.Str("separator"), a.Bool("digits")), nil
		},
		Example: `encode.nato_phonetic_encode("Go 42")`,
		Tags:    tagsWith("nato", "icao", "phonetic alphabet", "alpha bravo charlie", "spelling alphabet", "radio"),

		Keywords: []string{"nato", "phonetic"},
	})
	reg(tools.Func{
		Name: "nato_decode", Title: "NATO phonetic decode",
		Doc:     "Turns NATO phonetic words (Alpha, Bravo, ... also Alfa, Juliett, Niner, Fife, digit words) back into letters and digits, upper case unless lower=true. \"/\" becomes a space and single punctuation marks are kept; any other word is an error.",
		Params:  []tools.Param{pReq("text", "Phonetic words."), pBool("lower", "Return lower case.", false)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return natoDecode(a.Str("text"), a.Bool("lower")) },
		Accepts: func(s string) bool { _, err := natoDecode(s, false); return err == nil && len(natoTokens(s)) >= 2 },
		Example: `encode.nato_decode("Golf Oscar / Four Two")`,
		Tags:    tagsWith("nato", "icao", "phonetic alphabet", "alpha bravo charlie", "spelling alphabet", "radio", "decode"),

		Keywords: []string{"nato", "phonetic"},
	})
	reg(tools.Func{
		Name: "braille_encode", Title: "Braille encode",
		Doc:     "Converts text to Unicode Braille patterns. variant \"literary\" (default) is grade 1 English braille: capital sign before capitals (capitals=false drops it), number sign before digit runs, basic punctuation. variant \"ascii\" maps each ASCII character to its Braille ASCII (BRF) cell.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("variant", "Braille flavor.", "literary", "literary", "ascii"), pBool("capitals", "Emit the capital sign for upper-case letters.", true)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return brailleEncode(a.Str("text"), a.Str("variant"), a.Bool("capitals"))
		},
		Example: `encode.braille_encode("Hello 42")`,
		Tags:    tagsWith("braille", "grade 1", "brf", "tactile", "dots", "accessibility", "unicode braille"),

		Keywords: []string{"braille"},
	})
	reg(tools.Func{
		Name: "braille_decode", Title: "Braille decode",
		Doc:     "Converts Unicode Braille patterns (U+2800-U+28FF) back to text using grade 1 English rules (capital sign, number sign, basic punctuation) or, with variant \"ascii\", Braille ASCII. Non-braille characters pass through; text with no braille cells is an error.",
		Params:  []tools.Param{pReq("text", "Braille text."), pStr("variant", "Braille flavor.", "literary", "literary", "ascii")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return brailleDecode(a.Str("text"), a.Str("variant")) },
		Accepts: func(s string) bool {
			for _, r := range s {
				if r >= 0x2800 && r <= 0x28FF && r != 0x2800 {
					return true
				}
			}
			return false
		},
		Example: `encode.braille_decode("⠠⠓⠑⠇⠇⠕")`,
		Tags:    tagsWith("braille", "grade 1", "brf", "tactile", "dots", "accessibility", "unicode braille", "decode"),

		Keywords: []string{"braille"},
	})
	reg(tools.Func{
		Name: "tap_code_encode", Title: "Tap code encode",
		Doc:     "Encodes letters with the 5x5 tap code (Polybius square without K; K is written as C). style \"dots\" gives \".. ...\" (row dots, column dots), style \"numbers\" gives \"23\"; letters are separated by two spaces and words by \" / \". Non-letters are dropped.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("style", "Output style.", "dots", "dots", "numbers")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return tapEncode(a.Str("text"), a.Str("style")) },
		Example: `encode.tap_code_encode("hello")`,
		Tags:    tagsWith("tap code", "polybius", "prison code", "knock code", "5x5", "encode"),

		Keywords: []string{"tap", "tapcode"},
	})
	reg(tools.Func{
		Name: "tap_code_decode", Title: "Tap code decode",
		Doc:     "Decodes tap code written as dot groups (\".. ...\", row then column) or two-digit numbers (\"23\"); words are separated by /. Returns upper-case letters (C also stands for K). Fails on malformed groups.",
		Params:  []tools.Param{pReq("text", "Tap code.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return tapDecode(a.Str("text")) },
		Accepts: func(s string) bool {
			if strings.Trim(s, ". /12345\n") != "" {
				return false
			}
			_, err := tapDecode(s)
			return err == nil
		},
		Example: `encode.tap_code_decode("23 15 31 31 34")`,
		Tags:    tagsWith("tap code", "polybius", "prison code", "knock code", "5x5", "decode"),

		Keywords: []string{"tap", "tapcode"},
	})
	reg(tools.Func{
		Name: "a1z26_encode", Title: "A1Z26 encode (letters to numbers)",
		Doc:     "Replaces each letter with its alphabet position (A=1 ... Z=26), joined by separator (\"-\" default) within a word; words are separated by a space. Case-insensitive; characters that are not ASCII letters are dropped.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("separator", "Between letters of a word.", "-")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return a1z26Encode(a.Str("text"), a.Str("separator")), nil },
		Example: `encode.a1z26_encode("hello world")`,
		Tags:    tagsWith("a1z26", "letters to numbers", "alphabet position", "simple cipher", "numeric", "encode"),

		Keywords: []string{"a1z26", "letters"},
	})
	reg(tools.Func{
		Name: "a1z26_decode", Title: "A1Z26 decode (numbers to letters)",
		Doc:     "Turns numbers 1-26 back into upper-case letters. With \"-\" or \",\" in the input, those join letters and spaces separate words; otherwise spaces separate letters. \"/\" or a line break is always a word break. Numbers outside 1-26 are an error.",
		Params:  []tools.Param{pReq("text", "Numbers such as \"8-5-12-12-15 23-15-18-12-4\".")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return a1z26Decode(a.Str("text")) },
		Accepts: func(s string) bool {
			if strings.Trim(s, "0123456789-, /\n") != "" {
				return false
			}
			_, err := a1z26Decode(s)
			return err == nil && len(strings.FieldsFunc(s, func(r rune) bool { return r < '0' || r > '9' })) >= 2
		},
		Example: `encode.a1z26_decode("8-5-12-12-15")`,
		Tags:    tagsWith("a1z26", "numbers to letters", "alphabet position", "simple cipher", "numeric", "decode"),

		Keywords: []string{"a1z26", "letters"},
	})
	regAPI(tools.Func{
		Name: "ascii_table_lookup", Title: "ASCII table lookup",
		Doc:     "Looks up one character and returns a dict with char, codepoint (U+XXXX), decimal, hex, octal, binary, utf8_hex, ascii (\"true\"/\"false\"), class and, for control codes, name (NUL, LF, ESC ...); html_entity for known non-ASCII characters. More than one character is an error.",
		Params:  []tools.Param{pReq("char", "A single character, e.g. \"A\" or \"\\n\".")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return asciiLookup(a.Str("char")) },
		Example: `encode.ascii_table_lookup("A")`,
		Tags:    tagsWith("ascii", "ascii table", "char code", "codepoint", "control characters", "lookup", "ord", "unicode"),
	})
}
