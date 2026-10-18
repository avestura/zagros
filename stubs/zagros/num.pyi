"""Numbers and bases: Base conversion, the hex and binary calculator, bit operations, IEEE 754, roman numerals, number words, primes and checksums.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `num` is predeclared in main.star; call `num.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: number, base conversion, hex, binary, octal, decimal, roman numerals, number to words, prime, factorial, gcd, bit operations, ieee754, checksum digit, luhn, iban, isbn, human bytes
"""

from typing import Any, Dict, List

def base_all(value: str, from_base: int = 0) -> Dict[str, str]:
    """Shows one integer in binary, octal, decimal, hex, base32, base36 and base62 at once, plus the detected input base. Auto-detects a 0x/0b/0o prefix or hex letters; pass from_base to force a base. Works on arbitrarily large integers (strings in, strings out).

        value: Integer text, e.g. 255, 0xFF, 0b1010, -0o17, 1_000.
        from_base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.base_all("0xFF")
    Keywords: base, convert, converter, bin, oct, dec, hex, binary, octal, decimal, hexadecimal, radix
    Needs: none (pure function).
    """

def bits_view(value: str, bits: int = 0, base: int = 0) -> str:
    """Draws an integer as nibble-grouped binary with the bit index of each group's most significant bit above it. bits=0 picks 8, 16, 32 ... automatically (negatives use two's complement).

        value: Integer.
        bits: Field width 1..65536 (0 = automatic).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.bits_view("0xA5")
    Keywords: bits, binary, view, nibble, grouped, positions, bit, index, layout, register
    Needs: none (pure function).
    """

def calc_int(expr: str, base: int = 0, output: str = "dec") -> str:
    """Evaluates an integer expression with arbitrary precision and returns the result as a decimal string (or hex/bin/oct with output). Operators + - * / % ** & | ^ ~ << >> and parentheses; / truncates toward zero, ** is right-associative and size-capped, ^ is XOR. Literals may be hex 0xFF, binary 0b1010, octal 0o17 or decimal with underscores; functions abs, min, max, gcd, lcm, pow(a,b[,mod]), bits, popcount, isqrt, fact. Errors carry the character position.

        expr: Integer expression: literals (255, 0xFF, 0b1010, 0o17, 1_000), + - * / % ** & | ^ ~ << >> and parentheses.
        base: Base of unprefixed literals, 2..36 (0 = decimal). With base 16, write FF instead of 0xFF.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.calc_int("0xFF + 1")
    Keywords: hex calculator, binary calculator, programmer calculator, calculator, hex, binary, programmer, calc, bigint, evaluate, expression, 0xff, arithmetic
    Needs: none (pure function).
    """

def to_hex(value: str, from_base: int = 0, prefix: bool = False, uppercase: bool = False, group: int = 0) -> str:
    """Converts an integer to hexadecimal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

        value: Integer text (any prefix; decimal unless from_base or a prefix says otherwise).
        from_base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        prefix: Add a 0x prefix.
        uppercase: Use A-F instead of a-f.
        group: Insert a space every N digits (0 = none).

    Example: num.to_hex("255")
    Task: convert a number to hex
    Keywords: decimal to hex, hexadecimal, base 16, hex, base16, convert, decimal, tohex
    Needs: none (pure function).
    """

def calc_all(expr: str, base: int = 0) -> Dict[str, str]:
    """Evaluates an integer expression (same syntax as calc_int) and shows the result in dec, hex, bin and oct, with its bit length and the signed int8/16/32/64 and unsigned uint8/16/32/64 interpretations of its low bits.

        expr: Integer expression: literals (255, 0xFF, 0b1010, 0o17, 1_000), + - * / % ** & | ^ ~ << >> and parentheses.
        base: Base of unprefixed literals, 2..36 (0 = decimal). With base 16, write FF instead of 0xFF.

    Example: num.calc_all("0b1010 << 2")
    Keywords: calculator, hex, bin, dec, oct, programmer, signed, unsigned, int32, int64, twos, complement, evaluate
    Needs: none (pure function).
    """

def to_bin(value: str, from_base: int = 0, prefix: bool = False, group: int = 0) -> str:
    """Converts an integer to binary. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

        value: Integer text (any prefix; decimal unless from_base or a prefix says otherwise).
        from_base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        prefix: Add a 0b prefix.
        group: Insert a space every N digits (0 = none, 4 = nibbles).

    Example: num.to_bin("255")
    Task: convert a number to binary
    Keywords: decimal to binary, base 2, bits, bin, binary, base2, convert, decimal, tobin
    Needs: none (pure function).
    """

def roman_encode(value: str) -> str:
    """Converts a whole number 1..3999 to a Roman numeral using standard subtractive notation (1994 gives MCMXCIV). Numbers outside 1..3999 are an error because overline (vinculum) notation is not supported.

        value: Decimal integer 1..3999.

    Example: num.roman_encode("1994")
    Task: convert to roman numerals
    Keywords: integer to roman, number to roman numeral, roman, numeral, encode, number, to, I, V, X, L, C, M
    Needs: none (pure function).
    """

def to_oct(value: str, from_base: int = 0, prefix: bool = False) -> str:
    """Converts an integer to octal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

        value: Integer text (any prefix; decimal unless from_base or a prefix says otherwise).
        from_base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        prefix: Add a 0o prefix.

    Example: num.to_oct("255")
    Keywords: decimal to octal, base 8, oct, octal, base8, convert, decimal, tooct
    Needs: none (pure function).
    """

def roman_decode(value: str) -> str:
    """Converts a Roman numeral (case-insensitive, subtractive notation) to a decimal string; MCMXCIV gives 1994. Invalid letters or values outside 1..3999 are errors.

        value: Roman numeral.

    Example: num.roman_decode("MCMXCIV")
    Keywords: roman numeral to integer, parse roman numeral, roman, numeral, decode, to, number, I, V, X, L, C, M
    Needs: none (pure function).
    """

def to_dec(value: str, from_base: int = 0) -> str:
    """Converts an integer to decimal. Accepts 0x/0b/0o prefixes, underscores and a sign; any size. Negative numbers keep a minus sign (use to_twos for two's complement).

        value: Integer text (any prefix; decimal unless from_base or a prefix says otherwise).
        from_base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.to_dec("255")
    Keywords: dec, decimal, base10, convert, hex, binary, octal, todec
    Needs: none (pure function).
    """

def number_to_words(value: str, lang: str = "en", ordinal: bool = False) -> str:
    """Spells a whole number as English words (1234 gives "one thousand two hundred thirty-four"). Supports negatives ("minus") and any size below 10^39 (up to undecillions); pass the number as a decimal string. Set ordinal for ordinal words ("first", "twenty-first").

        value: Integer, decimal string.
        lang: Language; only "en" is supported. One of: en.
        ordinal: Return ordinal words (first, second, ...).

    Example: num.number_to_words("1234")
    Task: write a number in words
    Keywords: spell out number, number to text, say the number, number, words, spell, out, english, say, text, name, cardinal, ordinal, check, writing
    Needs: none (pure function).
    """

def to_base(value: str, from_base: int = 10, to_base: int = 16) -> str:
    """Converts an integer from one base to another (2..36 for output, 2..62 accepted for input, case-insensitive up to 36). Handles any size and a leading sign; prefixes 0x/0b/0o, underscores and spaces in the input are ignored.

        value: Integer text.
        from_base: Input base 2..62 (0 = auto-detect, a prefix always wins).
        to_base: Output base 2..36.

    Example: num.to_base("ff", 16, 2)
    Task: convert a number between bases
    Keywords: radix, change base, base n, base, convert, change, base36, base32, base3
    Needs: none (pure function).
    """

def ieee754_encode(value: str, bits: int = 32) -> Dict[str, str]:
    """Encodes a decimal number as an IEEE 754 binary32 (bits=32) or binary64 (bits=64) float and returns the hex bits, full binary, sign, biased exponent, mantissa, unbiased exponent, class and the exact stored decimal value. Accepts inf and nan.

        value: Decimal number, e.g. 0.1, -2.5e10, inf, nan.
        bits: 32 (single) or 64 (double).

    Example: num.ieee754_encode("0.1", 32)
    Keywords: float to bits, float binary representation, ieee754, float, floating, point, binary32, binary64, single, double, hex, bits, encode, representation, mantissa, exponent, sign
    Needs: none (pure function).
    """

def ordinal_words(value: str) -> str:
    """Spells a whole number as English ordinal words (1 gives first, 21 gives twenty-first, 100 gives one hundredth). Any size below 10^39.

        value: Integer, decimal string.

    Example: num.ordinal_words("21")
    Keywords: ordinal, words, first, second, third, spell, out
    Needs: none (pure function).
    """

def ieee754_decode(hex: str, bits: int = 0) -> str:
    """Decodes IEEE 754 hex bits (8 hex digits = binary32, 16 = binary64; 0x prefix and spaces ignored) into the decimal value, shortest round-trip form (3dcccccd gives 0.1). Returns NaN, +Inf or -Inf for the special patterns.

        hex: Hex bits, e.g. 3dcccccd or 0x3FB999999999999A.
        bits: 32 or 64 (0 = from the digit count).

    Example: num.ieee754_decode("3dcccccd")
    Keywords: ieee754, float, floating, point, decode, hex, bits, to, number, binary32, binary64
    Needs: none (pure function).
    """

def words_to_number(text: str) -> str:
    """Parses English number words ("two hundred and five", "minus twenty-one", "three million four", ordinals like "first") into a decimal string. Unknown words are errors.

        text: English number words.

    Example: num.words_to_number("two hundred and five")
    Keywords: words to digits, parse spelled number, words, to, number, parse, english, text, spelled, out, digits
    Needs: none (pure function).
    """

def float_to_fraction(x: str, max_denominator: int = 1000000) -> str:
    """Finds the closest fraction n/d to a decimal number with denominator at most max_denominator (continued fractions). A finite decimal like 0.125 gives exactly 1/8; 3.14159 with max 1000 gives 355/113. An integer gives just n.

        x: Decimal number or a fraction a/b.
        max_denominator: Largest allowed denominator (>= 1, up to 10^18).

    Example: num.float_to_fraction("3.14159", 1000)
    Keywords: fraction, rational, decimal, to, continued, approximate, numerator, denominator
    Needs: none (pure function).
    """

def ordinal(value: str) -> str:
    """Adds the English ordinal suffix to an integer: 1st, 2nd, 3rd, 4th, 11th, 12th, 13th, 21st, 101st; negatives keep their sign (-1st).

        value: Integer, decimal string.

    Example: num.ordinal("22")
    Keywords: ordinal, suffix, 1st, 2nd, 3rd, 4th, th, st, nd, rd
    Needs: none (pure function).
    """

def add_thousands_separators(value: str, sep: str = ",", group: int = 3) -> str:
    """Groups the integer part of a plain decimal number into threes with a separator: 1234567.891 becomes 1,234,567.891. Existing commas, spaces and underscores are removed first; the fraction is untouched.

        value: Plain decimal number.
        sep: Separator text.
        group: Digits per group.

    Example: num.add_thousands_separators("1234567.891")
    Keywords: commas in numbers, format with commas, thousands, separator, comma, group, digits, format, number, commas
    Needs: none (pure function).
    """

def nearest_float32(x: str) -> str:
    """Rounds a decimal number to the nearest IEEE 754 single-precision float and returns the exact decimal value of that float (0.1 gives 0.100000001490116119384765625). Values beyond float32 range are an error.

        x: Decimal number.

    Example: num.nearest_float32("0.1")
    Keywords: float32, single, precision, round, nearest, representable, error
    Needs: none (pure function).
    """

def bin_add(a: str, b: str) -> str:
    """Takes two binary numbers (prefix, sign and underscores allowed) and adds them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

        a: First binary number.
        b: Second binary number.

    Example: num.bin_add("1010", "11")
    Keywords: binary, calculator, add, arithmetic, plus, sum
    Needs: none (pure function).
    """

def bin_div(a: str, b: str) -> Dict[str, str]:
    """Takes two binary numbers and returns a dict with quotient and remainder as binary digits (division truncates toward zero; dividing by zero is an error).

        a: First binary number.
        b: Second binary number.

    Example: num.bin_div("1010", "11")
    Keywords: binary, calculator, div, arithmetic, quotient, remainder, divide
    Needs: none (pure function).
    """

def bin_mul(a: str, b: str) -> str:
    """Takes two binary numbers (prefix, sign and underscores allowed) and multiplies them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

        a: First binary number.
        b: Second binary number.

    Example: num.bin_mul("1010", "11")
    Keywords: binary, calculator, mul, arithmetic, times, product
    Needs: none (pure function).
    """

def bin_sub(a: str, b: str) -> str:
    """Takes two binary numbers (prefix, sign and underscores allowed) and subtracts them with arbitrary precision, returning binary digits without prefix (negative results keep a minus sign).

        a: First binary number.
        b: Second binary number.

    Example: num.bin_sub("1010", "11")
    Keywords: binary, calculator, sub, arithmetic, minus, difference
    Needs: none (pure function).
    """

def hex_add(a: str, b: str) -> str:
    """Takes two hex numbers (prefix, sign and underscores allowed) and adds them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

        a: First hex number.
        b: Second hex number.

    Example: num.hex_add("ff", "1")
    Keywords: hex, calculator, add, arithmetic, plus, sum
    Needs: none (pure function).
    """

def hex_div(a: str, b: str) -> Dict[str, str]:
    """Takes two hex numbers and returns a dict with quotient and remainder as hex digits (division truncates toward zero; dividing by zero is an error).

        a: First hex number.
        b: Second hex number.

    Example: num.hex_div("ff", "1")
    Keywords: hex, calculator, div, arithmetic, quotient, remainder, divide
    Needs: none (pure function).
    """

def hex_mul(a: str, b: str) -> str:
    """Takes two hex numbers (prefix, sign and underscores allowed) and multiplies them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

        a: First hex number.
        b: Second hex number.

    Example: num.hex_mul("ff", "1")
    Keywords: hex, calculator, mul, arithmetic, times, product
    Needs: none (pure function).
    """

def hex_sub(a: str, b: str) -> str:
    """Takes two hex numbers (prefix, sign and underscores allowed) and subtracts them with arbitrary precision, returning hex digits without prefix (negative results keep a minus sign).

        a: First hex number.
        b: Second hex number.

    Example: num.hex_sub("ff", "1")
    Keywords: hex, calculator, sub, arithmetic, minus, difference
    Needs: none (pure function).
    """

def strip_separators(value: str, decimal_point: str = ".") -> str:
    """Normalises a formatted number to plain digits with a dot decimal point by removing thousands marks (commas, spaces, underscores, apostrophes, non-breaking spaces). With decimal_point="," the dots are removed and the comma becomes the decimal point (1.234,56 gives 1234.56).

        value: Formatted number.
        decimal_point: Decimal mark used by the input: "." or ",". One of: ., ,.

    Example: num.strip_separators("1,234,567.89")
    Keywords: strip, remove, thousands, separators, commas, clean, number, normalize, parse, european
    Needs: none (pure function).
    """

def scientific(value: str, digits: int = 3) -> str:
    """Writes a number in scientific notation with `digits` digits after the point (123456789 with digits=3 gives 1.235e+8). Accepts decimals, fractions and exponents; works for huge integers.

        value: Number.
        digits: Digits after the decimal point (0..200).

    Example: num.scientific("123456789", 3)
    Keywords: scientific, notation, exponent, e, standard, form, engineering, mantissa
    Needs: none (pure function).
    """

def format_number(value: str, decimals: int = 2, thousands: str = ",", decimal_point: str = ".") -> str:
    """Formats a number with fixed decimals (rounded half away from zero), a thousands separator and a decimal point of your choice: 1234.5 becomes 1,234.50, or 1.234,50 with thousands="." and decimal_point=",". Exact for any size.

        value: Number.
        decimals: Digits after the decimal point (0..1000).
        thousands: Thousands separator (may be empty).
        decimal_point: Decimal mark.

    Example: num.format_number("1234567.891", 2)
    Keywords: thousands separator, group digits, pretty print number, format, number, decimals, thousands, separator, decimal, point, round, locale, currency
    Needs: none (pure function).
    """

def parse_human_number(text: str) -> str:
    """Parses shorthand like 1.5k, 2M, 3e6, 1,234.5, 2.5 billion or 50% into a plain decimal string (k=10^3, M=10^6, B=10^9, T=10^12, %=1/100). Currency symbols are ignored; unknown suffixes are errors.

        text: Shorthand number.

    Example: num.parse_human_number("1.5k")
    Keywords: parse, human, number, shorthand, k, m, b, t, million, billion, thousand, abbreviation, expand
    Needs: none (pure function).
    """

def humanize_bytes(value: str, binary: bool = False, decimals: int = 2) -> str:
    """Formats a byte count as a readable size: 1536 gives 1.54 KB (decimal, 1000-based) or 1.5 KiB with binary=true (1024-based). Trailing zeros are trimmed; negatives keep the sign.

        value: Number of bytes.
        binary: Use 1024-based KiB/MiB units instead of 1000-based KB/MB.
        decimals: Digits after the decimal point (0..6).

    Example: num.humanize_bytes("1536", binary=True)
    Task: format a byte count
    Keywords: file size, kb mb gb, human readable size, bytes to megabytes, bytes, humanize, human, readable, size, kb, mb, gb, kib, mib, gib, file, format
    Needs: none (pure function).
    """

def parse_bytes(text: str, binary: bool = False) -> int:
    """Parses a size like "1.5 GiB", "200MB" or "4k" into a byte count (int). KiB/MiB are 1024-based, KB/MB are 1000-based; bare K/M/G follow the binary flag. Results beyond int64 are an error.

        text: Size text.
        binary: Treat KB/MB/K/M as 1024-based.

    Example: num.parse_bytes("1.5 GiB")
    Keywords: mb to bytes, parse file size, size string to bytes, parse, bytes, size, gib, mib, kb, mb, convert, to, file
    Needs: none (pure function).
    """

def bytes_table(value: str) -> Dict[str, str]:
    """Shows a byte count in every unit: bit, B, KB, MB, GB, TB, PB (1000-based) and KiB, MiB, GiB, TiB, PiB (1024-based), each rounded to 6 decimals. Accepts a number or a size such as "1.5 GiB".

        value: Byte count or a size like 1.5GiB.

    Example: num.bytes_table("1073741824")
    Keywords: bytes, table, units, kb, mb, gb, tb, kib, mib, gib, conversion, size, chart
    Needs: none (pure function).
    """

def number_to_binary_bytes(value: str, size: int = 0, endian: str = "big", base: int = 0) -> List[str]:
    """Converts an integer to a list of two-digit hex bytes, big-endian by default (258 gives ["01", "02"]) or little-endian. size=0 uses the minimal width (negatives as two's complement); a fixed size pads or errors if the value does not fit (max 8192 bytes).

        value: Integer.
        size: Width in bytes (0 = minimal).
        endian: big or little. One of: big, little.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.number_to_binary_bytes("258")
    Keywords: bytes, to, endian, little, big, hex, array, serialize, integer
    Needs: none (pure function).
    """

def bytes_to_number(bytes: List[str], endian: str = "big", signed: bool = False) -> str:
    """Combines a list of hex bytes ("01", "0xff") into one integer, big-endian by default or little-endian, optionally as signed two's complement. Returns a decimal string.

        bytes: Hex byte strings.
        endian: big or little. One of: big, little.
        signed: Interpret as two's complement.

    Example: num.bytes_to_number(["01", "02"])
    Keywords: bytes, to, number, endian, little, big, hex, array, integer, deserialize
    Needs: none (pure function).
    """

def digits(value: str) -> List[str]:
    """Splits the absolute value of a decimal integer into its digits (1203 gives ["1", "2", "0", "3"]).

        value: Integer, decimal string.

    Example: num.digits("1203")
    Keywords: digits, split, digit, list, decimal, places
    Needs: none (pure function).
    """

def digit_sum(value: str) -> int:
    """Adds up the decimal digits of an integer's absolute value (9875 gives 29).

        value: Integer, decimal string.

    Example: num.digit_sum("9875")
    Keywords: digit, sum, add, digits, cross, checksum
    Needs: none (pure function).
    """

def from_bin(value: str) -> str:
    """Reads a binary integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

        value: The binary digits.

    Example: num.from_bin("1010")
    Keywords: binary to decimal, parse binary, binary, bin, decimal, convert, frombin
    Needs: none (pure function).
    """

def from_hex(value: str) -> str:
    """Reads a hexadecimal integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

        value: The hexadecimal digits.

    Example: num.from_hex("ff")
    Task: convert hex to a number
    Keywords: hex to decimal, parse hex, hexadecimal, hex, decimal, convert, fromhex
    Needs: none (pure function).
    """

def from_oct(value: str) -> str:
    """Reads a octal integer (prefix and underscores allowed) and returns it as a decimal string. Any size; garbage input is an error.

        value: The octal digits.

    Example: num.from_oct("17")
    Keywords: octal, oct, decimal, convert, fromoct
    Needs: none (pure function).
    """

def digital_root(value: str) -> int:
    """Repeatedly sums the digits until one digit remains (9875 gives 2); the root of 0 is 0. Uses the absolute value.

        value: Integer, decimal string.

    Example: num.digital_root("9875")
    Keywords: digital, root, repeated, digit, sum, casting, out, nines
    Needs: none (pure function).
    """

def reverse_digits(value: str) -> str:
    """Reverses the decimal digits of an integer, keeping the sign (1230 gives 321, -45 gives -54).

        value: Integer, decimal string.

    Example: num.reverse_digits("1230")
    Keywords: reverse, digits, flip, number, backwards
    Needs: none (pure function).
    """

def is_palindrome_number(value: str, base: int = 10) -> bool:
    """Tells whether an integer reads the same backwards in the given base (default 10; 585 is a palindrome, and so is 9 in base 2 as 1001). Negative numbers are never palindromes.

        value: Integer.
        base: Base 2..36 used for the digits.

    Example: num.is_palindrome_number("12321")
    Keywords: palindrome, number, symmetric, reads, same, backwards
    Needs: none (pure function).
    """

def detect_base(value: str) -> str:
    """Guesses the base of a number text: a 0x/0b/0o prefix gives hex/bin/oct, hex letters (a-f) mean hex, a plain digit run means dec. Returns "hex", "bin", "oct" or "dec", or an error if the text is not a number.

        value: Number text.

    Example: num.detect_base("0b1010")
    Keywords: is this hex or binary, guess number base, detect, guess, radix, base, prefix
    Needs: none (pure function).
    """

def is_prime(value: str) -> bool:
    """Tests whether an integer is prime. Exact for numbers below 2^64 and probabilistic (Miller-Rabin plus Baillie-PSW, 20 rounds) above; limited to 8192 bits. Numbers below 2 are not prime.

        value: Integer, decimal string.

    Example: num.is_prime("97")
    Task: check for a prime number
    Keywords: prime test, primality, prime, test, is, composite, miller, rabin
    Needs: none (pure function).
    """

def next_prime(value: str) -> str:
    """Returns the smallest prime strictly greater than the number (next_prime of 13 is 17; of -5 or 0 is 2). Limited to 1024-bit inputs.

        value: Integer, decimal string.

    Example: num.next_prime("100")
    Keywords: next, prime, following, greater
    Needs: none (pure function).
    """

def prev_prime(value: str) -> str:
    """Returns the largest prime strictly smaller than the number (prev_prime of 17 is 13). Numbers 2 and below have no previous prime and give an error. Limited to 1024-bit inputs.

        value: Integer, decimal string.

    Example: num.prev_prime("100")
    Keywords: previous, prime, smaller, lower
    Needs: none (pure function).
    """

def prime_factors(value: str) -> List[str]:
    """Returns the prime factors of an integer >= 2 with multiplicity, ascending (360 gives 2, 2, 2, 3, 3, 5). Uses trial division to 100000 and Pollard's rho; numbers over 512 bits or with a huge prime cofactor are an error.

        value: Integer, decimal string.

    Example: num.prime_factors("360")
    Task: factor a number
    Keywords: prime factorization, factorise, prime, factors, factorize, factor, decomposition, integer, factorization
    Needs: none (pure function).
    """

def factorization(value: str) -> str:
    """Writes the prime factorization of an integer >= 1 as a product with exponents: 360 gives "2^3 * 3^2 * 5", 1 gives "1". Same limits as prime_factors.

        value: Integer, decimal string.

    Example: num.factorization("360")
    Keywords: factorization, prime, power, exponent, factor, tree, decomposition
    Needs: none (pure function).
    """

def divisors(value: str) -> List[str]:
    """Lists all positive divisors of an integer >= 1 in ascending order (12 gives 1, 2, 3, 4, 6, 12). Limited to 512-bit numbers and 100000 divisors.

        value: Integer, decimal string.

    Example: num.divisors("28")
    Keywords: divisors, factors, all, proper, list
    Needs: none (pure function).
    """

def is_perfect(value: str) -> bool:
    """Tests whether a number equals the sum of its proper divisors (6, 28, 496, 8128, ...). Numbers below 2 are not perfect; limited to 128 bits.

        value: Integer, decimal string.

    Example: num.is_perfect("28")
    Keywords: perfect, number, divisors, sum, aliquot
    Needs: none (pure function).
    """

def is_armstrong(value: str) -> bool:
    """Tests whether a number equals the sum of its digits each raised to the number of digits (narcissistic number: 153, 370, 9474). Negative numbers are not.

        value: Integer, decimal string.

    Example: num.is_armstrong("153")
    Keywords: armstrong, narcissistic, number, digits, power
    Needs: none (pure function).
    """

def is_fibonacci(value: str) -> bool:
    """Tests whether a non-negative integer is a Fibonacci number (0, 1, 1, 2, 3, 5, 8, ...) using the 5n^2 +/- 4 perfect-square test.

        value: Integer, decimal string.

    Example: num.is_fibonacci("144")
    Keywords: fibonacci, number, test, is, sequence, member
    Needs: none (pure function).
    """

def fibonacci(value: str) -> str:
    """Returns the n-th Fibonacci number (F0=0, F1=1, F10=55) as a decimal string using fast doubling. n must be 0..100000.

        value: Integer, decimal string.

    Example: num.fibonacci("10")
    Keywords: nth fibonacci number, fibonacci, nth, number, sequence, golden
    Needs: none (pure function).
    """

def bit_and(a: str, b: str, base: int = 0, output: str = "dec") -> str:
    """Bitwise AND of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

        a: First integer.
        b: Second integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.bit_and("0xF0", "0x3C", output="hex")
    Keywords: bitwise and, mask bits, and, mask, bitwise, &, binary
    Needs: none (pure function).
    """

def bit_or(a: str, b: str, base: int = 0, output: str = "dec") -> str:
    """Bitwise OR of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

        a: First integer.
        b: Second integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.bit_or("0xF0", "0x3C", output="hex")
    Keywords: or, bitwise, |, binary, union
    Needs: none (pure function).
    """

def bit_xor(a: str, b: str, base: int = 0, output: str = "dec") -> str:
    """Bitwise XOR of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.

        a: First integer.
        b: Second integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.bit_xor("0xF0", "0x3C", output="hex")
    Keywords: xor, bitwise, ^, binary, exclusive
    Needs: none (pure function).
    """

def factorial(value: str) -> str:
    """Returns n! as a decimal string (0! = 1, 20! = 2432902008176640000). n must be 0..20000; larger values are an error to keep the result small.

        value: Integer, decimal string.

    Example: num.factorial("20")
    Keywords: n!, product of 1 to n, factorial, product, permutations
    Needs: none (pure function).
    """

def bit_not(value: str, bits: int = 0, base: int = 0, output: str = "dec") -> str:
    """Inverts the bits of an integer. With bits=0 returns -value-1 (infinite two's complement); with bits=N inverts within an N-bit field (0 < N <= 65536) and returns the unsigned pattern, zero-padded in hex/bin output.

        value: Integer.
        bits: Field width in bits (0 = unbounded).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.bit_not("0x0F", 8, output="bin")
    Keywords: not, invert, complement, ones, bitwise, ~, flip
    Needs: none (pure function).
    """

def collatz_steps(value: str) -> int:
    """Counts the steps of the 3n+1 (Collatz) sequence from n >= 1 until it reaches 1 (27 takes 111 steps). Gives an error if it exceeds one million steps.

        value: Integer, decimal string.

    Example: num.collatz_steps("27")
    Keywords: collatz, 3n+1, steps, hailstone, sequence
    Needs: none (pure function).
    """

def euler_phi(value: str) -> str:
    """Computes Euler's totient phi(n): how many numbers in 1..n are coprime to n (phi(36) = 12, phi(1) = 1). Uses the factorization, so it shares prime_factors' limits.

        value: Integer, decimal string.

    Example: num.euler_phi("36")
    Keywords: euler, phi, totient, coprime, count
    Needs: none (pure function).
    """

def shift_left(value: str, n: int = 1, base: int = 0, output: str = "dec") -> str:
    """Shifts a big integer left by n bits (multiplies by 2^n). n must be 0..1048576.

        value: Integer.
        n: Number of bit positions.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.shift_left("0b1010", 2, output="bin")
    Keywords: bit shift, multiply by power of two, shift, left, bitshift, <<, >>, multiply, divide, power, of, two
    Needs: none (pure function).
    """

def shift_right(value: str, n: int = 1, base: int = 0, output: str = "dec") -> str:
    """Shifts a big integer right by n bits (arithmetic shift, floors toward negative infinity). n must be 0..1048576.

        value: Integer.
        n: Number of bit positions.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.shift_right("0b1010", 2, output="bin")
    Keywords: shift, right, bitshift, <<, >>, multiply, divide, power, of, two
    Needs: none (pure function).
    """

def isqrt(value: str) -> str:
    """Returns floor(sqrt(n)) exactly for any size (isqrt of 99 is 9). Negative input is an error.

        value: Integer, decimal string.

    Example: num.isqrt("99")
    Keywords: integer, square, root, isqrt, floor, sqrt
    Needs: none (pure function).
    """

def rotate_left(value: str, n: int = 1, bits: int = 32, base: int = 0, output: str = "dec") -> str:
    """Rotates the low `bits` bits of value left by n positions (wrapping around) and returns the unsigned pattern, zero-padded in hex/bin output. Negative values are taken as two's complement; n is reduced modulo bits.

        value: Integer.
        n: Positions to rotate.
        bits: Field width in bits (1..65536).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.rotate_left("0x81", 1, 8, output="hex")
    Keywords: rotate, rol, ror, circular, shift, bits, left
    Needs: none (pure function).
    """

def rotate_right(value: str, n: int = 1, bits: int = 32, base: int = 0, output: str = "dec") -> str:
    """Rotates the low `bits` bits of value right by n positions (wrapping around) and returns the unsigned pattern, zero-padded in hex/bin output. Negative values are taken as two's complement; n is reduced modulo bits.

        value: Integer.
        n: Positions to rotate.
        bits: Field width in bits (1..65536).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.rotate_right("0x81", 1, 8, output="hex")
    Keywords: rotate, rol, ror, circular, shift, bits, right
    Needs: none (pure function).
    """

def clear_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str:
    """Sets one bit to 0. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

        value: Integer.
        index: Bit position, 0 = least significant.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.clear_bit("0b1000", 1)
    Keywords: bit, flag, clear, mask, index, position
    Needs: none (pure function).
    """

def icbrt(value: str) -> str:
    """Returns the cube root truncated toward zero (icbrt of 30 is 3, of -30 is -3), exact for any size.

        value: Integer, decimal string.

    Example: num.icbrt("1000")
    Keywords: integer, cube, root, icbrt, floor, cbrt
    Needs: none (pure function).
    """

def set_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str:
    """Sets one bit to 1. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

        value: Integer.
        index: Bit position, 0 = least significant.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.set_bit("0b1000", 1)
    Keywords: bit, flag, set, mask, index, position
    Needs: none (pure function).
    """

def test_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> bool:
    """Returns true if the bit is 1. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

        value: Integer.
        index: Bit position, 0 = least significant.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.test_bit("0b1000", 1)
    Keywords: bit, flag, test, mask, index, position
    Needs: none (pure function).
    """

def toggle_bit(value: str, index: int = 0, base: int = 0, output: str = "dec") -> str:
    """Flips one bit. Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0..1048576.

        value: Integer.
        index: Bit position, 0 = least significant.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.toggle_bit("0b1000", 1)
    Keywords: bit, flag, toggle, mask, index, position
    Needs: none (pure function).
    """

def bit_length(value: str, base: int = 0) -> int:
    """Returns the number of bits needed to represent the absolute value (0 for 0, 8 for 255, 9 for 256).

        value: Integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.bit_length("255")
    Keywords: bits, length, size, how, many, bitlen, width
    Needs: none (pure function).
    """

def nth_root(value: str, n: int = 2) -> str:
    """Returns the integer part of the n-th root of a number (nth_root(1000000, 6) is 10), exact for any size. n must be 1..4096; negative values need an odd n and truncate toward zero.

        value: Integer, decimal string.
        n: Root degree 1..4096.

    Example: num.nth_root("1024", 10)
    Keywords: nth, root, integer, radical, floor, exact
    Needs: none (pure function).
    """

def gcd(a: str, b: str) -> str:
    """Greatest common divisor of two integers of any size (gcd(12, 18) = 6; gcd with 0 returns the other absolute value; signs are ignored).

        a: First integer.
        b: Second integer.

    Example: num.gcd("12", "18")
    Task: greatest common divisor
    Keywords: gcf, highest common factor, gcd, greatest, common, divisor, hcf, euclid
    Needs: none (pure function).
    """

def popcount(value: str, bits: int = 0, base: int = 0) -> int:
    """Counts the 1 bits (Hamming weight). With bits=0 a negative number counts the bits of its absolute value; with bits=N the value is first reduced to an N-bit two's complement pattern.

        value: Integer.
        bits: Field width (0 = none).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.popcount("0xFF")
    Keywords: popcount, hamming, weight, count, ones, set, bits, bitcount
    Needs: none (pure function).
    """

def lcm(a: str, b: str) -> str:
    """Least common multiple of two integers of any size (lcm(4, 6) = 12; returns 0 if either is 0).

        a: First integer.
        b: Second integer.

    Example: num.lcm("4", "6")
    Task: least common multiple
    Keywords: lowest common multiple, lcm, least, common, multiple
    Needs: none (pure function).
    """

def parity(value: str, base: int = 0) -> int:
    """Returns 1 if the number of set bits (of the absolute value) is odd, else 0.

        value: Integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.parity("7")
    Keywords: parity, odd, even, bits, xor, bit
    Needs: none (pure function).
    """

def gcd_list(values: List[str]) -> str:
    """Folds a list of integer strings with gcd. The list must not be empty.

        values: Integers as decimal strings.

    Example: num.gcd_list(["12", "18", "30"])
    Keywords: gcd, lcm, list, many, numbers
    Needs: none (pure function).
    """

def lcm_list(values: List[str]) -> str:
    """Folds a list of integer strings with lcm. The list must not be empty.

        values: Integers as decimal strings.

    Example: num.lcm_list(["12", "18", "30"])
    Keywords: gcd, lcm, list, many, numbers
    Needs: none (pure function).
    """

def leading_zeros(value: str, bits: int = 32, base: int = 0) -> int:
    """Counts the zero bits above the highest 1 inside a `bits`-wide field. A value that does not fit the field is an error; negatives use two's complement (so 0 leading zeros).

        value: Integer.
        bits: Field width 1..65536.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.leading_zeros("1", 32)
    Keywords: leading, zeros, clz, count, bits, msb
    Needs: none (pure function).
    """

def modpow(base: str, exp: str, mod: str) -> str:
    """Computes base^exp mod m with arbitrary precision (modpow(4, 13, 497) = 445). m must be positive; a negative exp needs base invertible mod m.

        base: Base.
        exp: Exponent.
        mod: Modulus, > 0.

    Example: num.modpow("4", "13", "497")
    Keywords: modpow, modular, exponentiation, pow, mod, power, modulo, rsa
    Needs: none (pure function).
    """

def trailing_zeros(value: str, base: int = 0) -> int:
    """Counts the zero bits below the lowest 1 (ctz). Returns 0 for the value 0; the sign is ignored.

        value: Integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.trailing_zeros("8")
    Keywords: trailing, zeros, ctz, count, bits, lsb
    Needs: none (pure function).
    """

def modinv(a: str, m: str) -> str:
    """Finds x with a*x = 1 (mod m), 0 <= x < m (modinv(3, 11) = 4). Errors when gcd(a, m) is not 1 or m < 2.

        a: Number to invert.
        m: Modulus, >= 2.

    Example: num.modinv("3", "11")
    Keywords: modular, inverse, modinv, multiplicative, extended, euclid
    Needs: none (pure function).
    """

def reverse_bits(value: str, bits: int = 32, base: int = 0, output: str = "dec") -> str:
    """Reverses the order of the low `bits` bits (bit 0 becomes bit bits-1). Negatives use two's complement; hex/bin output is zero-padded to the width.

        value: Integer.
        bits: Field width 1..65536.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.reverse_bits("0b0001", 8, output="bin")
    Keywords: reverse, bits, mirror, flip, order, bit, reversal
    Needs: none (pure function).
    """

def swap_endian(value: str, bytes: int = 4, base: int = 0, output: str = "dec") -> str:
    """Reverses the byte order of the low `bytes` bytes (big-endian <-> little-endian), e.g. 0x12345678 becomes 0x78563412 for 4 bytes. Negatives use two's complement; hex output is zero-padded.

        value: Integer.
        bytes: Width in bytes (1..8192).
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.swap_endian("0x12345678", 4, output="hex")
    Keywords: byte order, big endian little endian, byteswap, endian, byte, swap, bswap, htonl, ntohl, little, big, order
    Needs: none (pure function).
    """

def to_twos(value: str, bits: int = 32, base: int = 0, output: str = "bin") -> str:
    """Converts a signed integer to its two's complement bit pattern in a `bits`-wide field (e.g. -1 in 8 bits is 0xFF). The value must satisfy -2^(bits-1) <= value < 2^bits. Hex/bin output is zero-padded to the width.

        value: Signed integer.
        bits: Field width 1..65536.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct.

    Example: num.to_twos("-5", 8, output="bin")
    Keywords: twos, complement, signed, negative, representation, binary, two's, encode
    Needs: none (pure function).
    """

def from_twos(value: str, bits: int = 32, base: int = 0) -> str:
    """Interprets a bit pattern as a `bits`-wide two's complement number and returns the signed decimal value (0xFF in 8 bits is -1). The pattern must be non-negative and fit the width.

        value: Bit pattern, e.g. 0xFB or 0b11111011.
        bits: Field width 1..65536.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.

    Example: num.from_twos("0xFB", 8)
    Keywords: twos, complement, signed, decode, negative, interpret, int8, int16, int32, int64
    Needs: none (pure function).
    """

def sign_extend(value: str, from_bits: int = 8, to_bits: int = 32, base: int = 0, output: str = "hex") -> str:
    """Sign-extends a from_bits-wide two's complement pattern to to_bits and returns the new unsigned pattern (0x80 from 8 to 16 bits is 0xFF80). to_bits must be >= from_bits.

        value: Bit pattern.
        from_bits: Source width 1..65536.
        to_bits: Target width 1..65536.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct.

    Example: num.sign_extend("0x80", 8, 16, output="hex")
    Keywords: sign, extend, sext, widen, bits, two's, complement
    Needs: none (pure function).
    """

def mask(bits: int, shift: int = 0, output: str = "hex") -> str:
    """Builds a mask of `bits` consecutive 1 bits, optionally shifted left (mask(8) = 0xFF, mask(4, 4) = 0xF0). bits 0..65536, shift 0..1048576.

        bits: Number of 1 bits.
        shift: Left shift applied to the mask.
        output: Output format: dec, hex, bin or oct. One of: dec, hex, bin, oct.

    Example: num.mask(8, 4, output="hex")
    Keywords: mask, bitmask, ones, field, width, generate
    Needs: none (pure function).
    """

def gray_encode(value: str, base: int = 0, output: str = "dec") -> str:
    """Converts a non-negative integer to its reflected binary Gray code (value XOR value>>1), so 2 becomes 3 and 7 becomes 4.

        value: Non-negative integer.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.gray_encode("7")
    Keywords: gray, code, reflected, binary, encode
    Needs: none (pure function).
    """

def gray_decode(value: str, base: int = 0, output: str = "dec") -> str:
    """Converts a Gray code back to the ordinary binary integer (inverse of gray_encode), so 4 becomes 7.

        value: Non-negative Gray code.
        base: Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.
        output: Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix). One of: dec, hex, bin, oct.

    Example: num.gray_decode("4")
    Keywords: gray, code, reflected, binary, decode
    Needs: none (pure function).
    """

def luhn_check(value: str) -> bool:
    """Validates a number with the Luhn (mod 10) algorithm used by credit cards, IMEI and many ids. Spaces and dashes are ignored; at least 2 digits are required; non-digits are an error.

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.luhn_check("4539 1488 0343 6467")
    Task: validate a card number checksum
    Keywords: credit card check, luhn validate, luhn, mod, 10, credit, card, validate, checksum, imei
    Needs: none (pure function).
    """

def luhn_generate(value: str) -> str:
    """Computes the single Luhn check digit to append to a number (7992739871 gives 3). Use luhn_complete to get the full number.

        value: Digits without the check digit; spaces, dashes and dots are ignored.

    Example: num.luhn_generate("7992739871")
    Keywords: luhn, check, digit, generate, credit, card, checksum
    Needs: none (pure function).
    """

def luhn_complete(value: str) -> str:
    """Appends the Luhn check digit to a number, giving a valid Luhn number (7992739871 gives 79927398713).

        value: Digits without the check digit; spaces, dashes and dots are ignored.

    Example: num.luhn_complete("7992739871")
    Keywords: luhn, complete, append, check, digit, credit, card
    Needs: none (pure function).
    """

def isbn10_check(value: str) -> bool:
    """Validates an ISBN-10 (weights 10..1, mod 11; the last character may be X). Hyphens and spaces are ignored.

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.isbn10_check("0-306-40615-2")
    Keywords: isbn, isbn10, book, validate, checksum
    Needs: none (pure function).
    """

def isbn13_check(value: str) -> bool:
    """Validates an ISBN-13 (EAN-13 check digit, must start with 978 or 979). Hyphens and spaces are ignored.

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.isbn13_check("978-3-16-148410-0")
    Task: validate an ISBN
    Keywords: book number check, isbn, isbn13, book, validate, checksum, ean
    Needs: none (pure function).
    """

def isbn10_to_13(value: str) -> str:
    """Converts a valid ISBN-10 to ISBN-13 by prefixing 978 and recomputing the check digit (0306406152 gives 9780306406157). Invalid input is an error.

        value: ISBN-10.

    Example: num.isbn10_to_13("0306406152")
    Keywords: isbn, convert, 10, to, 13, isbn13
    Needs: none (pure function).
    """

def isbn13_to_10(value: str) -> str:
    """Converts a valid 978-prefixed ISBN-13 to ISBN-10 (9780306406157 gives 0306406152). 979 numbers have no ISBN-10 and give an error.

        value: ISBN-13.

    Example: num.isbn13_to_10("9780306406157")
    Keywords: isbn, convert, 13, to, 10, isbn10
    Needs: none (pure function).
    """

def iban_check(value: str) -> bool:
    """Validates an IBAN: structure, the country's fixed length (for ~75 known countries) and the ISO 7064 mod-97 check. Spaces and dashes are ignored, case-insensitive. Returns false for a wrong checksum and an error for illegal characters.

        value: IBAN.

    Example: num.iban_check("GB82 WEST 1234 5698 7654 32")
    Task: validate an IBAN
    Keywords: bank account number check, iban, bank, account, validate, mod, 97, sepa, checksum
    Needs: none (pure function).
    """

def iban_format(value: str) -> str:
    """Formats an IBAN in groups of four characters, uppercase (GB82WEST12345698765432 gives "GB82 WEST 1234 5698 7654 32"). Does not validate the checksum.

        value: IBAN, with or without spaces.

    Example: num.iban_format("GB82WEST12345698765432")
    Keywords: iban, format, group, four, spaces, bank, account
    Needs: none (pure function).
    """

def ean13_check(value: str) -> bool:
    """Validates an EAN-13 barcode (13 digits, weights 1 and 3, mod 10).

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.ean13_check("4006381333931")
    Keywords: ean, ean13, barcode, gtin, validate, checksum
    Needs: none (pure function).
    """

def ean8_check(value: str) -> bool:
    """Validates an EAN-8 barcode (8 digits, weights 3 and 1, mod 10).

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.ean8_check("96385074")
    Keywords: ean, ean8, barcode, gtin, validate, checksum
    Needs: none (pure function).
    """

def upc_check(value: str) -> bool:
    """Validates a UPC-A barcode (12 digits, mod 10 check digit).

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.upc_check("036000291452")
    Keywords: upc, upca, barcode, gtin, validate, checksum
    Needs: none (pure function).
    """

def gtin_check(value: str) -> bool:
    """Validates any GTIN barcode number: GTIN-8, UPC-A (12), EAN-13 or GTIN-14, using the mod 10 check digit.

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.gtin_check("4006381333931")
    Keywords: gtin, barcode, ean, upc, validate, checksum
    Needs: none (pure function).
    """

def gtin_generate(value: str) -> str:
    """Computes the barcode check digit for 7, 11, 12 or 13 digits (GTIN-8, UPC-A, EAN-13, GTIN-14); 400638133393 gives 1.

        value: Digits without the check digit; spaces, dashes and dots are ignored.

    Example: num.gtin_generate("400638133393")
    Keywords: gtin, ean, upc, barcode, check, digit, generate
    Needs: none (pure function).
    """

def verhoeff_check(value: str) -> bool:
    """Validates a number with the Verhoeff dihedral-group checksum (used by Aadhaar); detects all single-digit and adjacent-transposition errors.

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.verhoeff_check("2363")
    Keywords: verhoeff, checksum, aadhaar, validate, dihedral
    Needs: none (pure function).
    """

def verhoeff_generate(value: str) -> str:
    """Computes the Verhoeff check digit to append to a number (236 gives 3).

        value: Digits without the check digit; spaces, dashes and dots are ignored.

    Example: num.verhoeff_generate("236")
    Keywords: verhoeff, check, digit, generate, aadhaar
    Needs: none (pure function).
    """

def damm_check(value: str) -> bool:
    """Validates a number with the Damm algorithm (quasigroup check digit; catches all single-digit and adjacent-transposition errors).

        value: Digits; spaces, dashes and dots are ignored.

    Example: num.damm_check("5724")
    Keywords: damm, checksum, validate, quasigroup
    Needs: none (pure function).
    """

def damm_generate(value: str) -> str:
    """Computes the Damm check digit to append to a number (572 gives 4).

        value: Digits without the check digit; spaces, dashes and dots are ignored.

    Example: num.damm_generate("572")
    Keywords: damm, check, digit, generate
    Needs: none (pure function).
    """
