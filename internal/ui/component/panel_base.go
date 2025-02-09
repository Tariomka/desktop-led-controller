package component

import raylib "github.com/gen2brain/raylib-go/raylib"

type PanelBase struct {
	Pos           raylib.Vector2
	Width, Height float32
}

// Default calculates properties with respect to window size, which only works after raylib init is called.
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func defaultPanelBase() PanelBase {
	panelWidth := float32(raylib.GetScreenWidth()) / 4
	return PanelBase{
		Pos:    raylib.NewVector2(float32(raylib.GetScreenWidth())-panelWidth, 0),
		Width:  panelWidth,
		Height: float32(raylib.GetScreenHeight()),
	}
}

func (pb *PanelBase) resize(additionalResizes ...func()) {
	if raylib.IsWindowResized() {
		previousPosition := pb.Pos
		panelWidth := float32(raylib.GetScreenWidth()) / 4

		if previousPosition.X > 0 {
			pb.Width = panelWidth
			pb.Pos.X = float32(raylib.GetScreenWidth()) - panelWidth
		}
		if previousPosition.Y > 0 {
			pb.Height = float32(raylib.GetScreenHeight()) - previousPosition.Y
		}

		for _, callback := range additionalResizes {
			callback()
		}
	}
}
