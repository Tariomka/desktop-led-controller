package component

import (
	"fmt"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type PanelConfigFunc func(*PanelBase)

type PanelBase struct {
	pos           raylib.Vector2
	width, height float32
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

func NewPanel[T Renderer](panelConfig ...PanelConfigFunc) Renderer {
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
		return &ConsolePanel{
			PanelBase:       base,
			messages:        make([]string, 0),
			maxMessageCount: 100,
		}
	case *PlaceholderPanel:
		return &PlaceholderPanel{PanelBase: base}
	default:
		panic("wrong renderer type")
	}
}

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
	nav.buttonWidth = nav.width / float32(panelCount)
	nav.buttonStates = make([]bool, panelCount)
}

func (nav *NavigationPanel) Update() {
	nav.resize(func() { nav.buttonWidth = nav.width / float32(len(nav.buttonStates)) })
}

func (nav *NavigationPanel) Render() {
	nav.index = 0
	for name, panel := range nav.parent.IteratePanels() {
		// TODO: Add tooltip to buttons.
		// Note: raygui-go does not have tooltip bindings so probably need to implement by hand.
		bounds := raylib.NewRectangle(
			nav.pos.X+float32(nav.index)*nav.buttonWidth,
			nav.pos.Y,
			nav.buttonWidth,
			nav.height)

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
	raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), "")
}

type ConsolePanel struct {
	PanelBase
	messages        []string
	maxMessageCount uint16
	test            int
	focus           int32
	scroll          int32
}

func (cp *ConsolePanel) Update() {
	cp.resize()
	cp.test++
	if cp.test%50 == 0 {
		cp.messages = append([]string{fmt.Sprintf("Message #%d: some ; aaaaaaaaaaaaaaaa", cp.test)}, cp.messages...)
	}
	if len(cp.messages) > int(cp.maxMessageCount) {
		cp.messages = cp.messages[:cp.maxMessageCount]
	}
}

func (cp *ConsolePanel) Render() {
	bounds := raylib.NewRectangle(cp.pos.X, cp.pos.Y, cp.width, cp.height)
	content := raylib.NewRectangle(cp.pos.X+10, cp.pos.Y+10, cp.width-20, cp.height-20)

	raygui.Panel(bounds, "")
	raygui.ListView(content, cp.getLineralizedText(), &cp.scroll, -1)
}

func (cp *ConsolePanel) getLineralizedText() string {
	return strings.Join(cp.messages, "\n")
}

type EditPanel struct{ PanelBase }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	raygui.Panel(raylib.NewRectangle(ep.pos.X, ep.pos.Y, ep.width, ep.height), "")
}

type PlaceholderPanel struct{ PanelBase }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() {
	raygui.Panel(raylib.NewRectangle(pp.pos.X, pp.pos.Y, pp.width, pp.height), "")
}

type NamedPanel struct {
	Renderer
	Title string
}

func NewNamedPanel[T Renderer](title string, panelConfig ...PanelConfigFunc) NamedPanel {
	return NamedPanel{
		Renderer: NewPanel[T](panelConfig...),
		Title:    title,
	}
}
