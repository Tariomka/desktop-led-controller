package component

import (
	"image/color"

	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// rowLayout places controls top to bottom in rows of equal-width cells.
type rowLayout struct {
	x, y, width, padding float32
}

func newRowLayout(panel raylib.Rectangle, padding float32) *rowLayout {
	return &rowLayout{
		x:       panel.X + padding,
		y:       panel.Y + padding,
		width:   panel.Width - 2*padding,
		padding: padding,
	}
}

func controlHeight() float32 { return style.TextSize + style.BorderWidth*2 + style.TextPadding*2 }

// row returns count cells of the next row.
func (this *rowLayout) row(count int) []raylib.Rectangle {
	return this.rowHeight(count, controlHeight())
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

func (this *rowLayout) label(text string, textColor color.RGBA) {
	utils.RenderText(text, raylib.NewRectangle(this.x, this.y, this.width, style.TextSize), textColor)
	this.y += style.TextSize + this.padding/2
}

func (this *rowLayout) separator() {
	raylib.DrawLine(
		int32(this.x), int32(this.y),
		int32(this.x+this.width), int32(this.y),
		style.BorderColor)
	this.y += this.padding
}

// button draws a button that is disabled when disabled is true.
func button(bounds raylib.Rectangle, text string, disabled bool) bool {
	clicked := false
	utils.RenderWithCondition(func() { clicked = raygui.Button(bounds, text) }, disabled)
	return clicked && !disabled
}

// Value-returning wrappers of raygui's pointer-based controls.

func comboBox(bounds raylib.Rectangle, text string, active int) int {
	value := int32(active)
	raygui.ComboBox(bounds, text, &value)
	return int(value)
}

func checkBox(bounds raylib.Rectangle, text string, checked bool) bool {
	raygui.CheckBox(bounds, text, &checked)
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
