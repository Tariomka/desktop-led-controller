package ui

import (
	"image/color"
	"math"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	voxelSize    = 1
	voxelSpacing = 2
)

var (
	wireColor      = raylib.Fade(raylib.Black, 0.35)
	unlitWireColor = raylib.Fade(raylib.Black, 0.08)
)

// CubeGrid is the 3-D view: the running preview, or else the selected graph colored with the
// selected palette, which can be painted with the mouse.
type CubeGrid struct {
	camera raylib.Camera
	screen raylib.Rectangle
	start  time.Time

	length, height, width int // size of the shown volume, to re-center the camera
}

func NewCubeGrid() component.Renderer {
	return &CubeGrid{
		camera: raylib.Camera{
			Up:         raylib.NewVector3(0, 1, 0),
			Fovy:       45,
			Projection: raylib.CameraPerspective,
		},
		screen: screenRectangle(),
		start:  time.Now(),
	}
}

func screenRectangle() raylib.Rectangle {
	return raylib.NewRectangle(
		0, 0,
		float32(raylib.GetScreenWidth())*style.RendererWidthCoeficient,
		float32(raylib.GetScreenHeight()))
}

func (this *CubeGrid) Update() {
	if raylib.IsWindowResized() {
		this.screen = screenRectangle()
	}
	if global.Preview != nil {
		global.Preview.Update(uint32(time.Since(this.start).Milliseconds()))
	}

	length, height, width := this.volume()
	if length != this.length || height != this.height || width != this.width {
		this.length, this.height, this.width = length, height, width
		this.resetCamera()
	}

	inside := raylib.CheckCollisionPointRec(raylib.GetMousePosition(), this.screen)
	painting := global.ShouldPaint && global.Preview == nil
	if inside && painting && raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		this.paint()
	}
	if inside && (raylib.IsMouseButtonDown(raylib.MouseRightButton) ||
		!painting && raylib.IsMouseButtonDown(raylib.MouseLeftButton)) {
		raylib.UpdateCamera(&this.camera, raylib.CameraThirdPerson)
	}
}

func (this *CubeGrid) Render() {
	raylib.ClearBackground(raylib.DarkGray)
	raylib.BeginMode3D(this.camera)
	this.forEachVoxel(func(position raylib.Vector3, voxel *byte, voxelColor color.RGBA) {
		size := raylib.NewVector3(voxelSize, voxelSize, voxelSize)
		if voxelColor.A > 0 {
			raylib.DrawCubeV(position, size, voxelColor)
			raylib.DrawCubeWiresV(position, size, wireColor)
		} else {
			raylib.DrawCubeWiresV(position, size, unlitWireColor)
		}
	})
	raylib.EndMode3D()
}

// volume is the size of the shown voxel space.
func (this *CubeGrid) volume() (length, height, width int) {
	if global.Preview != nil {
		for _, display := range global.Preview.Canvas.Displays {
			length = max(length, display.X+display.L)
			height = max(height, display.Y+display.H)
			width = max(width, display.Z+display.W)
		}
		return length, height, width
	}
	if graph := selectedGraph(); graph != nil {
		return graph.L, graph.H, graph.W
	}
	return 0, 0, 0
}

func (this *CubeGrid) resetCamera() {
	center := voxelPosition(this.length-1, this.height-1, this.width-1)
	center = raylib.Vector3Scale(center, 0.5)
	distance := float32(max(this.length, this.height, this.width, 1)) * voxelSpacing * 1.6
	this.camera.Target = center
	// In front of the z = 0 face, a little to the right and above, like the vendor simulator.
	this.camera.Position = raylib.NewVector3(center.X+distance*0.6, center.Y+distance*0.6, center.Z+distance*1.4)
}

// forEachVoxel calls draw for every visible voxel; voxel is nil for preview voxels.
func (this *CubeGrid) forEachVoxel(draw func(position raylib.Vector3, voxel *byte, voxelColor color.RGBA)) {
	if global.Preview != nil {
		for i := range global.Preview.Canvas.Displays {
			display := &global.Preview.Canvas.Displays[i]
			forEachCell(display.L, display.H, display.W, func(x, y, z, index int) {
				draw(voxelPosition(display.X+x, display.Y+y, display.Z+z), nil, common.VoxelColor(display.Voxels[index]))
			})
		}
		return
	}

	graph := selectedGraph()
	if graph == nil {
		return
	}
	palette := render.DefaultPalette
	if global.SelectedPalette >= 0 && global.SelectedPalette < len(global.Project.Palettes) {
		palette = string(global.Project.Palettes[global.SelectedPalette].Colors[:])
	}
	forEachCell(graph.L, graph.H, graph.W, func(x, y, z, index int) {
		if !layerVisible(y, z) {
			return
		}
		voxel := &graph.Voxels[index]
		voxelColor := common.ColorOff
		if *voxel != 0 && int(*voxel) < len(palette) {
			voxelColor = common.VoxelColor(palette[*voxel])
		}
		draw(voxelPosition(x, y, z), voxel, voxelColor)
	})
}

// paint sets the nearest voxel under the mouse to the selected palette index.
func (this *CubeGrid) paint() {
	ray := raylib.GetScreenToWorldRay(raylib.GetMousePosition(), this.camera)
	var target *byte
	nearest := float32(math.MaxFloat32)
	half := raylib.NewVector3(voxelSize/2.0, voxelSize/2.0, voxelSize/2.0)
	this.forEachVoxel(func(position raylib.Vector3, voxel *byte, _ color.RGBA) {
		box := raylib.NewBoundingBox(raylib.Vector3Subtract(position, half), raylib.Vector3Add(position, half))
		if hit := raylib.GetRayCollisionBox(ray, box); hit.Hit && hit.Distance < nearest {
			nearest, target = hit.Distance, voxel
		}
	})
	if target != nil {
		*target = byte(global.SelectedColor)
	}
}

func selectedGraph() *blob.GraphData {
	graphs := global.Project.Graphs
	if global.SelectedGraph < 0 || global.SelectedGraph >= len(graphs) {
		return nil
	}
	return &graphs[global.SelectedGraph]
}

func layerVisible(y, z int) bool {
	switch global.SelectedLayerState {
	case global.Layer:
		return y == int(global.SelectedLayer)
	case global.Column:
		return z == int(global.SelectedColumn)
	case global.Precise:
		return y == int(global.SelectedLayer) && z == int(global.SelectedColumn)
	}
	return true
}

// forEachCell walks a volume in voxel index order x + y*L + z*L*H (y is up).
func forEachCell(length, height, width int, callback func(x, y, z, index int)) {
	index := 0
	for z := range width {
		for y := range height {
			for x := range length {
				callback(x, y, z, index)
				index++
			}
		}
	}
}

// voxelPosition maps show coordinates (y up, z = 0 is the front) to the 3-D view, where the
// camera looks at the front along -Z.
func voxelPosition(x, y, z int) raylib.Vector3 {
	return raylib.NewVector3(float32(x)*voxelSpacing, float32(y)*voxelSpacing, -float32(z)*voxelSpacing)
}
