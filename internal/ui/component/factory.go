package component

import raylib "github.com/gen2brain/raylib-go/raylib"

type PanelConfigFunc func(*Panel)

type TypedPanelConfigFunc[Type Renderer] func(*Type) // TODO: try integrating this, or remove it

type NamedPanel struct {
	Renderer
	Title string
}

func NewNamedPanel[Type Renderer](title string, panelConfig ...PanelConfigFunc) NamedPanel {
	return NamedPanel{
		Renderer: NewPanel[Type](panelConfig...),
		Title:    title,
	}
}

func NewPanel[Type Renderer](panelConfig ...PanelConfigFunc) Renderer {
	base := defaultPanel()
	for _, config := range panelConfig {
		config(&base)
	}

	var placeholder Type
	switch any(placeholder).(type) {
	case *NavigationPanel:
		return &NavigationPanel{
			Panel:        base,
			parent:       nil,
			buttonStates: make([]bool, 0),
		}
	case *EditPanel:
		return newEditPanel(base)
	case *MenuPanel:
		return newMenuPanel(base)
	case *ConsolePanel:
		return newConsolePanel(base)
	case *PlaceholderPanel:
		return &PlaceholderPanel{Panel: base}
	default:
		panic("wrong renderer type")
	}
}

func NewElement[Type Renderer]() Renderer {
	var placeholder Type
	switch any(placeholder).(type) {
	case *ExitDialog:
		raylib.SetExitKey(0)
		return &ExitDialog{
			width:  250,
			height: 100,
		}
	// case *MessageListView:
	// 	return &MessageListView{}
	default:
		panic("wrong renderer type")
	}
}
