package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type ConsolePanel struct {
	Panel
	messageBounds raylib.Rectangle

	messages         *common.RingBuffer[string]
	visibleLineCount int

	currentScrollIndex  int
	endScrollIndex      int
	itemFocused         int
	useScrollbar        bool
	scrollbarSliderSize float32

	channel chan string
}

func newConsolePanel(base Panel) *ConsolePanel {
	consolePanel := &ConsolePanel{
		Panel:       base,
		messages:    common.NewRingBuffer[string](100),
		itemFocused: -1,
		channel:     make(chan string),
	}

	go consolePanel.channelLoop()
	global.Messenger.RegisterReceiver(
		constants.UIConsolePanel,
		func(message any) {
			stringMessage, ok := message.(string)
			if !ok {
				fmt.Printf("[CONSOLE_PANEL] Not a string message received, message %#v\n", message)
				return
			}

			consolePanel.channel <- stringMessage
		})

	return consolePanel
}

func (this *ConsolePanel) Update() {
	this.resize()

	this.messageBounds = raylib.NewRectangle(
		this.X+style.ListItemSpacing,
		this.Y+style.ListItemSpacing+style.BorderWidth,
		this.Width-style.ListItemSpacing-2*style.BorderWidth,
		style.ListItemHeight)

	if (style.ListItemHeight+style.ListItemSpacing)*float32(this.messages.Length()) > this.Height {
		this.useScrollbar = true
		this.messageBounds.Width -= style.ListScrollWidth
	}

	this.visibleLineCount = min(
		int(this.Height/(style.ListItemHeight+style.ListItemSpacing)),
		this.messages.Length())
	if this.currentScrollIndex < 0 || this.currentScrollIndex > this.messages.Length()-this.visibleLineCount {
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
}

func (this *ConsolePanel) updateScrollIndex() {
	mousePoint := raylib.GetMousePosition()
	if !raylib.CheckCollisionPointRec(mousePoint, this.Rectangle) {
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

	// TODO: add horizontal scrolling
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

	percentVisible := float32((this.endScrollIndex - this.currentScrollIndex)) / float32(this.messages.Length())
	this.scrollbarSliderSize = this.Height * percentVisible
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
		this.Height-style.BorderWidth)
	this.currentScrollIndex = int(raygui.ScrollBar(
		scrollBarBounds,
		int32(this.currentScrollIndex),
		0,
		int32(this.messages.Length()-this.visibleLineCount)))

	// this.currentScrollIndex = int(raygui.Slider(
	// 	raylib.NewRectangle(
	// 		this.X,
	// 		this.Y+this.Height-style.ListScrollWidth,
	// 		this.Width-style.ListScrollWidth,
	// 		style.ListScrollWidth),
	// 	"", "", // no text
	// 	float32(this.currentScrollIndex),
	// 	0,
	// 	float32(this.messages.Length()-this.visibleLineCount)))

	style.SetScrollbarStyle(prevSliderSize, prevScrollSpeed)
}

func (this *ConsolePanel) renderMessages() {
	if this.messages.Length() < 1 {
		return
	}

	for i := range this.visibleLineCount {
		textColor := style.ListTextColorNormal

		if this.currentScrollIndex+i == this.itemFocused {
			textColor = style.ListTextColorFocused
			utils.RenderRectangle( // Focused border
				this.messageBounds,
				style.ListBorderWidth,
				style.ListBorderColorFocused,
				style.ListBaseColorFocused)
		}

		utils.RenderText(
			this.messages.Get(this.currentScrollIndex+i),
			utils.GetTextBounds(this.messageBounds),
			textColor)

		// Update item rectangle y position for next item
		this.messageBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}
}

// Blocking state loop
func (this *ConsolePanel) channelLoop() {
	for {
		this.messages.Add(<-this.channel)
	}
}
