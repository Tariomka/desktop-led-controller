package ui

import (
	_ "embed"
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

//go:embed style.rgs
var style []byte

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
	IteratePanels() iter.Seq2[int, Panel]
	PanelCount() int
}

type PanelControler struct {
	navBar Panel

	panels        []Panel
	selectedPanel Panel

	dialogs []Element
}

func NewPanelControler(panelConfig ...PanelConfigFunc) Element {
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
			NewPanel[*PlaceholderPanel](append(panelConfig, shiftPosition, setTitle("Placeholder"))...),
			NewPanel[*MenuPanel](append(panelConfig, shiftPosition, setTitle("Menu"))...),
			NewPanel[*ConsolePanel](append(panelConfig, shiftPosition, setTitle("Console"))...),
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

func (pc *PanelControler) SetSelectedPanel(panel Panel) {
	pc.selectedPanel = panel
}

func (pc *PanelControler) IteratePanels() iter.Seq2[int, Panel] {
	return func(yield func(int, Panel) bool) {
		for index, panel := range pc.panels {
			if !yield(index, panel) {
				return
			}
		}
	}
}

func (pc *PanelControler) PanelCount() int {
	return len(pc.panels)
}

func (pc *PanelControler) setStyle() {
	// Base style
	raygui.LoadStyleFromMemory(style)

	// Updates, maybe set everything myself or create custom style seleretaly?
	raygui.SetStyle(0, raygui.BORDER_COLOR_FOCUSED, 0xff_00_00_7f)
	raygui.SetStyle(0, raygui.BASE_COLOR_FOCUSED, 0xff_00_00_2f)

	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED, 0xe0_3c_46_ff)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED, 0x5b_1e_20_ff)
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
	case *ConsolePanel:
		return (*ConsolePanel)(&base)
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
		raylib.SetExitKey(0)
		return &ExitDialog{}
	default:
		return nil
	}
}

type NavigationPanel struct {
	PanelBase
	parent       PanelSelector
	buttonWidth  float32
	buttonStates []bool
}

func (nav *NavigationPanel) SetParent(parent PanelSelector) {
	nav.parent = parent
	panelCount := nav.parent.PanelCount()
	nav.buttonWidth = nav.width / float32(panelCount)
	nav.buttonStates = make([]bool, panelCount)
}

func (nav *NavigationPanel) Update() {
	for index, panel := range nav.parent.IteratePanels() {
		// TODO: Add tooltip to buttons.
		// Note: raygui-go does not have tooltip bindings so probably need to implement by hand.
		if raygui.Button(
			raylib.NewRectangle(
				nav.pos.X+float32(index)*nav.buttonWidth,
				nav.pos.Y,
				nav.buttonWidth,
				nav.height),
			panel.Title()) {
			nav.parent.SetSelectedPanel(panel)

			for i := range nav.buttonStates {
				nav.buttonStates[i] = false
			}
			nav.buttonStates[index] = true
		}
	}
}

func (nav *NavigationPanel) Render() {
	for index, state := range nav.buttonStates {
		if state {
			rect := raylib.NewRectangle(
				nav.pos.X+float32(index)*nav.buttonWidth,
				nav.pos.Y,
				nav.buttonWidth,
				nav.height)

			raylib.DrawRectangleRec(
				rect,
				common.IntToRGBAWithAlpha(
					raygui.GetStyle(raygui.BUTTON, raygui.BASE_COLOR_PRESSED),
					77))
			raylib.DrawRectangleLinesEx(
				rect,
				float32(raygui.GetStyle(raygui.BUTTON, raygui.BORDER_WIDTH)),
				common.IntToRGBAWithAlpha(
					raygui.GetStyle(raygui.BUTTON, raygui.BORDER_COLOR_PRESSED),
					230))
		}
	}
}

func (nav *NavigationPanel) Title() string {
	return nav.title
}

type MenuPanel PanelBase

func (menu *MenuPanel) Update() {
}

func (menu *MenuPanel) Render() {
	raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), "")
	// raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), menu.title)
}

func (menu *MenuPanel) Title() string {
	return menu.title
}

type ConsolePanel PanelBase

func (cp *ConsolePanel) Update() {

}

func (cp *ConsolePanel) Render() {
	raygui.Panel(raylib.NewRectangle(cp.pos.X, cp.pos.Y, cp.width, cp.height), "")
	// raygui.Panel(raylib.NewRectangle(cp.pos.X, cp.pos.Y, cp.width, cp.height), cp.title)
}

func (cp *ConsolePanel) Title() string {
	return cp.title
}

type PlaceholderPanel PanelBase

func (pp *PlaceholderPanel) Update() {
}

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.pos.X, pp.pos.Y, pp.width, pp.height), "")
	// raygui.Panel(raylib.NewRectangle(pp.pos.X, pp.pos.Y, pp.width, pp.height), pp.title)
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
