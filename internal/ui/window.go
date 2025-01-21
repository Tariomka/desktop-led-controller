package ui

import (
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type WindowConfig struct {
	windowWidth, windowHeight int32
	cubeBaseSize, cubeHeight  uint8
}

func NewConfig() WindowConfig {
	return WindowConfig{
		windowWidth:  1280,
		windowHeight: 720,
		cubeBaseSize: 8,
		cubeHeight:   8,
	}
}

type Window struct {
	width, height int32
	camera        raylib.Camera
	panels        []Panel
	cubeGrid      *CubeGrid
}

func NewWindow(config WindowConfig) *Window {
	if config.windowWidth == 0 {
		config.windowWidth = 1280
	}
	if config.windowHeight == 0 {
		config.windowHeight = 720
	}

	return &Window{
		width:  config.windowWidth,
		height: config.windowHeight,
		camera: raylib.Camera{
			Position:   raylib.NewVector3(30.0, 30.0, 30.0),
			Target:     raylib.NewVector3(10.0, 0.0, 0.0),
			Up:         raylib.NewVector3(0.0, 1.0, 0.0),
			Fovy:       55.0,
			Projection: raylib.CameraPerspective,
		},
		panels: make([]Panel, 0),
		cubeGrid: NewCubeGrid(
			config.cubeBaseSize,
			config.cubeBaseSize,
			config.cubeHeight,
			raylib.NewVector3(1, 1, 1)),
	}
}

func (w *Window) Start() {
	raylib.InitWindow(w.width, w.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)
	raygui.SetStyle(0, raygui.BACKGROUND_COLOR, 0x2d2d2dff)

	for !raylib.WindowShouldClose() {
		if raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
			raylib.UpdateCamera(&w.camera, raylib.CameraThirdPerson)
		}
		for _, panel := range w.panels {
			panel.Update()
		}

		raylib.BeginDrawing()
		raylib.ClearBackground(raylib.DarkGray)
		raylib.BeginMode3D(w.camera)

		w.renderCubes()
		// for _, panel := range w.panels {
		// 	panel.Render()
		// }

		raylib.EndMode3D()
		raylib.EndDrawing()
	}
}

func (w *Window) renderCubes() {
	for cube := range w.cubeGrid.IterateCubes() {
		raylib.DrawCubeV(cube.pos, w.cubeGrid.size, cube.color)
		raylib.DrawCubeWiresV(cube.pos, w.cubeGrid.size, raylib.Black)
	}
}

func (w *Window) Stop() {
	raylib.CloseWindow()
}
