package component

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/Tariomka/led-common-lib/pkg/communication/uart"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Factory functions of KEY1..KEY7 (uploaded shows may remap them).
var keyNames = [7]string{"Prev", "Next", "Speed+", "Speed-", "Pause", "Model", "Display"}

// MenuPanel connects to the Pico and controls the cube: status, keys and cartoon selection.
type MenuPanel struct {
	Panel
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	retryToggle bool

	status         uart.Status
	statusReceived time.Time
	cartoon        int32
	cartoonEdit    bool

	channel chan any
}

func newMenuPanel(base Panel) *MenuPanel {
	menuPanel := &MenuPanel{
		Panel:   base,
		padding: raylib.NewVector2(10, 10),
		channel: make(chan any, 256),
	}

	global.RegisterMessageReceiver(
		constants.UIMenuPanel,
		func(message any) { menuPanel.channel <- message })

	return menuPanel
}

func (this *MenuPanel) Update() {
	this.resize()
	this.receive()

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
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y*2
	this.renderCubeStatus()
	this.renderControls()
}

func (this *MenuPanel) renderStatus() {
	message := ""
	innerColor := raylib.Blank
	outerColor := raylib.Blank
	switch global.ConnectionStatus {
	case global.Disconnect:
		innerColor, outerColor = raylib.Pink, raylib.Red
		message = "Disconnected"
	case global.Connected:
		innerColor, outerColor = raylib.Green, raylib.Lime
		message = "Connected"
	case global.Connecting, global.Disconnecting:
		innerColor, outerColor = raylib.SkyBlue, raylib.Blue
		message = "Processing"
	}

	additionalXPadding := this.itemBounds.Width / 4
	circleCenterXOffset := this.padding.X + this.itemBounds.Height/2
	raylib.DrawCircleGradient(
		raylib.NewVector2(
			this.itemBounds.X+circleCenterXOffset+additionalXPadding,
			this.itemBounds.Y+this.padding.Y+this.itemBounds.Height/2),
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

	raygui.CheckBox(toggleBounds, "Keep Retrying", &this.retryToggle)

	switch global.ConnectionStatus {
	case global.Disconnect:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connect")) {
			global.ConnectionStatus = global.Connecting
			global.SendMessage(constants.TCPClient, models.TCPConnectMessage{})
		}
	case global.Connecting:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_CLOSE, "Connecting..."))
		raygui.Enable()
	case global.Connected:
		if raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnect")) {
			global.ConnectionStatus = global.Disconnecting
			global.SendMessage(constants.TCPClient, models.TCPDisconnectMessage{})
		}
	case global.Disconnecting:
		raygui.Disable()
		raygui.Button(buttonBounds, raygui.IconText(raygui.ICON_LOCK_OPEN, "Disconnecting..."))
		raygui.Enable()
	}
}

func (this *MenuPanel) line(text string, textColor color.RGBA) {
	bounds := raylib.NewRectangle(this.X+this.padding.X, this.itemBounds.Y, this.Width-this.padding.X*2, style.TextSize)
	utils.RenderText(text, bounds, textColor)
	this.itemBounds.Y += style.TextSize + this.padding.Y/2
}

func (this *MenuPanel) renderCubeStatus() {
	if this.statusReceived.IsZero() {
		this.line("Cube status: not received yet", style.TextColorNormal)
		this.itemBounds.Y += this.padding.Y
		return
	}

	status := this.status
	age := time.Since(this.statusReceived).Round(time.Second)
	textColor := style.TextColorNormal
	if age > 5*time.Second {
		textColor = raylib.Orange
	}
	this.line(fmt.Sprintf("Cube status (%s ago)", age), textColor)
	this.line("Mode: "+status.Mode, style.TextColorNormal)
	if status.Mode != "Text" {
		this.line("Uploaded shows play only in Text mode", raylib.Orange)
		this.line("(DIP SW1 and SW2 ON)", raylib.Orange)
	}
	this.line(fmt.Sprintf("Cartoon: %d / %d", status.Cartoon+1, status.Count), style.TextColorNormal)
	this.line(fmt.Sprintf("Speed: %d.%02d", status.Speed/100, status.Speed%100), style.TextColorNormal)
	playModel := "repeat one"
	if status.PlayAll {
		playModel = "play all"
	}
	this.line("Play model: "+playModel, style.TextColorNormal)
	this.line(fmt.Sprintf("Paused: %t, display off: %t", status.Pause, status.DisplayOff), style.TextColorNormal)
	this.itemBounds.Y += this.padding.Y
}

func (this *MenuPanel) renderControls() {
	disabled := global.ConnectionStatus != global.Connected
	height := style.TextSize + style.BorderWidth*2 + style.TextPadding*2
	this.line("Keys 1-7 (factory functions)", style.TextColorNormal)

	columns := float32(4)
	width := (this.Width - this.padding.X*(columns+1)) / columns
	for key, name := range keyNames {
		row, column := float32(key/4), float32(key%4)
		bounds := raylib.NewRectangle(
			this.X+this.padding.X+column*(width+this.padding.X),
			this.itemBounds.Y+row*(height+this.padding.Y),
			width,
			height)
		utils.RenderWithCondition(func() {
			if raygui.Button(bounds, name) {
				this.command("Key " + strconv.Itoa(key+1))
			}
		}, disabled)
	}
	this.itemBounds.Y += (height + this.padding.Y) * 2

	width = (this.Width - this.padding.X*3) / 2
	bounds := raylib.NewRectangle(this.X+this.padding.X, this.itemBounds.Y, width, height)
	utils.RenderWithCondition(func() {
		if raygui.Button(bounds, "Replay") {
			this.command("Action 1")
		}
		bounds.X += width + this.padding.X
		if raygui.Button(bounds, "Speed reset") {
			this.command("Action 5")
		}
	}, disabled)
	this.itemBounds.Y += height + this.padding.Y

	maxCartoon := max(this.status.Count, 1)
	bounds = raylib.NewRectangle(this.X+this.padding.X, this.itemBounds.Y, width, height)
	utils.RenderWithCondition(func() {
		if raygui.Spinner(bounds, "", &this.cartoon, 1, maxCartoon, this.cartoonEdit) {
			this.cartoonEdit = !this.cartoonEdit
		}
		bounds.X += width + this.padding.X
		if raygui.Button(bounds, "Play cartoon") {
			this.command("Cartoon " + strconv.Itoa(int(this.cartoon)-1))
		}
	}, disabled)
	this.itemBounds.Y += height + this.padding.Y
}

func (this *MenuPanel) command(line string) {
	global.SendMessage(constants.TCPClient, models.TCPCommandMessage{Line: line})
}

// receive handles the messages queued since the last frame.
func (this *MenuPanel) receive() {
	for {
		select {
		case message := <-this.channel:
			this.handle(message)
		default:
			return
		}
	}
}

func (this *MenuPanel) handle(message any) {
	switch message := message.(type) {
	case models.ConnectedMessage:
		global.ConnectionStatus = global.Connected
	case models.DisconnectedMessage:
		if !this.retryToggle || global.ConnectionStatus == global.Disconnecting {
			global.ConnectionStatus = global.Disconnect
			return
		}
		global.ConnectionStatus = global.Connecting
		global.SendMessage(constants.TCPClient, models.TCPConnectMessage{})
	case models.ConsoleLineMessage:
		if message.Source == models.SourceCube {
			this.cubeLine(message.Text)
		}
	}
}

func (this *MenuPanel) cubeLine(line string) {
	event := uart.ParseLine(line)
	switch event.Kind {
	case uart.EventStatus:
		this.status, this.statusReceived = event.Status, time.Now()
		if this.cartoon < 1 {
			this.cartoon = 1
		}
	case uart.EventMode:
		this.status.Mode = event.Text
	case uart.EventCartoon:
		this.status.Cartoon = event.Values[0]
	case uart.EventSpeed:
		this.status.Speed = event.Values[0]
	case uart.EventPlayModel:
		this.status.PlayAll = event.Values[0] == 1
	case uart.EventPause:
		this.status.Pause = event.Values[0] == 1
	case uart.EventDisplayOff:
		this.status.DisplayOff = event.Values[0] == 1
	}
}
