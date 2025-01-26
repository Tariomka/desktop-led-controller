package ui

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type WindowConfigFunc func(*WindowConfig)

type WindowConfig struct {
	windowWidth, windowHeight int32
	cubeBaseSize, cubeHeight  uint8
}

func defaultConfig() WindowConfig {
	return WindowConfig{
		windowWidth:  1280,
		windowHeight: 720,
		cubeBaseSize: 8,
		cubeHeight:   8,
	}
}

type Window struct {
	width, height int32

	hud      component.Renderer
	cubeGrid component.Renderer
}

func NewWindow(configFuncs ...WindowConfigFunc) *Window {
	config := defaultConfig()
	for _, callback := range configFuncs {
		callback(&config)
	}

	return &Window{
		width:  config.windowWidth,
		height: config.windowHeight,

		cubeGrid: component.NewCubeGrid(
			config.cubeBaseSize,
			config.cubeBaseSize,
			config.cubeHeight,
			raylib.NewVector3(1, 1, 1)),
	}
}

func (w *Window) Start() {
	raylib.SetConfigFlags(raylib.FlagWindowResizable)
	raylib.InitWindow(w.width, w.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)
	w.hud = component.NewPanelControler()

	for !global.WindowShouldClose {
		global.WindowShouldClose = raylib.WindowShouldClose()

		w.cubeGrid.Update()
		w.hud.Update()

		raylib.BeginDrawing()

		w.cubeGrid.Render()
		w.hud.Render()

		raylib.EndDrawing()
	}
}

func (w *Window) Stop() {
	raylib.CloseWindow()
}
