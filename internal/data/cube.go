package data

import (
	"image/color"
	"iter"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type CubeIndex struct {
	X uint8 // Row
	Y uint8 // Column
	Z uint8 // Layer
}

type Cube struct {
	Pos   raylib.Vector3
	Color color.RGBA
}

type CubeFrame [][][]*Cube

func NewCubeFrame(xCount, yCount, zCount uint8, size raylib.Vector3) CubeFrame {
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
					Pos: raylib.NewVector3(
						sizeX*float32(x),
						sizeZ*float32(z),
						sizeY*float32(y)),
					Color: common.ColorOff,
				}
			}
		}
	}

	return grid
}

func (this CubeFrame) IterateCubes() iter.Seq[*Cube] {
	return func(yield func(*Cube) bool) {
		for _, z := range this {
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

func (this CubeFrame) IterateSingleOrAll(row, column, layer int) iter.Seq[*Cube] {
	return func(yield func(*Cube) bool) {
		for z := range common.IterateSingleOrAll(this, layer) {
			for y := range common.IterateSingleOrAll(z, column) {
				for cube := range common.IterateSingleOrAll(y, row) {
					if !yield(cube) {
						return
					}
				}
			}
		}
	}
}

func (this CubeFrame) IterateWithIndex() iter.Seq2[CubeIndex, *Cube] {
	return func(yield func(CubeIndex, *Cube) bool) {
		for layer, z := range this {
			for column, y := range z {
				for row, cube := range y {
					if !yield(CubeIndex{X: uint8(row), Y: uint8(column), Z: uint8(layer)}, cube) {
						return
					}
				}
			}
		}
	}
}

func (this CubeFrame) DeepClone() CubeFrame {
	buffer := make([][][]Cube, len(this))
	for zIndex, z := range this {
		buffer[zIndex] = make([][]Cube, len(this[zIndex]))
		for yIndex, y := range z {
			buffer[zIndex][yIndex] = make([]Cube, len(this[zIndex][yIndex]))
			for xIndex, cube := range y {
				buffer[zIndex][yIndex][xIndex] = *cube
			}
		}
	}

	clone := make(CubeFrame, len(buffer))
	for zIndex, z := range buffer {
		clone[zIndex] = make([][]*Cube, len(buffer[zIndex]))
		for yIndex, y := range z {
			clone[zIndex][yIndex] = make([]*Cube, len(buffer[zIndex][yIndex]))
			for xIndex, cube := range y {
				clone[zIndex][yIndex][xIndex] = &cube
			}
		}
	}

	return clone
}
