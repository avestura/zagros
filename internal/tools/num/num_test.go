package num

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "num") }

type m = map[string]any

func run(t *testing.T, name string, args m) (string, error) {
	t.Helper()
	v, err := tools.Call("num", name, args)
	if err != nil {
		return "", err
	}
	return tools.Format(v), nil
}

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		args m
		want string
	}{
		{"to_hex", m{"value": "255"}, "ff"},
		{"to_hex", m{"value": "255", "prefix": true, "uppercase": true}, "0xFF"},
		{"to_hex", m{"value": "65535", "group": int64(2)}, "ff ff"},
		{"to_hex", m{"value": "-255"}, "-ff"},
		{"to_hex", m{"value": "ff"}, "ff"},
		{"to_bin", m{"value": "10"}, "1010"},
		{"to_bin", m{"value": "0xFF", "prefix": true, "group": int64(4)}, "0b1111 1111"},
		{"to_oct", m{"value": "64"}, "100"},
		{"to_dec", m{"value": "0xFF"}, "255"},
		{"to_dec", m{"value": "1_000"}, "1000"},
		{"to_base", m{"value": "ff", "from_base": int64(16), "to_base": int64(2)}, "11111111"},
		{"to_base", m{"value": "35", "to_base": int64(36)}, "z"},
		{"from_bin", m{"value": "0b1010"}, "10"},
		{"from_oct", m{"value": "17"}, "15"},
		{"from_hex", m{"value": "0xFF"}, "255"},
		{"from_hex", m{"value": "0b1"}, "177"},
		{"detect_base", m{"value": "0b101"}, "bin"},
		{"detect_base", m{"value": "ff"}, "hex"},
		{"detect_base", m{"value": "99"}, "dec"},
		{"calc_int", m{"expr": "0xFF+1"}, "256"},
		{"calc_int", m{"expr": "2**64"}, "18446744073709551616"},
		{"calc_int", m{"expr": "0b1010<<2"}, "40"},
		{"calc_int", m{"expr": "-2**2"}, "-4"},
		{"calc_int", m{"expr": "2**3**2"}, "512"},
		{"calc_int", m{"expr": "(1+2)*3 - 10/3"}, "6"},
		{"calc_int", m{"expr": "-7/2"}, "-3"},
		{"calc_int", m{"expr": "-7%3"}, "-1"},
		{"calc_int", m{"expr": "0xF0 & 0x3C | 1 ^ 2"}, "51"},
		{"calc_int", m{"expr": "~0"}, "-1"},
		{"calc_int", m{"expr": "gcd(12,18)+lcm(4,6)+max(1,5,3)+min(4,2)+abs(-3)"}, "28"},
		{"calc_int", m{"expr": "pow(4,13,497)"}, "445"},
		{"calc_int", m{"expr": "bits(255)"}, "8"},
		{"calc_int", m{"expr": "FF+1", "base": int64(16)}, "256"},
		{"calc_int", m{"expr": "255", "output": "hex"}, "0xff"},
		{"calc_int", m{"expr": "1_000 * 1_000"}, "1000000"},
		{"hex_add", m{"a": "ff", "b": "1"}, "100"},
		{"hex_sub", m{"a": "0", "b": "0x10"}, "-10"},
		{"hex_mul", m{"a": "ffffffffffffffff", "b": "ffffffffffffffff"}, "FFFFFFFFFFFFFFFE0000000000000001"},
		{"hex_div", m{"a": "ff", "b": "10"}, "quotient: F\nremainder: F"},
		{"bin_add", m{"a": "1010", "b": "11"}, "1101"},
		{"bin_sub", m{"a": "11", "b": "100"}, "-1"},
		{"bin_mul", m{"a": "101", "b": "11"}, "1111"},
		{"bin_div", m{"a": "1111", "b": "10"}, "quotient: 111\nremainder: 1"},
		{"bit_and", m{"a": "0xF0", "b": "0x3C", "output": "hex"}, "0x30"},
		{"bit_or", m{"a": "0xF0", "b": "0x0F"}, "255"},
		{"bit_xor", m{"a": "5", "b": "3"}, "6"},
		{"bit_not", m{"value": "0x0F", "bits": int64(8), "output": "bin"}, "0b11110000"},
		{"bit_not", m{"value": "0"}, "-1"},
		{"shift_left", m{"value": "1", "n": int64(100)}, "1267650600228229401496703205376"},
		{"shift_right", m{"value": "-5", "n": int64(1)}, "-3"},
		{"rotate_left", m{"value": "0x81", "n": int64(1), "bits": int64(8), "output": "hex"}, "0x03"},
		{"rotate_right", m{"value": "0x01", "n": int64(1), "bits": int64(8), "output": "bin"}, "0b10000000"},
		{"set_bit", m{"value": "0", "index": int64(3)}, "8"},
		{"clear_bit", m{"value": "15", "index": int64(0)}, "14"},
		{"toggle_bit", m{"value": "5", "index": int64(1)}, "7"},
		{"test_bit", m{"value": "5", "index": int64(2)}, "true"},
		{"test_bit", m{"value": "5", "index": int64(1)}, "false"},
		{"bit_length", m{"value": "255"}, "8"},
		{"bit_length", m{"value": "256"}, "9"},
		{"popcount", m{"value": "0xFF"}, "8"},
		{"popcount", m{"value": "-1", "bits": int64(16)}, "16"},
		{"parity", m{"value": "7"}, "1"},
		{"leading_zeros", m{"value": "1", "bits": int64(32)}, "31"},
		{"trailing_zeros", m{"value": "8"}, "3"},
		{"reverse_bits", m{"value": "1", "bits": int64(8), "output": "bin"}, "0b10000000"},
		{"swap_endian", m{"value": "0x12345678", "bytes": int64(4), "output": "hex"}, "0x78563412"},
		{"bits_view", m{"value": "0xA5"}, "   7    3\n1010 0101"},
		{"to_twos", m{"value": "-5", "bits": int64(8)}, "0b11111011"},
		{"to_twos", m{"value": "-1", "bits": int64(16), "output": "hex"}, "0xffff"},
		{"from_twos", m{"value": "0xFB", "bits": int64(8)}, "-5"},
		{"from_twos", m{"value": "0x7F", "bits": int64(8)}, "127"},
		{"sign_extend", m{"value": "0x80", "from_bits": int64(8), "to_bits": int64(16)}, "0xff80"},
		{"mask", m{"bits": int64(8), "shift": int64(4)}, "0xff0"},
		{"gray_encode", m{"value": "7"}, "4"},
		{"gray_decode", m{"value": "4"}, "7"},
		{"gray_decode", m{"value": "1000"}, grayDecodeRef(1000)},
		{"ieee754_decode", m{"hex": "3dcccccd"}, "0.1"},
		{"ieee754_decode", m{"hex": "0x3FB999999999999A"}, "0.1"},
		{"ieee754_decode", m{"hex": "7fc00000"}, "NaN"},
		{"float_to_fraction", m{"x": "0.125"}, "1/8"},
		{"float_to_fraction", m{"x": "3.14159", "max_denominator": int64(1000)}, "355/113"},
		{"float_to_fraction", m{"x": "0.3333333333333", "max_denominator": int64(100)}, "1/3"},
		{"float_to_fraction", m{"x": "-2.5"}, "-5/2"},
		{"nearest_float32", m{"x": "0.1"}, "0.100000001490116119384765625"},
		{"roman_encode", m{"value": "1994"}, "MCMXCIV"},
		{"roman_encode", m{"value": "3999"}, "MMMCMXCIX"},
		{"roman_encode", m{"value": "4"}, "IV"},
		{"roman_decode", m{"value": "mcmxciv"}, "1994"},
		{"roman_decode", m{"value": "XLII"}, "42"},
		{"number_to_words", m{"value": "1234"}, "one thousand two hundred thirty-four"},
		{"number_to_words", m{"value": "0"}, "zero"},
		{"number_to_words", m{"value": "-15"}, "minus fifteen"},
		{"number_to_words", m{"value": "1000000"}, "one million"},
		{"number_to_words", m{"value": "21", "ordinal": true}, "twenty-first"},
		{"number_to_words", m{"value": "1000000000000000000000000000000000000"}, "one undecillion"},
		{"ordinal_words", m{"value": "100"}, "one hundredth"},
		{"ordinal_words", m{"value": "12"}, "twelfth"},
		{"ordinal_words", m{"value": "40"}, "fortieth"},
		{"words_to_number", m{"text": "two hundred and five"}, "205"},
		{"words_to_number", m{"text": "minus twenty-one"}, "-21"},
		{"words_to_number", m{"text": "one million two hundred thirty-four thousand five hundred sixty-seven"}, "1234567"},
		{"words_to_number", m{"text": "twenty-first"}, "21"},
		{"ordinal", m{"value": "1"}, "1st"},
		{"ordinal", m{"value": "12"}, "12th"},
		{"ordinal", m{"value": "23"}, "23rd"},
		{"ordinal", m{"value": "111"}, "111th"},
		{"ordinal", m{"value": "-2"}, "-2nd"},
		{"add_thousands_separators", m{"value": "1234567.891"}, "1,234,567.891"},
		{"add_thousands_separators", m{"value": "-1234", "sep": "."}, "-1.234"},
		{"strip_separators", m{"value": "1,234,567.89"}, "1234567.89"},
		{"strip_separators", m{"value": "1.234,56", "decimal_point": ","}, "1234.56"},
		{"scientific", m{"value": "123456789", "digits": int64(3)}, "1.235e+8"},
		{"scientific", m{"value": "0.00012", "digits": int64(1)}, "1.2e-4"},
		{"format_number", m{"value": "1234567.891", "decimals": int64(2)}, "1,234,567.89"},
		{"format_number", m{"value": "1234.5", "decimals": int64(2), "thousands": ".", "decimal_point": ","}, "1.234,50"},
		{"format_number", m{"value": "-0.001", "decimals": int64(2)}, "0.00"},
		{"format_number", m{"value": "2.5", "decimals": int64(0)}, "3"},
		{"parse_human_number", m{"text": "1.5k"}, "1500"},
		{"parse_human_number", m{"text": "2M"}, "2000000"},
		{"parse_human_number", m{"text": "3e6"}, "3000000"},
		{"parse_human_number", m{"text": "$1,234.5"}, "1234.5"},
		{"parse_human_number", m{"text": "2.5 billion"}, "2500000000"},
		{"parse_human_number", m{"text": "50%"}, "0.5"},
		{"humanize_bytes", m{"value": "1536", "binary": true}, "1.5 KiB"},
		{"humanize_bytes", m{"value": "1500"}, "1.5 KB"},
		{"humanize_bytes", m{"value": "999999"}, "1 MB"},
		{"humanize_bytes", m{"value": "12"}, "12 B"},
		{"parse_bytes", m{"text": "1.5 GiB"}, "1610612736"},
		{"parse_bytes", m{"text": "200MB"}, "200000000"},
		{"parse_bytes", m{"text": "4k", "binary": true}, "4096"},
		{"parse_bytes", m{"text": "10"}, "10"},
		{"number_to_binary_bytes", m{"value": "258"}, "01\n02"},
		{"number_to_binary_bytes", m{"value": "258", "endian": "little", "size": int64(4)}, "02\n01\n00\n00"},
		{"number_to_binary_bytes", m{"value": "-2"}, "fe"},
		{"number_to_binary_bytes", m{"value": "0"}, "00"},
		{"bytes_to_number", m{"bytes": []string{"01", "02"}}, "258"},
		{"bytes_to_number", m{"bytes": []string{"ff"}, "signed": true}, "-1"},
		{"digits", m{"value": "1203"}, "1\n2\n0\n3"},
		{"digit_sum", m{"value": "9875"}, "29"},
		{"digital_root", m{"value": "9875"}, "2"},
		{"digital_root", m{"value": "0"}, "0"},
		{"reverse_digits", m{"value": "-1230"}, "-321"},
		{"is_palindrome_number", m{"value": "12321"}, "true"},
		{"is_palindrome_number", m{"value": "9", "base": int64(2)}, "true"},
		{"is_palindrome_number", m{"value": "123"}, "false"},
		{"is_prime", m{"value": "97"}, "true"},
		{"is_prime", m{"value": "1"}, "false"},
		{"is_prime", m{"value": "170141183460469231731687303715884105727"}, "true"},
		{"next_prime", m{"value": "100"}, "101"},
		{"next_prime", m{"value": "-3"}, "2"},
		{"next_prime", m{"value": "2"}, "3"},
		{"prev_prime", m{"value": "100"}, "97"},
		{"prev_prime", m{"value": "3"}, "2"},
		{"prime_factors", m{"value": "360"}, "2\n2\n2\n3\n3\n5"},
		{"prime_factors", m{"value": "600851475143"}, "71\n839\n1471\n6857"},
		{"factorization", m{"value": "360"}, "2^3 * 3^2 * 5"},
		{"factorization", m{"value": "1"}, "1"},
		{"factorization", m{"value": "18446744073709551617"}, "274177 * 67280421310721"},
		{"divisors", m{"value": "28"}, "1\n2\n4\n7\n14\n28"},
		{"is_perfect", m{"value": "28"}, "true"},
		{"is_perfect", m{"value": "12"}, "false"},
		{"is_armstrong", m{"value": "153"}, "true"},
		{"is_armstrong", m{"value": "154"}, "false"},
		{"is_fibonacci", m{"value": "144"}, "true"},
		{"is_fibonacci", m{"value": "145"}, "false"},
		{"fibonacci", m{"value": "10"}, "55"},
		{"fibonacci", m{"value": "100"}, "354224848179261915075"},
		{"factorial", m{"value": "20"}, "2432902008176640000"},
		{"factorial", m{"value": "25"}, "15511210043330985984000000"},
		{"collatz_steps", m{"value": "27"}, "111"},
		{"collatz_steps", m{"value": "1"}, "0"},
		{"euler_phi", m{"value": "36"}, "12"},
		{"euler_phi", m{"value": "97"}, "96"},
		{"gcd", m{"a": "12", "b": "18"}, "6"},
		{"gcd", m{"a": "0", "b": "-5"}, "5"},
		{"lcm", m{"a": "4", "b": "6"}, "12"},
		{"gcd_list", m{"values": []string{"12", "18", "30"}}, "6"},
		{"lcm_list", m{"values": []string{"2", "3", "4"}}, "12"},
		{"modpow", m{"base": "4", "exp": "13", "mod": "497"}, "445"},
		{"modpow", m{"base": "2", "exp": "-1", "mod": "7"}, "4"},
		{"modinv", m{"a": "3", "m": "11"}, "4"},
		{"isqrt", m{"value": "99"}, "9"},
		{"isqrt", m{"value": "100000000000000000000"}, "10000000000"},
		{"icbrt", m{"value": "1000"}, "10"},
		{"icbrt", m{"value": "-30"}, "-3"},
		{"nth_root", m{"value": "1024", "n": int64(10)}, "2"},
		{"nth_root", m{"value": "1023", "n": int64(10)}, "1"},
		{"luhn_check", m{"value": "4539 1488 0343 6467"}, "true"},
		{"luhn_check", m{"value": "4539 1488 0343 6468"}, "false"},
		{"luhn_generate", m{"value": "7992739871"}, "3"},
		{"luhn_complete", m{"value": "7992739871"}, "79927398713"},
		{"isbn10_check", m{"value": "0-306-40615-2"}, "true"},
		{"isbn10_check", m{"value": "080442957X"}, "true"},
		{"isbn10_check", m{"value": "0306406153"}, "false"},
		{"isbn13_check", m{"value": "978-3-16-148410-0"}, "true"},
		{"isbn13_check", m{"value": "9783161484101"}, "false"},
		{"isbn10_to_13", m{"value": "0306406152"}, "9780306406157"},
		{"isbn13_to_10", m{"value": "9780306406157"}, "0306406152"},
		{"iban_check", m{"value": "GB82 WEST 1234 5698 7654 32"}, "true"},
		{"iban_check", m{"value": "GB82 WEST 1234 5698 7654 33"}, "false"},
		{"iban_check", m{"value": "DE89370400440532013000"}, "true"},
		{"iban_format", m{"value": "gb82west12345698765432"}, "GB82 WEST 1234 5698 7654 32"},
		{"ean13_check", m{"value": "4006381333931"}, "true"},
		{"ean13_check", m{"value": "4006381333932"}, "false"},
		{"ean8_check", m{"value": "96385074"}, "true"},
		{"upc_check", m{"value": "036000291452"}, "true"},
		{"gtin_check", m{"value": "4006381333931"}, "true"},
		{"gtin_generate", m{"value": "400638133393"}, "1"},
		{"verhoeff_check", m{"value": "2363"}, "true"},
		{"verhoeff_check", m{"value": "2364"}, "false"},
		{"verhoeff_generate", m{"value": "236"}, "3"},
		{"verhoeff_generate", m{"value": "12345"}, "1"},
		{"damm_check", m{"value": "5724"}, "true"},
		{"damm_check", m{"value": "5727"}, "false"},
		{"damm_generate", m{"value": "572"}, "4"},
	}
	for _, tc := range tests {
		got, err := run(t, tc.name, tc.args)
		if err != nil {
			t.Errorf("%s(%v): unexpected error %v", tc.name, tc.args, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s(%v) = %q, want %q", tc.name, tc.args, got, tc.want)
		}
	}
}

func grayDecodeRef(g uint64) string {
	n := g
	for s := g >> 1; s != 0; s >>= 1 {
		n ^= s
	}
	var b strings.Builder
	b.WriteString(itoa(n))
	return b.String()
}

func itoa(n uint64) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

func TestMaps(t *testing.T) {
	check := func(name string, args m, want map[string]string) {
		t.Helper()
		v, err := tools.Call("num", name, args)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := v.(map[string]string)
		for k, w := range want {
			if got[k] != w {
				t.Errorf("%s[%s] = %q, want %q", name, k, got[k], w)
			}
		}
	}
	check("base_all", m{"value": "0xFF"}, map[string]string{"bin": "11111111", "oct": "377", "dec": "255", "hex": "FF", "base32": "7v", "base36": "73", "detected_base": "16"})
	check("base_all", m{"value": "255"}, map[string]string{"hex": "FF", "detected_base": "10"})
	check("calc_all", m{"expr": "0xFF + 1"}, map[string]string{"dec": "256", "hex": "0x100", "bin": "0b100000000", "oct": "0o400", "bits": "9", "int8": "0", "uint8": "0", "int16": "256"})
	check("calc_all", m{"expr": "-1"}, map[string]string{"dec": "-1", "int8": "-1", "uint8": "255", "uint32": "4294967295", "bits": "1"})
	check("ieee754_encode", m{"value": "0.1", "bits": int64(32)}, map[string]string{
		"hex": "3dcccccd", "sign": "0", "exponent": "01111011", "mantissa": "10011001100110011001101",
		"exponent_value": "-4", "class": "normal", "value": "0.100000001490116119384765625"})
	check("ieee754_encode", m{"value": "0.1", "bits": int64(64)}, map[string]string{"hex": "3fb999999999999a", "exponent": "01111111011"})
	check("ieee754_encode", m{"value": "-2", "bits": int64(32)}, map[string]string{"hex": "c0000000", "sign": "1", "value": "-2"})
	check("ieee754_encode", m{"value": "inf", "bits": int64(32)}, map[string]string{"hex": "7f800000", "class": "inf"})
	check("bytes_table", m{"value": "1073741824"}, map[string]string{"GiB": "1", "GB": "1.073742", "bit": "8589934592", "MiB": "1024"})
	check("bytes_table", m{"value": "1.5 GiB"}, map[string]string{"GiB": "1.5"})
}

func TestErrors(t *testing.T) {
	bad := []struct {
		name string
		args m
	}{
		{"to_hex", m{"value": ""}},
		{"to_hex", m{"value": "xyz"}},
		{"to_hex", m{"value": "--5"}},
		{"to_hex", m{"value": "0x"}},
		{"to_base", m{"value": "5", "to_base": int64(1)}},
		{"to_base", m{"value": "5", "to_base": int64(99)}},
		{"from_bin", m{"value": "102"}},
		{"calc_int", m{"expr": ""}},
		{"calc_int", m{"expr": "1/0"}},
		{"calc_int", m{"expr": "1%0"}},
		{"calc_int", m{"expr": "2**-1"}},
		{"calc_int", m{"expr": "2**99999999"}},
		{"calc_int", m{"expr": "1<<99999999"}},
		{"calc_int", m{"expr": "(1+2"}},
		{"calc_int", m{"expr": "1 +"}},
		{"calc_int", m{"expr": "1 $ 2"}},
		{"calc_int", m{"expr": "foo(1)"}},
		{"calc_int", m{"expr": "abc"}},
		{"calc_int", m{"expr": strings.Repeat("(", 5000)}},
		{"calc_int", m{"expr": "fact(99999)"}},
		{"calc_int", m{"expr": "1", "output": "weird"}},
		{"hex_div", m{"a": "ff", "b": "0"}},
		{"hex_add", m{"a": "zz", "b": "1"}},
		{"bit_and", m{"a": "x", "b": "1"}},
		{"shift_left", m{"value": "1", "n": int64(1 << 40)}},
		{"shift_left", m{"value": "1", "n": int64(-1)}},
		{"rotate_left", m{"value": "1", "bits": int64(0)}},
		{"set_bit", m{"value": "1", "index": int64(-1)}},
		{"leading_zeros", m{"value": "256", "bits": int64(8)}},
		{"to_twos", m{"value": "200", "bits": int64(7)}},
		{"to_twos", m{"value": "-129", "bits": int64(8)}},
		{"from_twos", m{"value": "256", "bits": int64(8)}},
		{"sign_extend", m{"value": "1", "from_bits": int64(16), "to_bits": int64(8)}},
		{"mask", m{"bits": int64(1 << 40)}},
		{"gray_encode", m{"value": "-1"}},
		{"ieee754_encode", m{"value": "abc"}},
		{"ieee754_encode", m{"value": "1", "bits": int64(16)}},
		{"ieee754_decode", m{"hex": "123"}},
		{"ieee754_decode", m{"hex": "zzzzzzzz"}},
		{"float_to_fraction", m{"x": "abc"}},
		{"float_to_fraction", m{"x": "1e99999"}},
		{"float_to_fraction", m{"x": "1/0"}},
		{"nearest_float32", m{"x": "1e50"}},
		{"roman_encode", m{"value": "0"}},
		{"roman_encode", m{"value": "4000"}},
		{"roman_decode", m{"value": "ABC"}},
		{"roman_decode", m{"value": ""}},
		{"number_to_words", m{"value": "1" + strings.Repeat("0", 39)}},
		{"number_to_words", m{"value": "5", "lang": "fr"}},
		{"words_to_number", m{"text": "banana"}},
		{"words_to_number", m{"text": ""}},
		{"add_thousands_separators", m{"value": "12a"}},
		{"strip_separators", m{"value": "abc"}},
		{"scientific", m{"value": "1e99999"}},
		{"format_number", m{"value": "x"}},
		{"parse_human_number", m{"text": "5zz"}},
		{"parse_human_number", m{"text": ""}},
		{"parse_bytes", m{"text": "5 zorks"}},
		{"parse_bytes", m{"text": "99999999999 EB"}},
		{"number_to_binary_bytes", m{"value": "256", "size": int64(1)}},
		{"number_to_binary_bytes", m{"value": "1", "endian": "middle"}},
		{"bytes_to_number", m{"bytes": []string{"1ff"}}},
		{"bytes_to_number", m{"bytes": []string{}}},
		{"is_prime", m{"value": "abc"}},
		{"is_prime", m{"value": "1" + strings.Repeat("0", 3000)}},
		{"prev_prime", m{"value": "2"}},
		{"prime_factors", m{"value": "1"}},
		{"prime_factors", m{"value": "1" + strings.Repeat("0", 200)}},
		{"factorial", m{"value": "20001"}},
		{"factorial", m{"value": "-1"}},
		{"fibonacci", m{"value": "100001"}},
		{"collatz_steps", m{"value": "0"}},
		{"euler_phi", m{"value": "0"}},
		{"modpow", m{"base": "2", "exp": "3", "mod": "0"}},
		{"modinv", m{"a": "2", "m": "4"}},
		{"nth_root", m{"value": "-4", "n": int64(2)}},
		{"nth_root", m{"value": "4", "n": int64(0)}},
		{"isqrt", m{"value": "-1"}},
		{"luhn_check", m{"value": "12ab"}},
		{"isbn10_to_13", m{"value": "0306406153"}},
		{"isbn13_to_10", m{"value": "9791234567896"}},
		{"iban_check", m{"value": "!!"}},
		{"gtin_generate", m{"value": "123"}},
		{"verhoeff_check", m{"value": "abc"}},
	}
	for _, tc := range bad {
		if _, err := run(t, tc.name, tc.args); err == nil {
			t.Errorf("%s(%v): expected an error", tc.name, tc.args)
		}
	}
}

// TestGarbage feeds odd strings to every function with string parameters; it
// only requires that nothing panics and that results are one of the allowed types.
func TestGarbage(t *testing.T) {
	inputs := []string{"", " ", "-", "+", "0x", "0b", "0xZZ", "abc", "1e999999", "NaN", "inf", "\x00", "ÿ☃", "1/0", "((((", "--1", "1_", "_1", strings.Repeat("9", 300), "1 2 3", "٣", "0o9"}
	for _, f := range tools.All() {
		if f.Group != "num" {
			continue
		}
		for _, in := range inputs {
			args := m{}
			for _, p := range f.Params {
				if p.Required {
					switch p.Type {
					case "str":
						args[p.Name] = in
					case "int":
						args[p.Name] = int64(3)
					case "list[str]":
						args[p.Name] = []string{in}
					}
				}
			}
			if len(f.Params) > 0 && f.Params[0].Type == "str" {
				args[f.Params[0].Name] = in
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s(%q) panicked: %v", f.Key(), in, r)
					}
				}()
				if f.Accepts != nil {
					f.Accepts(in)
				}
				v, err := f.Invoke(args)
				if err == nil {
					switch v.(type) {
					case string, int64, float64, bool, []string, []float64, map[string]string, nil:
					default:
						t.Errorf("%s returned disallowed type %T", f.Key(), v)
					}
				}
			}()
		}
	}
}

func TestResultTypesMatchReturns(t *testing.T) {
	for _, f := range tools.All() {
		if f.Group != "num" || f.Example == "" {
			continue
		}
		// Examples are Starlark; just make sure they mention a real function.
		if _, ok := tools.Lookup("num", f.Name); !ok {
			t.Errorf("%s not registered", f.Name)
		}
	}
}

func TestBitsViewWide(t *testing.T) {
	got, err := run(t, "bits_view", m{"value": "-1"})
	if err != nil || got != "   7    3\n1111 1111" {
		t.Errorf("bits_view(-1) = %q, %v", got, err)
	}
	got, _ = run(t, "bits_view", m{"value": "0x1FF"})
	if !strings.HasSuffix(got, "0001 1111 1111") {
		t.Errorf("bits_view(0x1FF) = %q", got)
	}
}
