package component

import (
	"image/color"
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
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

	camera *raylib.Camera

	ray       raylib.Ray
	collision raylib.RayCollision
}

func NewCubeGrid(xCount, yCount, zCount uint8, size raylib.Vector3) Renderer {
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

	// for debugging purposes
	// Delete this block when done
	grid[7][0][1].color = common.ColorRed
	grid[7][0][3].color = common.ColorGreen
	grid[7][0][6].color = common.ColorBlue
	grid[7][4][2].color = common.ColorCyan
	grid[7][4][4].color = common.ColorYellow
	grid[7][4][6].color = common.ColorViolet
	grid[7][7][5].color = common.ColorWhite

	return &CubeGrid{
		cubes: grid,
		size:  size,
		camera: &raylib.Camera{
			Position:   raylib.NewVector3(30.0, 30.0, 30.0),
			Target:     raylib.NewVector3(10.0, 0.0, 0.0),
			Up:         raylib.NewVector3(0.0, 1.0, 0.0),
			Fovy:       55.0,
			Projection: raylib.CameraPerspective,
		},
	}
}

func (cg *CubeGrid) Update() {
	if global.ShouldChangeColor && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		cg.updateCollision()
	}
	if raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		raylib.UpdateCamera(cg.camera, raylib.CameraThirdPerson)
	}

}

func (cg *CubeGrid) Render() {
	raylib.ClearBackground(raylib.DarkGray)
	raylib.BeginMode3D(*cg.camera)

	for cube := range cg.IterateCubes() {
		raylib.DrawCubeV(cube.pos, cg.size, cube.color)
		raylib.DrawCubeWiresV(cube.pos, cg.size, raylib.Black)
	}

	raylib.EndMode3D()
}

func (cg *CubeGrid) IterateCubes() iter.Seq[*Cube] {
	return func(yield func(*Cube) bool) {
		for _, z := range cg.cubes {
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

func (cg *CubeGrid) updateCollision() {
	cg.ray = raylib.GetScreenToWorldRay(raylib.GetMousePosition(), *cg.camera)

	// TODO: add single slice iterating when slicing in editor panel is created
	for cube := range cg.IterateCubes() {
		// This hits multiple cubes, need to think on how to handle only a single collision
		cg.collision = raylib.GetRayCollisionBox(
			cg.ray,
			raylib.NewBoundingBox(
				raylib.NewVector3(cube.pos.X-cg.size.X/2, cube.pos.Y-cg.size.Y/2, cube.pos.Z-cg.size.Z/2),
				raylib.NewVector3(cube.pos.X+cg.size.X/2, cube.pos.Y+cg.size.Y/2, cube.pos.Z+cg.size.Z/2),
			))

		if cg.collision.Hit {
			cube.color = global.SelectedColor
		}
	}
}
