package global

import (
	"github.com/Tariomka/desktop-led-controller/internal/config"
	"github.com/Tariomka/desktop-led-controller/internal/data"
	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

type ConnectionState int

const (
	Disconnect ConnectionState = iota
	Connecting
	Connected
	Disconnecting
)

type SelectionKind int

const (
	SelectNone SelectionKind = iota
	SelectGraph
	SelectPalette
	SelectPath
	SelectCartoon
	SelectProp // Element/Prop of the selected cartoon
	SelectKeys
	SelectDisplay // Index = display, -1 = the table itself
)

// Selection is what the Props tab edits and the timeline shows.
type Selection struct {
	Kind    SelectionKind
	Index   int // graph / palette / path / display
	Group   int // cartoon group
	Cartoon int // cartoon in the group
	Element int
	Prop    int
}

type LayerState int32

const (
	All LayerState = iota
	Layer
	Column
	Precise
)

// Globally accessible state
var (
	messenger *data.Messenger = data.NewMessanger()

	ConnectionStatus ConnectionState
	PicoResponding   = true

	WindowShouldClose bool
)

// Editor state. It is used by the UI goroutine only; services hand over new projects through
// the message bus.
var (
	Settings config.Settings

	Project     = show.NewProject()
	ProjectName = "untitled"
	Dirty       bool

	Selected = Selection{Kind: SelectGraph}

	SelectedPalette int
	SelectedColor   = 1 // palette index painted into the selected graph
	ShouldPaint     bool

	SelectedLayerState LayerState
	SelectedLayer      int // y, vertical
	SelectedColumn     int // z

	// Preview is the running show preview; nil shows the selected graph in the 3-D view.
	Preview *show.Preview

	undo, redo []graphSnapshot
)

const undoLimit = 64

type graphSnapshot struct {
	index int
	graph blob.GraphData
}

func RegisterMessageReceiver(key string, receiver data.MessageReceiver) {
	messenger.RegisterReceiver(key, receiver)
}

func SendMessage(key string, message any) {
	messenger.Send(key, message)
}

// SetProject replaces the edited project and resets the selections.
func SetProject(name string, project *show.Project) {
	Project, ProjectName, Dirty = project, name, false
	Selected = Selection{Kind: SelectGraph}
	SelectedPalette, SelectedColor = 0, 1
	Preview = nil
	undo, redo = nil, nil
}

func MarkDirty() { Dirty = true }

// SelectedGraph returns the graph of the selection, or nil.
func SelectedGraph() *blob.GraphData {
	if Selected.Kind != SelectGraph || Selected.Index < 0 || Selected.Index >= len(Project.Graphs) {
		return nil
	}
	return &Project.Graphs[Selected.Index]
}

// SelectedCartoon returns the cartoon of a cartoon/prop selection, or nil.
func SelectedCartoon() *show.Cartoon {
	if Selected.Kind != SelectCartoon && Selected.Kind != SelectProp {
		return nil
	}
	if Selected.Group < 0 || Selected.Group >= blob.GroupCount {
		return nil
	}
	cartoons := Project.Groups[Selected.Group]
	if Selected.Cartoon < 0 || Selected.Cartoon >= len(cartoons) {
		return nil
	}
	return &cartoons[Selected.Cartoon]
}

// SelectedElement returns the element of a prop selection, or nil.
func SelectedElement() *blob.ElementData {
	cartoon := SelectedCartoon()
	if cartoon == nil || Selected.Kind != SelectProp || Selected.Element < 0 || Selected.Element >= len(cartoon.Elements) {
		return nil
	}
	return &cartoon.Elements[Selected.Element]
}

// PushUndo stores the selected graph before a change.
func PushUndo() {
	graph := SelectedGraph()
	if graph == nil {
		return
	}
	undo = append(undo, graphSnapshot{Selected.Index, show.CopyGraph(*graph, graph.Name)})
	if len(undo) > undoLimit {
		undo = undo[1:]
	}
	redo = nil
}

func CanUndo() bool { return len(undo) > 0 }
func CanRedo() bool { return len(redo) > 0 }

func Undo() { swapSnapshot(&undo, &redo) }
func Redo() { swapSnapshot(&redo, &undo) }

func swapSnapshot(from, to *[]graphSnapshot) {
	if len(*from) == 0 {
		return
	}
	snapshot := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]
	if snapshot.index < 0 || snapshot.index >= len(Project.Graphs) {
		return
	}
	current := Project.Graphs[snapshot.index]
	*to = append(*to, graphSnapshot{snapshot.index, show.CopyGraph(current, current.Name)})
	Project.Graphs[snapshot.index] = snapshot.graph
	Selected = Selection{Kind: SelectGraph, Index: snapshot.index}
	Dirty = true
}
