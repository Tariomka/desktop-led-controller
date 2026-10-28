package component

import (
	"fmt"
	"strconv"

	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	labelColumn = 170
	barRow      = 24
	edgeGrip    = 6
)

var barColors = [...]raylib.Color{
	blob.SetGraph:   raylib.NewColor(0x2f, 0x80, 0xed, 255),
	blob.DrawBall:   raylib.NewColor(0x27, 0xae, 0x60, 255),
	blob.TitleStyle: raylib.NewColor(0x8e, 0x44, 0xad, 255),
}

type dragMode int

const (
	dragNone dragMode = iota
	dragMove
	dragStart
	dragEnd
)

// Timeline shows the props of the selected cartoon as bars over the frames.
type Timeline struct {
	Actions *Actions
	Show    *ShowTab

	scroll  float32
	newKind int

	drag        dragMode
	dragElement int
	dragProp    int
	dragOffset  int // frames between the mouse and the bar start when moving
	dragged     bool
}

func (this *Timeline) Draw(bounds raylib.Rectangle) {
	cartoon := global.SelectedCartoon()
	layout := newRowLayout(bounds, Padding)
	if cartoon == nil {
		layout.label("Select a cartoon in the Assets panel to see its timeline.", MutedColor)
		return
	}
	if cartoon.Type == blob.Script {
		layout.label("Script cartoons have no timeline.", MutedColor)
		return
	}

	frames := this.frames(cartoon)
	this.drawToolbar(layout, cartoon, frames)
	area := layout.rowHeight(1, layout.remaining())[0]
	this.drawRows(area, cartoon, frames)
}

func (this *Timeline) frames(cartoon *show.Cartoon) int {
	if cartoon.Type == blob.Audio {
		return 64
	}
	return max(cartoon.Header.FrameCount, 1)
}

func (this *Timeline) drawToolbar(layout *rowLayout, cartoon *show.Cartoon, frames int) {
	cells := layout.split(4, 1.2, 1.2, 1.2, 1.2, 1.2, 1.4)
	text := fmt.Sprintf("%s - %d frames, %d ms", cartoon.Name, frames, cartoon.Header.Delay)
	if cartoon.Type == blob.Normal {
		text += fmt.Sprintf(", %d loop(s)", cartoon.Header.Loop)
	}
	if preview := global.Preview; preview != nil && this.previewing(preview) {
		text += fmt.Sprintf(" - frame %d", preview.Frame()+1)
	}
	layoutLabel(cells[0], text)

	this.newKind = comboBox(cells[1], "Graph;Ball;Title", this.newKind)
	if button(cells[2], "+ Element", false) {
		cartoon.Elements = append(cartoon.Elements, show.NewElement(this.newKind, frames))
		global.Selected.Kind, global.Selected.Element, global.Selected.Prop = global.SelectProp, len(cartoon.Elements)-1, 0
		global.MarkDirty()
		this.Actions.ShowTab(TabProps)
	}
	element := global.SelectedElement()
	if button(cells[3], "+ Prop", element == nil) {
		show.AddProp(element, frames)
		global.Selected.Prop = show.PropCount(*element) - 1
		global.MarkDirty()
	}
	if button(cells[4], "Duplicate", element == nil) {
		show.DuplicateProp(element, global.Selected.Prop)
		global.Selected.Prop++
		global.MarkDirty()
	}
	if button(cells[5], "Delete prop", element == nil) {
		show.DeleteProp(element, global.Selected.Prop)
		if show.PropCount(*element) == 0 {
			cartoon.Elements = append(cartoon.Elements[:global.Selected.Element], cartoon.Elements[global.Selected.Element+1:]...)
			global.Selected.Kind = global.SelectCartoon
		} else {
			global.Selected.Prop = min(global.Selected.Prop, show.PropCount(*element)-1)
		}
		global.MarkDirty()
	}
	if button(cells[6], "#131# Preview", false) {
		this.Show.PlaySelected()
	}
}

// previewing reports whether the preview runs the selected cartoon frame by frame.
func (this *Timeline) previewing(preview *show.Preview) bool {
	if preview.Mode != show.ModeCartoon {
		return false
	}
	flat := global.Project.Flatten()
	selected := global.Selected
	return selected.Group < len(flat) && selected.Cartoon < len(flat[selected.Group]) &&
		flat[selected.Group][selected.Cartoon] == preview.Current()
}

type barRowInfo struct {
	element, prop int
	label         string
}

func (this *Timeline) drawRows(area raylib.Rectangle, cartoon *show.Cartoon, frames int) {
	ruler := raylib.NewRectangle(area.X+labelColumn, area.Y, area.Width-labelColumn, 18)
	rowsArea := raylib.NewRectangle(area.X, area.Y+ruler.Height, area.Width, area.Height-ruler.Height)
	scale := ruler.Width / float32(frames)
	frameAt := func(x float32) int { return int((x - ruler.X) / scale) }

	// Ruler with ticks; clicking scrubs the preview.
	raylib.DrawRectangleRec(ruler, HeaderColor)
	step := 1
	for float32(step)*scale < 36 {
		step *= 2
	}
	for frame := 0; frame < frames; frame += step {
		x := ruler.X + float32(frame)*scale
		raylib.DrawLine(int32(x), int32(ruler.Y+ruler.Height-5), int32(x), int32(ruler.Y+ruler.Height), style.BorderColor)
		utils.RenderText(strconv.Itoa(frame), raylib.NewRectangle(x+2, ruler.Y, 40, ruler.Height-4), MutedColor)
	}
	preview := global.Preview
	if clicked(ruler) || (raylib.IsMouseButtonDown(raylib.MouseLeftButton) && mouseIn(ruler) && this.drag == dragNone) {
		if preview == nil || !this.previewing(preview) {
			this.Show.PlaySelected()
			preview = global.Preview
		}
		if preview != nil && this.previewing(preview) {
			preview.SetFrame(max(0, min(frameAt(raylib.GetMousePosition().X), frames-1)))
		}
	}

	// Rows: one per prop, grouped by element.
	var rows []barRowInfo
	for e, element := range cartoon.Elements {
		for p := range show.PropCount(element) {
			label := ""
			if p == 0 {
				label = fmt.Sprintf("E%d %s (%d)", e+1, nameOf(elementKinds[:], element.Kind), show.PropCount(element))
			}
			rows = append(rows, barRowInfo{e, p, label})
		}
		if show.PropCount(element) == 0 {
			rows = append(rows, barRowInfo{e, -1, fmt.Sprintf("E%d %s (empty)", e+1, nameOf(elementKinds[:], element.Kind))})
		}
	}
	content := float32(len(rows)) * barRow
	if mouseIn(rowsArea) {
		this.scroll -= raylib.GetMouseWheelMove() * barRow * 2
	}
	this.scroll = clampf(this.scroll, 0, max(0, content-rowsArea.Height))

	raylib.BeginScissorMode(int32(rowsArea.X), int32(rowsArea.Y), int32(rowsArea.Width), int32(rowsArea.Height))
	mouse := raylib.GetMousePosition()
	y := rowsArea.Y - this.scroll
	for _, row := range rows {
		rowBounds := raylib.NewRectangle(rowsArea.X, y, rowsArea.Width, barRow)
		y += barRow
		if rowBounds.Y+barRow < rowsArea.Y || rowBounds.Y > rowsArea.Y+rowsArea.Height {
			continue
		}
		raylib.DrawLine(int32(rowBounds.X), int32(rowBounds.Y+barRow-1), int32(rowBounds.X+rowBounds.Width), int32(rowBounds.Y+barRow-1), raylib.Fade(style.BorderColor, 0.5))
		if row.label != "" {
			utils.RenderText(row.label, raylib.NewRectangle(rowBounds.X+6, rowBounds.Y, labelColumn-12, barRow), style.TextColorNormal)
		}
		if row.prop < 0 {
			continue
		}
		element := &cartoon.Elements[row.element]
		start, end := show.PropSpan(*element, row.prop)
		bar := raylib.NewRectangle(ruler.X+float32(start)*scale, rowBounds.Y+4, float32(end-start+1)*scale, barRow-8)
		selected := global.Selected.Kind == global.SelectProp && global.Selected.Element == row.element && global.Selected.Prop == row.prop
		fill := barColors[min(element.Kind, len(barColors)-1)]
		if selected {
			fill = raylib.ColorBrightness(fill, 0.3)
		}
		raylib.DrawRectangleRounded(bar, 0.4, 4, fill)
		if selected {
			raylib.DrawRectangleRoundedLinesEx(bar, 0.4, 4, 2, raylib.White)
		}
		utils.RenderText(fmt.Sprintf("%d-%d", start, end), raylib.NewRectangle(bar.X+4, bar.Y, max(bar.Width-8, 10), bar.Height), raylib.White)

		if mouseIn(rowsArea) && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) && raylib.CheckCollisionPointRec(mouse, bar) && this.drag == dragNone {
			global.Selected.Kind, global.Selected.Element, global.Selected.Prop = global.SelectProp, row.element, row.prop
			this.Actions.ShowTab(TabProps)
			this.dragElement, this.dragProp, this.dragged = row.element, row.prop, false
			switch {
			case mouse.X < bar.X+edgeGrip:
				this.drag = dragStart
			case mouse.X > bar.X+bar.Width-edgeGrip:
				this.drag = dragEnd
			default:
				this.drag = dragMove
				this.dragOffset = frameAt(mouse.X) - start
			}
		}
	}
	raylib.EndScissorMode()

	// Frame cursor.
	if preview != nil && this.previewing(preview) {
		x := ruler.X + (float32(preview.Frame())+0.5)*scale
		raylib.DrawLineEx(raylib.NewVector2(x, ruler.Y), raylib.NewVector2(x, area.Y+area.Height), 2, raylib.Gold)
	}

	this.updateDrag(cartoon, frames, frameAt)
}

func (this *Timeline) updateDrag(cartoon *show.Cartoon, frames int, frameAt func(float32) int) {
	if this.drag == dragNone {
		return
	}
	if raylib.IsMouseButtonReleased(raylib.MouseLeftButton) {
		this.drag = dragNone
		if this.dragged {
			global.MarkDirty()
		}
		return
	}
	if this.dragElement >= len(cartoon.Elements) {
		this.drag = dragNone
		return
	}
	element := &cartoon.Elements[this.dragElement]
	start, end := show.PropSpan(*element, this.dragProp)
	frame := max(0, min(frameAt(raylib.GetMousePosition().X), frames-1))
	switch this.drag {
	case dragMove:
		length := end - start
		newStart := max(0, min(frame-this.dragOffset, frames-1-length))
		if newStart != start {
			show.SetPropSpan(element, this.dragProp, newStart, newStart+length)
			this.dragged = true
		}
	case dragStart:
		if frame != start && frame <= end {
			show.SetPropSpan(element, this.dragProp, frame, end)
			this.dragged = true
		}
	case dragEnd:
		if frame != end && frame >= start {
			show.SetPropSpan(element, this.dragProp, start, frame)
			this.dragged = true
		}
	}
}

// BottomPanel holds the Timeline and Console tabs.
type BottomPanel struct {
	Tab      int
	Timeline *Timeline
	Console  *Console
}

const (
	BottomTimeline = iota
	BottomConsole
)

func (this *BottomPanel) Draw(bounds raylib.Rectangle) {
	raylib.DrawRectangleRec(bounds, style.Panel)
	raylib.DrawRectangleLinesEx(bounds, 1, style.BorderColor)
	tabBounds := raylib.NewRectangle(bounds.X+Padding, bounds.Y+Padding+6, 184, ControlHeight()-4)
	if tab := tabs(tabBounds, "Timeline;Console", this.Tab); tab != this.Tab {
		this.Tab = tab
		StopEditing()
	}
	content := raylib.NewRectangle(bounds.X+2*Padding+184, bounds.Y+Padding, bounds.Width-2*Padding-184, bounds.Height-Padding)
	switch this.Tab {
	case BottomTimeline:
		this.Timeline.Draw(content)
	case BottomConsole:
		this.Console.Draw(content)
	}
}
