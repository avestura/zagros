"""UUIDs and IDs: UUID v1-v7, ULID, KSUID, Snowflake, ObjectID, nanoid: generate, parse, convert.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `uuid` is predeclared in main.star; call `uuid.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: uuid, guid, unique id, ulid, ksuid, snowflake, nanoid, object id, typeid, identifier, generate id, parse id, v4, v7
"""

from typing import Any, Dict, List

def uuid_v4(style: str = "canonical") -> str:
    """Generates a random version 4 UUID (RFC 9562) from crypto/rand, e.g. 3b241101-e2bb-4255-8caf-4136c566a962. Style selects the text form.

        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_v4()
    Task: generate a UUID
    Keywords: guid, random uuid, new id, unique identifier, unique id, generate, rfc 4122, identifier
    Needs: none (pure function).
    """

def uuid_v7(time: str = "", style: str = "canonical", seed: int = 0) -> str:
    """Generates a version 7 UUID (RFC 9562): 48-bit Unix millisecond time followed by random bits, so values sort by creation time and index well in databases. Pass time for a fixed instant.

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.uuid_v7()
    Task: generate a sortable UUID
    Keywords: time ordered uuid, uuid v7, database key, sortable uuid, database primary key, monotonic, rfc 9562
    Needs: none (pure function).
    """

def uuid_parse(u: str) -> Dict[str, str]:
    """Parses a UUID (canonical, upper case, braces, urn:uuid: or 32 hex digits) and returns version, variant, canonical and urn forms, integer value and hex; for v1/v6/v7 also the UTC timestamp, and for v1/v6 node and clock_seq. Errors if the text is not a UUID.

        u: UUID in any common notation.

    Example: uuid.uuid_parse("017f22e2-79b0-7cc3-98c4-dc0c0c07398f")
    Task: inspect a UUID
    Keywords: uuid version, decode uuid, what is this uuid, uuid timestamp, analyze, guid info, extract time
    Needs: none (pure function).
    """

def uuid_format(u: str, style: str = "canonical") -> str:
    """Converts a UUID to another text form: canonical, upper, no_dashes, braces, urn, base64 (22 chars URL-safe), base58 (22 chars) or base32 (26 chars).

        u: UUID in any common notation.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_format("3b241101-e2bb-4255-8caf-4136c566a962", "braces")
    Keywords: uppercase uuid, remove dashes, guid braces, urn, base64 uuid, base58 uuid, short uuid
    Needs: none (pure function).
    """

def ulid_new(time: str = "", seed: int = 0) -> str:
    """Generates a ULID: 26 Crockford base32 characters made of a 48-bit millisecond time and 80 random bits, sortable as text. Pass time to embed a fixed instant.

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.ulid_new()
    Task: generate a ULID
    Keywords: sortable id, lexicographic id, crockford base32, lexicographic, time ordered id, unique id
    Needs: none (pure function).
    """

def ulid_parse(ulid: str) -> Dict[str, str]:
    """Decodes a ULID (case-insensitive; I, L and O are read as 1, 1 and 0 per Crockford) into its UTC timestamp, Unix milliseconds, 80-bit entropy hex and equivalent UUID. Errors on wrong length, bad characters or overflow (first char above 7).

        ulid: 26-character ULID.

    Example: uuid.ulid_parse("01ARYZ6S41TSV4RRFFQ69G5FAV")
    Keywords: ulid timestamp, decode ulid, extract time, entropy
    Needs: none (pure function).
    """

def nanoid(size: int = 21, alphabet: str = "", seed: int = 0) -> str:
    """Generates a NanoID of size characters (1..10000, default 21) from alphabet (2..255 distinct characters, default A-Za-z0-9_-) with unbiased selection; crypto/rand unless seed is set.

        size: Number of characters.
        alphabet: Characters to draw from; empty means the default URL-safe alphabet.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.nanoid(12)
    Task: generate a short random id
    Keywords: short id, url safe id, compact identifier, compact unique id, random id
    Needs: none (pure function).
    """

def uuid_v1(time: str = "", style: str = "canonical", seed: int = 0) -> str:
    """Generates a version 1 UUID: 60-bit timestamp in 100 ns ticks since 1582-10-15, a random 14-bit clock sequence and a random node ID with the multicast bit set (so no real MAC address leaks).

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.uuid_v1()
    Keywords: time based uuid, timestamp uuid, mac address, rfc 4122, gregorian
    Needs: none (pure function).
    """

def uuid_v6(time: str = "", style: str = "canonical", seed: int = 0) -> str:
    """Generates a version 6 UUID (RFC 9562): the v1 timestamp reordered most-significant-first so it sorts chronologically as text or bytes.

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.uuid_v6()
    Keywords: sortable uuid, reordered v1, time ordered, rfc 9562
    Needs: none (pure function).
    """

def uuid_v3(namespace: str, name: str, style: str = "canonical") -> str:
    """Generates the deterministic version 3 UUID (md5) of name inside a namespace: the same inputs always give the same UUID. Namespace is dns, url, oid, x500 or any UUID.

        namespace: dns, url, oid, x500 or a namespace UUID.
        name: Name to hash, e.g. a domain or URL.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_v3("dns", "example.com")
    Keywords: deterministic uuid, namespace uuid, hash uuid, md5, stable id, rfc 4122
    Needs: none (pure function).
    """

def uuid_v5(namespace: str, name: str, style: str = "canonical") -> str:
    """Generates the deterministic version 5 UUID (sha1) of name inside a namespace: the same inputs always give the same UUID. Namespace is dns, url, oid, x500 or any UUID.

        namespace: dns, url, oid, x500 or a namespace UUID.
        name: Name to hash, e.g. a domain or URL.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_v5("dns", "example.com")
    Keywords: name based uuid, deterministic uuid from a string, namespace uuid, deterministic uuid, hash uuid, sha1, stable id, rfc 4122
    Needs: none (pure function).
    """

def uuid_validate(u: str, version: int = 0) -> bool:
    """Returns true when the text is a valid UUID in a common notation. With version set (1..8) it must also carry that version and the RFC variant.

        u: Text to check.
        version: Required version 1..8, or 0 for any.

    Example: uuid.uuid_validate("not-a-uuid")
    Task: check if text is a UUID
    Keywords: is valid guid, validate guid, check uuid, is uuid, guid validation, verify
    Needs: none (pure function).
    """

def uuid_to_int(u: str) -> str:
    """Returns the UUID as a 128-bit unsigned decimal integer string (an int would overflow).

        u: UUID in any common notation.

    Example: uuid.uuid_to_int("00000000-0000-0000-0000-0000000000ff")
    Keywords: uuid number, decimal, 128 bit, bigint
    Needs: none (pure function).
    """

def ksuid_new(time: str = "", seed: int = 0) -> str:
    """Generates a KSUID (Segment): 27 base62 characters made of a 32-bit timestamp in seconds since 2014-05-13 (epoch 1400000000) and 128 random bits. Errors for times before the KSUID epoch.

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.ksuid_new()
    Keywords: k-sortable id, segment id, segment, k-sortable, base62, unique id
    Needs: none (pure function).
    """

def ksuid_parse(ksuid: str) -> Dict[str, str]:
    """Decodes a 27-character KSUID into its UTC timestamp, Unix seconds, KSUID-epoch seconds and 16-byte payload hex.

        ksuid: 27-character base62 KSUID.

    Example: uuid.ksuid_parse("0ujtsYcgvSTl8PAuAdqWYSMnLOv")
    Keywords: decode ksuid, ksuid timestamp, segment id
    Needs: none (pure function).
    """

def snowflake_decode(id: str, epoch: str = "twitter") -> Dict[str, str]:
    """Decodes a 64-bit Snowflake ID (41-bit time, 5-bit worker, 5-bit process/datacenter, 12-bit sequence) into timestamp, worker, process and sequence. The epoch decides the date: twitter 1288834974657, discord 1420070400000, or custom milliseconds.

        id: Decimal snowflake ID.
        epoch: twitter (default), discord, unix, or a custom epoch in Unix milliseconds.

    Example: uuid.snowflake_decode("175928847299117063", "discord")
    Task: read a Snowflake ID
    Keywords: twitter id, discord id, snowflake timestamp, tweet id, timestamp from id, worker, sequence
    Needs: none (pure function).
    """

def objectid_new(time: str = "", seed: int = 0) -> str:
    """Generates a MongoDB ObjectID: 24 hex characters of a 4-byte Unix time in seconds, a 5-byte random value and a 3-byte counter started at a random number.

        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.objectid_new()
    Keywords: mongodb id, bson objectid, bson, document id, _id
    Needs: none (pure function).
    """

def objectid_parse(id: str) -> Dict[str, str]:
    """Decodes a MongoDB ObjectID into its UTC timestamp, the 5-byte random value (older drivers: 3-byte machine + 2-byte process) and the 3-byte counter.

        id: 24-digit hex ObjectID.

    Example: uuid.objectid_parse("507f1f77bcf86cd799439011")
    Keywords: decode objectid, mongodb timestamp, machine, counter, created at
    Needs: none (pure function).
    """

def short_uuid(u: str = "") -> str:
    """Returns the 22-character base58 spelling of a UUID, or of a fresh v4 UUID when u is empty. Reverse it with short_uuid_decode.

        u: UUID to shorten; empty generates a new v4.

    Example: uuid.short_uuid("3b241101-e2bb-4255-8caf-4136c566a962")
    Keywords: shortid, compact uuid, url friendly, base58 uuid
    Needs: none (pure function).
    """

def short_uuid_decode(s: str, style: str = "canonical") -> str:
    """Expands a 22-character base58 short UUID (see short_uuid) back to a canonical UUID.

        s: 22-character base58 string.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.short_uuid_decode("mhvXdrZT4jP5T8vBxuvm75")
    Keywords: expand short uuid, base58 decode, shortid
    Needs: none (pure function).
    """

def typeid_new(prefix: str = "", time: str = "", seed: int = 0) -> str:
    """Generates a TypeID: a type prefix (lowercase letters and underscores, up to 63, may be empty) plus an underscore and a 26-character base32 UUID v7, e.g. user_01h455vb4pex5vsknk084sn02q. Pass time for a fixed instant.

        prefix: Type prefix such as user or order.
        time: Instant to embed. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.typeid_new("user")
    Keywords: prefixed id, stripe style id, typed id, type safe id
    Needs: none (pure function).
    """

def typeid_parse(typeid: str) -> Dict[str, str]:
    """Decodes a TypeID into its prefix, canonical UUID and (for v7) the embedded UTC timestamp. Errors on a bad prefix, wrong suffix length or overflow.

        typeid: TypeID such as user_01h455vb4pex5vsknk084sn02q.

    Example: uuid.typeid_parse("user_01h455vb4pex5vsknk084sn02q")
    Keywords: decode typeid, prefix, uuid from typeid
    Needs: none (pure function).
    """

def id_detect(s: str) -> List[str]:
    """Guesses which identifier a string is and lists every plausible kind, best first: UUID (with version), ULID, KSUID, MongoDB ObjectID, Snowflake, TypeID, short UUID, NanoID, CUID2 and common hex digests. Returns an empty list when nothing fits.

        s: The identifier to classify.

    Example: uuid.id_detect("01ARZ3NDEKTSV4RRFFQ69G5FAV")
    Task: identify what kind of ID this is
    Keywords: which id format, is this uuid or ulid, identify id, guess id type, what kind of id, uuid or ulid, classify
    Needs: none (pure function).
    """

def cuid2(length: int = 24, seed: int = 0) -> str:
    """Generates a CUID2-style identifier: a random lowercase letter followed by base36 digits of a SHA-256 hash over time, counter and random entropy (length 2..32, default 24). It follows the CUID2 layout but is not byte-compatible with the reference library.

        length: Total length 2..32.
        seed: 0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.

    Example: uuid.cuid2()
    Keywords: collision resistant id, cuid, secure id, horizontal scaling
    Needs: none (pure function).
    """

def namespaces() -> Dict[str, str]:
    """Returns the predefined namespace UUIDs for name-based UUIDs: dns, url, oid and x500 (RFC 4122 appendix C).

    Example: uuid.namespaces()
    Keywords: namespace constants, dns, url, oid, x500, uuid v5
    Needs: none (pure function).
    """

def snowflake_encode(time: str = "", worker: int = 0, process: int = 0, sequence: int = 0, epoch: str = "twitter") -> str:
    """Builds a Snowflake ID from a time, worker (0..31), process (0..31) and sequence (0..4095) relative to an epoch. Errors if the time is before the epoch or a field is out of range.

        time: Instant to encode. RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.
        worker: Worker ID 0..31.
        process: Process/datacenter ID 0..31.
        sequence: Sequence 0..4095.
        epoch: twitter (default), discord, unix, or a custom epoch in Unix milliseconds.

    Example: uuid.snowflake_encode("2016-04-30T11:18:25.796Z", 1, 0, 7, "discord")
    Keywords: generate snowflake, twitter id, discord id, fabricate id
    Needs: none (pure function).
    """

def ulid_bulk(count: int = 10) -> List[str]:
    """Generates count (0..10000) ULIDs that are strictly increasing: within the same millisecond the entropy is incremented as in the monotonic ULID spec.

        count: How many ULIDs.

    Example: uuid.ulid_bulk(5)
    Keywords: monotonic ulid, batch, many ids
    Needs: none (pure function).
    """

def ulid_to_uuid(ulid: str, style: str = "canonical") -> str:
    """Converts a ULID to the UUID with the same 128 bits (the version/variant bits are not rewritten, so it is not a valid v7).

        ulid: 26-character ULID.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.ulid_to_uuid("01ARZ3NDEKTSV4RRFFQ69G5FAV")
    Keywords: convert ulid, ulid to guid
    Needs: none (pure function).
    """

def uuid_bulk(count: int = 10, version: str = "4", style: str = "canonical") -> List[str]:
    """Generates count UUIDs (0..10000) of version 1, 4, 6 or 7 (or nil/max) in the chosen style. v7 values are strictly increasing within the call.

        count: How many UUIDs.
        version: 1, 4, 6, 7, nil or max.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_bulk(5, "7")
    Task: generate many UUIDs
    Keywords: batch of guids, list of uuids, batch, many uuids, generate multiple, bulk
    Needs: none (pure function).
    """

def uuid_compare(a: str, b: str, by: str = "bytes") -> int:
    """Compares two UUIDs and returns -1, 0 or 1. by=bytes compares raw bytes (correct creation order for v6/v7); by=time compares the embedded timestamps (also correct for v1).

        a: First UUID.
        b: Second UUID.
        by: bytes (plain byte order) or time (embedded v1/v6/v7 timestamp; UUIDs without one sort last).

    Example: uuid.uuid_compare("017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "017f22e2-79b1-7000-8000-000000000000")
    Keywords: order, sort order, earlier, later, which is newer
    Needs: none (pure function).
    """

def uuid_from_int(value: str, style: str = "canonical") -> str:
    """Converts a decimal (or 0x-prefixed hex) integer in 0..2^128-1 to a canonical UUID.

        value: Decimal or 0x hex integer.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_from_int("255")
    Keywords: number to uuid, bigint, 128 bit
    Needs: none (pure function).
    """

def uuid_max() -> str:
    """Returns the max UUID ffffffff-ffff-ffff-ffff-ffffffffffff (all bits one, RFC 9562).

    Example: uuid.uuid_max()
    Keywords: maximum uuid, all ones, sentinel
    Needs: none (pure function).
    """

def uuid_nil() -> str:
    """Returns the nil UUID 00000000-0000-0000-0000-000000000000 (all bits zero).

    Example: uuid.uuid_nil()
    Keywords: empty uuid, zero uuid, null guid
    Needs: none (pure function).
    """

def uuid_sort(uuids: List[str], by: str = "bytes", reverse: bool = False) -> List[str]:
    """Sorts UUIDs (canonicalized) by raw bytes or by embedded timestamp; v7 UUIDs come out in creation order either way. Errors on the first invalid UUID. Set reverse for newest first.

        uuids: UUIDs in any common notation.
        by: bytes (plain byte order) or time (embedded v1/v6/v7 timestamp; UUIDs without one sort last).
        reverse: Descending order.

    Example: uuid.uuid_sort(["017f22e2-79b1-7000-8000-000000000000", "017f22e2-79b0-7cc3-98c4-dc0c0c07398f"])
    Keywords: order uuids, chronological, sort by time
    Needs: none (pure function).
    """

def uuid_to_typeid(u: str, prefix: str = "") -> str:
    """Encodes a UUID as a TypeID with the given prefix (may be empty).

        u: UUID in any common notation.
        prefix: Type prefix such as user.

    Example: uuid.uuid_to_typeid("01890a5d-ac96-774b-bcce-b302099a8057", "user")
    Keywords: convert uuid, typeid encode, prefixed id
    Needs: none (pure function).
    """

def uuid_to_ulid(u: str) -> str:
    """Converts a UUID to the 26-character ULID spelling of the same 128 bits; for a v7 UUID the ULID time prefix equals the UUID timestamp.

        u: UUID in any common notation.

    Example: uuid.uuid_to_ulid("01563df3-6481-d676-4c61-efb99302bd5b")
    Keywords: convert uuid, guid to ulid, crockford
    Needs: none (pure function).
    """

def uuid_v8(data: str, style: str = "canonical") -> str:
    """Builds a version 8 UUID (RFC 9562, vendor-specific) from up to 16 bytes given as hex (shorter input is zero padded); version and variant bits overwrite the corresponding bits.

        data: Up to 32 hex digits of custom data.
        style: Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32. One of: canonical, upper, no_dashes, braces, urn, base64, base58, base32.

    Example: uuid.uuid_v8("00112233445566778899aabbccddeeff")
    Keywords: custom uuid, vendor, rfc 9562, experimental
    Needs: none (pure function).
    """
