package models

// FetchMessage asks for the list of saved projects.
type FetchMessage struct{}

// LoadMessage loads a saved project by name.
type LoadMessage struct{ Name string }

// ImportMessage loads a file dropped on the window: .ls, .gc, .hex or .bin.
type ImportMessage struct{ Path string }

// SaveMessage stores an encoded project under a name.
type SaveMessage struct {
	Name string
	Data []byte
}
