package ui

import (
	"image/color"
	"iter"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Cube struct {
	pos   raylib.Vector3
	color color.RGBA
}

type CubeGrid struct {
	cubes [][][]*Cube
	size  raylib.Vector3
}

func NewCubeGrid(xCount, yCount, zCount uint8, size raylib.Vector3) *CubeGrid {
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
					color: raylib.Blank,
				}
			}
		}
	}

	// for debugging purposes
	// Delete this block when done
	grid[7][0][1].color = raylib.Red
	grid[7][0][6].color = raylib.Blue

	return &CubeGrid{
		cubes: grid,
		size:  size,
	}
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
