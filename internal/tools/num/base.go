package num

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func init() {
	reg("base_all", "All bases", "Shows one integer in binary, octal, decimal, hex, base32, base36 and base62 at once, plus the detected input base. Auto-detects a 0x/0b/0o prefix or hex letters; pass from_base to force a base. Works on arbitrarily large integers (strings in, strings out).",
		"dict[str, str]", `num.base_all("0xFF")`, 0, "base convert converter bin oct dec hex binary octal decimal hexadecimal radix",
		[]P{rs("value", "Integer text, e.g. 255, 0xFF, 0b1010, -0o17, 1_000."), baseP("from_base")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", int(a.Int("from_base")))
			if err != nil {
				return nil, err
			}
			from := int(a.Int("from_base"))
			if from == 0 {
				from = autoBase(a.Str("value"))
			}
			return map[string]string{
				"bin": fmtBase(n, 2, false, false, 0), "oct": fmtBase(n, 8, false, false, 0),
				"dec": n.String(), "hex": fmtBase(n, 16, false, true, 0),
				"base32": fmtBase(n, 32, false, false, 0), "base36": fmtBase(n, 36, false, false, 0),
				"base62": fmtBase(n, 62, false, false, 0), "detected_base": fmt.Sprint(from),
			}, nil
		}, kw("hex", "bin", "oct", "dec", "base", "radix"), accepts(parsesAuto))

	reg("to_base", "Convert base", "Converts an integer from one base to another (2..36 for output, 2..62 accepted for input, case-insensitive up to 36). Handles any size and a leading sign; prefixes 0x/0b/0o, underscores and spaces in the input are ignored.",
		"str", `num.to_base("ff", 16, 2)`, 5, "base convert radix change base36 base32 base3",
		[]P{rs("value", "Integer text."), oi("from_base", "Input base 2..62 (0 = auto-detect, a prefix always wins).", 10), oi("to_base", "Output base 2..36.", 16)},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", int(a.Int("from_base")))
			if err != nil {
				return nil, err
			}
			tb, err := intIn("to_base", a.Int("to_base"), 2, 62)
			if err != nil {
				return nil, err
			}
			return fmtBase(n, int(tb), false, false, 0), nil
		}, kw("base"), accepts(parsesAuto))

	conv := func(name, title string, base int, order int, words string) {
		kind := map[int]string{2: "binary", 8: "octal", 10: "decimal", 16: "hexadecimal"}[base]
		params := []P{rs("value", "Integer text (any prefix; decimal unless from_base or a prefix says otherwise)."), baseP("from_base")}
		if base == 16 {
			params = append(params, ob("prefix", "Add a 0x prefix.", false), ob("uppercase", "Use A-F instead of a-f.", false), oi("group", "Insert a space every N digits (0 = none).", 0))
		}
		if base == 2 {
			params = append(params, ob("prefix", "Add a 0b prefix.", false), oi("group", "Insert a space every N digits (0 = none, 4 = nibbles).", 0))
		}
		if base == 8 {
			params = append(params, ob("prefix", "Add a 0o prefix.", false))
		}
		reg(name, title, "Converts an integer to "+kind+". Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).",
			"str", `num.`+name+`("255")`, order, words,
			params, func(a tools.Args) (any, error) {
				fb := int(a.Int("from_base"))
				n, err := argBig(a, "value", fb)
				if err != nil {
					return nil, err
				}
				g := 0
				if base == 16 || base == 2 {
					g = int(a.Int("group"))
					if g < 0 || g > 1024 {
						return nil, fmt.Errorf("group must be 0..1024")
					}
				}
				return fmtBase(n, base, a.Bool("prefix"), a.Bool("uppercase"), g), nil
			}, kw(strings.TrimPrefix(name, "to_")), accepts(parsesAuto))
	}
	conv("to_hex", "To hex", 16, 1, "hex hexadecimal base16 convert decimal tohex")
	conv("to_bin", "To binary", 2, 2, "bin binary base2 convert decimal tobin bits")
	conv("to_oct", "To octal", 8, 3, "oct octal base8 convert decimal tooct")
	conv("to_dec", "To decimal", 10, 4, "dec decimal base10 convert hex binary octal todec")

	from := func(name, title string, base int, words string) {
		kind := map[int]string{2: "binary", 8: "octal", 16: "hexadecimal"}[base]
		reg(name, title, "Reads a "+kind+" integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.",
			"str", `num.`+name+`("`+map[int]string{2: "1010", 8: "17", 16: "ff"}[base]+`")`, 20, words,
			[]P{rs("value", "The "+kind+" digits.")},
			func(a tools.Args) (any, error) {
				n, err := argBig(a, "value", base)
				if err != nil {
					return nil, err
				}
				return n.String(), nil
			}, accepts(func(s string) bool { _, err := parseBig(s, base); return err == nil }))
	}
	from("from_bin", "Binary to decimal", 2, "binary bin decimal convert frombin")
	from("from_oct", "Octal to decimal", 8, "octal oct decimal convert fromoct")
	from("from_hex", "Hex to decimal", 16, "hexadecimal hex decimal convert fromhex")

	reg("detect_base", "Detect base", "Guesses the base of a number text: a 0x/0b/0o prefix gives hex/bin/oct, hex letters (a-f) mean hex, a plain digit run means dec. Returns \"hex\", \"bin\", \"oct\" or \"dec\", or an error if the text is not a number.",
		"str", `num.detect_base("0b1010")`, 30, "detect guess radix base prefix",
		[]P{rs("value", "Number text.")},
		func(a tools.Args) (any, error) {
			if _, err := argBig(a, "value", 0); err != nil {
				return nil, err
			}
			switch autoBase(a.Str("value")) {
			case 16:
				return "hex", nil
			case 2:
				return "bin", nil
			case 8:
				return "oct", nil
			}
			return "dec", nil
		}, accepts(parsesAuto))
}

var _ = big.NewInt
