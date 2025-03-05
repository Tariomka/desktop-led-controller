package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type ConsolePanel struct {
	PanelBase
	padding       raylib.Vector2
	messageBounds raylib.Rectangle

	// would be nice to have a ring buffer or something
	messages         []string
	maxMessageCount  uint16
	messageCount     int
	visibleLineCount int

	test int

	currentScrollIndex  int
	endScrollIndex      int
	itemFocused         int
	useScrollbar        bool
	scrollbarSliderSize float32
}

func (cp *ConsolePanel) Update() {
	cp.testData() // to be deleted
	cp.resize()

	cp.messageBounds = raylib.NewRectangle(
		cp.X+style.ListItemSpacing,
		cp.Y+style.ListItemSpacing+style.BorderWidth,
		cp.Width-2*style.ListItemSpacing-style.BorderWidth,
		style.ListItemHeight)
	cp.messageCount = len(cp.messages)

	if (style.ListItemHeight+style.ListItemSpacing)*float32(cp.messageCount) > cp.Height {
		cp.useScrollbar = true
		cp.messageBounds.Width -= style.ListScrollWidth
	}

	cp.visibleLineCount = min(
		int(cp.Height/(style.ListItemHeight+style.ListItemSpacing)),
		cp.messageCount)
	if cp.currentScrollIndex < 0 || cp.currentScrollIndex > cp.messageCount-cp.visibleLineCount {
		cp.currentScrollIndex = 0
	}
	cp.endScrollIndex = cp.currentScrollIndex + cp.visibleLineCount

	if style.GuiState != raygui.STATE_DISABLED && !raygui.IsLocked() {
		cp.updateScrollIndex()
		cp.messageBounds.Y = cp.Y + style.ListItemSpacing + style.BorderWidth // Reset message box height
	}

	cp.updateScrollbar()
}

func (cp *ConsolePanel) Render() {
	cp.renderPanel()
	cp.renderScrollbar()
	cp.renderMessages()
}

func (cp *ConsolePanel) updateScrollIndex() {
	mousePoint := raylib.GetMousePosition()
	if !raylib.CheckCollisionPointRec(mousePoint, cp.Rectangle) {
		cp.itemFocused = -1
		return
	}

	style.GuiState = raygui.STATE_FOCUSED

	for i := range cp.visibleLineCount {
		if raylib.CheckCollisionPointRec(mousePoint, cp.messageBounds) {
			cp.itemFocused = cp.currentScrollIndex + i
			break
		}

		// Update item rectangle y position for next item
		cp.messageBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}

	if !cp.useScrollbar {
		return
	}

	cp.currentScrollIndex -= int(raylib.GetMouseWheelMove())
	if cp.currentScrollIndex < 0 {
		cp.currentScrollIndex = 0
	} else if cp.currentScrollIndex > (cp.messageCount - cp.visibleLineCount) {
		cp.currentScrollIndex = cp.messageCount - cp.visibleLineCount
	}
	cp.endScrollIndex = min(cp.currentScrollIndex+cp.visibleLineCount, cp.messageCount)
}

func (cp *ConsolePanel) updateScrollbar() {
	if !cp.useScrollbar {
		return
	}

	percentVisible := float32((cp.endScrollIndex - cp.currentScrollIndex)) / float32(cp.messageCount)
	cp.scrollbarSliderSize = cp.Height * percentVisible
}

func (cp *ConsolePanel) renderScrollbar() {
	if !cp.useScrollbar {
		return
	}

	prevSliderSize, prevScrollSpeed := style.GetScrollbarStyle()
	style.SetScrollbarStyle(int64(cp.scrollbarSliderSize), int64(cp.messageCount-cp.visibleLineCount))

	scrollBarBounds := raylib.NewRectangle(
		cp.X+cp.Width-style.ListBorderWidth-style.ListScrollWidth,
		cp.Y+style.ListBorderWidth,
		style.ListScrollWidth,
		cp.Height-style.BorderWidth)
	cp.currentScrollIndex = int(raygui.ScrollBar(
		scrollBarBounds,
		int32(cp.currentScrollIndex),
		0,
		int32(cp.messageCount-cp.visibleLineCount)))

	style.SetScrollbarStyle(prevSliderSize, prevScrollSpeed)
}

func (cp *ConsolePanel) renderMessages() {
	if cp.messageCount < 1 {
		return
	}

	for i := range cp.visibleLineCount {
		textColor := style.ListTextColorNormal

		if cp.currentScrollIndex+i == cp.itemFocused {
			textColor = style.ListTextColorFocused
			utils.RenderRectangle( // Focused border
				cp.messageBounds,
				style.ListBorderWidth,
				style.ListBorderColorFocused,
				style.ListBaseColorFocused)
		}

		utils.RenderText(
			cp.messages[cp.currentScrollIndex+i],
			utils.GetTextBounds(cp.messageBounds),
			textColor)

		// Update item rectangle y position for next item
		cp.messageBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}
}

// remove later
func (cp *ConsolePanel) testData() {
	cp.test++
	if cp.test%50 == 0 {
		cp.messages = append([]string{fmt.Sprintf("Message #%d: some ; aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", cp.test)}, cp.messages...)
	}
	if len(cp.messages) > int(cp.maxMessageCount) {
		cp.messages = cp.messages[:cp.maxMessageCount]
	}
}
