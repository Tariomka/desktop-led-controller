package component

import (
	"fmt"
	"strconv"

	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// ToolsTab holds the graph editing tools; painting itself happens in the 3-D view.
type ToolsTab struct {
	Actions *Actions

	// TextTool renders text into a font strip graph; set by the window (needs raylib).
	TextTool func(text string, height int) (blob.GraphData, bool)
	text     string
	height   int
	shift    int
}

func (this *ToolsTab) Draw(layout *rowLayout) {
	graph := global.SelectedGraph()
	if graph == nil {
		layout.label("Select a graph in the Assets panel to edit it.", MutedColor)
		if global.Preview != nil {
			layout.label("Stop the preview to see the graph.", MutedColor)
		}
		this.drawTextTool(layout)
		return
	}
	if global.Preview != nil {
		layout.label("The preview is running; stop it in the Show tab to paint.", WarningColor)
	}

	layout.header("Paint - " + graph.Name)
	cells := layout.row(2)
	if paint := checkBox(cells[0], "Paint (left mouse)", global.ShouldPaint); paint != global.ShouldPaint {
		global.ShouldPaint = paint
	}
	if button(cells[1], "Erase (index 0)", false) {
		global.SelectedColor = 0
	}
	layout.label("Right mouse rotates, wheel zooms the view.", MutedColor)

	palette := this.drawPaletteSelector(layout)
	layout.label(fmt.Sprintf("Paint index %d", global.SelectedColor), MutedColor)
	if slot := swatches(layout, palette, global.SelectedColor); slot >= 0 {
		global.SelectedColor = slot
	}

	layout.header("Slice")
	cells = layout.row(1)
	global.SelectedLayerState = global.LayerState(comboBox(cells[0], "All;Layer (y);Column (z);Layer and column", int(global.SelectedLayerState)))
	if global.SelectedLayerState == global.Layer || global.SelectedLayerState == global.Precise {
		global.SelectedLayer = min(global.SelectedLayer, graph.H-1)
		cells = layout.split(1, 3)
		layoutLabel(cells[0], "Layer y")
		global.SelectedLayer = int(slider(cells[1], "", strconv.Itoa(global.SelectedLayer), float32(global.SelectedLayer), 0, float32(graph.H-1)) + 0.5)
	}
	if global.SelectedLayerState == global.Column || global.SelectedLayerState == global.Precise {
		global.SelectedColumn = min(global.SelectedColumn, graph.W-1)
		cells = layout.split(1, 3)
		layoutLabel(cells[0], "Column z")
		global.SelectedColumn = int(slider(cells[1], "", strconv.Itoa(global.SelectedColumn), float32(global.SelectedColumn), 0, float32(graph.W-1)) + 0.5)
	}
	layout.label("PageUp/PageDown move the layer.", MutedColor)
	cells = layout.row(2)
	if button(cells[0], "Fill visible", false) {
		this.apply(graph, func() { fillVisible(graph, byte(global.SelectedColor)) })
	}
	if button(cells[1], "Clear visible", false) {
		this.apply(graph, func() { fillVisible(graph, 0) })
	}

	layout.header("Transform")
	cells = layout.row(3)
	for axis, name := range [...]string{"Mirror X", "Mirror Y", "Mirror Z"} {
		if button(cells[axis], name, false) {
			this.apply(graph, func() { show.Mirror(graph, axis) })
		}
	}
	cells = layout.row(3)
	for axis, name := range [...]string{"Rotate X", "Rotate Y", "Rotate Z"} {
		if button(cells[axis], name, false) {
			this.apply(graph, func() { show.Rotate(graph, axis) })
		}
	}
	layout.label("Rotations are 90 degrees and swap the other two sizes.", MutedColor)
	cells = layout.row(3)
	for axis, name := range [...]string{"Shift X", "Shift Y", "Shift Z"} {
		if button(cells[axis], name, false) {
			this.apply(graph, func() { show.Shift(graph, axis, this.shiftBy()) })
		}
	}
	cells = layout.split(2, 1)
	layoutLabel(cells[0], "Shift by (negative = back)")
	if this.shift == 0 {
		this.shift = 1
	}
	valueBox("tool-shift", cells[1], &this.shift, -64, 64)

	layout.header("History")
	cells = layout.row(2)
	if button(cells[0], "Undo (Ctrl+Z)", !global.CanUndo()) {
		global.Undo()
	}
	if button(cells[1], "Redo (Ctrl+Y)", !global.CanRedo()) {
		global.Redo()
	}

	this.drawTextTool(layout)
}

func (this *ToolsTab) shiftBy() int {
	if this.shift == 0 {
		return 1
	}
	return this.shift
}

// apply records an undo step, runs the change and marks the project dirty.
func (this *ToolsTab) apply(graph *blob.GraphData, change func()) {
	_ = graph
	global.PushUndo()
	change()
	global.MarkDirty()
}

// drawPaletteSelector picks the palette used to show graphs; it returns its colors.
func (this *ToolsTab) drawPaletteSelector(layout *rowLayout) string {
	palettes := global.Project.Palettes
	if global.SelectedPalette >= len(palettes) {
		global.SelectedPalette = 0
	}
	name := "(default palette)"
	colors := defaultPalette()
	if len(palettes) > 0 {
		name = palettes[global.SelectedPalette].Name
		colors = string(palettes[global.SelectedPalette].Colors[:])
	}
	cells := layout.split(1, 4, 1)
	if button(cells[0], "<", len(palettes) < 2) {
		global.SelectedPalette = (global.SelectedPalette + len(palettes) - 1) % len(palettes)
	}
	layoutLabel(cells[1], "Palette: "+name)
	if button(cells[2], ">", len(palettes) < 2) {
		global.SelectedPalette = (global.SelectedPalette + 1) % len(palettes)
	}
	return colors
}

func (this *ToolsTab) drawTextTool(layout *rowLayout) {
	layout.header("Text tool")
	layout.label("Renders text into a new font strip graph for Title props.", MutedColor)
	changed := textBox("tool-text", layout.row(1)[0], &this.text, textLength)
	_ = changed
	if this.height == 0 {
		this.height = 8
	}
	cells := layout.split(1, 1, 2)
	layoutLabel(cells[0], "Height")
	valueBox("tool-text-height", cells[1], &this.height, 5, 32)
	if button(cells[2], "Create graph", this.text == "" || this.TextTool == nil) {
		graph, ok := this.TextTool(this.text, this.height)
		if !ok {
			this.Actions.Error("The text could not be rendered.")
			return
		}
		project := global.Project
		graph.Name = show.UniqueName(truncate(this.text, 11), func(name string) bool { return show.FindGraph(project, name) >= 0 })
		project.Graphs = append(project.Graphs, graph)
		global.Selected = global.Selection{Kind: global.SelectGraph, Index: len(project.Graphs) - 1}
		global.MarkDirty()
		this.Actions.Status(fmt.Sprintf("Created graph %q (%dx%dx%d)", graph.Name, graph.L, graph.H, graph.W), false)
	}
	layout.label("Strip: one column per pixel column, height rows, width 1; uses paint index.", MutedColor)
}

func truncate(text string, length int) string {
	if len(text) <= length {
		return text
	}
	return text[:length]
}

// fillVisible sets the voxels of the shown layer/column selection.
func fillVisible(graph *blob.GraphData, value byte) {
	for z := range graph.W {
		for y := range graph.H {
			if !layerVisible(y, z) {
				continue
			}
			for x := range graph.L {
				graph.Voxels[show.VoxelIndex(graph, x, y, z)] = value
			}
		}
	}
}

// LayerVisible reports whether a voxel at (y, z) is shown with the current slice.
func layerVisible(y, z int) bool {
	switch global.SelectedLayerState {
	case global.Layer:
		return y == global.SelectedLayer
	case global.Column:
		return z == global.SelectedColumn
	case global.Precise:
		return y == global.SelectedLayer && z == global.SelectedColumn
	}
	return true
}

// LayerVisible is the exported form for the 3-D view.
func LayerVisible(y, z int) bool { return layerVisible(y, z) }

var _ = style.TextSize
