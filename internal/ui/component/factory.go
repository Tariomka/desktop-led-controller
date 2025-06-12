package component

import (
	"reflect"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type BasePanelConfigFunc func(*Panel)

type PanelConfigFunc[Type Renderer] func(Type)

func NewPanel[Type Renderer](panelConfig ...PanelConfigFunc[Type]) Renderer {
	base := defaultTypedPanel[Type]()
	for _, config := range panelConfig {
		config(base)
	}

	switch panel := any(base).(type) {
	case *NavigationPanel:
		if panel.parent == nil {
			panel.buttonStates = make([]bool, 0)
		}
		return panel
	case *EditPanel:
		return newEditPanel(panel.Panel)
	case *MenuPanel:
		return newMenuPanel(panel.Panel)
	case *ConsolePanel:
		return newConsolePanel(panel.Panel)
	default:
		return base
	}
}

type NamedPanel struct {
	Renderer
	Title string
}

func NewNamedPanel[Type Renderer](title string, panelConfig ...BasePanelConfigFunc) NamedPanel {
	return NamedPanel{
		Renderer: NewPanelBase[Type](panelConfig...),
		Title:    title,
	}
}

func NewPanelBase[Type Renderer](panelConfig ...BasePanelConfigFunc) Renderer {
	return NewPanel(adaptPanelConfig[Type](panelConfig)...)
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
	default:
		panic("wrong renderer type")
	}
}

func adaptPanelConfig[Type Renderer](configs []BasePanelConfigFunc) []PanelConfigFunc[Type] {
	adapted := make([]PanelConfigFunc[Type], len(configs))
	for i, config := range configs {
		adapted[i] = func(panel Type) {
			rectangleField := reflect.ValueOf(panel).Elem().FieldByName("Rectangle")
			if rectangleField.IsValid() && rectangleField.CanSet() {
				temp := Panel{Rectangle: rectangleField.Interface().(raylib.Rectangle)}
				config(&temp)
				rectangleField.Set(reflect.ValueOf(temp.Rectangle))
			}
		}
	}
	return adapted
}

func defaultTypedPanel[Type Renderer]() Type {
	rendererType := reflect.TypeOf((*Type)(nil)).Elem()
	if rendererType.Kind() == reflect.Ptr {
		rendererType = rendererType.Elem()
	}
	renderer := reflect.New(rendererType)
	rectangleField := renderer.Elem().FieldByName("Rectangle")
	if rectangleField.IsValid() && rectangleField.CanSet() {
		rectangleField.Set(reflect.ValueOf(defaultRectangle()))
	}
	return renderer.Interface().(Type)
}

// Default calculates properties with respect to window size, which only works after raylib init is called.
// See:
//
//	import rl "github.com/gen2brain/raylib-go/raylib"
//	rl.InitWindow(...)
func defaultRectangle() raylib.Rectangle {
	panelWidth := float32(raylib.GetScreenWidth()) * style.PanelWidthCoeficient
	return raylib.NewRectangle(
		float32(raylib.GetScreenWidth())-panelWidth,
		0,
		panelWidth,
		float32(raylib.GetScreenHeight()))
}
