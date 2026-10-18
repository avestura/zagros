package num

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func bv(a tools.Args, name string) (*big.Int, error) {
	return argBig(a, name, int(a.Int("base")))
}

func bitsRange(a tools.Args, name string, lo int64) (int, error) {
	v, err := intIn(name, a.Int(name), lo, maxBits)
	return int(v), err
}

// padOut renders the non-negative pattern n at a fixed width for hex and bin.
func padOut(n *big.Int, bits int, output string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "hex", "16":
		s := n.Text(16)
		w := (bits + 3) / 4
		if len(s) < w {
			s = strings.Repeat("0", w-len(s)) + s
		}
		return "0x" + s, nil
	case "bin", "2":
		s := n.Text(2)
		if len(s) < bits {
			s = strings.Repeat("0", bits-len(s)) + s
		}
		return "0b" + s, nil
	}
	return outStr(n, output)
}

func fin(a tools.Args, n *big.Int) (any, error) { return outStr(n, a.Str("output")) }

// pattern returns the low `bits` bits of n as a non-negative number.
func pattern(n *big.Int, bits int) *big.Int { return new(big.Int).And(n, maskOf(bits)) }

func reverseBits(n *big.Int, bits int) *big.Int {
	out := new(big.Int)
	for i := 0; i < bits; i++ {
		if n.Bit(i) == 1 {
			out.SetBit(out, bits-1-i, 1)
		}
	}
	return out
}

func bitsViewOf(n *big.Int, bits int) string {
	p := pattern(n, bits)
	s := p.Text(2)
	if len(s) < bits {
		s = strings.Repeat("0", bits-len(s)) + s
	}
	var groups []string
	for len(s) > 0 {
		cut := len(s) - 4
		if cut < 0 {
			cut = 0
		}
		groups = append([]string{s[cut:]}, groups...)
		s = s[:cut]
	}
	var pos, bitsLine []string
	top := bits
	for _, g := range groups {
		top -= len(g)
		ps := fmt.Sprint(top + len(g) - 1)
		w := len(g)
		if len(ps) > w {
			w = len(ps)
		}
		pos = append(pos, strings.Repeat(" ", w-len(ps))+ps)
		bitsLine = append(bitsLine, strings.Repeat(" ", w-len(g))+g)
	}
	return strings.Join(pos, " ") + "\n" + strings.Join(bitsLine, " ")
}

func autoWidth(n *big.Int) int {
	if n.Sign() >= 0 {
		w := (n.BitLen() + 7) / 8 * 8
		if w < 8 {
			w = 8
		}
		return w
	}
	w := 8
	for new(big.Int).Neg(n).Cmp(pow2(w-1)) > 0 {
		w *= 2
	}
	return w
}

func init() {
	two := func(name, title, verb, tags string, f func(x, y *big.Int) *big.Int) {
		reg(name, title, "Bitwise "+verb+" of two big integers (any prefix; negatives behave as infinite two's complement). Returns a string in the chosen output format.",
			"str", `num.`+name+`("0xF0", "0x3C", output="hex")`, 40, tags,
			[]P{rs("a", "First integer."), rs("b", "Second integer."), baseP("base"), outP()},
			func(a tools.Args) (any, error) {
				x, err := bv(a, "a")
				if err != nil {
					return nil, err
				}
				y, err := bv(a, "b")
				if err != nil {
					return nil, err
				}
				return fin(a, f(x, y))
			})
	}
	two("bit_and", "Bitwise AND", "AND", "and mask bitwise & binary", func(x, y *big.Int) *big.Int { return new(big.Int).And(x, y) })
	two("bit_or", "Bitwise OR", "OR", "or bitwise | binary union", func(x, y *big.Int) *big.Int { return new(big.Int).Or(x, y) })
	two("bit_xor", "Bitwise XOR", "XOR", "xor bitwise ^ binary exclusive", func(x, y *big.Int) *big.Int { return new(big.Int).Xor(x, y) })

	reg("bit_not", "Bitwise NOT", "Inverts the bits of an integer. With bits=0 returns -value-1 (infinite two's complement); with bits=N inverts within an N-bit field (0 < N <= 65536) and returns the unsigned pattern, zero-padded in hex/bin output.",
		"str", `num.bit_not("0x0F", 8, output="bin")`, 41, "not invert complement ones complement bitwise ~ flip",
		[]P{rs("value", "Integer."), oi("bits", "Field width in bits (0 = unbounded).", 0), baseP("base"), outP()},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 0)
			if err != nil {
				return nil, err
			}
			if bits == 0 {
				return fin(a, new(big.Int).Not(v))
			}
			return padOut(new(big.Int).Xor(pattern(v, bits), maskOf(bits)), bits, a.Str("output"))
		}, accepts(parsesAuto))

	shift := func(name, title string, left bool) {
		dir := map[bool]string{true: "left", false: "right"}[left]
		reg(name, title, "Shifts a big integer "+dir+" by n bits ("+map[bool]string{true: "multiplies by 2^n", false: "arithmetic shift, floors toward negative infinity"}[left]+"). n must be 0.."+fmt.Sprint(maxShift)+".",
			"str", `num.`+name+`("0b1010", 2, output="bin")`, 42, "shift "+dir+" bitshift << >> multiply divide power of two",
			[]P{rs("value", "Integer."), oi("n", "Number of bit positions.", 1), baseP("base"), outP()},
			func(a tools.Args) (any, error) {
				v, err := bv(a, "value")
				if err != nil {
					return nil, err
				}
				n, err := intIn("n", a.Int("n"), 0, maxShift)
				if err != nil {
					return nil, err
				}
				if left {
					return fin(a, new(big.Int).Lsh(v, uint(n)))
				}
				return fin(a, new(big.Int).Rsh(v, uint(n)))
			}, accepts(parsesAuto))
	}
	shift("shift_left", "Shift left", true)
	shift("shift_right", "Shift right", false)

	rot := func(name, title string, left bool) {
		dir := map[bool]string{true: "left", false: "right"}[left]
		reg(name, title, "Rotates the low `bits` bits of value "+dir+" by n positions (wrapping around) and returns the unsigned pattern, zero-padded in hex/bin output. Negative values are taken as two's complement; n is reduced modulo bits.",
			"str", `num.`+name+`("0x81", 1, 8, output="hex")`, 43, "rotate rol ror circular shift bits "+dir,
			[]P{rs("value", "Integer."), oi("n", "Positions to rotate.", 1), oi("bits", "Field width in bits (1..65536).", 32), baseP("base"), outP()},
			func(a tools.Args) (any, error) {
				v, err := bv(a, "value")
				if err != nil {
					return nil, err
				}
				bits, err := bitsRange(a, "bits", 1)
				if err != nil {
					return nil, err
				}
				n := int(((a.Int("n") % int64(bits)) + int64(bits)) % int64(bits))
				if !left {
					n = (bits - n) % bits
				}
				p := pattern(v, bits)
				r := new(big.Int).Lsh(p, uint(n))
				r.Or(r, new(big.Int).Rsh(p, uint(bits-n)))
				return padOut(pattern(r, bits), bits, a.Str("output"))
			}, accepts(parsesAuto))
	}
	rot("rotate_left", "Rotate left", true)
	rot("rotate_right", "Rotate right", false)

	bitop := func(name, title, verb string, f func(v *big.Int, i int) any, ret string) {
		reg(name, title, verb+" Index 0 is the least significant bit; negatives use infinite two's complement. index must be 0.."+fmt.Sprint(maxShift)+".",
			ret, `num.`+name+`("0b1000", 1)`, 44, "bit flag "+strings.ReplaceAll(name, "_", " ")+" mask index position",
			[]P{rs("value", "Integer."), oi("index", "Bit position, 0 = least significant.", 0), baseP("base"), outP()},
			func(a tools.Args) (any, error) {
				v, err := bv(a, "value")
				if err != nil {
					return nil, err
				}
				i, err := intIn("index", a.Int("index"), 0, maxShift)
				if err != nil {
					return nil, err
				}
				r := f(v, int(i))
				if bn, ok := r.(*big.Int); ok {
					return fin(a, bn)
				}
				return r, nil
			}, accepts(parsesAuto))
	}
	bitop("set_bit", "Set bit", "Sets one bit to 1.", func(v *big.Int, i int) any { return new(big.Int).SetBit(v, i, 1) }, "str")
	bitop("clear_bit", "Clear bit", "Sets one bit to 0.", func(v *big.Int, i int) any { return new(big.Int).SetBit(v, i, 0) }, "str")
	bitop("toggle_bit", "Toggle bit", "Flips one bit.", func(v *big.Int, i int) any {
		return new(big.Int).SetBit(v, i, v.Bit(i)^1)
	}, "str")
	bitop("test_bit", "Test bit", "Returns true if the bit is 1.", func(v *big.Int, i int) any { return v.Bit(i) == 1 }, "bool")

	reg("bit_length", "Bit length", "Returns the number of bits needed to represent the absolute value (0 for 0, 8 for 255, 9 for 256).",
		"int", `num.bit_length("255")`, 45, "bits length size how many bits bitlen width",
		[]P{rs("value", "Integer."), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			return int64(v.BitLen()), nil
		}, accepts(parsesAuto))

	reg("popcount", "Population count", "Counts the 1 bits (Hamming weight). With bits=0 a negative number counts the bits of its absolute value; with bits=N the value is first reduced to an N-bit two's complement pattern.",
		"int", `num.popcount("0xFF")`, 46, "popcount hamming weight count ones set bits bitcount",
		[]P{rs("value", "Integer."), oi("bits", "Field width (0 = none).", 0), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 0)
			if err != nil {
				return nil, err
			}
			if bits > 0 {
				v = pattern(v, bits)
			}
			return int64(popcountAbs(v)), nil
		}, accepts(parsesAuto))

	reg("parity", "Bit parity", "Returns 1 if the number of set bits (of the absolute value) is odd, else 0.",
		"int", `num.parity("7")`, 47, "parity odd even bits xor parity bit",
		[]P{rs("value", "Integer."), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			return int64(popcountAbs(v) & 1), nil
		}, accepts(parsesAuto))

	reg("leading_zeros", "Leading zeros", "Counts the zero bits above the highest 1 inside a `bits`-wide field. A value that does not fit the field is an error; negatives use two's complement (so 0 leading zeros).",
		"int", `num.leading_zeros("1", 32)`, 48, "leading zeros clz count bits msb",
		[]P{rs("value", "Integer."), oi("bits", "Field width 1..65536.", 32), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 1)
			if err != nil {
				return nil, err
			}
			if v.Sign() >= 0 && v.BitLen() > bits {
				return nil, fmt.Errorf("value needs %d bits, more than the %d-bit field", v.BitLen(), bits)
			}
			return int64(bits - pattern(v, bits).BitLen()), nil
		}, accepts(parsesAuto))

	reg("trailing_zeros", "Trailing zeros", "Counts the zero bits below the lowest 1 (ctz). Returns 0 for the value 0; the sign is ignored.",
		"int", `num.trailing_zeros("8")`, 49, "trailing zeros ctz count bits lsb",
		[]P{rs("value", "Integer."), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			if v.Sign() == 0 {
				return int64(0), nil
			}
			return int64(new(big.Int).Abs(v).TrailingZeroBits()), nil
		}, accepts(parsesAuto))

	reg("reverse_bits", "Reverse bits", "Reverses the order of the low `bits` bits (bit 0 becomes bit bits-1). Negatives use two's complement; hex/bin output is zero-padded to the width.",
		"str", `num.reverse_bits("0b0001", 8, output="bin")`, 50, "reverse bits mirror flip order bit reversal",
		[]P{rs("value", "Integer."), oi("bits", "Field width 1..65536.", 32), baseP("base"), outP()},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 1)
			if err != nil {
				return nil, err
			}
			return padOut(reverseBits(pattern(v, bits), bits), bits, a.Str("output"))
		}, accepts(parsesAuto))

	reg("swap_endian", "Swap endianness", "Reverses the byte order of the low `bytes` bytes (big-endian <-> little-endian), e.g. 0x12345678 becomes 0x78563412 for 4 bytes. Negatives use two's complement; hex output is zero-padded.",
		"str", `num.swap_endian("0x12345678", 4, output="hex")`, 51, "endian byte swap bswap htonl ntohl little big byte order",
		[]P{rs("value", "Integer."), oi("bytes", "Width in bytes (1..8192).", 4), baseP("base"), outP()},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			nb, err := intIn("bytes", a.Int("bytes"), 1, maxBits/8)
			if err != nil {
				return nil, err
			}
			p := pattern(v, int(nb)*8)
			if v.Sign() >= 0 && v.BitLen() > int(nb)*8 {
				return nil, fmt.Errorf("value needs %d bytes, more than %d", (v.BitLen()+7)/8, nb)
			}
			buf := p.FillBytes(make([]byte, nb))
			for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
				buf[i], buf[j] = buf[j], buf[i]
			}
			return padOut(new(big.Int).SetBytes(buf), int(nb)*8, a.Str("output"))
		}, accepts(parsesAuto))

	reg("bits_view", "Bits view", "Draws an integer as nibble-grouped binary with the bit index of each group's most significant bit above it. bits=0 picks 8, 16, 32 ... automatically (negatives use two's complement).",
		"str", `num.bits_view("0xA5")`, 1, "bits binary view nibble grouped positions bit index layout register",
		[]P{rs("value", "Integer."), oi("bits", "Field width 1..65536 (0 = automatic).", 0), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 0)
			if err != nil {
				return nil, err
			}
			if bits == 0 {
				bits = autoWidth(v)
				if bits > maxBits {
					return nil, errors.New("value is too wide to draw")
				}
			}
			return bitsViewOf(v, bits), nil
		}, kw("bits", "binary", "nibbles"), accepts(parsesAuto))

	reg("to_twos", "To two's complement", "Converts a signed integer to its two's complement bit pattern in a `bits`-wide field (e.g. -1 in 8 bits is 0xFF). The value must satisfy -2^(bits-1) <= value < 2^bits. Hex/bin output is zero-padded to the width.",
		"str", `num.to_twos("-5", 8, output="bin")`, 52, "twos complement signed negative representation binary two's complement encode",
		[]P{rs("value", "Signed integer."), oi("bits", "Field width 1..65536.", 32), baseP("base"), os("output", "Output format: dec, hex, bin or oct.", "bin", outChoices...)},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 1)
			if err != nil {
				return nil, err
			}
			lo := new(big.Int).Neg(pow2(bits - 1))
			if v.Cmp(lo) < 0 || v.Cmp(pow2(bits)) >= 0 {
				return nil, fmt.Errorf("%s does not fit in %d bits", v, bits)
			}
			return padOut(pattern(v, bits), bits, a.Str("output"))
		}, accepts(parsesAuto))

	reg("from_twos", "From two's complement", "Interprets a bit pattern as a `bits`-wide two's complement number and returns the signed decimal value (0xFF in 8 bits is -1). The pattern must be non-negative and fit the width.",
		"str", `num.from_twos("0xFB", 8)`, 53, "twos complement signed decode negative interpret int8 int16 int32 int64",
		[]P{rs("value", "Bit pattern, e.g. 0xFB or 0b11111011."), oi("bits", "Field width 1..65536.", 32), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			bits, err := bitsRange(a, "bits", 1)
			if err != nil {
				return nil, err
			}
			if v.Sign() < 0 || v.BitLen() > bits {
				return nil, fmt.Errorf("pattern does not fit in %d bits", bits)
			}
			return signedWrap(v, bits).String(), nil
		}, accepts(parsesAuto))

	reg("sign_extend", "Sign extend", "Sign-extends a from_bits-wide two's complement pattern to to_bits and returns the new unsigned pattern (0x80 from 8 to 16 bits is 0xFF80). to_bits must be >= from_bits.",
		"str", `num.sign_extend("0x80", 8, 16, output="hex")`, 54, "sign extend sext widen bits two's complement",
		[]P{rs("value", "Bit pattern."), oi("from_bits", "Source width 1..65536.", 8), oi("to_bits", "Target width 1..65536.", 32), baseP("base"), os("output", "Output format: dec, hex, bin or oct.", "hex", outChoices...)},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			fb, err := bitsRange(a, "from_bits", 1)
			if err != nil {
				return nil, err
			}
			tb, err := bitsRange(a, "to_bits", 1)
			if err != nil {
				return nil, err
			}
			if tb < fb {
				return nil, errors.New("to_bits must be >= from_bits")
			}
			if v.Sign() < 0 || v.BitLen() > fb {
				return nil, fmt.Errorf("pattern does not fit in %d bits", fb)
			}
			return padOut(pattern(signedWrap(v, fb), tb), tb, a.Str("output"))
		}, accepts(parsesAuto))

	reg("mask", "Bit mask", "Builds a mask of `bits` consecutive 1 bits, optionally shifted left (mask(8) = 0xFF, mask(4, 4) = 0xF0). bits 0..65536, shift 0..1048576.",
		"str", `num.mask(8, 4, output="hex")`, 55, "mask bitmask ones field width generate",
		[]P{ri("bits", "Number of 1 bits."), oi("shift", "Left shift applied to the mask.", 0), os("output", "Output format: dec, hex, bin or oct.", "hex", outChoices...)},
		func(a tools.Args) (any, error) {
			bits, err := bitsRange(a, "bits", 0)
			if err != nil {
				return nil, err
			}
			sh, err := intIn("shift", a.Int("shift"), 0, maxShift)
			if err != nil {
				return nil, err
			}
			m := new(big.Int).Lsh(maskOf(bits), uint(sh))
			return outStr(m, a.Str("output"))
		})

	reg("gray_encode", "Gray code encode", "Converts a non-negative integer to its reflected binary Gray code (value XOR value>>1), so 2 becomes 3 and 7 becomes 4.",
		"str", `num.gray_encode("7")`, 56, "gray code reflected binary encode",
		[]P{rs("value", "Non-negative integer."), baseP("base"), outP()},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			if v.Sign() < 0 {
				return nil, errors.New("value must not be negative")
			}
			return fin(a, new(big.Int).Xor(v, new(big.Int).Rsh(v, 1)))
		}, accepts(parsesAuto))

	reg("gray_decode", "Gray code decode", "Converts a Gray code back to the ordinary binary integer (inverse of gray_encode), so 4 becomes 7.",
		"str", `num.gray_decode("4")`, 57, "gray code reflected binary decode",
		[]P{rs("value", "Non-negative Gray code."), baseP("base"), outP()},
		func(a tools.Args) (any, error) {
			v, err := bv(a, "value")
			if err != nil {
				return nil, err
			}
			if v.Sign() < 0 {
				return nil, errors.New("value must not be negative")
			}
			r := new(big.Int).Set(v)
			for sh := uint(1); sh < uint(v.BitLen()); sh <<= 1 {
				r.Xor(r, new(big.Int).Rsh(r, sh))
			}
			return fin(a, r)
		}, accepts(parsesAuto))
}
