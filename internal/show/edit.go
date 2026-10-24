package show

import (
	"encoding/binary"

	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// Axes of graph operations: X (length), Y (height, up), Z (width).
const (
	AxisX = 0
	AxisY = 1
	AxisZ = 2
)

// ---- Graph tools ----

// VoxelIndex is the index of (x, y, z) in a graph's voxels.
func VoxelIndex(graph *blob.GraphData, x, y, z int) int {
	return x + y*graph.L + z*graph.L*graph.H
}

// Mirror flips the graph along an axis.
func Mirror(graph *blob.GraphData, axis int) {
	result := make([]byte, len(graph.Voxels))
	for z := range graph.W {
		for y := range graph.H {
			for x := range graph.L {
				sx, sy, sz := x, y, z
				switch axis {
				case AxisX:
					sx = graph.L - 1 - x
				case AxisY:
					sy = graph.H - 1 - y
				case AxisZ:
					sz = graph.W - 1 - z
				}
				result[VoxelIndex(graph, x, y, z)] = graph.Voxels[VoxelIndex(graph, sx, sy, sz)]
			}
		}
	}
	graph.Voxels = result
}

// Rotate turns the graph 90 degrees about an axis; the two other sizes are swapped.
func Rotate(graph *blob.GraphData, axis int) {
	l, h, w := graph.L, graph.H, graph.W
	var result blob.GraphData
	switch axis {
	case AxisX: // y -> z
		result = NewGraph(graph.Name, l, w, h)
		for z := range w {
			for y := range h {
				for x := range l {
					result.Voxels[VoxelIndex(&result, x, z, h-1-y)] = graph.Voxels[VoxelIndex(graph, x, y, z)]
				}
			}
		}
	case AxisY: // z -> x
		result = NewGraph(graph.Name, w, h, l)
		for z := range w {
			for y := range h {
				for x := range l {
					result.Voxels[VoxelIndex(&result, w-1-z, y, x)] = graph.Voxels[VoxelIndex(graph, x, y, z)]
				}
			}
		}
	default: // AxisZ: x -> y
		result = NewGraph(graph.Name, h, l, w)
		for z := range w {
			for y := range h {
				for x := range l {
					result.Voxels[VoxelIndex(&result, h-1-y, x, z)] = graph.Voxels[VoxelIndex(graph, x, y, z)]
				}
			}
		}
	}
	*graph = result
}

// Shift moves the voxels by delta along an axis; voxels pushed out are lost.
func Shift(graph *blob.GraphData, axis, delta int) {
	result := make([]byte, len(graph.Voxels))
	for z := range graph.W {
		for y := range graph.H {
			for x := range graph.L {
				sx, sy, sz := x, y, z
				switch axis {
				case AxisX:
					sx -= delta
				case AxisY:
					sy -= delta
				case AxisZ:
					sz -= delta
				}
				if sx < 0 || sx >= graph.L || sy < 0 || sy >= graph.H || sz < 0 || sz >= graph.W {
					continue
				}
				result[VoxelIndex(graph, x, y, z)] = graph.Voxels[VoxelIndex(graph, sx, sy, sz)]
			}
		}
	}
	graph.Voxels = result
}

// FloodFill replaces the connected region (6-neighbourhood) of the same value around (x, y, z).
func FloodFill(graph *blob.GraphData, x, y, z int, value byte) {
	if x < 0 || x >= graph.L || y < 0 || y >= graph.H || z < 0 || z >= graph.W {
		return
	}
	target := graph.Voxels[VoxelIndex(graph, x, y, z)]
	if target == value {
		return
	}

	stack := [][3]int{{x, y, z}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if p[0] < 0 || p[0] >= graph.L || p[1] < 0 || p[1] >= graph.H || p[2] < 0 || p[2] >= graph.W {
			continue
		}
		index := VoxelIndex(graph, p[0], p[1], p[2])
		if graph.Voxels[index] != target {
			continue
		}
		graph.Voxels[index] = value
		stack = append(stack,
			[3]int{p[0] + 1, p[1], p[2]}, [3]int{p[0] - 1, p[1], p[2]},
			[3]int{p[0], p[1] + 1, p[2]}, [3]int{p[0], p[1] - 1, p[2]},
			[3]int{p[0], p[1], p[2] + 1}, [3]int{p[0], p[1], p[2] - 1})
	}
}

// CopyGraph returns a deep copy with a new name.
func CopyGraph(graph blob.GraphData, name string) blob.GraphData {
	graph.Name = name
	graph.Voxels = append([]byte(nil), graph.Voxels...)
	return graph
}

// ---- Paths (index data) ----

// Path points are int16 little-endian (x, z) pairs, 4 bytes per point.
type PathPoint struct{ X, Z int }

func PathPoints(index blob.IndexData) []PathPoint {
	points := make([]PathPoint, 0, len(index.Data)/4)
	for i := 0; i+4 <= len(index.Data); i += 4 {
		points = append(points, PathPoint{
			X: int(int16(binary.LittleEndian.Uint16(index.Data[i:]))),
			Z: int(int16(binary.LittleEndian.Uint16(index.Data[i+2:]))),
		})
	}
	return points
}

func PathData(points []PathPoint) []byte {
	data := make([]byte, 0, 4*len(points))
	for _, point := range points {
		data = binary.LittleEndian.AppendUint16(data, uint16(int16(point.X)))
		data = binary.LittleEndian.AppendUint16(data, uint16(int16(point.Z)))
	}
	return data
}

// EdgePath walks around the border of an l x w top view, starting at the front left, clockwise.
func EdgePath(l, w int) []PathPoint {
	var points []PathPoint
	for x := 0; x < l; x++ {
		points = append(points, PathPoint{x, 0})
	}
	for z := 1; z < w; z++ {
		points = append(points, PathPoint{l - 1, z})
	}
	for x := l - 2; x >= 0; x-- {
		points = append(points, PathPoint{x, w - 1})
	}
	for z := w - 2; z >= 1; z-- {
		points = append(points, PathPoint{0, z})
	}
	return points
}

// LinePath is a straight row along x at the given z.
func LinePath(l, z int) []PathPoint {
	points := make([]PathPoint, l)
	for x := range points {
		points[x] = PathPoint{x, z}
	}
	return points
}

func ReversePath(points []PathPoint) []PathPoint {
	result := make([]PathPoint, len(points))
	for i, point := range points {
		result[len(points)-1-i] = point
	}
	return result
}

// ---- Cartoons, elements and props ----

// Flatten returns, for each enabled cartoon, its position in the compiled show (groups in
// order), or -1 for disabled ones.
func (this *Project) Flatten() [blob.GroupCount][]int {
	var result [blob.GroupCount][]int
	index := 0
	for group, cartoons := range this.Groups {
		result[group] = make([]int, len(cartoons))
		for i, cartoon := range cartoons {
			if cartoon.Enabled {
				result[group][i] = index
				index++
			} else {
				result[group][i] = -1
			}
		}
	}
	return result
}

// NewElement returns an element of the kind with one default prop spanning all frames.
func NewElement(kind int, frames int) blob.ElementData {
	element := blob.ElementData{Kind: kind}
	AddProp(&element, frames)
	return element
}

// AddProp appends a default prop of the element's kind.
func AddProp(element *blob.ElementData, frames int) {
	end := max(frames-1, 0)
	switch element.Kind {
	case blob.SetGraph:
		element.Graphs = append(element.Graphs, blob.GraphProp{End: end})
	case blob.DrawBall:
		element.Balls = append(element.Balls, blob.BallProp{End: end, X: 350, Y: 350, Z: 350, R0: 100, R1: 300, ColorIn: 0x30, ColorOut: 0x30})
	case blob.TitleStyle:
		element.Titles = append(element.Titles, blob.TitleProp{End: end})
	}
}

// DuplicateProp copies prop i of the element after itself.
func DuplicateProp(element *blob.ElementData, i int) {
	switch element.Kind {
	case blob.SetGraph:
		if i >= 0 && i < len(element.Graphs) {
			element.Graphs = insertAt(element.Graphs, i+1, element.Graphs[i])
		}
	case blob.DrawBall:
		if i >= 0 && i < len(element.Balls) {
			element.Balls = insertAt(element.Balls, i+1, element.Balls[i])
		}
	case blob.TitleStyle:
		if i >= 0 && i < len(element.Titles) {
			element.Titles = insertAt(element.Titles, i+1, element.Titles[i])
		}
	}
}

func DeleteProp(element *blob.ElementData, i int) {
	switch element.Kind {
	case blob.SetGraph:
		element.Graphs = deleteAt(element.Graphs, i)
	case blob.DrawBall:
		element.Balls = deleteAt(element.Balls, i)
	case blob.TitleStyle:
		element.Titles = deleteAt(element.Titles, i)
	}
}

// PropSpan returns the start and end frame of prop i.
func PropSpan(element blob.ElementData, i int) (start, end int) {
	switch element.Kind {
	case blob.SetGraph:
		if i < len(element.Graphs) {
			return element.Graphs[i].Start, element.Graphs[i].End
		}
	case blob.DrawBall:
		if i < len(element.Balls) {
			return element.Balls[i].Start, element.Balls[i].End
		}
	case blob.TitleStyle:
		if i < len(element.Titles) {
			return element.Titles[i].Start, element.Titles[i].End
		}
	}
	return 0, 0
}

// SetPropSpan moves prop i to the frames start..end (kept in order and non-negative).
func SetPropSpan(element *blob.ElementData, i, start, end int) {
	start = max(start, 0)
	end = max(end, start)
	switch element.Kind {
	case blob.SetGraph:
		if i < len(element.Graphs) {
			element.Graphs[i].Start, element.Graphs[i].End = start, end
		}
	case blob.DrawBall:
		if i < len(element.Balls) {
			element.Balls[i].Start, element.Balls[i].End = start, end
		}
	case blob.TitleStyle:
		if i < len(element.Titles) {
			element.Titles[i].Start, element.Titles[i].End = start, end
		}
	}
}

// CopyCartoon returns a deep copy with a new name.
func CopyCartoon(cartoon Cartoon, name string) Cartoon {
	cartoon.Name = name
	cartoon.Script = append([]uint32(nil), cartoon.Script...)
	elements := make([]blob.ElementData, len(cartoon.Elements))
	for i, element := range cartoon.Elements {
		element.Graphs = append([]blob.GraphProp(nil), element.Graphs...)
		element.Balls = append([]blob.BallProp(nil), element.Balls...)
		element.Titles = append([]blob.TitleProp(nil), element.Titles...)
		elements[i] = element
	}
	cartoon.Elements = elements
	return cartoon
}

// UniqueName returns base, or base_N, that exists reports false for.
func UniqueName(base string, exists func(string) bool) string {
	if !exists(base) {
		return base
	}
	for i := 1; ; i++ {
		name := base + "_" + itoa(i)
		if !exists(name) {
			return name
		}
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

func insertAt[T any](items []T, i int, item T) []T {
	items = append(items, item)
	copy(items[i+1:], items[i:])
	items[i] = item
	return items
}

func deleteAt[T any](items []T, i int) []T {
	if i < 0 || i >= len(items) {
		return items
	}
	return append(items[:i], items[i+1:]...)
}

// MoveItem moves items[i] to position j.
func MoveItem[T any](items []T, i, j int) {
	if i < 0 || j < 0 || i >= len(items) || j >= len(items) || i == j {
		return
	}
	item := items[i]
	if i < j {
		copy(items[i:], items[i+1:j+1])
	} else {
		copy(items[j+1:], items[j:i])
	}
	items[j] = item
}
