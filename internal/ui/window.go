package ui

import (
	"image/color"

	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

var (
	windowShouldClose bool
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

	camera *raylib.Camera
	hud    Element

	cubeGrid      *CubeGrid
	selectedColor color.RGBA

	ray       raylib.Ray
	collision raylib.RayCollision
}

func NewWindow(configFuncs ...WindowConfigFunc) *Window {
	config := defaultConfig()
	for _, callback := range configFuncs {
		callback(&config)
	}

	return &Window{
		width:  config.windowWidth,
		height: config.windowHeight,
		camera: &raylib.Camera{
			Position:   raylib.NewVector3(30.0, 30.0, 30.0),
			Target:     raylib.NewVector3(10.0, 0.0, 0.0),
			Up:         raylib.NewVector3(0.0, 1.0, 0.0),
			Fovy:       55.0,
			Projection: raylib.CameraPerspective,
		},
		cubeGrid: NewCubeGrid(
			config.cubeBaseSize,
			config.cubeBaseSize,
			config.cubeHeight,
			raylib.NewVector3(1, 1, 1)),
		selectedColor: raylib.Gray,
	}
}

func (w *Window) Start() {
	raylib.InitWindow(w.width, w.height, "Led Cube Controller")
	raylib.SetTargetFPS(60)
	w.hud = NewPanelControler()
	raygui.SetStyle(0, raygui.BACKGROUND_COLOR, 0x2d2d2dff)
	raylib.SetExitKey(0)

	for !windowShouldClose {
		windowShouldClose = raylib.WindowShouldClose()
		w.updateCamera()
		w.updateCubes()
		w.hud.Update()

		raylib.BeginDrawing()

		w.render3D()
		w.hud.Render()

		raylib.EndDrawing()
	}
}

func (w *Window) Stop() {
	raylib.CloseWindow()
}

func (w *Window) updateCamera() {
	if raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		raylib.UpdateCamera(w.camera, raylib.CameraThirdPerson)
	}
}

func (w *Window) updateCubes() {
	if raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		w.ray = raylib.GetScreenToWorldRay(raylib.GetMousePosition(), *w.camera)

		for cube := range w.cubeGrid.IterateCubes() {
			w.collision = raylib.GetRayCollisionBox(
				w.ray,
				raylib.NewBoundingBox(
					raylib.NewVector3(
						cube.pos.X-w.cubeGrid.size.X/2,
						cube.pos.Y-w.cubeGrid.size.Y/2,
						cube.pos.Z-w.cubeGrid.size.Z/2),
					raylib.NewVector3(
						cube.pos.X+w.cubeGrid.size.X/2,
						cube.pos.Y+w.cubeGrid.size.Y/2,
						cube.pos.Z+w.cubeGrid.size.Z/2),
				))

			if w.collision.Hit {
				cube.color = w.selectedColor
			}
		}
	}
}

func (w *Window) render3D() {
	raylib.ClearBackground(raylib.DarkGray)
	raylib.BeginMode3D(*w.camera)

	for cube := range w.cubeGrid.IterateCubes() {
		raylib.DrawCubeV(cube.pos, w.cubeGrid.size, cube.color)
		raylib.DrawCubeWiresV(cube.pos, w.cubeGrid.size, raylib.Black)
	}

	raylib.EndMode3D()
}
