package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Element interface {
	Renderer
}

func NewElement[T Element]() Element {
	var placeholder T
	switch any(placeholder).(type) {
	case *ExitDialog:
		raylib.SetExitKey(0)
		return &ExitDialog{}
	default:
		return nil
	}
}

type ExitDialog struct{ show bool }

func (ed *ExitDialog) Update() {
	if raylib.IsKeyPressed(raylib.KeyEscape) {
		ed.show = !ed.show
	}
}

func (ed *ExitDialog) Render() {
	if ed.show {
		raylib.DrawRectangle(
			0,
			0,
			int32(raylib.GetScreenWidth()),
			int32(raylib.GetScreenHeight()),
			raylib.Fade(raylib.Black, 0.7))

		result := raygui.MessageBox(
			raylib.Rectangle{
				X:      float32(raylib.GetScreenWidth())/2 - 125,
				Y:      float32(raylib.GetScreenHeight())/2 - 50,
				Width:  250,
				Height: 100,
			},
			raygui.IconText(raygui.ICON_EXIT, "Close Window"),
			"Do you really want to exit?",
			"Yes;No")

		if raylib.IsKeyPressed(raylib.KeyEnter) {
			result = 1
		}

		if (result == 0) || (result == 2) {
			ed.show = false
		} else if result == 1 {
			global.WindowShouldClose = true
		}
	}
}
