package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Panel struct{ raylib.Rectangle }

func (this *Panel) renderPanel() { raygui.Panel(this.Rectangle, "") }

func (this *Panel) resize(additionalResizes ...func()) {
	if !raylib.IsWindowResized() {
		return
	}

	previousPosX, previousPosY := this.X, this.Y
	panelWidth := float32(raylib.GetScreenWidth()) * style.PanelWidthCoeficient

	if previousPosX > 0 {
		this.Width = panelWidth
		this.X = float32(raylib.GetScreenWidth()) - panelWidth
	}
	if previousPosY > 0 {
		this.Height = float32(raylib.GetScreenHeight()) - previousPosY
	}

	for _, callback := range additionalResizes {
		callback()
	}
}
