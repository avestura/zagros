"""Encoding and escaping module (generated from the tools registry; do not edit by hand).

Base16/32/58/62/64/85 and more, charsets, URL, HTML, punycode, escape and unescape, JWT decode.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
"""

from typing import Dict, List

def base64_encode(text: str, variant: str = "standard", wrap: int = 0, input_encoding: str = "utf-8") -> str:
    """Encodes text as Base64 (RFC 4648). Variants: standard with padding (default), url (URL-safe alphabet), standard_nopad, url_nopad, mime (wrapped at 76 columns with CRLF); wrap adds line breaks every N characters. Text is encoded as UTF-8 bytes unless input_encoding is hex/base64/etc.

        text: Text to encode.
        variant: Alphabet and padding. One of: standard, url, standard_nopad, url_nopad, mime.
        wrap: Insert a line break every N characters (0 = no wrapping; mime always wraps at 76).
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base64_encode("Hello, World!")
    Needs: none (pure function).
    """

def base64_decode(text: str, output: str = "text") -> str:
    """Decodes Base64 text. Tolerant: accepts the standard and URL-safe alphabets, missing padding, embedded whitespace/line breaks and a data: URI prefix. Returns text when the bytes are valid UTF-8, otherwise an error unless output is hex/base64/etc.

        text: Base64 text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base64_decode("SGVsbG8sIFdvcmxkIQ==")
    Needs: none (pure function).
    """

def url_encode(text: str, mode: str = "component", charset: str = "utf-8", input_encoding: str = "utf-8") -> str:
    """Percent-encodes text as UTF-8. Modes: component (encodeURIComponent semantics, default), query (form style: space becomes +), path (keeps / : @ & = + $ , ;), full (encodeURI semantics, keeps URL syntax characters) and rfc3986 (only unreserved characters stay). charset encodes with another charset first (e.g. windows-1252 for %E9).

        text: Text to encode.
        mode: Which characters stay unescaped. One of: component, query, path, full, rfc3986.
        charset: Charset used to turn text into bytes.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.url_encode("a b&c=d/é", mode="component")
    Needs: none (pure function).
    """

def url_decode(text: str, plus_as_space: bool = False, charset: str = "utf-8", output: str = "text") -> str:
    """Decodes %XX percent escapes (UTF-8 by default). Tolerant: malformed % sequences stay literal; plus_as_space turns + into a space (form data). charset decodes legacy encodings such as windows-1252; non-UTF-8 bytes are an error unless a charset or output is given.

        text: Percent-encoded text.
        plus_as_space: Treat + as a space (application/x-www-form-urlencoded).
        charset: Charset of the decoded bytes.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.url_decode("a%20b%26c%3Dd%2F%C3%A9")
    Needs: none (pure function).
    """

def html_escape(text: str, quotes: bool = True) -> str:
    """Escapes the five markup-significant characters & < > " ' as &amp; &lt; &gt; &quot; &#39; so text is safe in HTML text and attribute values. quotes=false leaves quotes alone (text content only). Non-ASCII characters are kept; use html_escape_all to escape them too.

        text: Text to escape.
        quotes: Also escape " and '.

    Example: encode.html_escape("<a href=\"x\">Tom & Jerry</a>")
    Needs: none (pure function).
    """

def html_unescape(text: str) -> str:
    """Decodes HTML character references: named entities (the full HTML5 table, with or without the trailing semicolon where HTML allows), decimal &#169; and hex &#xA9; numeric references. Unknown entities are left as they are.

        text: Text containing entities.

    Example: encode.html_unescape("Tom &amp; Jerry &copy; &#x1F600;")
    Needs: none (pure function).
    """

def hex_encode(text: str, case: str = "lower", separator: str = "", prefix: str = "", input_encoding: str = "utf-8") -> str:
    """Encodes text as hexadecimal bytes (UTF-8 unless input_encoding says otherwise). case chooses lower/upper digits, separator goes between bytes (":", " ", "-") and prefix before each byte ("0x", "\\x").

        text: Text to encode.
        case: Hex digit case. One of: lower, upper.
        separator: Text placed between bytes, e.g. ":", " " or ", ".
        prefix: Text placed before each byte, e.g. "0x" or "\\x".
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.hex_encode("Hello", separator=" ")
    Needs: none (pure function).
    """

def hex_decode(text: str, output: str = "text") -> str:
    """Decodes hexadecimal bytes to text. Tolerant of spaces, colons, dashes, commas and 0x, \x or % prefixes. Fails on odd digit counts or non-hex characters; returns text if valid UTF-8, otherwise an error unless output is base64/etc.

        text: Hex text, e.g. "48 65 6c 6c 6f" or "0x48,0x65".
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.hex_decode("48 65 6c 6c 6f")
    Needs: none (pure function).
    """

def detect(text: str) -> List[str]:
    """Lists the decoder functions of this module that plausibly apply to the text, best first (for example ["jwt_decode"], ["url_decode"], ["hex_decode", "base64_decode"]). Each entry is a function name callable as encode.<name>(text). Heuristic: it only reports decoders whose output would be readable text.

        text: Unknown encoded text.

    Example: encode.detect("SGVsbG8sIFdvcmxkIQ==")
    Needs: none (pure function).
    """

def morse_encode(text: str, dot: str = ".", dash: str = "-", sep: str = " ", word_sep: str = " / ") -> str:
    """Encodes text as ITU Morse code: letters separated by a space, words by " / ". Letters, digits and common punctuation are supported; accented letters are folded to their base letter (é becomes e) except ä ö ü ñ é ç å which have Morse forms; other characters are dropped. dot/dash/sep/word_sep customize the symbols.

        text: Text to encode.
        dot: Symbol for a dot.
        dash: Symbol for a dash.
        sep: Between letters.
        word_sep: Between words.

    Example: encode.morse_encode("SOS hello")
    Needs: none (pure function).
    """

def morse_decode(text: str) -> str:
    """Decodes Morse code made of . and - (also • · – —): letters separated by spaces, words by / or | or three or more spaces or a line break. Unknown dot/dash groups of valid characters decode to "?"; other characters are an error. Result is upper case.

        text: Morse code, e.g. "... --- ...".

    Example: encode.morse_decode("... --- ... / .... .. ")
    Needs: none (pure function).
    """

def nato_phonetic_encode(text: str, separator: str = " ", digits: bool = True) -> str:
    """Spells text with the NATO/ICAO phonetic alphabet (Alpha Bravo Charlie ...). Digits become words (Zero, One, ...) unless digits=false, spaces become "/", and other characters are kept as they are.

        text: Text to spell.
        separator: Between words.
        digits: Spell digits as words.

    Example: encode.nato_phonetic_encode("Go 42")
    Needs: none (pure function).
    """

def nato_decode(text: str, lower: bool = False) -> str:
    """Turns NATO phonetic words (Alpha, Bravo, ... also Alfa, Juliett, Niner, Fife, digit words) back into letters and digits, upper case unless lower=true. "/" becomes a space and single punctuation marks are kept; any other word is an error.

        text: Phonetic words.
        lower: Return lower case.

    Example: encode.nato_decode("Golf Oscar / Four Two")
    Needs: none (pure function).
    """

def braille_encode(text: str, variant: str = "literary", capitals: bool = True) -> str:
    """Converts text to Unicode Braille patterns. variant "literary" (default) is grade 1 English braille: capital sign before capitals (capitals=false drops it), number sign before digit runs, basic punctuation. variant "ascii" maps each ASCII character to its Braille ASCII (BRF) cell.

        text: Text to convert.
        variant: Braille flavor. One of: literary, ascii.
        capitals: Emit the capital sign for upper-case letters.

    Example: encode.braille_encode("Hello 42")
    Needs: none (pure function).
    """

def braille_decode(text: str, variant: str = "literary") -> str:
    """Converts Unicode Braille patterns (U+2800-U+28FF) back to text using grade 1 English rules (capital sign, number sign, basic punctuation) or, with variant "ascii", Braille ASCII. Non-braille characters pass through; text with no braille cells is an error.

        text: Braille text.
        variant: Braille flavor. One of: literary, ascii.

    Example: encode.braille_decode("⠠⠓⠑⠇⠇⠕")
    Needs: none (pure function).
    """

def tap_code_encode(text: str, style: str = "dots") -> str:
    """Encodes letters with the 5x5 tap code (Polybius square without K; K is written as C). style "dots" gives ".. ..." (row dots, column dots), style "numbers" gives "23"; letters are separated by two spaces and words by " / ". Non-letters are dropped.

        text: Text to encode.
        style: Output style. One of: dots, numbers.

    Example: encode.tap_code_encode("hello")
    Needs: none (pure function).
    """

def tap_code_decode(text: str) -> str:
    """Decodes tap code written as dot groups (".. ...", row then column) or two-digit numbers ("23"); words are separated by /. Returns upper-case letters (C also stands for K). Fails on malformed groups.

        text: Tap code.

    Example: encode.tap_code_decode("23 15 31 31 34")
    Needs: none (pure function).
    """

def a1z26_encode(text: str, separator: str = "-") -> str:
    """Replaces each letter with its alphabet position (A=1 ... Z=26), joined by separator ("-" default) within a word; words are separated by a space. Case-insensitive; characters that are not ASCII letters are dropped.

        text: Text to encode.
        separator: Between letters of a word.

    Example: encode.a1z26_encode("hello world")
    Needs: none (pure function).
    """

def a1z26_decode(text: str) -> str:
    """Turns numbers 1-26 back into upper-case letters. With "-" or "," in the input, those join letters and spaces separate words; otherwise spaces separate letters. "/" or a line break is always a word break. Numbers outside 1-26 are an error.

        text: Numbers such as "8-5-12-12-15 23-15-18-12-4".

    Example: encode.a1z26_decode("8-5-12-12-15")
    Needs: none (pure function).
    """

def ascii_table_lookup(char: str) -> Dict[str, str]:
    """Looks up one character and returns a dict with char, codepoint (U+XXXX), decimal, hex, octal, binary, utf8_hex, ascii ("true"/"false"), class and, for control codes, name (NUL, LF, ESC ...); html_entity for known non-ASCII characters. More than one character is an error.

        char: A single character, e.g. "A" or "\n".

    Example: encode.ascii_table_lookup("A")
    Needs: none (pure function).
    """

def base64url_encode(text: str, input_encoding: str = "utf-8") -> str:
    """Encodes text as URL-safe Base64 without padding (RFC 4648 section 5), the form used in JWTs and URLs. Text is encoded as UTF-8 unless input_encoding says otherwise.

        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base64url_encode("subjects?_d=1")
    Needs: none (pure function).
    """

def base64url_decode(text: str, output: str = "text") -> str:
    """Decodes URL-safe Base64 (with or without padding; the standard alphabet is accepted too). Returns text when valid UTF-8, otherwise an error unless output is hex/base64/etc.

        text: Base64url text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base64url_decode("c3ViamVjdHM_X2Q9MQ")
    Needs: none (pure function).
    """

def hex_to_base64(hex: str, variant: str = "standard") -> str:
    """Converts a hexadecimal byte string (separators, 0x and \x prefixes tolerated) to Base64, without going through text. Fails on odd length or non-hex characters.

        hex: Hex bytes, e.g. "48656c6c6f" or "48 65 6c 6c 6f".
        variant: Base64 variant. One of: standard, url, standard_nopad, url_nopad, mime.

    Example: encode.hex_to_base64("48656c6c6f")
    Needs: none (pure function).
    """

def base64_to_hex(text: str, case: str = "lower", separator: str = "") -> str:
    """Converts Base64 (standard or URL-safe, padding optional) to a hexadecimal byte string, without requiring the bytes to be text. Optional upper-case digits and a separator between bytes.

        text: Base64 text.
        case: Hex digit case. One of: lower, upper.
        separator: Text placed between bytes (e.g. ":" or " ").

    Example: encode.base64_to_hex("SGVsbG8=")
    Needs: none (pure function).
    """

def text_to_hex_bytes(text: str, prefix: str = "0x", separator: str = " ", case: str = "lower", input_encoding: str = "utf-8") -> str:
    """Writes the UTF-8 bytes of the text as hex bytes with a 0x prefix, separated by spaces by default (the common C/array literal style). Change prefix and separator for other styles.

        text: Text to convert.
        prefix: Prefix of each byte.
        separator: Separator between bytes.
        case: Hex digit case. One of: lower, upper.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.text_to_hex_bytes("Hi", separator=", ")
    Needs: none (pure function).
    """

def hex_bytes_to_text(text: str, output: str = "text") -> str:
    """Reads hex bytes written in any common style ("0x48, 0x65", "\\x48\\x65", "48:65", "4865") and returns the text. Same rules as hex_decode.

        text: Hex bytes.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.hex_bytes_to_text("0x48 0x65 0x6c 0x6c 0x6f")
    Needs: none (pure function).
    """

def base32_encode(text: str, variant: str = "rfc4648", padding: bool = True, input_encoding: str = "utf-8") -> str:
    """Encodes text as Base32. Variants: rfc4648 (default, A-Z2-7, padded), hex (extended hex alphabet), crockford (no padding, upper case), zbase32 (z-base-32, lower case, no padding). padding=false drops '=' for rfc4648/hex.

        text: Text to encode.
        variant: Alphabet. One of: rfc4648, hex, crockford, zbase32.
        padding: Add '=' padding (rfc4648 and hex only).
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base32_encode("Hello")
    Needs: none (pure function).
    """

def base32_decode(text: str, variant: str = "rfc4648", output: str = "text") -> str:
    """Decodes Base32 in the chosen variant (rfc4648, hex, crockford, zbase32). Case-insensitive, ignores whitespace and padding; Crockford also ignores hyphens and maps I/L to 1 and O to 0. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base32 text.
        variant: Alphabet. One of: rfc4648, hex, crockford, zbase32.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base32_decode("JBSWY3DP")
    Needs: none (pure function).
    """

def base36_encode(text: str, upper: bool = False, input_encoding: str = "utf-8") -> str:
    """Encodes the bytes of the text as one big number in base 36 (0-9a-z); leading zero bytes become leading '0' characters. Inputs above 32 KiB are refused. For plain integers use the num group.

        text: Text to encode.
        upper: Use upper-case letters.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base36_encode("Hello")
    Needs: none (pure function).
    """

def base36_decode(text: str, output: str = "text") -> str:
    """Decodes base 36 text (case-insensitive) into the bytes of a big number, reversing base36_encode. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base36 text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base36_decode("3yud78mn")
    Needs: none (pure function).
    """

def base45_encode(text: str, input_encoding: str = "utf-8") -> str:
    """Encodes text as Base45 (RFC 9285, used in EU digital COVID certificates and QR codes). Output uses upper-case letters, digits and ' $%*+-./:'.

        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base45_encode("ietf!")
    Needs: none (pure function).
    """

def base45_decode(text: str, output: str = "text") -> str:
    """Decodes Base45 (RFC 9285). Upper-case only; fails on characters outside the alphabet, invalid length or out-of-range groups. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base45 text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base45_decode("QED8WEX0")
    Needs: none (pure function).
    """

def base58_encode(text: str, variant: str = "bitcoin", input_encoding: str = "utf-8") -> str:
    """Encodes the bytes of the text in Base58 (Bitcoin alphabet by default; flickr and ripple variants available). Leading zero bytes become leading '1' (first alphabet character). Inputs above 32 KiB are refused.

        text: Text to encode.
        variant: Alphabet. One of: bitcoin, flickr, ripple.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base58_encode("Hello World")
    Needs: none (pure function).
    """

def base58_decode(text: str, variant: str = "bitcoin", output: str = "text") -> str:
    """Decodes Base58 (bitcoin, flickr or ripple alphabet). Fails on characters outside the alphabet (0, O, I, l in Bitcoin). Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base58 text.
        variant: Alphabet. One of: bitcoin, flickr, ripple.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base58_decode("JxF12TrwUP45BMd")
    Needs: none (pure function).
    """

def base58check_encode(text: str, version: int = 0, variant: str = "bitcoin", input_encoding: str = "utf-8") -> str:
    """Base58Check: prepends a version byte, appends the first 4 bytes of the double SHA-256 checksum and encodes in Base58 (Bitcoin addresses, WIF keys). Default version 0 (P2PKH); payload is the text bytes or, with input_encoding=hex, raw bytes.

        text: Payload.
        version: Version byte 0-255.
        variant: Alphabet. One of: bitcoin, flickr, ripple.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base58check_encode("00112233445566778899aabbccddeeff00112233", input_encoding="hex")
    Needs: none (pure function).
    """

def base58check_decode(text: str, variant: str = "bitcoin", output: str = "text") -> str:
    """Decodes Base58Check, verifies the 4-byte double SHA-256 checksum and returns the payload without the version byte (use base58check_parse to see the version). Fails with a checksum error on corrupted input; output hex/base64 for binary payloads.

        text: Base58Check text.
        variant: Alphabet. One of: bitcoin, flickr, ripple.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base58check_decode("1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr", output="hex")
    Needs: none (pure function).
    """

def base58check_parse(text: str, variant: str = "bitcoin") -> Dict[str, str]:
    """Parses a Base58Check string (Bitcoin address, WIF) into version, payload_hex, checksum and valid ("true"/"false"). Does not fail on a bad checksum, only on invalid characters or a too-short string.

        text: Base58Check text.
        variant: Alphabet. One of: bitcoin, flickr, ripple.

    Example: encode.base58check_parse("1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr")
    Needs: none (pure function).
    """

def base62_encode(text: str, variant: str = "standard", input_encoding: str = "utf-8") -> str:
    """Encodes the bytes of the text as one big number in base 62 (0-9A-Za-z by default; gmp = 0-9a-zA-Z, inverted = a-zA-Z0-9). Leading zero bytes become leading '0' characters. Inputs above 32 KiB are refused.

        text: Text to encode.
        variant: Alphabet order. One of: standard, gmp, inverted.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base62_encode("Hello")
    Needs: none (pure function).
    """

def base62_decode(text: str, variant: str = "standard", output: str = "text") -> str:
    """Decodes base 62 text (alphabet per variant) back to bytes, reversing base62_encode. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base62 text.
        variant: Alphabet order. One of: standard, gmp, inverted.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base62_decode("5TP3P3v")
    Needs: none (pure function).
    """

def base85_encode(text: str, variant: str = "ascii85", delimiters: bool = False, input_encoding: str = "utf-8") -> str:
    """Encodes text as Base85: ascii85 (Adobe/btoa style, 'z' for zero groups, optional <~ ~> delimiters), z85 (ZeroMQ) or rfc1924 (the IPv6 alphabet). z85 and rfc1924 handle a partial last group like ascii85 does.

        text: Text to encode.
        variant: Alphabet. One of: ascii85, z85, rfc1924.
        delimiters: Wrap ascii85 output in <~ and ~>.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base85_encode("Hello World", variant="ascii85")
    Needs: none (pure function).
    """

def base85_decode(text: str, variant: str = "ascii85", output: str = "text") -> str:
    """Decodes Base85 (ascii85 with optional <~ ~> delimiters and 'z' groups, z85, or rfc1924). Whitespace is ignored. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base85 text.
        variant: Alphabet. One of: ascii85, z85, rfc1924.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base85_decode("87cURD]i,\"Ebo7")
    Needs: none (pure function).
    """

def base91_encode(text: str, input_encoding: str = "utf-8") -> str:
    """Encodes text with basE91 (Joachim Henke), about 23% overhead versus 33% for Base64. Uses 91 printable ASCII characters including quotes and brackets.

        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base91_encode("Hello World")
    Needs: none (pure function).
    """

def base91_decode(text: str, output: str = "text") -> str:
    """Decodes basE91 text. Fails on characters outside the 91-character alphabet. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Base91 text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base91_decode(encode.base91_encode("Hello World"))
    Needs: none (pure function).
    """

def base_encode(algo: str, text: str, input_encoding: str = "utf-8") -> str:
    """Encodes text with the named base algorithm; see bases() for names (base64, base64url, base32, base32hex, base32_crockford, zbase32, base16, hex, base36, base45, base58, base62, ascii85, z85, base85_rfc1924, base91, ...). Aliases such as b64, b32, base85 work.

        algo: Algorithm name from bases().
        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.base_encode("base58", "Hello")
    Needs: none (pure function).
    """

def base_decode(algo: str, text: str, output: str = "text") -> str:
    """Decodes text with the named base algorithm; see bases() for names. Returns text if the bytes are valid UTF-8, otherwise an error unless output is hex/base64.

        algo: Algorithm name from bases().
        text: Encoded text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.base_decode("base58", "JxF12TrwUP45BMd")
    Needs: none (pure function).
    """

def bases() -> List[str]:
    """Lists the algorithm names accepted by base_encode and base_decode, sorted alphabetically.

    Example: encode.bases()
    Needs: none (pure function).
    """

def encode_charset(text: str, charset: str, output: str = "hex", errors: str = "strict") -> str:
    """Converts text to the bytes of a legacy or Unicode charset (utf-8, utf-16le/be, utf-32, iso-8859-x, windows-125x, koi8, ibm437..., shift_jis, euc-jp, gbk, gb18030, big5, euc-kr, ...; see charsets()). The bytes are returned as hex by default (also base64, percent, decimal, binary). Unrepresentable characters are an error unless errors="replace".

        text: Text to encode.
        charset: Charset name from charsets().
        output: Byte representation of the result. One of: hex, base64, percent, decimal, binary.
        errors: "strict" or "replace". One of: strict, replace.

    Example: encode.encode_charset("café", "windows-1252")
    Needs: none (pure function).
    """

def decode_charset(data: str, charset: str, input: str = "hex", errors: str = "strict") -> str:
    """Decodes bytes written as hex (default; also base64, percent, decimal, binary) using a charset from charsets() and returns the text. Invalid byte sequences are an error unless errors="replace", which substitutes U+FFFD.

        data: Bytes in the representation named by input.
        charset: Charset name from charsets().
        input: Representation of data. One of: hex, base64, percent, decimal, binary.
        errors: "strict" or "replace". One of: strict, replace.

    Example: encode.decode_charset("63 61 66 e9", "windows-1252")
    Needs: none (pure function).
    """

def charsets() -> List[str]:
    """Lists the canonical charset names accepted by encode_charset, decode_charset, text_to_bits and bits_to_text. Common aliases (latin1, cp1252, sjis, utf8, ...) are accepted too, as are other WHATWG labels.

    Example: encode.charsets()
    Needs: none (pure function).
    """

def mojibake_fix(text: str) -> str:
    """Repairs UTF-8 text that was wrongly decoded as windows-1252/latin-1 ("cafÃ©" becomes "café", "â€™" becomes an apostrophe), repeating for double-encoded text. Only runs of non-ASCII characters that form valid UTF-8 are changed; other text is untouched.

        text: Garbled text.

    Example: encode.mojibake_fix("cafÃ©")
    Needs: none (pure function).
    """

def mojibake_try_all(text: str) -> Dict[str, str]:
    """Tries to repair garbled text assuming each common source charset (windows-125x, iso-8859-x, koi8, cp437/850/852/866, macintosh) and returns a dict of charset to repaired text, including double-encoding rounds ("windows-1252 x2"). Empty when nothing can be repaired.

        text: Garbled text.

    Example: encode.mojibake_try_all("РџСЂРёРІРµС‚")
    Needs: none (pure function).
    """

def detect_charset(data_hex: str) -> List[str]:
    """Guesses the charset of bytes given as hex (separators and 0x tolerated), best guess first: BOMs, plain ASCII, valid UTF-8, interleaved NULs (UTF-16 without BOM), strict multibyte CJK decodes and single-byte statistics. A heuristic, not a guarantee.

        data_hex: Bytes as hex, e.g. "63 61 66 c3 a9".

    Example: encode.detect_charset("63 61 66 c3 a9")
    Needs: none (pure function).
    """

def normalize_utf8(text: str, form: str = "nfc") -> str:
    """Normalizes Unicode text to NFC (default), NFD, NFKC or NFKD, and replaces invalid UTF-8 bytes with U+FFFD. NFC composes (e + combining acute becomes é); NFKC/NFKD also fold compatibility characters such as ligatures and full-width forms.

        text: Text to normalize.
        form: Normalization form. One of: nfc, nfd, nfkc, nfkd.

    Example: encode.normalize_utf8("é", form="nfc")
    Needs: none (pure function).
    """

def js_escape(text: str) -> str:
    """Legacy JavaScript escape(): keeps A-Z a-z 0-9 @ * _ + - . / and writes other UTF-16 code units as %XX (below 256) or %uXXXX, upper-case hex. Not the same as URL encoding; deprecated but still seen in old cookies and scripts.

        text: Text to escape.

    Example: encode.js_escape("café 😀")
    Needs: none (pure function).
    """

def js_unescape(text: str) -> str:
    """Legacy JavaScript unescape(): decodes %XX (Latin-1 code unit) and %uXXXX sequences, joining surrogate pairs. Invalid sequences stay literal. Unlike url_decode, %E9 means é, not an invalid UTF-8 byte.

        text: Escaped text.

    Example: encode.js_unescape("caf%E9 %uD83D%uDE00")
    Needs: none (pure function).
    """

def json_escape(text: str, ascii_only: bool = False) -> str:
    """Escapes text for use inside a JSON string (without the surrounding quotes): \" \\ \n \t \r, other control characters as \u00XX. Non-ASCII is kept unless ascii_only=true, which writes \uXXXX with surrogate pairs. < > & are not escaped.

        text: Text to escape.
        ascii_only: Escape every non-ASCII character as \uXXXX.

    Example: encode.json_escape("line1\n\"quoted\"")
    Needs: none (pure function).
    """

def json_unescape(text: str) -> str:
    """Decodes JSON string escapes (\n \t \" \\ \/ \b \f \r \uXXXX with surrogate pairs). A surrounding pair of double quotes is removed. Fails on invalid escapes such as \q or truncated \u sequences.

        text: Escaped JSON string contents.

    Example: encode.json_unescape("line1\\nline2 \\u00e9")
    Needs: none (pure function).
    """

def json_quote(text: str, ascii_only: bool = False) -> str:
    """Returns the text as a complete JSON string literal, including the surrounding double quotes, with all required escapes. ascii_only=true escapes non-ASCII as \uXXXX.

        text: Text to quote.
        ascii_only: Escape every non-ASCII character.

    Example: encode.json_quote("say \"hi\"")
    Needs: none (pure function).
    """

def c_escape(text: str, control: str = "octal", non_ascii: str = "keep", quote: bool = False) -> str:
    """Escapes text for a C/C++ string literal: \\ \" \n \t \r \a \b \v \f, other control characters as octal (\001) or hex (\x01). non_ascii controls what happens to characters above 127: keep, octal or hex (UTF-8 bytes) or unicode (\uXXXX/\UXXXXXXXX). quote=true adds surrounding quotes.

        text: Text to escape.
        control: Style for other control characters. One of: octal, hex.
        non_ascii: How to treat non-ASCII characters. One of: keep, octal, hex, unicode.
        quote: Wrap the result in double quotes.

    Example: encode.c_escape("tab\there \"q\"")
    Needs: none (pure function).
    """

def c_unescape(text: str, output: str = "text") -> str:
    """Decodes C-style backslash escapes: \n \t \r \a \b \f \v \e \\ \' \" \?, octal \NNN (1-3 digits), hex \xHH, \uXXXX and \UXXXXXXXX. \x and octal produce raw bytes, so the result must be valid UTF-8 unless output is hex/base64. Unknown escapes are kept as written.

        text: Text with escapes.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.c_unescape("tab\\there \\x41\\101")
    Needs: none (pure function).
    """

def go_quote(text: str, ascii_only: bool = False) -> str:
    """Quotes text as a Go double-quoted string literal (strconv.Quote), escaping control characters, quotes and invalid UTF-8. ascii_only=true escapes every non-ASCII character as \u or \U.

        text: Text to quote.
        ascii_only: Escape non-ASCII characters.

    Example: encode.go_quote("héllo\n")
    Needs: none (pure function).
    """

def go_unquote(text: str) -> str:
    """Parses a Go string literal ("double-quoted", `raw` or 'c' character) into its value (strconv.Unquote). Input without quotes is treated as the contents of a double-quoted string. Fails on invalid escapes.

        text: Go string literal.

    Example: encode.go_unquote("\"h\\u00e9llo\\n\"")
    Needs: none (pure function).
    """

def python_repr(text: str) -> str:
    """Returns the text as Python would print repr(str): single quotes (double quotes if the text contains ' but no "), backslash escapes for control characters, non-printable characters as \xNN, \uNNNN or \UNNNNNNNN; printable Unicode is kept.

        text: Text to represent.

    Example: encode.python_repr("it's\ta test")
    Needs: none (pure function).
    """

def python_unrepr(text: str) -> str:
    """Parses a Python string literal ('...', "...", triple-quoted, with optional r/u prefix) or bare escaped contents into text: \n \t \xNN \uNNNN \UNNNNNNNN octal and line continuations. Raw (r) literals are returned verbatim.

        text: Python string literal.

    Example: encode.python_unrepr("'it\\'s\\ta test'")
    Needs: none (pure function).
    """

def unicode_escape(text: str, all: bool = False, style: str = "utf16", upper: bool = False) -> str:
    """Writes non-ASCII characters as \uXXXX escapes; characters above U+FFFF become surrogate pairs (\uD83D\uDE00, JavaScript/Java/JSON style) or \UXXXXXXXX with style=python. all=true escapes ASCII too; upper=true uses upper-case hex digits.

        text: Text to escape.
        all: Escape every character, ASCII included.
        style: How to write characters above U+FFFF. One of: utf16, python.
        upper: Upper-case hex digits.

    Example: encode.unicode_escape("héllo 😀")
    Needs: none (pure function).
    """

def unicode_unescape(text: str) -> str:
    """Decodes \uXXXX (surrogate pairs are joined), \u{1F600}, \UXXXXXXXX, \xHH and U+XXXX sequences to characters. Lone surrogates become U+FFFD; anything else is left as written.

        text: Text with Unicode escapes.

    Example: encode.unicode_unescape("h\\u00e9llo \\ud83d\\ude00")
    Needs: none (pure function).
    """

def unicode_codepoint_escape(text: str, all: bool = False, upper: bool = True) -> str:
    """Writes non-ASCII characters as ES6/Rust style \u{1F600} code point escapes (no surrogate pairs). all=true escapes ASCII too; upper=true uses upper-case hex.

        text: Text to escape.
        all: Escape every character.
        upper: Upper-case hex digits.

    Example: encode.unicode_codepoint_escape("a😀é")
    Needs: none (pure function).
    """

def unicode_codepoint_unescape(text: str) -> str:
    """Decodes \u{1F600} code point escapes (and the other forms handled by unicode_unescape) to characters.

        text: Text with \u{...} escapes.

    Example: encode.unicode_codepoint_unescape("a\\u{1F600}")
    Needs: none (pure function).
    """

def css_escape(text: str) -> str:
    """Escapes text for use as a CSS identifier or selector part, following CSS.escape(): special characters get a backslash, control characters and a leading digit become hex escapes (\31 ), NUL becomes U+FFFD.

        text: Text to escape.

    Example: encode.css_escape("1st .item:hover")
    Needs: none (pure function).
    """

def css_unescape(text: str) -> str:
    """Decodes CSS escapes: \ followed by 1-6 hex digits (plus one optional whitespace) is a code point, \ followed by any other character is that character, \ at a line end is dropped.

        text: CSS escaped text.

    Example: encode.css_unescape("\\31 st\\ .item")
    Needs: none (pure function).
    """

def regex_escape(text: str) -> str:
    """Escapes all regular expression metacharacters (\ . + * ? ( ) | [ ] { } ^ $) with a backslash so the text matches literally (Go regexp.QuoteMeta; also valid for PCRE, JavaScript and Python).

        text: Text to escape.

    Example: encode.regex_escape("a.b*c(d)")
    Needs: none (pure function).
    """

def regex_unescape(text: str) -> str:
    """Removes the backslash in front of punctuation (\. \* \( \\ ...), reversing regex_escape. Escapes of letters and digits (\d, \n, \1) are left alone because they are not literal escapes.

        text: Escaped regex text.

    Example: encode.regex_unescape("a\\.b\\*c\\(d\\)")
    Needs: none (pure function).
    """

def sql_escape(text: str, dialect: str = "standard") -> str:
    """Escapes text for a single-quoted SQL string literal: doubles single quotes (standard SQL, PostgreSQL, SQLite, SQL Server). dialect="mysql" also backslash-escapes \ " NUL newline CR and Ctrl-Z. Prefer parameterized queries over escaping.

        text: Text to escape.
        dialect: SQL dialect. One of: standard, mysql.

    Example: encode.sql_escape("O'Brien")
    Needs: none (pure function).
    """

def sql_unescape(text: str, dialect: str = "standard") -> str:
    """Reverses sql_escape: turns doubled single quotes back into one (standard) and, with dialect="mysql", decodes backslash escapes as well.

        text: Escaped SQL string contents.
        dialect: SQL dialect. One of: standard, mysql.

    Example: encode.sql_unescape("O''Brien")
    Needs: none (pure function).
    """

def sql_quote(text: str, dialect: str = "standard") -> str:
    """Escapes the text and wraps it in single quotes as a SQL string literal ('O''Brien'). dialect="mysql" uses backslash escapes in addition.

        text: Text to quote.
        dialect: SQL dialect. One of: standard, mysql.

    Example: encode.sql_quote("O'Brien")
    Needs: none (pure function).
    """

def sh_quote(text: str, always: bool = False) -> str:
    """Quotes text so a POSIX shell reads it as one literal word: wraps it in single quotes and rewrites embedded ' as '\''. Text made only of safe characters is returned unchanged unless always=true; empty text gives ''.

        text: Text to quote.
        always: Quote even when not necessary.

    Example: encode.sh_quote("it's a file.txt")
    Needs: none (pure function).
    """

def powershell_quote(text: str) -> str:
    """Quotes text as a PowerShell single-quoted (verbatim) string: wraps it in ' and doubles every ' and typographic single quote, so no variable expansion or escapes apply.

        text: Text to quote.

    Example: encode.powershell_quote("it's $HOME")
    Needs: none (pure function).
    """

def cmd_quote(text: str, for_cmd: bool = False) -> str:
    """Quotes an argument for the Windows command line using the CommandLineToArgvW rules: wrapped in double quotes when it has spaces or quotes, with backslashes before quotes doubled and quotes written as \". for_cmd=true also escapes cmd.exe metacharacters ( ) % ! ^ " & | < > with ^ for use at a cmd prompt.

        text: Argument to quote.
        for_cmd: Also caret-escape cmd.exe metacharacters.

    Example: encode.cmd_quote("C:\\Program Files\\App\\a \"b\".exe")
    Needs: none (pure function).
    """

def csv_escape(text: str, delimiter: str = ",", always: bool = False) -> str:
    """Quotes one CSV field when needed (it contains the delimiter, a double quote or a line break) and doubles embedded quotes. always=true quotes every field; delimiter can be ";" or a tab.

        text: Field value.
        delimiter: Field delimiter.
        always: Always add quotes.

    Example: encode.csv_escape("say \"hi\", ok")
    Needs: none (pure function).
    """

def csv_unescape(text: str) -> str:
    """Reverses csv_escape for one field: removes the surrounding double quotes and turns doubled quotes into single quotes. Text without surrounding quotes is returned unchanged.

        text: Quoted CSV field.

    Example: encode.csv_unescape("\"say \"\"hi\"\", ok\"")
    Needs: none (pure function).
    """

def quoted_printable_encode(text: str, input_encoding: str = "utf-8") -> str:
    """Encodes text as MIME quoted-printable (RFC 2045): non-ASCII bytes and '=' become =XX, trailing spaces are protected, lines are soft-wrapped with '=' before 76 columns and newlines become CRLF.

        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.quoted_printable_encode("café = good")
    Needs: none (pure function).
    """

def quoted_printable_decode(text: str, output: str = "text") -> str:
    """Decodes MIME quoted-printable: =XX bytes are decoded, soft line breaks (= at the end of a line) are removed. Malformed = sequences stay literal. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.

        text: Quoted-printable text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.quoted_printable_decode("caf=C3=A9 =3D good")
    Needs: none (pure function).
    """

def uuencode(text: str, filename: str = "data", mode: str = "644", input_encoding: str = "utf-8") -> str:
    """Encodes text in the classic Unix uuencode format: a "begin MODE NAME" header, lines of up to 45 bytes with a length prefix, a terminating backquote line and "end".

        text: Text to encode.
        filename: File name in the header.
        mode: Octal permission string in the header.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.uuencode("Hello", filename="hello.txt")
    Needs: none (pure function).
    """

def uudecode(text: str, output: str = "text") -> str:
    """Decodes uuencoded data. The "begin"/"end" lines are optional. Returns text if the bytes are valid UTF-8, otherwise an error unless output is hex/base64.

        text: UUencoded text.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.uudecode(encode.uuencode("Hello"))
    Needs: none (pure function).
    """

def unescape_auto(text: str) -> str:
    """Applies every common unescaping that matches, in order: JS %uXXXX, backslash escapes (\n \xHH \uXXXX octal), HTML entities, then %XX percent escapes (only if the result is valid UTF-8). Useful when the escaping type is unknown; text without any escapes is returned unchanged.

        text: Escaped text.

    Example: encode.unescape_auto("caf\\u00e9 &amp; %3Cb%3E")
    Needs: none (pure function).
    """

def html_escape_all(text: str, hex: bool = False) -> str:
    """Escapes markup characters and every non-ASCII character: named entities where an HTML 4 name exists (&eacute; &euro; &hearts;), numeric references otherwise (&#128512;, or &#x1F600; with hex=true). Output is pure ASCII.

        text: Text to escape.
        hex: Use hexadecimal numeric references for characters without a name.

    Example: encode.html_escape_all("café © 😀")
    Needs: none (pure function).
    """

def html_numeric_entities(text: str, hex: bool = False, all: bool = False) -> str:
    """Converts characters to numeric character references (&#233; decimal, &#xE9; with hex=true). By default converts non-ASCII characters and the markup characters & < > " '; all=true converts every character including letters.

        text: Text to convert.
        hex: Hexadecimal references.
        all: Convert every character, even plain ASCII.

    Example: encode.html_numeric_entities("Hi é", hex=true)
    Needs: none (pure function).
    """

def xml_escape(text: str) -> str:
    """Escapes & < > " ' as &amp; &lt; &gt; &quot; &apos; for XML text and attributes, and removes characters that are not allowed in XML 1.0 (most control characters).

        text: Text to escape.

    Example: encode.xml_escape("<a b='1'>x & y</a>")
    Needs: none (pure function).
    """

def xml_unescape(text: str) -> str:
    """Decodes the five predefined XML entities (&amp; &lt; &gt; &quot; &apos;) and numeric references (&#65; &#x41;). Other named entities and invalid numbers are left untouched; use html_unescape for HTML names.

        text: Text containing XML entities.

    Example: encode.xml_unescape("&lt;a&gt; &amp; &#65;")
    Needs: none (pure function).
    """

def strip_html_tags(text: str, keep_breaks: bool = True) -> str:
    """Removes HTML tags and comments, drops <script> and <style> contents, decodes entities and keeps the text. keep_breaks=true (default) turns block tags such as <p>, <br>, <li>, <div> into line breaks. Plain text without tags is returned trimmed.

        text: HTML text.
        keep_breaks: Insert line breaks at block-level tags.

    Example: encode.strip_html_tags("<p>Hello <b>world</b> &amp; all</p>")
    Needs: none (pure function).
    """

def html_entity_name(char: str) -> str:
    """Returns the HTML entity for a character: the named form ("&copy;", "&amp;", "&euro;") when an HTML 4 name exists, otherwise the decimal numeric reference ("&#128512;"). Uses the first character if several are given; empty input is an error.

        char: A single character.

    Example: encode.html_entity_name("©")
    Needs: none (pure function).
    """

def punycode_encode(text: str, prefix: bool = False) -> str:
    """Raw RFC 3492 Punycode of one string (a domain label): "bücher" gives "bcher-kva". No case folding or normalization; use idna_to_ascii for whole domains. prefix=true prepends "xn--". Pure-ASCII input yields the text plus a trailing hyphen.

        text: Label to encode.
        prefix: Prepend the ACE prefix xn--.

    Example: encode.punycode_encode("bücher")
    Needs: none (pure function).
    """

def punycode_decode(text: str) -> str:
    """Raw RFC 3492 Punycode decode of one label; a leading "xn--" is removed first. "bcher-kva" gives "bücher". Fails on invalid digits, overflow or truncated input.

        text: Punycode label, with or without xn--.

    Example: encode.punycode_decode("xn--bcher-kva")
    Needs: none (pure function).
    """

def idna_to_ascii(domain: str, profile: str = "lookup") -> str:
    """Converts an internationalized domain name to ASCII: each non-ASCII label becomes xn-- punycode, with IDNA mapping (lower-casing, normalization) and validation according to the profile (lookup by default). "münchen.de" gives "xn--mnchen-3ya.de".

        domain: Domain name such as "bücher.example".
        profile: IDNA profile; punycode skips mapping and validation. One of: lookup, registration, display, punycode.

    Example: encode.idna_to_ascii("münchen.de")
    Needs: none (pure function).
    """

def idna_to_unicode(domain: str, profile: str = "lookup") -> str:
    """Converts a domain with xn-- labels to its Unicode form: "xn--mnchen-3ya.de" gives "münchen.de". Labels without xn-- are kept. If strict validation fails the labels are still decoded with raw punycode when possible.

        domain: Domain name such as "xn--bcher-kva.example".
        profile: IDNA profile; punycode skips validation. One of: lookup, registration, display, punycode.

    Example: encode.idna_to_unicode("xn--mnchen-3ya.de")
    Needs: none (pure function).
    """

def email_to_ascii(email: str, profile: str = "lookup") -> str:
    """Converts the domain part of an email address to ASCII punycode (user@bücher.de gives user@xn--bcher-kva.de). The local part before the last @ is left unchanged, since internationalized local parts need SMTPUTF8.

        email: Email address.
        profile: IDNA profile. One of: lookup, registration, display, punycode.

    Example: encode.email_to_ascii("info@bücher.de")
    Needs: none (pure function).
    """

def email_to_unicode(email: str, profile: str = "lookup") -> str:
    """Converts the xn-- domain part of an email address to Unicode (user@xn--bcher-kva.de gives user@bücher.de). The local part before the last @ is left unchanged.

        email: Email address.
        profile: IDNA profile. One of: lookup, registration, display, punycode.

    Example: encode.email_to_unicode("info@xn--bcher-kva.de")
    Needs: none (pure function).
    """

def text_to_binary(text: str, separator: str = " ", input_encoding: str = "utf-8") -> str:
    """Writes the UTF-8 bytes of the text as 8-bit binary groups ("01001000 01101001"). separator sits between groups; use text_to_bits for UTF-16/UTF-32 or other charsets.

        text: Text to convert.
        separator: Text between 8-bit groups.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.text_to_binary("Hi")
    Needs: none (pure function).
    """

def binary_to_text(bits: str, output: str = "text") -> str:
    """Reads binary digits (8-bit groups, optional spaces, commas or 0b prefixes; 7-bit groups when every group has 7 digits) and returns the text. Fails when the bit count is not a multiple of 8 or the bytes are not UTF-8.

        bits: Binary digits, e.g. "01001000 01101001".
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.binary_to_text("01001000 01101001")
    Needs: none (pure function).
    """

def text_to_octal(text: str, separator: str = " ", prefix: str = "", input_encoding: str = "utf-8") -> str:
    """Writes each UTF-8 byte of the text as a 3-digit octal number ("110 151"), separated by separator. Set prefix to "\\" for C-style escapes.

        text: Text to convert.
        separator: Text between bytes.
        prefix: Prefix before each byte.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.text_to_octal("Hi")
    Needs: none (pure function).
    """

def octal_to_text(text: str, output: str = "text") -> str:
    """Reads octal byte values separated by spaces, commas or backslashes ("110 151", "\\110\\151") or a run of 3-digit groups, and returns the text. Values above 377 are rejected.

        text: Octal bytes.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.octal_to_text("110 151")
    Needs: none (pure function).
    """

def text_to_decimal_bytes(text: str, separator: str = " ", input_encoding: str = "utf-8") -> str:
    """Writes each UTF-8 byte of the text as a decimal number 0-255 ("72 101 108"), joined by separator. Equivalent to the byte values of the string.

        text: Text to convert.
        separator: Text between bytes.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.text_to_decimal_bytes("Hel")
    Needs: none (pure function).
    """

def decimal_bytes_to_text(text: str, output: str = "text") -> str:
    """Reads decimal byte values separated by spaces, commas or other non-digits ("72 101 108") and returns the text. Values above 255 are rejected.

        text: Decimal bytes.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.decimal_bytes_to_text("72 101 108")
    Needs: none (pure function).
    """

def text_to_bytes(text: str, base: int = 16, separator: str = " ", input_encoding: str = "utf-8") -> str:
    """Writes the UTF-8 bytes of the text in base 2, 8, 10 or 16 separated by separator (binary and octal are zero-padded, hex is lower case without prefix).

        text: Text to convert.
        base: Number base: 2, 8, 10 or 16.
        separator: Text between bytes.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.text_to_bytes("Hi", base=16)
    Needs: none (pure function).
    """

def bytes_to_text(data: str, base: int = 16, output: str = "text") -> str:
    """Reads byte values written in base 2, 8, 10 or 16 (separated by spaces, commas, etc.) and returns the text. Fails on out-of-range values or invalid UTF-8 unless output is hex/base64.

        data: Byte values.
        base: Number base: 2, 8, 10 or 16.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.bytes_to_text("48 69", base=16)
    Needs: none (pure function).
    """

def text_to_bits(text: str, charset: str = "utf-8", separator: str = " ") -> str:
    """Encodes the text in the given charset (utf-8 default; utf-16le, utf-16be, utf-32le, iso-8859-1, shift_jis, ... see charsets()) and writes every byte as 8 binary digits. Characters the charset cannot represent cause an error.

        text: Text to convert.
        charset: Charset name from charsets().
        separator: Text between bytes.

    Example: encode.text_to_bits("Hi", charset="utf-16le")
    Needs: none (pure function).
    """

def bits_to_text(bits: str, charset: str = "utf-8", errors: str = "strict") -> str:
    """Reads binary digits (groups of 8) and decodes the bytes in the given charset (utf-8 default; utf-16le, utf-32be, windows-1252, ...). Invalid sequences cause an error unless errors="replace".

        bits: Binary digits.
        charset: Charset name from charsets().
        errors: "strict" or "replace" (use U+FFFD). One of: strict, replace.

    Example: encode.bits_to_text("0100100000000000", charset="utf-16le")
    Needs: none (pure function).
    """

def url_encode_all(text: str, input_encoding: str = "utf-8") -> str:
    """Percent-encodes every byte of the text, including letters and digits ("A" becomes %41). Useful for obfuscation tests and WAF checks; url_decode reverses it.

        text: Text to encode.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.url_encode_all("admin")
    Needs: none (pure function).
    """

def double_url_decode(text: str, times: int = 2, plus_as_space: bool = False, output: str = "text") -> str:
    """Percent-decodes up to `times` rounds (default 2), stopping early when nothing changes, for double or triple encoded input such as %2527 or %253Cscript%253E. The final bytes must be valid UTF-8 unless output is hex/base64.

        text: Percent-encoded text.
        times: Maximum number of rounds (1-10).
        plus_as_space: Treat + as a space.
        output: Form of the result: "text" (default, fails if the bytes are not valid UTF-8), or a byte form "hex", "base64", "percent", "decimal", "binary". One of: text, hex, base64, percent, decimal, binary.

    Example: encode.double_url_decode("%253Cscript%253E")
    Needs: none (pure function).
    """

def query_string_parse(qs: str) -> Dict[str, str]:
    """Parses a query string (leading ? and #fragment ignored; a full URL is accepted and its query used) into a dict. Keys and values are percent-decoded with + as space; repeated keys are joined with ","; a key without = maps to "". Pairs split on & or ;.

        qs: Query string such as "a=1&b=two&a=3".

    Example: encode.query_string_parse("?a=1&b=two%20words&a=3")
    Needs: none (pure function).
    """

def query_string_build(params: Dict[str, str], prefix: bool = False) -> str:
    """Builds a query string from a dict: keys sorted alphabetically, keys and values form-encoded (space becomes +). prefix=true adds a leading ?. An empty dict gives an empty string.

        params: Key/value pairs.
        prefix: Add a leading ?.

    Example: encode.query_string_build({"q": "hello world", "page": "2"})
    Needs: none (pure function).
    """

def form_urlencode(params: Dict[str, str]) -> str:
    """Encodes a dict as application/x-www-form-urlencoded (sorted keys, space as +). Same as query_string_build without a prefix.

        params: Key/value pairs.

    Example: encode.form_urlencode({"name": "Ada Lovelace", "lang": "en"})
    Needs: none (pure function).
    """

def form_urldecode(text: str) -> Dict[str, str]:
    """Decodes an application/x-www-form-urlencoded body into a dict (+ is a space, repeated keys joined with ","). Same as query_string_parse.

        text: Form body such as "name=Ada+Lovelace&lang=en".

    Example: encode.form_urldecode("name=Ada+Lovelace&lang=en")
    Needs: none (pure function).
    """

def data_uri_encode(text: str, mime: str = "text/plain", base64: bool = True, charset: str = "", input_encoding: str = "utf-8") -> str:
    """Builds a data: URI (RFC 2397) from text, e.g. data:text/plain;base64,SGk=. base64=false percent-encodes the data instead; charset (e.g. "utf-8") adds a ;charset= parameter. Use input_encoding=base64 or hex to embed binary data.

        text: Content.
        mime: Media type.
        base64: Use base64 (otherwise percent-encoding).
        charset: Optional charset parameter, empty for none.
        input_encoding: How to read the text argument as bytes: "utf-8" (default, the text itself), or a byte form: "hex", "base64", "percent", "decimal" (space separated 0-255), "binary". One of: utf-8, hex, base64, percent, decimal, binary.

    Example: encode.data_uri_encode("<h1>Hi</h1>", mime="text/html")
    Needs: none (pure function).
    """

def data_uri_decode(uri: str) -> Dict[str, str]:
    """Parses a data: URI into a dict with mime, charset, base64 ("true"/"false"), size (bytes), text (decoded text, empty when the data is binary) and data_base64 (the raw bytes). Fails when the URI is not a data URI or the payload is malformed.

        uri: A data: URI.

    Example: encode.data_uri_decode("data:text/plain;base64,SGVsbG8=")
    Needs: none (pure function).
    """

def jwt_decode(token: str) -> Dict[str, str]:
    """Decodes a JSON Web Token WITHOUT verifying its signature: returns header and payload as compact JSON, the raw signature, and alg, typ, kid, iss, sub, aud, jti when present. Numeric exp, iat and nbf are returned as given plus a human UTC time (exp_utc, iat_utc, nbf_utc). A leading "Bearer " is ignored; encrypted JWTs (5 parts) are rejected.

        token: The JWT, header.payload.signature.

    Example: encode.jwt_decode("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyfQ.sig")
    Needs: none (pure function).
    """

def jwt_parts(token: str) -> List[str]:
    """Splits a JWT and returns [header JSON, payload JSON, signature] as three strings, the first two base64url-decoded and compacted. No signature verification.

        token: The JWT.

    Example: encode.jwt_parts("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhIn0.sig")
    Needs: none (pure function).
    """
