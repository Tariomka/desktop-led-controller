package global

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

// Globally accesable state
var (
	Messenger         *common.Messenger
	ShouldChangeColor bool
	SelectedColor     color.RGBA = common.ColorOff
	WindowShouldClose bool
)
