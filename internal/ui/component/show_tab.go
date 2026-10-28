package component

import (
	"fmt"
	"strconv"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// ShowTab compiles, uploads and previews the project.
type ShowTab struct {
	Actions *Actions
	Assets  *AssetsPanel

	compiled     []byte
	compileError string
	compiledFor  bool // compiled since the last change

	uploading    bool
	uploadAcked  int
	uploadTotal  int
	uploadResult string
	uploadOK     bool
}

// Compile compiles the project and returns whether it succeeded.
func (this *ShowTab) Compile() bool {
	this.compiled, this.compileError = nil, ""
	data, err := global.Project.Compile()
	if err != nil {
		this.compileError = err.Error()
		return false
	}
	this.compiled = data
	return true
}

// Compiled returns the compiled show data (compiling when needed).
func (this *ShowTab) Compiled() []byte {
	if this.compiled == nil && this.compileError == "" {
		this.Compile()
	}
	return this.compiled
}

// Invalidate forgets the compiled data after an edit.
func (this *ShowTab) Invalidate() { this.compiled, this.compileError = nil, "" }

func (this *ShowTab) Draw(layout *rowLayout) {
	this.drawCompile(layout)
	layout.separator()
	this.drawPreview(layout)
}

func (this *ShowTab) drawCompile(layout *rowLayout) {
	layout.header("Show data")
	data := this.Compiled()
	size := len(data)
	bounds := layout.rowHeight(1, 10)[0]
	progressBar(bounds, float32(size), float32(blob.MaxSize))
	if this.compileError != "" {
		layout.label("Compile error: "+this.compileError, ErrorColor)
	} else {
		layout.label(fmt.Sprintf("%d / %d bytes", size, blob.MaxSize), MutedColor)
	}

	connected := global.ConnectionStatus == global.Connected
	cells := layout.row(2)
	if button(cells[0], "Compile", false) {
		if this.Compile() {
			this.Actions.Status(fmt.Sprintf("Compiled: %d bytes", len(this.compiled)), false)
		} else {
			this.Actions.Error("Compile error: " + this.compileError)
		}
	}
	if button(cells[1], "Export .bin", data == nil) {
		global.SendMessage(constants.ServiceShow, models.ExportMessage{Name: global.ProjectName, Data: data})
	}
	cells = layout.row(2)
	if this.uploading {
		if button(cells[0], "Cancel upload", false) {
			global.SendMessage(constants.TCPClient, models.TCPUploadCancelMessage{})
		}
	} else if accentButton(cells[0], "#121#Upload to cube", !connected || data == nil) {
		this.upload()
	}
	if !connected {
		layout.label("Connect in the Cube tab to upload.", MutedColor)
	}
	if this.uploading || this.uploadResult != "" {
		progressBar(layout.rowHeight(1, 12)[0], float32(this.uploadAcked), float32(max(this.uploadTotal, 1)))
		if this.uploading {
			layout.label(fmt.Sprintf("Uploading %d / %d bytes", this.uploadAcked, this.uploadTotal), style.TextColorNormal)
		} else {
			textColor := OKColor
			if !this.uploadOK {
				textColor = ErrorColor
			}
			layout.label("Upload: "+this.uploadResult, textColor)
		}
	}
	layout.label("Uploaded shows play only in Text mode (DIP SW1+SW2 on).", MutedColor)
}

func (this *ShowTab) drawPreview(layout *rowLayout) {
	layout.header("Preview")
	preview := global.Preview
	cartoon := global.SelectedCartoon()
	cells := layout.row(2)
	if preview == nil {
		if accentButton(cells[0], "#131#Play show", false) {
			this.startPreview(func(preview *show.Preview) { preview.PlayShow(0) })
		}
		if button(cells[1], "Play cartoon", cartoon == nil) {
			this.playSelectedCartoon()
		}
		layout.label("The 3-D view shows the selected graph.", MutedColor)
		layout.label("Space plays/pauses, Left/Right step frames.", MutedColor)
		return
	}

	if button(cells[0], "#133#Stop", false) {
		global.Preview = nil
		return
	}
	pause := "#132#Pause"
	if preview.Paused {
		pause = "#131#Resume"
	}
	if button(cells[1], pause, false) {
		preview.TogglePause()
	}

	cells = layout.row(2)
	if toggleButton(cells[0], "Whole show", preview.Mode == show.ModeShow) && preview.Mode != show.ModeShow {
		preview.PlayShow(preview.Current())
	}
	if toggleButton(cells[1], "This cartoon", preview.Mode == show.ModeCartoon) && preview.Mode != show.ModeCartoon {
		preview.PlayCartoon(preview.Current())
	}

	current := preview.Current()
	text := fmt.Sprintf("Cartoon %d / %d", current+1, preview.Count())
	if preview.Mode == show.ModeCartoon {
		text += fmt.Sprintf(" - frame %d / %d", preview.Frame()+1, preview.FrameCount())
	}
	layout.label(text+" - speed "+preview.SpeedText(), style.TextColorNormal)

	cells = layout.row(4)
	if button(cells[0], "#129#", false) {
		this.changeCartoon(-1)
	}
	if button(cells[1], "#134#", false) {
		this.changeCartoon(1)
	}
	if button(cells[2], "#114#", preview.Mode != show.ModeCartoon) {
		preview.Step(-1)
	}
	if button(cells[3], "#115#", preview.Mode != show.ModeCartoon) {
		preview.Step(1)
	}
	cells = layout.row(4)
	if button(cells[0], "Speed -", false) {
		preview.AddSpeed(-1)
	}
	if button(cells[1], "Speed +", false) {
		preview.AddSpeed(1)
	}
	if button(cells[2], "Model", preview.Mode != show.ModeShow) {
		preview.Player.Action(lightshow.ActionPlayModel)
	}
	if button(cells[3], "Replay", preview.Mode != show.ModeShow) {
		preview.Player.Action(lightshow.ActionReplay)
		preview.Restart()
	}

	bounds := layout.split(1, 3)
	layoutLabel(bounds[0], "Audio")
	preview.Level = int(sliderBar(bounds[1], "", strconv.Itoa(preview.Level), float32(preview.Level), 0, 63))

	layout.label("Keys of this show (preview)", MutedColor)
	cells = layout.row(7)
	for key := range 7 {
		if button(cells[key], "K"+strconv.Itoa(key+1), preview.Mode != show.ModeShow) {
			preview.Player.Key(key)
			preview.Restart()
		}
	}
}

func (this *ShowTab) startPreview(start func(*show.Preview)) {
	preview, err := show.NewPreview(global.Project)
	if err != nil {
		this.Actions.Error("Compile error: " + err.Error())
		return
	}
	global.Preview = preview
	start(preview)
}

// playSelectedCartoon previews the selected cartoon alone.
func (this *ShowTab) playSelectedCartoon() {
	selected := global.Selected
	cartoon := global.SelectedCartoon()
	if cartoon == nil {
		return
	}
	if !cartoon.Enabled {
		this.Actions.Status("The cartoon is disabled; enable it to preview it", true)
		return
	}
	index := global.Project.Flatten()[selected.Group][selected.Cartoon]
	if global.Preview == nil {
		this.startPreview(func(preview *show.Preview) { preview.PlayCartoon(index) })
		return
	}
	global.Preview.PlayCartoon(index)
}

// PlaySelected starts or restarts the preview for the selected cartoon (Space in the timeline).
func (this *ShowTab) PlaySelected() { this.playSelectedCartoon() }

func (this *ShowTab) changeCartoon(delta int) {
	preview := global.Preview
	if preview == nil || preview.Count() == 0 {
		return
	}
	target := (preview.Current() + delta + preview.Count()) % preview.Count()
	if preview.Mode == show.ModeShow {
		preview.PlayShow(target)
	} else {
		preview.PlayCartoon(target)
	}
	// Keep the assets selection in step with the previewed cartoon.
	flat := global.Project.Flatten()
	for group := range flat {
		for i, index := range flat[group] {
			if index == target {
				this.Assets.SelectCartoon(group, i)
			}
		}
	}
}

func (this *ShowTab) upload() {
	data := this.Compiled()
	if data == nil {
		this.Actions.Error("Compile error: " + this.compileError)
		return
	}
	this.uploading, this.uploadAcked, this.uploadTotal, this.uploadResult = true, 0, len(data), ""
	global.SendMessage(constants.TCPClient, models.TCPUploadMessage{Data: data})
}

// Handle takes the upload messages from the client.
func (this *ShowTab) Handle(message any) {
	switch message := message.(type) {
	case models.UploadProgressMessage:
		this.uploadAcked, this.uploadTotal = message.Acked, message.Total
	case models.UploadResultMessage:
		this.uploading, this.uploadOK, this.uploadResult = false, message.OK, message.Text
		if message.OK {
			this.uploadAcked = this.uploadTotal
			this.Actions.Status("Upload finished: "+message.Text, false)
		} else {
			this.Actions.Status("Upload failed: "+message.Text, true)
		}
	case models.DisconnectedMessage:
		if this.uploading {
			this.uploading, this.uploadOK, this.uploadResult = false, false, "connection closed"
		}
	}
}
