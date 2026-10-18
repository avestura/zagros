# Hashes (`hash`)

Digests (MD5, SHA-1/2/3, BLAKE, CRC, Adler, FNV, xxHash), HMAC, PBKDF2, bcrypt, scrypt, Argon2, TOTP.

No permission is needed: these are pure functions. Generated from the tools registry.

112 functions. Use them as `hash.<name>(...)` in any extension.

## hash.algorithms

`hash.algorithms() -> List[str]`

Returns the sorted names accepted by hash.hash(), hash.hash_many() and hash.hash_length_bits(). HMAC accepts the subset that has a streaming form (md5, sha1, sha2, sha3, blake2, ripemd160, sm3, keccak).

Returns `List[str]`. Example: `hash.algorithms()`

## hash.argon2_verify

`hash.argon2_verify(password: str, phc: str) -> bool`

Checks a password against a PHC string produced by argon2id_hash (or any $argon2id$ / $argon2i$ v=19 hash) and returns true on a match, false on a wrong password. Malformed strings, or parameters above the caps (time 10, 256 MiB), are errors.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password to check. |
| `phc` | `str` | PHC string such as $argon2id$v=19$m=65536,t=3,p=4$salt$hash. |

Returns `bool`. Example: `hash.argon2_verify("hunter2", "$argon2id$v=19$m=65536,t=3,p=4$...$...")`

## hash.argon2i

`hash.argon2i(password: str, salt: str, time: int = 3, memory_kib: int = 65536, threads: int = 4, length: int = 32, salt_encoding: str = "utf-8", output: str = "hex") -> str`

Derives a raw key with argon2i (RFC 9106, version 1.3). Defaults time=3, memory_kib=65536 (64 MiB), threads=4, length=32; caps time 10, memory 256 MiB, threads 64. Password is UTF-8, salt follows salt_encoding (use at least 8 bytes); output hex by default. For a self-describing stored hash use argon2id_hash.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password (UTF-8). |
| `salt` | `str` | Salt per salt_encoding, at least 8 bytes recommended. |
| `time` | `int` (default `3`) | Passes over memory, 1 to 10 (default 3). |
| `memory_kib` | `int` (default `65536`) | Memory in KiB, 8 to 262144 (default 65536 = 64 MiB). |
| `threads` | `int` (default `4`) | Parallelism 1 to 64 (default 4). |
| `length` | `int` (default `32`) | Tag length in bytes, 4 to 1024 (default 32). |
| `salt_encoding` | `str` (default `"utf-8"`) | How the salt is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.argon2i("password", "somesalt", time=2, memory_kib=19456, threads=1)`

## hash.argon2id

`hash.argon2id(password: str, salt: str, time: int = 3, memory_kib: int = 65536, threads: int = 4, length: int = 32, salt_encoding: str = "utf-8", output: str = "hex") -> str`

Derives a raw key with argon2id (RFC 9106, version 1.3). Defaults time=3, memory_kib=65536 (64 MiB), threads=4, length=32; caps time 10, memory 256 MiB, threads 64. Password is UTF-8, salt follows salt_encoding (use at least 8 bytes); output hex by default. For a self-describing stored hash use argon2id_hash.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password (UTF-8). |
| `salt` | `str` | Salt per salt_encoding, at least 8 bytes recommended. |
| `time` | `int` (default `3`) | Passes over memory, 1 to 10 (default 3). |
| `memory_kib` | `int` (default `65536`) | Memory in KiB, 8 to 262144 (default 65536 = 64 MiB). |
| `threads` | `int` (default `4`) | Parallelism 1 to 64 (default 4). |
| `length` | `int` (default `32`) | Tag length in bytes, 4 to 1024 (default 32). |
| `salt_encoding` | `str` (default `"utf-8"`) | How the salt is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.argon2id("password", "somesalt", time=2, memory_kib=19456, threads=1)`

## hash.argon2id_hash

`hash.argon2id_hash(password: str, time: int = 3, memory_kib: int = 65536, threads: int = 4, length: int = 32) -> str`

Hashes a password with Argon2id and a fresh random 16-byte salt, returning the standard PHC string ($argon2id$v=19$m=...,t=...,p=...$salt$hash). Defaults time=3, memory_kib=65536, threads=4, length=32 (caps: time 10, 256 MiB, 64 threads). Not deterministic; verify with argon2_verify.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password (UTF-8). |
| `time` | `int` (default `3`) | Passes over memory, 1 to 10 (default 3). |
| `memory_kib` | `int` (default `65536`) | Memory in KiB, 8 to 262144 (default 65536 = 64 MiB). |
| `threads` | `int` (default `4`) | Parallelism 1 to 64 (default 4). |
| `length` | `int` (default `32`) | Tag length in bytes, 4 to 1024 (default 32). |

Returns `str`. Example: `hash.argon2id_hash("hunter2")`

## hash.bcrypt_hash

`hash.bcrypt_hash(password: str, cost: int = 10) -> str`

Hashes a password with bcrypt and a random salt, returning the 60-character $2a$ string. cost is 4 to 14 (default 10, each step doubles the time); passwords longer than 72 bytes are an error rather than silently truncated.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password, at most 72 bytes. |
| `cost` | `int` (default `10`) | Work factor 4 to 14 (default 10). |

Returns `str`. Example: `hash.bcrypt_hash("hunter2", cost=10)`

## hash.bcrypt_verify

`hash.bcrypt_verify(password: str, hash: str) -> bool`

Checks a password against a bcrypt hash ($2a$, $2b$ or $2y$) and returns true on a match, false on a wrong password. A malformed hash string is an error.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password to check. |
| `hash` | `str` | bcrypt hash string. |

Returns `bool`. Example: `hash.bcrypt_verify("hunter2", "$2a$10$...")`

## hash.checksum_compare

`hash.checksum_compare(a: str, b: str) -> bool`

Compares two hex digests in constant time after lowercasing and dropping whitespace, ':' and '-' separators and any 0x prefix; returns true when equal. Non-hex input falls back to an exact constant-time string comparison.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First checksum. |
| `b` | `str` | Second checksum. |

Returns `bool`. Example: `hash.checksum_compare("D41D8CD9 8F00B204", "d41d8cd98f00b204")`

## hash.crc_custom

`hash.crc_custom(text: str, width: int, poly: str, init: str = "0", xorout: str = "0", reflect_in: bool = False, reflect_out: bool = False, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes a CRC with explicit Rocksoft-model parameters: width in bits (8 to 64), polynomial, initial value and final xor as hex strings, and the input/output reflection flags. Useful for protocols not covered by the named variants; the result is fixed-width hex by default.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `width` | `int` | CRC width in bits, 8 to 64. |
| `poly` | `str` | Polynomial in normal (MSB-first) form as hex, e.g. 04C11DB7. |
| `init` | `str` (default `"0"`) | Initial register value as hex (default 0). |
| `xorout` | `str` (default `"0"`) | Final xor value as hex (default 0). |
| `reflect_in` | `bool` (default `False`) | Reflect each input byte (LSB first). |
| `reflect_out` | `bool` (default `False`) | Reflect the final register. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc_custom("123456789", 16, "1021", init="FFFF")`

## hash.hash

`hash.hash(algo: str, text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Hashes text with the named algorithm (see algorithms(): md5, sha256, sha3_512, blake2b_256, blake3, crc32, xxhash64 and many more; '-' and '/' in names are accepted, so 'SHA-256' works). Algorithms with extra options (seed, key, length) use their defaults here; call their own function to change them. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `algo` | `str` | Algorithm name, for example sha256, md5, blake2b_512, crc32c. |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hash("sha256", "hello")`

## hash.hash_all

`hash.hash_all(text: str, input_encoding: str = "utf-8", output: str = "hex") -> Dict[str, str]`

Hashes text with the common algorithms at once (md5, sha1, sha2 family, sha3, keccak256, blake2, blake3, ripemd160, crc32, adler32, fnv64a, xxhash64, murmur3_32) and returns a dict of algorithm name to digest. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `Dict[str, str]`. Example: `hash.hash_all("hello")`

## hash.hash_length_bits

`hash.hash_length_bits(algo: str) -> int`

Returns the default digest size in bits of an algorithm (md5 is 128, sha256 is 256, crc32 is 32). Variable-length ones report their default (shake128 256, shake256 512, blake3 256).

| Parameter | Type | Description |
|---|---|---|
| `algo` | `str` | Algorithm name. |

Returns `int`. Example: `hash.hash_length_bits("sha256")`

## hash.hash_many

`hash.hash_many(text: str, algos: List[str], input_encoding: str = "utf-8", output: str = "hex") -> Dict[str, str]`

Hashes text with each of the listed algorithms and returns a dict of algorithm name to digest. Unknown algorithm names are an error.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `algos` | `List[str]` | Algorithm names, for example ["md5", "sha256"]. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `Dict[str, str]`. Example: `hash.hash_many("hello", ["md5", "sha256"])`

## hash.hkdf

`hash.hkdf(algo: str, secret: str, salt: str = "", info: str = "", length: int = 32, key_encoding: str = "utf-8", salt_encoding: str = "utf-8", output: str = "hex") -> str`

Derives length bytes (1 to 1024) of key material with HKDF (RFC 5869). secret is decoded with key_encoding; salt and info with salt_encoding (utf-8, hex or base64); empty salt is allowed. Output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `algo` | `str` | Hash algorithm, e.g. sha256. |
| `secret` | `str` | Input key material. |
| `salt` | `str` (default `""`) | Optional salt. |
| `info` | `str` (default `""`) | Optional context/application info. |
| `length` | `int` (default `32`) | Output length in bytes (1 to 1024, default 32). |
| `key_encoding` | `str` (default `"utf-8"`) | How the secret is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `salt_encoding` | `str` (default `"utf-8"`) | How salt and info is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hkdf("sha256", "input key", salt="salt", info="app", length=32)`

## hash.hmac

`hash.hmac(algo: str, key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the HMAC (RFC 2104) of text under a secret key with the named hash (sha256, sha512, sha1, md5, sha3_256, blake2b_256, ... see algorithms()). Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `algo` | `str` | Hash algorithm name, e.g. sha256, sha512, sha3_256, blake2b_256. |
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac("sha256", "secret", "message")`

## hash.hmac_verify

`hash.hmac_verify(algo: str, key: str, text: str, mac: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8") -> bool`

Recomputes the HMAC of text under key and compares it to mac in constant time; returns true on a match. mac may be hex or base64 (case and whitespace of hex are ignored); an undecodable mac is simply false.

| Parameter | Type | Description |
|---|---|---|
| `algo` | `str` | Hash algorithm name, e.g. sha256. |
| `key` | `str` | Secret key. |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `mac` | `str` | Expected MAC as hex or base64. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |

Returns `bool`. Example: `hash.hmac_verify("sha256", "secret", "message", "8b5f48702995c1598c573db1e21866a9b825d4a794d169d7060a03605796360b")`

## hash.hotp

`hash.hotp(secret: str, counter: int, digits: int = 6, algo: str = "sha1") -> str`

Computes the HOTP one-time code (RFC 4226) for a base32 secret and counter. Returns a zero-padded string of `digits` digits (default 6, HMAC-SHA1).

| Parameter | Type | Description |
|---|---|---|
| `secret` | `str` | Shared secret in base32 (spaces, dashes, case and '=' padding ignored). |
| `counter` | `int` | Counter value, 0 or more. |
| `digits` | `int` (default `6`) | Code length, 1 to 10 (default 6). |
| `algo` | `str` (default `"sha1"`) | HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512. One of: sha1, sha256, sha512. |

Returns `str`. Example: `hash.hotp("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", 0)`

## hash.hotp_uri

`hash.hotp_uri(secret: str, account: str, issuer: str = "", digits: int = 6, algo: str = "sha1", counter: int = 0) -> str`

Builds the otpauth://hotp/ provisioning URI from a base32 secret, account, issuer and initial counter.

| Parameter | Type | Description |
|---|---|---|
| `secret` | `str` | Shared secret in base32 (spaces, dashes, case and '=' padding ignored). |
| `account` | `str` | Account name, e.g. an email address. |
| `issuer` | `str` (default `""`) | Service name shown in the authenticator app. |
| `digits` | `int` (default `6`) | Code length, 1 to 10 (default 6). |
| `algo` | `str` (default `"sha1"`) | HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512. One of: sha1, sha256, sha512. |
| `counter` | `int` (default `0`) | Initial counter (default 0). |

Returns `str`. Example: `hash.hotp_uri("JBSWY3DPEHPK3PXP", "alice@example.com", issuer="Example")`

## hash.pbkdf2

`hash.pbkdf2(password: str, salt: str, iterations: int = 600000, length: int = 32, algo: str = "sha256", salt_encoding: str = "utf-8", output: str = "hex") -> str`

Derives a key from a password with PBKDF2 (RFC 8018). Iterations default to 600000 and are capped at 5,000,000; length is 1 to 1024 bytes. The password is UTF-8; salt follows salt_encoding. Output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password (UTF-8). |
| `salt` | `str` | Salt, per salt_encoding. |
| `iterations` | `int` (default `600000`) | Iteration count, 1 to 5000000 (default 600000). |
| `length` | `int` (default `32`) | Derived key length in bytes, 1 to 1024 (default 32). |
| `algo` | `str` (default `"sha256"`) | Underlying hash, default sha256 (sha1, sha512, ... also work). |
| `salt_encoding` | `str` (default `"utf-8"`) | How the salt is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.pbkdf2("password", "salt", iterations=100000)`

## hash.scrypt

`hash.scrypt(password: str, salt: str, n: int = 32768, r: int = 8, p: int = 1, length: int = 32, salt_encoding: str = "utf-8", output: str = "hex") -> str`

Derives a key from a password with scrypt (RFC 7914). Defaults n=32768, r=8, p=1, length=32. Caps: n a power of two up to 2^20, memory (128*n*r bytes) at most 256 MiB, length up to 1024. Password is UTF-8, salt follows salt_encoding; output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `password` | `str` | Password (UTF-8). |
| `salt` | `str` | Salt, per salt_encoding. |
| `n` | `int` (default `32768`) | CPU/memory cost, a power of two from 2 to 1048576 (default 32768). |
| `r` | `int` (default `8`) | Block size, default 8. |
| `p` | `int` (default `1`) | Parallelization, default 1. |
| `length` | `int` (default `32`) | Derived key length in bytes, 1 to 1024 (default 32). |
| `salt_encoding` | `str` (default `"utf-8"`) | How the salt is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.scrypt("password", "NaCl", n=1024, r=8, p=16)`

## hash.totp_uri

`hash.totp_uri(secret: str, account: str, issuer: str = "", digits: int = 6, algo: str = "sha1", period: int = 30) -> str`

Builds the otpauth://totp/ provisioning URI (the content of 2FA QR codes) from a base32 secret, account and issuer. The secret is normalised to uppercase unpadded base32; period is included (default 30).

| Parameter | Type | Description |
|---|---|---|
| `secret` | `str` | Shared secret in base32 (spaces, dashes, case and '=' padding ignored). |
| `account` | `str` | Account name, e.g. an email address. |
| `issuer` | `str` (default `""`) | Service name shown in the authenticator app. |
| `digits` | `int` (default `6`) | Code length, 1 to 10 (default 6). |
| `algo` | `str` (default `"sha1"`) | HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512. One of: sha1, sha256, sha512. |
| `period` | `int` (default `30`) | Time step in seconds (default 30). |

Returns `str`. Example: `hash.totp_uri("JBSWY3DPEHPK3PXP", "alice@example.com", issuer="Example")`

## hash.totp_verify

`hash.totp_verify(secret: str, code: str, time: int = 0, period: int = 30, digits: int = 6, algo: str = "sha1", window: int = 1) -> bool`

Checks a TOTP code against a base32 secret, accepting `window` time steps before and after the given time (default 1, to absorb clock drift). Time 0 means now. Compares in constant time and returns true or false.

| Parameter | Type | Description |
|---|---|---|
| `secret` | `str` | Shared secret in base32 (spaces, dashes, case and '=' padding ignored). |
| `code` | `str` | Code to check. |
| `time` | `int` (default `0`) | Unix time in seconds; 0 (default) means now. |
| `period` | `int` (default `30`) | Time step in seconds (default 30). |
| `digits` | `int` (default `6`) | Code length, 1 to 10 (default 6). |
| `algo` | `str` (default `"sha1"`) | HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512. One of: sha1, sha256, sha512. |
| `window` | `int` (default `1`) | Accepted steps either side of the current one, 0 to 10 (default 1). |

Returns `bool`. Example: `hash.totp_verify("JBSWY3DPEHPK3PXP", "123456")`

## hash.md5

`hash.md5(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the MD5 digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.md5("hello")`

## hash.sha1

`hash.sha1(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-1 digest of text (160 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha1("hello")`

## hash.sha256

`hash.sha256(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-256 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha256("hello")`

## hash.sha512

`hash.sha512(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-512 digest of text (512 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha512("hello")`

## hash.sha3_256

`hash.sha3_256(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA3-256 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha3_256("hello")`

## hash.blake2b_256

`hash.blake2b_256(text: str, input_encoding: str = "utf-8", key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the BLAKE2b-256 digest of text, optionally keyed (key up to 64 bytes, a MAC when set; key_encoding utf-8, hex or base64). Text is UTF-8 unless input_encoding is hex or base64; output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `key` | `str` (default `""`) | Optional MAC key, up to 64 bytes. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake2b_256("hello")`

## hash.crc32

`hash.crc32(text: str, variant: str = "ieee", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes a CRC-32 checksum of text; variant picks the parameter set (default ieee; the catalogue check value of each variant over "123456789" is verified). Text is UTF-8 unless input_encoding is hex or base64; output is fixed-width hex by default, or dec for the plain number. Non-cryptographic.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `variant` | `str` (default `"ieee"`) | Parameter set: bzip2, castagnoli, cksum, crc32c, d, ieee, jamcrc, koopman, mpeg2, posix, q. Default ieee. One of: bzip2, castagnoli, cksum, crc32c, d, ieee, jamcrc, koopman, mpeg2, posix, q. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32("hello")`

## hash.adler32

`hash.adler32(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Adler-32 digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.adler32("hello")`

## hash.md4

`hash.md4(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the MD4 digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.md4("hello")`

## hash.sha224

`hash.sha224(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-224 digest of text (224 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha224("hello")`

## hash.sha384

`hash.sha384(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-384 digest of text (384 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha384("hello")`

## hash.sha512_224

`hash.sha512_224(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-512/224 digest of text (224 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha512_224("hello")`

## hash.sha512_256

`hash.sha512_256(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA-512/256 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha512_256("hello")`

## hash.sha256d

`hash.sha256d(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Double SHA-256 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). SHA-256 applied twice, as used by Bitcoin.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha256d("hello")`

## hash.blake2b_384

`hash.blake2b_384(text: str, input_encoding: str = "utf-8", key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the BLAKE2b-384 digest of text, optionally keyed (key up to 64 bytes, a MAC when set; key_encoding utf-8, hex or base64). Text is UTF-8 unless input_encoding is hex or base64; output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `key` | `str` (default `""`) | Optional MAC key, up to 64 bytes. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake2b_384("hello")`

## hash.blake2b_512

`hash.blake2b_512(text: str, input_encoding: str = "utf-8", key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the BLAKE2b-512 digest of text, optionally keyed (key up to 64 bytes, a MAC when set; key_encoding utf-8, hex or base64). Text is UTF-8 unless input_encoding is hex or base64; output is hex by default.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `key` | `str` (default `""`) | Optional MAC key, up to 64 bytes. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake2b_512("hello")`

## hash.hash160

`hash.hash160(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the HASH160 (RIPEMD160 of SHA-256) digest of text (160 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). RIPEMD-160 of SHA-256, as used by Bitcoin addresses.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hash160("hello")`

## hash.blake2s_256

`hash.blake2s_256(text: str, input_encoding: str = "utf-8", key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the BLAKE2s-256 digest of text, optionally keyed (key up to 32 bytes). Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `key` | `str` (default `""`) | Optional MAC key, up to 32 bytes. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake2s_256("hello")`

## hash.sha3_224

`hash.sha3_224(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA3-224 digest of text (224 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha3_224("hello")`

## hash.blake3

`hash.blake3(text: str, input_encoding: str = "utf-8", length: int = 32, key: str = "", key_encoding: str = "utf-8", context: str = "", output: str = "hex") -> str`

Computes the BLAKE3 digest of text (32 bytes by default, any length up to 65536 via length). With a 32-byte key it is a keyed MAC; with context it derives a key from text as key material (key and context are mutually exclusive). Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `length` | `int` (default `32`) | Output length in bytes (1 to 65536, default 32). |
| `key` | `str` (default `""`) | Optional 32-byte key for keyed hashing. Empty means plain hashing. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `context` | `str` (default `""`) | Optional context string; when set, derives a key from text (derive_key mode). |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake3("hello")`

## hash.sha3_384

`hash.sha3_384(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA3-384 digest of text (384 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha3_384("hello")`

## hash.sha3_512

`hash.sha3_512(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SHA3-512 digest of text (512 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sha3_512("hello")`

## hash.keccak256

`hash.keccak256(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Keccak-256 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Original Keccak padding (as in Ethereum), which differs from NIST SHA3.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.keccak256("hello")`

## hash.keccak512

`hash.keccak512(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Keccak-512 digest of text (512 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Original Keccak padding (as in Ethereum), which differs from NIST SHA3.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.keccak512("hello")`

## hash.ripemd160

`hash.ripemd160(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the RIPEMD-160 digest of text (160 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.ripemd160("hello")`

## hash.sm3

`hash.sm3(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SM3 digest of text (256 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sm3("hello")`

## hash.whirlpool

`hash.whirlpool(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Whirlpool digest of text (512 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). 

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.whirlpool("hello")`

## hash.ntlm

`hash.ntlm(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the NTLM (NT hash) digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Windows NT hash: MD4 of the UTF-16LE password. Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.ntlm("hello")`

## hash.lm

`hash.lm(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the LM hash (LAN Manager) digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). Legacy LAN Manager hash: password uppercased, cut to 14 bytes, two DES halves. Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.lm("hello")`

## hash.mysql41

`hash.mysql41(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the MySQL PASSWORD() hash digest of text (160 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). MySQL PASSWORD()/mysql_native_password hash: SHA1(SHA1(password)); MySQL stores it as '*' plus uppercase hex. Legacy algorithm: fine for fingerprints and compatibility, not for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.mysql41("hello")`

## hash.crc8

`hash.crc8(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-8 (SMBus) digest of text (8 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc8("hello")`

## hash.crc8_maxim

`hash.crc8_maxim(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-8/MAXIM (Dallas 1-Wire) digest of text (8 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc8_maxim("hello")`

## hash.crc16_arc

`hash.crc16_arc(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/ARC digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_arc("hello")`

## hash.crc16_ccitt_false

`hash.crc16_ccitt_false(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/CCITT-FALSE digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_ccitt_false("hello")`

## hash.crc16_xmodem

`hash.crc16_xmodem(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/XMODEM digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_xmodem("hello")`

## hash.crc16_kermit

`hash.crc16_kermit(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/KERMIT digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_kermit("hello")`

## hash.crc16_modbus

`hash.crc16_modbus(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/MODBUS digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_modbus("hello")`

## hash.crc16_x25

`hash.crc16_x25(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/X-25 (IBM-SDLC) digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_x25("hello")`

## hash.crc16_usb

`hash.crc16_usb(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/USB digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_usb("hello")`

## hash.crc16_genibus

`hash.crc16_genibus(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/GENIBUS digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_genibus("hello")`

## hash.crc16_mcrf4xx

`hash.crc16_mcrf4xx(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/MCRF4XX digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_mcrf4xx("hello")`

## hash.crc16_dnp

`hash.crc16_dnp(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/DNP digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_dnp("hello")`

## hash.crc16_buypass

`hash.crc16_buypass(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/BUYPASS digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_buypass("hello")`

## hash.crc16_t10_dif

`hash.crc16_t10_dif(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/T10-DIF digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_t10_dif("hello")`

## hash.crc16_cdma2000

`hash.crc16_cdma2000(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/CDMA2000 digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_cdma2000("hello")`

## hash.crc16_aug_ccitt

`hash.crc16_aug_ccitt(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-16/AUG-CCITT digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16_aug_ccitt("hello")`

## hash.crc24

`hash.crc24(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-24 (OpenPGP) digest of text (24 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc24("hello")`

## hash.crc32_castagnoli

`hash.crc32_castagnoli(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32C (Castagnoli) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_castagnoli("hello")`

## hash.crc32_koopman

`hash.crc32_koopman(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32K (Koopman) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_koopman("hello")`

## hash.crc32_bzip2

`hash.crc32_bzip2(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32/BZIP2 digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_bzip2("hello")`

## hash.crc32_mpeg2

`hash.crc32_mpeg2(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32/MPEG-2 digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_mpeg2("hello")`

## hash.crc32_posix

`hash.crc32_posix(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32/POSIX (cksum) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_posix("hello")`

## hash.crc32_jamcrc

`hash.crc32_jamcrc(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32/JAMCRC digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_jamcrc("hello")`

## hash.crc32_q

`hash.crc32_q(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32Q digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_q("hello")`

## hash.crc32_d

`hash.crc32_d(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32D digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32_d("hello")`

## hash.crc64_iso

`hash.crc64_iso(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-64/ISO (GO-ISO) digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc64_iso("hello")`

## hash.crc64_ecma

`hash.crc64_ecma(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-64/ECMA (XZ) digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc64_ecma("hello")`

## hash.crc64_ecma182

`hash.crc64_ecma182(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-64/ECMA-182 digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc64_ecma182("hello")`

## hash.crc64_we

`hash.crc64_we(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-64/WE digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc64_we("hello")`

## hash.crc32c

`hash.crc32c(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the CRC-32C (Castagnoli) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc32c("hello")`

## hash.fletcher16

`hash.fletcher16(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Fletcher-16 digest of text (16 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fletcher16("hello")`

## hash.fletcher32

`hash.fletcher32(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Fletcher-32 digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fletcher32("hello")`

## hash.fnv32

`hash.fnv32(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1 32-bit digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv32("hello")`

## hash.fnv32a

`hash.fnv32a(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1a 32-bit digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv32a("hello")`

## hash.fnv64

`hash.fnv64(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1 64-bit digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv64("hello")`

## hash.fnv64a

`hash.fnv64a(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1a 64-bit digest of text (64 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv64a("hello")`

## hash.fnv128

`hash.fnv128(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1 128-bit digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv128("hello")`

## hash.fnv128a

`hash.fnv128a(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the FNV-1a 128-bit digest of text (128 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.fnv128a("hello")`

## hash.djb2

`hash.djb2(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the DJB2 (Bernstein) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.djb2("hello")`

## hash.djb2a

`hash.djb2a(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the DJB2a (xor variant) digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.djb2a("hello")`

## hash.sdbm

`hash.sdbm(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the SDBM digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.sdbm("hello")`

## hash.jenkins_oaat

`hash.jenkins_oaat(text: str, input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes the Jenkins one-at-a-time digest of text (32 bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). This is a non-cryptographic checksum: use it for error detection or hash tables, never for security.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.jenkins_oaat("hello")`

## hash.murmur3_128

`hash.murmur3_128(text: str, input_encoding: str = "utf-8", seed: int = 0, output: str = "hex") -> str`

Computes MurmurHash3 (x64, 128-bit) of text with an optional integer seed (default 0). Non-cryptographic: for hash tables and sharding, not security. Text is UTF-8 unless input_encoding is hex or base64; output is hex by default (big-endian).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `seed` | `int` (default `0`) | Seed value (default 0). Truncated to the width of the algorithm. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.murmur3_128("hello")`

## hash.murmur3_32

`hash.murmur3_32(text: str, input_encoding: str = "utf-8", seed: int = 0, output: str = "hex") -> str`

Computes MurmurHash3 (x86, 32-bit) of text with an optional integer seed (default 0). Non-cryptographic: for hash tables and sharding, not security. Text is UTF-8 unless input_encoding is hex or base64; output is hex by default (big-endian).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `seed` | `int` (default `0`) | Seed value (default 0). Truncated to the width of the algorithm. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.murmur3_32("hello")`

## hash.xxhash32

`hash.xxhash32(text: str, input_encoding: str = "utf-8", seed: int = 0, output: str = "hex") -> str`

Computes xxHash32 of text with an optional integer seed (default 0). Non-cryptographic: for hash tables and sharding, not security. Text is UTF-8 unless input_encoding is hex or base64; output is hex by default (big-endian).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `seed` | `int` (default `0`) | Seed value (default 0). Truncated to the width of the algorithm. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.xxhash32("hello")`

## hash.xxhash64

`hash.xxhash64(text: str, input_encoding: str = "utf-8", seed: int = 0, output: str = "hex") -> str`

Computes xxHash64 of text with an optional integer seed (default 0). Non-cryptographic: for hash tables and sharding, not security. Text is UTF-8 unless input_encoding is hex or base64; output is hex by default (big-endian).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `seed` | `int` (default `0`) | Seed value (default 0). Truncated to the width of the algorithm. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.xxhash64("hello")`

## hash.crc16

`hash.crc16(text: str, variant: str = "arc", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes a CRC-16 checksum of text; variant picks the parameter set (default arc; the catalogue check value of each variant over "123456789" is verified). Text is UTF-8 unless input_encoding is hex or base64; output is fixed-width hex by default, or dec for the plain number. Non-cryptographic.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `variant` | `str` (default `"arc"`) | Parameter set: arc, aug_ccitt, buypass, ccitt_false, cdma2000, dnp, genibus, kermit, mcrf4xx, modbus, t10_dif, usb, x25, xmodem. Default arc. One of: arc, aug_ccitt, buypass, ccitt_false, cdma2000, dnp, genibus, kermit, mcrf4xx, modbus, t10_dif, usb, x25, xmodem. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc16("hello")`

## hash.crc64

`hash.crc64(text: str, variant: str = "ecma", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes a CRC-64 checksum of text; variant picks the parameter set (default ecma; the catalogue check value of each variant over "123456789" is verified). Text is UTF-8 unless input_encoding is hex or base64; output is fixed-width hex by default, or dec for the plain number. Non-cryptographic.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `variant` | `str` (default `"ecma"`) | Parameter set: ecma, ecma182, iso, we, xz. Default ecma. One of: ecma, ecma182, iso, we, xz. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.crc64("hello")`

## hash.siphash

`hash.siphash(text: str, input_encoding: str = "utf-8", key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes SipHash-2-4, a keyed 64-bit hash for hash tables. The key must be 16 bytes (empty means 16 zero bytes); the 8-byte result is output in the reference little-endian byte order. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `key` | `str` (default `""`) | 16-byte key. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.siphash("hello")`

## hash.shake128

`hash.shake128(text: str, input_encoding: str = "utf-8", length: int = 32, output: str = "hex") -> str`

Computes the SHAKE128 extendable-output function (SHA-3 family) of text and returns length bytes (default 32, max 65536) encoded per output. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `length` | `int` (default `32`) | Output length in bytes (1 to 65536, default 32). |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.shake128("hello")`

## hash.shake256

`hash.shake256(text: str, input_encoding: str = "utf-8", length: int = 64, output: str = "hex") -> str`

Computes the SHAKE256 extendable-output function (SHA-3 family) of text and returns length bytes (default 64, max 65536) encoded per output. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `length` | `int` (default `64`) | Output length in bytes (1 to 65536, default 64). |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.shake256("hello")`

## hash.blake2b

`hash.blake2b(text: str, input_encoding: str = "utf-8", length: int = 64, key: str = "", key_encoding: str = "utf-8", output: str = "hex") -> str`

Computes BLAKE2b with a digest size of 1 to 64 bytes (default 64) and an optional key of up to 64 bytes. Text is UTF-8 unless input_encoding is hex or base64.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `length` | `int` (default `64`) | Output length in bytes (1 to 64, default 64). |
| `key` | `str` (default `""`) | Optional MAC key, up to 64 bytes. Empty means unkeyed. |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.blake2b("hello")`

## hash.hmac_md5

`hash.hmac_md5(key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes HMAC-MD5 of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac_md5("secret", "message")`

## hash.hmac_sha1

`hash.hmac_sha1(key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes HMAC-SHA1 of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac_sha1("secret", "message")`

## hash.hmac_sha256

`hash.hmac_sha256(key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes HMAC-SHA256 of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac_sha256("secret", "message")`

## hash.hmac_sha384

`hash.hmac_sha384(key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes HMAC-SHA384 of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac_sha384("secret", "message")`

## hash.hmac_sha512

`hash.hmac_sha512(key: str, text: str, key_encoding: str = "utf-8", input_encoding: str = "utf-8", output: str = "hex") -> str`

Computes HMAC-SHA512 of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.

| Parameter | Type | Description |
|---|---|---|
| `key` | `str` | Secret key (UTF-8 unless key_encoding is hex or base64). |
| `text` | `str` | Input text (UTF-8 bytes unless input_encoding says otherwise). |
| `key_encoding` | `str` (default `"utf-8"`) | How the key is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `input_encoding` | `str` (default `"utf-8"`) | How text is turned into bytes: utf-8 (default), hex or base64. One of: utf-8, hex, base64. |
| `output` | `str` (default `"hex"`) | Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number). One of: hex, HEX, base64, base64url, base32, dec. |

Returns `str`. Example: `hash.hmac_sha512("secret", "message")`

## hash.hash_identify

`hash.hash_identify(hash: str) -> List[str]`

Guesses which algorithms could have produced a hash string, from its prefix ($2b$ bcrypt, $argon2id$, $6$ sha512crypt, {SSHA}, ...) or its length (32 hex digits = md5/ntlm/..., 64 = sha256/...). Returns an ordered list of candidates, empty when nothing matches. Only a hint, since many algorithms share a length.

| Parameter | Type | Description |
|---|---|---|
| `hash` | `str` | Hash string, hex, base64 or a modular-crypt/PHC string. |

Returns `List[str]`. Example: `hash.hash_identify("5d41402abc4b2a76b9719d911017c592")`

## hash.hash_to_base64

`hash.hash_to_base64(hex: str) -> str`

Converts a hex digest (separators and 0x prefix tolerated) to standard base64 with padding, the form used by SRI integrity attributes and HTTP Digest headers. Odd or non-hex input is an error.

| Parameter | Type | Description |
|---|---|---|
| `hex` | `str` | Hex digest. |

Returns `str`. Example: `hash.hash_to_base64("d41d8cd98f00b204e9800998ecf8427e")`

## hash.totp

`hash.totp(secret: str, time: int = 0, period: int = 30, digits: int = 6, algo: str = "sha1") -> str`

Computes the TOTP code (RFC 6238) for a base32 secret at a Unix time (0 means now, the only use of the clock in this module). Defaults: 30 s period, 6 digits, HMAC-SHA1, as in Google Authenticator. Returns a zero-padded string.

| Parameter | Type | Description |
|---|---|---|
| `secret` | `str` | Shared secret in base32 (spaces, dashes, case and '=' padding ignored). |
| `time` | `int` (default `0`) | Unix time in seconds; 0 (default) means now. |
| `period` | `int` (default `30`) | Time step in seconds (default 30). |
| `digits` | `int` (default `6`) | Code length, 1 to 10 (default 6). |
| `algo` | `str` (default `"sha1"`) | HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512. One of: sha1, sha256, sha512. |

Returns `str`. Example: `hash.totp("JBSWY3DPEHPK3PXP")`

