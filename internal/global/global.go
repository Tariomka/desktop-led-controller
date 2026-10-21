package global

import (
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/show"
)

type LayerState int32

const (
	All LayerState = iota
	Layer
	Column
	Precise
)

type ConnectionState int

const (
	Disconnect ConnectionState = iota
	Connecting
	Connected
	Disconnecting
)

// Globally accesable state
var (
	messenger *data.Messenger = data.NewMessanger()

	ConnectionStatus ConnectionState

	WindowShouldClose bool
)

// Editor state. It is used by the UI goroutine only; services hand over new projects through
// the message bus.
var (
	Project     = show.NewProject()
	ProjectName = "untitled"

	SelectedGraph   int
	SelectedPalette int
	SelectedColor   = 1 // palette index painted into the selected graph
	ShouldPaint     bool

	SelectedLayerState LayerState
	SelectedLayer      uint8 // y, vertical
	SelectedColumn     uint8 // z

	// Preview is the running show preview; nil shows the selected graph in the 3-D view.
	Preview *show.Preview
)

// SetProject replaces the edited project and resets the selections.
func SetProject(name string, project *show.Project) {
	Project, ProjectName = project, name
	SelectedGraph, SelectedPalette, SelectedColor = 0, 0, 1
	Preview = nil
}

func RegisterMessageReceiver(key string, receiver data.MessageReceiver) {
	messenger.RegisterReceiver(key, receiver)
}

func SendMessage(key string, message any) {
	messenger.Send(key, message)
}
