package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	sectionGraphs = iota
	sectionPalettes
	sectionPaths
	sectionCartoons
	sectionKeys
	sectionDisplays
	sectionCount
)

var sectionNames = [sectionCount]string{"Graphs", "Palettes", "Paths", "Cartoons", "Keys", "Displays"}

type assetRow struct {
	text      string
	header    int // section index for header rows, -1 otherwise
	selection global.Selection
	selected  bool
	muted     bool
}

// AssetsPanel is the tree of everything in the project; clicking selects for the Props tab.
type AssetsPanel struct {
	collapsed [sectionCount]bool
	group     int
	scroll    float32
	Actions   *Actions
}

func (this *AssetsPanel) Draw(bounds raylib.Rectangle) {
	raylib.DrawRectangleRec(bounds, style.BackgroundColor)
	raylib.DrawRectangleLinesEx(bounds, 1, style.BorderColor)
	layout := newRowLayout(bounds, Padding)
	utils.RenderText("Assets", raylib.NewRectangle(layout.x, layout.y, layout.width-24, style.TextSize), raylib.White)
	layout.space(style.TextSize + Padding)

	buttons := layout.remaining() - 2*ControlHeight() - 2*layout.padding
	listBounds := layout.rowHeight(1, buttons)[0]
	this.drawTree(listBounds)
	this.drawButtons(layout)
}

func (this *AssetsPanel) rows() []assetRow {
	project := global.Project
	selected := global.Selected
	var rows []assetRow
	header := func(section int, count string) {
		arrow := "v "
		if this.collapsed[section] {
			arrow = "> "
		}
		rows = append(rows, assetRow{text: arrow + sectionNames[section] + count, header: section})
	}
	item := func(text string, selection global.Selection, muted bool) {
		rows = append(rows, assetRow{text: "    " + text, header: -1, selection: selection, muted: muted,
			selected: selection.Kind == selected.Kind && selection.Index == selected.Index &&
				selection.Group == selected.Group && selection.Cartoon == selected.Cartoon &&
				(selection.Kind != global.SelectProp || selected.Kind == global.SelectProp)})
	}

	header(sectionGraphs, fmt.Sprintf(" (%d)", len(project.Graphs)))
	if !this.collapsed[sectionGraphs] {
		for i, graph := range project.Graphs {
			item(fmt.Sprintf("%s  %dx%dx%d", graph.Name, graph.L, graph.H, graph.W), global.Selection{Kind: global.SelectGraph, Index: i}, false)
		}
	}
	header(sectionPalettes, fmt.Sprintf(" (%d)", len(project.Palettes)))
	if !this.collapsed[sectionPalettes] {
		for i, palette := range project.Palettes {
			item(palette.Name, global.Selection{Kind: global.SelectPalette, Index: i}, false)
		}
	}
	header(sectionPaths, fmt.Sprintf(" (%d)", len(project.Indexes)))
	if !this.collapsed[sectionPaths] {
		for i, index := range project.Indexes {
			item(fmt.Sprintf("%s  %d points", index.Name, len(index.Data)/4), global.Selection{Kind: global.SelectPath, Index: i}, false)
		}
	}
	header(sectionCartoons, fmt.Sprintf(" - group %d (%d)", this.group+1, len(project.Groups[this.group])))
	if !this.collapsed[sectionCartoons] {
		for i, cartoon := range project.Groups[this.group] {
			text := fmt.Sprintf("%d  %s", i+1, cartoon.Name)
			if !cartoon.Enabled {
				text += "  (off)"
			}
			// A prop selection keeps its cartoon highlighted.
			selection := global.Selection{Kind: global.SelectCartoon, Group: this.group, Cartoon: i}
			row := assetRow{text: "    " + text, header: -1, selection: selection, muted: !cartoon.Enabled,
				selected: (selected.Kind == global.SelectCartoon || selected.Kind == global.SelectProp) &&
					selected.Group == this.group && selected.Cartoon == i}
			rows = append(rows, row)
		}
	}
	header(sectionKeys, "")
	if !this.collapsed[sectionKeys] {
		item("KEY1-KEY7 functions", global.Selection{Kind: global.SelectKeys}, false)
	}
	header(sectionDisplays, fmt.Sprintf(" (%d)", len(project.Config.Displays)))
	if !this.collapsed[sectionDisplays] {
		for i, display := range project.Config.Displays {
			l, h, w := display.Size()
			item(fmt.Sprintf("%d  at %d,%d,%d  %dx%dx%d", i+1, display.X, display.Y, display.Z, l, h, w), global.Selection{Kind: global.SelectDisplay, Index: i}, false)
		}
	}
	return rows
}

func (this *AssetsPanel) drawTree(bounds raylib.Rectangle) {
	rows := this.rows()
	rowHeight := style.TextSize + 6
	content := float32(len(rows)) * rowHeight
	if mouseIn(bounds) {
		this.scroll -= raylib.GetMouseWheelMove() * rowHeight * 3
	}
	this.scroll = clampf(this.scroll, 0, max(0, content-bounds.Height))

	raylib.BeginScissorMode(int32(bounds.X), int32(bounds.Y), int32(bounds.Width), int32(bounds.Height))
	y := bounds.Y - this.scroll
	for _, row := range rows {
		rowBounds := raylib.NewRectangle(bounds.X, y, bounds.Width, rowHeight)
		y += rowHeight
		if rowBounds.Y+rowHeight < bounds.Y || rowBounds.Y > bounds.Y+bounds.Height {
			continue
		}
		hover := mouseIn(bounds) && mouseIn(rowBounds)
		switch {
		case row.header >= 0:
			raylib.DrawRectangleRec(rowBounds, HeaderColor)
		case row.selected:
			raylib.DrawRectangleRec(rowBounds, SelectColor)
		case hover:
			raylib.DrawRectangleRec(rowBounds, raylib.Fade(SelectColor, 0.4))
		}
		textColor := style.TextColorNormal
		if row.muted {
			textColor = MutedColor
		}
		if row.header >= 0 {
			textColor = raylib.White
		}
		utils.RenderText(row.text, raylib.NewRectangle(rowBounds.X+6, rowBounds.Y, rowBounds.Width-12, rowBounds.Height), textColor)

		if hover && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
			if row.header >= 0 {
				this.collapsed[row.header] = !this.collapsed[row.header]
			} else {
				this.selectRow(row.selection)
			}
		}
	}
	raylib.EndScissorMode()

	// Group selector drawn over the cartoons header's right side.
	for i, row := range rows {
		if row.header != sectionCartoons {
			continue
		}
		rowY := bounds.Y - this.scroll + float32(i)*rowHeight
		if rowY < bounds.Y || rowY+rowHeight > bounds.Y+bounds.Height {
			break
		}
		x := bounds.X + bounds.Width - 4*(rowHeight+2) - 4
		for group := range blob.GroupCount {
			cell := raylib.NewRectangle(x+float32(group)*(rowHeight+2), rowY+1, rowHeight, rowHeight-2)
			if toggleButton(cell, fmt.Sprint(group+1), group == this.group) && group != this.group {
				this.group = group
				this.collapsed[sectionCartoons] = false
			}
		}
		break
	}
}

func (this *AssetsPanel) selectRow(selection global.Selection) {
	global.Selected = selection
	switch selection.Kind {
	case global.SelectPalette:
		global.SelectedPalette = selection.Index
	case global.SelectCartoon:
		this.group = selection.Group
	}
	if this.Actions != nil && this.Actions.ShowTab != nil {
		this.Actions.ShowTab(TabProps)
	}
}

// SelectCartoon selects a cartoon by group and index (used by the Show tab).
func (this *AssetsPanel) SelectCartoon(group, index int) {
	this.group = group
	global.Selected = global.Selection{Kind: global.SelectCartoon, Group: group, Cartoon: index}
}

func (this *AssetsPanel) drawButtons(layout *rowLayout) {
	project := global.Project
	selected := global.Selected
	cells := layout.row(3)
	if button(cells[0], "#8#Add", false) {
		this.add()
	}
	canCopy := selected.Kind == global.SelectGraph || selected.Kind == global.SelectPalette ||
		selected.Kind == global.SelectPath || selected.Kind == global.SelectCartoon || selected.Kind == global.SelectProp
	if button(cells[1], "Duplicate", !canCopy) {
		this.duplicate()
	}
	canDelete := canCopy || (selected.Kind == global.SelectDisplay && len(project.Config.Displays) > 1)
	if button(cells[2], "Delete", !canDelete) {
		this.delete()
	}
	cells = layout.row(2)
	if button(cells[0], "Move up", !canCopy) {
		this.move(-1)
	}
	if button(cells[1], "Move down", !canCopy) {
		this.move(1)
	}
}

func (this *AssetsPanel) add() {
	project := global.Project
	switch global.Selected.Kind {
	case global.SelectPalette:
		name := show.UniqueName("Palette", func(name string) bool { return show.FindPalette(project, name) >= 0 })
		palette := blob.Palette{Name: name}
		copy(palette.Colors[:], defaultPalette())
		project.Palettes = append(project.Palettes, palette)
		global.Selected = global.Selection{Kind: global.SelectPalette, Index: len(project.Palettes) - 1}
		global.SelectedPalette = global.Selected.Index
	case global.SelectPath:
		name := show.UniqueName("Path", func(name string) bool { return findIndex(project, name) >= 0 })
		project.Indexes = append(project.Indexes, blob.IndexData{Name: name, Data: show.PathData(show.EdgePath(8, 8))})
		global.Selected = global.Selection{Kind: global.SelectPath, Index: len(project.Indexes) - 1}
	case global.SelectCartoon, global.SelectProp:
		cartoons := &project.Groups[this.group]
		name := show.UniqueName("Cartoon", func(name string) bool {
			for _, cartoon := range *cartoons {
				if cartoon.Name == name {
					return true
				}
			}
			return false
		})
		*cartoons = append(*cartoons, show.NewCartoon(name, blob.Normal))
		global.Selected = global.Selection{Kind: global.SelectCartoon, Group: this.group, Cartoon: len(*cartoons) - 1}
	case global.SelectDisplay:
		project.Config.Displays = append(project.Config.Displays, blob.Display{CubeType: 1})
		global.Selected = global.Selection{Kind: global.SelectDisplay, Index: len(project.Config.Displays) - 1}
	default:
		name := show.UniqueName("Graph", func(name string) bool { return show.FindGraph(project, name) >= 0 })
		project.Graphs = append(project.Graphs, show.NewGraph(name, 8, 8, 8))
		global.Selected = global.Selection{Kind: global.SelectGraph, Index: len(project.Graphs) - 1}
	}
	global.MarkDirty()
}

func (this *AssetsPanel) duplicate() {
	project := global.Project
	selected := global.Selected
	switch selected.Kind {
	case global.SelectGraph:
		if graph := global.SelectedGraph(); graph != nil {
			name := show.UniqueName(graph.Name, func(name string) bool { return show.FindGraph(project, name) >= 0 })
			project.Graphs = append(project.Graphs, show.CopyGraph(*graph, name))
			global.Selected.Index = len(project.Graphs) - 1
		}
	case global.SelectPalette:
		if selected.Index < len(project.Palettes) {
			palette := project.Palettes[selected.Index]
			palette.Name = show.UniqueName(palette.Name, func(name string) bool { return show.FindPalette(project, name) >= 0 })
			project.Palettes = append(project.Palettes, palette)
			global.Selected.Index = len(project.Palettes) - 1
		}
	case global.SelectPath:
		if selected.Index < len(project.Indexes) {
			index := project.Indexes[selected.Index]
			index.Name = show.UniqueName(index.Name, func(name string) bool { return findIndex(project, name) >= 0 })
			index.Data = append([]byte(nil), index.Data...)
			project.Indexes = append(project.Indexes, index)
			global.Selected.Index = len(project.Indexes) - 1
		}
	case global.SelectCartoon, global.SelectProp:
		if cartoon := global.SelectedCartoon(); cartoon != nil {
			cartoons := &project.Groups[selected.Group]
			*cartoons = append(*cartoons, show.CopyCartoon(*cartoon, cartoon.Name+" copy"))
			global.Selected = global.Selection{Kind: global.SelectCartoon, Group: selected.Group, Cartoon: len(*cartoons) - 1}
		}
	}
	global.MarkDirty()
}

func (this *AssetsPanel) delete() {
	project := global.Project
	selected := global.Selected
	switch selected.Kind {
	case global.SelectGraph:
		if selected.Index < len(project.Graphs) {
			project.Graphs = append(project.Graphs[:selected.Index], project.Graphs[selected.Index+1:]...)
			global.Selected.Index = max(0, min(selected.Index, len(project.Graphs)-1))
		}
	case global.SelectPalette:
		if selected.Index < len(project.Palettes) {
			project.Palettes = append(project.Palettes[:selected.Index], project.Palettes[selected.Index+1:]...)
			global.Selected.Index = max(0, min(selected.Index, len(project.Palettes)-1))
			global.SelectedPalette = global.Selected.Index
		}
	case global.SelectPath:
		if selected.Index < len(project.Indexes) {
			project.Indexes = append(project.Indexes[:selected.Index], project.Indexes[selected.Index+1:]...)
			global.Selected.Index = max(0, min(selected.Index, len(project.Indexes)-1))
		}
	case global.SelectCartoon, global.SelectProp:
		cartoons := &project.Groups[selected.Group]
		if selected.Cartoon < len(*cartoons) {
			*cartoons = append((*cartoons)[:selected.Cartoon], (*cartoons)[selected.Cartoon+1:]...)
			global.Selected = global.Selection{Kind: global.SelectCartoon, Group: selected.Group, Cartoon: max(0, min(selected.Cartoon, len(*cartoons)-1))}
		}
	case global.SelectDisplay:
		displays := &project.Config.Displays
		if selected.Index < len(*displays) && len(*displays) > 1 {
			*displays = append((*displays)[:selected.Index], (*displays)[selected.Index+1:]...)
			global.Selected.Index = max(0, min(selected.Index, len(*displays)-1))
		}
	}
	global.MarkDirty()
}

func (this *AssetsPanel) move(delta int) {
	project := global.Project
	selected := &global.Selected
	switch selected.Kind {
	case global.SelectGraph:
		if target := selected.Index + delta; target >= 0 && target < len(project.Graphs) {
			show.MoveItem(project.Graphs, selected.Index, target)
			selected.Index = target
		}
	case global.SelectPalette:
		if target := selected.Index + delta; target >= 0 && target < len(project.Palettes) {
			show.MoveItem(project.Palettes, selected.Index, target)
			selected.Index = target
			global.SelectedPalette = target
		}
	case global.SelectPath:
		if target := selected.Index + delta; target >= 0 && target < len(project.Indexes) {
			show.MoveItem(project.Indexes, selected.Index, target)
			selected.Index = target
		}
	case global.SelectCartoon, global.SelectProp:
		cartoons := project.Groups[selected.Group]
		if target := selected.Cartoon + delta; target >= 0 && target < len(cartoons) {
			show.MoveItem(cartoons, selected.Cartoon, target)
			selected.Cartoon = target
		}
	}
	global.MarkDirty()
}

func findIndex(project *show.Project, name string) int {
	for i, index := range project.Indexes {
		if index.Name == name {
			return i
		}
	}
	return -1
}

func clampf(value, low, high float32) float32 { return max(low, min(value, high)) }
