package ui

import (
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type WindowConfigFunc func(*WindowConfig)

type WindowConfig struct {
	WindowWidth, WindowHeight int32
	CubeBaseSize, CubeHeight  uint8
}

func defaultConfig() WindowConfig {
	return WindowConfig{
		WindowWidth:  1280,
		WindowHeight: 720,
		CubeBaseSize: 8,
		CubeHeight:   8,
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
		width:  config.WindowWidth,
		height: config.WindowHeight,

		cubeGrid: NewCubeGrid(
			config.CubeBaseSize,
			config.CubeBaseSize,
			config.CubeHeight,
			raylib.NewVector3(1, 1, 1)),
	}
}

func (w *Window) Start() {
	raylib.SetConfigFlags(raylib.FlagWindowResizable)
	raylib.InitWindow(w.width, w.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)
	style.LoadStyle()

	// Try move to factory and use update update style seperately?
	w.hud = NewPanelControler()

	w.renderLoop()
}

func (w *Window) Stop() {
	raylib.CloseWindow()
}

func (w *Window) renderLoop() {
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
