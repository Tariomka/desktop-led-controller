package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type MenuPanel struct {
	PanelBase
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	connectionStatus int

	test int
}

func (menu *MenuPanel) Update() {
	menu.testData() // FIXME: to be deleted
	menu.resize()

	menu.itemBounds = raylib.NewRectangle(
		menu.X+style.BorderWidth,
		menu.Y+style.BorderWidth,
		menu.Width-2*style.BorderWidth,
		style.TextSize+style.TextLineSpacing)
}

func (menu *MenuPanel) Render() {
	menu.renderPanel()
	// status text with colored bubble
	// connect button with connect icon / when connected - disconnect button
	menu.renderStatus()
	menu.renderConnect()
	// retry toggle button, probably on the same line as connect button
	// maybe a text area with the last message? or maybe just pass a message to ConsolePanel
}

func (menu *MenuPanel) renderStatus() {

}

func (menu *MenuPanel) renderConnect() {
	toggleWidth := menu.itemBounds.Width/2 - menu.padding.X*2
	buttonWidth := toggleWidth
	buttonBounds := raylib.NewRectangle(
		menu.itemBounds.X+toggleWidth+menu.padding.X*3,
		menu.itemBounds.Y+menu.padding.Y,
		buttonWidth,
		menu.itemBounds.Height)

	switch menu.connectionStatus {
	case 0:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connect")) {
			menu.connectionStatus = 1
		}
	case 1:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connecting..."))
		raygui.Enable()
		if menu.test > 200 {
			menu.test = 0
			menu.connectionStatus = 2
		}
	case 2:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnect")) {
			menu.connectionStatus = 3
		}
	case 3:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnecting..."))
		raygui.Enable()
		if menu.test > 200 {
			menu.test = 0
			menu.connectionStatus = 0
		}
	default:
	}
}

// FIXME: remove later
func (menu *MenuPanel) testData() {
	if menu.connectionStatus == 1 || menu.connectionStatus == 3 {
		menu.test++
	}
}
