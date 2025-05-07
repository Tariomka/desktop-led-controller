package ui

import (
	"image/color"
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Cube struct {
	pos   raylib.Vector3
	color color.RGBA
}

type CubeGrid struct {
	cubes [][][]*Cube
	size  raylib.Vector3

	camera    *raylib.Camera
	screen    raylib.Rectangle
	ray       raylib.Ray
	collision raylib.RayCollision
}

func NewCubeGrid(xCount, yCount, zCount uint8, size raylib.Vector3, window raylib.Vector2) component.Renderer {
	sizeX := 1 + size.X
	sizeY := 1 + size.Y
	sizeZ := 1 + size.Z

	grid := make([][][]*Cube, zCount)
	for z := range grid {
		grid[z] = make([][]*Cube, yCount)
		for y := range grid[z] {
			grid[z][y] = make([]*Cube, xCount)
			for x := range grid[z][y] {
				grid[z][y][x] = &Cube{
					// this is not a mistake. 'y' and 'z' are switched
					// to keep the same perspecive as on the physical cube
					pos: raylib.NewVector3(
						sizeX*float32(x),
						sizeZ*float32(z),
						sizeY*float32(y),
					),
					color: common.ColorOff,
				}
			}
		}
	}

	cubeGrid := &CubeGrid{
		cubes: grid,
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
			window.X,
			window.Y),
	}
	return cubeGrid
}

func (this *CubeGrid) Update() {
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

	// for cube := range cg.IterateCubes() {
	// for cube := range this.IterateCubesExtended(3, -1, -1) {
	for cube := range this.IterateCubesStateful() {
		raylib.DrawCubeV(cube.pos, this.size, cube.color)
		raylib.DrawCubeWiresV(cube.pos, this.size, raylib.Black)
	}

	raylib.EndMode3D()
}

func (this *CubeGrid) IterateCubes() iter.Seq[*Cube] {
	return func(yield func(*Cube) bool) {
		for _, z := range this.cubes {
			for _, y := range z {
				for _, cube := range y {
					if !yield(cube) {
						return
					}
				}
			}
		}
	}
}

func (this *CubeGrid) IterateCubesExtended(row, column, layer int) iter.Seq[*Cube] {
	iterateZ := func() iter.Seq2[int, [][]*Cube] {
		return func(yield func(int, [][]*Cube) bool) {
			if layer > -1 && layer < len(this.cubes) {
				yield(layer, this.cubes[layer])
				return
			}
			for i, z := range this.cubes {
				if !yield(i, z) {
					return
				}
			}
		}
	}

	iterateY := func(zIndex int) iter.Seq2[int, []*Cube] {
		return func(yield func(int, []*Cube) bool) {
			if column > -1 && column < len(this.cubes[zIndex]) {
				yield(column, this.cubes[zIndex][column])
				return
			}
			for i, y := range this.cubes[zIndex] {
				if !yield(i, y) {
					return
				}
			}
		}
	}

	iterateX := func(zIndex, yIndex int) iter.Seq[*Cube] {
		return func(yield func(*Cube) bool) {
			if row > -1 && row < len(this.cubes[zIndex][yIndex]) {
				yield(this.cubes[zIndex][yIndex][row])
				return
			}
			for _, x := range this.cubes[zIndex][yIndex] {
				if !yield(x) {
					return
				}
			}
		}
	}

	return func(yield func(*Cube) bool) {
		for layerIndex, _ := range iterateZ() {
			for columnIndex, _ := range iterateY(layerIndex) {
				for cube := range iterateX(layerIndex, columnIndex) {
					if !yield(cube) {
						return
					}
				}
			}
		}
	}
}

func (this *CubeGrid) IterateCubesStateful() iter.Seq[*Cube] {
	xIndex, yIndex, zIndex := -1, -1, -1
	if global.SelectedLayerState == global.Layer || global.SelectedLayerState == global.Precise {
		zIndex = int(global.SelectedLayer)
	}
	if global.SelectedLayerState == global.Column || global.SelectedLayerState == global.Precise {
		yIndex = int(global.SelectedColumn)
	}
	return this.IterateCubesExtended(xIndex, yIndex, zIndex)
}

func (this *CubeGrid) updateCollision() {
	this.ray = raylib.GetScreenToWorldRay(raylib.GetMousePosition(), *this.camera)

	// TODO: add single slice iterating when slicing in editor panel is created
	for cube := range this.IterateCubesStateful() {
		// This hits multiple cubes, need to think on how to handle only a single collision
		this.collision = raylib.GetRayCollisionBox(
			this.ray,
			raylib.NewBoundingBox(
				raylib.NewVector3(cube.pos.X-this.size.X/2, cube.pos.Y-this.size.Y/2, cube.pos.Z-this.size.Z/2),
				raylib.NewVector3(cube.pos.X+this.size.X/2, cube.pos.Y+this.size.Y/2, cube.pos.Z+this.size.Z/2),
			))

		if this.collision.Hit {
			cube.color = global.SelectedColor
		}
	}
}
