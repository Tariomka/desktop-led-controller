package ui

import (
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Panel interface {
	Update()
	Render()
}

type PanelType interface {
	PanelBase | MenuPanel
}

// type PanelType interface {
// 	PanelBase
// 	Panel
// }

type PanelBase struct {
	pos           raylib.Vector2
	width, height float32
	title         string
}

func NewPanel[T PanelType](position raylib.Vector2, width, height float32) *T {
	return &T{
		pos:    position,
		height: height,
		width:  width,
	}
}

type MenuPanel PanelBase

func (menu *MenuPanel) Update() {
}

func (menu *MenuPanel) Render() {
	raygui.Panel(raylib.NewRectangle(menu.pos.X, menu.pos.Y, menu.width, menu.height), menu.title)
}
