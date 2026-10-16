package ui

import (
	"fmt"
	"strings"
	"unicode"
)

// Hotkey modifier flags (match the Win32 MOD_* values).
const (
	HotAlt     = 0x1
	HotControl = 0x2
	HotShift   = 0x4
	HotWin     = 0x8
)

var namedKeys = map[string]uint32{
	"space": 0x20, "enter": 0x0D, "return": 0x0D, "tab": 0x09, "esc": 0x1B, "escape": 0x1B,
	"backspace": 0x08, "delete": 0x2E, "del": 0x2E, "insert": 0x2D, "ins": 0x2D,
	"home": 0x24, "end": 0x23, "pageup": 0x21, "pgup": 0x21, "pagedown": 0x22, "pgdn": 0x22,
	"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
	"`": 0xC0, "backtick": 0xC0, "grave": 0xC0, "-": 0xBD, "=": 0xBB, ",": 0xBC, ".": 0xBE,
	"/": 0xBF, ";": 0xBA, "'": 0xDE, "[": 0xDB, "]": 0xDD, "\\": 0xDC,
	"minus": 0xBD, "plus": 0xBB, "equals": 0xBB, "comma": 0xBC, "period": 0xBE, "slash": 0xBF,
}

// ParseHotkey parses strings like "Alt+Space", "Ctrl+Shift+K", "Win+`", "F9".
// It returns the modifier mask and the virtual-key code.
func ParseHotkey(s string) (mods, vk uint32, err error) {
	parts := strings.Split(strings.TrimSpace(s), "+")
	if len(parts) == 0 || strings.TrimSpace(s) == "" {
		return 0, 0, fmt.Errorf("empty hotkey")
	}
	// "Ctrl++" style: trailing empty part means the key is '+'.
	for i, p := range parts {
		p = strings.TrimSpace(p)
		low := strings.ToLower(p)
		last := i == len(parts)-1
		switch low {
		case "alt":
			mods |= HotAlt
			continue
		case "ctrl", "control":
			mods |= HotControl
			continue
		case "shift":
			mods |= HotShift
			continue
		case "win", "super", "meta", "cmd":
			mods |= HotWin
			continue
		}
		if !last {
			return 0, 0, fmt.Errorf("hotkey %q: key %q must be last", s, p)
		}
		if k, ok := namedKeys[low]; ok {
			vk = k
		} else if len(low) > 1 && low[0] == 'f' && isDigits(low[1:]) {
			var n int
			fmt.Sscanf(low[1:], "%d", &n)
			if n < 1 || n > 24 {
				return 0, 0, fmt.Errorf("hotkey %q: bad function key", s)
			}
			vk = uint32(0x70 + n - 1)
		} else if r := []rune(p); len(r) == 1 && r[0] < 128 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0])) {
			vk = uint32(unicode.ToUpper(r[0]))
		} else {
			return 0, 0, fmt.Errorf("hotkey %q: unknown key %q", s, p)
		}
	}
	if vk == 0 {
		return 0, 0, fmt.Errorf("hotkey %q: no key", s)
	}
	if mods == 0 && !(vk >= 0x70 && vk <= 0x87) {
		return 0, 0, fmt.Errorf("hotkey %q: needs a modifier", s)
	}
	return mods, vk, nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
