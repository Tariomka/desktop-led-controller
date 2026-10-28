package ui

import (
	"path/filepath"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/config"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/gen2brain/raylib-go/raygui"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

// Window is the application window: layout, panels, dialogs, shortcuts and the message bus
// receiver for everything that reaches the UI.
type Window struct {
	layout  Layout
	actions component.Actions

	grid    *CubeGrid
	top     component.TopBar
	status  component.StatusBar
	assets  *component.AssetsPanel
	right   *component.RightPanel
	bottom  *component.BottomPanel
	dialogs component.Dialogs

	channel                     chan any
	settingsDirty               bool
	lastRightTab, lastBottomTab int
}

func NewWindow() *Window {
	window := &Window{channel: make(chan any, 256)}
	global.RegisterMessageReceiver(constants.UIApp, func(message any) { window.channel <- message })
	return window
}

func (this *Window) Start() {
	settings := &global.Settings
	raylib.SetConfigFlags(raylib.FlagWindowResizable | raylib.FlagWindowHighdpi)
	raylib.InitWindow(int32(settings.Window.Width), int32(settings.Window.Height), "LED Cube Studio")
	raylib.SetWindowMinSize(config.MinWidth, config.MinHeight)
	raylib.SetExitKey(0)
	raylib.SetTargetFPS(60)
	style.LoadStyle()
	raygui.EnableTooltip()

	this.actions = component.Actions{
		New:      this.newProject,
		Open:     func() { this.confirmDiscard("Open another project?", this.dialogs.OpenProject) },
		Save:     this.save,
		SaveAs:   this.dialogs.SaveAs,
		Import:   func() { this.confirmDiscard("Import another file?", this.dialogs.Import) },
		Settings: this.dialogs.Settings,
		ShowTab:  func(tab int) { this.right.Tab = tab },
		Pick:     this.dialogs.Pick,
		Status:   this.setStatus,
		Error:    func(text string) { this.dialogs.ShowError("Error", text) },
	}
	this.dialogs.Actions = &this.actions
	this.assets = &component.AssetsPanel{Actions: &this.actions}
	this.right = component.NewRightPanel(&this.actions, this.assets)
	this.right.Tab = settings.Panels.RightTab
	this.right.Tools.TextTool = RenderText
	this.bottom = &component.BottomPanel{
		Tab:      settings.Panels.BottomTab,
		Timeline: &component.Timeline{Actions: &this.actions, Show: this.right.Show},
		Console:  component.NewConsole(),
	}
	this.grid = NewCubeGrid()
	this.setStatus("Ready - drop .ls, .gc, .hex or .bin files on the window to open them", false)
}

func (this *Window) Stop() {
	this.saveSettings()
	this.grid.Unload()
	raylib.CloseWindow()
}

// Render is the blocking frame loop.
func (this *Window) Render() {
	for !global.WindowShouldClose {
		if raylib.WindowShouldClose() {
			this.confirmDiscard("Exit without saving?", func() { global.WindowShouldClose = true })
		}
		this.Frame()
	}
}

// Frame runs one update/draw cycle.
func (this *Window) Frame() {
	this.receive()
	this.importDroppedFiles()
	this.frame()
}

func (this *Window) frame() {
	width, height := float32(raylib.GetScreenWidth()), float32(raylib.GetScreenHeight())
	panels := &global.Settings.Panels
	mouseFree := !this.dialogs.Active()
	if this.layout.Update(panels, width, height, mouseFree) {
		this.settingsDirty = true
	}
	if raylib.IsWindowResized() {
		this.settingsDirty = true
	}
	this.shortcuts()
	this.grid.Update(this.layout.Center, mouseFree && !this.layout.Dragging())

	raylib.BeginDrawing()
	raylib.ClearBackground(style.BackgroundColor)
	raylib.SetMouseCursor(int32(raylib.MouseCursorDefault))
	if this.dialogs.Active() {
		raygui.Lock()
	}
	this.grid.Render()
	if panels.LeftOpen {
		this.assets.Draw(this.layout.Left)
	}
	if panels.RightOpen {
		this.right.Draw(this.layout.Right)
	}
	if panels.BottomOpen {
		this.bottom.Draw(this.layout.Bottom)
	}
	this.layout.DrawChrome(panels)
	this.top.Draw(this.layout.TopBar, &this.actions, panels)
	this.status.Draw(this.layout.StatusBar, this.right.Cube)
	raygui.Unlock()
	this.dialogs.Draw(raylib.NewRectangle(0, 0, width, height))
	raylib.EndDrawing()

	if this.right.Tab != panels.RightTab || this.bottom.Tab != panels.BottomTab {
		// Tabs changed by a shortcut, a panel or the settings: keep both in step.
		if this.right.Tab != this.lastRightTab {
			panels.RightTab = this.right.Tab
		} else {
			this.right.Tab = panels.RightTab
		}
		if this.bottom.Tab != this.lastBottomTab {
			panels.BottomTab = this.bottom.Tab
		} else {
			this.bottom.Tab = panels.BottomTab
		}
		this.settingsDirty = true
	}
	this.lastRightTab, this.lastBottomTab = this.right.Tab, this.bottom.Tab
	if this.settingsDirty && !this.layout.Dragging() && !raylib.IsMouseButtonDown(raylib.MouseLeftButton) {
		this.saveSettings()
	}
}

func (this *Window) saveSettings() {
	this.settingsDirty = false
	if !raylib.IsWindowMinimized() {
		global.Settings.Window = config.Window{Width: raylib.GetScreenWidth(), Height: raylib.GetScreenHeight()}
	}
	if err := global.Settings.Save(); err != nil {
		this.setStatus("Saving settings failed: "+err.Error(), true)
	}
}

func (this *Window) shortcuts() {
	if this.dialogs.Active() || component.EditingText() {
		return
	}
	control := raylib.IsKeyDown(raylib.KeyLeftControl) || raylib.IsKeyDown(raylib.KeyRightControl)
	shift := raylib.IsKeyDown(raylib.KeyLeftShift) || raylib.IsKeyDown(raylib.KeyRightShift)
	panels := &global.Settings.Panels
	switch {
	case control && raylib.IsKeyPressed(raylib.KeyN):
		this.actions.New()
	case control && raylib.IsKeyPressed(raylib.KeyO):
		this.actions.Open()
	case control && shift && raylib.IsKeyPressed(raylib.KeyS):
		this.actions.SaveAs()
	case control && raylib.IsKeyPressed(raylib.KeyS):
		this.actions.Save()
	case control && raylib.IsKeyPressed(raylib.KeyI):
		this.actions.Import()
	case control && raylib.IsKeyPressed(raylib.KeyZ):
		global.Undo()
	case control && raylib.IsKeyPressed(raylib.KeyY):
		global.Redo()
	case raylib.IsKeyPressed(raylib.KeyF1):
		panels.LeftOpen = !panels.LeftOpen
		this.settingsDirty = true
	case raylib.IsKeyPressed(raylib.KeyF2):
		panels.RightOpen = !panels.RightOpen
		this.settingsDirty = true
	case raylib.IsKeyPressed(raylib.KeyF3):
		panels.BottomOpen = !panels.BottomOpen
		this.settingsDirty = true
	case raylib.IsKeyPressed(raylib.KeyF12):
		panels.BottomOpen = true
		this.bottom.Tab = component.BottomConsole
		this.bottom.Console.Focus()
	case raylib.IsKeyPressed(raylib.KeySpace):
		if global.Preview == nil {
			this.right.Show.PlaySelected()
		} else {
			global.Preview.TogglePause()
		}
	case raylib.IsKeyPressed(raylib.KeyLeft) || raylib.IsKeyPressedRepeat(raylib.KeyLeft):
		if global.Preview != nil {
			global.Preview.Step(-1)
		}
	case raylib.IsKeyPressed(raylib.KeyRight) || raylib.IsKeyPressedRepeat(raylib.KeyRight):
		if global.Preview != nil {
			global.Preview.Step(1)
		}
	case raylib.IsKeyPressed(raylib.KeyPageUp) || raylib.IsKeyPressedRepeat(raylib.KeyPageUp):
		this.moveLayer(1)
	case raylib.IsKeyPressed(raylib.KeyPageDown) || raylib.IsKeyPressedRepeat(raylib.KeyPageDown):
		this.moveLayer(-1)
	}
}

func (this *Window) moveLayer(delta int) {
	graph := global.SelectedGraph()
	if graph == nil {
		return
	}
	if global.SelectedLayerState == global.All || global.SelectedLayerState == global.Column {
		global.SelectedLayerState = global.Layer
	}
	global.SelectedLayer = max(0, min(global.SelectedLayer+delta, graph.H-1))
}

// confirmDiscard runs action directly, or after a confirmation when the project has changes.
func (this *Window) confirmDiscard(question string, action func()) {
	if !global.Dirty {
		action()
		return
	}
	this.dialogs.Confirm("Unsaved changes", "The project has unsaved changes. "+question, action)
}

func (this *Window) newProject() {
	this.confirmDiscard("Start a new project?", func() {
		global.SetProject("untitled", show.NewProject())
		this.right.Show.Invalidate()
		this.setStatus("New project", false)
	})
}

func (this *Window) save() {
	if global.ProjectName == "" || global.ProjectName == "untitled" {
		this.dialogs.SaveAs()
		return
	}
	data, err := global.Project.Marshal()
	if err != nil {
		this.actions.Error("Saving failed: " + err.Error())
		return
	}
	global.SendMessage(constants.ServiceShow, models.SaveMessage{Name: global.ProjectName, Data: data})
	global.Dirty = false
}

func (this *Window) setStatus(text string, isError bool) {
	this.status.Text, this.status.IsError = text, isError
}

// importDroppedFiles loads files dragged onto the window (.ls, .gc, .hex, .bin).
func (this *Window) importDroppedFiles() {
	if !raylib.IsFileDropped() {
		return
	}
	paths := raylib.LoadDroppedFiles()
	raylib.UnloadDroppedFiles()
	if len(paths) == 0 {
		return
	}
	path := paths[0]
	this.confirmDiscard("Import "+filepath.Base(path)+"?", func() {
		global.SendMessage(constants.ServiceShow, models.ImportMessage{Path: path})
	})
}

// receive handles the messages queued since the last frame.
func (this *Window) receive() {
	for {
		select {
		case message := <-this.channel:
			this.handle(message)
		default:
			return
		}
	}
}

func (this *Window) handle(message any) {
	switch message := message.(type) {
	case models.ConsoleLineMessage:
		this.bottom.Console.Add(message)
		this.right.Cube.Handle(message)
	case models.ConnectedMessage, models.DisconnectedMessage, models.LinkHealthMessage:
		this.right.Cube.Handle(message)
		this.right.Show.Handle(message)
	case models.UploadProgressMessage, models.UploadResultMessage:
		this.right.Show.Handle(message)
	case models.ProjectListMessage:
		this.dialogs.Projects = message.Names
	case models.ProjectLoadedMessage:
		global.SetProject(message.Name, message.Project)
		this.right.Show.Invalidate()
		if message.Path != "" {
			if !strings.EqualFold(filepath.Ext(message.Path), ".ls") {
				global.Dirty = true // imported: not saved as a project yet
			}
			global.Settings.AddRecent(message.Path)
			this.settingsDirty = true
		}
	case models.StatusTextMessage:
		this.setStatus(message.Text, message.Error)
	}
}
