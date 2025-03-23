package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type ExitDialog struct {
	show                      bool
	width, height             float32
	screenWidth, screenHeight int
}

func (ed *ExitDialog) Update() {
	ed.screenWidth = raylib.GetScreenWidth()
	ed.screenHeight = raylib.GetScreenHeight()

	if raylib.IsKeyPressed(raylib.KeyEscape) {
		ed.show = !ed.show
	}
}

func (ed *ExitDialog) Render() {
	if ed.show {
		raylib.DrawRectangle(0, 0, int32(ed.screenWidth), int32(ed.screenHeight), raylib.Fade(raylib.Black, 0.7))

		pos := raylib.NewVector2((float32(ed.screenWidth)-ed.width)/2, (float32(ed.screenHeight)-ed.height)/2)
		result := raygui.MessageBox(
			raylib.NewRectangle(pos.X, pos.Y, ed.width, ed.height),
			raygui.IconText(raygui.ICON_EXIT, "Close Window"),
			"Do you really want to exit?",
			"Yes;No")

		if raylib.IsKeyPressed(raylib.KeyEnter) {
			result = 1
		}

		if result == 0 || result == 2 {
			ed.show = false
		} else if result == 1 {
			global.WindowShouldClose = true
		}
	}
}

type MessageListView struct{}

func (mlv *MessageListView) Update() {}

func (mlv *MessageListView) Render() {}
