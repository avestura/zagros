package theme

import (
	"fmt"
	"strings"
)

// UnmarshalText parses "#rrggbb", "rrggbb" or "#rgb".
func (c *Color) UnmarshalText(b []byte) error {
	s := strings.TrimPrefix(strings.TrimSpace(string(b)), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return fmt.Errorf("invalid color %q: want #rrggbb", b)
	}
	var v uint32
	for i := 0; i < 6; i++ {
		d, ok := hexDigit(s[i])
		if !ok {
			return fmt.Errorf("invalid color %q: bad digit %q", b, s[i])
		}
		v = v<<4 | uint32(d)
	}
	*c = Color(v)
	return nil
}

// MarshalText renders the color as "#rrggbb".
func (c Color) MarshalText() ([]byte, error) { return []byte(c.String()), nil }

func (c Color) String() string { return fmt.Sprintf("#%06x", uint32(c)&0xffffff) }

// RGB returns the 8-bit channels.
func (c Color) RGB() (r, g, b uint8) {
	return uint8(c >> 16), uint8(c >> 8), uint8(c)
}

// COLORREF returns the Win32 0x00BBGGRR value used by GDI.
func (c Color) COLORREF() uint32 {
	r, g, b := c.RGB()
	return uint32(b)<<16 | uint32(g)<<8 | uint32(r)
}

func hexDigit(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
