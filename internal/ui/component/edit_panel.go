package component

import (
	"fmt"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type EditPanel struct {
	Panel
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

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

// TODO: add frame preview?
// TODO: probably in menu panel add sending of frames show via tcp
// TODO: test if frame sending works correctly both here and on RPI
// TODO: ???

func (this *EditPanel) Render() {
	this.renderPanel()
	this.renderLayerSelection()
	this.renderColorSelection()
	this.renderFramePreview()
	this.renderFrameControl()
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
	global.SelectedLayerState = global.LayerState(raygui.ComboBox(
		this.itemBounds,
		"All;Layer;Column;Precise",
		int32(global.SelectedLayerState)))
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y

	if global.SelectedLayerState != global.Layer && global.SelectedLayerState != global.Precise {
		raygui.Disable()
	}
	global.SelectedLayer = uint8(raygui.Slider(
		this.itemBounds,
		"Layer",
		fmt.Sprintf("%d", global.SelectedLayer+1),
		float32(global.SelectedLayer),
		0, 7)) // TODO: need to get cube size instead of hardcode
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y

	if global.SelectedLayerState != global.Column && global.SelectedLayerState != global.Precise {
		raygui.Disable()
	}
	global.SelectedColumn = uint8(raygui.Slider(
		this.itemBounds,
		"Column",
		fmt.Sprintf("%d", global.SelectedColumn+1),
		float32(global.SelectedColumn),
		0, 7))
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.renderSegmentLine()
}

func (this *EditPanel) renderColorSelection() {
	toggleBounds := raylib.NewRectangle(
		this.X+this.padding.X*2+style.TextLineSpacing,
		this.itemBounds.Y,
		style.TextSize+style.BorderWidth*2,
		style.TextSize+style.BorderWidth*2)
	global.ShouldChangeColor = raygui.CheckBox(
		toggleBounds,
		"Enable Coloring",
		global.ShouldChangeColor)

	buttonWidth := (this.Width - this.padding.X*4) / 3
	reverseOffsetX := this.X + this.Width - (this.padding.X*2 + style.TextLineSpacing + buttonWidth)
	fillButtonBounds := raylib.NewRectangle(
		reverseOffsetX,
		toggleBounds.Y,
		buttonWidth,
		toggleBounds.Height)
	if !global.ShouldChangeColor {
		raygui.Disable()
	}
	if raygui.Button(fillButtonBounds, "Fill-in") {
		global.SendMessage(constants.UICubeGrid, models.FillVisibleCubesMessage{})
	}
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	if raygui.DropdownBox(
		this.itemBounds,
		"OFF;GREEN;BLUE;RED;CYAN;YELLOW;VIOLET;WHITE", // TODO: move to a config, to be able to switch between mono and RGB
		&this.activeColor,
		this.colorSelector) {
		this.colorSelector = !this.colorSelector
		this.colorChanged = true
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.renderSegmentLine()
}

func (this *EditPanel) renderFramePreview() {
	this.itemBounds.Y += (this.itemBounds.Height + style.BorderWidth + this.padding.Y) * 2

	utils.RenderText(
		fmt.Sprintf(
			"Currently viewing %d/%d frame",
			global.SelectedFrame+1,
			global.TotalFrameCount+1),
		utils.GetTextBounds(this.itemBounds),
		style.TextColorNormal)
	this.itemBounds.Y += (this.itemBounds.Height + style.BorderWidth + this.padding.Y) * 2

	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.renderSegmentLine()
}

func (this *EditPanel) renderFrameControl() {
	// TODO: PREVIOUS and NEXT FRAME
	buttonWidth := (this.Width - this.padding.X*4) / 2

	buttonBounds := raylib.NewRectangle(
		this.X+this.padding.X*1.5,
		this.itemBounds.Y,
		buttonWidth,
		style.TextSize+style.BorderWidth*2)

	if global.SelectedFrame <= 0 {
		raygui.Disable()
	}
	if raygui.Button(buttonBounds, "Previous Frame") {
		global.SendMessage(
			constants.ServiceLedProcessor,
			models.LoadFrameMessage{Index: global.SelectedFrame - 1})
	}
	raygui.Enable()

	buttonBounds.X += buttonWidth + this.padding.X
	if global.SelectedFrame >= global.TotalFrameCount {
		raygui.Disable()
	}
	if raygui.Button(buttonBounds, "Next Frame") {
		global.SendMessage(
			constants.ServiceLedProcessor,
			models.LoadFrameMessage{Index: global.SelectedFrame + 1})
	}
	raygui.Enable()
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	buttonWidth = (this.Width - this.padding.X*4) / 3
	buttonBounds = raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		buttonWidth,
		buttonBounds.Height)
	if raygui.Button(buttonBounds, "Reset") {
		global.SendMessage(constants.UICubeGrid, models.ResetMessage{})
		global.SendMessage(constants.ServiceLedProcessor, models.ResetMessage{})
	}

	buttonBounds.X += buttonWidth + this.padding.X
	if raygui.Button(buttonBounds, "New Frame") {
		global.SendMessage(constants.UICubeGrid, models.SaveMessage{})
	}

	buttonBounds.X += buttonWidth + this.padding.X
	if raygui.Button(buttonBounds, "Save") {
		global.SendMessage(constants.ServiceLedProcessor, models.SaveMessage{})
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	this.renderSegmentLine()
}
