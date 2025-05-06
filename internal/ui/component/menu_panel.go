package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type MenuPanel struct {
	Panel
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	connectionStatus int
	retryToggle      bool

	channel chan any
}

func newMenuPanel(base Panel) *MenuPanel {
	menuPanel := &MenuPanel{
		Panel:   base,
		padding: raylib.NewVector2(10, 10),
		channel: make(chan any, 1),
	}

	go menuPanel.channelLoop()
	global.Messenger.RegisterReceiver(
		constants.UIMenuPanel,
		func(message any) { menuPanel.channel <- message })

	return menuPanel
}

func (this *MenuPanel) Update() {
	this.resize()

	this.itemBounds = raylib.NewRectangle(
		this.X+style.BorderWidth,
		this.Y+style.BorderWidth,
		this.Width-2*style.BorderWidth,
		style.TextSize+style.TextLineSpacing)
}

func (this *MenuPanel) Render() {
	this.renderPanel()
	this.itemBounds.Y += this.itemBounds.Height / 2
	this.renderStatus()
	this.itemBounds.Y += this.itemBounds.Height / 2
	this.renderConnect()
	// TODO: maybe a text area with the last message? or maybe just pass a message to ConsolePanel
	// TODO: add ip/port input area, maybe also save option to overwrite config.json
}

func (this *MenuPanel) renderStatus() {
	message := ""
	innerColor := raylib.Blank
	outerColor := raylib.Blank
	switch this.connectionStatus {
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

	additionalXPadding := this.itemBounds.Width / 4
	circleCenterXOffset := this.padding.X + this.itemBounds.Height/2
	raylib.DrawCircleGradient(
		int32(this.itemBounds.X+circleCenterXOffset+additionalXPadding),
		int32(this.itemBounds.Y+this.padding.Y+this.itemBounds.Height/2),
		style.TextSize,
		innerColor,
		outerColor)

	messageBounds := raylib.NewRectangle(
		this.itemBounds.X+circleCenterXOffset*2+additionalXPadding,
		this.itemBounds.Y+this.padding.Y,
		this.itemBounds.Width-circleCenterXOffset*2-additionalXPadding,
		this.itemBounds.Height)
	utils.RenderText(message, messageBounds, style.TextColorNormal)

	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y*2
}

func (this *MenuPanel) renderConnect() {
	toggleWidth := this.itemBounds.Width/2 - this.padding.X*2
	toggleBounds := raylib.NewRectangle(
		this.itemBounds.X+this.padding.X+style.TextLineSpacing,
		this.itemBounds.Y+this.padding.Y+style.TextLineSpacing/2,
		style.TextSize,
		style.TextSize)
	buttonWidth := toggleWidth
	buttonBounds := raylib.NewRectangle(
		this.itemBounds.X+toggleWidth+this.padding.X*3,
		this.itemBounds.Y+this.padding.Y,
		buttonWidth,
		this.itemBounds.Height)

	this.retryToggle = raygui.CheckBox(toggleBounds, "Keep Retrying", this.retryToggle)

	switch this.connectionStatus {
	case 0:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connect")) {
			this.connectionStatus = 1
			global.Messenger.Send(constants.TCPClient, models.TCPConnectMessage{})
		}
	case 1:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connecting..."))
		raygui.Enable()
	case 2:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnect")) {
			this.connectionStatus = 3
			global.Messenger.Send(constants.TCPClient, models.TCPDisconnectMessage{})
		}
	case 3:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnecting..."))
		raygui.Enable()
	default:
	}
}

// Blocking state loop
func (this *MenuPanel) channelLoop() {
	for {
		switch (<-this.channel).(type) {
		case models.ConnectedMessage:
			this.connectionStatus = 2
		case models.DisconnectedMessage:
			if !this.retryToggle {
				this.connectionStatus = 0
				continue
			}
			this.connectionStatus = 1
			global.Messenger.Send(constants.TCPClient, models.TCPConnectMessage{})
		}
	}
}
