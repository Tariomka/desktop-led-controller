package global

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

type LayerState int32

const (
	All LayerState = iota
	Layer
	Column
	Precise
)

// Globally accesable state
var (
	Messenger *common.Messenger

	ShouldChangeColor bool
	SelectedColor     color.RGBA = common.ColorOff

	SelectedLayerState LayerState
	SelectedLayer      uint8
	SelectedColumn     uint8

	WindowShouldClose bool
)
