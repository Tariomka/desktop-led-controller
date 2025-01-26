package ui

import (
	_ "embed"
	"fmt"
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

//go:embed panel_style.rgs
var style []byte

type Renderer interface {
	Update()
	Render()
}

type Element interface {
	Renderer
}

type Panel interface {
	Renderer
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

type PanelConfigFunc func(*PanelBase)

type PanelBase struct {
	pos           raylib.Vector2
	width, height float32
	title         string
}

// Default calculates properties with respect to window size, which only works after raylib init is called.
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func defaultPanelBase() PanelBase {
	panelWidth := float32(raylib.GetScreenWidth()) / 4
	return PanelBase{
		pos:    raylib.NewVector2(float32(raylib.GetScreenWidth())-panelWidth, 0),
		width:  panelWidth,
		height: float32(raylib.GetScreenHeight()),
	}
}

func (pb *PanelBase) resize(additionalResizes ...func()) {
	if raylib.IsWindowResized() {
		previousPosition := pb.pos
		panelWidth := float32(raylib.GetScreenWidth()) / 4

		if previousPosition.X > 0 {
			pb.width = panelWidth
			pb.pos.X = float32(raylib.GetScreenWidth()) - panelWidth
		}
		if previousPosition.Y > 0 {
			pb.height = float32(raylib.GetScreenHeight()) - previousPosition.Y
		}

		for _, callback := range additionalResizes {
			callback()
		}
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
			PanelBase:    base,
			parent:       nil,
			buttonStates: make([]bool, 0),
		}
	case *EditPanel:
		return &EditPanel{PanelBase: base}
	case *MenuPanel:
		return &MenuPanel{PanelBase: base}
	case *ConsolePanel:
		text := "Placeholder"
		return &ConsolePanel{
			PanelBase: base,
			output:    &text,
		}
	case *PlaceholderPanel:
		return &PlaceholderPanel{PanelBase: base}
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
	if panelCount > 0 {
		nav.buttonStates[0] = true
	}
}

func (nav *NavigationPanel) Update() {
	nav.resize(func() { nav.buttonWidth = nav.width / float32(len(nav.buttonStates)) })
}

func (nav *NavigationPanel) Render() {
	for index, panel := range nav.parent.IteratePanels() {
		// TODO: Add tooltip to buttons.
		// Note: raygui-go does not have tooltip bindings so probably need to implement by hand.
		bounds := raylib.NewRectangle(
			nav.pos.X+float32(index)*nav.buttonWidth,
			nav.pos.Y,
			nav.buttonWidth,
			nav.height)
		if !raygui.Button(bounds, panel.Title()) {
			if nav.buttonStates[index] {
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

			continue
		}

		nav.parent.SetSelectedPanel(panel)

		for i := range nav.buttonStates {
			nav.buttonStates[i] = false
		}
		nav.buttonStates[index] = true
	}
}

func (nav *NavigationPanel) Title() string { return nav.title }

type MenuPanel struct{ PanelBase }

func (menu *MenuPanel) Update() { menu.resize() }

func (menu *MenuPanel) Render() {
	raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), "")
}

func (menu *MenuPanel) Title() string { return menu.title }

type ConsolePanel struct {
	PanelBase
	output *string
	test   int
}

func (cp *ConsolePanel) Update() {
	cp.resize()
	cp.test++
	if cp.test%50 == 0 {
		text := *cp.output + fmt.Sprintf("\nMessage #%d", cp.test)
		cp.output = &text
	}
}

func (cp *ConsolePanel) Render() {
	bounds := raylib.NewRectangle(cp.pos.X, cp.pos.Y, cp.width, cp.height)
	content := raylib.NewRectangle(cp.pos.X+10, cp.pos.Y+10, cp.width-20, cp.height-20)
	// raygui.Panel(bounds, "")
	scroll := raylib.NewVector2(10, 10)
	view := raylib.NewRectangle(cp.pos.X+20, cp.pos.Y+20, cp.width-40, cp.height-40)

	raygui.ScrollPanel(
		bounds,
		"",
		content,
		&scroll,
		&view,
	)
	// raygui.ListView(
	// 	raylib.NewRectangle(cp.pos.X+10, cp.pos.Y+10, cp.width-20, cp.height-20),
	// 	*cp.output,
	// 	nil,
	// 	-1,
	// )
	raygui.TextBox(
		view,
		cp.output,
		10,
		false,
	)
}

func (cp *ConsolePanel) Title() string { return cp.title }

type EditPanel struct{ PanelBase }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	raygui.Panel(raylib.NewRectangle(ep.pos.X, ep.pos.Y, ep.width, ep.height), "")
}

func (ep *EditPanel) Title() string { return ep.title }

type PlaceholderPanel struct{ PanelBase }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.pos.X, pp.pos.Y, pp.width, pp.height), "")
}

func (pp *PlaceholderPanel) Title() string { return pp.title }

type ExitDialog struct{ show bool }

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
