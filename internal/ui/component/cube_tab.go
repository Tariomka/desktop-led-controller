package component

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/led-common-lib/pkg/communication/uart"
)

// Factory functions of KEY1..KEY7 (uploaded shows may remap them).
var keyNames = [7]string{"Prev", "Next", "Spd+", "Spd-", "Pause", "Model", "Disp"}

// CubeTab connects to the Pico and controls the cube: status, keys, cartoon and registry.
type CubeTab struct {
	retry bool

	status         uart.Status
	statusReceived time.Time
	restartedAt    time.Time
	cartoon        int

	regKey, regValue string
	registry         []string // lines of the last RegList
	listing          bool
}

func (this *CubeTab) Draw(layout *rowLayout) {
	this.drawConnection(layout)
	layout.separator()
	this.drawStatus(layout)
	layout.separator()
	this.drawRemote(layout)
	layout.separator()
	this.drawRegistry(layout)
}

func (this *CubeTab) drawConnection(layout *rowLayout) {
	layout.header("Connection")
	layout.label(fmt.Sprintf("Pico at %s:%d (change in Settings)", global.Settings.IP, global.Settings.Port), MutedColor)
	cells := layout.row(2)
	switch global.ConnectionStatus {
	case global.Disconnect:
		if accentButton(cells[0], "Connect", false) {
			global.ConnectionStatus = global.Connecting
			global.SendMessage(constants.TCPClient, models.TCPConnectMessage{})
		}
	case global.Connecting:
		button(cells[0], "Connecting...", true)
	case global.Connected:
		if button(cells[0], "Disconnect", false) {
			global.ConnectionStatus = global.Disconnecting
			global.SendMessage(constants.TCPClient, models.TCPDisconnectMessage{})
		}
	case global.Disconnecting:
		button(cells[0], "Disconnecting...", true)
	}
	this.retry = checkBox(cells[1], "Keep retrying", this.retry)
	if global.ConnectionStatus == global.Connected && !global.PicoResponding {
		layout.label("The Pico does not answer pings.", WarningColor)
	}
}

func (this *CubeTab) drawStatus(layout *rowLayout) {
	layout.header("Cube status")
	if this.statusReceived.IsZero() {
		layout.label("No status received yet.", MutedColor)
		return
	}
	status := this.status
	age := time.Since(this.statusReceived).Round(time.Second)
	textColor := MutedColor
	if age > 5*time.Second {
		textColor = WarningColor
	}
	layout.label(fmt.Sprintf("Received %s ago", age), textColor)
	layout.label("Mode: "+status.Mode, style.TextColorNormal)
	if status.Mode != "Text" {
		layout.label("Uploaded shows play only in Text mode (DIP SW1 and SW2 on).", WarningColor)
	}
	playModel := "repeat one"
	if status.PlayAll {
		playModel = "play all"
	}
	layout.label(fmt.Sprintf("Cartoon %d / %d - speed %d.%02d - %s", status.Cartoon+1, status.Count, status.Speed/100, status.Speed%100, playModel), style.TextColorNormal)
	layout.label(fmt.Sprintf("Paused: %t - display off: %t - device %d", status.Pause, status.DisplayOff, status.DeviceNum), style.TextColorNormal)
	if !this.restartedAt.IsZero() && time.Since(this.restartedAt) < 30*time.Second {
		layout.label("The cube restarted "+time.Since(this.restartedAt).Round(time.Second).String()+" ago.", WarningColor)
	}
}

func (this *CubeTab) drawRemote(layout *rowLayout) {
	layout.header("Remote")
	disabled := global.ConnectionStatus != global.Connected
	layout.label("Keys 1-7 (factory functions)", MutedColor)
	cells := layout.row(7)
	for key, name := range keyNames {
		if button(cells[key], name, disabled) {
			this.command("Key " + strconv.Itoa(key+1))
		}
	}
	cells = layout.row(2)
	if button(cells[0], "Replay", disabled) {
		this.command("Action 1")
	}
	if button(cells[1], "Speed reset", disabled) {
		this.command("Action 5")
	}
	cells = layout.split(1, 1, 1)
	layoutLabel(cells[0], "Cartoon")
	if this.cartoon < 1 {
		this.cartoon = 1
	}
	spinner("cube-cartoon", cells[1], &this.cartoon, 1, max(this.status.Count, 1))
	if button(cells[2], "Play", disabled) {
		this.command("Cartoon " + strconv.Itoa(this.cartoon-1))
	}
	cells = layout.row(2)
	if button(cells[0], "Status now", disabled) {
		this.command("Status")
	}
	if button(cells[1], "Help", disabled) {
		this.command("Help")
	}
}

func (this *CubeTab) drawRegistry(layout *rowLayout) {
	layout.header("Registry (cube settings)")
	disabled := global.ConnectionStatus != global.Connected
	cells := layout.split(1, 1)
	changedKey := textBox("reg-key", cells[0], &this.regKey, 32)
	_ = changedKey
	textBox("reg-value", cells[1], &this.regValue, 32)
	cells = layout.row(4)
	if button(cells[0], "List", disabled) {
		this.registry, this.listing = nil, true
		this.command("RegList")
	}
	if button(cells[1], "Get", disabled || this.regKey == "") {
		this.command("GetReg " + this.regKey)
	}
	if button(cells[2], "Set", disabled || this.regKey == "") {
		this.command("SetReg " + this.regKey + " " + this.regValue)
	}
	if button(cells[3], "Delete", disabled || this.regKey == "") {
		this.command("DelReg " + this.regKey)
	}
	if button(layout.row(1)[0], "Clear all registry entries", disabled) {
		this.command("ClrReg")
	}
	if len(this.registry) > 0 {
		layout.label(fmt.Sprintf("%d entries (click to pick):", len(this.registry)), MutedColor)
		for _, entry := range this.registry {
			bounds := layout.row(1)[0]
			if button(bounds, entry, false) {
				this.regKey, this.regValue, _ = strings.Cut(entry, "=")
			}
		}
	}
}

func (this *CubeTab) command(line string) {
	global.SendMessage(constants.TCPClient, models.TCPCommandMessage{Line: line})
}

// Summary is the status bar text.
func (this *CubeTab) Summary() string {
	if global.ConnectionStatus != global.Connected {
		return "Cube: not connected"
	}
	if this.statusReceived.IsZero() {
		return "Cube: waiting for status"
	}
	status := this.status
	return fmt.Sprintf("Cube: %s - cartoon %d/%d - speed %d.%02d - %s ago",
		status.Mode, status.Cartoon+1, status.Count, status.Speed/100, status.Speed%100, time.Since(this.statusReceived).Round(time.Second))
}

// Handle takes connection events and cube lines.
func (this *CubeTab) Handle(message any) {
	switch message := message.(type) {
	case models.ConnectedMessage:
		global.ConnectionStatus = global.Connected
		global.PicoResponding = true
	case models.DisconnectedMessage:
		this.statusReceived = time.Time{}
		if !this.retry || global.ConnectionStatus == global.Disconnecting {
			global.ConnectionStatus = global.Disconnect
			return
		}
		global.ConnectionStatus = global.Connecting
		global.SendMessage(constants.TCPClient, models.TCPConnectMessage{})
	case models.LinkHealthMessage:
		global.PicoResponding = message.Responding
	case models.ConsoleLineMessage:
		if message.Source == models.SourceCube {
			this.cubeLine(message.Text)
		}
	}
}

func (this *CubeTab) cubeLine(line string) {
	if this.listing {
		switch {
		case strings.HasPrefix(line, "DoGetRegeditList"), strings.HasPrefix(line, "Regedit Count"):
		case strings.Contains(line, "=") && !strings.HasPrefix(line, "Status:"):
			this.registry = append(this.registry, line)
			return
		default:
			this.listing = false
		}
	}

	event := uart.ParseLine(line)
	switch event.Kind {
	case uart.EventStatus:
		this.status, this.statusReceived = event.Status, time.Now()
	case uart.EventMode:
		this.status.Mode = event.Text
	case uart.EventCartoon:
		this.status.Cartoon = event.Values[0]
	case uart.EventSpeed:
		this.status.Speed = event.Values[0]
	case uart.EventPlayModel:
		this.status.PlayAll = event.Values[0] == 1
	case uart.EventPause:
		this.status.Pause = event.Values[0] == 1
	case uart.EventDisplayOff:
		this.status.DisplayOff = event.Values[0] == 1
	case uart.EventStartup:
		this.restartedAt = time.Now()
	}
}
