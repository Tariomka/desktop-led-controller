package component

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

func getRGBNamedColors() []data.Tuple[string, color.RGBA] {
	return []data.Tuple[string, color.RGBA]{
		data.NewTuple("OFF", common.ColorOff),
		data.NewTuple("GREEN", common.ColorGreen),
		data.NewTuple("BLUE", common.ColorBlue),
		data.NewTuple("RED", common.ColorRed),
		data.NewTuple("CYAN", common.ColorCyan),
		data.NewTuple("YELLOW", common.ColorYellow),
		data.NewTuple("VIOLET", common.ColorViolet),
		data.NewTuple("WHITE", common.ColorWhite),
	}
}

type EditPanel struct {
	Panel
	padding    raylib.Vector2
	itemBounds raylib.Rectangle

	colorData    *data.DropdownData[color.RGBA]
	colorChanged bool

	lightShowData *data.DropdownData[string]

	channel chan any
}

func newEditPanel(base Panel) *EditPanel {
	editPanel := &EditPanel{
		Panel:         base,
		padding:       raylib.NewVector2(10, 10),
		colorData:     data.NewDropdownDataWithValues(getRGBNamedColors()...), // TODO: make it more dynamic
		lightShowData: data.NewDropdownData[string](),
		channel:       make(chan any, 1),
	}

	go editPanel.channelLoop()
	global.RegisterMessageReceiver(
		constants.UIEditPanel,
		func(message any) { editPanel.channel <- message })
	return editPanel
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
		colorValue, _ := this.colorData.GetSelectedValue()
		global.SelectedColor = colorValue
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
	this.renderNetworkControl()
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

	utils.RenderWithCondition(
		func() {
			global.SelectedLayer = uint8(raygui.Slider(
				this.itemBounds,
				"Layer",
				fmt.Sprintf("%d", global.SelectedLayer+1),
				float32(global.SelectedLayer),
				0, 7)) // TODO: need to get cube size instead of hardcode
		},
		global.SelectedLayerState != global.Layer && global.SelectedLayerState != global.Precise)
	this.itemBounds.Y += this.itemBounds.Height + this.padding.Y

	utils.RenderWithCondition(
		func() {
			global.SelectedColumn = uint8(raygui.Slider(
				this.itemBounds,
				"Column",
				fmt.Sprintf("%d", global.SelectedColumn+1),
				float32(global.SelectedColumn),
				0, 7))
		},
		global.SelectedLayerState != global.Column && global.SelectedLayerState != global.Precise)
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
	utils.RenderWithCondition(
		func() {
			if raygui.Button(fillButtonBounds, "Fill-in") {
				global.SendMessage(constants.UICubeGrid, models.FillVisibleCubesMessage{})
			}
		},
		!global.ShouldChangeColor)
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	if raygui.DropdownBox(
		this.itemBounds,
		this.colorData.GetText(),
		this.colorData.GetIndex(),
		this.colorData.IsActive()) {
		this.colorData.SwitchActive()
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
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth + this.padding.Y

	this.itemBounds.Y += (this.itemBounds.Height + style.BorderWidth + this.padding.Y) * 2

	this.renderSegmentLine()
}

func (this *EditPanel) renderFrameControl() {
	buttonWidth := (this.Width - this.padding.X*3) / 2
	localItemBounds := raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		buttonWidth,
		style.TextSize+style.BorderWidth*2)

	utils.RenderWithCondition(
		func() {
			if raygui.Button(localItemBounds, "Previous Frame") {
				global.SendMessage(
					constants.ServiceLedProcessor,
					models.RenderFrameMessage{Index: global.SelectedFrame - 1})
			}
		},
		global.SelectedFrame <= 0)
	localItemBounds.X += localItemBounds.Width + this.padding.X

	utils.RenderWithCondition(
		func() {
			if raygui.Button(localItemBounds, "Next Frame") {
				global.SendMessage(
					constants.ServiceLedProcessor,
					models.RenderFrameMessage{Index: global.SelectedFrame + 1})
			}
		},
		global.SelectedFrame >= global.TotalFrameCount)
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	// --- New line ---
	buttonWidth = (this.Width - this.padding.X*4) / 3
	localItemBounds = raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		buttonWidth,
		localItemBounds.Height)
	if raygui.Button(localItemBounds, "Reset") {
		global.SendMessage(constants.UICubeGrid, models.ResetMessage{})
		global.SendMessage(constants.ServiceLedProcessor, models.ResetMessage{})
	}
	localItemBounds.X += localItemBounds.Width + this.padding.X

	utils.RenderWithCondition(
		func() {
			if raygui.Button(localItemBounds, "Remove Frame") {
				global.SendMessage(
					constants.ServiceLedProcessor,
					models.RemoveFrameMessage{Index: global.SelectedFrame})
			}
		},
		global.SelectedLayer == uint8(global.TotalFrameCount))
	localItemBounds.X += localItemBounds.Width + this.padding.X

	if raygui.Button(localItemBounds, "Save Frame") {
		global.SendMessage(constants.UICubeGrid, models.SaveMessage{})
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	// --- New line ---
	buttonWidth = (this.Width - this.padding.X*3) / 2
	localItemBounds = raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		buttonWidth,
		localItemBounds.Height)

	localItemBounds.X += (localItemBounds.Width + this.padding.X) / 2
	if raygui.Button(localItemBounds, "Fetch Light Shows") {
		global.SendMessage(constants.ServiceLedProcessor, models.FetchMessage{})
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	// --- New line ---
	selectionWidth := (this.Width - this.padding.X*3) / 3 * 2
	localItemBounds = raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		selectionWidth,
		localItemBounds.Height)

	utils.RenderWithCondition(
		func() {
			if raygui.DropdownBox(
				localItemBounds,
				this.lightShowData.GetText(),
				this.lightShowData.GetIndex(),
				this.lightShowData.IsActive()) {
				this.lightShowData.SwitchActive()
			}
		},
		this.lightShowData.IsEmpty())
	localItemBounds.X += localItemBounds.Width + this.padding.X

	buttonWidth = (this.Width - this.padding.X*4) / 3
	localItemBounds.Width = buttonWidth
	utils.RenderWithCondition(
		func() {
			if raygui.Button(localItemBounds, "Load") {
				lightShowName, _ := this.lightShowData.GetSelectedValue()
				global.SendMessage(constants.ServiceLedProcessor, models.LoadMessage{Name: lightShowName})
			}
		},
		this.lightShowData.IsEmpty() || !this.lightShowData.IsSelected())
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	// --- New line ---
	localItemBounds = raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		selectionWidth,
		localItemBounds.Height)

	localItemBounds.X += localItemBounds.Width + this.padding.X
	localItemBounds.Width = buttonWidth
	if raygui.Button(localItemBounds, "Save") {
		global.SendMessage(constants.ServiceLedProcessor, models.SaveMessage{})
	}
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	this.renderSegmentLine()
}

func (this *EditPanel) renderNetworkControl() {
	buttonWidth := (this.Width - this.padding.X*3) / 2
	buttonBounds := raylib.NewRectangle(
		this.X+this.padding.X,
		this.itemBounds.Y,
		buttonWidth,
		style.TextSize+style.BorderWidth*2)

	buttonBounds.X += (buttonBounds.Width + this.padding.X) / 2
	utils.RenderWithCondition(
		func() {
			if raygui.Button(buttonBounds, "Send Light Show") {
				global.SendMessage(constants.ServiceLedProcessor, models.SendMessage{})
			}
		},
		global.ConnectionStatus != global.Connected)
	this.itemBounds.Y += this.itemBounds.Height + style.BorderWidth*3 + this.padding.Y

	this.renderSegmentLine()
}

// Blocking message loop
func (this *EditPanel) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.SetLightShowsMessage:
			var pairs []data.Tuple[string, string]
			for _, name := range message.Names {
				pairs = append(pairs, data.NewTuple(strings.ToUpper(name), name))
			}
			this.lightShowData.SetData(pairs...)
		}
	}
}
