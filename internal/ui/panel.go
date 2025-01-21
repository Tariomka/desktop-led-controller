package ui

import raylib "github.com/gen2brain/raylib-go/raylib"

type Panel interface {
	Update()
	Render()
}

type PanelBase struct {
	position      raylib.Vector2
	height, width float32
	title         string
}

type MenuPanel struct {
	PanelBase
}
