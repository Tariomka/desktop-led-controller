package component

import raylib "github.com/gen2brain/raylib-go/raylib"

type PanelBase struct{ raylib.Rectangle }

// Default calculates properties with respect to window size, which only works after raylib init is called.
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func defaultPanelBase() PanelBase {
	panelWidth := float32(raylib.GetScreenWidth()) / 4
	return PanelBase{
		Rectangle: raylib.NewRectangle(
			float32(raylib.GetScreenWidth())-panelWidth,
			0,
			panelWidth,
			float32(raylib.GetScreenHeight())),
	}
}

func (pb *PanelBase) resize(additionalResizes ...func()) {
	if raylib.IsWindowResized() {
		previousPosX, previousPosY := pb.X, pb.Y
		panelWidth := float32(raylib.GetScreenWidth()) / 4

		if previousPosX > 0 {
			pb.Width = panelWidth
			pb.X = float32(raylib.GetScreenWidth()) - panelWidth
		}
		if previousPosY > 0 {
			pb.Height = float32(raylib.GetScreenHeight()) - previousPosY
		}

		for _, callback := range additionalResizes {
			callback()
		}
	}
}
