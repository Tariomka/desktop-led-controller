package global

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

// Globally accesable state
var (
	ShouldChangeColor bool
	SelectedColor     color.RGBA = common.ColorGray
	WindowShouldClose bool
)
