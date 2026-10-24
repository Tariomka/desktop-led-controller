package component

import (
	"github.com/Tariomka/desktop-led-controller/internal/config"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Actions are the application commands the panels can trigger; the window implements them.
type Actions struct {
	New, Open, Save, SaveAs, Import, Settings func()
	// ShowTab switches the right panel to a tab (TabShow..TabCube).
	ShowTab func(tab int)
	// Pick opens a modal list and calls onPick with the chosen name.
	Pick   func(title string, names []string, onPick func(string))
	Status func(text string, isError bool)
	Error  func(text string)
}

// TopBar holds the file actions, the project name and the panel toggles.
type TopBar struct{}

func (this *TopBar) Draw(bounds raylib.Rectangle, actions *Actions, panels *config.Panels) {
	raylib.DrawRectangleRec(bounds, style.BackgroundColor)
	raylib.DrawLine(int32(bounds.X), int32(bounds.Y+bounds.Height-1), int32(bounds.X+bounds.Width), int32(bounds.Y+bounds.Height-1), style.BorderColor)

	height := ControlHeight()
	y := bounds.Y + (bounds.Height-height)/2
	x := bounds.X + Padding
	items := []struct {
		text   string
		action func()
		width  float32
	}{
		{"#8#New", actions.New, 68},
		{"#5#Open", actions.Open, 72},
		{"#6#Save", actions.Save, 70},
		{"#1#Import", actions.Import, 84},
		{"#141#Settings", actions.Settings, 92},
	}
	for _, item := range items {
		if raygui.Button(raylib.NewRectangle(x, y, item.width, height), item.text) && item.action != nil {
			item.action()
		}
		x += item.width + Padding/2
	}

	name := global.ProjectName
	if global.Dirty {
		name += " *"
	}
	utils.RenderText(name, raylib.NewRectangle(bounds.X+bounds.Width/2-200, y, 400, height), raylib.White)

	// Panel toggles and the connection state on the right.
	x = bounds.X + bounds.Width - Padding - 220
	utils.RenderText(connectionText(), raylib.NewRectangle(x, y, 220, height), connectionColor())
	x -= Padding + 3*(height+2)
	toggles := []struct {
		text string
		open *bool
	}{{"L", &panels.LeftOpen}, {"R", &panels.RightOpen}, {"B", &panels.BottomOpen}}
	for _, toggle := range toggles {
		if toggleButton(raylib.NewRectangle(x, y, height, height), toggle.text, *toggle.open) {
			*toggle.open = !*toggle.open
		}
		x += height + 2
	}
}

func connectionText() string {
	switch global.ConnectionStatus {
	case global.Connected:
		if !global.PicoResponding {
			return "Pico not responding"
		}
		return "Connected to the Pico"
	case global.Connecting:
		return "Connecting..."
	case global.Disconnecting:
		return "Disconnecting..."
	}
	return "Not connected"
}

func connectionColor() raylib.Color {
	switch global.ConnectionStatus {
	case global.Connected:
		if !global.PicoResponding {
			return WarningColor
		}
		return OKColor
	case global.Connecting, global.Disconnecting:
		return WarningColor
	}
	return MutedColor
}

// StatusBar shows the last message on the left and the cube state on the right.
type StatusBar struct {
	Text    string
	IsError bool
}

func (this *StatusBar) Draw(bounds raylib.Rectangle, cube *CubeTab) {
	raylib.DrawRectangleRec(bounds, style.BackgroundColor)
	raylib.DrawLine(int32(bounds.X), int32(bounds.Y), int32(bounds.X+bounds.Width), int32(bounds.Y), style.BorderColor)
	textColor := MutedColor
	if this.IsError {
		textColor = ErrorColor
	}
	utils.RenderText(this.Text, raylib.NewRectangle(bounds.X+Padding, bounds.Y, bounds.Width*0.6, bounds.Height), textColor)
	right := raylib.NewRectangle(bounds.X+bounds.Width*0.6, bounds.Y, bounds.Width*0.4-Padding, bounds.Height)
	// LABEL only: DEFAULT would propagate the alignment to every control.
	raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_RIGHT)
	raygui.Label(right, cube.Summary())
	raygui.SetStyle(raygui.LABEL, raygui.TEXT_ALIGNMENT, raygui.TEXT_ALIGN_LEFT)
}
