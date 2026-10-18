# Numbers and bases (`num`)

Base conversion, the hex and binary calculator, bit operations, IEEE 754, roman numerals, number words, primes and checksums.

No permission is needed: these are pure functions. Generated from the tools registry.

111 functions. Use them as `num.<name>(...)` in any extension.

## num.base_all

`num.base_all(value: str, from_base: int = 0) -> Dict[str, str]`

Shows one integer in binary, octal, decimal, hex, base32, base36 and base62 at once, plus the detected input base. Auto-detects a 0x/0b/0o prefix or hex letters; pass from_base to force a base. Works on arbitrarily large integers (strings in, strings out).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text, e.g. 255, 0xFF, 0b1010, -0o17, 1_000. |
| `from_base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `Dict[str, str]`. Example: `num.base_all("0xFF")`

Keywords: base, convert, converter, bin, oct, dec, hex, binary, octal, decimal, hexadecimal, radix

## num.bits_view

`num.bits_view(value: str, bits: int = 0, base: int = 0) -> str`

Draws an integer as nibble-grouped binary with the bit index of each group's most significant bit above it. bits=0 picks 8, 16, 32 ... automatically (negatives use two's complement).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bits` | `int` (default `0`) | Field width 1..65536 (0 = automatic). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `str`. Example: `num.bits_view("0xA5")`

Keywords: bits, binary, view, nibble, grouped, positions, bit, index, layout, register

## num.calc_int

`num.calc_int(expr: str, base: int = 0, output: str = "dec") -> str`

Evaluates an integer expression with arbitrary precision and returns the result as a decimal string (or hex/bin/oct with output). Operators + - * / % ** & | ^ ~ &lt;&lt; &gt;&gt; and parentheses; / truncates toward zero, ** is right-associative and size-capped, ^ is XOR. Literals may be hex 0xFF, binary 0b1010, octal 0o17 or decimal with underscores; functions abs, min, max, gcd, lcm, pow(a,b\[,mod\]), bits, popcount, isqrt, fact. Errors carry the character position.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | Integer expression: literals (255, 0xFF, 0b1010, 0o17, 1_000), + - * / % ** & \| ^ ~ << >> and parentheses. |
| `base` | `int` (default `0`) | Base of unprefixed literals, 2..36 (0 = decimal). With base 16, write FF instead of 0xFF. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.calc_int("0xFF + 1")`

Keywords: hex calculator, binary calculator, programmer calculator, calculator, hex, binary, programmer, calc, bigint, evaluate, expression, 0xff, arithmetic

## num.to_hex

`num.to_hex(value: str, from_base: int = 0, prefix: bool = False, uppercase: bool = False, group: int = 0) -> str`

Converts an integer to hexadecimal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text (any prefix; decimal unless from_base or a prefix says otherwise). |
| `from_base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `prefix` | `bool` (default `False`) | Add a 0x prefix. |
| `uppercase` | `bool` (default `False`) | Use A-F instead of a-f. |
| `group` | `int` (default `0`) | Insert a space every N digits (0 = none). |

Returns `str`. Example: `num.to_hex("255")`

Keywords: decimal to hex, hexadecimal, base 16, hex, base16, convert, decimal, tohex

## num.calc_all

`num.calc_all(expr: str, base: int = 0) -> Dict[str, str]`

Evaluates an integer expression (same syntax as calc_int) and shows the result in dec, hex, bin and oct, with its bit length and the signed int8/16/32/64 and unsigned uint8/16/32/64 interpretations of its low bits.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | Integer expression: literals (255, 0xFF, 0b1010, 0o17, 1_000), + - * / % ** & \| ^ ~ << >> and parentheses. |
| `base` | `int` (default `0`) | Base of unprefixed literals, 2..36 (0 = decimal). With base 16, write FF instead of 0xFF. |

Returns `Dict[str, str]`. Example: `num.calc_all("0b1010 << 2")`

Keywords: calculator, hex, bin, dec, oct, programmer, signed, unsigned, int32, int64, twos, complement, evaluate

## num.to_bin

`num.to_bin(value: str, from_base: int = 0, prefix: bool = False, group: int = 0) -> str`

Converts an integer to binary. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text (any prefix; decimal unless from_base or a prefix says otherwise). |
| `from_base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `prefix` | `bool` (default `False`) | Add a 0b prefix. |
| `group` | `int` (default `0`) | Insert a space every N digits (0 = none, 4 = nibbles). |

Returns `str`. Example: `num.to_bin("255")`

Keywords: decimal to binary, base 2, bits, bin, binary, base2, convert, decimal, tobin

## num.roman_encode

`num.roman_encode(value: str) -> str`

Converts a whole number 1..3999 to a Roman numeral using standard subtractive notation (1994 gives MCMXCIV). Numbers outside 1..3999 are an error because overline (vinculum) notation is not supported.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Decimal integer 1..3999. |

Returns `str`. Example: `num.roman_encode("1994")`

Keywords: integer to roman, number to roman numeral, roman, numeral, encode, number, to, I, V, X, L, C, M

## num.to_oct

`num.to_oct(value: str, from_base: int = 0, prefix: bool = False) -> str`

Converts an integer to octal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text (any prefix; decimal unless from_base or a prefix says otherwise). |
| `from_base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `prefix` | `bool` (default `False`) | Add a 0o prefix. |

Returns `str`. Example: `num.to_oct("255")`

Keywords: decimal to octal, base 8, oct, octal, base8, convert, decimal, tooct

## num.roman_decode

`num.roman_decode(value: str) -> str`

Converts a Roman numeral (case-insensitive, subtractive notation) to a decimal string; MCMXCIV gives 1994. Invalid letters or values outside 1..3999 are errors.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Roman numeral. |

Returns `str`. Example: `num.roman_decode("MCMXCIV")`

Keywords: roman numeral to integer, parse roman numeral, roman, numeral, decode, to, number, I, V, X, L, C, M

## num.to_dec

`num.to_dec(value: str, from_base: int = 0) -> str`

Converts an integer to decimal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text (any prefix; decimal unless from_base or a prefix says otherwise). |
| `from_base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `str`. Example: `num.to_dec("255")`

Keywords: dec, decimal, base10, convert, hex, binary, octal, todec

## num.number_to_words

`num.number_to_words(value: str, lang: str = "en", ordinal: bool = False) -> str`

Spells a whole number as English words (1234 gives "one thousand two hundred thirty-four"). Supports negatives ("minus") and any size below 10^39 (up to undecillions); pass the number as a decimal string. Set ordinal for ordinal words ("first", "twenty-first").

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |
| `lang` | `str` (default `"en"`) | Language; only "en" is supported. One of: en. |
| `ordinal` | `bool` (default `False`) | Return ordinal words (first, second, ...). |

Returns `str`. Example: `num.number_to_words("1234")`

Keywords: spell out number, number to text, say the number, number, words, spell, out, english, say, text, name, cardinal, ordinal, check, writing

## num.to_base

`num.to_base(value: str, from_base: int = 10, to_base: int = 16) -> str`

Converts an integer from one base to another (2..36 for output, 2..62 accepted for input, case-insensitive up to 36). Handles any size and a leading sign; prefixes 0x/0b/0o, underscores and spaces in the input are ignored.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer text. |
| `from_base` | `int` (default `10`) | Input base 2..62 (0 = auto-detect, a prefix always wins). |
| `to_base` | `int` (default `16`) | Output base 2..36. |

Returns `str`. Example: `num.to_base("ff", 16, 2)`

Keywords: radix, change base, base n, base, convert, change, base36, base32, base3

## num.ieee754_encode

`num.ieee754_encode(value: str, bits: int = 32) -> Dict[str, str]`

Encodes a decimal number as an IEEE 754 binary32 (bits=32) or binary64 (bits=64) float and returns the hex bits, full binary, sign, biased exponent, mantissa, unbiased exponent, class and the exact stored decimal value. Accepts inf and nan.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Decimal number, e.g. 0.1, -2.5e10, inf, nan. |
| `bits` | `int` (default `32`) | 32 (single) or 64 (double). |

Returns `Dict[str, str]`. Example: `num.ieee754_encode("0.1", 32)`

Keywords: float to bits, float binary representation, ieee754, float, floating, point, binary32, binary64, single, double, hex, bits, encode, representation, mantissa, exponent, sign

## num.ordinal_words

`num.ordinal_words(value: str) -> str`

Spells a whole number as English ordinal words (1 gives first, 21 gives twenty-first, 100 gives one hundredth). Any size below 10^39.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.ordinal_words("21")`

Keywords: ordinal, words, first, second, third, spell, out

## num.ieee754_decode

`num.ieee754_decode(hex: str, bits: int = 0) -> str`

Decodes IEEE 754 hex bits (8 hex digits = binary32, 16 = binary64; 0x prefix and spaces ignored) into the decimal value, shortest round-trip form (3dcccccd gives 0.1). Returns NaN, +Inf or -Inf for the special patterns.

| Parameter | Type | Description |
|---|---|---|
| `hex` | `str` | Hex bits, e.g. 3dcccccd or 0x3FB999999999999A. |
| `bits` | `int` (default `0`) | 32 or 64 (0 = from the digit count). |

Returns `str`. Example: `num.ieee754_decode("3dcccccd")`

Keywords: ieee754, float, floating, point, decode, hex, bits, to, number, binary32, binary64

## num.words_to_number

`num.words_to_number(text: str) -> str`

Parses English number words ("two hundred and five", "minus twenty-one", "three million four", ordinals like "first") into a decimal string. Unknown words are errors.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | English number words. |

Returns `str`. Example: `num.words_to_number("two hundred and five")`

Keywords: words to digits, parse spelled number, words, to, number, parse, english, text, spelled, out, digits

## num.float_to_fraction

`num.float_to_fraction(x: str, max_denominator: int = 1000000) -> str`

Finds the closest fraction n/d to a decimal number with denominator at most max_denominator (continued fractions). A finite decimal like 0.125 gives exactly 1/8; 3.14159 with max 1000 gives 355/113. An integer gives just n.

| Parameter | Type | Description |
|---|---|---|
| `x` | `str` | Decimal number or a fraction a/b. |
| `max_denominator` | `int` (default `1000000`) | Largest allowed denominator (>= 1, up to 10^18). |

Returns `str`. Example: `num.float_to_fraction("3.14159", 1000)`

Keywords: fraction, rational, decimal, to, continued, approximate, numerator, denominator

## num.ordinal

`num.ordinal(value: str) -> str`

Adds the English ordinal suffix to an integer: 1st, 2nd, 3rd, 4th, 11th, 12th, 13th, 21st, 101st; negatives keep their sign (-1st).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.ordinal("22")`

Keywords: ordinal, suffix, 1st, 2nd, 3rd, 4th, th, st, nd, rd

## num.add_thousands_separators

`num.add_thousands_separators(value: str, sep: str = ",", group: int = 3) -> str`

Groups the integer part of a plain decimal number into threes with a separator: 1234567.891 becomes 1,234,567.891. Existing commas, spaces and underscores are removed first; the fraction is untouched.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Plain decimal number. |
| `sep` | `str` (default `","`) | Separator text. |
| `group` | `int` (default `3`) | Digits per group. |

Returns `str`. Example: `num.add_thousands_separators("1234567.891")`

Keywords: commas in numbers, format with commas, thousands, separator, comma, group, digits, format, number, commas

## num.nearest_float32

`num.nearest_float32(x: str) -> str`

Rounds a decimal number to the nearest IEEE 754 single-precision float and returns the exact decimal value of that float (0.1 gives 0.100000001490116119384765625). Values beyond float32 range are an error.

| Parameter | Type | Description |
|---|---|---|
| `x` | `str` | Decimal number. |

Returns `str`. Example: `num.nearest_float32("0.1")`

Keywords: float32, single, precision, round, nearest, representable, error

## num.bin_add

`num.bin_add(a: str, b: str) -> str`

Takes two binary numbers (prefix, sign and underscores allowed) and adds them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First binary number. |
| `b` | `str` | Second binary number. |

Returns `str`. Example: `num.bin_add("1010", "11")`

Keywords: binary, calculator, add, arithmetic, plus, sum

## num.bin_div

`num.bin_div(a: str, b: str) -> Dict[str, str]`

Takes two binary numbers and returns a dict with quotient and remainder as binary digits (division truncates toward zero; dividing by zero is an error).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First binary number. |
| `b` | `str` | Second binary number. |

Returns `Dict[str, str]`. Example: `num.bin_div("1010", "11")`

Keywords: binary, calculator, div, arithmetic, quotient, remainder, divide

## num.bin_mul

`num.bin_mul(a: str, b: str) -> str`

Takes two binary numbers (prefix, sign and underscores allowed) and multiplies them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First binary number. |
| `b` | `str` | Second binary number. |

Returns `str`. Example: `num.bin_mul("1010", "11")`

Keywords: binary, calculator, mul, arithmetic, times, product

## num.bin_sub

`num.bin_sub(a: str, b: str) -> str`

Takes two binary numbers (prefix, sign and underscores allowed) and subtracts them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First binary number. |
| `b` | `str` | Second binary number. |

Returns `str`. Example: `num.bin_sub("1010", "11")`

Keywords: binary, calculator, sub, arithmetic, minus, difference

## num.hex_add

`num.hex_add(a: str, b: str) -> str`

Takes two hex numbers (prefix, sign and underscores allowed) and adds them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First hex number. |
| `b` | `str` | Second hex number. |

Returns `str`. Example: `num.hex_add("ff", "1")`

Keywords: hex, calculator, add, arithmetic, plus, sum

## num.hex_div

`num.hex_div(a: str, b: str) -> Dict[str, str]`

Takes two hex numbers and returns a dict with quotient and remainder as hex digits (division truncates toward zero; dividing by zero is an error).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First hex number. |
| `b` | `str` | Second hex number. |

Returns `Dict[str, str]`. Example: `num.hex_div("ff", "1")`

Keywords: hex, calculator, div, arithmetic, quotient, remainder, divide

## num.hex_mul

`num.hex_mul(a: str, b: str) -> str`

Takes two hex numbers (prefix, sign and underscores allowed) and multiplies them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First hex number. |
| `b` | `str` | Second hex number. |

Returns `str`. Example: `num.hex_mul("ff", "1")`

Keywords: hex, calculator, mul, arithmetic, times, product

## num.hex_sub

`num.hex_sub(a: str, b: str) -> str`

Takes two hex numbers (prefix, sign and underscores allowed) and subtracts them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First hex number. |
| `b` | `str` | Second hex number. |

Returns `str`. Example: `num.hex_sub("ff", "1")`

Keywords: hex, calculator, sub, arithmetic, minus, difference

## num.strip_separators

`num.strip_separators(value: str, decimal_point: str = ".") -> str`

Normalises a formatted number to plain digits with a dot decimal point by removing thousands marks (commas, spaces, underscores, apostrophes, non-breaking spaces). With decimal_point="," the dots are removed and the comma becomes the decimal point (1.234,56 gives 1234.56).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Formatted number. |
| `decimal_point` | `str` (default `"."`) | Decimal mark used by the input: "." or ",". One of: ., ,. |

Returns `str`. Example: `num.strip_separators("1,234,567.89")`

Keywords: strip, remove, thousands, separators, commas, clean, number, normalize, parse, european

## num.scientific

`num.scientific(value: str, digits: int = 3) -> str`

Writes a number in scientific notation with `digits` digits after the point (123456789 with digits=3 gives 1.235e+8). Accepts decimals, fractions and exponents; works for huge integers.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Number. |
| `digits` | `int` (default `3`) | Digits after the decimal point (0..200). |

Returns `str`. Example: `num.scientific("123456789", 3)`

Keywords: scientific, notation, exponent, e, standard, form, engineering, mantissa

## num.format_number

`num.format_number(value: str, decimals: int = 2, thousands: str = ",", decimal_point: str = ".") -> str`

Formats a number with fixed decimals (rounded half away from zero), a thousands separator and a decimal point of your choice: 1234.5 becomes 1,234.50, or 1.234,50 with thousands="." and decimal_point=",". Exact for any size.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Number. |
| `decimals` | `int` (default `2`) | Digits after the decimal point (0..1000). |
| `thousands` | `str` (default `","`) | Thousands separator (may be empty). |
| `decimal_point` | `str` (default `"."`) | Decimal mark. |

Returns `str`. Example: `num.format_number("1234567.891", 2)`

Keywords: thousands separator, group digits, pretty print number, format, number, decimals, thousands, separator, decimal, point, round, locale, currency

## num.parse_human_number

`num.parse_human_number(text: str) -> str`

Parses shorthand like 1.5k, 2M, 3e6, 1,234.5, 2.5 billion or 50% into a plain decimal string (k=10^3, M=10^6, B=10^9, T=10^12, %=1/100). Currency symbols are ignored; unknown suffixes are errors.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Shorthand number. |

Returns `str`. Example: `num.parse_human_number("1.5k")`

Keywords: parse, human, number, shorthand, k, m, b, t, million, billion, thousand, abbreviation, expand

## num.humanize_bytes

`num.humanize_bytes(value: str, binary: bool = False, decimals: int = 2) -> str`

Formats a byte count as a readable size: 1536 gives 1.54 KB (decimal, 1000-based) or 1.5 KiB with binary=true (1024-based). Trailing zeros are trimmed; negatives keep the sign.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Number of bytes. |
| `binary` | `bool` (default `False`) | Use 1024-based KiB/MiB units instead of 1000-based KB/MB. |
| `decimals` | `int` (default `2`) | Digits after the decimal point (0..6). |

Returns `str`. Example: `num.humanize_bytes("1536", binary=True)`

Keywords: file size, kb mb gb, human readable size, bytes to megabytes, bytes, humanize, human, readable, size, kb, mb, gb, kib, mib, gib, file, format

## num.parse_bytes

`num.parse_bytes(text: str, binary: bool = False) -> int`

Parses a size like "1.5 GiB", "200MB" or "4k" into a byte count (int). KiB/MiB are 1024-based, KB/MB are 1000-based; bare K/M/G follow the binary flag. Results beyond int64 are an error.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Size text. |
| `binary` | `bool` (default `False`) | Treat KB/MB/K/M as 1024-based. |

Returns `int`. Example: `num.parse_bytes("1.5 GiB")`

Keywords: mb to bytes, parse file size, size string to bytes, parse, bytes, size, gib, mib, kb, mb, convert, to, file

## num.bytes_table

`num.bytes_table(value: str) -> Dict[str, str]`

Shows a byte count in every unit: bit, B, KB, MB, GB, TB, PB (1000-based) and KiB, MiB, GiB, TiB, PiB (1024-based), each rounded to 6 decimals. Accepts a number or a size such as "1.5 GiB".

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Byte count or a size like 1.5GiB. |

Returns `Dict[str, str]`. Example: `num.bytes_table("1073741824")`

Keywords: bytes, table, units, kb, mb, gb, tb, kib, mib, gib, conversion, size, chart

## num.number_to_binary_bytes

`num.number_to_binary_bytes(value: str, size: int = 0, endian: str = "big", base: int = 0) -> List[str]`

Converts an integer to a list of two-digit hex bytes, big-endian by default (258 gives \["01", "02"\]) or little-endian. size=0 uses the minimal width (negatives as two's complement); a fixed size pads or errors if the value does not fit (max 8192 bytes).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `size` | `int` (default `0`) | Width in bytes (0 = minimal). |
| `endian` | `str` (default `"big"`) | big or little. One of: big, little. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `List[str]`. Example: `num.number_to_binary_bytes("258")`

Keywords: bytes, to, endian, little, big, hex, array, serialize, integer

## num.bytes_to_number

`num.bytes_to_number(bytes: List[str], endian: str = "big", signed: bool = False) -> str`

Combines a list of hex bytes ("01", "0xff") into one integer, big-endian by default or little-endian, optionally as signed two's complement. Returns a decimal string.

| Parameter | Type | Description |
|---|---|---|
| `bytes` | `List[str]` | Hex byte strings. |
| `endian` | `str` (default `"big"`) | big or little. One of: big, little. |
| `signed` | `bool` (default `False`) | Interpret as two's complement. |

Returns `str`. Example: `num.bytes_to_number(["01", "02"])`

Keywords: bytes, to, number, endian, little, big, hex, array, integer, deserialize

## num.digits

`num.digits(value: str) -> List[str]`

Splits the absolute value of a decimal integer into its digits (1203 gives \["1", "2", "0", "3"\]).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `List[str]`. Example: `num.digits("1203")`

Keywords: digits, split, digit, list, decimal, places

## num.digit_sum

`num.digit_sum(value: str) -> int`

Adds up the decimal digits of an integer's absolute value (9875 gives 29).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `int`. Example: `num.digit_sum("9875")`

Keywords: digit, sum, add, digits, cross, checksum

## num.from_bin

`num.from_bin(value: str) -> str`

Reads a binary integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | The binary digits. |

Returns `str`. Example: `num.from_bin("1010")`

Keywords: binary to decimal, parse binary, binary, bin, decimal, convert, frombin

## num.from_hex

`num.from_hex(value: str) -> str`

Reads a hexadecimal integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | The hexadecimal digits. |

Returns `str`. Example: `num.from_hex("ff")`

Keywords: hex to decimal, parse hex, hexadecimal, hex, decimal, convert, fromhex

## num.from_oct

`num.from_oct(value: str) -> str`

Reads a octal integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | The octal digits. |

Returns `str`. Example: `num.from_oct("17")`

Keywords: octal, oct, decimal, convert, fromoct

## num.digital_root

`num.digital_root(value: str) -> int`

Repeatedly sums the digits until one digit remains (9875 gives 2); the root of 0 is 0. Uses the absolute value.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `int`. Example: `num.digital_root("9875")`

Keywords: digital, root, repeated, digit, sum, casting, out, nines

## num.reverse_digits

`num.reverse_digits(value: str) -> str`

Reverses the decimal digits of an integer, keeping the sign (1230 gives 321, -45 gives -54).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.reverse_digits("1230")`

Keywords: reverse, digits, flip, number, backwards

## num.is_palindrome_number

`num.is_palindrome_number(value: str, base: int = 10) -> bool`

Tells whether an integer reads the same backwards in the given base (default 10; 585 is a palindrome, and so is 9 in base 2 as 1001). Negative numbers are never palindromes.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `base` | `int` (default `10`) | Base 2..36 used for the digits. |

Returns `bool`. Example: `num.is_palindrome_number("12321")`

Keywords: palindrome, number, symmetric, reads, same, backwards

## num.detect_base

`num.detect_base(value: str) -> str`

Guesses the base of a number text: a 0x/0b/0o prefix gives hex/bin/oct, hex letters (a-f) mean hex, a plain digit run means dec. Returns "hex", "bin", "oct" or "dec", or an error if the text is not a number.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Number text. |

Returns `str`. Example: `num.detect_base("0b1010")`

Keywords: is this hex or binary, guess number base, detect, guess, radix, base, prefix

## num.is_prime

`num.is_prime(value: str) -> bool`

Tests whether an integer is prime. Exact for numbers below 2^64 and probabilistic (Miller-Rabin plus Baillie-PSW, 20 rounds) above; limited to 8192 bits. Numbers below 2 are not prime.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `bool`. Example: `num.is_prime("97")`

Keywords: prime test, primality, prime, test, is, composite, miller, rabin

## num.next_prime

`num.next_prime(value: str) -> str`

Returns the smallest prime strictly greater than the number (next_prime of 13 is 17; of -5 or 0 is 2). Limited to 1024-bit inputs.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.next_prime("100")`

Keywords: next, prime, following, greater

## num.prev_prime

`num.prev_prime(value: str) -> str`

Returns the largest prime strictly smaller than the number (prev_prime of 17 is 13). Numbers 2 and below have no previous prime and give an error. Limited to 1024-bit inputs.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.prev_prime("100")`

Keywords: previous, prime, smaller, lower

## num.prime_factors

`num.prime_factors(value: str) -> List[str]`

Returns the prime factors of an integer &gt;= 2 with multiplicity, ascending (360 gives 2, 2, 2, 3, 3, 5). Uses trial division to 100000 and Pollard's rho; numbers over 512 bits or with a huge prime cofactor are an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `List[str]`. Example: `num.prime_factors("360")`

Keywords: prime factorization, factorise, prime, factors, factorize, factor, decomposition, integer, factorization

## num.factorization

`num.factorization(value: str) -> str`

Writes the prime factorization of an integer &gt;= 1 as a product with exponents: 360 gives "2^3 * 3^2 * 5", 1 gives "1". Same limits as prime_factors.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.factorization("360")`

Keywords: factorization, prime, power, exponent, factor, tree, decomposition

## num.divisors

`num.divisors(value: str) -> List[str]`

Lists all positive divisors of an integer &gt;= 1 in ascending order (12 gives 1, 2, 3, 4, 6, 12). Limited to 512-bit numbers and 100000 divisors.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `List[str]`. Example: `num.divisors("28")`

Keywords: divisors, factors, all, proper, list

## num.is_perfect

`num.is_perfect(value: str) -> bool`

Tests whether a number equals the sum of its proper divisors (6, 28, 496, 8128, ...). Numbers below 2 are not perfect; limited to 128 bits.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `bool`. Example: `num.is_perfect("28")`

Keywords: perfect, number, divisors, sum, aliquot

## num.is_armstrong

`num.is_armstrong(value: str) -> bool`

Tests whether a number equals the sum of its digits each raised to the number of digits (narcissistic number: 153, 370, 9474). Negative numbers are not.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `bool`. Example: `num.is_armstrong("153")`

Keywords: armstrong, narcissistic, number, digits, power

## num.is_fibonacci

`num.is_fibonacci(value: str) -> bool`

Tests whether a non-negative integer is a Fibonacci number (0, 1, 1, 2, 3, 5, 8, ...) using the 5n^2 +/- 4 perfect-square test.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `bool`. Example: `num.is_fibonacci("144")`

Keywords: fibonacci, number, test, is, sequence, member

## num.fibonacci

`num.fibonacci(value: str) -> str`

Returns the n-th Fibonacci number (F0=0, F1=1, F10=55) as a decimal string using fast doubling. n must be 0..100000.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.fibonacci("10")`

Keywords: nth fibonacci number, fibonacci, nth, number, sequence, golden

## num.bit_and

`num.bit_and(a: str, b: str, base: int = 0, output: str = "dec") -> str`

Bitwise AND of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First integer. |
| `b` | `str` | Second integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.bit_and("0xF0", "0x3C", output="hex")`

Keywords: bitwise and, mask bits, and, mask, bitwise, &, binary

## num.bit_or

`num.bit_or(a: str, b: str, base: int = 0, output: str = "dec") -> str`

Bitwise OR of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First integer. |
| `b` | `str` | Second integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.bit_or("0xF0", "0x3C", output="hex")`

Keywords: or, bitwise, |, binary, union

## num.bit_xor

`num.bit_xor(a: str, b: str, base: int = 0, output: str = "dec") -> str`

Bitwise XOR of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First integer. |
| `b` | `str` | Second integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.bit_xor("0xF0", "0x3C", output="hex")`

Keywords: xor, bitwise, ^, binary, exclusive

## num.factorial

`num.factorial(value: str) -> str`

Returns n! as a decimal string (0! = 1, 20! = 2432902008176640000). n must be 0..20000; larger values are an error to keep the result small.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.factorial("20")`

Keywords: n!, product of 1 to n, factorial, product, permutations

## num.bit_not

`num.bit_not(value: str, bits: int = 0, base: int = 0, output: str = "dec") -> str`

Inverts the bits of an integer. With bits=0 returns -value-1 (infinite two's complement); with bits=N inverts within an N-bit field (0 &lt; N &lt;= 65536) and returns the unsigned pattern, zero-padded in hex/bin output.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bits` | `int` (default `0`) | Field width in bits (0 = unbounded). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.bit_not("0x0F", 8, output="bin")`

Keywords: not, invert, complement, ones, bitwise, ~, flip

## num.collatz_steps

`num.collatz_steps(value: str) -> int`

Counts the steps of the 3n+1 (Collatz) sequence from n &gt;= 1 until it reaches 1 (27 takes 111 steps). Gives an error if it exceeds one million steps.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `int`. Example: `num.collatz_steps("27")`

Keywords: collatz, 3n+1, steps, hailstone, sequence

## num.euler_phi

`num.euler_phi(value: str) -> str`

Computes Euler's totient phi(n): how many numbers in 1..n are coprime to n (phi(36) = 12, phi(1) = 1). Uses the factorization, so it shares prime_factors' limits.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.euler_phi("36")`

Keywords: euler, phi, totient, coprime, count

## num.shift_left

`num.shift_left(value: str, n: int = 1, base: int = 0, output: str = "dec") -> str`

Shifts a big integer left by n bits (multiplies by 2^n). n must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `n` | `int` (default `1`) | Number of bit positions. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.shift_left("0b1010", 2, output="bin")`

Keywords: bit shift, multiply by power of two, shift, left, bitshift, &lt;&lt;, &gt;&gt;, multiply, divide, power, of, two

## num.shift_right

`num.shift_right(value: str, n: int = 1, base: int = 0, output: str = "dec") -> str`

Shifts a big integer right by n bits (arithmetic shift, floors toward negative infinity). n must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `n` | `int` (default `1`) | Number of bit positions. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.shift_right("0b1010", 2, output="bin")`

Keywords: shift, right, bitshift, &lt;&lt;, &gt;&gt;, multiply, divide, power, of, two

## num.isqrt

`num.isqrt(value: str) -> str`

Returns floor(sqrt(n)) exactly for any size (isqrt of 99 is 9). Negative input is an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.isqrt("99")`

Keywords: integer, square, root, isqrt, floor, sqrt

## num.rotate_left

`num.rotate_left(value: str, n: int = 1, bits: int = 32, base: int = 0, output: str = "dec") -> str`

Rotates the low `bits` bits of value left by n positions (wrapping around) and returns the unsigned pattern, zero-padded in hex/bin output. Negative values are taken as two's complement; n is reduced modulo bits.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `n` | `int` (default `1`) | Positions to rotate. |
| `bits` | `int` (default `32`) | Field width in bits (1..65536). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.rotate_left("0x81", 1, 8, output="hex")`

Keywords: rotate, rol, ror, circular, shift, bits, left

## num.rotate_right

`num.rotate_right(value: str, n: int = 1, bits: int = 32, base: int = 0, output: str = "dec") -> str`

Rotates the low `bits` bits of value right by n positions (wrapping around) and returns the unsigned pattern, zero-padded in hex/bin output. Negative values are taken as two's complement; n is reduced modulo bits.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `n` | `int` (default `1`) | Positions to rotate. |
| `bits` | `int` (default `32`) | Field width in bits (1..65536). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.rotate_right("0x81", 1, 8, output="hex")`

Keywords: rotate, rol, ror, circular, shift, bits, right

## num.clear_bit

`num.clear_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str`

Sets one bit to 0. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `index` | `int` (default `0`) | Bit position, 0 = least significant. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.clear_bit("0b1000", 1)`

Keywords: bit, flag, clear, mask, index, position

## num.icbrt

`num.icbrt(value: str) -> str`

Returns the cube root truncated toward zero (icbrt of 30 is 3, of -30 is -3), exact for any size.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |

Returns `str`. Example: `num.icbrt("1000")`

Keywords: integer, cube, root, icbrt, floor, cbrt

## num.set_bit

`num.set_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str`

Sets one bit to 1. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `index` | `int` (default `0`) | Bit position, 0 = least significant. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.set_bit("0b1000", 1)`

Keywords: bit, flag, set, mask, index, position

## num.test_bit

`num.test_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> bool`

Returns true if the bit is 1. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `index` | `int` (default `0`) | Bit position, 0 = least significant. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `bool`. Example: `num.test_bit("0b1000", 1)`

Keywords: bit, flag, test, mask, index, position

## num.toggle_bit

`num.toggle_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str`

Flips one bit. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `index` | `int` (default `0`) | Bit position, 0 = least significant. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.toggle_bit("0b1000", 1)`

Keywords: bit, flag, toggle, mask, index, position

## num.bit_length

`num.bit_length(value: str, base: int = 0) -> int`

Returns the number of bits needed to represent the absolute value (0 for 0, 8 for 255, 9 for 256).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `int`. Example: `num.bit_length("255")`

Keywords: bits, length, size, how, many, bitlen, width

## num.nth_root

`num.nth_root(value: str, n: int = 2) -> str`

Returns the integer part of the n-th root of a number (nth_root(1000000, 6) is 10), exact for any size. n must be 1..4096; negative values need an odd n and truncate toward zero.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer, decimal string. |
| `n` | `int` (default `2`) | Root degree 1..4096. |

Returns `str`. Example: `num.nth_root("1024", 10)`

Keywords: nth, root, integer, radical, floor, exact

## num.gcd

`num.gcd(a: str, b: str) -> str`

Greatest common divisor of two integers of any size (gcd(12, 18) = 6; gcd with 0 returns the other absolute value; signs are ignored).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First integer. |
| `b` | `str` | Second integer. |

Returns `str`. Example: `num.gcd("12", "18")`

Keywords: gcf, highest common factor, gcd, greatest, common, divisor, hcf, euclid

## num.popcount

`num.popcount(value: str, bits: int = 0, base: int = 0) -> int`

Counts the 1 bits (Hamming weight). With bits=0 a negative number counts the bits of its absolute value; with bits=N the value is first reduced to an N-bit two's complement pattern.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bits` | `int` (default `0`) | Field width (0 = none). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `int`. Example: `num.popcount("0xFF")`

Keywords: popcount, hamming, weight, count, ones, set, bits, bitcount

## num.lcm

`num.lcm(a: str, b: str) -> str`

Least common multiple of two integers of any size (lcm(4, 6) = 12; returns 0 if either is 0).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First integer. |
| `b` | `str` | Second integer. |

Returns `str`. Example: `num.lcm("4", "6")`

Keywords: lowest common multiple, lcm, least, common, multiple

## num.parity

`num.parity(value: str, base: int = 0) -> int`

Returns 1 if the number of set bits (of the absolute value) is odd, else 0.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `int`. Example: `num.parity("7")`

Keywords: parity, odd, even, bits, xor, bit

## num.gcd_list

`num.gcd_list(values: List[str]) -> str`

Folds a list of integer strings with gcd. The list must not be empty.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[str]` | Integers as decimal strings. |

Returns `str`. Example: `num.gcd_list(["12", "18", "30"])`

Keywords: gcd, lcm, list, many, numbers

## num.lcm_list

`num.lcm_list(values: List[str]) -> str`

Folds a list of integer strings with lcm. The list must not be empty.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[str]` | Integers as decimal strings. |

Returns `str`. Example: `num.lcm_list(["12", "18", "30"])`

Keywords: gcd, lcm, list, many, numbers

## num.leading_zeros

`num.leading_zeros(value: str, bits: int = 32, base: int = 0) -> int`

Counts the zero bits above the highest 1 inside a `bits`-wide field. A value that does not fit the field is an error; negatives use two's complement (so 0 leading zeros).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bits` | `int` (default `32`) | Field width 1..65536. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `int`. Example: `num.leading_zeros("1", 32)`

Keywords: leading, zeros, clz, count, bits, msb

## num.modpow

`num.modpow(base: str, exp: str, mod: str) -> str`

Computes base^exp mod m with arbitrary precision (modpow(4, 13, 497) = 445). m must be positive; a negative exp needs base invertible mod m.

| Parameter | Type | Description |
|---|---|---|
| `base` | `str` | Base. |
| `exp` | `str` | Exponent. |
| `mod` | `str` | Modulus, > 0. |

Returns `str`. Example: `num.modpow("4", "13", "497")`

Keywords: modpow, modular, exponentiation, pow, mod, power, modulo, rsa

## num.trailing_zeros

`num.trailing_zeros(value: str, base: int = 0) -> int`

Counts the zero bits below the lowest 1 (ctz). Returns 0 for the value 0; the sign is ignored.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `int`. Example: `num.trailing_zeros("8")`

Keywords: trailing, zeros, ctz, count, bits, lsb

## num.modinv

`num.modinv(a: str, m: str) -> str`

Finds x with a*x = 1 (mod m), 0 &lt;= x &lt; m (modinv(3, 11) = 4). Errors when gcd(a, m) is not 1 or m &lt; 2.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | Number to invert. |
| `m` | `str` | Modulus, >= 2. |

Returns `str`. Example: `num.modinv("3", "11")`

Keywords: modular, inverse, modinv, multiplicative, extended, euclid

## num.reverse_bits

`num.reverse_bits(value: str, bits: int = 32, base: int = 0, output: str = "dec") -> str`

Reverses the order of the low `bits` bits (bit 0 becomes bit bits-1). Negatives use two's complement; hex/bin output is zero-padded to the width.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bits` | `int` (default `32`) | Field width 1..65536. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.reverse_bits("0b0001", 8, output="bin")`

Keywords: reverse, bits, mirror, flip, order, bit, reversal

## num.swap_endian

`num.swap_endian(value: str, bytes: int = 4, base: int = 0, output: str = "dec") -> str`

Reverses the byte order of the low `bytes` bytes (big-endian &lt;-&gt; little-endian), e.g. 0x12345678 becomes 0x78563412 for 4 bytes. Negatives use two's complement; hex output is zero-padded.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Integer. |
| `bytes` | `int` (default `4`) | Width in bytes (1..8192). |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.swap_endian("0x12345678", 4, output="hex")`

Keywords: byte order, big endian little endian, byteswap, endian, byte, swap, bswap, htonl, ntohl, little, big, order

## num.to_twos

`num.to_twos(value: str, bits: int = 32, base: int = 0, output: str = "bin") -> str`

Converts a signed integer to its two's complement bit pattern in a `bits`-wide field (e.g. -1 in 8 bits is 0xFF). The value must satisfy -2^(bits-1) &lt;= value &lt; 2^bits. Hex/bin output is zero-padded to the width.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Signed integer. |
| `bits` | `int` (default `32`) | Field width 1..65536. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"bin"`) | Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.to_twos("-5", 8, output="bin")`

Keywords: twos, complement, signed, negative, representation, binary, two's, encode

## num.from_twos

`num.from_twos(value: str, bits: int = 32, base: int = 0) -> str`

Interprets a bit pattern as a `bits`-wide two's complement number and returns the signed decimal value (0xFF in 8 bits is -1). The pattern must be non-negative and fit the width.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Bit pattern, e.g. 0xFB or 0b11111011. |
| `bits` | `int` (default `32`) | Field width 1..65536. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |

Returns `str`. Example: `num.from_twos("0xFB", 8)`

Keywords: twos, complement, signed, decode, negative, interpret, int8, int16, int32, int64

## num.sign_extend

`num.sign_extend(value: str, from_bits: int = 8, to_bits: int = 32, base: int = 0, output: str = "hex") -> str`

Sign-extends a from_bits-wide two's complement pattern to to_bits and returns the new unsigned pattern (0x80 from 8 to 16 bits is 0xFF80). to_bits must be &gt;= from_bits.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Bit pattern. |
| `from_bits` | `int` (default `8`) | Source width 1..65536. |
| `to_bits` | `int` (default `32`) | Target width 1..65536. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"hex"`) | Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.sign_extend("0x80", 8, 16, output="hex")`

Keywords: sign, extend, sext, widen, bits, two's, complement

## num.mask

`num.mask(bits: int, shift: int = 0, output: str = "hex") -> str`

Builds a mask of `bits` consecutive 1 bits, optionally shifted left (mask(8) = 0xFF, mask(4, 4) = 0xF0). bits 0..65536, shift 0..1048576.

| Parameter | Type | Description |
|---|---|---|
| `bits` | `int` | Number of 1 bits. |
| `shift` | `int` (default `0`) | Left shift applied to the mask. |
| `output` | `str` (default `"hex"`) | Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.mask(8, 4, output="hex")`

Keywords: mask, bitmask, ones, field, width, generate

## num.gray_encode

`num.gray_encode(value: str, base: int = 0, output: str = "dec") -> str`

Converts a non-negative integer to its reflected binary Gray code (value XOR value&gt;&gt;1), so 2 becomes 3 and 7 becomes 4.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Non-negative integer. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.gray_encode("7")`

Keywords: gray, code, reflected, binary, encode

## num.gray_decode

`num.gray_decode(value: str, base: int = 0, output: str = "dec") -> str`

Converts a Gray code back to the ordinary binary integer (inverse of gray_encode), so 4 becomes 7.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Non-negative Gray code. |
| `base` | `int` (default `0`) | Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal. |
| `output` | `str` (default `"dec"`) | Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct. |

Returns `str`. Example: `num.gray_decode("4")`

Keywords: gray, code, reflected, binary, decode

## num.luhn_check

`num.luhn_check(value: str) -> bool`

Validates a number with the Luhn (mod 10) algorithm used by credit cards, IMEI and many ids. Spaces and dashes are ignored; at least 2 digits are required; non-digits are an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.luhn_check("4539 1488 0343 6467")`

Keywords: credit card check, luhn validate, luhn, mod, 10, credit, card, validate, checksum, imei

## num.luhn_generate

`num.luhn_generate(value: str) -> str`

Computes the single Luhn check digit to append to a number (7992739871 gives 3). Use luhn_complete to get the full number.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits without the check digit; spaces, dashes and dots are ignored. |

Returns `str`. Example: `num.luhn_generate("7992739871")`

Keywords: luhn, check, digit, generate, credit, card, checksum

## num.luhn_complete

`num.luhn_complete(value: str) -> str`

Appends the Luhn check digit to a number, giving a valid Luhn number (7992739871 gives 79927398713).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits without the check digit; spaces, dashes and dots are ignored. |

Returns `str`. Example: `num.luhn_complete("7992739871")`

Keywords: luhn, complete, append, check, digit, credit, card

## num.isbn10_check

`num.isbn10_check(value: str) -> bool`

Validates an ISBN-10 (weights 10..1, mod 11; the last character may be X). Hyphens and spaces are ignored.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.isbn10_check("0-306-40615-2")`

Keywords: isbn, isbn10, book, validate, checksum

## num.isbn13_check

`num.isbn13_check(value: str) -> bool`

Validates an ISBN-13 (EAN-13 check digit, must start with 978 or 979). Hyphens and spaces are ignored.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.isbn13_check("978-3-16-148410-0")`

Keywords: book number check, isbn, isbn13, book, validate, checksum, ean

## num.isbn10_to_13

`num.isbn10_to_13(value: str) -> str`

Converts a valid ISBN-10 to ISBN-13 by prefixing 978 and recomputing the check digit (0306406152 gives 9780306406157). Invalid input is an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | ISBN-10. |

Returns `str`. Example: `num.isbn10_to_13("0306406152")`

Keywords: isbn, convert, 10, to, 13, isbn13

## num.isbn13_to_10

`num.isbn13_to_10(value: str) -> str`

Converts a valid 978-prefixed ISBN-13 to ISBN-10 (9780306406157 gives 0306406152). 979 numbers have no ISBN-10 and give an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | ISBN-13. |

Returns `str`. Example: `num.isbn13_to_10("9780306406157")`

Keywords: isbn, convert, 13, to, 10, isbn10

## num.iban_check

`num.iban_check(value: str) -> bool`

Validates an IBAN: structure, the country's fixed length (for ~75 known countries) and the ISO 7064 mod-97 check. Spaces and dashes are ignored, case-insensitive. Returns false for a wrong checksum and an error for illegal characters.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | IBAN. |

Returns `bool`. Example: `num.iban_check("GB82 WEST 1234 5698 7654 32")`

Keywords: bank account number check, iban, bank, account, validate, mod, 97, sepa, checksum

## num.iban_format

`num.iban_format(value: str) -> str`

Formats an IBAN in groups of four characters, uppercase (GB82WEST12345698765432 gives "GB82 WEST 1234 5698 7654 32"). Does not validate the checksum.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | IBAN, with or without spaces. |

Returns `str`. Example: `num.iban_format("GB82WEST12345698765432")`

Keywords: iban, format, group, four, spaces, bank, account

## num.ean13_check

`num.ean13_check(value: str) -> bool`

Validates an EAN-13 barcode (13 digits, weights 1 and 3, mod 10).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.ean13_check("4006381333931")`

Keywords: ean, ean13, barcode, gtin, validate, checksum

## num.ean8_check

`num.ean8_check(value: str) -> bool`

Validates an EAN-8 barcode (8 digits, weights 3 and 1, mod 10).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.ean8_check("96385074")`

Keywords: ean, ean8, barcode, gtin, validate, checksum

## num.upc_check

`num.upc_check(value: str) -> bool`

Validates a UPC-A barcode (12 digits, mod 10 check digit).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.upc_check("036000291452")`

Keywords: upc, upca, barcode, gtin, validate, checksum

## num.gtin_check

`num.gtin_check(value: str) -> bool`

Validates any GTIN barcode number: GTIN-8, UPC-A (12), EAN-13 or GTIN-14, using the mod 10 check digit.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.gtin_check("4006381333931")`

Keywords: gtin, barcode, ean, upc, validate, checksum

## num.gtin_generate

`num.gtin_generate(value: str) -> str`

Computes the barcode check digit for 7, 11, 12 or 13 digits (GTIN-8, UPC-A, EAN-13, GTIN-14); 400638133393 gives 1.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits without the check digit; spaces, dashes and dots are ignored. |

Returns `str`. Example: `num.gtin_generate("400638133393")`

Keywords: gtin, ean, upc, barcode, check, digit, generate

## num.verhoeff_check

`num.verhoeff_check(value: str) -> bool`

Validates a number with the Verhoeff dihedral-group checksum (used by Aadhaar); detects all single-digit and adjacent-transposition errors.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.verhoeff_check("2363")`

Keywords: verhoeff, checksum, aadhaar, validate, dihedral

## num.verhoeff_generate

`num.verhoeff_generate(value: str) -> str`

Computes the Verhoeff check digit to append to a number (236 gives 3).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits without the check digit; spaces, dashes and dots are ignored. |

Returns `str`. Example: `num.verhoeff_generate("236")`

Keywords: verhoeff, check, digit, generate, aadhaar

## num.damm_check

`num.damm_check(value: str) -> bool`

Validates a number with the Damm algorithm (quasigroup check digit; catches all single-digit and adjacent-transposition errors).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits; spaces, dashes and dots are ignored. |

Returns `bool`. Example: `num.damm_check("5724")`

Keywords: damm, checksum, validate, quasigroup

## num.damm_generate

`num.damm_generate(value: str) -> str`

Computes the Damm check digit to append to a number (572 gives 4).

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Digits without the check digit; spaces, dashes and dots are ignored. |

Returns `str`. Example: `num.damm_generate("572")`

Keywords: damm, check, digit, generate

