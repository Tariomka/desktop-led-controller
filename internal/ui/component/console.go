package component

import (
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	commandMaxLength = 180
	historyLength    = 50
)

// Console shows cube, Pico and app lines with filters, and sends console commands.
type Console struct {
	messages   *data.RingBuffer[models.ConsoleLineMessage]
	show       [3]bool // per source
	autoScroll bool
	scroll     float32

	command   string
	history   []string
	historyAt int
}

func NewConsole() *Console {
	return &Console{
		messages:   data.NewRingBuffer[models.ConsoleLineMessage](500),
		show:       [3]bool{true, true, true},
		autoScroll: true,
		historyAt:  -1,
	}
}

// Add appends a line; the periodic Status replies are hidden.
func (this *Console) Add(line models.ConsoleLineMessage) {
	if line.Source == models.SourceCube && strings.HasPrefix(line.Text, "Status:") {
		return
	}
	this.messages.Add(line)
}

func (this *Console) visible() []models.ConsoleLineMessage {
	var lines []models.ConsoleLineMessage
	// The ring buffer keeps the newest first.
	for i := this.messages.Length() - 1; i >= 0; i-- {
		line := this.messages.Get(i)
		if this.show[line.Source] {
			lines = append(lines, line)
		}
	}
	return lines
}

func lineColor(source models.ConsoleSource) raylib.Color {
	switch source {
	case models.SourcePico:
		return raylib.SkyBlue
	case models.SourceApp:
		return raylib.Gold
	}
	return style.ListTextColorNormal
}

func (this *Console) Draw(bounds raylib.Rectangle) {
	layout := newRowLayout(bounds, Padding)
	cells := layout.split(1, 1, 1, 1, 2)
	names := [...]string{"App", "Cube", "Pico"}
	for source := range names {
		if toggleButton(cells[source], names[source], this.show[source]) {
			this.show[source] = !this.show[source]
		}
	}
	if toggleButton(cells[3], "Auto-scroll", this.autoScroll) {
		this.autoScroll = !this.autoScroll
	}
	if button(cells[4], "Clear", false) {
		this.messages = data.NewRingBuffer[models.ConsoleLineMessage](500)
	}

	listBounds := layout.rowHeight(1, layout.remaining()-ControlHeight()-layout.padding)[0]
	this.drawLines(listBounds)

	cells = layout.split(5, 1)
	this.drawInput(cells[0], cells[1])
}

func (this *Console) drawLines(bounds raylib.Rectangle) {
	raylib.DrawRectangleRec(bounds, raylib.Fade(raylib.Black, 0.25))
	raylib.DrawRectangleLinesEx(bounds, 1, style.BorderColor)
	lines := this.visible()
	rowHeight := style.TextSize + 4
	content := float32(len(lines)) * rowHeight
	maxScroll := max(0, content-bounds.Height+4)
	if mouseIn(bounds) {
		if wheel := raylib.GetMouseWheelMove(); wheel != 0 {
			this.scroll -= wheel * rowHeight * 3
			this.autoScroll = false
		}
	}
	if this.autoScroll {
		this.scroll = maxScroll
	}
	this.scroll = clampf(this.scroll, 0, maxScroll)

	raylib.BeginScissorMode(int32(bounds.X), int32(bounds.Y), int32(bounds.Width), int32(bounds.Height))
	y := bounds.Y + 2 - this.scroll
	for _, line := range lines {
		if y+rowHeight >= bounds.Y && y <= bounds.Y+bounds.Height {
			utils.RenderText(line.Text, raylib.NewRectangle(bounds.X+6, y, bounds.Width-12, rowHeight), lineColor(line.Source))
		}
		y += rowHeight
	}
	raylib.EndScissorMode()
}

func (this *Console) drawInput(textBounds, buttonBounds raylib.Rectangle) {
	disabled := global.ConnectionStatus != global.Connected
	if editing == "console-command" {
		if raylib.IsKeyPressed(raylib.KeyUp) && len(this.history) > 0 {
			this.historyAt = min(this.historyAt+1, len(this.history)-1)
			this.command = this.history[len(this.history)-1-this.historyAt]
		}
		if raylib.IsKeyPressed(raylib.KeyDown) {
			this.historyAt = max(this.historyAt-1, -1)
			this.command = ""
			if this.historyAt >= 0 {
				this.command = this.history[len(this.history)-1-this.historyAt]
			}
		}
	}
	wasEditing := editing == "console-command"
	textBox("console-command", textBounds, &this.command, commandMaxLength)
	if wasEditing && editing != "console-command" && raylib.IsKeyPressed(raylib.KeyEnter) {
		this.send()
		editing = "console-command"
	}
	if button(buttonBounds, "Send", disabled) {
		this.send()
	}
	if this.command == "" && editing != "console-command" {
		utils.RenderText("  Cube command: Status, Key 2, Cartoon 5, Help ...", textBounds, MutedColor)
	}
}

func (this *Console) send() {
	line := strings.TrimSpace(this.command)
	if line == "" || global.ConnectionStatus != global.Connected {
		return
	}
	global.SendMessage(constants.TCPClient, models.TCPCommandMessage{Line: line})
	this.history = append(this.history, line)
	if len(this.history) > historyLength {
		this.history = this.history[1:]
	}
	this.command, this.historyAt = "", -1
}

// Focus puts the keyboard into the command box.
func (this *Console) Focus() { editing = "console-command" }
