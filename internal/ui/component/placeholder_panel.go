package component

import raylib "github.com/gen2brain/raylib-go/raylib"

type PlaceholderPanel struct{ Panel }

func (this *PlaceholderPanel) Update() { this.resize() }

func (this *PlaceholderPanel) Render() {
	this.renderPanel()
	raylib.DrawFPS(this.ToInt32().X+10, this.ToInt32().Y+10)
}
