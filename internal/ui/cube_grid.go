package ui

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
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
	hoverColor     = raylib.Fade(raylib.White, 0.5)
	viewBackground = style.View
)

// CubeGrid is the 3-D view: the running preview, or else the selected graph colored with the
// selected palette, which can be painted with the mouse.
type CubeGrid struct {
	camera raylib.Camera
	bounds raylib.Rectangle
	start  time.Time

	target raylib.RenderTexture2D // the view is rendered off-screen so the camera fits its rectangle
	scale  float32                // DPI scale of the render texture

	length, height, width int // size of the shown volume, to re-center the camera
	distance              float32
	hover                 *byte
	painting              bool
}

func NewCubeGrid() *CubeGrid {
	return &CubeGrid{
		camera: raylib.Camera{
			Up:         raylib.NewVector3(0, 1, 0),
			Fovy:       45,
			Projection: raylib.CameraPerspective,
		},
		start: time.Now(),
	}
}

// Update handles the preview clock and the mouse; mouseFree is false while a dialog or a
// splitter owns the mouse.
func (this *CubeGrid) Update(bounds raylib.Rectangle, mouseFree bool) {
	this.bounds = bounds
	this.ensureTarget()
	if global.Preview != nil {
		global.Preview.Update(uint32(time.Since(this.start).Milliseconds()))
	}

	length, height, width := this.volume()
	if length != this.length || height != this.height || width != this.width {
		this.length, this.height, this.width = length, height, width
		this.ResetCamera()
	}

	inside := mouseFree && raylib.CheckCollisionPointRec(raylib.GetMousePosition(), bounds)
	if inside {
		if wheel := raylib.GetMouseWheelMove(); wheel != 0 {
			this.zoom(wheel)
		}
	}
	painting := global.ShouldPaint && global.Preview == nil && global.SelectedGraph() != nil
	this.hover = nil
	if inside && global.Preview == nil {
		this.hover = this.pick()
	}
	if inside && painting && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) && this.hover != nil {
		global.PushUndo()
		this.painting = true
	}
	if this.painting && raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		if this.hover != nil && *this.hover != byte(global.SelectedColor) {
			*this.hover = byte(global.SelectedColor)
			global.MarkDirty()
		}
	} else {
		this.painting = false
	}
	if inside && (raylib.IsMouseButtonDown(raylib.MouseRightButton) ||
		!painting && raylib.IsMouseButtonDown(raylib.MouseLeftButton)) {
		this.orbit(raylib.GetMouseDelta())
	}
}

// ensureTarget keeps the render texture at the view's size (in physical pixels).
func (this *CubeGrid) ensureTarget() {
	scale := raylib.GetWindowScaleDPI().X
	if scale <= 0 {
		scale = 1
	}
	width, height := int32(this.bounds.Width*scale), int32(this.bounds.Height*scale)
	if width < 1 || height < 1 {
		return
	}
	if this.target.Texture.Width == width && this.target.Texture.Height == height {
		return
	}
	if this.target.ID != 0 {
		raylib.UnloadRenderTexture(this.target)
	}
	this.target = raylib.LoadRenderTexture(width, height)
	raylib.SetTextureFilter(this.target.Texture, raylib.FilterBilinear)
	this.scale = scale
}

// Unload frees the render texture.
func (this *CubeGrid) Unload() {
	if this.target.ID != 0 {
		raylib.UnloadRenderTexture(this.target)
		this.target = raylib.RenderTexture2D{}
	}
}

func (this *CubeGrid) Render() {
	if this.target.ID == 0 {
		return
	}
	raylib.BeginTextureMode(this.target)
	raylib.ClearBackground(viewBackground)
	raylib.BeginMode3D(this.camera)
	this.forEachVoxel(func(position raylib.Vector3, voxel *byte, voxelColor color.RGBA) {
		size := raylib.NewVector3(voxelSize, voxelSize, voxelSize)
		if voxelColor.A > 0 {
			raylib.DrawCubeV(position, size, voxelColor)
			raylib.DrawCubeWiresV(position, size, wireColor)
		} else {
			raylib.DrawCubeWiresV(position, size, unlitWireColor)
		}
		if voxel != nil && voxel == this.hover {
			raylib.DrawCubeWiresV(position, raylib.NewVector3(voxelSize*1.15, voxelSize*1.15, voxelSize*1.15), hoverColor)
		}
	})
	raylib.EndMode3D()
	raylib.EndTextureMode()

	// Render textures are upside down in OpenGL: flip the source height.
	source := raylib.NewRectangle(0, 0, float32(this.target.Texture.Width), -float32(this.target.Texture.Height))
	raylib.DrawTexturePro(this.target.Texture, source, this.bounds, raylib.NewVector2(0, 0), 0, raylib.White)
	this.drawOverlay()
}

// viewSize is the render texture size in logical pixels, for projections.
func (this *CubeGrid) viewSize() (int32, int32) {
	return int32(this.bounds.Width), int32(this.bounds.Height)
}

// toView converts a window position to the view's own coordinates.
func (this *CubeGrid) toView(position raylib.Vector2) raylib.Vector2 {
	return raylib.NewVector2(position.X-this.bounds.X, position.Y-this.bounds.Y)
}

func (this *CubeGrid) drawOverlay() {
	// Axis gizmo: project the three unit axes from the origin corner.
	origin := voxelPosition(-1, -1, -1)
	axes := [...]struct {
		end   raylib.Vector3
		color raylib.Color
		label string
	}{
		{voxelPosition(2, -1, -1), raylib.NewColor(0xe7, 0x4c, 0x3c, 255), "x"},
		{voxelPosition(-1, 2, -1), raylib.NewColor(0x2e, 0xcc, 0x71, 255), "y"},
		{voxelPosition(-1, -1, 2), raylib.NewColor(0x2f, 0x80, 0xed, 255), "z (front)"},
	}
	center := raylib.NewVector2(this.bounds.X+60, this.bounds.Y+this.bounds.Height-50)
	width, height := this.viewSize()
	from := raylib.GetWorldToScreenEx(origin, this.camera, width, height)
	for _, axis := range axes {
		to := raylib.GetWorldToScreenEx(axis.end, this.camera, width, height)
		direction := raylib.Vector2Normalize(raylib.Vector2Subtract(to, from))
		if math.IsNaN(float64(direction.X)) {
			continue
		}
		end := raylib.Vector2Add(center, raylib.Vector2Scale(direction, 34))
		raylib.DrawLineEx(center, end, 2, axis.color)
		utils.RenderText(axis.label, raylib.NewRectangle(end.X+2, end.Y-style.TextSize/2, 70, style.TextSize), axis.color)
	}

	text := ""
	switch {
	case global.Preview != nil:
		text = fmt.Sprintf("Preview - cartoon %d / %d", global.Preview.Current()+1, global.Preview.Count())
		if global.Preview.Paused {
			text += " (paused)"
		}
	case global.SelectedGraph() != nil:
		graph := global.SelectedGraph()
		text = fmt.Sprintf("Graph %q  %dx%dx%d", graph.Name, graph.L, graph.H, graph.W)
		if global.ShouldPaint {
			text += fmt.Sprintf("  - painting index %d", global.SelectedColor)
		}
		switch global.SelectedLayerState {
		case global.Layer:
			text += fmt.Sprintf("  - layer y=%d", global.SelectedLayer)
		case global.Column:
			text += fmt.Sprintf("  - column z=%d", global.SelectedColumn)
		case global.Precise:
			text += fmt.Sprintf("  - y=%d z=%d", global.SelectedLayer, global.SelectedColumn)
		}
	default:
		text = "Select a graph or play the preview"
	}
	utils.RenderText(text, raylib.NewRectangle(this.bounds.X+10, this.bounds.Y+6, this.bounds.Width-120, style.TextSize), component.MutedColor)
	if component.Button(raylib.NewRectangle(this.bounds.X+this.bounds.Width-96, this.bounds.Y+6, 86, component.ControlHeight()), "Reset view") {
		this.ResetCamera()
	}
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
	if graph := global.SelectedGraph(); graph != nil {
		return graph.L, graph.H, graph.W
	}
	return 0, 0, 0
}

// ResetCamera looks at the front of the volume from slightly right and above, like the vendor
// simulator.
func (this *CubeGrid) ResetCamera() {
	this.distance = float32(max(this.length, this.height, this.width, 1)) * voxelSpacing * 1.6
	this.camera.Target = this.center()
	this.camera.Position = raylib.Vector3Add(this.center(), raylib.Vector3Scale(raylib.Vector3Normalize(raylib.NewVector3(0.6, 0.6, 1.4)), this.distance*1.7))
}

func (this *CubeGrid) center() raylib.Vector3 {
	return raylib.Vector3Scale(voxelPosition(this.length-1, this.height-1, this.width-1), 0.5)
}

func (this *CubeGrid) zoom(wheel float32) {
	offset := raylib.Vector3Subtract(this.camera.Position, this.camera.Target)
	length := raylib.Vector3Length(offset) * (1 - wheel*0.1)
	length = max(this.distance*0.3, min(length, this.distance*4))
	this.camera.Position = raylib.Vector3Add(this.camera.Target, raylib.Vector3Scale(raylib.Vector3Normalize(offset), length))
}

// orbit rotates the camera around the target by the mouse movement.
func (this *CubeGrid) orbit(delta raylib.Vector2) {
	offset := raylib.Vector3Subtract(this.camera.Position, this.camera.Target)
	yaw := -delta.X * 0.005
	pitch := -delta.Y * 0.005
	offset = raylib.Vector3RotateByAxisAngle(offset, raylib.NewVector3(0, 1, 0), yaw)
	right := raylib.Vector3Normalize(raylib.Vector3CrossProduct(this.camera.Up, offset))
	rotated := raylib.Vector3RotateByAxisAngle(offset, right, pitch)
	// Keep away from the poles so Up stays valid.
	if math.Abs(float64(raylib.Vector3Normalize(rotated).Y)) < 0.98 {
		offset = rotated
	}
	this.camera.Position = raylib.Vector3Add(this.camera.Target, offset)
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

	graph := global.SelectedGraph()
	if graph == nil {
		return
	}
	palette := render.DefaultPalette
	if global.SelectedPalette >= 0 && global.SelectedPalette < len(global.Project.Palettes) {
		palette = string(global.Project.Palettes[global.SelectedPalette].Colors[:])
	}
	forEachCell(graph.L, graph.H, graph.W, func(x, y, z, index int) {
		if !component.LayerVisible(y, z) {
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

// pick returns the nearest voxel under the mouse.
func (this *CubeGrid) pick() *byte {
	width, height := this.viewSize()
	ray := raylib.GetScreenToWorldRayEx(this.toView(raylib.GetMousePosition()), this.camera, width, height)
	var target *byte
	nearest := float32(math.MaxFloat32)
	half := raylib.NewVector3(voxelSize/2.0, voxelSize/2.0, voxelSize/2.0)
	this.forEachVoxel(func(position raylib.Vector3, voxel *byte, _ color.RGBA) {
		box := raylib.NewBoundingBox(raylib.Vector3Subtract(position, half), raylib.Vector3Add(position, half))
		if hit := raylib.GetRayCollisionBox(ray, box); hit.Hit && hit.Distance < nearest {
			nearest, target = hit.Distance, voxel
		}
	})
	return target
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
