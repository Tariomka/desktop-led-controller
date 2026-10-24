package models

// FetchMessage asks for the list of saved projects.
type FetchMessage struct{}

// LoadMessage loads a saved project by name.
type LoadMessage struct{ Name string }

// ImportMessage loads a file by path: .ls, .gc, .hex or .bin.
type ImportMessage struct{ Path string }

// SaveMessage stores an encoded project under a name.
type SaveMessage struct {
	Name string
	Data []byte
}

// RenameProjectMessage renames a saved project.
type RenameProjectMessage struct{ From, To string }

// DeleteProjectMessage deletes a saved project.
type DeleteProjectMessage struct{ Name string }

// ExportMessage writes compiled show data as <name>.bin next to the projects.
type ExportMessage struct {
	Name string
	Data []byte
}
