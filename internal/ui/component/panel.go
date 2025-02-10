package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
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
			nav.Pos.X+float32(nav.index)*nav.buttonWidth,
			nav.Pos.Y,
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
	raygui.Panel(raylib.NewRectangle(menu.Pos.X, menu.Pos.Y, menu.Width, menu.Height), "")
}

type ConsolePanel struct {
	PanelBase
	messages        []string
	maxMessageCount uint16
	test            int
	scroll          int32
	itemFocused     int
	useScrollBar    bool
}

func (cp *ConsolePanel) Update() {
	cp.resize()
	cp.test++
	if cp.test%50 == 0 {
		cp.messages = append([]string{fmt.Sprintf("Message #%d: some ; aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", cp.test)}, cp.messages...)
	}
	if len(cp.messages) > int(cp.maxMessageCount) {
		cp.messages = cp.messages[:cp.maxMessageCount]
	}

}

func (cp *ConsolePanel) Render() {
	bounds := raylib.NewRectangle(cp.Pos.X, cp.Pos.Y, cp.Width, cp.Height)
	raygui.Panel(bounds, "")

	content := raylib.NewRectangle(cp.Pos.X+5, cp.Pos.Y+5, cp.Width-10, cp.Height-10)
	GuiListViewEx(cp.messages, content, &cp.scroll)
}

type EditPanel struct{ PanelBase }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	raygui.Panel(raylib.NewRectangle(ep.Pos.X, ep.Pos.Y, ep.Width, ep.Height), "")
}

type PlaceholderPanel struct{ PanelBase }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.Pos.X, pp.Pos.Y, pp.Width, pp.Height), "")
}
