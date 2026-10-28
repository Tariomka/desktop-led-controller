package component

import (
	"strconv"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

type DialogKind int

const (
	DialogNone DialogKind = iota
	DialogOpen
	DialogSaveAs
	DialogImport
	DialogSettings
	DialogConfirm
	DialogError
	DialogPick
)

// Dialogs draws one modal dialog at a time over the window.
type Dialogs struct {
	Kind DialogKind

	Projects []string
	selected int
	scroll   int32
	name     string
	path     string
	ip       string
	port     int

	title, message string
	onYes          func()
	names          []string
	onPick         func(string)
	Actions        *Actions
}

func (this *Dialogs) Active() bool { return this.Kind != DialogNone }

func (this *Dialogs) open(kind DialogKind) {
	this.Kind = kind
	StopEditing()
}

func (this *Dialogs) Close() { this.Kind = DialogNone }

func (this *Dialogs) OpenProject() {
	this.name = ""
	this.open(DialogOpen)
	global.SendMessage(constants.ServiceShow, models.FetchMessage{})
}

func (this *Dialogs) SaveAs() {
	this.name = global.ProjectName
	this.open(DialogSaveAs)
}

func (this *Dialogs) Import() {
	this.path = ""
	this.selected = -1
	this.open(DialogImport)
}

func (this *Dialogs) Settings() {
	this.ip, this.port = global.Settings.IP, int(global.Settings.Port)
	this.open(DialogSettings)
}

// Confirm asks a yes/no question and runs onYes for yes.
func (this *Dialogs) Confirm(title, message string, onYes func()) {
	this.title, this.message, this.onYes = title, message, onYes
	this.open(DialogConfirm)
}

func (this *Dialogs) ShowError(title, message string) {
	this.title, this.message = title, message
	this.open(DialogError)
}

// Pick lets the user choose one of names.
func (this *Dialogs) Pick(title string, names []string, onPick func(string)) {
	this.title, this.names, this.onPick, this.selected = title, names, onPick, -1
	this.open(DialogPick)
}

func (this *Dialogs) Draw(screen raylib.Rectangle) {
	if this.Kind == DialogNone {
		return
	}
	raylib.DrawRectangleRec(screen, raylib.Fade(raylib.Black, 0.6))
	if raylib.IsKeyPressed(raylib.KeyEscape) {
		this.Close()
		return
	}

	switch this.Kind {
	case DialogOpen:
		this.drawOpen(center(screen, 460, 360))
	case DialogSaveAs:
		this.drawSaveAs(center(screen, 420, 150))
	case DialogImport:
		this.drawImport(center(screen, 560, 330))
	case DialogSettings:
		this.drawSettings(center(screen, 420, 190))
	case DialogConfirm, DialogError:
		this.drawMessage(center(screen, 460, 150))
	case DialogPick:
		this.drawPick(center(screen, 420, 420))
	}
}

func (this *Dialogs) drawPick(bounds raylib.Rectangle) {
	layout := this.frame(bounds, this.title)
	if layout == nil {
		return
	}
	listBounds := layout.rowHeight(1, layout.remaining()-ControlHeight()-layout.padding)[0]
	picked := list(listBounds, this.names, this.selected, &this.scroll)
	if picked != this.selected || (picked >= 0 && raylib.IsMouseButtonPressed(raylib.MouseLeftButton) && mouseIn(listBounds) && this.selected >= 0 && picked == this.selected) {
		this.selected = picked
	}
	cells := layout.row(2)
	if accentButton(cells[0], "Choose", this.selected < 0 || this.selected >= len(this.names)) || (this.selected >= 0 && raylib.IsKeyPressed(raylib.KeyEnter)) {
		onPick, name := this.onPick, this.names[this.selected]
		this.Close()
		if onPick != nil {
			onPick(name)
		}
	}
	if button(cells[1], "Cancel", false) {
		this.Close()
	}
}

func center(screen raylib.Rectangle, width, height float32) raylib.Rectangle {
	return raylib.NewRectangle(screen.X+(screen.Width-width)/2, screen.Y+(screen.Height-height)/2, width, height)
}

// frame draws the dialog window; it returns the content layout or nil when closed.
func (this *Dialogs) frame(bounds raylib.Rectangle, title string) *rowLayout {
	if raygui.WindowBox(bounds, title) {
		this.Close()
		return nil
	}
	content := raylib.NewRectangle(bounds.X, bounds.Y+24, bounds.Width, bounds.Height-24)
	return newRowLayout(content, Padding*1.5)
}

func (this *Dialogs) drawOpen(bounds raylib.Rectangle) {
	layout := this.frame(bounds, "Open project")
	if layout == nil {
		return
	}
	if len(this.Projects) == 0 {
		layout.label("No saved projects in light_shows/.", MutedColor)
	}
	listBounds := layout.rowHeight(1, layout.remaining()-3*ControlHeight()-3*layout.padding)[0]
	this.selected = list(listBounds, this.Projects, this.selected, &this.scroll)
	valid := this.selected >= 0 && this.selected < len(this.Projects)

	layout.label("Rename to", MutedColor)
	cells := layout.split(3, 1)
	textBox("dialog-rename", cells[0], &this.name, textLength)
	if button(cells[1], "Rename", !valid || this.name == "" || this.name == this.Projects[this.selected]) {
		global.SendMessage(constants.ServiceShow, models.RenameProjectMessage{From: this.Projects[this.selected], To: this.name})
	}

	cells = layout.row(3)
	if accentButton(cells[0], "Load", !valid) {
		global.SendMessage(constants.ServiceShow, models.LoadMessage{Name: this.Projects[this.selected]})
		this.Close()
	}
	if button(cells[1], "Delete", !valid) {
		name := this.Projects[this.selected]
		this.Confirm("Delete project", "Delete "+name+".ls? This cannot be undone.", func() {
			global.SendMessage(constants.ServiceShow, models.DeleteProjectMessage{Name: name})
			this.OpenProject()
		})
	}
	if button(cells[2], "Cancel", false) {
		this.Close()
	}
}

func (this *Dialogs) drawSaveAs(bounds raylib.Rectangle) {
	layout := this.frame(bounds, "Save project as")
	if layout == nil {
		return
	}
	layout.label("Name (saved to light_shows/<name>.ls)", MutedColor)
	textBox("dialog-save-name", layout.row(1)[0], &this.name, textLength)
	cells := layout.row(2)
	if accentButton(cells[0], "Save", this.name == "") || (this.name != "" && raylib.IsKeyPressed(raylib.KeyEnter)) {
		global.ProjectName = this.name
		this.Close()
		if this.Actions != nil && this.Actions.Save != nil {
			this.Actions.Save()
		}
	}
	if button(cells[1], "Cancel", false) {
		this.Close()
	}
}

func (this *Dialogs) drawImport(bounds raylib.Rectangle) {
	layout := this.frame(bounds, "Import a file")
	if layout == nil {
		return
	}
	layout.label("Path of a .ls, .gc (3D Software), .hex (Update.hex) or .bin (show data) file.", MutedColor)
	layout.label("You can also drop files on the window.", MutedColor)
	textBox("dialog-import-path", layout.row(1)[0], &this.path, 512)

	layout.label("Recent", MutedColor)
	listBounds := layout.rowHeight(1, layout.remaining()-ControlHeight()-layout.padding)[0]
	recent := global.Settings.Recent
	if picked := list(listBounds, recent, this.selected, &this.scroll); picked != this.selected && picked >= 0 && picked < len(recent) {
		this.selected = picked
		this.path = recent[picked]
	}

	cells := layout.row(2)
	if accentButton(cells[0], "Import", this.path == "") || (this.path != "" && raylib.IsKeyPressed(raylib.KeyEnter)) {
		global.SendMessage(constants.ServiceShow, models.ImportMessage{Path: this.path})
		this.Close()
	}
	if button(cells[1], "Cancel", false) {
		this.Close()
	}
}

func (this *Dialogs) drawSettings(bounds raylib.Rectangle) {
	layout := this.frame(bounds, "Settings")
	if layout == nil {
		return
	}
	layout.label("Pico W address (the IP the router gave it)", MutedColor)
	cells := layout.split(3, 1)
	textBox("dialog-ip", cells[0], &this.ip, 64)
	valueBox("dialog-port", cells[1], &this.port, 1, 65535)
	layout.label("Panel sizes and the window size are saved automatically.", MutedColor)
	cells = layout.row(2)
	if accentButton(cells[0], "Save", this.ip == "") {
		global.Settings.IP, global.Settings.Port = this.ip, uint16(this.port)
		if err := global.Settings.Save(); err != nil && this.Actions != nil {
			this.Actions.Status("Saving settings failed: "+err.Error(), true)
		} else if this.Actions != nil {
			this.Actions.Status("Settings saved: "+this.ip+":"+strconv.Itoa(this.port), false)
		}
		global.SendMessage(constants.TCPClient, models.TCPConfigureMessage{IP: this.ip, Port: uint16(this.port)})
		this.Close()
	}
	if button(cells[1], "Cancel", false) {
		this.Close()
	}
}

func (this *Dialogs) drawMessage(bounds raylib.Rectangle) {
	layout := this.frame(bounds, this.title)
	if layout == nil {
		return
	}
	textColor := style.TextColorNormal
	if this.Kind == DialogError {
		textColor = ErrorColor
	}
	utils.RenderText(this.message, layout.rowHeight(1, layout.remaining()-ControlHeight()-layout.padding)[0], textColor)
	if this.Kind == DialogError {
		if accentButton(layout.row(1)[0], "OK", false) || raylib.IsKeyPressed(raylib.KeyEnter) {
			this.Close()
		}
		return
	}
	cells := layout.row(2)
	if accentButton(cells[0], "Yes", false) || raylib.IsKeyPressed(raylib.KeyEnter) {
		onYes := this.onYes
		this.Close()
		if onYes != nil {
			onYes()
		}
	}
	if button(cells[1], "No", false) {
		this.Close()
	}
}
