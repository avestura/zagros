package hashx

import (
	"crypto/des"
	"crypto/sha1"
	"encoding/binary"
	"hash"
	"math/bits"
	"strings"
	"unicode/utf16"

	"golang.org/x/crypto/md4"
)

// bufHash adapts a one-shot function to hash.Hash (used for HMAC).
type bufHash struct {
	buf   []byte
	f     func([]byte) []byte
	size  int
	block int
}

func (h *bufHash) Write(p []byte) (int, error) { h.buf = append(h.buf, p...); return len(p), nil }
func (h *bufHash) Sum(b []byte) []byte         { return append(b, h.f(h.buf)...) }
func (h *bufHash) Reset()                      { h.buf = nil }
func (h *bufHash) Size() int                   { return h.size }
func (h *bufHash) BlockSize() int              { return h.block }

func newSM3() hash.Hash { return &bufHash{f: sm3Sum, size: 32, block: 64} }

// ---- SM3 (GB/T 32905-2016) ----

func sm3Sum(data []byte) []byte {
	v := [8]uint32{0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600, 0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e}
	msg := append([]byte(nil), data...)
	msg = append(msg, 0x80)
	for len(msg)%64 != 56 {
		msg = append(msg, 0)
	}
	msg = binary.BigEndian.AppendUint64(msg, uint64(len(data))*8)
	p0 := func(x uint32) uint32 { return x ^ bits.RotateLeft32(x, 9) ^ bits.RotateLeft32(x, 17) }
	p1 := func(x uint32) uint32 { return x ^ bits.RotateLeft32(x, 15) ^ bits.RotateLeft32(x, 23) }
	for off := 0; off < len(msg); off += 64 {
		var w [68]uint32
		var wp [64]uint32
		for j := 0; j < 16; j++ {
			w[j] = binary.BigEndian.Uint32(msg[off+4*j:])
		}
		for j := 16; j < 68; j++ {
			w[j] = p1(w[j-16]^w[j-9]^bits.RotateLeft32(w[j-3], 15)) ^ bits.RotateLeft32(w[j-13], 7) ^ w[j-6]
		}
		for j := 0; j < 64; j++ {
			wp[j] = w[j] ^ w[j+4]
		}
		a, b, c, d, e, f, g, h := v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7]
		for j := 0; j < 64; j++ {
			t := uint32(0x79cc4519)
			if j >= 16 {
				t = 0x7a879d8a
			}
			ss1 := bits.RotateLeft32(bits.RotateLeft32(a, 12)+e+bits.RotateLeft32(t, j%32), 7)
			ss2 := ss1 ^ bits.RotateLeft32(a, 12)
			var ff, gg uint32
			if j < 16 {
				ff = a ^ b ^ c
				gg = e ^ f ^ g
			} else {
				ff = (a & b) | (a & c) | (b & c)
				gg = (e & f) | (^e & g)
			}
			tt1 := ff + d + ss2 + wp[j]
			tt2 := gg + h + ss1 + w[j]
			d = c
			c = bits.RotateLeft32(b, 9)
			b = a
			a = tt1
			h = g
			g = bits.RotateLeft32(f, 19)
			f = e
			e = p0(tt2)
		}
		v[0] ^= a
		v[1] ^= b
		v[2] ^= c
		v[3] ^= d
		v[4] ^= e
		v[5] ^= f
		v[6] ^= g
		v[7] ^= h
	}
	out := make([]byte, 32)
	for i, x := range v {
		binary.BigEndian.PutUint32(out[4*i:], x)
	}
	return out
}

// ---- Fletcher ----

func fletcher16(data []byte) []byte {
	var s1, s2 uint32
	for _, b := range data {
		s1 = (s1 + uint32(b)) % 255
		s2 = (s2 + s1) % 255
	}
	return []byte{byte(s2), byte(s1)}
}

func fletcher32(data []byte) []byte {
	var s1, s2 uint32
	for i := 0; i < len(data); i += 2 {
		w := uint32(data[i])
		if i+1 < len(data) {
			w |= uint32(data[i+1]) << 8
		}
		s1 = (s1 + w) % 65535
		s2 = (s2 + s1) % 65535
	}
	return binary.BigEndian.AppendUint32(nil, s2<<16|s1)
}

// ---- Murmur3 ----

func murmur3x86_32(data []byte, seed uint32) uint32 {
	const c1, c2 = 0xcc9e2d51, 0x1b873593
	h := seed
	n := len(data) / 4
	for i := 0; i < n; i++ {
		k := binary.LittleEndian.Uint32(data[4*i:])
		k *= c1
		k = bits.RotateLeft32(k, 15)
		k *= c2
		h ^= k
		h = bits.RotateLeft32(h, 13)
		h = h*5 + 0xe6546b64
	}
	tail := data[4*n:]
	var k uint32
	switch len(tail) {
	case 3:
		k ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(tail[0])
		k *= c1
		k = bits.RotateLeft32(k, 15)
		k *= c2
		h ^= k
	}
	h ^= uint32(len(data))
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

func murmur3x64_128(data []byte, seed uint32) (uint64, uint64) {
	const c1, c2 = 0x87c37b91114253d5, 0x4cf5ad432745937f
	h1, h2 := uint64(seed), uint64(seed)
	n := len(data) / 16
	for i := 0; i < n; i++ {
		k1 := binary.LittleEndian.Uint64(data[16*i:])
		k2 := binary.LittleEndian.Uint64(data[16*i+8:])
		k1 *= c1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2
		h1 ^= k1
		h1 = bits.RotateLeft64(h1, 27)
		h1 += h2
		h1 = h1*5 + 0x52dce729
		k2 *= c2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1
		h2 ^= k2
		h2 = bits.RotateLeft64(h2, 31)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}
	tail := data[16*n:]
	var k1, k2 uint64
	for i := len(tail) - 1; i >= 8; i-- {
		k2 ^= uint64(tail[i]) << (8 * uint(i-8))
	}
	if len(tail) > 8 {
		k2 *= c2
		k2 = bits.RotateLeft64(k2, 33)
		k2 *= c1
		h2 ^= k2
	}
	for i := min(len(tail), 8) - 1; i >= 0; i-- {
		k1 ^= uint64(tail[i]) << (8 * uint(i))
	}
	if len(tail) > 0 {
		k1 *= c1
		k1 = bits.RotateLeft64(k1, 31)
		k1 *= c2
		h1 ^= k1
	}
	h1 ^= uint64(len(data))
	h2 ^= uint64(len(data))
	h1 += h2
	h2 += h1
	h1 = fmix64(h1)
	h2 = fmix64(h2)
	h1 += h2
	h2 += h1
	return h1, h2
}

// ---- xxHash ----

func xxh32(data []byte, seed uint32) uint32 {
	const p1, p2, p3, p4, p5 = 2654435761, 2246822519, 3266489917, 668265263, 374761393
	round := func(acc, in uint32) uint32 { return bits.RotateLeft32(acc+in*p2, 13) * p1 }
	n := len(data)
	var h uint32
	i := 0
	if n >= 16 {
		v1, v2, v3, v4 := seed+p1+p2, seed+p2, seed, seed-p1
		for ; i+16 <= n; i += 16 {
			v1 = round(v1, binary.LittleEndian.Uint32(data[i:]))
			v2 = round(v2, binary.LittleEndian.Uint32(data[i+4:]))
			v3 = round(v3, binary.LittleEndian.Uint32(data[i+8:]))
			v4 = round(v4, binary.LittleEndian.Uint32(data[i+12:]))
		}
		h = bits.RotateLeft32(v1, 1) + bits.RotateLeft32(v2, 7) + bits.RotateLeft32(v3, 12) + bits.RotateLeft32(v4, 18)
	} else {
		h = seed + p5
	}
	h += uint32(n)
	for ; i+4 <= n; i += 4 {
		h += binary.LittleEndian.Uint32(data[i:]) * p3
		h = bits.RotateLeft32(h, 17) * p4
	}
	for ; i < n; i++ {
		h += uint32(data[i]) * p5
		h = bits.RotateLeft32(h, 11) * p1
	}
	h ^= h >> 15
	h *= p2
	h ^= h >> 13
	h *= p3
	h ^= h >> 16
	return h
}

func xxh64(data []byte, seed uint64) uint64 {
	const p1, p2, p3, p4, p5 uint64 = 11400714785074694791, 14029467366897019727, 1609587929392839161, 9650029242287828579, 2870177450012600261
	round := func(acc, in uint64) uint64 { return bits.RotateLeft64(acc+in*p2, 31) * p1 }
	merge := func(acc, v uint64) uint64 { return (acc^round(0, v))*p1 + p4 }
	n := len(data)
	var h uint64
	i := 0
	if n >= 32 {
		v1, v2, v3, v4 := seed+p1+p2, seed+p2, seed, seed-p1
		for ; i+32 <= n; i += 32 {
			v1 = round(v1, binary.LittleEndian.Uint64(data[i:]))
			v2 = round(v2, binary.LittleEndian.Uint64(data[i+8:]))
			v3 = round(v3, binary.LittleEndian.Uint64(data[i+16:]))
			v4 = round(v4, binary.LittleEndian.Uint64(data[i+24:]))
		}
		h = bits.RotateLeft64(v1, 1) + bits.RotateLeft64(v2, 7) + bits.RotateLeft64(v3, 12) + bits.RotateLeft64(v4, 18)
		h = merge(h, v1)
		h = merge(h, v2)
		h = merge(h, v3)
		h = merge(h, v4)
	} else {
		h = seed + p5
	}
	h += uint64(n)
	for ; i+8 <= n; i += 8 {
		h ^= round(0, binary.LittleEndian.Uint64(data[i:]))
		h = bits.RotateLeft64(h, 27)*p1 + p4
	}
	if i+4 <= n {
		h ^= uint64(binary.LittleEndian.Uint32(data[i:])) * p1
		h = bits.RotateLeft64(h, 23)*p2 + p3
		i += 4
	}
	for ; i < n; i++ {
		h ^= uint64(data[i]) * p5
		h = bits.RotateLeft64(h, 11) * p1
	}
	h ^= h >> 33
	h *= p2
	h ^= h >> 29
	h *= p3
	h ^= h >> 32
	return h
}

// ---- SipHash-2-4 ----

func siphash24(key [16]byte, data []byte) uint64 {
	k0 := binary.LittleEndian.Uint64(key[:8])
	k1 := binary.LittleEndian.Uint64(key[8:])
	v0 := k0 ^ 0x736f6d6570736575
	v1 := k1 ^ 0x646f72616e646f6d
	v2 := k0 ^ 0x6c7967656e657261
	v3 := k1 ^ 0x7465646279746573
	rnd := func() {
		v0 += v1
		v1 = bits.RotateLeft64(v1, 13)
		v1 ^= v0
		v0 = bits.RotateLeft64(v0, 32)
		v2 += v3
		v3 = bits.RotateLeft64(v3, 16)
		v3 ^= v2
		v0 += v3
		v3 = bits.RotateLeft64(v3, 21)
		v3 ^= v0
		v2 += v1
		v1 = bits.RotateLeft64(v1, 17)
		v1 ^= v2
		v2 = bits.RotateLeft64(v2, 32)
	}
	n := len(data)
	i := 0
	for ; i+8 <= n; i += 8 {
		m := binary.LittleEndian.Uint64(data[i:])
		v3 ^= m
		rnd()
		rnd()
		v0 ^= m
	}
	last := uint64(n) << 56
	for j := 0; i+j < n; j++ {
		last |= uint64(data[i+j]) << (8 * uint(j))
	}
	v3 ^= last
	rnd()
	rnd()
	v0 ^= last
	v2 ^= 0xff
	rnd()
	rnd()
	rnd()
	rnd()
	return v0 ^ v1 ^ v2 ^ v3
}

// ---- simple string hashes ----

func be32(v uint32) []byte { return binary.BigEndian.AppendUint32(nil, v) }
func be64(v uint64) []byte { return binary.BigEndian.AppendUint64(nil, v) }

func djb2(data []byte) []byte {
	h := uint32(5381)
	for _, c := range data {
		h = h*33 + uint32(c)
	}
	return be32(h)
}

func djb2a(data []byte) []byte {
	h := uint32(5381)
	for _, c := range data {
		h = (h * 33) ^ uint32(c)
	}
	return be32(h)
}

func sdbm(data []byte) []byte {
	var h uint32
	for _, c := range data {
		h = uint32(c) + (h << 6) + (h << 16) - h
	}
	return be32(h)
}

func jenkinsOAAT(data []byte) []byte {
	var h uint32
	for _, c := range data {
		h += uint32(c)
		h += h << 10
		h ^= h >> 6
	}
	h += h << 3
	h ^= h >> 11
	h += h << 15
	return be32(h)
}

// ---- Windows / MySQL password hashes ----

func ntlm(data []byte) []byte {
	u := utf16.Encode([]rune(string(data)))
	b := make([]byte, 2*len(u))
	for i, c := range u {
		binary.LittleEndian.PutUint16(b[2*i:], c)
	}
	h := md4.New()
	h.Write(b)
	return h.Sum(nil)
}

func desKey7(k []byte) []byte {
	// Expand 7 key bytes into 8 DES key bytes (parity bits are ignored by Go).
	return []byte{
		k[0] & 0xfe,
		k[0]<<7 | k[1]>>1,
		k[1]<<6 | k[2]>>2,
		k[2]<<5 | k[3]>>3,
		k[3]<<4 | k[4]>>4,
		k[4]<<3 | k[5]>>5,
		k[5]<<2 | k[6]>>6,
		k[6] << 1,
	}
}

// lmHash computes the legacy LAN Manager hash (uppercased, 14 bytes max).
func lmHash(data []byte) []byte {
	pw := []byte(strings.ToUpper(string(data)))
	if len(pw) > 14 {
		pw = pw[:14]
	}
	key := make([]byte, 14)
	copy(key, pw)
	out := make([]byte, 0, 16)
	for i := 0; i < 2; i++ {
		c, err := des.NewCipher(desKey7(key[7*i : 7*i+7]))
		if err != nil {
			return nil
		}
		blk := make([]byte, 8)
		c.Encrypt(blk, []byte("KGS!@#$%"))
		out = append(out, blk...)
	}
	return out
}

// mysql41 is the MySQL PASSWORD() hash: SHA1(SHA1(pw)), shown without the
// leading "*" that MySQL prepends.
func mysql41(data []byte) []byte {
	a := sha1.Sum(data)
	b := sha1.Sum(a[:])
	return b[:]
}
