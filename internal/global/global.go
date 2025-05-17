package global

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/data"
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
	messenger *data.Messenger = data.NewMessanger()

	ShouldChangeColor bool
	SelectedColor     color.RGBA = common.ColorOff

	SelectedLayerState LayerState
	SelectedLayer      uint8
	SelectedColumn     uint8

	SelectedFrame   uint32
	TotalFrameCount uint32

	WindowShouldClose bool
)

func RegisterMessageReceiver(key string, receiver data.MessageReceiver) {
	messenger.RegisterReceiver(key, receiver)
}

func SendMessage(key string, message any) {
	messenger.Send(key, message)
}
