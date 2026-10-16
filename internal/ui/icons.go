//go:build windows

package ui

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/windows"
)

// iconFromICO builds an HICON of about the given size from .ico file bytes.
func iconFromICO(data []byte, want int) windows.Handle {
	if len(data) < 6 || binary.LittleEndian.Uint16(data[2:]) != 1 {
		return 0
	}
	n := int(binary.LittleEndian.Uint16(data[4:]))
	bestOff, bestSize, bestW, bestDiff := 0, 0, 0, 1<<30
	for i := 0; i < n && 6+16*(i+1) <= len(data); i++ {
		e := data[6+16*i:]
		w := int(e[0])
		if w == 0 {
			w = 256
		}
		size := int(binary.LittleEndian.Uint32(e[8:]))
		off := int(binary.LittleEndian.Uint32(e[12:]))
		if off < 0 || size <= 0 || off+size > len(data) {
			continue
		}
		diff := w - want
		if diff < 0 {
			diff = -diff * 2 // prefer downscaling
		}
		if diff < bestDiff {
			bestDiff, bestOff, bestSize, bestW = diff, off, size, w
		}
	}
	if bestSize == 0 {
		return 0
	}
	h, _, _ := pCreateIconFromResEx.Call(uintptr(unsafe.Pointer(&data[bestOff])), uintptr(bestSize), 1,
		0x00030000, uintptr(want), uintptr(want), 0)
	_ = bestW
	return windows.Handle(h)
}
