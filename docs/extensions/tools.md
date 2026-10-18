# Built-in tools API

Zagros ships a large library of pure utility functions. The same functions power the launcher's tool families (Tab opens the whole family for what you typed) and are available to extensions as modules, with no permission needed.

| Module | What it covers | Functions |
|---|---|---|
| [`cipher`](tools-cipher.md) | AES, ChaCha20, RSA, Ed25519, XOR and classic ciphers, each with a decrypt counterpart. | 76 |
| [`dt`](tools-dt.md) | Parse, convert and calculate dates, times, time zones, durations and cron schedules. | 66 |
| [`encode`](tools-encode.md) | Base16/32/58/62/64/85 and more, charsets, URL, HTML, punycode, escape and unescape, JWT decode. | 117 |
| [`extract`](tools-extract.md) | Pull emails, URLs, domains, IPs, numbers, phone numbers, dates and more out of any text. | 62 |
| [`hash`](tools-hash.md) | Digests (MD5, SHA-1/2/3, BLAKE, CRC, Adler, FNV, xxHash), HMAC, PBKDF2, bcrypt, scrypt, Argon2, TOTP. | 112 |
| [`lines`](tools-lines.md) | Work on multi-line text: sort, unique, filter, join, columns, CSV and list conversions. | 60 |
| [`math`](tools-math.md) | An expression evaluator, percentages, statistics, algebra, geometry and finance helpers. | 116 |
| [`net`](tools-net.md) | CIDR and subnet calculators, IP and MAC tools, ports, HTTP status, MIME types, URLs and domains. | 92 |
| [`num`](tools-num.md) | Base conversion, the hex and binary calculator, bit operations, IEEE 754, roman numerals, number words, primes and checksums. | 111 |
| [`rand`](tools-rand.md) | Random numbers, dice, strings, passwords, passphrases, colors and picks, with optional seeds. | 55 |
| [`regex`](tools-regex.md) | Test, match, replace and explain regular expressions (Go RE2 syntax). | 25 |
| [`text`](tools-text.md) | Change, measure and analyse strings: case styles, palindromes, lengths, normalisation, similarity, Unicode. | 107 |
| [`uuid`](tools-uuid.md) | UUID v1-v7, ULID, KSUID, Snowflake, ObjectID, nanoid: generate, parse, convert. | 37 |

```python
def on_query(ctx, query):
    return [item(id = "h", title = hash.sha256(query), action = action_copy(hash.sha256(query)))]
```

Compose them with the host modules: read text with `clipboard.get()` (needs `clipboard.read`), transform it with these functions, and write it back with `clipboard.set(...)` (needs `clipboard.write`).

The exact signatures are in the API stubs (`zagros-stubs/zagros/<module>.pyi`); coding agents can also search them through the Zagros MCP server.
