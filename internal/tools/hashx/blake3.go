package hashx

import (
	"encoding/binary"
	"math/bits"
)

// BLAKE3 implemented from the specification (hash, keyed_hash and derive_key
// modes, extendable output). Whole-input, single-threaded.

const (
	b3ChunkStart = 1 << 0
	b3ChunkEnd   = 1 << 1
	b3Parent     = 1 << 2
	b3Root       = 1 << 3
	b3Keyed      = 1 << 4
	b3DeriveCtx  = 1 << 5
	b3DeriveMat  = 1 << 6
	b3ChunkLen   = 1024
	b3BlockLen   = 64
)

var b3IV = [8]uint32{0x6A09E667, 0xBB67AE85, 0x3C6EF372, 0xA54FF53A, 0x510E527F, 0x9B05688C, 0x1F83D9AB, 0x5BE0CD19}

var b3Perm = [16]int{2, 6, 3, 10, 7, 0, 4, 13, 1, 11, 12, 5, 9, 14, 15, 8}

func b3g(s *[16]uint32, a, b, c, d int, mx, my uint32) {
	s[a] = s[a] + s[b] + mx
	s[d] = bits.RotateLeft32(s[d]^s[a], -16)
	s[c] = s[c] + s[d]
	s[b] = bits.RotateLeft32(s[b]^s[c], -12)
	s[a] = s[a] + s[b] + my
	s[d] = bits.RotateLeft32(s[d]^s[a], -8)
	s[c] = s[c] + s[d]
	s[b] = bits.RotateLeft32(s[b]^s[c], -7)
}

func b3compress(cv [8]uint32, block [16]uint32, counter uint64, blen, flags uint32) [16]uint32 {
	s := [16]uint32{
		cv[0], cv[1], cv[2], cv[3], cv[4], cv[5], cv[6], cv[7],
		b3IV[0], b3IV[1], b3IV[2], b3IV[3],
		uint32(counter), uint32(counter >> 32), blen, flags,
	}
	m := block
	for r := 0; r < 7; r++ {
		b3g(&s, 0, 4, 8, 12, m[0], m[1])
		b3g(&s, 1, 5, 9, 13, m[2], m[3])
		b3g(&s, 2, 6, 10, 14, m[4], m[5])
		b3g(&s, 3, 7, 11, 15, m[6], m[7])
		b3g(&s, 0, 5, 10, 15, m[8], m[9])
		b3g(&s, 1, 6, 11, 12, m[10], m[11])
		b3g(&s, 2, 7, 8, 13, m[12], m[13])
		b3g(&s, 3, 4, 9, 14, m[14], m[15])
		var p [16]uint32
		for i := range p {
			p[i] = m[b3Perm[i]]
		}
		m = p
	}
	for i := 0; i < 8; i++ {
		s[i] ^= s[i+8]
		s[i+8] ^= cv[i]
	}
	return s
}

func b3words(b []byte) [16]uint32 {
	var buf [64]byte
	copy(buf[:], b)
	var w [16]uint32
	for i := range w {
		w[i] = binary.LittleEndian.Uint32(buf[4*i:])
	}
	return w
}

// b3output is the input of the final compression of a node; it can yield a
// chaining value or (with the ROOT flag) any amount of output bytes.
type b3output struct {
	cv      [8]uint32
	block   [16]uint32
	counter uint64
	blen    uint32
	flags   uint32
}

func (o b3output) chainingValue() [8]uint32 {
	s := b3compress(o.cv, o.block, o.counter, o.blen, o.flags)
	var cv [8]uint32
	copy(cv[:], s[:8])
	return cv
}

func (o b3output) root(n int) []byte {
	out := make([]byte, 0, n+64)
	for t := uint64(0); len(out) < n; t++ {
		s := b3compress(o.cv, o.block, t, o.blen, o.flags|b3Root)
		for _, w := range s {
			out = binary.LittleEndian.AppendUint32(out, w)
		}
	}
	return out[:n]
}

func b3chunk(key [8]uint32, data []byte, idx uint64, flags uint32) b3output {
	cv := key
	nblocks := (len(data) + b3BlockLen - 1) / b3BlockLen
	if nblocks == 0 {
		nblocks = 1
	}
	for i := 0; i < nblocks; i++ {
		start := i * b3BlockLen
		end := min(start+b3BlockLen, len(data))
		fl := flags
		if i == 0 {
			fl |= b3ChunkStart
		}
		blk := b3words(data[start:end])
		if i == nblocks-1 {
			return b3output{cv, blk, idx, uint32(end - start), fl | b3ChunkEnd}
		}
		s := b3compress(cv, blk, idx, b3BlockLen, fl)
		copy(cv[:], s[:8])
	}
	panic("unreachable")
}

func b3parent(key [8]uint32, l, r [8]uint32, flags uint32) b3output {
	var blk [16]uint32
	copy(blk[:8], l[:])
	copy(blk[8:], r[:])
	return b3output{key, blk, 0, b3BlockLen, flags | b3Parent}
}

func b3node(key [8]uint32, data []byte, idx uint64, flags uint32) b3output {
	if len(data) <= b3ChunkLen {
		return b3chunk(key, data, idx, flags)
	}
	total := (len(data) + b3ChunkLen - 1) / b3ChunkLen
	left := 1 << (bits.Len(uint(total-1)) - 1)
	l := b3node(key, data[:left*b3ChunkLen], idx, flags)
	r := b3node(key, data[left*b3ChunkLen:], idx+uint64(left), flags)
	return b3parent(key, l.chainingValue(), r.chainingValue(), flags)
}

func b3keyWords(k []byte) [8]uint32 {
	var w [8]uint32
	for i := range w {
		w[i] = binary.LittleEndian.Uint32(k[4*i:])
	}
	return w
}

// blake3Hash returns n output bytes in default hash mode.
func blake3Hash(data []byte, n int) []byte { return b3node(b3IV, data, 0, 0).root(n) }

// blake3Keyed requires a 32-byte key.
func blake3Keyed(key, data []byte, n int) []byte {
	return b3node(b3keyWords(key), data, 0, b3Keyed).root(n)
}

// blake3Derive is the key derivation mode.
func blake3Derive(context string, material []byte, n int) []byte {
	ck := b3node(b3IV, []byte(context), 0, b3DeriveCtx).root(32)
	return b3node(b3keyWords(ck), material, 0, b3DeriveMat).root(n)
}
