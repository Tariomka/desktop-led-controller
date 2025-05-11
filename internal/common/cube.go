package common

import (
	"image/color"

	raylib "github.com/gen2brain/raylib-go/raylib"
)

type Cube struct {
	Pos   raylib.Vector3
	Color color.RGBA
}

type CubeLayout [][][]*Cube
