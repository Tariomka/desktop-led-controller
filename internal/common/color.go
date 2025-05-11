package common

import (
	"encoding/binary"
	"image/color"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	alpha = 120
	on    = 255
	off   = 0
)

var (
	ColorOff    = raylib.Blank
	ColorGreen  = raylib.NewColor(off, on, off, alpha)
	ColorBlue   = raylib.NewColor(off, off, on, alpha)
	ColorRed    = raylib.NewColor(on, off, off, alpha)
	ColorCyan   = raylib.NewColor(off, on, on, alpha)
	ColorYellow = raylib.NewColor(on, on, off, alpha)
	ColorViolet = raylib.NewColor(on, off, on, alpha)
	ColorWhite  = raylib.NewColor(on, on, on, alpha)
)

func IntToRGBA(value int64) color.RGBA {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(value))
	return color.RGBA{
		R: bytes[3],
		G: bytes[2],
		B: bytes[1],
		A: bytes[0],
	}
}

func IntToRGBAExtended(value int64, alpha uint8) color.RGBA {
	base := IntToRGBA(value)
	base.A = alpha
	return base
}
