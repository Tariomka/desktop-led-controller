package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type layerSelection int32

const (
	all layerSelection = iota
	layer
	column
	precise
)

type EditPanel struct {
	Panel
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	selectedLayer layerSelection
	row           uint8
	column        uint8
	colorSelector bool
	activeColor   int32
	colorChanged  bool
}

func newEditPanel(base Panel) *EditPanel {
	return &EditPanel{
		Panel:   base,
		padding: raylib.NewVector2(10, 10),
	}
}

func (this *EditPanel) Update() {
	this.resize()

	offset := this.Width / 5
	this.itemBounds = raylib.NewRectangle(
		this.X+style.BorderWidth+offset,
		this.Y+style.BorderWidth+this.padding.Y,
		this.Width-2*(style.BorderWidth+offset),
		style.TextSize)

	if this.colorChanged {
		switch this.activeColor {
		case 1:
			global.SelectedColor = common.ColorGreen
		case 2:
			global.SelectedColor = common.ColorBlue
		case 3:
			global.SelectedColor = common.ColorRed
		case 4:
			global.SelectedColor = common.ColorCyan
		case 5:
			global.SelectedColor = common.ColorYellow
		case 6:
			global.SelectedColor = common.ColorViolet
		case 7:
			global.SelectedColor = common.ColorWhite
		default:
			global.SelectedColor = common.ColorOff
		}
		this.colorChanged = false
	}
}

// TODO: add new frame button
// TODO: add frame browser
// TODO: add frame preview?
// TODO: add frame show saving/loading
// TODO: probably in menu panel add sending of frames show via tcp
// TODO: test if frame sending works correctly both here and on RPI
// TODO: ???

func (this *EditPanel) Render() {
	this.renderPanel()
	this.renderLayerSelection()
	this.renderColorSelection()
}

func (this *EditPanel) renderSegmentLine() {
	raylib.DrawLine(
		this.ToInt32().X+int32(style.BorderWidth),
		this.itemBounds.ToInt32().Y,
		this.ToInt32().X+this.ToInt32().Width-int32(style.BorderWidth),
		this.itemBounds.ToInt32().Y,
		style.BorderColor)
	this.itemBounds.Y += style.BorderWidth + this.padding.Y
}

func (this *EditPanel) renderLayerSelection() {
	// TODO: send layer selection values (row + column) to cube renderer. maybe make it global state?
	this.selectedLayer = layerSelection(raygui.ComboBox(
		this.itemBounds,
		"All;Layer;Column;Precise",
		int32(this.selectedLayer)))
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y

	if this.selectedLayer != layer && this.selectedLayer != precise {
		raygui.Disable()
	}
	this.row = uint8(raygui.Slider(
		this.itemBounds,
		"Layer",
		fmt.Sprintf("%d", this.row+1),
		float32(this.row),
		0, 7))
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y

	if this.selectedLayer != column && this.selectedLayer != precise {
		raygui.Disable()
	}
	this.column = uint8(raygui.Slider(
		this.itemBounds,
		"Column",
		fmt.Sprintf("%d", this.column+1),
		float32(this.column),
		0, 7))
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.renderSegmentLine()
}

func (this *EditPanel) renderColorSelection() {
	toggleBounds := raylib.NewRectangle(
		this.itemBounds.X+this.padding.X+style.TextLineSpacing,
		this.itemBounds.Y,
		style.TextSize,
		style.TextSize)
	global.ShouldChangeColor = raygui.CheckBox(
		toggleBounds,
		"Enable Coloring",
		global.ShouldChangeColor)
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	if raygui.DropdownBox(
		this.itemBounds,
		"OFF;GREEN;BLUE;RED;CYAN;YELLOW;VIOLET;WHITE",
		&this.activeColor,
		this.colorSelector) {
		this.colorSelector = !this.colorSelector
		this.colorChanged = true
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.renderSegmentLine()
}
