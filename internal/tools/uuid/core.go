package uuid

import (
	crand "crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

const maxCount = 10000

// UUID is the 16 raw bytes of a UUID.
type UUID [16]byte

const gregorianOffset = 122192928000000000 // 100ns ticks from 1582-10-15 to 1970-01-01

var (
	nsDNS  = mustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	nsURL  = mustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	nsOID  = mustParse("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	nsX500 = mustParse("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
)

func mustParse(s string) UUID {
	u, err := parseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// parseUUID accepts canonical, upper case, braces, urn:uuid: and 32 hex digits.
func parseUUID(s string) (UUID, error) {
	var u UUID
	t := strings.TrimSpace(s)
	low := strings.ToLower(t)
	low = strings.TrimPrefix(low, "urn:uuid:")
	low = strings.TrimPrefix(low, "uuid:")
	if strings.HasPrefix(low, "{") && strings.HasSuffix(low, "}") {
		low = low[1 : len(low)-1]
	}
	switch len(low) {
	case 36:
		if low[8] != '-' || low[13] != '-' || low[18] != '-' || low[23] != '-' {
			return u, fmt.Errorf("%q is not a UUID: dashes are misplaced", s)
		}
		low = low[0:8] + low[9:13] + low[14:18] + low[19:23] + low[24:]
	case 32:
	default:
		return u, fmt.Errorf("%q is not a UUID: expected 32 hex digits", s)
	}
	b, err := hex.DecodeString(low)
	if err != nil {
		return u, fmt.Errorf("%q is not a UUID: invalid hex digits", s)
	}
	copy(u[:], b)
	return u, nil
}

func (u UUID) String() string {
	h := hex.EncodeToString(u[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

func (u UUID) isNil() bool { return u == UUID{} }
func (u UUID) isMax() bool {
	for _, b := range u {
		if b != 0xff {
			return false
		}
	}
	return true
}

func (u UUID) version() int { return int(u[6] >> 4) }

func (u UUID) variant() string {
	switch b := u[8]; {
	case b&0x80 == 0:
		return "NCS (reserved)"
	case b&0xc0 == 0x80:
		return "RFC 4122/9562"
	case b&0xe0 == 0xc0:
		return "Microsoft"
	}
	return "reserved (future)"
}

func (u UUID) isRFC() bool { return u[8]&0xc0 == 0x80 }

func (u UUID) bigInt() *big.Int { return new(big.Int).SetBytes(u[:]) }

// timestamp extracts the embedded time of v1, v6 and v7 UUIDs.
func (u UUID) timestamp() (time.Time, bool) {
	if !u.isRFC() {
		return time.Time{}, false
	}
	switch u.version() {
	case 1:
		low := uint64(binary.BigEndian.Uint32(u[0:4]))
		mid := uint64(binary.BigEndian.Uint16(u[4:6]))
		hi := uint64(binary.BigEndian.Uint16(u[6:8]) & 0x0fff)
		return fromGregorian(hi<<48 | mid<<32 | low), true
	case 6:
		high := uint64(binary.BigEndian.Uint32(u[0:4]))
		mid := uint64(binary.BigEndian.Uint16(u[4:6]))
		low := uint64(binary.BigEndian.Uint16(u[6:8]) & 0x0fff)
		return fromGregorian(high<<28 | mid<<12 | low), true
	case 7:
		ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
		return time.UnixMilli(ms).UTC(), true
	}
	return time.Time{}, false
}

func fromGregorian(ts uint64) time.Time {
	u := int64(ts) - gregorianOffset // 100ns ticks since the Unix epoch
	sec := u / 1e7
	rem := u % 1e7
	if rem < 0 {
		sec--
		rem += 1e7
	}
	return time.Unix(sec, rem*100).UTC()
}

func toGregorian(t time.Time) uint64 {
	return uint64(t.Unix()*1e7 + int64(t.Nanosecond()/100) + gregorianOffset)
}

func buildV1(ts uint64, clockSeq uint16, node [6]byte) UUID {
	var u UUID
	binary.BigEndian.PutUint32(u[0:4], uint32(ts))
	binary.BigEndian.PutUint16(u[4:6], uint16(ts>>32))
	binary.BigEndian.PutUint16(u[6:8], uint16(ts>>48)&0x0fff|0x1000)
	binary.BigEndian.PutUint16(u[8:10], 0x8000|clockSeq&0x3fff)
	copy(u[10:], node[:])
	return u
}

func buildV6(ts uint64, clockSeq uint16, node [6]byte) UUID {
	var u UUID
	binary.BigEndian.PutUint32(u[0:4], uint32(ts>>28))
	binary.BigEndian.PutUint16(u[4:6], uint16(ts>>12))
	binary.BigEndian.PutUint16(u[6:8], uint16(ts)&0x0fff|0x6000)
	binary.BigEndian.PutUint16(u[8:10], 0x8000|clockSeq&0x3fff)
	copy(u[10:], node[:])
	return u
}

// buildV7 packs a Unix millisecond time, 12 bits randA and 62 bits randB.
func buildV7(ms int64, randA uint16, randB uint64) UUID {
	var u UUID
	u[0], u[1], u[2], u[3], u[4], u[5] = byte(ms>>40), byte(ms>>32), byte(ms>>24), byte(ms>>16), byte(ms>>8), byte(ms)
	binary.BigEndian.PutUint16(u[6:8], 0x7000|randA&0x0fff)
	binary.BigEndian.PutUint64(u[8:16], 0x8000000000000000|randB&0x3fffffffffffffff)
	return u
}

func withVersion(u UUID, v byte) UUID {
	u[6] = u[6]&0x0f | v<<4
	u[8] = u[8]&0x3f | 0x80
	return u
}

// ---- randomness ----

type cryptoSrc struct{}

func (cryptoSrc) Uint64() uint64 {
	var b [8]byte
	_, _ = crand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

func newRng(seed int64) *rand.Rand {
	if seed == 0 {
		return rand.New(cryptoSrc{})
	}
	return rand.New(rand.NewPCG(uint64(seed), 0x9e3779b97f4a7c15))
}

func randBytes(r *rand.Rand, n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.Uint32())
	}
	return b
}

// ---- helpers shared by registrations ----

func p(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

func seedParam() tools.Param {
	return p("seed", "int", "0 (default) draws randomness from crypto/rand; any other value makes the random parts reproducible.", int64(0))
}

func timeParam(doc string) tools.Param {
	return p("time", "str", doc+" RFC 3339, YYYY-MM-DD, or Unix seconds/milliseconds; empty means now.", "")
}

var keywords = []string{"uuid", "guid", "ulid", "ksuid", "nanoid", "snowflake", "objectid"}

func reg(f tools.Func) {
	f.Group = "uuid"
	if f.Order == 0 {
		f.Order = 100 // unset: behind the headline generators in the launcher
	}
	if f.Input != "" {
		f.Family = "uuid"
		f.Keywords = append(append([]string(nil), keywords...), f.Keywords...)
	}
	tools.Register(f)
}

// parseTime reads an instant; empty means now.
func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().UTC(), nil
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && len(strings.TrimLeft(s, "-")) != 4 {
		if n > 99999999999 || n < -99999999999 { // milliseconds
			return time.UnixMilli(n).UTC(), nil
		}
		return time.Unix(n, 0).UTC(), nil
	}
	for _, l := range []string{time.RFC3339Nano, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", time.RFC1123Z, time.RFC1123} {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot read time %q (use RFC 3339, YYYY-MM-DD or Unix seconds/milliseconds)", s)
}

func rfc(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z07:00") }

// ---- encodings ----

const (
	b58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	b62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	crockford   = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

func encodeBase(n *big.Int, alphabet string, pad int) string {
	base := big.NewInt(int64(len(alphabet)))
	x := new(big.Int).Set(n)
	var out []byte
	m := new(big.Int)
	for x.Sign() > 0 {
		x.DivMod(x, base, m)
		out = append(out, alphabet[m.Int64()])
	}
	for len(out) < pad {
		out = append(out, alphabet[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func decodeBase(s, alphabet string) (*big.Int, error) {
	n := new(big.Int)
	base := big.NewInt(int64(len(alphabet)))
	for _, c := range s {
		i := strings.IndexRune(alphabet, c)
		if i < 0 {
			return nil, fmt.Errorf("invalid character %q", c)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(i)))
	}
	return n, nil
}

func bigToBytes(n *big.Int, size int) ([]byte, error) {
	if n.BitLen() > size*8 {
		return nil, errors.New("value does not fit")
	}
	return n.FillBytes(make([]byte, size)), nil
}

// formatUUID renders u in one of the supported styles.
func formatUUID(u UUID, style string) (string, error) {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(style), "-", "_")) {
	case "", "canonical", "lower", "default":
		return u.String(), nil
	case "upper", "uppercase":
		return strings.ToUpper(u.String()), nil
	case "no_dashes", "nodashes", "hex", "compact", "plain":
		return hex.EncodeToString(u[:]), nil
	case "braces", "guid", "microsoft":
		return "{" + u.String() + "}", nil
	case "urn":
		return "urn:uuid:" + u.String(), nil
	case "base64", "base64url":
		return base64.RawURLEncoding.EncodeToString(u[:]), nil
	case "base58", "short":
		return encodeBase(u.bigInt(), b58Alphabet, 22), nil
	case "base32":
		return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(u[:]), nil
	}
	return "", fmt.Errorf("style must be canonical, upper, no_dashes, braces, urn, base64, base58 or base32")
}

var styleChoices = []string{"canonical", "upper", "no_dashes", "braces", "urn", "base64", "base58", "base32"}

func styleParam() tools.Param {
	pp := p("style", "str", "Output style: canonical, upper, no_dashes, braces, urn, base64 (unpadded URL-safe), base58 or base32.", "canonical")
	pp.Choices = styleChoices
	return pp
}
