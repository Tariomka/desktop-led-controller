package ui

import (
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type CubeGrid struct {
	frame data.CubeFrame
	size  raylib.Vector3

	camera    *raylib.Camera
	screen    raylib.Rectangle
	ray       raylib.Ray
	collision raylib.RayCollision

	channel chan any
}

func NewCubeGrid(xCount, yCount, zCount uint8, size raylib.Vector3) component.Renderer {
	cubeGrid := &CubeGrid{
		frame: data.NewCubeFrame(xCount, yCount, zCount, size),
		size:  size,
		camera: &raylib.Camera{
			Position:   raylib.NewVector3(30.0, 30.0, 30.0),
			Target:     raylib.NewVector3(10.0, 0.0, 0.0),
			Up:         raylib.NewVector3(0.0, 1.0, 0.0),
			Fovy:       float32(yCount)*float32(zCount) - float32(xCount),
			Projection: raylib.CameraPerspective,
		},
		screen: raylib.NewRectangle(
			0, 0,
			float32(raylib.GetScreenWidth())*style.RendererWidthCoeficient,
			float32(raylib.GetScreenHeight())),
		channel: make(chan any, 1),
	}

	go cubeGrid.channelLoop()
	global.RegisterMessageReceiver(
		constants.UICubeGrid,
		func(message any) { cubeGrid.channel <- message })

	return cubeGrid
}

func (this *CubeGrid) Update() {
	if raylib.IsWindowResized() {
		this.screen.Width = float32(raylib.GetScreenWidth()) * style.RendererWidthCoeficient
		this.screen.Height = float32(raylib.GetScreenHeight())
	}

	if global.ShouldChangeColor && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		this.updateCollision()
	}

	if raylib.IsMouseButtonDown(raylib.MouseLeftButton) &&
		raylib.CheckCollisionPointRec(raylib.GetMousePosition(), this.screen) {
		raylib.UpdateCamera(this.camera, raylib.CameraThirdPerson)
	}
}

func (this *CubeGrid) Render() {
	raylib.ClearBackground(raylib.DarkGray)
	raylib.BeginMode3D(*this.camera)

	for cube := range this.iterateCubesSelected() {
		raylib.DrawCubeV(cube.Pos, this.size, cube.Color)
		raylib.DrawCubeWiresV(cube.Pos, this.size, raylib.Black)
	}

	raylib.EndMode3D()
}

func (this *CubeGrid) iterateCubesSelected() iter.Seq[*data.Cube] {
	if global.SelectedLayerState == global.All {
		return this.frame.IterateCubes()
	}

	xIndex, yIndex, zIndex := -1, -1, -1
	if global.SelectedLayerState == global.Layer || global.SelectedLayerState == global.Precise {
		zIndex = int(global.SelectedLayer)
	}
	if global.SelectedLayerState == global.Column || global.SelectedLayerState == global.Precise {
		yIndex = int(global.SelectedColumn)
	}
	return this.frame.IterateSingleOrAll(xIndex, yIndex, zIndex)
}

func (this *CubeGrid) updateCollision() {
	this.ray = raylib.GetScreenToWorldRay(raylib.GetMousePosition(), *this.camera)

	for cube := range this.iterateCubesSelected() {
		// TODO: This hits multiple cubes, need to think on how to handle only a single collision
		this.collision = raylib.GetRayCollisionBox(
			this.ray,
			raylib.NewBoundingBox(
				raylib.NewVector3(
					cube.Pos.X-this.size.X/2,
					cube.Pos.Y-this.size.Y/2,
					cube.Pos.Z-this.size.Z/2),
				raylib.NewVector3(
					cube.Pos.X+this.size.X/2,
					cube.Pos.Y+this.size.Y/2,
					cube.Pos.Z+this.size.Z/2)))

		if this.collision.Hit {
			cube.Color = global.SelectedColor
		}
	}
}

func (this *CubeGrid) fillInVisibleCubes() {
	for cube := range this.iterateCubesSelected() {
		cube.Color = global.SelectedColor
	}
}

func (this *CubeGrid) resetCubes() {
	for cube := range this.frame.IterateCubes() {
		cube.Color = common.ColorOff
	}
}

func (this *CubeGrid) overwriteCubes(frame data.CubeFrame) {
	this.frame = frame.DeepClone()
}

// Blocking message loop
func (this *CubeGrid) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.ResetMessage:
			this.resetCubes()
		case models.SaveMessage:
			global.SendMessage(
				constants.ServiceLedProcessor,
				models.AddToBufferMessage{Frame: this.frame.DeepClone()})
		case models.FillVisibleCubesMessage:
			this.fillInVisibleCubes()
		case models.SetFrameMessage:
			this.overwriteCubes(message.Frame)
		}
	}
}
