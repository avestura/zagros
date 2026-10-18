package uuid

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

// ---- ULID ----

type ulidGen struct {
	started bool
	ms      int64
	hi      uint16
	lo      uint64
}

func (g *ulidGen) next(r *rand.Rand, ms int64) [16]byte {
	if g.started && ms <= g.ms {
		ms = g.ms
		g.lo++
		if g.lo == 0 {
			g.hi++
			if g.hi == 0 {
				ms++
				g.hi, g.lo = uint16(r.Uint32()), r.Uint64()
			}
		}
	} else {
		g.hi, g.lo = uint16(r.Uint32()), r.Uint64()
	}
	g.started, g.ms = true, ms
	return packULID(ms, g.hi, g.lo)
}

func packULID(ms int64, hi uint16, lo uint64) [16]byte {
	var b [16]byte
	b[0], b[1], b[2], b[3], b[4], b[5] = byte(ms>>40), byte(ms>>32), byte(ms>>24), byte(ms>>16), byte(ms>>8), byte(ms)
	binary.BigEndian.PutUint16(b[6:8], hi)
	binary.BigEndian.PutUint64(b[8:16], lo)
	return b
}

func encodeULID(b [16]byte) string { return encodeBase(new(big.Int).SetBytes(b[:]), crockford, 26) }

func decodeULID(s string) ([16]byte, error) {
	var out [16]byte
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) != 26 {
		return out, fmt.Errorf("a ULID has 26 characters, got %d", len(s))
	}
	s = strings.NewReplacer("I", "1", "L", "1", "O", "0").Replace(s)
	if s[0] > '7' {
		return out, fmt.Errorf("ULID overflows 128 bits (first character must be 0-7)")
	}
	n, err := decodeBase(s, crockford)
	if err != nil {
		return out, fmt.Errorf("not a ULID: %w", err)
	}
	copy(out[:], n.FillBytes(make([]byte, 16)))
	return out, nil
}

func ulidTime(b [16]byte) time.Time {
	ms := int64(b[0])<<40 | int64(b[1])<<32 | int64(b[2])<<24 | int64(b[3])<<16 | int64(b[4])<<8 | int64(b[5])
	return time.UnixMilli(ms).UTC()
}

func acceptsULID(s string) bool { _, err := decodeULID(s); return err == nil }

// ---- KSUID ----

const ksuidEpoch = 1400000000

func encodeKSUID(t time.Time, payload []byte) (string, error) {
	sec := t.Unix() - ksuidEpoch
	if sec < 0 || sec > 0xffffffff {
		return "", fmt.Errorf("time is outside the KSUID range (2014-05-13 to 2150)")
	}
	var b [20]byte
	binary.BigEndian.PutUint32(b[:4], uint32(sec))
	copy(b[4:], payload)
	return encodeBase(new(big.Int).SetBytes(b[:]), b62Alphabet, 27), nil
}

func decodeKSUID(s string) ([20]byte, error) {
	var out [20]byte
	s = strings.TrimSpace(s)
	if len(s) != 27 {
		return out, fmt.Errorf("a KSUID has 27 characters, got %d", len(s))
	}
	n, err := decodeBase(s, b62Alphabet)
	if err != nil {
		return out, fmt.Errorf("not a KSUID: %w", err)
	}
	b, err := bigToBytes(n, 20)
	if err != nil {
		return out, fmt.Errorf("not a KSUID: value exceeds 160 bits")
	}
	copy(out[:], b)
	return out, nil
}

func ksuidTime(b [20]byte) time.Time {
	return time.Unix(int64(binary.BigEndian.Uint32(b[:4]))+ksuidEpoch, 0).UTC()
}

func acceptsKSUID(s string) bool { _, err := decodeKSUID(s); return err == nil }

// ---- Snowflake ----

var snowflakeEpochs = map[string]int64{
	"twitter": 1288834974657, "x": 1288834974657, "discord": 1420070400000,
	"unix": 0, "instagram": 1314220021721,
}

func snowflakeEpoch(s string) (int64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		s = "twitter"
	}
	if e, ok := snowflakeEpochs[s]; ok {
		return e, nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 {
		return n, nil
	}
	return 0, fmt.Errorf("epoch must be twitter, discord, unix or a millisecond Unix time, got %q", s)
}

// ---- ObjectID ----

func decodeObjectID(s string) ([12]byte, error) {
	var out [12]byte
	s = strings.TrimSpace(s)
	if len(s) != 24 {
		return out, fmt.Errorf("an ObjectID has 24 hex digits, got %d characters", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return out, fmt.Errorf("not an ObjectID: invalid hex")
	}
	copy(out[:], b)
	return out, nil
}

func acceptsObjectID(s string) bool { _, err := decodeObjectID(s); return err == nil }

// ---- TypeID ----

const typeidAlphabet = "0123456789abcdefghjkmnpqrstvwxyz"

var typeidPrefixRe = regexp.MustCompile(`^([a-z]([a-z_]{0,61}[a-z])?)?$`)

func typeidEncode(prefix string, u UUID) string {
	s := encodeBase(u.bigInt(), typeidAlphabet, 26)
	if prefix == "" {
		return s
	}
	return prefix + "_" + s
}

func typeidDecode(s string) (string, UUID, error) {
	var u UUID
	prefix, suffix := "", s
	if i := strings.LastIndex(s, "_"); i >= 0 {
		prefix, suffix = s[:i], s[i+1:]
		if prefix == "" {
			return "", u, fmt.Errorf("empty prefix before underscore")
		}
	}
	if !typeidPrefixRe.MatchString(prefix) || len(prefix) > 63 {
		return "", u, fmt.Errorf("invalid TypeID prefix %q (lowercase letters and underscores, max 63)", prefix)
	}
	if len(suffix) != 26 || suffix[0] > '7' {
		return "", u, fmt.Errorf("TypeID suffix must be 26 base32 characters starting with 0-7")
	}
	n, err := decodeBase(suffix, typeidAlphabet)
	if err != nil {
		return "", u, fmt.Errorf("invalid TypeID suffix: %w", err)
	}
	copy(u[:], n.FillBytes(make([]byte, 16)))
	return prefix, u, nil
}

// ---- nanoid / cuid2 ----

const nanoidDefault = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(r *rand.Rand, alphabet []rune, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(alphabet[r.IntN(len(alphabet))])
	}
	return b.String()
}

func init() {
	// ---- ULID ----
	reg(tools.Func{
		Name: "ulid_new", Title: "ULID", Order: 5, Input: "time", Keywords: []string{"ulid"},
		Doc:     "Generates a ULID: 26 Crockford base32 characters made of a 48-bit millisecond time and 80 random bits, sortable as text. Pass time to embed a fixed instant.",
		Params:  []tools.Param{timeParam("Instant to embed."), seedParam()},
		Returns: "str", Example: `uuid.ulid_new()`,
		Tags: []string{"sortable id", "crockford base32", "lexicographic", "time ordered id", "unique id"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			if t.UnixMilli() < 0 || t.UnixMilli() >= 1<<48 {
				return nil, fmt.Errorf("time is outside the ULID range (1970 to 10889)")
			}
			var g ulidGen
			return encodeULID(g.next(newRng(a.Int("seed")), t.UnixMilli())), nil
		},
	})
	reg(tools.Func{
		Name: "ulid_bulk", Title: "Many ULIDs",
		Doc:     "Generates count (0..10000) ULIDs that are strictly increasing: within the same millisecond the entropy is incremented as in the monotonic ULID spec.",
		Params:  []tools.Param{p("count", "int", "How many ULIDs.", int64(10))},
		Returns: "list[str]", Example: `uuid.ulid_bulk(5)`,
		Tags: []string{"monotonic ulid", "batch", "many ids"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("count")
			if n < 0 || n > maxCount {
				return nil, fmt.Errorf("count must be 0..%d", maxCount)
			}
			var g ulidGen
			r := newRng(0)
			out := make([]string, n)
			for i := range out {
				out[i] = encodeULID(g.next(r, time.Now().UnixMilli()))
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "ulid_parse", Title: "Inspect ULID", Order: 6, Input: "ulid", Accepts: acceptsULID, Keywords: []string{"parse"},
		Doc:     "Decodes a ULID (case-insensitive; I, L and O are read as 1, 1 and 0 per Crockford) into its UTC timestamp, Unix milliseconds, 80-bit entropy hex and equivalent UUID. Errors on wrong length, bad characters or overflow (first char above 7).",
		Params:  []tools.Param{req("ulid", "str", "26-character ULID.")},
		Returns: "dict[str, str]", Example: `uuid.ulid_parse("01ARYZ6S41TSV4RRFFQ69G5FAV")`,
		Tags: []string{"decode ulid", "ulid timestamp", "extract time", "entropy"},
		Fn: func(a tools.Args) (any, error) {
			b, err := decodeULID(a.Str("ulid"))
			if err != nil {
				return nil, err
			}
			t := ulidTime(b)
			return map[string]string{
				"canonical": encodeULID(b), "timestamp": rfc(t), "unix_ms": strconv.FormatInt(t.UnixMilli(), 10),
				"entropy": hex.EncodeToString(b[6:]), "uuid": UUID(b).String(),
			}, nil
		},
	})
	reg(tools.Func{
		Name: "ulid_to_uuid", Title: "ULID to UUID", Input: "ulid", Accepts: acceptsULID,
		Doc:     "Converts a ULID to the UUID with the same 128 bits (the version/variant bits are not rewritten, so it is not a valid v7).",
		Params:  []tools.Param{req("ulid", "str", "26-character ULID."), styleParam()},
		Returns: "str", Example: `uuid.ulid_to_uuid("01ARZ3NDEKTSV4RRFFQ69G5FAV")`,
		Tags: []string{"convert ulid", "ulid to guid"},
		Fn: func(a tools.Args) (any, error) {
			b, err := decodeULID(a.Str("ulid"))
			if err != nil {
				return nil, err
			}
			return fmtOut(UUID(b), a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_to_ulid", Title: "UUID to ULID", Input: "u", Accepts: acceptsUUID,
		Doc:     "Converts a UUID to the 26-character ULID spelling of the same 128 bits; for a v7 UUID the ULID time prefix equals the UUID timestamp.",
		Params:  []tools.Param{req("u", "str", "UUID in any common notation.")},
		Returns: "str", Example: `uuid.uuid_to_ulid("01563df3-6481-d676-4c61-efb99302bd5b")`,
		Tags: []string{"convert uuid", "guid to ulid", "crockford"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return nil, err
			}
			return encodeULID(u), nil
		},
	})

	// ---- KSUID ----
	reg(tools.Func{
		Name: "ksuid_new", Title: "KSUID", Order: 30, Input: "time", Keywords: []string{"ksuid"},
		Doc:     "Generates a KSUID (Segment): 27 base62 characters made of a 32-bit timestamp in seconds since 2014-05-13 (epoch 1400000000) and 128 random bits. Errors for times before the KSUID epoch.",
		Params:  []tools.Param{timeParam("Instant to embed."), seedParam()},
		Returns: "str", Example: `uuid.ksuid_new()`,
		Tags: []string{"segment", "k-sortable", "base62", "unique id"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			return encodeKSUID(t, randBytes(newRng(a.Int("seed")), 16))
		},
	})
	reg(tools.Func{
		Name: "ksuid_parse", Title: "Inspect KSUID", Order: 31, Input: "ksuid", Accepts: acceptsKSUID,
		Doc:     "Decodes a 27-character KSUID into its UTC timestamp, Unix seconds, KSUID-epoch seconds and 16-byte payload hex.",
		Params:  []tools.Param{req("ksuid", "str", "27-character base62 KSUID.")},
		Returns: "dict[str, str]", Example: `uuid.ksuid_parse("0ujtsYcgvSTl8PAuAdqWYSMnLOv")`,
		Tags: []string{"decode ksuid", "ksuid timestamp", "segment id"},
		Fn: func(a tools.Args) (any, error) {
			b, err := decodeKSUID(a.Str("ksuid"))
			if err != nil {
				return nil, err
			}
			t := ksuidTime(b)
			return map[string]string{
				"timestamp": rfc(t), "unix": strconv.FormatInt(t.Unix(), 10),
				"ksuid_seconds": strconv.FormatInt(t.Unix()-ksuidEpoch, 10), "payload": strings.ToUpper(hex.EncodeToString(b[4:])),
			}, nil
		},
	})

	// ---- Snowflake ----
	epochParam := p("epoch", "str", "twitter (default), discord, unix, or a custom epoch in Unix milliseconds.", "twitter")
	reg(tools.Func{
		Name: "snowflake_decode", Title: "Decode snowflake ID", Order: 32, Input: "id", Keywords: []string{"snowflake", "discord", "twitter"},
		Accepts: func(s string) bool { _, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64); return err == nil },
		Doc:     "Decodes a 64-bit Snowflake ID (41-bit time, 5-bit worker, 5-bit process/datacenter, 12-bit sequence) into timestamp, worker, process and sequence. The epoch decides the date: twitter 1288834974657, discord 1420070400000, or custom milliseconds.",
		Params:  []tools.Param{req("id", "str", "Decimal snowflake ID."), epochParam},
		Returns: "dict[str, str]", Example: `uuid.snowflake_decode("175928847299117063", "discord")`,
		Tags: []string{"twitter id", "discord id", "tweet id", "timestamp from id", "worker", "sequence"},
		Fn: func(a tools.Args) (any, error) {
			id, err := strconv.ParseUint(strings.TrimSpace(a.Str("id")), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%q is not a decimal snowflake ID", a.Str("id"))
			}
			ep, err := snowflakeEpoch(a.Str("epoch"))
			if err != nil {
				return nil, err
			}
			ms := int64(id>>22) + ep
			t := time.UnixMilli(ms).UTC()
			return map[string]string{
				"timestamp": rfc(t), "unix_ms": strconv.FormatInt(ms, 10),
				"worker": strconv.FormatUint(id>>17&0x1f, 10), "process": strconv.FormatUint(id>>12&0x1f, 10),
				"sequence": strconv.FormatUint(id&0xfff, 10),
			}, nil
		},
	})
	reg(tools.Func{
		Name: "snowflake_encode", Title: "Build snowflake ID", Input: "time",
		Doc:     "Builds a Snowflake ID from a time, worker (0..31), process (0..31) and sequence (0..4095) relative to an epoch. Errors if the time is before the epoch or a field is out of range.",
		Params:  []tools.Param{timeParam("Instant to encode."), p("worker", "int", "Worker ID 0..31.", int64(0)), p("process", "int", "Process/datacenter ID 0..31.", int64(0)), p("sequence", "int", "Sequence 0..4095.", int64(0)), epochParam},
		Returns: "str", Example: `uuid.snowflake_encode("2016-04-30T11:18:25.796Z", 1, 0, 7, "discord")`,
		Tags: []string{"generate snowflake", "twitter id", "discord id", "fabricate id"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			ep, err := snowflakeEpoch(a.Str("epoch"))
			if err != nil {
				return nil, err
			}
			w, pr, sq := a.Int("worker"), a.Int("process"), a.Int("sequence")
			if w < 0 || w > 31 || pr < 0 || pr > 31 || sq < 0 || sq > 4095 {
				return nil, fmt.Errorf("worker and process must be 0..31, sequence 0..4095")
			}
			d := t.UnixMilli() - ep
			if d < 0 || d >= 1<<41 {
				return nil, fmt.Errorf("time is outside the 41-bit range of this epoch")
			}
			return strconv.FormatUint(uint64(d)<<22|uint64(w)<<17|uint64(pr)<<12|uint64(sq), 10), nil
		},
	})

	// ---- ObjectID ----
	reg(tools.Func{
		Name: "objectid_new", Title: "MongoDB ObjectID", Order: 33, Input: "time", Keywords: []string{"objectid", "mongo", "mongodb", "oid"},
		Doc:     "Generates a MongoDB ObjectID: 24 hex characters of a 4-byte Unix time in seconds, a 5-byte random value and a 3-byte counter started at a random number.",
		Params:  []tools.Param{timeParam("Instant to embed."), seedParam()},
		Returns: "str", Example: `uuid.objectid_new()`,
		Tags: []string{"mongodb id", "bson", "document id", "_id"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			if t.Unix() < 0 || t.Unix() > 0xffffffff {
				return nil, fmt.Errorf("time is outside the ObjectID range (1970 to 2106)")
			}
			r := newRng(a.Int("seed"))
			var b [12]byte
			binary.BigEndian.PutUint32(b[:4], uint32(t.Unix()))
			copy(b[4:], randBytes(r, 8))
			return hex.EncodeToString(b[:]), nil
		},
	})
	reg(tools.Func{
		Name: "objectid_parse", Title: "Inspect ObjectID", Order: 34, Input: "id", Accepts: acceptsObjectID,
		Doc:     "Decodes a MongoDB ObjectID into its UTC timestamp, the 5-byte random value (older drivers: 3-byte machine + 2-byte process) and the 3-byte counter.",
		Params:  []tools.Param{req("id", "str", "24-digit hex ObjectID.")},
		Returns: "dict[str, str]", Example: `uuid.objectid_parse("507f1f77bcf86cd799439011")`,
		Tags: []string{"decode objectid", "mongodb timestamp", "machine", "counter", "created at"},
		Fn: func(a tools.Args) (any, error) {
			b, err := decodeObjectID(a.Str("id"))
			if err != nil {
				return nil, err
			}
			t := time.Unix(int64(binary.BigEndian.Uint32(b[:4])), 0).UTC()
			return map[string]string{
				"timestamp": rfc(t), "unix": strconv.FormatInt(t.Unix(), 10),
				"random": hex.EncodeToString(b[4:9]), "machine": hex.EncodeToString(b[4:7]), "process": hex.EncodeToString(b[7:9]),
				"counter": strconv.Itoa(int(b[9])<<16 | int(b[10])<<8 | int(b[11])),
			}, nil
		},
	})

	// ---- nanoid, short uuid, cuid2, typeid ----
	reg(tools.Func{
		Name: "nanoid", Title: "Nano ID", Order: 7, Input: "alphabet", Keywords: []string{"nanoid", "short id"},
		Doc:     "Generates a NanoID of size characters (1..10000, default 21) from alphabet (2..255 distinct characters, default A-Za-z0-9_-) with unbiased selection; crypto/rand unless seed is set.",
		Params:  []tools.Param{p("size", "int", "Number of characters.", int64(21)), p("alphabet", "str", "Characters to draw from; empty means the default URL-safe alphabet.", ""), seedParam()},
		Returns: "str", Example: `uuid.nanoid(12)`,
		Tags: []string{"short id", "url safe id", "compact unique id", "random id"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("size")
			if n < 1 || n > maxCount {
				return nil, fmt.Errorf("size must be 1..%d", maxCount)
			}
			al := a.Str("alphabet")
			if al == "" {
				al = nanoidDefault
			}
			seen := map[rune]bool{}
			var rs []rune
			for _, c := range al {
				if !seen[c] {
					seen[c] = true
					rs = append(rs, c)
				}
			}
			if len(rs) < 2 || len(rs) > 255 {
				return nil, fmt.Errorf("alphabet needs 2..255 distinct characters")
			}
			return randomString(newRng(a.Int("seed")), rs, int(n)), nil
		},
	})
	reg(tools.Func{
		Name: "short_uuid", Title: "Short UUID (base58)", Order: 35, Input: "u", Keywords: []string{"short", "base58"},
		Doc:     "Returns the 22-character base58 spelling of a UUID, or of a fresh v4 UUID when u is empty. Reverse it with short_uuid_decode.",
		Params:  []tools.Param{p("u", "str", "UUID to shorten; empty generates a new v4.", "")},
		Returns: "str", Example: `uuid.short_uuid("3b241101-e2bb-4255-8caf-4136c566a962")`,
		Tags: []string{"shortid", "compact uuid", "url friendly", "base58 uuid"},
		Fn: func(a tools.Args) (any, error) {
			u := newV4()
			if strings.TrimSpace(a.Str("u")) != "" {
				var err error
				if u, err = parseUUID(a.Str("u")); err != nil {
					return nil, err
				}
			}
			return encodeBase(u.bigInt(), b58Alphabet, 22), nil
		},
	})
	reg(tools.Func{
		Name: "short_uuid_decode", Title: "Short UUID to UUID", Order: 36, Input: "s",
		Accepts: func(s string) bool { _, err := shortDecode(s); return err == nil },
		Doc:     "Expands a 22-character base58 short UUID (see short_uuid) back to a canonical UUID.",
		Params:  []tools.Param{req("s", "str", "22-character base58 string."), styleParam()},
		Returns: "str", Example: `uuid.short_uuid_decode("mhvXdrZT4jP5T8vBxuvm75")`,
		Tags: []string{"expand short uuid", "base58 decode", "shortid"},
		Fn: func(a tools.Args) (any, error) {
			u, err := shortDecode(a.Str("s"))
			if err != nil {
				return nil, err
			}
			return fmtOut(u, a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "cuid2", Title: "CUID2-style ID",
		Doc:     "Generates a CUID2-style identifier: a random lowercase letter followed by base36 digits of a SHA-256 hash over time, counter and random entropy (length 2..32, default 24). It follows the CUID2 layout but is not byte-compatible with the reference library.",
		Params:  []tools.Param{p("length", "int", "Total length 2..32.", int64(24)), seedParam()},
		Returns: "str", Example: `uuid.cuid2()`,
		Tags: []string{"collision resistant id", "cuid", "secure id", "horizontal scaling"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("length")
			if n < 2 || n > 32 {
				return nil, fmt.Errorf("length must be 2..32")
			}
			r := newRng(a.Int("seed"))
			h := sha256.New()
			var ts [8]byte
			binary.BigEndian.PutUint64(ts[:], uint64(time.Now().UnixNano()))
			h.Write(ts[:])
			h.Write(randBytes(r, 32))
			body := encodeBase(new(big.Int).SetBytes(h.Sum(nil)), "0123456789abcdefghijklmnopqrstuvwxyz", 50)
			return string(rune('a'+r.IntN(26))) + body[:n-1], nil
		},
	})
	reg(tools.Func{
		Name: "typeid_new", Title: "TypeID", Order: 37, Input: "prefix", Keywords: []string{"typeid", "stripe"},
		Doc:     "Generates a TypeID: a type prefix (lowercase letters and underscores, up to 63, may be empty) plus an underscore and a 26-character base32 UUID v7, e.g. user_01h455vb4pex5vsknk084sn02q. Pass time for a fixed instant.",
		Params:  []tools.Param{p("prefix", "str", "Type prefix such as user or order.", ""), timeParam("Instant to embed."), seedParam()},
		Returns: "str", Example: `uuid.typeid_new("user")`,
		Tags: []string{"prefixed id", "stripe style id", "typed id", "type safe id"},
		Fn: func(a tools.Args) (any, error) {
			if !typeidPrefixRe.MatchString(a.Str("prefix")) || len(a.Str("prefix")) > 63 {
				return nil, fmt.Errorf("prefix must be lowercase letters and underscores (max 63), starting and ending with a letter")
			}
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			var g v7gen
			return typeidEncode(a.Str("prefix"), g.next(newRng(a.Int("seed")), t.UnixMilli())), nil
		},
	})
	reg(tools.Func{
		Name: "typeid_parse", Title: "Inspect TypeID", Order: 38, Input: "typeid",
		Accepts: func(s string) bool { _, _, err := typeidDecode(s); return err == nil },
		Doc:     "Decodes a TypeID into its prefix, canonical UUID and (for v7) the embedded UTC timestamp. Errors on a bad prefix, wrong suffix length or overflow.",
		Params:  []tools.Param{req("typeid", "str", "TypeID such as user_01h455vb4pex5vsknk084sn02q.")},
		Returns: "dict[str, str]", Example: `uuid.typeid_parse("user_01h455vb4pex5vsknk084sn02q")`,
		Tags: []string{"decode typeid", "prefix", "uuid from typeid"},
		Fn: func(a tools.Args) (any, error) {
			prefix, u, err := typeidDecode(strings.TrimSpace(a.Str("typeid")))
			if err != nil {
				return nil, err
			}
			m := map[string]string{"prefix": prefix, "uuid": u.String(), "version": strconv.Itoa(u.version())}
			if t, ok := u.timestamp(); ok {
				m["timestamp"] = rfc(t)
			}
			return m, nil
		},
	})
	reg(tools.Func{
		Name: "uuid_to_typeid", Title: "UUID to TypeID", Input: "u", Accepts: acceptsUUID,
		Doc:     "Encodes a UUID as a TypeID with the given prefix (may be empty).",
		Params:  []tools.Param{req("u", "str", "UUID in any common notation."), p("prefix", "str", "Type prefix such as user.", "")},
		Returns: "str", Example: `uuid.uuid_to_typeid("01890a5d-ac96-774b-bcce-b302099a8057", "user")`,
		Tags: []string{"convert uuid", "typeid encode", "prefixed id"},
		Fn: func(a tools.Args) (any, error) {
			if !typeidPrefixRe.MatchString(a.Str("prefix")) || len(a.Str("prefix")) > 63 {
				return nil, fmt.Errorf("invalid TypeID prefix")
			}
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return nil, err
			}
			return typeidEncode(a.Str("prefix"), u), nil
		},
	})

	// ---- detection ----
	reg(tools.Func{
		Name: "id_detect", Title: "Which kind of ID is this?", Order: 39, Input: "s", Keywords: []string{"detect", "identify", "what"},
		Doc:     "Guesses which identifier a string is and lists every plausible kind, best first: UUID (with version), ULID, KSUID, MongoDB ObjectID, Snowflake, TypeID, short UUID, NanoID, CUID2 and common hex digests. Returns an empty list when nothing fits.",
		Params:  []tools.Param{req("s", "str", "The identifier to classify.")},
		Returns: "list[str]", Example: `uuid.id_detect("01ARZ3NDEKTSV4RRFFQ69G5FAV")`,
		Tags: []string{"identify id", "guess id type", "what kind of id", "uuid or ulid", "classify"},
		Fn:   func(a tools.Args) (any, error) { return detectID(strings.TrimSpace(a.Str("s"))), nil },
	})
}

func shortDecode(s string) (UUID, error) {
	var u UUID
	s = strings.TrimSpace(s)
	if len(s) != 22 {
		return u, fmt.Errorf("a short UUID has 22 characters, got %d", len(s))
	}
	n, err := decodeBase(s, b58Alphabet)
	if err != nil {
		return u, fmt.Errorf("not a short UUID: %w", err)
	}
	b, err := bigToBytes(n, 16)
	if err != nil {
		return u, fmt.Errorf("not a short UUID: value exceeds 128 bits")
	}
	copy(u[:], b)
	return u, nil
}

var (
	hexRe     = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	digitsRe  = regexp.MustCompile(`^[0-9]+$`)
	nanoidRe  = regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)
	cuid2Re   = regexp.MustCompile(`^[a-z][a-z0-9]{23}$`)
	base58Re  = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{22}$`)
	base62Re  = regexp.MustCompile(`^[0-9A-Za-z]{27}$`)
	crockRe   = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{26}$`)
	typeidSfx = regexp.MustCompile(`^[0-7][0-9a-hjkmnp-tv-z]{25}$`)
)

func detectID(s string) []string {
	var out []string
	if s == "" {
		return out
	}
	if u, err := parseUUID(s); err == nil {
		label := "uuid"
		switch {
		case u.isNil():
			label = "uuid nil"
		case u.isMax():
			label = "uuid max"
		case u.isRFC() && u.version() >= 1 && u.version() <= 8:
			label = fmt.Sprintf("uuid v%d", u.version())
			if t, ok := u.timestamp(); ok {
				label += " (" + rfc(t) + ")"
			}
		default:
			label = "uuid (" + u.variant() + " variant, version nibble " + strconv.Itoa(u.version()) + ")"
		}
		if len(s) == 32 {
			label += " without dashes"
		}
		out = append(out, label)
		if len(s) == 32 {
			out = append(out, "md5 digest (32 hex)")
		}
		return out
	}
	if prefix, u, err := typeidDecode(s); err == nil && prefix != "" {
		out = append(out, "typeid (prefix "+prefix+", uuid "+u.String()+")")
	}
	if b, err := decodeULID(s); err == nil && crockRe.MatchString(s) {
		out = append(out, "ulid ("+rfc(ulidTime(b))+")")
		if typeidSfx.MatchString(strings.ToLower(s)) {
			out = append(out, "typeid without prefix")
		}
	}
	if base62Re.MatchString(s) {
		if b, err := decodeKSUID(s); err == nil {
			if t := ksuidTime(b); t.Year() >= 2014 && t.Year() <= 2100 {
				out = append(out, "ksuid ("+rfc(t)+")")
			}
		}
	}
	if len(s) == 24 && hexRe.MatchString(s) {
		b, _ := decodeObjectID(s)
		t := time.Unix(int64(binary.BigEndian.Uint32(b[:4])), 0).UTC()
		if t.Year() >= 2000 && t.Year() <= 2100 {
			out = append(out, "mongodb objectid ("+rfc(t)+")")
		}
	}
	if digitsRe.MatchString(s) && len(s) >= 15 && len(s) <= 20 {
		if id, err := strconv.ParseUint(s, 10, 64); err == nil && id>>22 > 0 {
			dis := time.UnixMilli(int64(id>>22) + snowflakeEpochs["discord"]).UTC()
			tw := time.UnixMilli(int64(id>>22) + snowflakeEpochs["twitter"]).UTC()
			if dis.Year() >= 2015 && dis.Year() <= 2100 {
				out = append(out, "snowflake, discord epoch ("+rfc(dis)+")")
			}
			if tw.Year() >= 2010 && tw.Year() <= 2100 {
				out = append(out, "snowflake, twitter epoch ("+rfc(tw)+")")
			}
		}
	}
	if base58Re.MatchString(s) {
		if _, err := shortDecode(s); err == nil {
			out = append(out, "short uuid (base58)")
		}
	}
	if cuid2Re.MatchString(s) {
		out = append(out, "cuid2 (24 chars)")
	}
	if nanoidRe.MatchString(s) {
		out = append(out, "nanoid (default 21 chars)")
	}
	if hexRe.MatchString(s) {
		switch len(s) {
		case 40:
			out = append(out, "sha-1 digest (40 hex)")
		case 64:
			out = append(out, "sha-256 digest (64 hex)")
		}
	}
	return out
}
