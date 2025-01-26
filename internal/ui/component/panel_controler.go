package component

import (
	_ "embed"
	"iter"

	"github.com/gen2brain/raylib-go/raygui"
)

//go:embed panel_style.rgs
var style []byte

type PanelSelector interface {
	SetSelectedPanel(panel Panel)
	IteratePanels() iter.Seq2[int, Panel]
	PanelCount() int
}

type PanelControler struct {
	navBar Panel

	panels        []Panel
	selectedPanel Panel

	dialogs []Element
}

func NewPanelControler(panelConfig ...PanelConfigFunc) Renderer {
	navBarHeight := float32(24)
	navBarPosition := func(pb *PanelBase) { pb.height = navBarHeight }
	shiftPosition := func(pb *PanelBase) {
		pb.pos.Y += navBarHeight
		pb.height -= navBarHeight
	}
	setTitle := func(name string) PanelConfigFunc {
		return func(pb *PanelBase) { pb.title = name }
	}

	controller := &PanelControler{
		navBar: NewPanel[*NavigationPanel](append(panelConfig, navBarPosition)...),
		panels: []Panel{
			NewPanel[*EditPanel](append(panelConfig, shiftPosition, setTitle("Edit"))...),
			NewPanel[*MenuPanel](append(panelConfig, shiftPosition, setTitle("Menu"))...),
			NewPanel[*ConsolePanel](append(panelConfig, shiftPosition, setTitle("Console"))...),
			NewPanel[*PlaceholderPanel](append(panelConfig, shiftPosition, setTitle("Placeholder"))...),
		},
		dialogs: []Element{
			NewElement[*ExitDialog](),
		},
	}

	controller.selectedPanel = controller.panels[0]
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

func (pc *PanelControler) SetSelectedPanel(panel Panel) { pc.selectedPanel = panel }

func (pc *PanelControler) IteratePanels() iter.Seq2[int, Panel] {
	return func(yield func(int, Panel) bool) {
		for index, panel := range pc.panels {
			if !yield(index, panel) {
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
