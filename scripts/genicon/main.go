// Command genicon writes assets/zagros.ico: a square-cornered Z mark.
//
//	go run ./scripts/genicon [-o assets/zagros.ico]
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

var (
	bg = color.NRGBA{0x0f, 0x62, 0xfe, 0xff} // Boxy blue
	fg = color.NRGBA{0xff, 0xff, 0xff, 0xff}
)

// inZ reports whether the point (u,v) in the unit square lies on the Z.
func inZ(u, v float64) bool {
	const lo, hi, t = 0.22, 0.78, 0.15
	if u < lo || u > hi || v < lo || v > hi {
		return false
	}
	if v < lo+t || v > hi-t {
		return true
	}
	// diagonal from top-right to bottom-left
	frac := (v - lo) / (hi - lo)
	cx := hi - frac*(hi-lo)
	return math.Abs(u-cx) < t*0.95
}

func render(n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	const ss = 4
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			hits := 0
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					u := (float64(x) + (float64(sx)+0.5)/ss) / float64(n)
					v := (float64(y) + (float64(sy)+0.5)/ss) / float64(n)
					if inZ(u, v) {
						hits++
					}
				}
			}
			a := float64(hits) / (ss * ss)
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(float64(bg.R)*(1-a) + float64(fg.R)*a),
				G: uint8(float64(bg.G)*(1-a) + float64(fg.G)*a),
				B: uint8(float64(bg.B)*(1-a) + float64(fg.B)*a),
				A: 0xff,
			})
		}
	}
	return img
}

// bmpEntry encodes img as an ICO DIB image (BITMAPINFOHEADER + XOR + AND mask).
func bmpEntry(img *image.NRGBA) []byte {
	n := img.Bounds().Dx()
	var b bytes.Buffer
	hdr := []any{uint32(40), int32(n), int32(2 * n), uint16(1), uint16(32), uint32(0), uint32(0), int32(0), int32(0), uint32(0), uint32(0)}
	for _, v := range hdr {
		binary.Write(&b, binary.LittleEndian, v)
	}
	for y := n - 1; y >= 0; y-- {
		for x := 0; x < n; x++ {
			c := img.NRGBAAt(x, y)
			b.Write([]byte{c.B, c.G, c.R, c.A})
		}
	}
	b.Write(make([]byte, n*((n+31)/32*4)))
	return b.Bytes()
}

func main() {
	out := flag.String("o", filepath.Join("assets", "zagros.ico"), "output .ico")
	flag.Parse()
	sizes := []int{16, 24, 32, 48, 64, 256}
	var datas [][]byte
	for _, n := range sizes {
		img := render(n)
		if n == 256 {
			var p bytes.Buffer
			if err := png.Encode(&p, img); err != nil {
				log.Fatal(err)
			}
			datas = append(datas, p.Bytes())
		} else {
			datas = append(datas, bmpEntry(img))
		}
	}
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, [3]uint16{0, 1, uint16(len(sizes))})
	off := 6 + 16*len(sizes)
	for i, n := range sizes {
		w := byte(n)
		if n >= 256 {
			w = 0
		}
		b.Write([]byte{w, w, 0, 0})
		binary.Write(&b, binary.LittleEndian, uint16(1))
		binary.Write(&b, binary.LittleEndian, uint16(32))
		binary.Write(&b, binary.LittleEndian, uint32(len(datas[i])))
		binary.Write(&b, binary.LittleEndian, uint32(off))
		off += len(datas[i])
	}
	for _, d := range datas {
		b.Write(d)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, b.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d bytes)", *out, b.Len())
}
