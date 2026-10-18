package num

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// digitsOnly strips spaces, dashes and dots and requires only digits.
func digitsOnly(s string, allowX bool) (string, error) {
	t := strings.NewReplacer(" ", "", "-", "", ".", "", "_", "").Replace(strings.TrimSpace(s))
	if t == "" {
		return "", errors.New("empty input")
	}
	if len(t) > 5000 {
		return "", errors.New("input is too long")
	}
	for i, c := range t {
		if c >= '0' && c <= '9' || allowX && (c == 'X' || c == 'x') && i == len(t)-1 {
			continue
		}
		return "", fmt.Errorf("unexpected character %q at position %d", string(c), i+1)
	}
	return t, nil
}

// luhnSum returns the Luhn sum of digits, doubling every second digit from the
// right starting with the digit at offset (0 = rightmost is not doubled).
func luhnSum(d string, doubleFirst bool) int {
	sum := 0
	dbl := doubleFirst
	for i := len(d) - 1; i >= 0; i-- {
		v := int(d[i] - '0')
		if dbl {
			v *= 2
			if v > 9 {
				v -= 9
			}
		}
		sum += v
		dbl = !dbl
	}
	return sum
}

func luhnDigit(d string) int { return (10 - luhnSum(d, true)%10) % 10 }

func luhnValid(d string) bool { return len(d) >= 2 && luhnSum(d, false)%10 == 0 }

func gtinValid(d string) bool {
	switch len(d) {
	case 8, 12, 13, 14:
	default:
		return false
	}
	return gtinDigit(d[:len(d)-1]) == int(d[len(d)-1]-'0')
}

// gtinDigit computes the check digit for the body (weights 3,1 from the right).
func gtinDigit(body string) int {
	sum := 0
	for i := len(body) - 1; i >= 0; i-- {
		w := 1
		if (len(body)-1-i)%2 == 0 {
			w = 3
		}
		sum += int(body[i]-'0') * w
	}
	return (10 - sum%10) % 10
}

func isbn10Digit(body string) byte {
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(body[i]-'0') * (10 - i)
	}
	c := (11 - sum%11) % 11
	if c == 10 {
		return 'X'
	}
	return byte('0' + c)
}

func isbn10Valid(d string) bool {
	if len(d) != 10 {
		return false
	}
	sum := 0
	for i := 0; i < 10; i++ {
		v := int(d[i] - '0')
		if d[i] == 'X' || d[i] == 'x' {
			if i != 9 {
				return false
			}
			v = 10
		}
		sum += v * (10 - i)
	}
	return sum%11 == 0
}

var ibanLen = map[string]int{
	"AD": 24, "AE": 23, "AL": 28, "AT": 20, "AZ": 28, "BA": 20, "BE": 16, "BG": 22, "BH": 22, "BR": 29, "BY": 28,
	"CH": 21, "CR": 22, "CY": 28, "CZ": 24, "DE": 22, "DK": 18, "DO": 28, "EE": 20, "EG": 29, "ES": 24, "FI": 18,
	"FO": 18, "FR": 27, "GB": 22, "GE": 22, "GI": 23, "GL": 18, "GR": 27, "GT": 28, "HR": 21, "HU": 28, "IE": 22,
	"IL": 23, "IQ": 23, "IS": 26, "IT": 27, "JO": 30, "KW": 30, "KZ": 20, "LB": 28, "LC": 32, "LI": 21, "LT": 20,
	"LU": 20, "LV": 21, "MC": 27, "MD": 24, "ME": 22, "MK": 19, "MR": 27, "MT": 31, "MU": 30, "NL": 18, "NO": 15,
	"PK": 24, "PL": 28, "PS": 29, "PT": 25, "QA": 29, "RO": 24, "RS": 22, "SA": 24, "SE": 24, "SI": 19, "SK": 24,
	"SM": 27, "TN": 24, "TR": 26, "UA": 29, "VA": 22, "VG": 24, "XK": 20,
}

func ibanClean(s string) (string, error) {
	t := strings.ToUpper(strings.NewReplacer(" ", "", "-", "", " ", "").Replace(strings.TrimSpace(s)))
	if len(t) < 5 || len(t) > 34 {
		return "", errors.New("an IBAN has 5..34 characters")
	}
	for i, c := range t {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z') {
			return "", fmt.Errorf("unexpected character %q at position %d", string(c), i+1)
		}
	}
	return t, nil
}

func ibanMod97(t string) int {
	r := t[4:] + t[:4]
	rem := 0
	for _, c := range r {
		if c >= 'A' {
			rem = (rem*100 + int(c-'A') + 10) % 97
		} else {
			rem = (rem*10 + int(c-'0')) % 97
		}
	}
	return rem
}

func ibanValid(s string) (bool, error) {
	t, err := ibanClean(s)
	if err != nil {
		return false, err
	}
	if t[0] < 'A' || t[1] < 'A' || t[2] < '0' || t[2] > '9' || t[3] < '0' || t[3] > '9' {
		return false, nil
	}
	if want, ok := ibanLen[t[:2]]; ok && want != len(t) {
		return false, nil
	}
	return ibanMod97(t) == 1, nil
}

var verhoeffD = [10][10]int{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, {1, 2, 3, 4, 0, 6, 7, 8, 9, 5}, {2, 3, 4, 0, 1, 7, 8, 9, 5, 6},
	{3, 4, 0, 1, 2, 8, 9, 5, 6, 7}, {4, 0, 1, 2, 3, 9, 5, 6, 7, 8}, {5, 9, 8, 7, 6, 0, 4, 3, 2, 1},
	{6, 5, 9, 8, 7, 1, 0, 4, 3, 2}, {7, 6, 5, 9, 8, 2, 1, 0, 4, 3}, {8, 7, 6, 5, 9, 3, 2, 1, 0, 4},
	{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
}

var verhoeffInv = [10]int{0, 4, 3, 2, 1, 5, 6, 7, 8, 9}

var verhoeffP = func() (p [8][10]int) {
	p[0] = [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	p[1] = [10]int{1, 5, 7, 6, 2, 8, 3, 0, 9, 4}
	for i := 2; i < 8; i++ {
		for j := 0; j < 10; j++ {
			p[i][j] = p[i-1][p[1][j]]
		}
	}
	return
}()

func verhoeffSum(d string, shift int) int {
	c := 0
	for i := len(d) - 1; i >= 0; i-- {
		c = verhoeffD[c][verhoeffP[(len(d)-1-i+shift)%8][d[i]-'0']]
	}
	return c
}

func verhoeffDigit(d string) int { return verhoeffInv[verhoeffSum(d, 1)] }

var dammTable = [10][10]int{
	{0, 3, 1, 7, 5, 9, 8, 6, 4, 2}, {7, 0, 9, 2, 1, 5, 4, 8, 6, 3}, {4, 2, 0, 6, 8, 7, 1, 3, 5, 9},
	{1, 7, 5, 0, 9, 8, 3, 4, 2, 6}, {6, 1, 2, 3, 0, 4, 5, 9, 7, 8}, {3, 6, 7, 4, 2, 0, 9, 5, 8, 1},
	{5, 8, 6, 9, 7, 2, 0, 1, 3, 4}, {8, 9, 4, 5, 3, 6, 2, 0, 1, 7}, {9, 4, 3, 8, 6, 1, 7, 2, 0, 5},
	{2, 5, 8, 1, 4, 3, 6, 7, 9, 0},
}

func dammInterim(d string) int {
	c := 0
	for i := 0; i < len(d); i++ {
		c = dammTable[c][d[i]-'0']
	}
	return c
}

func looksDigits(minLen int) func(string) bool {
	return func(s string) bool {
		d, err := digitsOnly(s, true)
		return err == nil && len(d) >= minLen
	}
}

func init() {
	check := func(name, title, doc, ex, tags string, order int, minLen int, f func(d string) (bool, error), allowX bool) {
		reg(name, title, doc, "bool", ex, order, tags,
			[]P{rs("value", "Digits; spaces, dashes and dots are ignored.")},
			func(a tools.Args) (any, error) {
				d, err := digitsOnly(a.Str("value"), allowX)
				if err != nil {
					return nil, err
				}
				return f(d)
			}, accepts(looksDigits(minLen)), kw(strings.TrimSuffix(name, "_check")))
	}
	generate := func(name, title, doc, ex, tags string, order int, f func(d string) (string, error)) {
		reg(name, title, doc, "str", ex, order, tags,
			[]P{rs("value", "Digits without the check digit; spaces, dashes and dots are ignored.")},
			func(a tools.Args) (any, error) {
				d, err := digitsOnly(a.Str("value"), false)
				if err != nil {
					return nil, err
				}
				return f(d)
			}, accepts(looksDigits(1)))
	}

	check("luhn_check", "Luhn check", "Validates a number with the Luhn (mod 10) algorithm used by credit cards, IMEI and many ids. Spaces and dashes are ignored; at least 2 digits are required; non-digits are an error.",
		`num.luhn_check("4539 1488 0343 6467")`, "luhn mod 10 credit card validate checksum imei", 60, 2,
		func(d string) (bool, error) { return luhnValid(d), nil }, false)
	generate("luhn_generate", "Luhn check digit", "Computes the single Luhn check digit to append to a number (7992739871 gives 3). Use luhn_complete to get the full number.",
		`num.luhn_generate("7992739871")`, "luhn check digit generate credit card checksum", 61,
		func(d string) (string, error) { return fmt.Sprint(luhnDigit(d)), nil })
	generate("luhn_complete", "Luhn complete", "Appends the Luhn check digit to a number, giving a valid Luhn number (7992739871 gives 79927398713).",
		`num.luhn_complete("7992739871")`, "luhn complete append check digit credit card", 62,
		func(d string) (string, error) { return d + fmt.Sprint(luhnDigit(d)), nil })

	check("isbn10_check", "ISBN-10 check", "Validates an ISBN-10 (weights 10..1, mod 11; the last character may be X). Hyphens and spaces are ignored.",
		`num.isbn10_check("0-306-40615-2")`, "isbn isbn10 book validate checksum", 63, 10,
		func(d string) (bool, error) { return isbn10Valid(d), nil }, true)
	check("isbn13_check", "ISBN-13 check", "Validates an ISBN-13 (EAN-13 check digit, must start with 978 or 979). Hyphens and spaces are ignored.",
		`num.isbn13_check("978-3-16-148410-0")`, "isbn isbn13 book validate checksum ean", 64, 13,
		func(d string) (bool, error) {
			return len(d) == 13 && (strings.HasPrefix(d, "978") || strings.HasPrefix(d, "979")) && gtinValid(d), nil
		}, false)
	reg("isbn10_to_13", "ISBN-10 to ISBN-13", "Converts a valid ISBN-10 to ISBN-13 by prefixing 978 and recomputing the check digit (0306406152 gives 9780306406157). Invalid input is an error.",
		"str", `num.isbn10_to_13("0306406152")`, 65, "isbn convert 10 to 13 isbn13",
		[]P{rs("value", "ISBN-10.")},
		func(a tools.Args) (any, error) {
			d, err := digitsOnly(a.Str("value"), true)
			if err != nil {
				return nil, err
			}
			if !isbn10Valid(d) {
				return nil, errors.New("not a valid ISBN-10")
			}
			body := "978" + d[:9]
			return body + fmt.Sprint(gtinDigit(body)), nil
		}, accepts(looksDigits(10)))
	reg("isbn13_to_10", "ISBN-13 to ISBN-10", "Converts a valid 978-prefixed ISBN-13 to ISBN-10 (9780306406157 gives 0306406152). 979 numbers have no ISBN-10 and give an error.",
		"str", `num.isbn13_to_10("9780306406157")`, 66, "isbn convert 13 to 10 isbn10",
		[]P{rs("value", "ISBN-13.")},
		func(a tools.Args) (any, error) {
			d, err := digitsOnly(a.Str("value"), false)
			if err != nil {
				return nil, err
			}
			if len(d) != 13 || !gtinValid(d) {
				return nil, errors.New("not a valid ISBN-13")
			}
			if !strings.HasPrefix(d, "978") {
				return nil, errors.New("only 978-prefixed ISBN-13 numbers have an ISBN-10")
			}
			body := d[3:12]
			return body + string(isbn10Digit(body)), nil
		}, accepts(looksDigits(13)))

	reg("iban_check", "IBAN check", "Validates an IBAN: structure, the country's fixed length (for ~75 known countries) and the ISO 7064 mod-97 check. Spaces and dashes are ignored, case-insensitive. Returns false for a wrong checksum and an error for illegal characters.",
		"bool", `num.iban_check("GB82 WEST 1234 5698 7654 32")`, 67, "iban bank account validate mod 97 sepa checksum",
		[]P{rs("value", "IBAN.")},
		func(a tools.Args) (any, error) { return ibanValid(a.Str("value")) },
		accepts(func(s string) bool {
			t, err := ibanClean(s)
			return err == nil && len(t) >= 15 && t[0] >= 'A' && t[0] <= 'Z'
		}), kw("iban"))
	reg("iban_format", "IBAN format", "Formats an IBAN in groups of four characters, uppercase (GB82WEST12345698765432 gives \"GB82 WEST 1234 5698 7654 32\"). Does not validate the checksum.",
		"str", `num.iban_format("GB82WEST12345698765432")`, 68, "iban format group four spaces bank account",
		[]P{rs("value", "IBAN, with or without spaces.")},
		func(a tools.Args) (any, error) {
			t, err := ibanClean(a.Str("value"))
			if err != nil {
				return nil, err
			}
			var parts []string
			for len(t) > 4 {
				parts, t = append(parts, t[:4]), t[4:]
			}
			return strings.Join(append(parts, t), " "), nil
		}, accepts(func(s string) bool { t, err := ibanClean(s); return err == nil && len(t) >= 15 && t[0] >= 'A' }))

	check("ean13_check", "EAN-13 check", "Validates an EAN-13 barcode (13 digits, weights 1 and 3, mod 10).",
		`num.ean13_check("4006381333931")`, "ean ean13 barcode gtin validate checksum", 69, 13,
		func(d string) (bool, error) { return len(d) == 13 && gtinValid(d), nil }, false)
	check("ean8_check", "EAN-8 check", "Validates an EAN-8 barcode (8 digits, weights 3 and 1, mod 10).",
		`num.ean8_check("96385074")`, "ean ean8 barcode gtin validate checksum", 70, 8,
		func(d string) (bool, error) { return len(d) == 8 && gtinValid(d), nil }, false)
	check("upc_check", "UPC-A check", "Validates a UPC-A barcode (12 digits, mod 10 check digit).",
		`num.upc_check("036000291452")`, "upc upca barcode gtin validate checksum", 71, 12,
		func(d string) (bool, error) { return len(d) == 12 && gtinValid(d), nil }, false)
	check("gtin_check", "GTIN check", "Validates any GTIN barcode number: GTIN-8, UPC-A (12), EAN-13 or GTIN-14, using the mod 10 check digit.",
		`num.gtin_check("4006381333931")`, "gtin barcode ean upc validate checksum", 72, 8,
		func(d string) (bool, error) { return gtinValid(d), nil }, false)
	generate("gtin_generate", "GTIN check digit", "Computes the barcode check digit for 7, 11, 12 or 13 digits (GTIN-8, UPC-A, EAN-13, GTIN-14); 400638133393 gives 1.",
		`num.gtin_generate("400638133393")`, "gtin ean upc barcode check digit generate", 73,
		func(d string) (string, error) {
			switch len(d) {
			case 7, 11, 12, 13:
				return fmt.Sprint(gtinDigit(d)), nil
			}
			return "", errors.New("expected 7, 11, 12 or 13 digits without the check digit")
		})

	check("verhoeff_check", "Verhoeff check", "Validates a number with the Verhoeff dihedral-group checksum (used by Aadhaar); detects all single-digit and adjacent-transposition errors.",
		`num.verhoeff_check("2363")`, "verhoeff checksum aadhaar validate dihedral", 74, 2,
		func(d string) (bool, error) { return len(d) >= 2 && verhoeffSum(d, 0) == 0, nil }, false)
	generate("verhoeff_generate", "Verhoeff check digit", "Computes the Verhoeff check digit to append to a number (236 gives 3).",
		`num.verhoeff_generate("236")`, "verhoeff check digit generate aadhaar", 75,
		func(d string) (string, error) { return fmt.Sprint(verhoeffDigit(d)), nil })
	check("damm_check", "Damm check", "Validates a number with the Damm algorithm (quasigroup check digit; catches all single-digit and adjacent-transposition errors).",
		`num.damm_check("5724")`, "damm checksum validate quasigroup", 76, 2,
		func(d string) (bool, error) { return len(d) >= 2 && dammInterim(d) == 0, nil }, false)
	generate("damm_generate", "Damm check digit", "Computes the Damm check digit to append to a number (572 gives 4).",
		`num.damm_generate("572")`, "damm check digit generate", 77,
		func(d string) (string, error) { return fmt.Sprint(dammInterim(d)), nil })
}

var _ = big.NewInt
