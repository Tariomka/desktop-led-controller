package component

import raylib "github.com/gen2brain/raylib-go/raylib"

type PanelConfigFunc func(*PanelBase)

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

func NewElement[T Renderer]() Renderer {
	var placeholder T
	switch any(placeholder).(type) {
	case *ExitDialog:
		raylib.SetExitKey(0)
		return &ExitDialog{}
	case *MessageListView:
		return &MessageListView{}
	default:
		panic("wrong renderer type")
	}
}

type GenericPanelConfigFunc[T Renderer] func(Renderer)

func test[T Renderer](configs ...GenericPanelConfigFunc[T]) []GenericPanelConfigFunc[T] {
	return append(
		configs,
		func(r Renderer) {
			r.(*ConsolePanel).maxMessageCount = 100
		})
}
