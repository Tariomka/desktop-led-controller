package ui

import (
	"iter"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Element interface {
	Update()
	Render()
}

type Panel interface {
	Element
	Title() string
}

type PanelSelector interface {
	SetSelectedPanel(panel Panel)
	IteratePanels() iter.Seq[Panel]
}

type PanelControler struct {
	navBar Panel

	panels        []Panel
	selectedPanel Panel

	dialogs []Element
}

func NewPanelControler(panelConfig ...PanelConfigFunc) Element {
	navBarHeight := float32(20)
	navBarPosition := func(pb *PanelBase) { pb.height = navBarHeight }
	shiftPosition := func(pb *PanelBase) { pb.pos.Y += navBarHeight }
	setTitle := func(name string) PanelConfigFunc {
		return func(pb *PanelBase) { pb.title = name }
	}

	controller := &PanelControler{
		navBar: NewPanel[*NavigationPanel](append(panelConfig, navBarPosition)...),
		panels: []Panel{
			NewPanel[*PlaceholderPanel](append(panelConfig, shiftPosition, setTitle("Placeholder"))...),
			NewPanel[*MenuPanel](append(panelConfig, shiftPosition, setTitle("Menu"))...),
		},
		dialogs: []Element{
			NewElement[*ExitDialog](),
		},
	}

	controller.selectedPanel = controller.panels[0]

	controller.navBar.(*NavigationPanel).parent = controller
	controller.navBar.(*NavigationPanel).buttonWidth =
		controller.navBar.(*NavigationPanel).width / float32(len(controller.panels))

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

func (pc *PanelControler) Title() string {
	return ""
}

func (pc *PanelControler) SetSelectedPanel(panel Panel) {
	pc.selectedPanel = panel
}

func (pc *PanelControler) IteratePanels() iter.Seq[Panel] {
	return func(yield func(Panel) bool) {
		for _, panel := range pc.panels {
			if !yield(panel) {
				return
			}
		}
	}
}

type PanelConfigFunc func(*PanelBase)

type PanelBase struct {
	pos           raylib.Vector2
	width, height float32
	title         string
}

func defaultPanelBase() PanelBase {
	panelWidth := float32(raylib.GetScreenWidth()) / 4
	return PanelBase{
		pos:    raylib.NewVector2(float32(raylib.GetScreenWidth())-panelWidth, 0),
		width:  panelWidth,
		height: float32(raylib.GetScreenHeight()),
	}
}

func NewPanel[T Panel](panelConfig ...PanelConfigFunc) Panel {
	base := defaultPanelBase()
	for _, config := range panelConfig {
		config(&base)
	}

	var placeholder T
	switch any(placeholder).(type) {
	case *NavigationPanel:
		return &NavigationPanel{
			PanelBase: base,
			parent:    nil,
		}
	case *MenuPanel:
		return (*MenuPanel)(&base)
	case *PlaceholderPanel:
		return (*PlaceholderPanel)(&base)
	default:
		return nil
	}
}

func NewElement[T Element]() Element {
	var placeholder T
	switch any(placeholder).(type) {
	case *ExitDialog:
		return &ExitDialog{}
	default:
		return nil
	}
}

type NavigationPanel struct {
	PanelBase
	parent      PanelSelector
	buttonWidth float32
	buttonIndex int
}

func (nav *NavigationPanel) Update() {
}

func (nav *NavigationPanel) Render() {
	nav.buttonIndex = 0

	for panel := range nav.parent.IteratePanels() {
		if raygui.Button(
			raylib.NewRectangle(
				nav.pos.X+float32(nav.buttonIndex)*nav.buttonWidth,
				nav.pos.Y,
				nav.buttonWidth,
				nav.height),
			panel.Title()) {
			nav.parent.SetSelectedPanel(panel)
		}

		nav.buttonIndex++
	}
}

func (nav *NavigationPanel) Title() string {
	return nav.title
}

type MenuPanel PanelBase

func (menu *MenuPanel) Update() {
}

func (menu *MenuPanel) Render() {
	raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), menu.title)
}
func (menu *MenuPanel) Title() string {
	return menu.title
}

type PlaceholderPanel PanelBase

func (pp *PlaceholderPanel) Update() {
}

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.pos.X, pp.pos.Y, pp.width, pp.height), pp.title)
}

func (pp *PlaceholderPanel) Title() string {
	return pp.title
}

type ExitDialog struct {
	show bool
}

func (ed *ExitDialog) Update() {
	if raylib.IsKeyPressed(raylib.KeyEscape) {
		ed.show = !ed.show
	}
}

func (ed *ExitDialog) Render() {
	if ed.show {
		raylib.DrawRectangle(
			0,
			0,
			int32(raylib.GetScreenWidth()),
			int32(raylib.GetScreenHeight()),
			raylib.Fade(raylib.Black, 0.7))

		result := raygui.MessageBox(
			raylib.Rectangle{
				X:      float32(raylib.GetScreenWidth())/2 - 125,
				Y:      float32(raylib.GetScreenHeight())/2 - 50,
				Width:  250,
				Height: 100,
			},
			raygui.IconText(raygui.ICON_EXIT, "Close Window"),
			"Do you really want to exit?",
			"Yes;No")

		if raylib.IsKeyPressed(raylib.KeyEnter) {
			result = 1
		}

		if (result == 0) || (result == 2) {
			ed.show = false
		} else if result == 1 {
			windowShouldClose = true
		}
	}
}
