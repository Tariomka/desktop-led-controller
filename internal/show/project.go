// Package show holds the editable light show project and its file formats: the JSON project
// file (.ls), cube show data (.bin), Intel HEX firmware images and vendor 3D Software projects (.gc).
package show

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
)

// ProjectVersion is the version of the .ls format this package writes.
const ProjectVersion = 1

var (
	ErrProjectVersion    = errors.New("unsupported project version")
	ErrScriptNotCompiled = errors.New("script cartoon has no compiled code")
)

// Project is the content of a .ls file: the show data model plus what only the editor needs
// (cartoon names, enabled flags, script sources), like a vendor .gc project.
type Project struct {
	Version  int
	Config   blob.Config
	Palettes []blob.Palette
	Graphs   []blob.GraphData
	Indexes  []blob.IndexData
	Groups   [blob.GroupCount][]Cartoon
}

type Cartoon struct {
	Name    string
	Enabled bool
	blob.CartoonData
	// ScriptSource is the vendor script text of an imported .gc cartoon; only Script words can be compiled.
	ScriptSource string `json:",omitempty"`
	ScriptRtf    bool   `json:",omitempty"`
}

// Default key functions of the factory shows: previous, next, speed up, speed down, pause,
// play model, display off.
var DefaultKeys = [7]byte{2, 3, 4, 6, 8, 7, 9}

// NewProject returns a small project for one 8x8x8 cube with one empty cartoon.
func NewProject() *Project {
	palette := blob.Palette{Name: "Palette"}
	copy(palette.Colors[:], render.DefaultPalette)

	return &Project{
		Version:  ProjectVersion,
		Config:   blob.Config{Keys: DefaultKeys, Displays: []blob.Display{{CubeType: 1}}},
		Palettes: []blob.Palette{palette},
		Graphs:   []blob.GraphData{NewGraph("Graph", 8, 8, 8)},
		Groups:   [blob.GroupCount][]Cartoon{{NewCartoon("Cartoon", blob.Normal)}},
	}
}

func NewGraph(name string, l, h, w int) blob.GraphData {
	return blob.GraphData{Name: name, L: l, H: h, W: w, Voxels: make([]byte, l*h*w)}
}

// ResizeGraph changes the graph size, keeping the voxels that still fit.
func ResizeGraph(graph *blob.GraphData, l, h, w int) {
	resized := NewGraph(graph.Name, l, h, w)
	for z := range min(w, graph.W) {
		for y := range min(h, graph.H) {
			for x := range min(l, graph.L) {
				resized.Voxels[x+y*l+z*l*h] = graph.Voxels[x+y*graph.L+z*graph.L*graph.H]
			}
		}
	}
	*graph = resized
}

// FindGraph returns the index of the named graph, or -1.
func FindGraph(project *Project, name string) int {
	for i, graph := range project.Graphs {
		if graph.Name == name {
			return i
		}
	}
	return -1
}

// FindPalette returns the index of the named palette, or -1.
func FindPalette(project *Project, name string) int {
	for i, palette := range project.Palettes {
		if palette.Name == name {
			return i
		}
	}
	return -1
}

// NewCartoon returns an enabled cartoon with the vendor defaults (10 frames, 100 ms, 1 loop).
func NewCartoon(name string, kind int) Cartoon {
	return Cartoon{
		Name:        name,
		Enabled:     true,
		CartoonData: blob.CartoonData{Type: kind, Header: blob.FrameHeader{FrameCount: 10, Loop: 1, Delay: 100}},
	}
}

// LoadProject reads a .ls file.
func LoadProject(data []byte) (*Project, error) {
	project := &Project{}
	if err := json.Unmarshal(data, project); err != nil {
		return nil, err
	}
	if project.Version < 1 || project.Version > ProjectVersion {
		return nil, ErrProjectVersion
	}
	return project, nil
}

// Marshal encodes the project as a .ls file.
func (this *Project) Marshal() ([]byte, error) {
	this.Version = ProjectVersion
	return json.MarshalIndent(this, "", "\t")
}

// Show returns the show data model of the enabled cartoons. Elements without props are left
// out, as the vendor compiler does.
func (this *Project) Show() (*blob.Show, error) {
	result := &blob.Show{
		Config:   this.Config,
		Palettes: this.Palettes,
		Graphs:   this.Graphs,
		Indexes:  this.Indexes,
	}

	for group, cartoons := range this.Groups {
		for _, cartoon := range cartoons {
			if !cartoon.Enabled {
				continue
			}
			if cartoon.Type == blob.Script && len(cartoon.Script) == 0 {
				return nil, fmt.Errorf("group %d, %q: %w", group+1, cartoon.Name, ErrScriptNotCompiled)
			}

			data := cartoon.CartoonData
			data.Elements = nil
			for _, element := range cartoon.Elements {
				if PropCount(element) > 0 {
					data.Elements = append(data.Elements, element)
				}
			}
			result.Groups[group] = append(result.Groups[group], data)
		}
	}
	return result, nil
}

// PropCount is the number of props of the element's kind.
func PropCount(element blob.ElementData) int {
	switch element.Kind {
	case blob.SetGraph:
		return len(element.Graphs)
	case blob.DrawBall:
		return len(element.Balls)
	case blob.TitleStyle:
		return len(element.Titles)
	}
	return 0
}

// Compile returns the show data to upload to the cube.
func (this *Project) Compile() ([]byte, error) {
	model, err := this.Show()
	if err != nil {
		return nil, err
	}
	return model.Marshal()
}

// FromShowData imports compiled show data; cartoons get generated names.
func FromShowData(data []byte) (*Project, error) {
	parsed, err := blob.Parse(string(data))
	if err != nil {
		return nil, err
	}
	model, err := blob.FromBlob(parsed)
	if err != nil {
		return nil, err
	}

	project := &Project{
		Version:  ProjectVersion,
		Config:   model.Config,
		Palettes: model.Palettes,
		Graphs:   model.Graphs,
		Indexes:  model.Indexes,
	}
	for group, cartoons := range model.Groups {
		for i, cartoon := range cartoons {
			project.Groups[group] = append(project.Groups[group], Cartoon{
				Name:        fmt.Sprintf("Cartoon %d", i+1),
				Enabled:     true,
				CartoonData: cartoon,
			})
		}
	}
	return project, nil
}
