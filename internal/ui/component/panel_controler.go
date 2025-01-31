package component

import (
	_ "embed"
	"iter"

	"github.com/gen2brain/raylib-go/raygui"
)

//go:embed panel_style.rgs
var style []byte

type PanelSelector interface {
	SetSelectedPanel(panel Renderer)
	IteratePanels() iter.Seq2[string, Renderer]
	PanelCount() int
}

type PanelControler struct {
	navBar Renderer

	panels        []NamedPanel
	selectedPanel Renderer

	dialogs []Element
}

func NewPanelControler(panelConfig ...PanelConfigFunc) Renderer {
	navBarHeight := float32(24)
	navBarPosition := func(pb *PanelBase) { pb.height = navBarHeight }
	shiftPosition := func(pb *PanelBase) {
		pb.pos.Y += navBarHeight
		pb.height -= navBarHeight
	}

	controller := &PanelControler{
		navBar: NewPanel[*NavigationPanel](append(panelConfig, navBarPosition)...),
		panels: []NamedPanel{
			NewNamedPanel[*EditPanel]("Edit", append(panelConfig, shiftPosition)...),
			NewNamedPanel[*MenuPanel]("Menu", append(panelConfig, shiftPosition)...),
			NewNamedPanel[*ConsolePanel]("Console", append(panelConfig, shiftPosition)...),
			NewNamedPanel[*PlaceholderPanel]("Placeholder", append(panelConfig, shiftPosition)...),
		},
		dialogs: []Element{
			NewElement[*ExitDialog](),
		},
	}

	controller.navBar.(*NavigationPanel).SetParent(controller)
	controller.setStyle()

	return controller
}

func (pc *PanelControler) Update() {
	pc.navBar.Update()
	for _, panel := range pc.panels {
		panel.Update()
	}
	for _, dialog := range pc.dialogs {
		dialog.Update()
	}
}

func (pc *PanelControler) Render() {
	if pc.selectedPanel != nil {
		pc.selectedPanel.Render()
	}
	pc.navBar.Render()
	for _, dialog := range pc.dialogs {
		dialog.Render()
	}
}

func (pc *PanelControler) SetSelectedPanel(panel Renderer) { pc.selectedPanel = panel }

func (pc *PanelControler) IteratePanels() iter.Seq2[string, Renderer] {
	return func(yield func(string, Renderer) bool) {
		for _, panel := range pc.panels {
			if !yield(panel.Title, panel.Renderer) {
				return
			}
		}
	}
}

func (pc *PanelControler) PanelCount() int { return len(pc.panels) }

func (pc *PanelControler) setStyle() {
	// Base style
	raygui.LoadStyleFromMemory(style)

	// Updates, maybe set everything myself or create custom style seleretaly?
	raygui.SetStyle(0, raygui.BORDER_COLOR_FOCUSED, 0xff_00_00_7f)
	raygui.SetStyle(0, raygui.BASE_COLOR_FOCUSED, 0xff_00_00_2f)

	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED, 0xe0_3c_46_ff)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED, 0x5b_1e_20_ff)
}
