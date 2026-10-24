package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Names of the key function codes 0..19 (see show-engine.md).
var keyFunctions = [...]string{
	"None", "Replay", "Previous", "Next", "Speed up", "Speed reset", "Speed down",
	"Play model", "Pause", "Display off",
	"Cartoon 1", "Cartoon 2", "Cartoon 3", "Cartoon 4", "Cartoon 5",
	"Cartoon 6", "Cartoon 7", "Cartoon 8", "Cartoon 9", "Cartoon 10",
}

var (
	cartoonKinds = [...]string{blob.Normal: "Normal", blob.Script: "Script", blob.Audio: "Audio"}
	elementKinds = [...]string{blob.SetGraph: "Graph", blob.DrawBall: "Ball", blob.TitleStyle: "Title"}
	cubeTypes    = [...]string{"custom", "8x8x8", "8x16x8", "16x16x16", "8x32x8"}
)

// PropsTab edits the selected asset, cartoon or prop.
type PropsTab struct {
	Actions *Actions

	resize    [3]int
	resizeFor *blob.GraphData
	slot      int // palette slot
	paletteID string
	pathPoint int
}

// changed marks the project dirty when a control reports a change.
func changed(flag bool) bool {
	if flag {
		global.MarkDirty()
	}
	return flag
}

func nameOf(names []string, index int) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "Code " + strconv.Itoa(index)
}

func (this *PropsTab) Draw(layout *rowLayout) {
	switch global.Selected.Kind {
	case global.SelectGraph:
		this.drawGraph(layout)
	case global.SelectPalette:
		this.drawPalette(layout)
	case global.SelectPath:
		this.drawPath(layout)
	case global.SelectCartoon:
		this.drawCartoon(layout)
	case global.SelectProp:
		this.drawCartoon(layout)
		layout.separator()
		this.drawProp(layout)
	case global.SelectKeys:
		this.drawKeys(layout)
	case global.SelectDisplay:
		this.drawDisplay(layout)
	default:
		layout.label("Select something in the Assets panel.", MutedColor)
	}
}

// field draws a labelled integer box in the given cell.
func field(id string, cell raylib.Rectangle, label string, value *int, minimum, maximum int) {
	if label != "" {
		labelWidth := float32(len(label))*7 + 6
		utils.RenderText(label, raylib.NewRectangle(cell.X, cell.Y, labelWidth, cell.Height), MutedColor)
		cell.X += labelWidth
		cell.Width -= labelWidth
	}
	changed(valueBox(id, cell, value, minimum, maximum))
}

func (this *PropsTab) drawGraph(layout *rowLayout) {
	graph := global.SelectedGraph()
	if graph == nil {
		layout.label("No graph selected.", MutedColor)
		return
	}
	if graph != this.resizeFor {
		this.resize, this.resizeFor = [3]int{graph.L, graph.H, graph.W}, graph
	}
	layout.header("Graph")
	layout.label("Name (15 characters)", MutedColor)
	changed(textBox("graph-name", layout.row(1)[0], &graph.Name, nameLength))

	layout.label("Size L x H x W (H is up)", MutedColor)
	cells := layout.row(4)
	for i := range this.resize {
		valueBox("graph-size-"+strconv.Itoa(i), cells[i], &this.resize[i], 1, 255)
	}
	if button(cells[3], "Resize", this.resize == [3]int{graph.L, graph.H, graph.W}) {
		global.PushUndo()
		show.ResizeGraph(graph, this.resize[0], this.resize[1], this.resize[2])
		this.resizeFor = nil
		global.MarkDirty()
	}
	lit := 0
	for _, voxel := range graph.Voxels {
		if voxel != 0 {
			lit++
		}
	}
	layout.label(fmt.Sprintf("%d of %d voxels lit", lit, len(graph.Voxels)), MutedColor)
	layout.label("Paint it in the 3-D view with the Tools tab.", MutedColor)
	this.drawUsers(layout, func(element blob.ElementData) []string {
		var names []string
		for _, prop := range element.Graphs {
			names = append(names, prop.Graph)
		}
		for _, prop := range element.Titles {
			names = append(names, prop.Graph)
		}
		return names
	}, graph.Name)
}

// drawUsers lists the cartoons whose props reference the name.
func (this *PropsTab) drawUsers(layout *rowLayout, names func(blob.ElementData) []string, name string) {
	var users []string
	for group, cartoons := range global.Project.Groups {
		for _, cartoon := range cartoons {
			for _, element := range cartoon.Elements {
				for _, used := range names(element) {
					if used == name || strings.Contains(used, "$") {
						users = append(users, fmt.Sprintf("%d: %s", group+1, cartoon.Name))
						goto next
					}
				}
			}
		next:
		}
	}
	if len(users) == 0 {
		layout.label("Not used by any cartoon.", MutedColor)
		return
	}
	layout.label(fmt.Sprintf("Used by %d cartoon(s): %s", len(users), strings.Join(users, ", ")), MutedColor)
}

func (this *PropsTab) drawPalette(layout *rowLayout) {
	palettes := global.Project.Palettes
	if global.Selected.Index < 0 || global.Selected.Index >= len(palettes) {
		layout.label("No palette selected.", MutedColor)
		return
	}
	palette := &palettes[global.Selected.Index]
	layout.header("Palette")
	layout.label("Name (15 characters)", MutedColor)
	changed(textBox("palette-name", layout.row(1)[0], &palette.Name, nameLength))

	layout.label("Slots (click to edit)", MutedColor)
	if slot := swatches(layout, string(palette.Colors[:]), this.slot); slot >= 0 {
		this.slot = slot
	}
	value := palette.Colors[this.slot]
	layout.label(fmt.Sprintf("Slot %d: 0x%02X = 0b%06b (RRGGBB)", this.slot, value, value), style.TextColorNormal)
	if edited := colorPicker(layout, "palette-color", value); edited != value {
		palette.Colors[this.slot] = edited
		global.MarkDirty()
	}
	cells := layout.row(2)
	if button(cells[0], "Default colors", false) {
		copy(palette.Colors[:], defaultPalette())
		global.MarkDirty()
	}
	if button(cells[1], "Clear all", false) {
		palette.Colors = [16]byte{}
		global.MarkDirty()
	}
	this.drawUsers(layout, func(element blob.ElementData) []string {
		var names []string
		for _, prop := range element.Graphs {
			names = append(names, prop.Palette)
		}
		for _, prop := range element.Titles {
			names = append(names, prop.Palette)
		}
		return names
	}, palette.Name)
}

func (this *PropsTab) drawPath(layout *rowLayout) {
	indexes := global.Project.Indexes
	if global.Selected.Index < 0 || global.Selected.Index >= len(indexes) {
		layout.label("No path selected.", MutedColor)
		return
	}
	index := &indexes[global.Selected.Index]
	points := show.PathPoints(*index)
	layout.header("Path (title route)")
	layout.label("Name (15 characters)", MutedColor)
	changed(textBox("path-name", layout.row(1)[0], &index.Name, nameLength))

	// Top view of the first display: x to the right, z down (front row at the top).
	l, w := 8, 8
	if displays := global.Project.Config.Displays; len(displays) > 0 {
		l, _, w = displays[0].Size()
	}
	layout.label(fmt.Sprintf("%d points. Click cells to append; x right, z down (z = 0 is the front).", len(points)), MutedColor)
	cell := min((layout.width-2)/float32(l), 28)
	gridBounds := layout.rowHeight(1, cell*float32(w)+2)[0]
	for z := range w {
		for x := range l {
			bounds := raylib.NewRectangle(gridBounds.X+float32(x)*cell, gridBounds.Y+float32(z)*cell, cell-1, cell-1)
			raylib.DrawRectangleRec(bounds, HeaderColor)
			order := -1
			for i, point := range points {
				if point.X == x && point.Z == z {
					order = i
				}
			}
			if order >= 0 {
				shade := raylib.Fade(AccentColor, 0.4+0.6*float32(order+1)/float32(len(points)))
				raylib.DrawRectangleRec(bounds, shade)
				utils.RenderText(strconv.Itoa(order+1), bounds, raylib.White)
			}
			if clicked(bounds) {
				points = append(points, show.PathPoint{X: x, Z: z})
				index.Data = show.PathData(points)
				global.MarkDirty()
			}
		}
	}

	cells := layout.row(3)
	if button(cells[0], "Undo point", len(points) == 0) {
		index.Data = show.PathData(points[:len(points)-1])
		global.MarkDirty()
	}
	if button(cells[1], "Clear", len(points) == 0) {
		index.Data = nil
		global.MarkDirty()
	}
	if button(cells[2], "Reverse", len(points) < 2) {
		index.Data = show.PathData(show.ReversePath(points))
		global.MarkDirty()
	}
	cells = layout.row(2)
	if button(cells[0], "Around the edge", false) {
		index.Data = show.PathData(show.EdgePath(l, w))
		global.MarkDirty()
	}
	if button(cells[1], "Front row", false) {
		index.Data = show.PathData(show.LinePath(l, 0))
		global.MarkDirty()
	}
	this.drawUsers(layout, func(element blob.ElementData) []string {
		var names []string
		for _, prop := range element.Titles {
			names = append(names, prop.Index)
		}
		return names
	}, index.Name)
}

func (this *PropsTab) drawCartoon(layout *rowLayout) {
	cartoon := global.SelectedCartoon()
	if cartoon == nil {
		layout.label("No cartoon selected.", MutedColor)
		return
	}
	layout.header(fmt.Sprintf("Cartoon - group %d, #%d", global.Selected.Group+1, global.Selected.Cartoon+1))
	layout.label("Name", MutedColor)
	changed(textBox("cartoon-name", layout.row(1)[0], &cartoon.Name, textLength))

	cells := layout.row(2)
	if enabled := checkBox(cells[0], "Enabled (compiled)", cartoon.Enabled); enabled != cartoon.Enabled {
		cartoon.Enabled = enabled
		global.MarkDirty()
	}
	if cartoon.Type < 0 || cartoon.Type >= len(cartoonKinds) {
		layoutLabel(cells[1], nameOf(cartoonKinds[:], cartoon.Type))
	} else if kind := comboBox(cells[1], "Normal;Script;Audio", cartoon.Type); kind != cartoon.Type {
		cartoon.Type = kind
		if kind == blob.Audio {
			cartoon.Header.FrameCount, cartoon.Header.Loop = 0, 0
		} else if kind == blob.Normal && cartoon.Header.FrameCount == 0 {
			cartoon.Header.FrameCount, cartoon.Header.Loop = 10, 1
		}
		global.MarkDirty()
	}

	switch cartoon.Type {
	case blob.Normal:
		cells = layout.row(3)
		field("cartoon-frames", cells[0], "Frames", &cartoon.Header.FrameCount, 1, 32767)
		field("cartoon-loops", cells[1], "Loops", &cartoon.Header.Loop, 1, 65535)
		field("cartoon-delay", cells[2], "ms", &cartoon.Header.Delay, 1, 65535)
	case blob.Audio:
		cells = layout.row(3)
		field("cartoon-time", cells[0], "Secs", &cartoon.Header.FrameCount, 0, 32767)
		field("cartoon-mode", cells[1], "Mode", &cartoon.Header.Loop, 0, 65535)
		field("cartoon-delay", cells[2], "ms", &cartoon.Header.Delay, 1, 65535)
		layout.label("Secs 0 = forever; the frame follows the audio level (0-63).", MutedColor)
	case blob.Script:
		if len(cartoon.Script) > 0 {
			layout.label(fmt.Sprintf("Compiled script: %d words (from imported show data).", len(cartoon.Script)), MutedColor)
		} else {
			layout.label("Script source only: this app cannot compile scripts yet.", WarningColor)
		}
		if cartoon.ScriptSource != "" {
			lines := strings.Split(strings.ReplaceAll(cartoon.ScriptSource, "\r\n", "\n"), "\n")
			layout.label(fmt.Sprintf("Source: %d lines (read-only)", len(lines)), MutedColor)
			for _, line := range lines[:min(len(lines), 12)] {
				layout.label(line, style.TextColorNormal)
			}
		}
		return
	}
	layout.label(fmt.Sprintf("%d elements - add and arrange props in the Timeline.", len(cartoon.Elements)), MutedColor)
}

func (this *PropsTab) drawProp(layout *rowLayout) {
	element := global.SelectedElement()
	if element == nil {
		layout.label("Select a bar in the timeline.", MutedColor)
		return
	}
	prop := global.Selected.Prop
	count := show.PropCount(*element)
	if prop < 0 || prop >= count {
		layout.label("Select a bar in the timeline.", MutedColor)
		return
	}
	layout.header(fmt.Sprintf("Element %d (%s) - prop %d of %d", global.Selected.Element+1, nameOf(elementKinds[:], element.Kind), prop+1, count))

	start, end := show.PropSpan(*element, prop)
	cells := layout.row(2)
	field("prop-start", cells[0], "Start", &start, 0, 32767)
	field("prop-end", cells[1], "End", &end, 0, 32767)
	if s, e := show.PropSpan(*element, prop); s != start || e != end {
		show.SetPropSpan(element, prop, start, end)
	}

	switch element.Kind {
	case blob.SetGraph:
		this.drawGraphProp(layout, &element.Graphs[prop])
	case blob.DrawBall:
		this.drawBallProp(layout, &element.Balls[prop])
	case blob.TitleStyle:
		this.drawTitleProp(layout, &element.Titles[prop])
	}
}

// namePicker is a text box plus a button that opens a list of the existing names (names may
// also be expressions like Fly_$C%4+1, typed directly).
func (this *PropsTab) namePicker(layout *rowLayout, id, label string, value *string, names []string) {
	layout.label(label, MutedColor)
	cells := layout.split(3, 1)
	changed(textBox(id, cells[0], value, nameLength))
	if button(cells[1], "Pick...", len(names) == 0) {
		this.Actions.Pick(label, names, func(name string) {
			*value = name
			global.MarkDirty()
		})
	}
}

func graphNames() []string {
	names := make([]string, len(global.Project.Graphs))
	for i, graph := range global.Project.Graphs {
		names[i] = graph.Name
	}
	return names
}

func paletteNames() []string {
	names := make([]string, len(global.Project.Palettes))
	for i, palette := range global.Project.Palettes {
		names[i] = palette.Name
	}
	return names
}

func pathNames() []string {
	names := make([]string, len(global.Project.Indexes))
	for i, index := range global.Project.Indexes {
		names[i] = index.Name
	}
	return names
}

func (this *PropsTab) drawGraphProp(layout *rowLayout, prop *blob.GraphProp) {
	this.namePicker(layout, "prop-graph", "Graph", &prop.Graph, graphNames())
	this.namePicker(layout, "prop-palette", "Palette (empty = default)", &prop.Palette, paletteNames())
	layout.label("Position from -> to", MutedColor)
	cells := layout.row(3)
	field("prop-x0", cells[0], "x", &prop.X0, -32768, 32767)
	field("prop-y0", cells[1], "y", &prop.Y0, -32768, 32767)
	field("prop-z0", cells[2], "z", &prop.Z0, -32768, 32767)
	cells = layout.row(3)
	field("prop-x1", cells[0], "x", &prop.X1, -32768, 32767)
	field("prop-y1", cells[1], "y", &prop.Y1, -32768, 32767)
	field("prop-z1", cells[2], "z", &prop.Z1, -32768, 32767)
	layout.label("Angle from -> to (degrees), axis, mode", MutedColor)
	cells = layout.row(4)
	field("prop-a0", cells[0], "", &prop.Angle0, -32768, 32767)
	field("prop-a1", cells[1], "", &prop.Angle1, -32768, 32767)
	if axis := comboBox(cells[2], "X;Y;Z", prop.Axis); axis != prop.Axis {
		prop.Axis, _ = axis, changed(true)
	}
	if mode := comboBox(cells[3], "Set;Mix", prop.Mode); mode != prop.Mode {
		prop.Mode, _ = mode, changed(true)
	}
	cells = layout.row(2)
	if button(cells[0], "Copy from -> to", false) {
		prop.X1, prop.Y1, prop.Z1, prop.Angle1 = prop.X0, prop.Y0, prop.Z0, prop.Angle0
		global.MarkDirty()
	}
	if button(cells[1], "Center on cube", false) {
		prop.X0, prop.Y0, prop.Z0, prop.X1, prop.Y1, prop.Z1 = 0, 0, 0, 0, 0, 0
		global.MarkDirty()
	}
}

func (this *PropsTab) drawBallProp(layout *rowLayout, prop *blob.BallProp) {
	layout.label("Centre x y z (1/100 voxel: 350 = 3.5)", MutedColor)
	cells := layout.row(3)
	field("ball-x", cells[0], "x", &prop.X, -32768, 32767)
	field("ball-y", cells[1], "y", &prop.Y, -32768, 32767)
	field("ball-z", cells[2], "z", &prop.Z, -32768, 32767)
	layout.label("Radius from -> to (1/100 voxel), hollow", MutedColor)
	cells = layout.row(3)
	field("ball-r0", cells[0], "", &prop.R0, 0, 32767)
	field("ball-r1", cells[1], "", &prop.R1, 0, 32767)
	if hollow := checkBox(cells[2], "Hollow", prop.Hollow); hollow != prop.Hollow {
		prop.Hollow, _ = hollow, changed(true)
	}
	layout.label("Inner color", MutedColor)
	if value := colorPicker(layout, "ball-in", prop.ColorIn); value != prop.ColorIn {
		prop.ColorIn, _ = value, changed(true)
	}
	layout.label("Outer color", MutedColor)
	if value := colorPicker(layout, "ball-out", prop.ColorOut); value != prop.ColorOut {
		prop.ColorOut, _ = value, changed(true)
	}
}

func (this *PropsTab) drawTitleProp(layout *rowLayout, prop *blob.TitleProp) {
	this.namePicker(layout, "title-graph", "Font strip graph", &prop.Graph, graphNames())
	this.namePicker(layout, "title-palette", "Palette (empty = default)", &prop.Palette, paletteNames())
	this.namePicker(layout, "title-path", "Path", &prop.Index, pathNames())
	cells := layout.row(2)
	field("title-high", cells[0], "Height y", &prop.High, -32768, 32767)
	if style := comboBox(cells[1], "Forward;Reverse", prop.Style); style != prop.Style {
		prop.Style, _ = style, changed(true)
	}
	layout.label("The strip scrolls along the path; frames = points + strip columns.", MutedColor)
}

func (this *PropsTab) drawKeys(layout *rowLayout) {
	config := &global.Project.Config
	layout.header("Cube keys")
	layout.label("Function of KEY1-KEY7 while this show plays", MutedColor)
	for key := range config.Keys {
		cells := layout.split(1, 1, 2)
		layoutLabel(cells[0], "KEY"+strconv.Itoa(key+1))
		value := int(config.Keys[key])
		if changed(spinner("key-"+strconv.Itoa(key), cells[1], &value, 0, len(keyFunctions)-1)) {
			config.Keys[key] = byte(value)
		}
		layoutLabel(cells[2], nameOf(keyFunctions[:], int(config.Keys[key])))
	}
	if button(layout.row(1)[0], "Factory defaults", false) {
		config.Keys = show.DefaultKeys
		global.MarkDirty()
	}
}

func (this *PropsTab) drawDisplay(layout *rowLayout) {
	displays := global.Project.Config.Displays
	if global.Selected.Index < 0 || global.Selected.Index >= len(displays) {
		layout.label("No display selected.", MutedColor)
		return
	}
	display := &displays[global.Selected.Index]
	layout.header(fmt.Sprintf("Display (cube) %d of %d", global.Selected.Index+1, len(displays)))
	layout.label("Position in the show space", MutedColor)
	cells := layout.row(3)
	field("display-x", cells[0], "x", &display.X, -32768, 32767)
	field("display-y", cells[1], "y", &display.Y, -32768, 32767)
	field("display-z", cells[2], "z", &display.Z, -32768, 32767)
	cells = layout.row(2)
	if cubeType := comboBox(cells[0], strings.Join(cubeTypes[:], ";"), display.CubeType); cubeType != display.CubeType {
		display.CubeType, _ = cubeType, changed(true)
	}
	if dir := comboBox(cells[1], "Bottom;Top", display.Dir); dir != display.Dir {
		display.Dir, _ = dir, changed(true)
	}
	layout.label("Type 0 is treated as 8x8x8 by the cube. Dir Top turns the picture 180 degrees.", MutedColor)
}
