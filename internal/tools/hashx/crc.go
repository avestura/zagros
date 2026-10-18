package hashx

import "math/bits"

type crcSpec struct {
	width  uint
	poly   uint64
	init   uint64
	refin  bool
	refout bool
	xorout uint64
}

func reflect(v uint64, width uint) uint64 {
	return bits.Reverse64(v) >> (64 - width)
}

func (s crcSpec) sum(data []byte) uint64 {
	mask := ^uint64(0)
	if s.width < 64 {
		mask = (uint64(1) << s.width) - 1
	}
	top := uint64(1) << (s.width - 1)
	crc := s.init & mask
	for _, b := range data {
		x := uint64(b)
		if s.refin {
			x = uint64(bits.Reverse8(b))
		}
		crc ^= x << (s.width - 8)
		for i := 0; i < 8; i++ {
			if crc&top != 0 {
				crc = (crc << 1) ^ s.poly
			} else {
				crc <<= 1
			}
			crc &= mask
		}
	}
	if s.refout {
		crc = reflect(crc, s.width)
	}
	return (crc ^ s.xorout) & mask
}

func (s crcSpec) bytes(data []byte) []byte {
	v := s.sum(data)
	n := int(s.width+7) / 8
	out := make([]byte, n)
	for i := n - 1; i >= 0; i-- {
		out[i] = byte(v)
		v >>= 8
	}
	return out
}

const all64 = ^uint64(0)

// crcCatalog maps names to CRC models; the check value of each over
// "123456789" is asserted in the tests.
var crcCatalog = map[string]crcSpec{
	"crc8":              {8, 0x07, 0, false, false, 0},
	"crc8_maxim":        {8, 0x31, 0, true, true, 0},
	"crc16_arc":         {16, 0x8005, 0, true, true, 0},
	"crc16_ccitt_false": {16, 0x1021, 0xFFFF, false, false, 0},
	"crc16_xmodem":      {16, 0x1021, 0, false, false, 0},
	"crc16_kermit":      {16, 0x1021, 0, true, true, 0},
	"crc16_modbus":      {16, 0x8005, 0xFFFF, true, true, 0},
	"crc16_x25":         {16, 0x1021, 0xFFFF, true, true, 0xFFFF},
	"crc16_usb":         {16, 0x8005, 0xFFFF, true, true, 0xFFFF},
	"crc16_genibus":     {16, 0x1021, 0xFFFF, false, false, 0xFFFF},
	"crc16_mcrf4xx":     {16, 0x1021, 0xFFFF, true, true, 0},
	"crc16_dnp":         {16, 0x3D65, 0, true, true, 0xFFFF},
	"crc16_buypass":     {16, 0x8005, 0, false, false, 0},
	"crc16_t10_dif":     {16, 0x8BB7, 0, false, false, 0},
	"crc16_cdma2000":    {16, 0xC867, 0xFFFF, false, false, 0},
	"crc16_aug_ccitt":   {16, 0x1021, 0x1D0F, false, false, 0},
	"crc24":             {24, 0x864CFB, 0xB704CE, false, false, 0},
	"crc32_ieee":        {32, 0x04C11DB7, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
	"crc32_castagnoli":  {32, 0x1EDC6F41, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
	"crc32_koopman":     {32, 0x741B8CD7, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
	"crc32_bzip2":       {32, 0x04C11DB7, 0xFFFFFFFF, false, false, 0xFFFFFFFF},
	"crc32_mpeg2":       {32, 0x04C11DB7, 0xFFFFFFFF, false, false, 0},
	"crc32_posix":       {32, 0x04C11DB7, 0, false, false, 0xFFFFFFFF},
	"crc32_jamcrc":      {32, 0x04C11DB7, 0xFFFFFFFF, true, true, 0},
	"crc32_q":           {32, 0x814141AB, 0, false, false, 0},
	"crc32_d":           {32, 0xA833982B, 0xFFFFFFFF, true, true, 0xFFFFFFFF},
	"crc64_iso":         {64, 0x1B, all64, true, true, all64},
	"crc64_ecma":        {64, 0x42F0E1EBA9EA3693, all64, true, true, all64},
	"crc64_ecma182":     {64, 0x42F0E1EBA9EA3693, 0, false, false, 0},
	"crc64_we":          {64, 0x42F0E1EBA9EA3693, all64, false, false, all64},
}

// Variant selectors for the crc16, crc32 and crc64 functions.
var (
	crc16Variants = map[string]string{
		"arc": "crc16_arc", "ccitt_false": "crc16_ccitt_false", "xmodem": "crc16_xmodem", "kermit": "crc16_kermit",
		"modbus": "crc16_modbus", "x25": "crc16_x25", "usb": "crc16_usb", "genibus": "crc16_genibus",
		"mcrf4xx": "crc16_mcrf4xx", "dnp": "crc16_dnp", "buypass": "crc16_buypass", "t10_dif": "crc16_t10_dif",
		"cdma2000": "crc16_cdma2000", "aug_ccitt": "crc16_aug_ccitt",
	}
	crc32Variants = map[string]string{
		"ieee": "crc32_ieee", "castagnoli": "crc32_castagnoli", "crc32c": "crc32_castagnoli", "koopman": "crc32_koopman",
		"bzip2": "crc32_bzip2", "mpeg2": "crc32_mpeg2", "posix": "crc32_posix", "cksum": "crc32_posix",
		"jamcrc": "crc32_jamcrc", "q": "crc32_q", "d": "crc32_d",
	}
	crc64Variants = map[string]string{
		"ecma": "crc64_ecma", "xz": "crc64_ecma", "iso": "crc64_iso", "ecma182": "crc64_ecma182", "we": "crc64_we",
	}
)
