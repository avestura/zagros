package hashx

import "hash"

// Whirlpool (final version, ISO/IEC 10118-3), built from the published
// structure: S-box from three 4-bit mini-boxes, circulant diffusion matrix
// over GF(2^8) with polynomial 0x11D, 10 rounds, Miyaguchi-Preneel.

var wpSbox = func() [256]byte {
	e := [16]byte{0x1, 0xB, 0x9, 0xC, 0xD, 0x6, 0xF, 0x3, 0xE, 0x8, 0x7, 0x4, 0xA, 0x2, 0x5, 0x0}
	r := [16]byte{0x7, 0xC, 0xB, 0xD, 0xE, 0x4, 0x9, 0xF, 0x6, 0x3, 0x8, 0xA, 0x2, 0x5, 0x1, 0x0}
	var einv [16]byte
	for i, v := range e {
		einv[v] = byte(i)
	}
	var s [256]byte
	for u := 0; u < 256; u++ {
		u1 := e[u>>4]
		u2 := einv[u&15]
		rr := r[u1^u2]
		s[u] = e[u1^rr]<<4 | einv[u2^rr]
	}
	return s
}()

func gfMul(a, b byte) byte {
	var p byte
	for i := 0; i < 8; i++ {
		if b&1 != 0 {
			p ^= a
		}
		hi := a & 0x80
		a <<= 1
		if hi != 0 {
			a ^= 0x1D // 0x11D without the overflow bit
		}
		b >>= 1
	}
	return p
}

var wpCirc = [8]byte{1, 1, 4, 1, 8, 5, 2, 9}

// wpRound applies gamma, pi, theta and the key addition to a 64-byte state
// (row-major 8x8).
func wpRound(x, k *[64]byte) [64]byte {
	var g, p, out [64]byte
	for i := range x {
		g[i] = wpSbox[x[i]]
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			p[8*i+j] = g[8*((i-j+8)%8)+j]
		}
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			var v byte
			for k := 0; k < 8; k++ {
				v ^= gfMul(p[8*i+k], wpCirc[(j-k+8)%8])
			}
			out[8*i+j] = v
		}
	}
	for i := range out {
		out[i] ^= k[i]
	}
	return out
}

func whirlpoolSum(data []byte) []byte {
	msg := append([]byte(nil), data...)
	msg = append(msg, 0x80)
	for len(msg)%64 != 32 {
		msg = append(msg, 0)
	}
	var l [32]byte
	n := uint64(len(data)) * 8
	for i := 0; i < 8; i++ {
		l[31-i] = byte(n >> (8 * uint(i)))
	}
	msg = append(msg, l[:]...)

	var h [64]byte
	for off := 0; off < len(msg); off += 64 {
		var m [64]byte
		copy(m[:], msg[off:off+64])
		k := h
		var st [64]byte
		for i := range st {
			st[i] = m[i] ^ k[i]
		}
		for r := 0; r < 10; r++ {
			var c [64]byte
			for j := 0; j < 8; j++ {
				c[j] = wpSbox[8*r+j]
			}
			k = wpRound(&k, &c)
			st = wpRound(&st, &k)
		}
		for i := range h {
			h[i] ^= st[i] ^ m[i]
		}
	}
	return h[:]
}

func newWhirlpool() hash.Hash { return &bufHash{f: whirlpoolSum, size: 64, block: 64} }
