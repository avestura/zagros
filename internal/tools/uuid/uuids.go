package uuid

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

// v7gen produces monotonically increasing UUID v7 values (RFC 9562 method 1).
type v7gen struct {
	started bool
	ms      int64
	a       uint16
	b       uint64
}

func (g *v7gen) next(r *rand.Rand, ms int64) UUID {
	if g.started && ms <= g.ms {
		ms = g.ms
		g.b = (g.b + 1) & 0x3fffffffffffffff
		if g.b == 0 {
			g.a = (g.a + 1) & 0x0fff
			if g.a == 0 {
				ms++
				g.a, g.b = uint16(r.Uint32())&0x0fff, r.Uint64()&0x3fffffffffffffff
			}
		}
	} else {
		g.a, g.b = uint16(r.Uint32())&0x0fff, r.Uint64()&0x3fffffffffffffff
	}
	g.started, g.ms = true, ms
	return buildV7(ms, g.a, g.b)
}

func randNode(r *rand.Rand) [6]byte {
	var n [6]byte
	copy(n[:], randBytes(r, 6))
	n[0] |= 0x01 // multicast bit: marks a random, non-MAC node (RFC 9562 section 6.10)
	return n
}

func newV1(t time.Time, r *rand.Rand) UUID {
	return buildV1(toGregorian(t), uint16(r.Uint32()), randNode(r))
}

func newV6(t time.Time, r *rand.Rand) UUID {
	return buildV6(toGregorian(t), uint16(r.Uint32()), randNode(r))
}

func newV4() UUID {
	r := newRng(0)
	var u UUID
	copy(u[:], randBytes(r, 16))
	return withVersion(u, 4)
}

func parseNamespace(s string) (UUID, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dns":
		return nsDNS, nil
	case "url":
		return nsURL, nil
	case "oid":
		return nsOID, nil
	case "x500", "x.500":
		return nsX500, nil
	case "nil":
		return UUID{}, nil
	}
	u, err := parseUUID(s)
	if err != nil {
		return u, fmt.Errorf("namespace must be dns, url, oid, x500 or a UUID: %w", err)
	}
	return u, nil
}

func newNameBased(version int, ns UUID, name string) UUID {
	var u UUID
	if version == 3 {
		h := md5.New()
		h.Write(ns[:])
		h.Write([]byte(name))
		copy(u[:], h.Sum(nil))
	} else {
		h := sha1.New()
		h.Write(ns[:])
		h.Write([]byte(name))
		copy(u[:], h.Sum(nil)[:16])
	}
	return withVersion(u, byte(version))
}

func fmtOut(u UUID, style string) (any, error) {
	s, err := formatUUID(u, style)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func acceptsUUID(in string) bool { _, err := parseUUID(in); return err == nil }

// generate creates a UUID of the named version for uuid_bulk.
func generate(version string, g *v7gen, r *rand.Rand) (UUID, error) {
	v := strings.ToLower(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(version), "uuid"), "v"))
	switch v {
	case "1":
		return newV1(time.Now().UTC(), r), nil
	case "4", "":
		return newV4(), nil
	case "6":
		return newV6(time.Now().UTC(), r), nil
	case "7":
		return g.next(r, time.Now().UnixMilli()), nil
	case "nil", "0":
		return UUID{}, nil
	case "max":
		return mustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), nil
	}
	return UUID{}, fmt.Errorf("version must be 1, 4, 6, 7, nil or max (v3 and v5 need a name)")
}

func init() {
	reg(tools.Func{
		Name: "uuid_v4", Title: "UUID v4 (random)", Order: 1, Input: "style", Keywords: []string{"v4", "random"},
		Doc:     "Generates a random version 4 UUID (RFC 9562) from crypto/rand, e.g. 3b241101-e2bb-4255-8caf-4136c566a962. Style selects the text form.",
		Params:  []tools.Param{styleParam()},
		Returns: "str", Example: `uuid.uuid_v4()`,
		Tags: []string{"guid", "random uuid", "unique id", "generate", "rfc 4122", "identifier"},
		Fn:   func(a tools.Args) (any, error) { return fmtOut(newV4(), a.Str("style")) },
	})
	reg(tools.Func{
		Name: "uuid_v7", Title: "UUID v7 (time-ordered)", Order: 2, Input: "time", Keywords: []string{"v7", "sortable", "time"},
		Doc:     "Generates a version 7 UUID (RFC 9562): 48-bit Unix millisecond time followed by random bits, so values sort by creation time and index well in databases. Pass time for a fixed instant.",
		Params:  []tools.Param{timeParam("Instant to embed."), styleParam(), seedParam()},
		Returns: "str", Example: `uuid.uuid_v7()`,
		Tags: []string{"time ordered uuid", "sortable uuid", "database primary key", "monotonic", "rfc 9562"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			var g v7gen
			return fmtOut(g.next(newRng(a.Int("seed")), t.UnixMilli()), a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_v1", Title: "UUID v1 (time + node)", Order: 8, Input: "time", Keywords: []string{"v1"},
		Doc:     "Generates a version 1 UUID: 60-bit timestamp in 100 ns ticks since 1582-10-15, a random 14-bit clock sequence and a random node ID with the multicast bit set (so no real MAC address leaks).",
		Params:  []tools.Param{timeParam("Instant to embed."), styleParam(), seedParam()},
		Returns: "str", Example: `uuid.uuid_v1()`,
		Tags: []string{"time based uuid", "timestamp uuid", "mac address", "rfc 4122", "gregorian"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			return fmtOut(newV1(t, newRng(a.Int("seed"))), a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_v6", Title: "UUID v6 (reordered time)", Order: 9, Input: "time", Keywords: []string{"v6"},
		Doc:     "Generates a version 6 UUID (RFC 9562): the v1 timestamp reordered most-significant-first so it sorts chronologically as text or bytes.",
		Params:  []tools.Param{timeParam("Instant to embed."), styleParam(), seedParam()},
		Returns: "str", Example: `uuid.uuid_v6()`,
		Tags: []string{"sortable uuid", "reordered v1", "time ordered", "rfc 9562"},
		Fn: func(a tools.Args) (any, error) {
			t, err := parseTime(a.Str("time"))
			if err != nil {
				return nil, err
			}
			return fmtOut(newV6(t, newRng(a.Int("seed"))), a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_v8", Title: "UUID v8 (custom)",
		Doc:     "Builds a version 8 UUID (RFC 9562, vendor-specific) from up to 16 bytes given as hex (shorter input is zero padded); version and variant bits overwrite the corresponding bits.",
		Params:  []tools.Param{req("data", "str", "Up to 32 hex digits of custom data."), styleParam()},
		Returns: "str", Example: `uuid.uuid_v8("00112233445566778899aabbccddeeff")`,
		Tags: []string{"custom uuid", "vendor", "rfc 9562", "experimental"},
		Fn: func(a tools.Args) (any, error) {
			h := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(a.Str("data")), " ", ""), "-", "")
			if len(h) > 32 || len(h)%2 != 0 {
				return nil, fmt.Errorf("data must be an even number of hex digits, at most 32")
			}
			b, err := hex.DecodeString(h)
			if err != nil {
				return nil, fmt.Errorf("data is not valid hex")
			}
			var u UUID
			copy(u[:], b)
			return fmtOut(withVersion(u, 8), a.Str("style"))
		},
	})
	nameBased := func(v int, hash string) {
		reg(tools.Func{
			Name: "uuid_v" + strconv.Itoa(v), Title: fmt.Sprintf("UUID v%d (%s name-based)", v, hash), Order: 10 + v, Input: "name", Keywords: []string{"v" + strconv.Itoa(v), "namespace"},
			Doc: fmt.Sprintf("Generates the deterministic version %d UUID (%s) of name inside a namespace: the same inputs always give the same UUID. Namespace is dns, url, oid, x500 or any UUID.", v, hash),
			Params: []tools.Param{req("namespace", "str", "dns, url, oid, x500 or a namespace UUID."), req("name", "str", "Name to hash, e.g. a domain or URL."),
				styleParam()},
			Returns: "str", Example: fmt.Sprintf(`uuid.uuid_v%d("dns", "example.com")`, v),
			Tags: []string{"deterministic uuid", "namespace uuid", "hash uuid", hash, "stable id", "rfc 4122"},
			Fn: func(a tools.Args) (any, error) {
				ns, err := parseNamespace(a.Str("namespace"))
				if err != nil {
					return nil, err
				}
				return fmtOut(newNameBased(v, ns, a.Str("name")), a.Str("style"))
			},
		})
	}
	nameBased(3, "md5")
	nameBased(5, "sha1")
	reg(tools.Func{
		Name: "uuid_nil", Title: "Nil UUID",
		Doc:     "Returns the nil UUID 00000000-0000-0000-0000-000000000000 (all bits zero).",
		Returns: "str", Example: `uuid.uuid_nil()`,
		Tags: []string{"empty uuid", "zero uuid", "null guid"},
		Fn:   func(a tools.Args) (any, error) { return UUID{}.String(), nil },
	})
	reg(tools.Func{
		Name: "uuid_max", Title: "Max UUID",
		Doc:     "Returns the max UUID ffffffff-ffff-ffff-ffff-ffffffffffff (all bits one, RFC 9562).",
		Returns: "str", Example: `uuid.uuid_max()`,
		Tags: []string{"maximum uuid", "all ones", "sentinel"},
		Fn:   func(a tools.Args) (any, error) { return "ffffffff-ffff-ffff-ffff-ffffffffffff", nil },
	})
	reg(tools.Func{
		Name: "namespaces", Title: "Predefined namespaces",
		Doc:     "Returns the predefined namespace UUIDs for name-based UUIDs: dns, url, oid and x500 (RFC 4122 appendix C).",
		Returns: "dict[str, str]", Example: `uuid.namespaces()`,
		Tags: []string{"namespace constants", "dns", "url", "oid", "x500", "uuid v5"},
		Fn: func(a tools.Args) (any, error) {
			return map[string]string{"dns": nsDNS.String(), "url": nsURL.String(), "oid": nsOID.String(), "x500": nsX500.String()}, nil
		},
	})
	reg(tools.Func{
		Name: "uuid_parse", Title: "Inspect UUID", Order: 3, Input: "u", Accepts: acceptsUUID, Keywords: []string{"parse", "inspect", "decode"},
		Doc:     "Parses a UUID (canonical, upper case, braces, urn:uuid: or 32 hex digits) and returns version, variant, canonical and urn forms, integer value and hex; for v1/v6/v7 also the UTC timestamp, and for v1/v6 node and clock_seq. Errors if the text is not a UUID.",
		Params:  []tools.Param{req("u", "str", "UUID in any common notation.")},
		Returns: "dict[str, str]", Example: `uuid.uuid_parse("017f22e2-79b0-7cc3-98c4-dc0c0c07398f")`,
		Tags: []string{"decode uuid", "uuid version", "uuid timestamp", "analyze", "guid info", "extract time"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return nil, err
			}
			m := map[string]string{
				"canonical": u.String(), "urn": "urn:uuid:" + u.String(), "hex": hex.EncodeToString(u[:]),
				"int": u.bigInt().String(), "variant": u.variant(), "version": strconv.Itoa(u.version()),
			}
			switch {
			case u.isNil():
				m["version"], m["kind"] = "nil", "nil UUID"
			case u.isMax():
				m["version"], m["kind"] = "max", "max UUID"
			}
			if t, ok := u.timestamp(); ok {
				m["timestamp"] = rfc(t)
				m["unix_ms"] = strconv.FormatInt(t.UnixMilli(), 10)
			}
			if u.isRFC() && (u.version() == 1 || u.version() == 6) {
				m["node"] = fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", u[10], u[11], u[12], u[13], u[14], u[15])
				m["clock_seq"] = strconv.Itoa(int(u[8]&0x3f)<<8 | int(u[9]))
				m["node_multicast"] = strconv.FormatBool(u[10]&1 == 1)
			}
			return m, nil
		},
	})
	reg(tools.Func{
		Name: "uuid_validate", Title: "Is it a UUID?", Order: 20, Input: "u", Accepts: acceptsUUID,
		Doc:     "Returns true when the text is a valid UUID in a common notation. With version set (1..8) it must also carry that version and the RFC variant.",
		Params:  []tools.Param{req("u", "str", "Text to check."), p("version", "int", "Required version 1..8, or 0 for any.", int64(0))},
		Returns: "bool", Example: `uuid.uuid_validate("not-a-uuid")`,
		Tags: []string{"check uuid", "is uuid", "guid validation", "verify"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return false, nil
			}
			if v := a.Int("version"); v != 0 {
				return u.isRFC() && int64(u.version()) == v, nil
			}
			return true, nil
		},
	})
	reg(tools.Func{
		Name: "uuid_format", Title: "Reformat UUID", Order: 4, Input: "u", Accepts: acceptsUUID, Keywords: []string{"format", "convert"},
		Doc:     "Converts a UUID to another text form: canonical, upper, no_dashes, braces, urn, base64 (22 chars URL-safe), base58 (22 chars) or base32 (26 chars).",
		Params:  []tools.Param{req("u", "str", "UUID in any common notation."), styleParam()},
		Returns: "str", Example: `uuid.uuid_format("3b241101-e2bb-4255-8caf-4136c566a962", "braces")`,
		Tags: []string{"uppercase uuid", "remove dashes", "guid braces", "urn", "base64 uuid", "base58 uuid", "short uuid"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return nil, err
			}
			return fmtOut(u, a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_to_int", Title: "UUID to integer", Order: 21, Input: "u", Accepts: acceptsUUID,
		Doc:     "Returns the UUID as a 128-bit unsigned decimal integer string (an int would overflow).",
		Params:  []tools.Param{req("u", "str", "UUID in any common notation.")},
		Returns: "str", Example: `uuid.uuid_to_int("00000000-0000-0000-0000-0000000000ff")`,
		Tags: []string{"uuid number", "decimal", "128 bit", "bigint"},
		Fn: func(a tools.Args) (any, error) {
			u, err := parseUUID(a.Str("u"))
			if err != nil {
				return nil, err
			}
			return u.bigInt().String(), nil
		},
	})
	reg(tools.Func{
		Name: "uuid_from_int", Title: "Integer to UUID",
		Doc:     "Converts a decimal (or 0x-prefixed hex) integer in 0..2^128-1 to a canonical UUID.",
		Params:  []tools.Param{req("value", "str", "Decimal or 0x hex integer."), styleParam()},
		Returns: "str", Example: `uuid.uuid_from_int("255")`,
		Tags: []string{"number to uuid", "bigint", "128 bit"},
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("value"))
			n, ok := new(big.Int).SetString(s, 0)
			if !ok || n.Sign() < 0 {
				return nil, fmt.Errorf("%q is not a non-negative integer", s)
			}
			b, err := bigToBytes(n, 16)
			if err != nil {
				return nil, fmt.Errorf("%q does not fit in 128 bits", s)
			}
			var u UUID
			copy(u[:], b)
			return fmtOut(u, a.Str("style"))
		},
	})
	reg(tools.Func{
		Name: "uuid_bulk", Title: "Many UUIDs",
		Doc:     "Generates count UUIDs (0..10000) of version 1, 4, 6 or 7 (or nil/max) in the chosen style. v7 values are strictly increasing within the call.",
		Params:  []tools.Param{p("count", "int", "How many UUIDs.", int64(10)), p("version", "str", "1, 4, 6, 7, nil or max.", "4"), styleParam()},
		Returns: "list[str]", Example: `uuid.uuid_bulk(5, "7")`,
		Tags: []string{"batch", "many uuids", "list of uuids", "generate multiple", "bulk"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("count")
			if n < 0 || n > maxCount {
				return nil, fmt.Errorf("count must be 0..%d", maxCount)
			}
			var g v7gen
			r := newRng(0)
			out := make([]string, n)
			for i := range out {
				u, err := generate(a.Str("version"), &g, r)
				if err != nil {
					return nil, err
				}
				if out[i], err = formatUUID(u, a.Str("style")); err != nil {
					return nil, err
				}
			}
			return out, nil
		},
	})
	cmpBy := p("by", "str", "bytes (plain byte order) or time (embedded v1/v6/v7 timestamp; UUIDs without one sort last).", "bytes")
	less := func(by string) (func(a, b UUID) int, error) {
		switch strings.ToLower(by) {
		case "bytes", "":
			return func(a, b UUID) int { return strings.Compare(string(a[:]), string(b[:])) }, nil
		case "time":
			return func(a, b UUID) int {
				ta, oka := a.timestamp()
				tb, okb := b.timestamp()
				switch {
				case oka && okb:
					if c := ta.Compare(tb); c != 0 {
						return c
					}
					return strings.Compare(string(a[:]), string(b[:]))
				case oka:
					return -1
				case okb:
					return 1
				}
				return strings.Compare(string(a[:]), string(b[:]))
			}, nil
		}
		return nil, fmt.Errorf("by must be bytes or time")
	}
	reg(tools.Func{
		Name: "uuid_compare", Title: "Compare UUIDs",
		Doc:     "Compares two UUIDs and returns -1, 0 or 1. by=bytes compares raw bytes (correct creation order for v6/v7); by=time compares the embedded timestamps (also correct for v1).",
		Params:  []tools.Param{req("a", "str", "First UUID."), req("b", "str", "Second UUID."), cmpBy},
		Returns: "int", Example: `uuid.uuid_compare("017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "017f22e2-79b1-7000-8000-000000000000")`,
		Tags: []string{"order", "sort order", "earlier", "later", "which is newer"},
		Fn: func(a tools.Args) (any, error) {
			x, err := parseUUID(a.Str("a"))
			if err != nil {
				return nil, err
			}
			y, err := parseUUID(a.Str("b"))
			if err != nil {
				return nil, err
			}
			f, err := less(a.Str("by"))
			if err != nil {
				return nil, err
			}
			return int64(f(x, y)), nil
		},
	})
	reg(tools.Func{
		Name: "uuid_sort", Title: "Sort UUIDs",
		Doc:     "Sorts UUIDs (canonicalized) by raw bytes or by embedded timestamp; v7 UUIDs come out in creation order either way. Errors on the first invalid UUID. Set reverse for newest first.",
		Params:  []tools.Param{req("uuids", "list[str]", "UUIDs in any common notation."), cmpBy, p("reverse", "bool", "Descending order.", false)},
		Returns: "list[str]", Example: `uuid.uuid_sort(["017f22e2-79b1-7000-8000-000000000000", "017f22e2-79b0-7cc3-98c4-dc0c0c07398f"])`,
		Tags: []string{"order uuids", "chronological", "sort by time"},
		Fn: func(a tools.Args) (any, error) {
			f, err := less(a.Str("by"))
			if err != nil {
				return nil, err
			}
			us := make([]UUID, 0, len(a.List("uuids")))
			for _, s := range a.List("uuids") {
				u, err := parseUUID(s)
				if err != nil {
					return nil, err
				}
				us = append(us, u)
			}
			sort.SliceStable(us, func(i, j int) bool {
				if a.Bool("reverse") {
					return f(us[i], us[j]) > 0
				}
				return f(us[i], us[j]) < 0
			})
			out := make([]string, len(us))
			for i, u := range us {
				out[i] = u.String()
			}
			return out, nil
		},
	})
}
