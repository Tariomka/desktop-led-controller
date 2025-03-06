package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/common"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type PanelConfigFunc func(*PanelBase)

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
		return &MenuPanel{
			PanelBase: base,
			padding:   raylib.NewVector2(10, 10),
		}
	case *ConsolePanel:
		return &ConsolePanel{
			PanelBase:   base,
			messages:    common.NewRingArray[string](100),
			itemFocused: -1,
		}
	case *PlaceholderPanel:
		return &PlaceholderPanel{PanelBase: base}
	default:
		panic("wrong renderer type")
	}
}

func NewElement[T Renderer]() Renderer {
	var placeholder T
	switch any(placeholder).(type) {
	case *ExitDialog:
		raylib.SetExitKey(0)
		return &ExitDialog{
			width:  250,
			height: 100,
		}
	case *MessageListView:
		return &MessageListView{}
	default:
		panic("wrong renderer type")
	}
}
