package encode

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

type charsetEntry struct {
	name    string
	enc     encoding.Encoding // nil for the special cases handled in code
	aliases []string
}

var charsetList []charsetEntry
var charsetIndex = map[string]*charsetEntry{}

func addCS(name string, enc encoding.Encoding, aliases ...string) {
	charsetList = append(charsetList, charsetEntry{name, enc, aliases})
}

func init() {
	addCS("utf-8", nil, "utf8", "utf_8", "unicode-1-1-utf-8")
	addCS("utf-8-sig", nil, "utf8-sig", "utf-8-bom", "utf8bom")
	addCS("ascii", nil, "us-ascii", "usascii", "ansi_x3.4-1968", "iso646-us")
	addCS("utf-16le", unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "utf16le", "utf-16-le", "ucs-2le", "ucs2le")
	addCS("utf-16be", unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "utf16be", "utf-16-be", "ucs-2be", "ucs2be")
	addCS("utf-16", unicode.UTF16(unicode.LittleEndian, unicode.UseBOM), "utf16", "ucs-2", "ucs2", "unicode")
	addCS("utf-32le", utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM), "utf32le", "utf-32-le", "ucs-4le")
	addCS("utf-32be", utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM), "utf32be", "utf-32-be", "ucs-4be")
	addCS("utf-32", utf32.UTF32(utf32.LittleEndian, utf32.UseBOM), "utf32", "ucs-4", "ucs4")
	isos := []*charmap.Charmap{charmap.ISO8859_1, charmap.ISO8859_2, charmap.ISO8859_3, charmap.ISO8859_4, charmap.ISO8859_5, charmap.ISO8859_6, charmap.ISO8859_7, charmap.ISO8859_8, charmap.ISO8859_9, charmap.ISO8859_10}
	for i, c := range isos {
		n := i + 1
		al := []string{fmt.Sprintf("iso8859-%d", n), fmt.Sprintf("iso_8859-%d", n)}
		switch n {
		case 1:
			al = append(al, "latin1", "latin-1", "l1", "iso-ir-100", "cp819")
		case 2, 3, 4:
			al = append(al, fmt.Sprintf("latin%d", n), fmt.Sprintf("l%d", n))
		case 5:
			al = append(al, "cyrillic")
		case 6:
			al = append(al, "arabic")
		case 7:
			al = append(al, "greek")
		case 8:
			al = append(al, "hebrew")
		case 9:
			al = append(al, "latin5", "l5")
		case 10:
			al = append(al, "latin6", "l6")
		}
		addCS(fmt.Sprintf("iso-8859-%d", n), c, al...)
	}
	for _, p := range []struct {
		n int
		c *charmap.Charmap
	}{{13, charmap.ISO8859_13}, {14, charmap.ISO8859_14}, {15, charmap.ISO8859_15}, {16, charmap.ISO8859_16}} {
		al := []string{fmt.Sprintf("iso8859-%d", p.n), fmt.Sprintf("iso_8859-%d", p.n)}
		switch p.n {
		case 13:
			al = append(al, "latin7", "l7")
		case 14:
			al = append(al, "latin8", "l8")
		case 15:
			al = append(al, "latin9", "l9")
		case 16:
			al = append(al, "latin10", "l10")
		}
		addCS(fmt.Sprintf("iso-8859-%d", p.n), p.c, al...)
	}
	wins := map[int]*charmap.Charmap{1250: charmap.Windows1250, 1251: charmap.Windows1251, 1252: charmap.Windows1252, 1253: charmap.Windows1253, 1254: charmap.Windows1254, 1255: charmap.Windows1255, 1256: charmap.Windows1256, 1257: charmap.Windows1257, 1258: charmap.Windows1258}
	for n := 1250; n <= 1258; n++ {
		addCS(fmt.Sprintf("windows-%d", n), wins[n], fmt.Sprintf("cp%d", n), fmt.Sprintf("win%d", n), fmt.Sprintf("windows%d", n), fmt.Sprintf("x-cp%d", n))
	}
	addCS("windows-874", charmap.Windows874, "cp874", "win874", "iso-8859-11", "tis-620", "thai")
	addCS("koi8-r", charmap.KOI8R, "koi8r", "koi8_r")
	addCS("koi8-u", charmap.KOI8U, "koi8u", "koi8_u")
	addCS("ibm437", charmap.CodePage437, "cp437", "437", "dos-us")
	addCS("ibm850", charmap.CodePage850, "cp850", "850")
	addCS("ibm852", charmap.CodePage852, "cp852", "852")
	addCS("ibm855", charmap.CodePage855, "cp855", "855")
	addCS("ibm858", charmap.CodePage858, "cp858", "858")
	addCS("ibm860", charmap.CodePage860, "cp860", "860")
	addCS("ibm862", charmap.CodePage862, "cp862", "862")
	addCS("ibm863", charmap.CodePage863, "cp863", "863")
	addCS("ibm865", charmap.CodePage865, "cp865", "865")
	addCS("ibm866", charmap.CodePage866, "cp866", "866")
	addCS("ibm037", charmap.CodePage037, "cp037", "ebcdic-cp-us", "ebcdic")
	addCS("macintosh", charmap.Macintosh, "mac", "macroman", "mac-roman", "x-mac-roman")
	addCS("x-mac-cyrillic", charmap.MacintoshCyrillic, "mac-cyrillic", "maccyrillic")
	addCS("shift_jis", japanese.ShiftJIS, "shift-jis", "sjis", "shiftjis", "ms_kanji", "cp932", "windows-31j", "x-sjis")
	addCS("euc-jp", japanese.EUCJP, "eucjp", "ujis")
	addCS("iso-2022-jp", japanese.ISO2022JP, "iso2022jp", "jis", "csiso2022jp")
	addCS("gbk", simplifiedchinese.GBK, "cp936", "gb2312", "euc-cn", "euccn", "x-gbk", "chinese")
	addCS("gb18030", simplifiedchinese.GB18030)
	addCS("hz-gb-2312", simplifiedchinese.HZGB2312, "hz", "hzgb2312", "hz-gb2312")
	addCS("big5", traditionalchinese.Big5, "big-5", "big5-hkscs", "cp950", "x-x-big5")
	addCS("euc-kr", korean.EUCKR, "euckr", "cp949", "uhc", "ks_c_5601-1987", "korean")

	for i := range charsetList {
		e := &charsetList[i]
		charsetIndex[e.name] = e
		for _, a := range e.aliases {
			if _, ok := charsetIndex[a]; !ok {
				charsetIndex[a] = e
			}
		}
	}
}

func charsetNames() []string {
	out := make([]string, len(charsetList))
	for i, e := range charsetList {
		out[i] = e.name
	}
	sort.Strings(out)
	return out
}

func findCS(name string) (*charsetEntry, encoding.Encoding, error) {
	n := normName(name)
	if n == "" {
		n = "utf-8"
	}
	if e, ok := charsetIndex[n]; ok {
		return e, e.enc, nil
	}
	if e, ok := charsetIndex[strings.ReplaceAll(n, "_", "-")]; ok {
		return e, e.enc, nil
	}
	if enc, err := htmlindex.Get(n); err == nil {
		return &charsetEntry{name: n}, enc, nil
	}
	return nil, nil, fmt.Errorf("unknown charset %q; see encode.charsets() for the supported names", name)
}

// csEncode converts text to bytes in the named charset.
func csEncode(text, name string, replace bool) ([]byte, error) {
	e, enc, err := findCS(name)
	if err != nil {
		return nil, err
	}
	switch e.name {
	case "utf-8":
		if !utf8.ValidString(text) {
			text = strings.ToValidUTF8(text, "�")
		}
		return []byte(text), nil
	case "utf-8-sig":
		return append([]byte{0xEF, 0xBB, 0xBF}, text...), nil
	case "ascii":
		out := make([]byte, 0, len(text))
		for _, r := range text {
			if r > 0x7f {
				if !replace {
					return nil, fmt.Errorf("character %q (U+%04X) cannot be encoded in ascii", r, r)
				}
				r = '?'
			}
			out = append(out, byte(r))
		}
		return out, nil
	}
	en := enc.NewEncoder()
	if replace {
		en = encoding.ReplaceUnsupported(en)
	}
	b, err := en.Bytes([]byte(text))
	if err != nil {
		return nil, fmt.Errorf("cannot encode the text in %s: %v (use errors=\"replace\" to substitute)", e.name, err)
	}
	return b, nil
}

// csDecode converts bytes in the named charset to text.
func csDecode(b []byte, name string, replace bool) (string, error) {
	e, enc, err := findCS(name)
	if err != nil {
		return "", err
	}
	switch e.name {
	case "utf-8":
		if !utf8.Valid(b) {
			if !replace {
				return "", fmt.Errorf("bytes are not valid UTF-8 (use errors=\"replace\" or another charset)")
			}
			return strings.ToValidUTF8(string(b), "�"), nil
		}
		return string(b), nil
	case "utf-8-sig":
		b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
		if !utf8.Valid(b) {
			if !replace {
				return "", fmt.Errorf("bytes are not valid UTF-8 (use errors=\"replace\" or another charset)")
			}
			return strings.ToValidUTF8(string(b), "�"), nil
		}
		return string(b), nil
	case "ascii":
		var sb strings.Builder
		for _, c := range b {
			if c > 0x7f {
				if !replace {
					return "", fmt.Errorf("byte 0x%02X is not ASCII", c)
				}
				sb.WriteRune('�')
				continue
			}
			sb.WriteByte(c)
		}
		return sb.String(), nil
	}
	out, err := enc.NewDecoder().Bytes(b)
	if err != nil {
		return "", fmt.Errorf("cannot decode the bytes as %s: %v", e.name, err)
	}
	s := string(out)
	if !replace && strings.ContainsRune(s, '�') {
		return "", fmt.Errorf("the bytes contain sequences that are invalid in %s (use errors=\"replace\" to substitute U+FFFD)", e.name)
	}
	return s, nil
}

// ---- mojibake ----

// fixRuns reinterprets runs of non-ASCII characters: each rune is mapped back
// to the byte it would have in charset cs and, when those bytes form valid
// UTF-8, the run is replaced by the decoded text.
func fixRuns(text, cs string) (string, bool) {
	_, enc, err := findCS(cs)
	if err != nil {
		return text, false
	}
	en := enc.NewEncoder()
	rs := []rune(text)
	var out []rune
	changed := false
	for i := 0; i < len(rs); {
		if rs[i] < 0x80 {
			out = append(out, rs[i])
			i++
			continue
		}
		// extend the run while runes are non-ASCII and encodable
		j := i
		var bs []byte
		for j < len(rs) && rs[j] >= 0x80 {
			b, err := en.Bytes([]byte(string(rs[j])))
			if err != nil || len(b) != 1 {
				break
			}
			bs = append(bs, b[0])
			j++
		}
		if j == i {
			out = append(out, rs[i])
			i++
			continue
		}
		fixed := false
		if utf8.Valid(bs) {
			dec := []rune(string(bs))
			if len(dec) < j-i && !hasCtl(dec) {
				out = append(out, dec...)
				changed, fixed = true, true
			}
		}
		if !fixed {
			out = append(out, rs[i:j]...)
		}
		i = j
	}
	return string(out), changed
}

func hasCtl(rs []rune) bool {
	for _, r := range rs {
		if r < 0x20 || r >= 0x7f && r < 0xa0 || r == 0xfffd {
			return true
		}
	}
	return false
}

func mojibakeFix(text string) string {
	for i := 0; i < 4; i++ {
		changed := false
		for _, cs := range []string{"windows-1252", "iso-8859-1"} {
			if t, ok := fixRuns(text, cs); ok {
				text, changed = t, true
				break
			}
		}
		if !changed {
			break
		}
	}
	return text
}

var mojibakeCharsets = []string{"windows-1252", "iso-8859-1", "windows-1250", "windows-1251", "windows-1253", "windows-1254", "windows-1255", "windows-1256", "windows-1257", "windows-1258", "iso-8859-2", "iso-8859-15", "koi8-r", "koi8-u", "ibm437", "ibm850", "ibm852", "ibm866", "macintosh"}

func mojibakeAll(text string) map[string]string {
	res := map[string]string{}
	for _, cs := range mojibakeCharsets {
		cur := text
		for round := 1; round <= 3; round++ {
			t, ok := fixRuns(cur, cs)
			if !ok {
				break
			}
			cur = t
			key := cs
			if round > 1 {
				key = fmt.Sprintf("%s x%d", cs, round)
			}
			res[key] = cur
		}
	}
	return res
}

// ---- charset detection ----

func detectCharset(b []byte) []string {
	var out []string
	add := func(s string) {
		for _, o := range out {
			if o == s {
				return
			}
		}
		out = append(out, s)
	}
	if len(b) == 0 {
		return []string{"ascii", "utf-8"}
	}
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		add("utf-8-sig")
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE, 0, 0}):
		add("utf-32")
	case bytes.HasPrefix(b, []byte{0, 0, 0xFE, 0xFF}):
		add("utf-32be")
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		add("utf-16")
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		add("utf-16be")
	}
	ascii := true
	for _, c := range b {
		if c > 0x7f {
			ascii = false
			break
		}
	}
	if ascii {
		hasESC := bytes.IndexByte(b, 0x1b) >= 0
		if hasESC {
			add("iso-2022-jp")
		}
		// NUL interleaving suggests UTF-16 without BOM
		even, odd := 0, 0
		for i, c := range b {
			if c == 0 {
				if i%2 == 0 {
					even++
				} else {
					odd++
				}
			}
		}
		if len(b) >= 4 && len(b)%2 == 0 {
			if odd*2 >= len(b)-2 && even == 0 {
				add("utf-16le")
			}
			if even*2 >= len(b)-2 && odd == 0 {
				add("utf-16be")
			}
		}
		if len(out) == 0 || !bytes.Contains(b, []byte{0}) {
			add("ascii")
			add("utf-8")
		}
		return out
	}
	if utf8.Valid(b) {
		add("utf-8")
	}
	// strict multibyte candidates
	for _, cs := range []string{"shift_jis", "euc-jp", "gbk", "big5", "euc-kr"} {
		if utf8.Valid(b) {
			break
		}
		if s, err := csDecode(b, cs, false); err == nil && !strings.ContainsAny(s, "\x00") {
			// reject if most characters are rare CJK noise: accept when decode round trips
			if re, err := csEncode(s, cs, false); err == nil && bytes.Equal(re, b) {
				add(cs)
			}
		}
	}
	// single-byte guesses by byte statistics
	c1, high, cyr, nonLatin := 0, 0, 0, 0
	for _, c := range b {
		if c >= 0x80 {
			high++
		}
		if c >= 0x80 && c < 0xA0 {
			c1++
		}
		if c >= 0xC0 {
			cyr++
		}
		if c == 0x81 || c == 0x8D || c == 0x8F || c == 0x90 || c == 0x9D {
			nonLatin++
		}
	}
	if !utf8.Valid(b) {
		if c1 > 0 && nonLatin == 0 {
			add("windows-1252")
		}
		if cyr*2 > high && high*3 >= len(b) {
			add("windows-1251")
			add("koi8-r")
		}
		if c1 == 0 {
			add("iso-8859-1")
			add("windows-1252")
			add("iso-8859-15")
		}
		add("windows-1252")
		add("ibm437")
	}
	return out
}
