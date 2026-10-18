# UUIDs and IDs (`uuid`)

UUID v1-v7, ULID, KSUID, Snowflake, ObjectID, nanoid: generate, parse, convert.

No permission is needed: these are pure functions. Generated from the tools registry.

37 functions. Use them as `uuid.<name>(...)` in any extension.

## uuid.uuid_v4

`uuid.uuid_v4(style: str = "canonical") -> str`

Generates a random version 4 UUID (RFC 9562) from crypto/rand, e.g. 3b241101-e2bb-4255-8caf-4136c566a962. Style selects the text form.

| Parameter | Type | Description |
|---|---|---|
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_v4()`

## uuid.uuid_v7

`uuid.uuid_v7(time: str = "", style: str = "canonical", seed: int = 0) -> str`

Generates a version 7 UUID (RFC 9562): 48-bit Unix millisecond time followed by random bits, so values sort by creation time and index well in databases. Pass time for a fixed instant.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.uuid_v7()`

## uuid.uuid_parse

`uuid.uuid_parse(u: str) -> Dict[str, str]`

Parses a UUID (canonical, upper case, braces, urn:uuid: or 32 hex digits) and returns version, variant, canonical and urn forms, integer value and hex; for v1/v6/v7 also the UTC timestamp, and for v1/v6 node and clock_seq. Errors if the text is not a UUID.

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | UUID in any common notation. |

Returns `Dict[str, str]`. Example: `uuid.uuid_parse("017f22e2-79b0-7cc3-98c4-dc0c0c07398f")`

## uuid.uuid_format

`uuid.uuid_format(u: str, style: str = "canonical") -> str`

Converts a UUID to another text form: canonical, upper, no_dashes, braces, urn, base64 (22 chars URL-safe), base58 (22 chars) or base32 (26 chars).

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | UUID in any common notation. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_format("3b241101-e2bb-4255-8caf-4136c566a962", "braces")`

## uuid.ulid_new

`uuid.ulid_new(time: str = "", seed: int = 0) -> str`

Generates a ULID: 26 Crockford base32 characters made of a 48-bit millisecond time and 80 random bits, sortable as text. Pass time to embed a fixed instant.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.ulid_new()`

## uuid.ulid_parse

`uuid.ulid_parse(ulid: str) -> Dict[str, str]`

Decodes a ULID (case-insensitive; I, L and O are read as 1, 1 and 0 per Crockford) into its UTC timestamp, Unix milliseconds, 80-bit entropy hex and equivalent UUID. Errors on wrong length, bad characters or overflow (first char above 7).

| Parameter | Type | Description |
|---|---|---|
| `ulid` | `str` | 26-character ULID. |

Returns `Dict[str, str]`. Example: `uuid.ulid_parse("01ARYZ6S41TSV4RRFFQ69G5FAV")`

## uuid.nanoid

`uuid.nanoid(size: int = 21, alphabet: str = "", seed: int = 0) -> str`

Generates a NanoID of size characters (1..10000, default 21) from alphabet (2..255 distinct characters, default A-Za-z0-9_-) with unbiased selection; crypto/rand unless seed is set.

| Parameter | Type | Description |
|---|---|---|
| `size` | `int` (default `21`) | Number of characters. |
| `alphabet` | `str` (default `""`) | Characters to draw from; empty means the default URL-safe alphabet. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.nanoid(12)`

## uuid.uuid_v1

`uuid.uuid_v1(time: str = "", style: str = "canonical", seed: int = 0) -> str`

Generates a version 1 UUID: 60-bit timestamp in 100 ns ticks since 1582-10-15, a random 14-bit clock sequence and a random node ID with the multicast bit set (so no real MAC address leaks).

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.uuid_v1()`

## uuid.uuid_v6

`uuid.uuid_v6(time: str = "", style: str = "canonical", seed: int = 0) -> str`

Generates a version 6 UUID (RFC 9562): the v1 timestamp reordered most-significant-first so it sorts chronologically as text or bytes.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.uuid_v6()`

## uuid.uuid_v3

`uuid.uuid_v3(namespace: str, name: str, style: str = "canonical") -> str`

Generates the deterministic version 3 UUID (md5) of name inside a namespace: the same inputs always give the same UUID. Namespace is dns, url, oid, x500 or any UUID.

| Parameter | Type | Description |
|---|---|---|
| `namespace` | `str` | dns, url, oid, x500 or a namespace UUID. |
| `name` | `str` | Name to hash, e.g. a domain or URL. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_v3("dns", "example.com")`

## uuid.uuid_v5

`uuid.uuid_v5(namespace: str, name: str, style: str = "canonical") -> str`

Generates the deterministic version 5 UUID (sha1) of name inside a namespace: the same inputs always give the same UUID. Namespace is dns, url, oid, x500 or any UUID.

| Parameter | Type | Description |
|---|---|---|
| `namespace` | `str` | dns, url, oid, x500 or a namespace UUID. |
| `name` | `str` | Name to hash, e.g. a domain or URL. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_v5("dns", "example.com")`

## uuid.uuid_validate

`uuid.uuid_validate(u: str, version: int = 0) -> bool`

Returns true when the text is a valid UUID in a common notation. With version set (1..8) it must also carry that version and the RFC variant.

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | Text to check. |
| `version` | `int` (default `0`) | Required version 1..8, or 0 for any. |

Returns `bool`. Example: `uuid.uuid_validate("not-a-uuid")`

## uuid.uuid_to_int

`uuid.uuid_to_int(u: str) -> str`

Returns the UUID as a 128-bit unsigned decimal integer string (an int would overflow).

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | UUID in any common notation. |

Returns `str`. Example: `uuid.uuid_to_int("00000000-0000-0000-0000-0000000000ff")`

## uuid.ksuid_new

`uuid.ksuid_new(time: str = "", seed: int = 0) -> str`

Generates a KSUID (Segment): 27 base62 characters made of a 32-bit timestamp in seconds since 2014-05-13 (epoch 1400000000) and 128 random bits. Errors for times before the KSUID epoch.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.ksuid_new()`

## uuid.ksuid_parse

`uuid.ksuid_parse(ksuid: str) -> Dict[str, str]`

Decodes a 27-character KSUID into its UTC timestamp, Unix seconds, KSUID-epoch seconds and 16-byte payload hex.

| Parameter | Type | Description |
|---|---|---|
| `ksuid` | `str` | 27-character base62 KSUID. |

Returns `Dict[str, str]`. Example: `uuid.ksuid_parse("0ujtsYcgvSTl8PAuAdqWYSMnLOv")`

## uuid.snowflake_decode

`uuid.snowflake_decode(id: str, epoch: str = "twitter") -> Dict[str, str]`

Decodes a 64-bit Snowflake ID (41-bit time, 5-bit worker, 5-bit process/datacenter, 12-bit sequence) into timestamp, worker, process and sequence. The epoch decides the date: twitter 1288834974657, discord 1420070400000, or custom milliseconds.

| Parameter | Type | Description |
|---|---|---|
| `id` | `str` | Decimal snowflake ID. |
| `epoch` | `str` (default `"twitter"`) | twitter (default), discord, unix, or a custom epoch in Unix milliseconds. |

Returns `Dict[str, str]`. Example: `uuid.snowflake_decode("175928847299117063", "discord")`

## uuid.objectid_new

`uuid.objectid_new(time: str = "", seed: int = 0) -> str`

Generates a MongoDB ObjectID: 24 hex characters of a 4-byte Unix time in seconds, a 5-byte random value and a 3-byte counter started at a random number.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.objectid_new()`

## uuid.objectid_parse

`uuid.objectid_parse(id: str) -> Dict[str, str]`

Decodes a MongoDB ObjectID into its UTC timestamp, the 5-byte random value (older drivers: 3-byte machine + 2-byte process) and the 3-byte counter.

| Parameter | Type | Description |
|---|---|---|
| `id` | `str` | 24-digit hex ObjectID. |

Returns `Dict[str, str]`. Example: `uuid.objectid_parse("507f1f77bcf86cd799439011")`

## uuid.short_uuid

`uuid.short_uuid(u: str = "") -> str`

Returns the 22-character base58 spelling of a UUID, or of a fresh v4 UUID when u is empty. Reverse it with short_uuid_decode.

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` (default `""`) | UUID to shorten; empty generates a new v4. |

Returns `str`. Example: `uuid.short_uuid("3b241101-e2bb-4255-8caf-4136c566a962")`

## uuid.short_uuid_decode

`uuid.short_uuid_decode(s: str, style: str = "canonical") -> str`

Expands a 22-character base58 short UUID (see short_uuid) back to a canonical UUID.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | 22-character base58 string. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.short_uuid_decode("mhvXdrZT4jP5T8vBxuvm75")`

## uuid.typeid_new

`uuid.typeid_new(prefix: str = "", time: str = "", seed: int = 0) -> str`

Generates a TypeID: a type prefix (lowercase letters and underscores, up to 63, may be empty) plus an underscore and a 26-character base32 UUID v7, e.g. user_01h455vb4pex5vsknk084sn02q. Pass time for a fixed instant.

| Parameter | Type | Description |
|---|---|---|
| `prefix` | `str` (default `""`) | Type prefix such as user or order. |
| `time` | `str` (default `""`) | Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.typeid_new("user")`

## uuid.typeid_parse

`uuid.typeid_parse(typeid: str) -> Dict[str, str]`

Decodes a TypeID into its prefix, canonical UUID and (for v7) the embedded UTC timestamp. Errors on a bad prefix, wrong suffix length or overflow.

| Parameter | Type | Description |
|---|---|---|
| `typeid` | `str` | TypeID such as user_01h455vb4pex5vsknk084sn02q. |

Returns `Dict[str, str]`. Example: `uuid.typeid_parse("user_01h455vb4pex5vsknk084sn02q")`

## uuid.id_detect

`uuid.id_detect(s: str) -> List[str]`

Guesses which identifier a string is and lists every plausible kind, best first: UUID (with version), ULID, KSUID, MongoDB ObjectID, Snowflake, TypeID, short UUID, NanoID, CUID2 and common hex digests. Returns an empty list when nothing fits.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | The identifier to classify. |

Returns `List[str]`. Example: `uuid.id_detect("01ARZ3NDEKTSV4RRFFQ69G5FAV")`

## uuid.cuid2

`uuid.cuid2(length: int = 24, seed: int = 0) -> str`

Generates a CUID2-style identifier: a random lowercase letter followed by base36 digits of a SHA-256 hash over time, counter and random entropy (length 2..32, default 24). It follows the CUID2 layout but is not byte-compatible with the reference library.

| Parameter | Type | Description |
|---|---|---|
| `length` | `int` (default `24`) | Total length 2..32. |
| `seed` | `int` (default `0`) | 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible. |

Returns `str`. Example: `uuid.cuid2()`

## uuid.namespaces

`uuid.namespaces() -> Dict[str, str]`

Returns the predefined namespace UUIDs for name-based UUIDs: dns, url, oid and x500 (RFC 4122 appendix C).

Returns `Dict[str, str]`. Example: `uuid.namespaces()`

## uuid.snowflake_encode

`uuid.snowflake_encode(time: str = "", worker: int = 0, process: int = 0, sequence: int = 0, epoch: str = "twitter") -> str`

Builds a Snowflake ID from a time, worker (0..31), process (0..31) and sequence (0..4095) relative to an epoch. Errors if the time is before the epoch or a field is out of range.

| Parameter | Type | Description |
|---|---|---|
| `time` | `str` (default `""`) | Instant to encode. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now. |
| `worker` | `int` (default `0`) | Worker ID 0..31. |
| `process` | `int` (default `0`) | Process/datacenter ID 0..31. |
| `sequence` | `int` (default `0`) | Sequence 0..4095. |
| `epoch` | `str` (default `"twitter"`) | twitter (default), discord, unix, or a custom epoch in Unix milliseconds. |

Returns `str`. Example: `uuid.snowflake_encode("2016-04-30T11:18:25.796Z", 1, 0, 7, "discord")`

## uuid.ulid_bulk

`uuid.ulid_bulk(count: int = 10) -> List[str]`

Generates count (0..10000) ULIDs that are strictly increasing: within the same millisecond the entropy is incremented as in the monotonic ULID spec.

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `10`) | How many ULIDs. |

Returns `List[str]`. Example: `uuid.ulid_bulk(5)`

## uuid.ulid_to_uuid

`uuid.ulid_to_uuid(ulid: str, style: str = "canonical") -> str`

Converts a ULID to the UUID with the same 128 bits (the version/variant bits are not rewritten, so it is not a valid v7).

| Parameter | Type | Description |
|---|---|---|
| `ulid` | `str` | 26-character ULID. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.ulid_to_uuid("01ARZ3NDEKTSV4RRFFQ69G5FAV")`

## uuid.uuid_bulk

`uuid.uuid_bulk(count: int = 10, version: str = "4", style: str = "canonical") -> List[str]`

Generates count UUIDs (0..10000) of version 1, 4, 6 or 7 (or nil/max) in the chosen style. v7 values are strictly increasing within the call.

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `10`) | How many UUIDs. |
| `version` | `str` (default `"4"`) | 1, 4, 6, 7, nil or max. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `List[str]`. Example: `uuid.uuid_bulk(5, "7")`

## uuid.uuid_compare

`uuid.uuid_compare(a: str, b: str, by: str = "bytes") -> int`

Compares two UUIDs and returns -1, 0 or 1. by=bytes compares raw bytes (correct creation order for v6/v7); by=time compares the embedded timestamps (also correct for v1).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First UUID. |
| `b` | `str` | Second UUID. |
| `by` | `str` (default `"bytes"`) | bytes (plain byte order) or time (embedded v1/v6/v7 timestamp; UUIDs without one sort last). |

Returns `int`. Example: `uuid.uuid_compare("017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "017f22e2-79b1-7000-8000-000000000000")`

## uuid.uuid_from_int

`uuid.uuid_from_int(value: str, style: str = "canonical") -> str`

Converts a decimal (or 0x-prefixed hex) integer in 0..2^128-1 to a canonical UUID.

| Parameter | Type | Description |
|---|---|---|
| `value` | `str` | Decimal or 0x hex integer. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_from_int("255")`

## uuid.uuid_max

`uuid.uuid_max() -> str`

Returns the max UUID ffffffff-ffff-ffff-ffff-ffffffffffff (all bits one, RFC 9562).

Returns `str`. Example: `uuid.uuid_max()`

## uuid.uuid_nil

`uuid.uuid_nil() -> str`

Returns the nil UUID 00000000-0000-0000-0000-000000000000 (all bits zero).

Returns `str`. Example: `uuid.uuid_nil()`

## uuid.uuid_sort

`uuid.uuid_sort(uuids: List[str], by: str = "bytes", reverse: bool = False) -> List[str]`

Sorts UUIDs (canonicalized) by raw bytes or by embedded timestamp; v7 UUIDs come out in creation order either way. Errors on the first invalid UUID. Set reverse for newest first.

| Parameter | Type | Description |
|---|---|---|
| `uuids` | `List[str]` | UUIDs in any common notation. |
| `by` | `str` (default `"bytes"`) | bytes (plain byte order) or time (embedded v1/v6/v7 timestamp; UUIDs without one sort last). |
| `reverse` | `bool` (default `False`) | Descending order. |

Returns `List[str]`. Example: `uuid.uuid_sort(["017f22e2-79b1-7000-8000-000000000000", "017f22e2-79b0-7cc3-98c4-dc0c0c07398f"])`

## uuid.uuid_to_typeid

`uuid.uuid_to_typeid(u: str, prefix: str = "") -> str`

Encodes a UUID as a TypeID with the given prefix (may be empty).

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | UUID in any common notation. |
| `prefix` | `str` (default `""`) | Type prefix such as user. |

Returns `str`. Example: `uuid.uuid_to_typeid("01890a5d-ac96-774b-bcce-b302099a8057", "user")`

## uuid.uuid_to_ulid

`uuid.uuid_to_ulid(u: str) -> str`

Converts a UUID to the 26-character ULID spelling of the same 128 bits; for a v7 UUID the ULID time prefix equals the UUID timestamp.

| Parameter | Type | Description |
|---|---|---|
| `u` | `str` | UUID in any common notation. |

Returns `str`. Example: `uuid.uuid_to_ulid("01563df3-6481-d676-4c61-efb99302bd5b")`

## uuid.uuid_v8

`uuid.uuid_v8(data: str, style: str = "canonical") -> str`

Builds a version 8 UUID (RFC 9562, vendor-specific) from up to 16 bytes given as hex (shorter input is zero padded); version and variant bits overwrite the corresponding bits.

| Parameter | Type | Description |
|---|---|---|
| `data` | `str` | Up to 32 hex digits of custom data. |
| `style` | `str` (default `"canonical"`) | Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32. |

Returns `str`. Example: `uuid.uuid_v8("00112233445566778899aabbccddeeff")`

