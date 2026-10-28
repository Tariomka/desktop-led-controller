package show_test

import (
	"reflect"
	"testing"

	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

func graph3() blob.GraphData {
	// 3x2x2 with a distinct value per voxel.
	graph := show.NewGraph("g", 3, 2, 2)
	for i := range graph.Voxels {
		graph.Voxels[i] = byte(i + 1)
	}
	return graph
}

func at(graph *blob.GraphData, x, y, z int) byte {
	return graph.Voxels[show.VoxelIndex(graph, x, y, z)]
}

func TestMirror(t *testing.T) {
	graph := graph3()
	original := graph3()
	show.Mirror(&graph, show.AxisX)
	for z := range 2 {
		for y := range 2 {
			for x := range 3 {
				if at(&graph, x, y, z) != at(&original, 2-x, y, z) {
					t.Fatalf("mirror x wrong at %d,%d,%d", x, y, z)
				}
			}
		}
	}
	show.Mirror(&graph, show.AxisX)
	if !reflect.DeepEqual(graph, original) {
		t.Fatal("mirroring twice is not the identity")
	}
}

func TestRotateFourTimesIsIdentity(t *testing.T) {
	for _, axis := range []int{show.AxisX, show.AxisY, show.AxisZ} {
		graph := graph3()
		for range 4 {
			show.Rotate(&graph, axis)
		}
		if !reflect.DeepEqual(graph, graph3()) {
			t.Fatalf("axis %d: four rotations differ", axis)
		}
	}
	graph := graph3()
	show.Rotate(&graph, show.AxisY)
	if graph.L != 2 || graph.H != 2 || graph.W != 3 {
		t.Fatalf("rotate y: size %dx%dx%d", graph.L, graph.H, graph.W)
	}
	// x -> z: voxel (0,0,0) moves to z = 0, x = W-1-0 = 1
	if at(&graph, 1, 0, 0) != at(&[]blob.GraphData{graph3()}[0], 0, 0, 0) {
		t.Fatal("rotate y moved the corner wrongly")
	}
}

func TestShift(t *testing.T) {
	graph := graph3()
	show.Shift(&graph, show.AxisX, 1)
	if at(&graph, 0, 0, 0) != 0 || at(&graph, 1, 0, 0) != 1 || at(&graph, 2, 0, 0) != 2 {
		t.Fatalf("shift x: %v", graph.Voxels)
	}
	show.Shift(&graph, show.AxisX, -1)
	if at(&graph, 0, 0, 0) != 1 || at(&graph, 2, 0, 0) != 0 {
		t.Fatalf("shift back: %v", graph.Voxels)
	}
}

func TestFloodFill(t *testing.T) {
	graph := show.NewGraph("g", 4, 1, 1)
	graph.Voxels[2] = 9 // wall
	show.FloodFill(&graph, 0, 0, 0, 5)
	if !reflect.DeepEqual(graph.Voxels, []byte{5, 5, 9, 0}) {
		t.Fatalf("got %v", graph.Voxels)
	}
}

func TestPathRoundTrip(t *testing.T) {
	points := show.EdgePath(8, 8)
	if len(points) != 28 {
		t.Fatalf("edge path has %d points", len(points))
	}
	data := show.PathData(points)
	if len(data) != 28*4 {
		t.Fatalf("path data is %d bytes", len(data))
	}
	back := show.PathPoints(blob.IndexData{Data: data})
	if !reflect.DeepEqual(back, points) {
		t.Fatal("path points differ after encoding")
	}
	if reversed := show.ReversePath(points); reversed[0] != points[len(points)-1] {
		t.Fatal("reverse failed")
	}
	negative := show.PathData([]show.PathPoint{{X: -1, Z: -2}})
	if got := show.PathPoints(blob.IndexData{Data: negative}); got[0].X != -1 || got[0].Z != -2 {
		t.Fatalf("negative points: %+v", got)
	}
}

func TestPropOperations(t *testing.T) {
	element := show.NewElement(blob.DrawBall, 20)
	if show.PropCount(element) != 1 {
		t.Fatal("new element has no prop")
	}
	if start, end := show.PropSpan(element, 0); start != 0 || end != 19 {
		t.Fatalf("span %d-%d", start, end)
	}
	show.AddProp(&element, 20)
	show.DuplicateProp(&element, 0)
	if show.PropCount(element) != 3 {
		t.Fatalf("got %d props", show.PropCount(element))
	}
	show.SetPropSpan(&element, 1, 5, 3)
	if start, end := show.PropSpan(element, 1); start != 5 || end != 5 {
		t.Fatalf("span not normalized: %d-%d", start, end)
	}
	show.DeleteProp(&element, 0)
	show.DeleteProp(&element, 0)
	show.DeleteProp(&element, 0)
	show.DeleteProp(&element, 0) // out of range is ignored
	if show.PropCount(element) != 0 {
		t.Fatal("props left")
	}
}

func TestFlattenAndCopy(t *testing.T) {
	project := show.NewProject()
	disabled := show.NewCartoon("off", blob.Normal)
	disabled.Enabled = false
	project.Groups[0] = append(project.Groups[0], disabled, show.NewCartoon("c", blob.Normal))
	project.Groups[2] = append(project.Groups[2], show.NewCartoon("d", blob.Normal))
	flat := project.Flatten()
	if flat[0][0] != 0 || flat[0][1] != -1 || flat[0][2] != 1 || flat[2][0] != 2 {
		t.Fatalf("flatten: %v", flat)
	}

	original := project.Groups[0][2]
	original.Elements = append(original.Elements, show.NewElement(blob.SetGraph, 10))
	clone := show.CopyCartoon(original, "copy")
	clone.Elements[0].Graphs[0].X0 = 7
	if original.Elements[0].Graphs[0].X0 != 0 || clone.Name != "copy" {
		t.Fatal("copy shares props")
	}

	show.MoveItem(project.Groups[0], 0, 2)
	if project.Groups[0][2].Name != "Cartoon" || project.Groups[0][0].Name != "off" {
		t.Fatalf("move: %q, %q", project.Groups[0][0].Name, project.Groups[0][2].Name)
	}
	if name := show.UniqueName("Graph", func(name string) bool { return name == "Graph" || name == "Graph_1" }); name != "Graph_2" {
		t.Fatalf("unique name %q", name)
	}
}
