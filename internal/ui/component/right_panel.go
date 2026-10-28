package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	TabShow = iota
	TabProps
	TabTools
	TabCube
)

// RightPanel holds the Show, Props, Tools and Cube tabs.
type RightPanel struct {
	Tab   int
	Show  *ShowTab
	Props *PropsTab
	Tools *ToolsTab
	Cube  *CubeTab

	scroll  float32
	content float32 // height drawn last frame, for scrolling
}

func NewRightPanel(actions *Actions, assets *AssetsPanel) *RightPanel {
	return &RightPanel{
		Show:  &ShowTab{Actions: actions, Assets: assets},
		Props: &PropsTab{Actions: actions},
		Tools: &ToolsTab{Actions: actions},
		Cube:  &CubeTab{},
	}
}

func (this *RightPanel) Draw(bounds raylib.Rectangle) {
	raylib.DrawRectangleRec(bounds, style.Panel)
	raylib.DrawRectangleLinesEx(bounds, 1, style.BorderColor)

	tabBounds := raylib.NewRectangle(bounds.X+20, bounds.Y+Padding, bounds.Width-20-Padding, ControlHeight())
	if tab := tabs(tabBounds, "Show;Props;Tools;Cube", this.Tab); tab != this.Tab {
		this.Tab, this.scroll = tab, 0
		StopEditing()
	}
	top := bounds.Y + Padding + ControlHeight() + Padding
	raylib.DrawLine(int32(bounds.X), int32(top), int32(bounds.X+bounds.Width), int32(top), style.BorderColor)
	content := raylib.NewRectangle(bounds.X, top, bounds.Width, bounds.Y+bounds.Height-top)

	if mouseIn(content) && !EditingText() {
		this.scroll -= raylib.GetMouseWheelMove() * 40
	}
	this.scroll = clampf(this.scroll, 0, max(0, this.content-content.Height))
	// Scrolled content is drawn in a scissor; mouse interaction uses the shifted bounds directly.
	raylib.BeginScissorMode(int32(content.X), int32(content.Y), int32(content.Width), int32(content.Height))
	shifted := content
	shifted.Y -= this.scroll
	shifted.Height += this.scroll
	layout := newRowLayout(shifted, Padding)
	layout.x += 2
	layout.width -= 4
	switch this.Tab {
	case TabShow:
		this.Show.Draw(layout)
	case TabProps:
		this.Props.Draw(layout)
	case TabTools:
		this.Tools.Draw(layout)
	case TabCube:
		this.Cube.Draw(layout)
	}
	raylib.EndScissorMode()
	this.content = layout.y - shifted.Y + Padding
}

func defaultPalette() string { return render.DefaultPalette }
