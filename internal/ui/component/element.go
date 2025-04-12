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

func (this *ExitDialog) Update() {
	this.screenWidth = raylib.GetScreenWidth()
	this.screenHeight = raylib.GetScreenHeight()

	if raylib.IsKeyPressed(raylib.KeyEscape) {
		this.show = !this.show
	}
}

func (this *ExitDialog) Render() {
	if !this.show {
		return
	}

	raylib.DrawRectangle(
		0, 0,
		int32(this.screenWidth),
		int32(this.screenHeight),
		raylib.Fade(raylib.Black, 0.7))

	pos := raylib.NewVector2(
		(float32(this.screenWidth)-this.width)/2,
		(float32(this.screenHeight)-this.height)/2)
	result := raygui.MessageBox(
		raylib.NewRectangle(pos.X, pos.Y, this.width, this.height),
		raygui.IconText(raygui.ICON_EXIT, "Close Window"),
		"Do you really want to exit?",
		"Yes;No")

	if raylib.IsKeyPressed(raylib.KeyEnter) {
		result = 1
	}

	if result == 0 || result == 2 {
		this.show = false
	} else if result == 1 {
		global.WindowShouldClose = true
	}
}

type MessageListView struct{}

func (mlv *MessageListView) Update() {}

func (mlv *MessageListView) Render() {}
