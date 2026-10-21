package show_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Tariomka/desktop-led-controller/internal/show"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
)

// Vendor demo projects and their compiled firmware images.
var demos = []string{
	"1-3D8RGB-Demo-20250128",
	"2-8X8X32RGB-Demo-20250128",
	"3-3D16RGB-Demo-20240619",
	"4-3D16S-and-3D16MINI-Demo-20240930",
	"5-3D8Px1Cube-Demo-20241109",
}

func read(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func hexShow(t *testing.T, name string) []byte {
	t.Helper()
	data, err := show.ShowDataFromHex(read(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return data
}

// model decodes show data into the library's model, the common ground for comparisons.
func model(t *testing.T, data []byte) *blob.Show {
	t.Helper()
	parsed, err := blob.Parse(string(data))
	if err != nil {
		t.Fatal(err)
	}
	result, err := blob.FromBlob(parsed)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func compile(t *testing.T, project *show.Project) []byte {
	t.Helper()
	data, err := project.Compile()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestHexFactoryShowIsOriginal(t *testing.T) {
	data := hexShow(t, demos[0]+".hex")
	if !bytes.Equal(data, read(t, "original.bin")) {
		t.Fatalf("show data from the factory hex differs from original.bin (%d bytes)", len(data))
	}
}

func TestHexShowSizes(t *testing.T) {
	sizes := map[string]int{"Update2.hex": 32128, "Update3.hex": 9532}
	for name, size := range sizes {
		if data := hexShow(t, name); len(data) != size {
			t.Errorf("%s: %d bytes, want %d", name, len(data), size)
		}
	}
}

func TestHexChecksumInShowData(t *testing.T) {
	lines := strings.Split(string(read(t, demos[0]+".hex")), "\n")
	for i, line := range lines {
		// first data record of the show data: upper address 0x0800 is set, offset B000
		if strings.HasPrefix(line, ":10B00000") {
			lines[i] = line[:len(line)-3] + "00" + line[len(line)-1:]
			break
		}
	}
	if _, err := show.ShowDataFromHex([]byte(strings.Join(lines, "\n"))); !errors.Is(err, show.ErrHexChecksum) {
		t.Fatalf("got %v, want ErrHexChecksum", err)
	}
}

func TestHexWithoutShow(t *testing.T) {
	if _, err := show.ShowDataFromHex([]byte(":00000001FF\n")); !errors.Is(err, show.ErrHexNoShow) {
		t.Fatalf("got %v, want ErrHexNoShow", err)
	}
}

func TestShowDataRoundTrip(t *testing.T) {
	names := []string{"Update2.hex", "Update3.hex"}
	for _, demo := range demos {
		names = append(names, demo+".hex")
	}
	for _, name := range names {
		data := hexShow(t, name)
		project, err := show.FromShowData(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(model(t, compile(t, project)), model(t, data)) {
			t.Errorf("%s: recompiled show differs", name)
		}
	}
}

func TestGCMatchesCompiledDemo(t *testing.T) {
	for _, demo := range demos {
		project, err := show.FromGC(read(t, demo+".gc"))
		if err != nil {
			t.Fatalf("%s: %v", demo, err)
		}
		got, want := model(t, compile(t, project)), model(t, hexShow(t, demo+".hex"))
		// .gc names are Unicode, the vendor compiles them as GB2312; this app keeps UTF-8.
		normalizeNames(got)
		normalizeNames(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: compiled .gc differs from the vendor image", demo)
			describeDifference(t, got, want)
		}
	}
}

// normalizeNames replaces non-ASCII names by their order of appearance.
func normalizeNames(model *blob.Show) {
	names := map[string]string{}
	rename := func(name *string) {
		for _, char := range []byte(*name) {
			if char >= 0x80 {
				if _, ok := names[*name]; !ok {
					names[*name] = fmt.Sprintf("#%d", len(names))
				}
				*name = names[*name]
				return
			}
		}
	}

	for i := range model.Palettes {
		rename(&model.Palettes[i].Name)
	}
	for i := range model.Graphs {
		rename(&model.Graphs[i].Name)
	}
	for i := range model.Indexes {
		rename(&model.Indexes[i].Name)
	}
	for _, cartoons := range model.Groups {
		for _, cartoon := range cartoons {
			for _, element := range cartoon.Elements {
				for i := range element.Graphs {
					rename(&element.Graphs[i].Graph)
					rename(&element.Graphs[i].Palette)
				}
				for i := range element.Titles {
					rename(&element.Titles[i].Graph)
					rename(&element.Titles[i].Palette)
					rename(&element.Titles[i].Index)
				}
			}
		}
	}
}

func describeDifference(t *testing.T, got, want *blob.Show) {
	t.Helper()
	if !reflect.DeepEqual(got.Config, want.Config) {
		t.Logf("config: got %+v, want %+v", got.Config, want.Config)
	}
	if !reflect.DeepEqual(got.Palettes, want.Palettes) {
		t.Logf("palettes differ: got %d, want %d", len(got.Palettes), len(want.Palettes))
	}
	if !reflect.DeepEqual(got.Graphs, want.Graphs) {
		t.Logf("graphs differ: got %d, want %d", len(got.Graphs), len(want.Graphs))
		for i := range min(len(got.Graphs), len(want.Graphs)) {
			g, w := got.Graphs[i], want.Graphs[i]
			if reflect.DeepEqual(g, w) {
				continue
			}
			t.Logf("graph %d: got %q %dx%dx%d, want %q %dx%dx%d", i, g.Name, g.L, g.H, g.W, w.Name, w.L, w.H, w.W)
			for v := range min(len(g.Voxels), len(w.Voxels)) {
				if g.Voxels[v] != w.Voxels[v] {
					t.Logf("first voxel difference at %d: got %d, want %d", v, g.Voxels[v], w.Voxels[v])
					break
				}
			}
			break
		}
	}
	if !reflect.DeepEqual(got.Indexes, want.Indexes) {
		t.Logf("indexes differ: got %d, want %d", len(got.Indexes), len(want.Indexes))
	}
	for group := range got.Groups {
		g, w := got.Groups[group], want.Groups[group]
		if len(g) != len(w) {
			t.Logf("group %d: got %d cartoons, want %d", group+1, len(g), len(w))
			continue
		}
		for i := range g {
			if !reflect.DeepEqual(g[i], w[i]) {
				t.Logf("group %d cartoon %d:\n got  %+v\n want %+v", group+1, i, g[i], w[i])
				return
			}
		}
	}
}

func TestGCNames(t *testing.T) {
	project, err := show.FromGC(read(t, demos[0]+".gc"))
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Groups[0]) == 0 || project.Groups[0][0].Name == "" {
		t.Fatal("cartoon names missing")
	}
	if len(project.Graphs) != 79 {
		t.Errorf("got %d graphs, want 79", len(project.Graphs))
	}
}

func TestGCVersion(t *testing.T) {
	data := bytes.Replace(read(t, demos[0]+".gc"), []byte("\x041.03"), []byte("\x041.04"), 1)
	if _, err := show.FromGC(data); !errors.Is(err, show.ErrGCVersion) {
		t.Fatalf("got %v, want ErrGCVersion", err)
	}
}

func TestGCTruncated(t *testing.T) {
	data := read(t, demos[0]+".gc")
	if _, err := show.FromGC(data[:len(data)/2]); err == nil {
		t.Fatal("truncated file accepted")
	}
}

func TestProjectFileRoundTrip(t *testing.T) {
	project, err := show.FromGC(read(t, demos[1]+".gc"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := project.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := show.LoadProject(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, project) {
		t.Fatal("loaded project differs")
	}
}

func TestProjectVersion(t *testing.T) {
	if _, err := show.LoadProject([]byte(`{"Version":99}`)); !errors.Is(err, show.ErrProjectVersion) {
		t.Fatalf("got %v, want ErrProjectVersion", err)
	}
}

func TestDisabledCartoonsAreSkipped(t *testing.T) {
	project := show.NewProject()
	disabled := show.NewCartoon("Off", blob.Normal)
	disabled.Enabled = false
	project.Groups[0] = append(project.Groups[0], disabled)

	if got := model(t, compile(t, project)).Groups[0]; len(got) != 1 {
		t.Fatalf("got %d cartoons, want 1", len(got))
	}
}

func TestPreviewRendersFactoryShow(t *testing.T) {
	project, err := show.FromShowData(read(t, "original.bin"))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := show.NewPreview(project)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Canvas.Displays) != 1 || len(preview.Canvas.Displays[0].Voxels) != 512 {
		t.Fatalf("unexpected canvas %+v", preview.Canvas.Displays)
	}

	frames, lit := 0, false
	for now := uint32(0); now < 5000; now += 10 {
		if preview.Update(now) {
			frames++
			for _, voxel := range preview.Canvas.Displays[0].Voxels {
				lit = lit || voxel != 0
			}
		}
	}
	if frames == 0 || !lit {
		t.Fatalf("%d frames rendered, lit voxels: %t", frames, lit)
	}
}

func TestResizeGraphKeepsVoxels(t *testing.T) {
	graph := show.NewGraph("g", 2, 2, 2)
	graph.Voxels[1+1*2+1*4] = 5
	show.ResizeGraph(&graph, 3, 3, 1)
	if len(graph.Voxels) != 9 || graph.Voxels[1+1*3] != 0 {
		t.Fatalf("got %v", graph.Voxels)
	}
	show.ResizeGraph(&graph, 2, 2, 2)
	graph.Voxels[1+1*2] = 7
	show.ResizeGraph(&graph, 4, 4, 4)
	if graph.Voxels[1+1*4] != 7 {
		t.Fatalf("voxel moved: %v", graph.Voxels)
	}
}

func TestScriptWithoutCode(t *testing.T) {
	project := show.NewProject()
	script := show.NewCartoon("Script", blob.Script)
	script.ScriptSource = "void main() {}"
	project.Groups[1] = append(project.Groups[1], script)

	if _, err := project.Compile(); !errors.Is(err, show.ErrScriptNotCompiled) {
		t.Fatalf("got %v, want ErrScriptNotCompiled", err)
	}
}
