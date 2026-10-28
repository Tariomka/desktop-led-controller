// Package component holds the UI panels. All of it runs on the UI goroutine.
package component

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	Padding    = 8
	nameLength = 16 // show names: 15 bytes and the terminator
	textLength = 64
)

var (
	AccentColor  = style.Accent
	WarningColor = raylib.Orange
	ErrorColor   = raylib.NewColor(0xe7, 0x4c, 0x3c, 255)
	OKColor      = raylib.NewColor(0x2e, 0xcc, 0x71, 255)
	MutedColor   = style.TextMuted
	HeaderColor  = style.Bar
	SelectColor  = style.Selection
	StripColor   = style.Bar
)

// editing is the id of the text/value control in edit mode; only one at a time.
var editing string

// EditingText reports whether a text or value box has the keyboard, so shortcuts must wait.
func EditingText() bool { return editing != "" }

// StopEditing ends text input (e.g. when a dialog opens).
func StopEditing() { editing = "" }

func toggleEditing(id string) {
	if editing == id {
		editing = ""
	} else {
		editing = id
	}
}

// rowLayout places controls top to bottom in rows of equal-width cells.
type rowLayout struct {
	x, y, width, padding float32
	bottom               float32
}

func newRowLayout(panel raylib.Rectangle, padding float32) *rowLayout {
	return &rowLayout{
		x:       panel.X + padding,
		y:       panel.Y + padding,
		width:   panel.Width - 2*padding,
		padding: padding,
		bottom:  panel.Y + panel.Height,
	}
}

// ControlHeight is the height of buttons and boxes.
func ControlHeight() float32 { return style.TextSize + style.BorderWidth*2 + style.TextPadding*2 }

// row returns count cells of the next row.
func (this *rowLayout) row(count int) []raylib.Rectangle {
	return this.rowHeight(count, ControlHeight())
}

func (this *rowLayout) rowHeight(count int, height float32) []raylib.Rectangle {
	cellWidth := (this.width - this.padding*float32(count-1)) / float32(count)
	cells := make([]raylib.Rectangle, count)
	for i := range cells {
		cells[i] = raylib.NewRectangle(this.x+float32(i)*(cellWidth+this.padding), this.y, cellWidth, height)
	}
	this.y += height + this.padding
	return cells
}

// split returns cells with the given relative widths.
func (this *rowLayout) split(weights ...float32) []raylib.Rectangle {
	total := float32(0)
	for _, weight := range weights {
		total += weight
	}
	available := this.width - this.padding*float32(len(weights)-1)
	cells := make([]raylib.Rectangle, len(weights))
	x := this.x
	for i, weight := range weights {
		width := available * weight / total
		cells[i] = raylib.NewRectangle(x, this.y, width, ControlHeight())
		x += width + this.padding
	}
	this.y += ControlHeight() + this.padding
	return cells
}

func (this *rowLayout) label(text string, textColor color.RGBA) {
	utils.RenderText(text, raylib.NewRectangle(this.x, this.y, this.width, style.TextSize), textColor)
	this.y += style.TextSize + this.padding/2
}

func (this *rowLayout) header(text string) {
	bounds := raylib.NewRectangle(this.x, this.y, this.width, style.TextSize+8)
	raylib.DrawRectangleRec(bounds, HeaderColor)
	utils.RenderText(text, raylib.NewRectangle(this.x+8, this.y, this.width-16, bounds.Height), style.TextBright)
	this.y += bounds.Height + this.padding
}

func (this *rowLayout) separator() {
	raylib.DrawLine(int32(this.x), int32(this.y), int32(this.x+this.width), int32(this.y), style.BorderColor)
	this.y += this.padding
}

func (this *rowLayout) space(height float32) { this.y += height }

// remaining is the height left in the panel.
func (this *rowLayout) remaining() float32 { return this.bottom - this.y - this.padding }

// button draws a button that is disabled when disabled is true.
func button(bounds raylib.Rectangle, text string, disabled bool) bool {
	clicked := false
	utils.RenderWithCondition(func() { clicked = raygui.Button(bounds, text) }, disabled)
	return clicked && !disabled
}

// Button is the exported plain button for other ui packages.
func Button(bounds raylib.Rectangle, text string) bool { return raygui.Button(bounds, text) }

// accentButton is a button drawn with the accent colour (primary action).
func accentButton(bounds raylib.Rectangle, text string, disabled bool) bool {
	base := raygui.GetStyle(raygui.BUTTON, raygui.BASE_COLOR_NORMAL)
	border := raygui.GetStyle(raygui.BUTTON, raygui.BORDER_COLOR_NORMAL)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_NORMAL, raygui.NewColorPropertyValue(AccentColor))
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_NORMAL, raygui.NewColorPropertyValue(AccentColor))
	clicked := button(bounds, text, disabled)
	raygui.SetStyle(raygui.BUTTON, raygui.BASE_COLOR_NORMAL, base)
	raygui.SetStyle(raygui.BUTTON, raygui.BORDER_COLOR_NORMAL, border)
	return clicked
}

// toggleButton is a button that looks pressed while active.
func toggleButton(bounds raylib.Rectangle, text string, active bool) bool {
	if active {
		raylib.DrawRectangleRec(bounds, style.Background)
	} else {
		raylib.DrawRectangleRec(bounds, style.Panel)
	}
	clicked := raygui.LabelButton(bounds, "")
	edge := style.Border
	if active {
		edge = AccentColor
	}
	if mouseIn(bounds) && !active {
		edge = style.ControlEdge
	}
	raylib.DrawRectangleLinesEx(bounds, style.BorderWidth, edge)
	utils.RenderText(text, utils.GetTextBounds(bounds), style.TextColorNormal)
	return clicked
}

// Value-returning wrappers of raygui's pointer-based controls.

// comboBox draws the options as a toggle group filling bounds and returns the active index.
func comboBox(bounds raylib.Rectangle, text string, active int) int {
	count := strings.Count(text, ";") + 1
	gap := float32(raygui.GetStyle(raygui.TOGGLE, raygui.GROUP_PADDING))
	cell := bounds
	cell.Width = (bounds.Width - gap*float32(count-1)) / float32(count)
	value := int32(active)
	raygui.ToggleGroup(cell, text, &value)
	return int(value)
}

func checkBox(bounds raylib.Rectangle, text string, checked bool) bool {
	raygui.CheckBox(raylib.NewRectangle(bounds.X, bounds.Y+(bounds.Height-style.TextSize)/2, style.TextSize, style.TextSize), text, &checked)
	return checked
}

func slider(bounds raylib.Rectangle, textLeft, textRight string, value, minimum, maximum float32) float32 {
	raygui.Slider(bounds, textLeft, textRight, &value, minimum, maximum)
	return value
}

func sliderBar(bounds raylib.Rectangle, textLeft, textRight string, value, minimum, maximum float32) float32 {
	raygui.SliderBar(bounds, textLeft, textRight, &value, minimum, maximum)
	return value
}

func progressBar(bounds raylib.Rectangle, value, maximum float32) {
	raygui.ProgressBar(bounds, "", "", &value, 0, maximum)
}

// textBox edits a string; it reports whether the text changed this frame.
func textBox(id string, bounds raylib.Rectangle, text *string, size int) bool {
	before := *text
	if raygui.TextBox(bounds, text, size, editing == id) {
		toggleEditing(id)
	}
	return *text != before
}

// valueBox edits an int; it reports whether the value changed this frame.
func valueBox(id string, bounds raylib.Rectangle, value *int, minimum, maximum int) bool {
	number := int32(*value)
	if raygui.ValueBox(bounds, "", &number, minimum, maximum, editing == id) {
		toggleEditing(id)
	}
	changed := int(number) != *value
	*value = int(number)
	return changed
}

func spinner(id string, bounds raylib.Rectangle, value *int, minimum, maximum int) bool {
	number := int32(*value)
	if raygui.Spinner(bounds, "", &number, minimum, maximum, editing == id) {
		toggleEditing(id)
	}
	changed := int(number) != *value
	*value = int(number)
	return changed
}

// tabs draws a toggle group filling bounds and returns the active index.
func tabs(bounds raylib.Rectangle, names string, active int) int {
	return comboBox(bounds, names, active)
}

// swatches draws 16 palette colours in two rows; it returns the clicked slot or -1.
func swatches(layout *rowLayout, palette string, selected int) int {
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
			utils.RenderText(strconv.Itoa(slot), bounds, raylib.Fade(raylib.White, 0.6))
			if raylib.IsMouseButtonPressed(raylib.MouseLeftButton) &&
				raylib.CheckCollisionPointRec(raylib.GetMousePosition(), bounds) {
				clicked = slot
			}
		}
	}
	return clicked
}

// colorPicker edits a 6-bit 0bRRGGBB value with three level sliders.
func colorPicker(layout *rowLayout, id string, value byte) byte {
	channels := [...]struct {
		name  string
		shift byte
	}{{"R", 4}, {"G", 2}, {"B", 0}}
	cells := layout.split(1, 1, 1, 1)
	for i, channel := range channels {
		bounds := cells[i]
		bounds.X += 14
		bounds.Width -= 28
		level := value >> channel.shift & 3
		level = byte(slider(bounds, channel.name, strconv.Itoa(int(level)), float32(level), 0, 3) + 0.5)
		value = value&^(3<<channel.shift) | level<<channel.shift
	}
	raylib.DrawRectangleRec(cells[3], raylib.Black)
	raylib.DrawRectangleRec(cells[3], common.VoxelColor(value))
	raylib.DrawRectangleLinesEx(cells[3], style.BorderWidth, style.BorderColor)
	_ = id
	return value
}

// list draws a scrolling list of names and returns the clicked index, or current.
func list(bounds raylib.Rectangle, names []string, current int, scroll *int32) int {
	text := ""
	for i, name := range names {
		if i > 0 {
			text += ";"
		}
		text += name
	}
	active := int32(current)
	raygui.ListView(bounds, text, scroll, &active)
	if active < 0 || int(active) >= len(names) {
		return current
	}
	return int(active)
}

func mouseIn(bounds raylib.Rectangle) bool {
	return raylib.CheckCollisionPointRec(raylib.GetMousePosition(), bounds)
}

// layoutLabel draws a text vertically centred in bounds.
func layoutLabel(bounds raylib.Rectangle, text string) {
	utils.RenderText(text, bounds, style.TextColorNormal)
}

func clicked(bounds raylib.Rectangle) bool {
	return raylib.IsMouseButtonPressed(raylib.MouseLeftButton) && mouseIn(bounds)
}
