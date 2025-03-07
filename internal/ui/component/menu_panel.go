package component

import (
	"math/rand"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type MenuPanel struct {
	PanelBase
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	connectionStatus int
	retryToggle      bool

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
	menu.itemBounds.Y += menu.itemBounds.Height / 2
	menu.renderStatus()
	menu.itemBounds.Y += menu.itemBounds.Height / 2
	menu.renderConnect()
	// TODO: maybe a text area with the last message? or maybe just pass a message to ConsolePanel
}

func (menu *MenuPanel) renderStatus() {
	message := ""
	innerColor := raylib.Blank
	outerColor := raylib.Blank
	switch menu.connectionStatus {
	case 0:
		innerColor, outerColor = raylib.Pink, raylib.Red
		message = "Disconnected"
	case 2:
		innerColor, outerColor = raylib.Green, raylib.Lime
		message = "Connected"
	case 1, 3:
		innerColor, outerColor = raylib.SkyBlue, raylib.Blue
		message = "Processing"
	}

	additionalXPadding := menu.itemBounds.Width / 4
	circleCenterXOffset := menu.padding.X + menu.itemBounds.Height/2
	raylib.DrawCircleGradient(
		int32(menu.itemBounds.X+circleCenterXOffset+additionalXPadding),
		int32(menu.itemBounds.Y+menu.padding.Y+menu.itemBounds.Height/2),
		style.TextSize,
		innerColor,
		outerColor)

	messageBounds := raylib.NewRectangle(
		menu.itemBounds.X+circleCenterXOffset*2+additionalXPadding,
		menu.itemBounds.Y+menu.padding.Y,
		menu.itemBounds.Width-circleCenterXOffset*2-additionalXPadding,
		menu.itemBounds.Height)
	utils.RenderText(message, messageBounds, style.TextColorNormal)

	menu.itemBounds.Y += menu.itemBounds.Height + menu.padding.Y*2
}

func (menu *MenuPanel) renderConnect() {
	toggleWidth := menu.itemBounds.Width/2 - menu.padding.X*2
	toggleBounds := raylib.NewRectangle(
		menu.itemBounds.X+menu.padding.X+style.TextLineSpacing,
		menu.itemBounds.Y+menu.padding.Y+style.TextLineSpacing/2,
		style.TextSize,
		style.TextSize)
	buttonWidth := toggleWidth
	buttonBounds := raylib.NewRectangle(
		menu.itemBounds.X+toggleWidth+menu.padding.X*3,
		menu.itemBounds.Y+menu.padding.Y,
		buttonWidth,
		menu.itemBounds.Height)

	menu.retryToggle = raygui.CheckBox(toggleBounds, "Keep Retrying", menu.retryToggle)

	switch menu.connectionStatus {
	case 0:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connect")) {
			menu.connectionStatus = 1
		}
	case 1:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connecting..."))
		raygui.Enable()
		if menu.test > 100 {
			if menu.retryToggle && rand.Intn(3) == 0 {
				menu.connectionStatus = 2
			} else if !menu.retryToggle {
				menu.connectionStatus = 2
			} else {
				println("connection retrying")
			}
			menu.test = 0
		}
	case 2:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnect")) {
			menu.connectionStatus = 3
		}
	case 3:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnecting..."))
		raygui.Enable()
		if menu.test > 100 {
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
