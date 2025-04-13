package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type ConsolePanel struct {
	Panel
	messageBounds raylib.Rectangle

	messages         *common.RingArray[string]
	visibleLineCount int

	test int

	currentScrollIndex  int
	endScrollIndex      int
	itemFocused         int
	useScrollbar        bool
	scrollbarSliderSize float32
}

func newConsolePanel(base Panel) *ConsolePanel {
	return &ConsolePanel{
		Panel:       base,
		messages:    common.NewRingArray[string](100),
		itemFocused: -1,
	}
}

func (cp *ConsolePanel) Update() {
	cp.testData() // FIXME: to be deleted
	cp.resize()

	cp.messageBounds = raylib.NewRectangle(
		cp.X+style.ListItemSpacing,
		cp.Y+style.ListItemSpacing+style.BorderWidth,
		cp.Width-style.ListItemSpacing-2*style.BorderWidth,
		style.ListItemHeight)

	if (style.ListItemHeight+style.ListItemSpacing)*float32(cp.messages.Length()) > cp.Height {
		cp.useScrollbar = true
		cp.messageBounds.Width -= style.ListScrollWidth
	}

	cp.visibleLineCount = min(
		int(cp.Height/(style.ListItemHeight+style.ListItemSpacing)),
		cp.messages.Length())
	if cp.currentScrollIndex < 0 || cp.currentScrollIndex > cp.messages.Length()-cp.visibleLineCount {
		cp.currentScrollIndex = 0
	}
	cp.endScrollIndex = cp.currentScrollIndex + cp.visibleLineCount

	if style.GuiState != raygui.STATE_DISABLED && !raygui.IsLocked() {
		cp.updateScrollIndex()
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

	initialBoundY := cp.messageBounds.Y
	messageHeight := style.ListItemHeight + style.ListItemSpacing
	for i := range cp.visibleLineCount {
		if raylib.CheckCollisionPointRec(mousePoint, cp.messageBounds) {
			cp.itemFocused = cp.currentScrollIndex + i
			break
		}

		cp.messageBounds.Y += messageHeight // Update message box y position for next message line
	}
	cp.messageBounds.Y = initialBoundY // Reset message box Y position

	if !cp.useScrollbar {
		return
	}

	// TODO: add horizontal scrolling
	cp.currentScrollIndex -= int(raylib.GetMouseWheelMoveV().Y)
	if cp.currentScrollIndex < 0 {
		cp.currentScrollIndex = 0
	} else if cp.currentScrollIndex > (cp.messages.Length() - cp.visibleLineCount) {
		cp.currentScrollIndex = cp.messages.Length() - cp.visibleLineCount
	}
	cp.endScrollIndex = min(cp.currentScrollIndex+cp.visibleLineCount, cp.messages.Length())
}

func (cp *ConsolePanel) updateScrollbar() {
	if !cp.useScrollbar {
		return
	}

	percentVisible := float32((cp.endScrollIndex - cp.currentScrollIndex)) / float32(cp.messages.Length())
	cp.scrollbarSliderSize = cp.Height * percentVisible
}

func (cp *ConsolePanel) renderScrollbar() {
	if !cp.useScrollbar {
		return
	}

	prevSliderSize, prevScrollSpeed := style.GetScrollbarStyle()
	style.SetScrollbarStyle(
		int64(cp.scrollbarSliderSize),
		int64(cp.messages.Length()-cp.visibleLineCount))

	scrollBarBounds := raylib.NewRectangle(
		cp.X+cp.Width-style.ListBorderWidth-style.ListScrollWidth,
		cp.Y+style.ListBorderWidth,
		style.ListScrollWidth,
		cp.Height-style.BorderWidth)
	cp.currentScrollIndex = int(raygui.ScrollBar(
		scrollBarBounds,
		int32(cp.currentScrollIndex),
		0,
		int32(cp.messages.Length()-cp.visibleLineCount)))

	// cp.currentScrollIndex = int(raygui.Slider(
	// 	raylib.NewRectangle(
	// 		cp.X,
	// 		cp.Y+cp.Height-style.ListScrollWidth,
	// 		cp.Width-style.ListScrollWidth,
	// 		style.ListScrollWidth),
	// 	"", "", // no text
	// 	float32(cp.currentScrollIndex),
	// 	0,
	// 	float32(cp.messages.Length()-cp.visibleLineCount)))

	style.SetScrollbarStyle(prevSliderSize, prevScrollSpeed)
}

func (cp *ConsolePanel) renderMessages() {
	if cp.messages.Length() < 1 {
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
			cp.messages.Get(cp.currentScrollIndex+i),
			utils.GetTextBounds(cp.messageBounds),
			textColor)

		// Update item rectangle y position for next item
		cp.messageBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}
}

// FIXME: remove later
func (cp *ConsolePanel) testData() {
	cp.test++
	if cp.test%50 == 0 {
		cp.messages.Add("")
		cp.messages.Add(fmt.Sprintf("Message #%d: some ; aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", cp.test))

		// for r := 'a'; r < 'z'; r++ {
		// 	R := unicode.ToUpper(r)

		// 	fmt.Printf("Char - %c; Width - %f\n", r, utils.GetTextWidth(string(r)))
		// 	fmt.Printf("Char - %c; Width - %f\n", R, utils.GetTextWidth(string(R)))
		// }

		// fmt.Printf("Space; Width - %f\n", utils.GetTextWidth(string(' ')))
		// fmt.Printf("Dot; Width - %f\n", utils.GetTextWidth(string('.')))
		// fmt.Printf("Slash; Width - %f\n", utils.GetTextWidth(string('/')))

		// move := raylib.GetMouseWheelMove()
		// moveBoth := raylib.GetMouseWheelMoveV()
		// fmt.Printf("Vertical scroll: move - %f; moveV - %f \n", move, moveBoth.Y)
		// fmt.Printf("Horizontal scroll: moveV - %f \n", moveBoth.X)
	}
}
