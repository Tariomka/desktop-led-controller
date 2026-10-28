package ui

import (
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// RenderText draws text with raylib's default font into a font strip graph: L = text width in
// pixels, H = height, W = 1, lit pixels get the current paint index. Needs an open window.
func RenderText(text string, height int) (blob.GraphData, bool) {
	text = strings.TrimSpace(text)
	if text == "" || height < 1 {
		return blob.GraphData{}, false
	}
	// The default font is 10 px high with the glyphs in the upper 8 rows; render large and
	// scale down to the wanted height for a cleaner result.
	fontSize := float32(height * 4)
	image := raylib.ImageTextEx(raylib.GetFontDefault(), text, fontSize, fontSize/10, raylib.White)
	defer raylib.UnloadImage(image)
	if image.Width <= 0 || image.Height <= 0 {
		return blob.GraphData{}, false
	}

	colors := raylib.LoadImageColors(image)
	width, rows := int(image.Width), int(image.Height)
	// Trim empty columns and rows.
	left, right, top, bottom := width, -1, rows, -1
	for y := range rows {
		for x := range width {
			if colors[y*width+x].A > 128 {
				left, right, top, bottom = min(left, x), max(right, x), min(top, y), max(bottom, y)
			}
		}
	}
	if right < left {
		return blob.GraphData{}, false
	}
	scale := float64(bottom-top+1) / float64(height)
	length := max(1, int(float64(right-left+1)/scale+0.5))
	graph := show.NewGraph(text, length, height, 1)
	value := byte(max(global.SelectedColor, 1))
	for y := range height {
		for x := range length {
			// Sample the centre of each destination cell; rows are flipped because y is up.
			sx := left + int((float64(x)+0.5)*scale)
			sy := top + int((float64(height-1-y)+0.5)*scale)
			if sx <= right && sy <= bottom && colors[sy*width+sx].A > 128 {
				graph.Voxels[show.VoxelIndex(&graph, x, y, 0)] = value
			}
		}
	}
	return graph, true
}
