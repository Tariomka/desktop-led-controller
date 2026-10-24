package models

import "github.com/Tariomka/desktop-led-controller/internal/show"

type ConnectedMessage struct{}

type DisconnectedMessage struct{}

// LinkHealthMessage reports whether the Pico answers pings.
type LinkHealthMessage struct{ Responding bool }

type ConsoleSource int

const (
	SourceApp ConsoleSource = iota
	SourceCube
	SourcePico
)

// ConsoleLineMessage is a line for the console: from the cube, the Pico or this app.
type ConsoleLineMessage struct {
	Source ConsoleSource
	Text   string
}

type UploadProgressMessage struct{ Acked, Total int }

type UploadResultMessage struct {
	OK   bool
	Text string
}

// ProjectListMessage lists the saved projects (file names without extension).
type ProjectListMessage struct{ Names []string }

// ProjectLoadedMessage hands a loaded or imported project to the UI, which owns it from then on.
type ProjectLoadedMessage struct {
	Name    string
	Path    string // source file, empty for saved projects
	Project *show.Project
}

// StatusTextMessage reports the result of a file operation.
type StatusTextMessage struct {
	Text  string
	Error bool
}
