package ext

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// binary is a pure helper module (no permission): struct-style pack/unpack,
// hex, integers and byte-slice helpers. It is not called "bytes" because that
// is a Starlark builtin.
func (e *Extension) binaryModule() *starlarkstruct.Module {
	return module("binary", starlark.StringDict{
		"pack": starlark.NewBuiltin("binary.pack", biPack),
		"unpack": starlark.NewBuiltin("binary.unpack", func(th *starlark.Thread, b *starlark.Builtin, a starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
			return biUnpack("binary.unpack", a, kw, true)
		}),
		"unpack_from": starlark.NewBuiltin("binary.unpack_from", func(th *starlark.Thread, b *starlark.Builtin, a starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
			return biUnpack("binary.unpack_from", a, kw, false)
		}),
		"calcsize": starlark.NewBuiltin("binary.calcsize", biCalcsize),
		"hex":      starlark.NewBuiltin("binary.hex", biHex),
		"unhex":    starlark.NewBuiltin("binary.unhex", biUnhex),
		"to_int":   starlark.NewBuiltin("binary.to_int", biToInt),
		"from_int": starlark.NewBuiltin("binary.from_int", biFromInt),
		"xor":      starlark.NewBuiltin("binary.xor", biXor),
		"chunks":   starlark.NewBuiltin("binary.chunks", biChunks),
		"reverse":  starlark.NewBuiltin("binary.reverse", biReverse),
	})
}

type byteOrd interface {
	binary.ByteOrder
	binary.AppendByteOrder
}

type fmtItem struct {
	code byte
	n    int
}

func parseFmt(f string) (byteOrd, []fmtItem, error) {
	var order byteOrd = binary.LittleEndian
	f = strings.TrimSpace(f)
	if f != "" {
		switch f[0] {
		case '<', '=', '@':
			f = f[1:]
		case '>', '!':
			order = binary.BigEndian
			f = f[1:]
		}
	}
	var items []fmtItem
	for i := 0; i < len(f); {
		c := f[i]
		if c == ' ' {
			i++
			continue
		}
		n, hasN := 0, false
		for i < len(f) && f[i] >= '0' && f[i] <= '9' {
			n = n*10 + int(f[i]-'0')
			hasN = true
			i++
			if n > 1<<20 {
				return nil, nil, errors.New("repeat count too large")
			}
		}
		if i >= len(f) {
			return nil, nil, errors.New("format ends with a count")
		}
		c = f[i]
		i++
		if !hasN {
			n = 1
		}
		if strings.IndexByte("xcbB?hHiIlLqQfds", c) < 0 {
			return nil, nil, fmt.Errorf("unknown format code %q (use x c b B ? h H i I l L q Q f d s)", string(c))
		}
		items = append(items, fmtItem{c, n})
	}
	return order, items, nil
}

func (it fmtItem) width() int {
	switch it.code {
	case 'x', 'c', 'b', 'B', '?':
		return it.n
	case 's':
		return it.n
	case 'h', 'H':
		return 2 * it.n
	case 'i', 'I', 'l', 'L', 'f':
		return 4 * it.n
	}
	return 8 * it.n // q Q d
}

func fmtSize(items []fmtItem) int {
	t := 0
	for _, it := range items {
		t += it.width()
	}
	return t
}

func biCalcsize(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var f string
	if err := starlark.UnpackArgs("binary.calcsize", args, kw, "fmt", &f); err != nil {
		return nil, err
	}
	_, items, err := parseFmt(f)
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt(fmtSize(items)), nil
}

func intArg(v starlark.Value, lo, hi *big.Int) (*big.Int, error) {
	i, ok := v.(starlark.Int)
	if !ok {
		return nil, fmt.Errorf("want int, got %s", v.Type())
	}
	n := i.BigInt()
	if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
		return nil, fmt.Errorf("%s out of range for the format code (%s..%s)", n, lo, hi)
	}
	return n, nil
}

func biPack(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if len(kw) > 0 || len(args) == 0 {
		return nil, errors.New("binary.pack: want (fmt, *values)")
	}
	f, ok := starlark.AsString(args[0])
	if !ok {
		return nil, errors.New("binary.pack: fmt must be a string")
	}
	order, items, err := parseFmt(f)
	if err != nil {
		return nil, fmt.Errorf("binary.pack: %w", err)
	}
	vals := args[1:]
	out := make([]byte, 0, fmtSize(items))
	next := func() (starlark.Value, error) {
		if len(vals) == 0 {
			return nil, errors.New("binary.pack: not enough values for the format")
		}
		v := vals[0]
		vals = vals[1:]
		return v, nil
	}
	for _, it := range items {
		switch it.code {
		case 'x':
			out = append(out, make([]byte, it.n)...)
		case 's':
			v, err := next()
			if err != nil {
				return nil, err
			}
			bs, err := asBytes(v)
			if err != nil {
				return nil, fmt.Errorf("binary.pack: s wants bytes or str: %w", err)
			}
			pad := make([]byte, it.n)
			copy(pad, bs)
			out = append(out, pad...)
		default:
			for k := 0; k < it.n; k++ {
				v, err := next()
				if err != nil {
					return nil, err
				}
				if out, err = packOne(out, order, it.code, v); err != nil {
					return nil, fmt.Errorf("binary.pack: %w", err)
				}
			}
		}
	}
	if len(vals) > 0 {
		return nil, errors.New("binary.pack: too many values for the format")
	}
	return starlark.Bytes(out), nil
}

var (
	bigMinI8, bigMaxI8   = big.NewInt(math.MinInt8), big.NewInt(math.MaxInt8)
	bigMaxU8             = big.NewInt(math.MaxUint8)
	bigMinI16, bigMaxI16 = big.NewInt(math.MinInt16), big.NewInt(math.MaxInt16)
	bigMaxU16            = big.NewInt(math.MaxUint16)
	bigMinI32, bigMaxI32 = big.NewInt(math.MinInt32), big.NewInt(math.MaxInt32)
	bigMaxU32            = big.NewInt(math.MaxUint32)
	bigMinI64, bigMaxI64 = big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)
	bigMaxU64            = new(big.Int).SetUint64(math.MaxUint64)
	bigZero              = big.NewInt(0)
)

func packOne(out []byte, o byteOrd, code byte, v starlark.Value) ([]byte, error) {
	switch code {
	case 'c':
		bs, err := asBytes(v)
		if err != nil || len(bs) != 1 {
			return nil, errors.New("c wants one byte")
		}
		return append(out, bs[0]), nil
	case '?':
		return append(out, boolByte(v.Truth() == starlark.True)), nil
	case 'f', 'd':
		var x float64
		switch n := v.(type) {
		case starlark.Float:
			x = float64(n)
		case starlark.Int:
			x = float64(n.Float())
		default:
			return nil, fmt.Errorf("%c wants a number, got %s", code, v.Type())
		}
		if code == 'f' {
			return o.AppendUint32(out, math.Float32bits(float32(x))), nil
		}
		return o.AppendUint64(out, math.Float64bits(x)), nil
	}
	var lo, hi *big.Int
	var size int
	switch code {
	case 'b':
		lo, hi, size = bigMinI8, bigMaxI8, 1
	case 'B':
		lo, hi, size = bigZero, bigMaxU8, 1
	case 'h':
		lo, hi, size = bigMinI16, bigMaxI16, 2
	case 'H':
		lo, hi, size = bigZero, bigMaxU16, 2
	case 'i', 'l':
		lo, hi, size = bigMinI32, bigMaxI32, 4
	case 'I', 'L':
		lo, hi, size = bigZero, bigMaxU32, 4
	case 'q':
		lo, hi, size = bigMinI64, bigMaxI64, 8
	default: // Q
		lo, hi, size = bigZero, bigMaxU64, 8
	}
	n, err := intArg(v, lo, hi)
	if err != nil {
		return nil, err
	}
	var u uint64
	if n.Sign() < 0 {
		u = uint64(n.Int64())
	} else {
		u = n.Uint64()
	}
	switch size {
	case 1:
		return append(out, byte(u)), nil
	case 2:
		return o.AppendUint16(out, uint16(u)), nil
	case 4:
		return o.AppendUint32(out, uint32(u)), nil
	}
	return o.AppendUint64(out, u), nil
}

func boolByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func biUnpack(name string, args starlark.Tuple, kw []starlark.Tuple, exact bool) (starlark.Value, error) {
	var f string
	var data starlark.Value
	offset := 0
	var err error
	if exact {
		err = starlark.UnpackArgs(name, args, kw, "fmt", &f, "data", &data)
	} else {
		err = starlark.UnpackArgs(name, args, kw, "fmt", &f, "data", &data, "offset?", &offset)
	}
	if err != nil {
		return nil, err
	}
	bs, err := asBytes(data)
	if err != nil {
		return nil, err
	}
	order, items, err := parseFmt(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if offset < 0 || offset > len(bs) {
		return nil, fmt.Errorf("%s: offset out of range", name)
	}
	bs = bs[offset:]
	size := fmtSize(items)
	if exact && len(bs) != size || len(bs) < size {
		return nil, fmt.Errorf("%s: format needs %d bytes, got %d", name, size, len(bs))
	}
	var out []starlark.Value
	pos := 0
	for _, it := range items {
		switch it.code {
		case 'x':
			pos += it.n
		case 's':
			out = append(out, starlark.Bytes(bs[pos:pos+it.n]))
			pos += it.n
		default:
			for k := 0; k < it.n; k++ {
				v, w := unpackOne(bs[pos:], order, it.code)
				out = append(out, v)
				pos += w
			}
		}
	}
	return starlark.Tuple(out), nil
}

func unpackOne(b []byte, o byteOrd, code byte) (starlark.Value, int) {
	switch code {
	case 'c':
		return starlark.Bytes(b[:1]), 1
	case '?':
		return starlark.Bool(b[0] != 0), 1
	case 'b':
		return starlark.MakeInt(int(int8(b[0]))), 1
	case 'B':
		return starlark.MakeInt(int(b[0])), 1
	case 'h':
		return starlark.MakeInt(int(int16(o.Uint16(b)))), 2
	case 'H':
		return starlark.MakeInt(int(o.Uint16(b))), 2
	case 'i', 'l':
		return starlark.MakeInt64(int64(int32(o.Uint32(b)))), 4
	case 'I', 'L':
		return starlark.MakeUint64(uint64(o.Uint32(b))), 4
	case 'q':
		return starlark.MakeInt64(int64(o.Uint64(b))), 8
	case 'Q':
		return starlark.MakeUint64(o.Uint64(b)), 8
	case 'f':
		return starlark.Float(math.Float32frombits(o.Uint32(b))), 4
	}
	return starlark.Float(math.Float64frombits(o.Uint64(b))), 8 // d
}

func biHex(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var data starlark.Value
	sep := ""
	if err := starlark.UnpackArgs("binary.hex", args, kw, "data", &data, "sep?", &sep); err != nil {
		return nil, err
	}
	bs, err := asBytes(data)
	if err != nil {
		return nil, err
	}
	h := hex.EncodeToString(bs)
	if sep == "" || len(bs) < 2 {
		return starlark.String(h), nil
	}
	parts := make([]string, len(bs))
	for i := range bs {
		parts[i] = h[2*i : 2*i+2]
	}
	return starlark.String(strings.Join(parts, sep)), nil
}

func biUnhex(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs("binary.unhex", args, kw, "text", &s); err != nil {
		return nil, err
	}
	s = strings.TrimPrefix(strings.TrimSpace(s), "0x")
	s = strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", ":", "", "-", "").Replace(s)
	bs, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("binary.unhex: %w", err)
	}
	return starlark.Bytes(bs), nil
}

func byteOrder(s string) (bool, error) {
	switch s {
	case "big":
		return true, nil
	case "little":
		return false, nil
	}
	return false, fmt.Errorf("byteorder must be \"big\" or \"little\", got %q", s)
}

func biToInt(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var data starlark.Value
	order := "big"
	signed := false
	if err := starlark.UnpackArgs("binary.to_int", args, kw, "data", &data, "byteorder?", &order, "signed?", &signed); err != nil {
		return nil, err
	}
	bs, err := asBytes(data)
	if err != nil {
		return nil, err
	}
	isBig, err := byteOrder(order)
	if err != nil {
		return nil, err
	}
	if len(bs) > 4096 {
		return nil, errors.New("binary.to_int: at most 4096 bytes")
	}
	be := append([]byte(nil), bs...)
	if !isBig {
		for i, j := 0, len(be)-1; i < j; i, j = i+1, j-1 {
			be[i], be[j] = be[j], be[i]
		}
	}
	n := new(big.Int).SetBytes(be)
	if signed && len(be) > 0 && be[0]&0x80 != 0 {
		n.Sub(n, new(big.Int).Lsh(big.NewInt(1), uint(8*len(be))))
	}
	return starlark.MakeBigInt(n), nil
}

func biFromInt(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var v starlark.Int
	var size int
	order := "big"
	signed := false
	if err := starlark.UnpackArgs("binary.from_int", args, kw, "value", &v, "size", &size, "byteorder?", &order, "signed?", &signed); err != nil {
		return nil, err
	}
	isBig, err := byteOrder(order)
	if err != nil {
		return nil, err
	}
	if size < 0 || size > 4096 {
		return nil, errors.New("binary.from_int: size must be 0..4096")
	}
	n := v.BigInt()
	lim := new(big.Int).Lsh(big.NewInt(1), uint(8*size))
	if signed {
		half := new(big.Int).Rsh(lim, 1)
		if n.Cmp(new(big.Int).Neg(half)) < 0 || n.Cmp(half) >= 0 {
			return nil, fmt.Errorf("binary.from_int: %s does not fit in %d signed bytes", n, size)
		}
		if n.Sign() < 0 {
			n = new(big.Int).Add(n, lim)
		}
	} else if n.Sign() < 0 || n.Cmp(lim) >= 0 {
		return nil, fmt.Errorf("binary.from_int: %s does not fit in %d unsigned bytes", n, size)
	}
	out := make([]byte, size)
	n.FillBytes(out)
	if !isBig {
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return starlark.Bytes(out), nil
}

func biXor(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var d, k starlark.Value
	if err := starlark.UnpackArgs("binary.xor", args, kw, "data", &d, "key", &k); err != nil {
		return nil, err
	}
	data, err := asBytes(d)
	if err != nil {
		return nil, err
	}
	key, err := asBytes(k)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, errors.New("binary.xor: key is empty")
	}
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ key[i%len(key)]
	}
	return starlark.Bytes(out), nil
}

func biChunks(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var d starlark.Value
	var size int
	if err := starlark.UnpackArgs("binary.chunks", args, kw, "data", &d, "size", &size); err != nil {
		return nil, err
	}
	data, err := asBytes(d)
	if err != nil {
		return nil, err
	}
	if size <= 0 {
		return nil, errors.New("binary.chunks: size must be positive")
	}
	var out []starlark.Value
	for i := 0; i < len(data); i += size {
		out = append(out, starlark.Bytes(data[i:min(i+size, len(data))]))
	}
	return starlark.NewList(out), nil
}

func biReverse(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var d starlark.Value
	if err := starlark.UnpackArgs("binary.reverse", args, kw, "data", &d); err != nil {
		return nil, err
	}
	data, err := asBytes(d)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	for i := range data {
		out[len(data)-1-i] = data[i]
	}
	return starlark.Bytes(out), nil
}

// attempt(fn, *args, **kwargs) calls fn and returns (value, None) or
// (None, message) instead of raising. Cancellation (timeout, step limit) is
// never caught.
func biAttempt(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return nil, errors.New("attempt: want (fn, *args, **kwargs)")
	}
	fn, ok := args[0].(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("attempt: first argument must be callable, got %s", args[0].Type())
	}
	v, err := starlark.Call(th, fn, args[1:], kw)
	if err == nil {
		return starlark.Tuple{v, starlark.None}, nil
	}
	msg := err.Error()
	var ee *starlark.EvalError
	if errors.As(err, &ee) {
		msg = ee.Msg
	}
	if callCtx(th).Err() != nil || strings.Contains(msg, "Starlark computation cancelled") {
		return nil, err
	}
	return starlark.Tuple{starlark.None, starlark.String(msg)}, nil
}
