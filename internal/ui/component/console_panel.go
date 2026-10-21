package component

import (
	"image/color"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const commandMaxLength = 180

// ConsolePanel shows the cube's and the Pico's log lines and sends console commands.
type ConsolePanel struct {
	Panel
	listBounds    raylib.Rectangle
	messageBounds raylib.Rectangle
	inputBounds   raylib.Rectangle

	messages         *data.RingBuffer[models.ConsoleLineMessage]
	visibleLineCount int

	currentScrollIndex  int
	endScrollIndex      int
	itemFocused         int
	useScrollbar        bool
	scrollbarSliderSize float32

	command     string
	commandEdit bool

	channel chan models.ConsoleLineMessage
}

func newConsolePanel(base Panel) *ConsolePanel {
	consolePanel := &ConsolePanel{
		Panel:       base,
		messages:    data.NewRingBuffer[models.ConsoleLineMessage](300),
		itemFocused: -1,
		channel:     make(chan models.ConsoleLineMessage, 256),
	}

	global.RegisterMessageReceiver(
		constants.UIConsolePanel,
		func(message any) {
			if line, ok := message.(models.ConsoleLineMessage); ok {
				consolePanel.channel <- line
			}
		})

	return consolePanel
}

func (this *ConsolePanel) Update() {
	this.resize()
	this.receive()

	inputHeight := style.TextSize + style.BorderWidth*2 + style.TextPadding*2
	this.listBounds = this.Rectangle
	this.listBounds.Height -= inputHeight + style.ListItemSpacing*2
	this.inputBounds = raylib.NewRectangle(
		this.X+style.ListItemSpacing,
		this.Y+this.Height-inputHeight-style.ListItemSpacing,
		this.Width-2*style.ListItemSpacing,
		inputHeight)

	this.messageBounds = raylib.NewRectangle(
		this.X+style.ListItemSpacing,
		this.Y+style.ListItemSpacing+style.BorderWidth,
		this.Width-style.ListItemSpacing-2*style.BorderWidth,
		style.ListItemHeight)

	this.useScrollbar = (style.ListItemHeight+style.ListItemSpacing)*float32(this.messages.Length()) > this.listBounds.Height
	if this.useScrollbar {
		this.messageBounds.Width -= style.ListScrollWidth
	}

	this.visibleLineCount = min(
		int(this.listBounds.Height/(style.ListItemHeight+style.ListItemSpacing)),
		this.messages.Length())
	if this.currentScrollIndex < 0 ||
		this.currentScrollIndex > this.messages.Length()-this.visibleLineCount {
		this.currentScrollIndex = 0
	}
	this.endScrollIndex = this.currentScrollIndex + this.visibleLineCount

	if style.GuiState != raygui.STATE_DISABLED && !raygui.IsLocked() {
		this.updateScrollIndex()
	}

	this.updateScrollbar()
}

func (this *ConsolePanel) Render() {
	this.renderPanel()
	this.renderScrollbar()
	this.renderMessages()
	this.renderInput()
}

// receive takes the lines queued since the last frame; the periodic Status replies are hidden.
func (this *ConsolePanel) receive() {
	for {
		select {
		case line := <-this.channel:
			if line.Source == models.SourceCube && strings.HasPrefix(line.Text, "Status:") {
				continue
			}
			this.messages.Add(line)
		default:
			return
		}
	}
}

func (this *ConsolePanel) renderInput() {
	buttonWidth := this.inputBounds.Width / 4
	textBounds := this.inputBounds
	textBounds.Width -= buttonWidth + style.ListItemSpacing
	buttonBounds := raylib.NewRectangle(
		textBounds.X+textBounds.Width+style.ListItemSpacing,
		this.inputBounds.Y,
		buttonWidth,
		this.inputBounds.Height)

	if raygui.TextBox(textBounds, &this.command, commandMaxLength, this.commandEdit) {
		this.commandEdit = !this.commandEdit
		if !this.commandEdit && raylib.IsKeyPressed(raylib.KeyEnter) {
			this.send()
		}
	}

	utils.RenderWithCondition(
		func() {
			if raygui.Button(buttonBounds, "Send") {
				this.send()
			}
		},
		global.ConnectionStatus != global.Connected)
}

func (this *ConsolePanel) send() {
	line := strings.TrimSpace(this.command)
	if line == "" || global.ConnectionStatus != global.Connected {
		return
	}
	global.SendMessage(constants.TCPClient, models.TCPCommandMessage{Line: line})
	this.command = ""
}

func (this *ConsolePanel) updateScrollIndex() {
	mousePoint := raylib.GetMousePosition()
	if !raylib.CheckCollisionPointRec(mousePoint, this.listBounds) {
		this.itemFocused = -1
		return
	}

	initialBoundY := this.messageBounds.Y
	messageHeight := style.ListItemHeight + style.ListItemSpacing
	for i := range this.visibleLineCount {
		if raylib.CheckCollisionPointRec(mousePoint, this.messageBounds) {
			this.itemFocused = this.currentScrollIndex + i
			break
		}

		this.messageBounds.Y += messageHeight // Update message box y position for next message line
	}
	this.messageBounds.Y = initialBoundY // Reset message box Y position

	if !this.useScrollbar {
		return
	}

	this.currentScrollIndex -= int(raylib.GetMouseWheelMoveV().Y)
	if this.currentScrollIndex < 0 {
		this.currentScrollIndex = 0
	} else if this.currentScrollIndex > (this.messages.Length() - this.visibleLineCount) {
		this.currentScrollIndex = this.messages.Length() - this.visibleLineCount
	}
	this.endScrollIndex = min(this.currentScrollIndex+this.visibleLineCount, this.messages.Length())
}

func (this *ConsolePanel) updateScrollbar() {
	if !this.useScrollbar {
		return
	}

	percentVisible := float32(this.endScrollIndex-this.currentScrollIndex) / float32(this.messages.Length())
	this.scrollbarSliderSize = this.listBounds.Height * percentVisible
}

func (this *ConsolePanel) renderScrollbar() {
	if !this.useScrollbar {
		return
	}

	prevSliderSize, prevScrollSpeed := style.GetScrollbarStyle()
	style.SetScrollbarStyle(
		int64(this.scrollbarSliderSize),
		int64(this.messages.Length()-this.visibleLineCount))

	scrollBarBounds := raylib.NewRectangle(
		this.X+this.Width-style.ListBorderWidth-style.ListScrollWidth,
		this.Y+style.ListBorderWidth,
		style.ListScrollWidth,
		this.listBounds.Height-style.BorderWidth)
	this.currentScrollIndex = int(raygui.ScrollBar(
		scrollBarBounds,
		int32(this.currentScrollIndex),
		0,
		int32(this.messages.Length()-this.visibleLineCount)))

	style.SetScrollbarStyle(prevSliderSize, prevScrollSpeed)
}

func lineColor(source models.ConsoleSource) color.RGBA {
	switch source {
	case models.SourcePico:
		return raylib.SkyBlue
	case models.SourceApp:
		return raylib.Gold
	}
	return style.ListTextColorNormal
}

func (this *ConsolePanel) renderMessages() {
	if this.messages.Length() < 1 {
		return
	}

	for i := range this.visibleLineCount {
		message := this.messages.Get(this.currentScrollIndex + i)
		textColor := lineColor(message.Source)

		if this.currentScrollIndex+i == this.itemFocused {
			textColor = style.ListTextColorFocused
			utils.RenderRectangle( // Focused border
				this.messageBounds,
				style.ListBorderWidth,
				style.ListBorderColorFocused,
				style.ListBaseColorFocused)
		}

		utils.RenderText(message.Text, utils.GetTextBounds(this.messageBounds), textColor)

		// Update item rectangle y position for next item
		this.messageBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}
}
