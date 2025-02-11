package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type NavigationPanel struct {
	PanelBase
	parent       PanelSelector
	buttonWidth  float32
	buttonStates []bool
	index        int
}

func (nav *NavigationPanel) SetParent(parent PanelSelector) {
	nav.parent = parent
	panelCount := nav.parent.PanelCount()
	nav.buttonWidth = nav.Width / float32(panelCount)
	nav.buttonStates = make([]bool, panelCount)
}

func (nav *NavigationPanel) Update() {
	nav.resize(func() { nav.buttonWidth = nav.Width / float32(len(nav.buttonStates)) })
}

func (nav *NavigationPanel) Render() {
	nav.index = 0
	for name, panel := range nav.parent.IteratePanels() {
		// TODO: Add tooltip to buttons.
		// Note: raygui-go does not have tooltip bindings so probably need to implement by hand.
		bounds := raylib.NewRectangle(
			nav.X+float32(nav.index)*nav.buttonWidth,
			nav.Y,
			nav.buttonWidth,
			nav.Height)

		if !raygui.Button(bounds, name) {
			if nav.buttonStates[nav.index] {
				raylib.DrawRectangleRec(
					bounds,
					common.IntToRGBAEx(
						raygui.GetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED),
						125))
				raylib.DrawRectangleLinesEx(
					bounds,
					float32(raygui.GetStyle(raygui.BUTTON, raygui.BORDER_WIDTH)),
					common.IntToRGBAEx(
						raygui.GetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED),
						230))
			}

			nav.index++
			continue
		}

		previousState := nav.buttonStates[nav.index]
		for i := range nav.buttonStates {
			nav.buttonStates[i] = false
		}
		nav.buttonStates[nav.index] = !previousState
		if nav.buttonStates[nav.index] {
			nav.parent.SetSelectedPanel(panel)
		} else {
			nav.parent.SetSelectedPanel(nil)
		}
		nav.index++
	}
}

type MenuPanel struct{ PanelBase }

func (menu *MenuPanel) Update() { menu.resize() }

func (menu *MenuPanel) Render() {
	raygui.Panel(raylib.NewRectangle(menu.X, menu.Y, menu.Width, menu.Height), "")
}

type ConsolePanel struct {
	PanelBase
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	messages         []string
	maxMessageCount  uint16
	messageCount     int
	visibleLineCount int

	test int

	scrollIndex         int
	itemFocused         int
	useScrollbar        bool
	scrollbarSliderSize float32
}

func (cp *ConsolePanel) Update() {
	cp.testData() // to be deleted

	cp.resize()
	cp.messageCount = len(cp.messages)

	// Check if we need a scroll bar
	if (style.ListItemHeight+style.ListItemSpacing)*float32(cp.messageCount) > cp.Height {
		cp.useScrollbar = true
	}

	// Define base item rectangle [0]
	cp.itemBounds = raylib.NewRectangle(
		cp.X+style.ListItemSpacing,
		cp.Y+style.ListItemSpacing+style.BorderWidth,
		cp.Width-2*style.ListItemSpacing-style.BorderWidth,
		style.ListItemHeight)
	if cp.useScrollbar {
		cp.itemBounds.Width -= style.ListScrollWidth
	}

	// Get items on the list
	cp.visibleLineCount = int(cp.Height / (style.ListItemHeight + style.ListItemSpacing))
	if cp.visibleLineCount > cp.messageCount {
		cp.visibleLineCount = cp.messageCount
	}

	if cp.scrollIndex < 0 || cp.scrollIndex > cp.messageCount-cp.visibleLineCount {
		cp.scrollIndex = 0
	}
	endIndex := cp.scrollIndex + cp.visibleLineCount

	// Update control
	//--------------------------------------------------------------------
	if style.GuiState != raygui.STATE_DISABLED && !raygui.IsLocked() {
		mousePoint := raylib.GetMousePosition()

		// Check mouse inside list view
		if raylib.CheckCollisionPointRec(mousePoint, cp.Rectangle) {
			style.GuiState = raygui.STATE_FOCUSED

			// Check focused and selected item
			for i := 0; i < cp.visibleLineCount; i++ {
				if raylib.CheckCollisionPointRec(mousePoint, cp.itemBounds) {
					cp.itemFocused = cp.scrollIndex + i
					break
				}

				// Update item rectangle y position for next item
				cp.itemBounds.Y += style.ListItemHeight + style.ListItemSpacing
			}

			if cp.useScrollbar {
				wheelMove := int(raylib.GetMouseWheelMove())
				cp.scrollIndex -= wheelMove

				if cp.scrollIndex < 0 {
					cp.scrollIndex = 0
				} else if cp.scrollIndex > (cp.messageCount - cp.visibleLineCount) {
					cp.scrollIndex = cp.messageCount - cp.visibleLineCount
				}

				endIndex = cp.scrollIndex + cp.visibleLineCount
				if endIndex > cp.messageCount {
					endIndex = cp.messageCount
				}
			}
		} else {
			cp.itemFocused = -1
		}

		// Reset item rectangle y to [0]
		cp.itemBounds.Y = cp.Y + style.ListItemSpacing + style.BorderWidth
	}

	if cp.useScrollbar {
		// Calculate percentage of visible items and apply same percentage to scrollbar
		percentVisible := float32((endIndex - cp.scrollIndex) / cp.messageCount)
		cp.scrollbarSliderSize = cp.Height * percentVisible
	}
}

func (cp *ConsolePanel) Render() {
	raygui.Panel(cp.Rectangle, "")

	// Draw visible items
	for i := 0; (i < cp.visibleLineCount) && (cp.messageCount > 0); i++ {
		if cp.scrollIndex+i == cp.itemFocused {
			// Draw item focused
			utils.RenderRectangle(
				cp.itemBounds,
				style.ListBorderWidth,
				style.ListBorderColorFocused,
				style.ListBaseColorFocused)
			utils.RenderText(
				cp.messages[cp.scrollIndex+i],
				utils.GetTextBounds(cp.itemBounds),
				style.ListTextColorFocused)
		} else {
			// Draw item normal
			utils.RenderText(
				cp.messages[cp.scrollIndex+i],
				utils.GetTextBounds(cp.itemBounds),
				style.ListTextColorNormal)
		}

		// Update item rectangle y position for next item
		cp.itemBounds.Y += style.ListItemHeight + style.ListItemSpacing
	}

	cp.renderScrollbar()
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
	cp.scrollIndex = int(raygui.ScrollBar(
		scrollBarBounds,
		int32(cp.scrollIndex),
		0,
		int32(cp.messageCount-cp.visibleLineCount)))

	style.SetScrollbarStyle(prevSliderSize, prevScrollSpeed)
}

func (cp *ConsolePanel) testData() {
	cp.test++
	if cp.test%50 == 0 {
		cp.messages = append([]string{fmt.Sprintf("Message #%d: some ; aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", cp.test)}, cp.messages...)
	}
	if len(cp.messages) > int(cp.maxMessageCount) {
		cp.messages = cp.messages[:cp.maxMessageCount]
	}
}

type EditPanel struct{ PanelBase }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	raygui.Panel(raylib.NewRectangle(ep.X, ep.Y, ep.Width, ep.Height), "")
}

type PlaceholderPanel struct{ PanelBase }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.X, pp.Y, pp.Width, pp.Height), "")
}
