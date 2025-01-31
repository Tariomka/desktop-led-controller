package common

import raylib "github.com/gen2brain/raylib-go/raylib"

var (
	ColorOff    = raylib.Blank
	ColorGray   = raylib.Gray
	ColorGreen  = raylib.NewColor(off, on, off, alpha)
	ColorBlue   = raylib.NewColor(off, off, on, alpha)
	ColorRed    = raylib.NewColor(on, off, off, alpha)
	ColorCyan   = raylib.NewColor(off, on, on, alpha)
	ColorYellow = raylib.NewColor(on, on, off, alpha)
	ColorViolet = raylib.NewColor(on, off, on, alpha)
	ColorWhite  = raylib.NewColor(on, on, on, alpha)
)

const (
	alpha = 120
	on    = 255
	off   = 0
)
