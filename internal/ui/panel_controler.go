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

// Creates controller for all panels shown on screen.
//
// This controller should be created only after raylib window is initialized
// as panels could possibly use unitialized window parameters.
//
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func NewPanelControler(panelConfig ...component.PanelConfigFunc) component.Renderer {
	navBarHeight := style.TextLineSpacing
	navBarConfig := append(panelConfig, func(pb *component.Panel) { pb.Height = navBarHeight })
	panelConfig = append(
		panelConfig,
		func(pb *component.Panel) {
			pb.Y += navBarHeight
			pb.Height -= navBarHeight
		})

	controller := &PanelControler{
		navBar: component.NewPanel[*component.NavigationPanel](navBarConfig...),
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

	// TODO: Move to customizations
	controller.navBar.(*component.NavigationPanel).SetParent(controller)
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
