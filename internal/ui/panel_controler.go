package ui

import (
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
)

type PanelControler struct {
	navBar component.Renderer

	panels        []component.NamedPanel
	selectedPanel component.Renderer

	dialogs []component.Renderer
}

// NewPanelControler creates controller for all panels shown on screen.
//
// This controller should be created only after raylib window is initialized
// as panels could possibly use unitialized window parameters.
//
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func NewPanelControler(panelConfig ...component.BasePanelConfigFunc) component.Renderer {
	navBarHeight := style.TextLineSpacing
	panelConfig = append(
		panelConfig,
		func(pb *component.Panel) {
			pb.Y += navBarHeight
			pb.Height -= navBarHeight
		})

	controller := &PanelControler{
		panels: []component.NamedPanel{
			component.NewNamedPanel[*component.EditPanel]("Edit", panelConfig...),
			component.NewNamedPanel[*component.MenuPanel]("Menu", panelConfig...),
			component.NewNamedPanel[*component.ConsolePanel]("Console", panelConfig...),
			component.NewNamedPanel[*component.PlaceholderPanel]("Placeholder", panelConfig...),
		},
		dialogs: []component.Renderer{
			component.NewElement[*component.ExitDialog](),
		},
	}

	controller.navBar = component.NewPanel(
		func(np *component.NavigationPanel) { np.Height = navBarHeight },
		func(np *component.NavigationPanel) { np.SetParent(controller) },
	)

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

func (pc *PanelControler) SetSelectedPanel(panel component.Renderer) { pc.selectedPanel = panel }

func (pc *PanelControler) IteratePanels() iter.Seq2[string, component.Renderer] {
	return func(yield func(string, component.Renderer) bool) {
		for _, panel := range pc.panels {
			if !yield(panel.Title, panel.Renderer) {
				return
			}
		}
	}
}

func (pc *PanelControler) PanelCount() int { return len(pc.panels) }
