package ui

import (
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type WindowConfigFunc func(*WindowConfig)

type WindowConfig struct {
	WindowWidth, WindowHeight int32
}

func defaultConfig() WindowConfig {
	return WindowConfig{
		WindowWidth:  1280,
		WindowHeight: 720,
	}
}

type Window struct {
	width, height int32

	hud      component.Renderer
	cubeGrid component.Renderer

	config WindowConfig
}

func NewWindow(configFuncs ...WindowConfigFunc) *Window {
	config := defaultConfig()
	for _, callback := range configFuncs {
		callback(&config)
	}

	return &Window{
		width:  config.WindowWidth,
		height: config.WindowHeight,
		config: config,
	}
}

func (this *Window) Start() {
	raylib.SetConfigFlags(raylib.FlagWindowResizable)
	raylib.InitWindow(this.width, this.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)

	style.LoadStyle()
	this.hud = NewPanelControler()
	this.cubeGrid = NewCubeGrid()
}

func (_ *Window) Stop() {
	raylib.CloseWindow()
}

// Render is a blocking infinite loop
func (this *Window) Render() {
	for !global.WindowShouldClose {
		global.WindowShouldClose = raylib.WindowShouldClose()

		// Update
		this.importDroppedFiles()
		this.cubeGrid.Update()
		this.hud.Update()

		// Render
		raylib.BeginDrawing()

		this.cubeGrid.Render()
		this.hud.Render()

		raylib.EndDrawing()
	}
}

// importDroppedFiles loads files dragged onto the window (.ls, .gc, .hex, .bin).
func (this *Window) importDroppedFiles() {
	if !raylib.IsFileDropped() {
		return
	}
	for _, path := range raylib.LoadDroppedFiles() {
		global.SendMessage(constants.ServiceShow, models.ImportMessage{Path: path})
	}
	raylib.UnloadDroppedFiles()
}
