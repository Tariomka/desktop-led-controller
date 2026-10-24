package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	tabGraph = iota
	tabPalette
	tabCartoon
	tabKeys
)

// nameLength is the TextBox buffer size for show names: 15 bytes and the terminator.
const nameLength = 16

// Names of the key function codes 0..19 (see show-engine.md).
var keyFunctions = [...]string{
	"None", "Replay", "Previous", "Next", "Speed up", "Speed reset", "Speed down",
	"Play model", "Pause", "Display off",
	"Cartoon 1", "Cartoon 2", "Cartoon 3", "Cartoon 4", "Cartoon 5",
	"Cartoon 6", "Cartoon 7", "Cartoon 8", "Cartoon 9", "Cartoon 10",
}

var cartoonKinds = [...]string{blob.Normal: "Normal", blob.Script: "Script", blob.Audio: "Audio"}

// nameOf returns names[index], or a placeholder for codes this editor doesn't know.
func nameOf(names []string, index int) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "Code " + strconv.Itoa(index)
}

// EditPanel edits the project: graphs (painted in the 3-D view), palettes, cartoons and keys.
type EditPanel struct {
	Panel
	padding float32

	tab     int32
	editing string // id of the text or value box in edit mode
	scroll  int32

	newSize [3]int // graph size for Resize
	sizeFor *blob.GraphData

	paletteSlot int

	group   int32
	cartoon int
}

func newEditPanel(base Panel) *EditPanel {
	return &EditPanel{Panel: base, padding: 10, newSize: [3]int{8, 8, 8}}
}

func (this *EditPanel) Update() { this.resize() }

func (this *EditPanel) Render() {
	this.renderPanel()
	layout := newRowLayout(this.Rectangle, this.padding)
	tab := this.tab
	raygui.ToggleGroup(layout.row(4)[0], "Graph;Palette;Cartoon;Keys", &tab)
	if tab != this.tab {
		this.tab, this.editing, this.scroll = tab, "", 0
	}
	layout.separator()

	switch this.tab {
	case tabGraph:
		this.renderGraph(layout)
	case tabPalette:
		this.renderPalette(layout)
	case tabCartoon:
		this.renderCartoon(layout)
	case tabKeys:
		this.renderKeys(layout)
	}
}

// textBox edits a string; only one box is in edit mode at a time.
func (this *EditPanel) textBox(id string, bounds raylib.Rectangle, text *string, size int) {
	if raygui.TextBox(bounds, text, size, this.editing == id) {
		if this.editing == id {
			this.editing = ""
		} else {
			this.editing = id
		}
	}
}

func (this *EditPanel) valueBox(id string, bounds raylib.Rectangle, value *int, minimum, maximum int) {
	number := int32(*value)
	if raygui.ValueBox(bounds, "", &number, minimum, maximum, this.editing == id) {
		if this.editing == id {
			this.editing = ""
		} else {
			this.editing = id
		}
	}
	*value = int(number)
}

// list draws a list view and returns the clicked item, or current when nothing was clicked.
func (this *EditPanel) list(layout *rowLayout, rows int, names []string, current int) int {
	bounds := layout.rowHeight(1, float32(rows)*(style.ListItemHeight+style.ListItemSpacing)+style.ListItemSpacing)[0]
	active := int32(current)
	raygui.ListView(bounds, strings.Join(names, ";"), &this.scroll, &active)
	if active < 0 || int(active) >= len(names) {
		return current
	}
	return int(active)
}

func uniqueName(base string, exists func(string) bool) string {
	for i := 1; ; i++ {
		name := base + "_" + strconv.Itoa(i)
		if !exists(name) {
			return name
		}
	}
}

func (this *EditPanel) renderGraph(layout *rowLayout) {
	project := global.Project
	names := make([]string, len(project.Graphs))
	for i, graph := range project.Graphs {
		names[i] = fmt.Sprintf("%s (%dx%dx%d)", graph.Name, graph.L, graph.H, graph.W)
	}
	if selected := this.list(layout, 5, names, global.SelectedGraph); selected != global.SelectedGraph {
		global.SelectedGraph, this.editing = selected, ""
	}

	cells := layout.row(2)
	if button(cells[0], "Add", false) {
		name := uniqueName("Graph", func(name string) bool { return show.FindGraph(project, name) >= 0 })
		project.Graphs = append(project.Graphs, show.NewGraph(name, 8, 8, 8))
		global.SelectedGraph = len(project.Graphs) - 1
	}
	graph := this.graph()
	if button(cells[1], "Delete", graph == nil) {
		project.Graphs = append(project.Graphs[:global.SelectedGraph], project.Graphs[global.SelectedGraph+1:]...)
		global.SelectedGraph = max(0, min(global.SelectedGraph, len(project.Graphs)-1))
		return
	}
	if graph == nil {
		return
	}
	if graph != this.sizeFor {
		this.newSize, this.sizeFor = [3]int{graph.L, graph.H, graph.W}, graph
	}

	layout.label("Name", style.TextColorNormal)
	this.textBox("graph-name", layout.row(1)[0], &graph.Name, nameLength)

	layout.label("Size L x H x W (H is up)", style.TextColorNormal)
	cells = layout.row(4)
	for i := range this.newSize {
		this.valueBox("graph-size-"+strconv.Itoa(i), cells[i], &this.newSize[i], 1, 64)
	}
	if button(cells[3], "Resize", false) {
		show.ResizeGraph(graph, this.newSize[0], this.newSize[1], this.newSize[2])
		this.sizeFor = nil
	}
	layout.separator()

	palette := this.renderPaletteSelector(layout)
	layout.label(fmt.Sprintf("Paint color: index %d", global.SelectedColor), style.TextColorNormal)
	if slot := this.renderSwatches(layout, palette, global.SelectedColor); slot >= 0 {
		global.SelectedColor = slot
	}

	cells = layout.row(1)
	global.ShouldPaint = checkBox(
		raylib.NewRectangle(cells[0].X, cells[0].Y, style.TextSize, style.TextSize),
		"Paint with the left mouse button", global.ShouldPaint)
	layout.label("Right mouse button rotates the view.", style.TextColorNormal)

	cells = layout.row(1)
	global.SelectedLayerState = global.LayerState(comboBox(cells[0], "All;Layer;Column;Precise", int(global.SelectedLayerState)))
	if global.SelectedLayerState == global.Layer || global.SelectedLayerState == global.Precise {
		global.SelectedLayer = uint8(min(int(global.SelectedLayer), graph.H-1))
		bounds := layout.row(1)[0]
		bounds.X += 60
		bounds.Width -= 90
		global.SelectedLayer = uint8(slider(bounds, "Layer y", strconv.Itoa(int(global.SelectedLayer)), float32(global.SelectedLayer), 0, float32(graph.H-1)))
	}
	if global.SelectedLayerState == global.Column || global.SelectedLayerState == global.Precise {
		global.SelectedColumn = uint8(min(int(global.SelectedColumn), graph.W-1))
		bounds := layout.row(1)[0]
		bounds.X += 60
		bounds.Width -= 90
		global.SelectedColumn = uint8(slider(bounds, "Column z", strconv.Itoa(int(global.SelectedColumn)), float32(global.SelectedColumn), 0, float32(graph.W-1)))
	}

	cells = layout.row(2)
	if button(cells[0], "Fill visible", false) {
		this.fillVisible(graph, byte(global.SelectedColor))
	}
	if button(cells[1], "Clear visible", false) {
		this.fillVisible(graph, 0)
	}
}

func (this *EditPanel) graph() *blob.GraphData {
	graphs := global.Project.Graphs
	if global.SelectedGraph < 0 || global.SelectedGraph >= len(graphs) {
		return nil
	}
	return &graphs[global.SelectedGraph]
}

// fillVisible sets the voxels of the shown layer/column selection.
func (this *EditPanel) fillVisible(graph *blob.GraphData, value byte) {
	index := 0
	for z := range graph.W {
		for y := range graph.H {
			for range graph.L {
				visible := true
				switch global.SelectedLayerState {
				case global.Layer:
					visible = y == int(global.SelectedLayer)
				case global.Column:
					visible = z == int(global.SelectedColumn)
				case global.Precise:
					visible = y == int(global.SelectedLayer) && z == int(global.SelectedColumn)
				}
				if visible {
					graph.Voxels[index] = value
				}
				index++
			}
		}
	}
}

// renderPaletteSelector picks the palette used to show graphs; it returns its colors.
func (this *EditPanel) renderPaletteSelector(layout *rowLayout) string {
	palettes := global.Project.Palettes
	if global.SelectedPalette >= len(palettes) {
		global.SelectedPalette = 0
	}
	name := "(default palette)"
	colors := render.DefaultPalette
	if len(palettes) > 0 {
		name = palettes[global.SelectedPalette].Name
		colors = string(palettes[global.SelectedPalette].Colors[:])
	}

	cells := layout.row(4)
	if button(cells[0], "<", len(palettes) < 2) {
		global.SelectedPalette = (global.SelectedPalette + len(palettes) - 1) % len(palettes)
	}
	cells[1].Width = cells[1].Width*2 + layout.padding
	raygui.Label(cells[1], "Palette: "+name)
	if button(cells[3], ">", len(palettes) < 2) {
		global.SelectedPalette = (global.SelectedPalette + 1) % len(palettes)
	}
	return colors
}

// renderSwatches draws the 16 palette colors; it returns the clicked slot or -1.
func (this *EditPanel) renderSwatches(layout *rowLayout, palette string, selected int) int {
	clicked := -1
	for row := range 2 {
		cells := layout.row(8)
		for column, bounds := range cells {
			slot := row*8 + column
			value := byte(0)
			if slot < len(palette) {
				value = palette[slot]
			}
			raylib.DrawRectangleRec(bounds, raylib.Black)
			raylib.DrawRectangleRec(bounds, common.VoxelColor(value))
			border := style.BorderColor
			if slot == selected {
				border = raylib.White
			}
			raylib.DrawRectangleLinesEx(bounds, 2, border)
			if raylib.IsMouseButtonPressed(raylib.MouseLeftButton) &&
				raylib.CheckCollisionPointRec(raylib.GetMousePosition(), bounds) {
				clicked = slot
			}
		}
	}
	return clicked
}

func (this *EditPanel) renderPalette(layout *rowLayout) {
	project := global.Project
	names := make([]string, len(project.Palettes))
	for i, palette := range project.Palettes {
		names[i] = palette.Name
	}
	if selected := this.list(layout, 5, names, global.SelectedPalette); selected != global.SelectedPalette {
		global.SelectedPalette, this.editing = selected, ""
	}

	cells := layout.row(2)
	if button(cells[0], "Add", false) {
		palette := blob.Palette{Name: uniqueName("Palette", func(name string) bool { return show.FindPalette(project, name) >= 0 })}
		copy(palette.Colors[:], render.DefaultPalette)
		project.Palettes = append(project.Palettes, palette)
		global.SelectedPalette = len(project.Palettes) - 1
	}
	if button(cells[1], "Delete", len(project.Palettes) == 0) {
		project.Palettes = append(project.Palettes[:global.SelectedPalette], project.Palettes[global.SelectedPalette+1:]...)
		global.SelectedPalette = max(0, min(global.SelectedPalette, len(project.Palettes)-1))
		return
	}
	if global.SelectedPalette >= len(project.Palettes) {
		return
	}
	palette := &project.Palettes[global.SelectedPalette]

	layout.label("Name", style.TextColorNormal)
	this.textBox("palette-name", layout.row(1)[0], &palette.Name, nameLength)

	layout.label("Slots (click to edit)", style.TextColorNormal)
	if slot := this.renderSwatches(layout, string(palette.Colors[:]), this.paletteSlot); slot >= 0 {
		this.paletteSlot = slot
	}

	value := palette.Colors[this.paletteSlot]
	layout.label(fmt.Sprintf("Slot %d: 0x%02X (0b%06b)", this.paletteSlot, value, value), style.TextColorNormal)
	channels := [...]struct {
		name  string
		shift byte
	}{{"Red", 4}, {"Green", 2}, {"Blue", 0}}
	for _, channel := range channels {
		bounds := layout.row(1)[0]
		bounds.X += 50
		bounds.Width -= 80
		level := value >> channel.shift & 3
		level = byte(slider(bounds, channel.name, strconv.Itoa(int(level)), float32(level), 0, 3) + 0.5)
		value = value&^(3<<channel.shift) | level<<channel.shift
	}
	palette.Colors[this.paletteSlot] = value
}

func (this *EditPanel) renderCartoon(layout *rowLayout) {
	project := global.Project
	group := this.group
	raygui.ToggleGroup(layout.row(4)[0], "Group 1;Group 2;Group 3;Group 4", &group)
	if group != this.group {
		this.group, this.cartoon, this.editing, this.scroll = group, 0, "", 0
	}
	cartoons := project.Groups[this.group]

	names := make([]string, len(cartoons))
	for i, cartoon := range cartoons {
		names[i] = cartoon.Name + " (" + nameOf(cartoonKinds[:], cartoon.Type) + ")"
		if !cartoon.Enabled {
			names[i] = "[off] " + names[i]
		}
	}
	if selected := this.list(layout, 6, names, this.cartoon); selected != this.cartoon {
		this.cartoon, this.editing = selected, ""
	}

	cells := layout.row(4)
	if button(cells[0], "Add", false) {
		name := uniqueName("Cartoon", func(name string) bool {
			for _, cartoon := range cartoons {
				if cartoon.Name == name {
					return true
				}
			}
			return false
		})
		project.Groups[this.group] = append(cartoons, show.NewCartoon(name, blob.Normal))
		this.cartoon = len(project.Groups[this.group]) - 1
		return
	}
	valid := this.cartoon >= 0 && this.cartoon < len(cartoons)
	if button(cells[1], "Delete", !valid) {
		project.Groups[this.group] = append(cartoons[:this.cartoon], cartoons[this.cartoon+1:]...)
		this.cartoon = max(0, min(this.cartoon, len(project.Groups[this.group])-1))
		return
	}
	if button(cells[2], "Up", !valid || this.cartoon == 0) {
		cartoons[this.cartoon-1], cartoons[this.cartoon] = cartoons[this.cartoon], cartoons[this.cartoon-1]
		this.cartoon--
	}
	if button(cells[3], "Down", !valid || this.cartoon >= len(cartoons)-1) {
		cartoons[this.cartoon+1], cartoons[this.cartoon] = cartoons[this.cartoon], cartoons[this.cartoon+1]
		this.cartoon++
	}
	if !valid {
		return
	}
	cartoon := &cartoons[this.cartoon]

	layout.label("Name", style.TextColorNormal)
	this.textBox("cartoon-name", layout.row(1)[0], &cartoon.Name, projectNameLength)

	cells = layout.row(2)
	cartoon.Enabled = checkBox(
		raylib.NewRectangle(cells[0].X, cells[0].Y, style.TextSize, style.TextSize),
		"Enabled", cartoon.Enabled)
	if cartoon.Type < 0 || cartoon.Type >= len(cartoonKinds) {
		raygui.Label(cells[1], nameOf(cartoonKinds[:], cartoon.Type))
		return
	}
	kind := comboBox(cells[1], "Normal;Script;Audio", cartoon.Type)
	if kind != cartoon.Type {
		cartoon.Type = kind
		if kind == blob.Audio {
			cartoon.Header.FrameCount, cartoon.Header.Loop = 0, 0
		}
	}

	switch cartoon.Type {
	case blob.Normal:
		layout.label("Frames / loops / delay (ms)", style.TextColorNormal)
		cells = layout.row(3)
		this.valueBox("cartoon-frames", cells[0], &cartoon.Header.FrameCount, 1, 32767)
		this.valueBox("cartoon-loops", cells[1], &cartoon.Header.Loop, 1, 65535)
		this.valueBox("cartoon-delay", cells[2], &cartoon.Header.Delay, 1, 65535)
	case blob.Audio:
		layout.label("Duration s (0 = forever) / level mode / delay (ms)", style.TextColorNormal)
		cells = layout.row(3)
		this.valueBox("cartoon-time", cells[0], &cartoon.Header.FrameCount, 0, 32767)
		this.valueBox("cartoon-mode", cells[1], &cartoon.Header.Loop, 0, 65535)
		this.valueBox("cartoon-delay", cells[2], &cartoon.Header.Delay, 1, 65535)
	case blob.Script:
		if len(cartoon.Script) > 0 {
			layout.label(fmt.Sprintf("Compiled script: %d words", len(cartoon.Script)), style.TextColorNormal)
		} else {
			layout.label("Script source only: not compilable yet", raylib.Orange)
		}
		return
	}

	counts := [3]int{}
	for _, element := range cartoon.Elements {
		if element.Kind >= 0 && element.Kind < len(counts) {
			counts[element.Kind] += show.PropCount(element)
		}
	}
	layout.label(fmt.Sprintf("Elements: %d", len(cartoon.Elements)), style.TextColorNormal)
	layout.label(fmt.Sprintf("Props: %d graph, %d ball, %d title", counts[blob.SetGraph], counts[blob.DrawBall], counts[blob.TitleStyle]), style.TextColorNormal)
}

func (this *EditPanel) renderKeys(layout *rowLayout) {
	config := &global.Project.Config
	layout.label("Function of each cube key in this show", style.TextColorNormal)
	for key := range config.Keys {
		cells := layout.row(3)
		raygui.Label(cells[0], "KEY"+strconv.Itoa(key+1))
		value := int32(config.Keys[key])
		previous := value
		id := "key-" + strconv.Itoa(key)
		if raygui.Spinner(cells[1], "", &value, 0, len(keyFunctions)-1, this.editing == id) {
			if this.editing == id {
				this.editing = ""
			} else {
				this.editing = id
			}
		}
		if value != previous {
			config.Keys[key] = byte(max(0, min(int(value), len(keyFunctions)-1)))
		}
		raygui.Label(cells[2], nameOf(keyFunctions[:], int(config.Keys[key])))
	}
	cells := layout.row(1)
	if button(cells[0], "Factory defaults", false) {
		config.Keys = show.DefaultKeys
	}

	layout.separator()
	layout.label(fmt.Sprintf("Displays (cubes): %d", len(config.Displays)), style.TextColorNormal)
	for i, display := range config.Displays {
		l, h, w := display.Size()
		layout.label(fmt.Sprintf("%d: at %d,%d,%d, %dx%dx%d, dir %d", i+1, display.X, display.Y, display.Z, l, h, w, display.Dir), style.TextColorNormal)
	}
}
