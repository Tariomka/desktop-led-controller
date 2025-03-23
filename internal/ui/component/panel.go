package component

import (
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Panel struct{ raylib.Rectangle }

// Default calculates properties with respect to window size, which only works after raylib init is called.
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func defaultPanel() Panel {
	panelWidth := float32(raylib.GetScreenWidth()) / 4
	return Panel{
		Rectangle: raylib.NewRectangle(
			float32(raylib.GetScreenWidth())-panelWidth,
			0,
			panelWidth,
			float32(raylib.GetScreenHeight())),
	}
}

func (p *Panel) renderPanel() { raygui.Panel(p.Rectangle, "") }

func (p *Panel) resize(additionalResizes ...func()) {
	if !raylib.IsWindowResized() {
		return
	}

	previousPosX, previousPosY := p.X, p.Y
	panelWidth := float32(raylib.GetScreenWidth()) / 4

	if previousPosX > 0 {
		p.Width = panelWidth
		p.X = float32(raylib.GetScreenWidth()) - panelWidth
	}
	if previousPosY > 0 {
		p.Height = float32(raylib.GetScreenHeight()) - previousPosY
	}

	for _, callback := range additionalResizes {
		callback()
	}
}
