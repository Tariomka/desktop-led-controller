package component

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/lightshow"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const projectNameLength = 64

// ShowPanel manages the project file, compiles and uploads it, and runs the preview.
type ShowPanel struct {
	Panel
	padding float32

	projects      []string
	projectIndex  int32
	nameEdit      bool
	statusText    string
	statusIsError bool

	compiled     []byte
	compileError string

	uploading    bool
	uploadAcked  int
	uploadTotal  int
	uploadResult string
	uploadOK     bool

	channel chan any
}

func newShowPanel(base Panel) *ShowPanel {
	showPanel := &ShowPanel{
		Panel:   base,
		padding: 10,
		channel: make(chan any, 64),
	}

	global.RegisterMessageReceiver(
		constants.UIShowPanel,
		func(message any) { showPanel.channel <- message })
	global.SendMessage(constants.ServiceShow, models.FetchMessage{})
	showPanel.compile()

	return showPanel
}

func (this *ShowPanel) Update() {
	this.resize()
	for {
		select {
		case message := <-this.channel:
			this.handle(message)
		default:
			return
		}
	}
}

func (this *ShowPanel) Render() {
	this.renderPanel()
	layout := newRowLayout(this.Rectangle, this.padding)
	this.renderProject(layout)
	layout.separator()
	this.renderUpload(layout)
	layout.separator()
	this.renderPreview(layout)
}

func (this *ShowPanel) renderProject(layout *rowLayout) {
	layout.label("Project name", style.TextColorNormal)
	if raygui.TextBox(layout.row(1)[0], &global.ProjectName, projectNameLength, this.nameEdit) {
		this.nameEdit = !this.nameEdit
	}

	cells := layout.row(2)
	if button(cells[0], "New", false) {
		global.SetProject("untitled", show.NewProject())
		this.compile()
	}
	if button(cells[1], "Save", false) {
		this.save()
	}

	cells = layout.row(3)
	names := strings.Join(this.projects, ";")
	if names == "" {
		names = "(no saved projects)"
	}
	cells[0].Width = cells[0].Width*2 + layout.padding
	raygui.ComboBox(cells[0], names, &this.projectIndex)
	if button(cells[2], "Load", len(this.projects) == 0) {
		global.SendMessage(constants.ServiceShow, models.LoadMessage{Name: this.projects[this.projectIndex]})
	}

	layout.label("Drop .ls, .gc, .hex or .bin files", style.TextColorNormal)
	layout.label("on the window to open them.", style.TextColorNormal)
	if this.statusText != "" {
		layout.label(this.statusText, statusColor(this.statusIsError))
	}
}

func (this *ShowPanel) renderUpload(layout *rowLayout) {
	cells := layout.row(2)
	if button(cells[0], "Compile", false) {
		this.compile()
	}
	connected := global.ConnectionStatus == global.Connected
	if button(cells[1], "Upload", !connected || this.uploading) {
		this.upload()
	}

	if this.compileError != "" {
		layout.label(this.compileError, raylib.Red)
	} else {
		layout.label(fmt.Sprintf("Compiled: %d / %d bytes", len(this.compiled), blob.MaxSize), style.TextColorNormal)
	}

	if this.uploading || this.uploadResult != "" {
		bounds := layout.row(1)[0]
		progressBar(bounds, float32(this.uploadAcked), float32(max(this.uploadTotal, 1)))
		if this.uploading {
			layout.label(fmt.Sprintf("Uploading %d / %d bytes", this.uploadAcked, this.uploadTotal), style.TextColorNormal)
		} else {
			layout.label("Upload: "+this.uploadResult, statusColor(!this.uploadOK))
		}
	}
}

func (this *ShowPanel) renderPreview(layout *rowLayout) {
	preview := global.Preview
	cells := layout.row(2)
	if preview == nil {
		if button(cells[0], "Play preview", false) {
			this.startPreview()
		}
		layout.label("The 3-D view shows the selected graph.", style.TextColorNormal)
		return
	}
	if button(cells[0], "Stop preview", false) {
		global.Preview = nil
		return
	}

	player := preview.Player
	pause := "Pause"
	if player.Paused {
		pause = "Resume"
	}
	if button(cells[1], pause, false) {
		player.Action(lightshow.ActionPause)
		preview.Restart()
	}

	layout.label(fmt.Sprintf("Cartoon %d / %d, speed %s", player.Current()+1, player.Count(), player.SpeedText()), style.TextColorNormal)
	cells = layout.row(4)
	actions := [...]struct {
		text   string
		action int
	}{
		{"Prev", lightshow.ActionPrevious},
		{"Next", lightshow.ActionNext},
		{"Speed+", lightshow.ActionSpeedUp},
		{"Speed-", lightshow.ActionSpeedDown},
	}
	for i, item := range actions {
		if button(cells[i], item.text, false) {
			player.Action(item.action)
			preview.Restart()
		}
	}

	layout.label("Keys of this show", style.TextColorNormal)
	cells = append(layout.row(4), layout.row(4)[:3]...)
	for key := range 7 {
		if button(cells[key], "KEY"+strconv.Itoa(key+1), false) {
			player.Key(key)
			preview.Restart()
		}
	}

	layout.label("Audio level", style.TextColorNormal)
	bounds := layout.row(1)[0]
	bounds.Width -= 30
	preview.Level = int(sliderBar(bounds, "", strconv.Itoa(preview.Level), float32(preview.Level), 0, 63))
}

func (this *ShowPanel) compile() {
	this.compiled, this.compileError = nil, ""
	data, err := global.Project.Compile()
	if err != nil {
		this.compileError = "Compile error: " + err.Error()
		return
	}
	this.compiled = data
}

func (this *ShowPanel) startPreview() {
	preview, err := show.NewPreview(global.Project)
	if err != nil {
		this.compileError = "Compile error: " + err.Error()
		return
	}
	global.Preview = preview
}

func (this *ShowPanel) save() {
	data, err := global.Project.Marshal()
	if err != nil {
		this.setStatus("Saving failed: "+err.Error(), true)
		return
	}
	global.SendMessage(constants.ServiceShow, models.SaveMessage{Name: global.ProjectName, Data: data})
}

func (this *ShowPanel) upload() {
	this.compile()
	if this.compiled == nil {
		return
	}
	this.uploading, this.uploadAcked, this.uploadTotal, this.uploadResult = true, 0, len(this.compiled), ""
	global.SendMessage(constants.TCPClient, models.TCPUploadMessage{Data: this.compiled})
}

func (this *ShowPanel) setStatus(text string, isError bool) {
	this.statusText, this.statusIsError = text, isError
}

func (this *ShowPanel) handle(message any) {
	switch message := message.(type) {
	case models.ProjectListMessage:
		this.projects = message.Names
		if int(this.projectIndex) >= len(this.projects) {
			this.projectIndex = 0
		}
	case models.ProjectLoadedMessage:
		global.SetProject(message.Name, message.Project)
		this.compile()
	case models.StatusTextMessage:
		this.setStatus(message.Text, message.Error)
	case models.UploadProgressMessage:
		this.uploadAcked, this.uploadTotal = message.Acked, message.Total
	case models.UploadResultMessage:
		this.uploading, this.uploadOK, this.uploadResult = false, message.OK, message.Text
		if message.OK {
			this.uploadAcked = this.uploadTotal
		}
	}
}

func statusColor(isError bool) color.RGBA {
	if isError {
		return raylib.Red
	}
	return raylib.Lime
}
