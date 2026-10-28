package show

import (
	"errors"

	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// GCVersion is the vendor 3D Software project file version this package reads.
const GCVersion = "1.03"

var ErrGCVersion = errors.New("unsupported 3D Software project version (only " + GCVersion + ")")

// FromGC imports a vendor 3D Software project (.gc). The file is a sequence of .NET
// BinaryFormatter streams: version, config, palettes, graphs, indexes and the 4 cartoon groups.
func FromGC(data []byte) (*Project, error) {
	reader := &nrbfReader{data: data}
	next := func() gcStream {
		stream, root := reader.stream()
		return gcStream{stream, root}
	}

	version := next()
	if reader.err != nil {
		return nil, reader.err
	}
	if version.text(version.root, "version") != GCVersion {
		return nil, ErrGCVersion
	}

	project := &Project{Version: ProjectVersion}
	config := next()
	keys := [...]string{"m_key1", "m_key2", "m_key3", "m_key4", "m_key5", "m_key6", "m_key7"}
	for i, key := range keys {
		project.Config.Keys[i] = byte(config.number(config.root, key))
	}
	for _, item := range config.list(config.field(config.root, "m_displayList")) {
		project.Config.Displays = append(project.Config.Displays, blob.Display{
			X:   config.number(item, "X"),
			Y:   config.number(item, "Y"),
			Z:   config.number(item, "Z"),
			Dir: config.number(item, "Dir"),
			CubeType: cubeType(
				config.number(item, "Long"), config.number(item, "Height"), config.number(item, "Width")),
		})
	}

	palettes := next()
	for _, item := range palettes.sourceList() {
		palette := blob.Palette{Name: palettes.text(item, "m_name")}
		copy(palette.Colors[:], palettes.bytes(item, "m_buff"))
		project.Palettes = append(project.Palettes, palette)
	}

	graphs := next()
	for _, item := range graphs.sourceList() {
		graph := blob.GraphData{
			Name:   graphs.text(item, "m_name"),
			L:      graphs.number(item, "m_long"),
			H:      graphs.number(item, "m_height"),
			W:      graphs.number(item, "m_width"),
			Voxels: graphs.bytes(item, "m_buff"),
		}
		if len(graph.Voxels) != graph.L*graph.H*graph.W {
			return nil, ErrNRBF
		}
		project.Graphs = append(project.Graphs, graph)
	}

	indexes := next()
	for _, item := range indexes.sourceList() {
		project.Indexes = append(project.Indexes, blob.IndexData{
			Name: indexes.text(item, "m_name"),
			Data: indexes.bytes(item, "m_buff"),
		})
	}

	for group := range project.Groups {
		cartoons := next()
		for _, item := range cartoons.list(cartoons.field(cartoons.root, "m_list")) {
			project.Groups[group] = append(project.Groups[group], cartoons.cartoon(item))
		}
	}

	if reader.err != nil {
		return nil, reader.err
	}
	return project, nil
}

// cubeType is the display type code of a cube size, as the vendor DebugSolid.CType.
func cubeType(l, h, w int) int {
	switch {
	case l == 8 && h == 8 && w == 8:
		return 1
	case l == 8 && h == 16 && w == 8:
		return 2
	case l == 16 && h == 16 && w == 16:
		return 3
	case l == 8 && h == 32 && w == 8:
		return 4
	}
	return 0
}

type gcStream struct {
	*nrbfStream
	root any
}

func (this gcStream) cartoon(item any) Cartoon {
	frames := this.field(item, "m_frames")
	cartoon := Cartoon{
		Name:         this.text(item, "m_name"),
		Enabled:      this.flag(item, "m_isUse"),
		ScriptSource: this.text(item, "m_scriptText"),
		ScriptRtf:    this.flag(item, "m_isRtf"),
	}
	cartoon.Type = this.number(item, "m_cartoonType")
	cartoon.Header = blob.FrameHeader{
		FrameCount: this.number(frames, "m_frameCount"),
		Loop:       this.number(frames, "m_loop"),
		Delay:      this.number(frames, "m_delay"),
	}
	if cartoon.Type == blob.Audio {
		cartoon.Header.FrameCount = this.number(item, "m_audioTime")
		cartoon.Header.Loop = this.number(item, "m_audioType")
	}
	if cartoon.Type != blob.Script {
		cartoon.ScriptSource, cartoon.ScriptRtf = "", false
	}

	for _, element := range this.list(this.field(frames, "m_element")) {
		data := blob.ElementData{Kind: this.number(element, "m_t")}
		for _, prop := range this.list(this.field(element, "m_properties")) {
			this.addProp(&data, prop)
		}
		cartoon.Elements = append(cartoon.Elements, data)
	}
	return cartoon
}

func (this gcStream) asObject(value any) *nrbfObject {
	object, _ := this.value(value).(*nrbfObject)
	return object
}

func (this gcStream) field(value any, name string) any {
	object := this.asObject(value)
	if object == nil {
		this.fail()
		return nil
	}
	return this.value(object.get(name))
}

// number reads an integer member; enums are objects with a value__ member.
func (this gcStream) number(value any, name string) int {
	switch number := this.field(value, name).(type) {
	case int32:
		return int(number)
	case int16:
		return int(number)
	case uint16:
		return int(number)
	case byte:
		return int(number)
	case int8:
		return int(number)
	case uint32:
		return int(number)
	case int64:
		return int(number)
	case *nrbfObject:
		return this.number(number, "value__")
	}
	this.fail()
	return 0
}

func (this gcStream) text(value any, name string) string {
	switch text := this.field(value, name).(type) {
	case string:
		return text
	case nil:
		return ""
	}
	this.fail()
	return ""
}

func (this gcStream) flag(value any, name string) bool {
	flag, ok := this.field(value, name).(bool)
	if !ok {
		this.fail()
	}
	return flag
}

func (this gcStream) bytes(value any, name string) []byte {
	data, ok := this.field(value, name).([]byte)
	if !ok {
		this.fail()
	}
	return data
}

// list returns the items of a System.Collections.Generic.List.
func (this gcStream) list(value any) []any {
	if this.asObject(value) == nil {
		this.fail()
		return nil
	}
	items, _ := this.field(value, "_items").([]any)
	size := this.number(value, "_size")
	if size < 0 || size > len(items) {
		this.fail()
		return nil
	}
	result := make([]any, size)
	for i := range result {
		result[i] = this.value(items[i])
	}
	return result
}

// sourceList returns the items of the root SourceList<T>.
func (this gcStream) sourceList() []any {
	return this.list(this.field(this.root, "m_list"))
}

// addProp appends a FrameGraph, FrameBall or FrameTitle prop to the element of its kind.
func (this gcStream) addProp(element *blob.ElementData, prop any) {
	start, end := this.number(prop, "m_startIndex"), this.number(prop, "m_endIndex")
	switch element.Kind {
	case blob.SetGraph:
		element.Graphs = append(element.Graphs, blob.GraphProp{
			Start: start, End: end,
			Graph: this.text(prop, "graphName"), Palette: this.text(prop, "palleteName"),
			X0: this.number(prop, "startX"), X1: this.number(prop, "endX"),
			Y0: this.number(prop, "startY"), Y1: this.number(prop, "endY"),
			Z0: this.number(prop, "startZ"), Z1: this.number(prop, "endZ"),
			Angle0: this.number(prop, "startAngle"), Angle1: this.number(prop, "endAngle"),
			Axis: this.number(prop, "axis"), Mode: this.number(prop, "graphType"),
		})
	case blob.DrawBall:
		element.Balls = append(element.Balls, blob.BallProp{
			Start: start, End: end,
			X: this.number(prop, "x"), Y: this.number(prop, "y"), Z: this.number(prop, "z"),
			R0: this.number(prop, "startRadius"), R1: this.number(prop, "endRadius"),
			Hollow:  this.flag(prop, "isEmpty"),
			ColorIn: byte(this.number(prop, "colorIn")), ColorOut: byte(this.number(prop, "colorOut")),
		})
	case blob.TitleStyle:
		element.Titles = append(element.Titles, blob.TitleProp{
			Start: start, End: end,
			Graph: this.text(prop, "graphName"), Palette: this.text(prop, "palleteName"),
			Index: this.text(prop, "indexName"),
			High:  this.number(prop, "high"), Style: this.number(prop, "style"),
		})
	default:
		this.fail()
	}
}
