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
			raylib.NewVector3(1, 1, 1),
			raylib.NewVector2(
				float32(config.WindowWidth*3/4), // TODO: remove hardcode? defaultPanel() has the final part
				float32(config.WindowHeight))),
	}
}

func (this *Window) Start() {
	raylib.SetConfigFlags(raylib.FlagWindowResizable)
	raylib.InitWindow(this.width, this.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)

	style.LoadStyle()
	this.hud = NewPanelControler()
}

func (_ *Window) Stop() {
	raylib.CloseWindow()
}

// Rendering method is a blocking infinite loop
func (this *Window) Render() {
	for !global.WindowShouldClose {
		global.WindowShouldClose = raylib.WindowShouldClose()

		// Update
		this.cubeGrid.Update()
		this.hud.Update()

		// Render
		raylib.BeginDrawing()

		this.cubeGrid.Render()
		this.hud.Render()

		raylib.EndDrawing()
	}
}
