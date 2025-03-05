package component

import (
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
