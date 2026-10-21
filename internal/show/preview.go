package show

import (
	"github.com/Tariomka/led-common-lib/pkg/lightshow"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
)

// Preview plays the compiled project with the cube's show engine, on a canvas with one display
// per cube of the show's display table.
type Preview struct {
	Player *lightshow.Player
	Canvas *render.Canvas
	// Level is the audio level 0..63 used for Audio cartoons.
	Level int

	next uint32
}

func NewPreview(project *Project) (*Preview, error) {
	data, err := project.Compile()
	if err != nil {
		return nil, err
	}
	parsed, err := blob.Parse(string(data))
	if err != nil {
		return nil, err
	}

	canvas := &render.Canvas{}
	for _, display := range project.Config.Displays {
		l, h, w := display.Size()
		canvas.Displays = append(canvas.Displays, render.NewDisplay(display.X, display.Y, display.Z, l, h, w))
	}
	if len(canvas.Displays) == 0 {
		canvas.Displays = append(canvas.Displays, render.NewDisplay(0, 0, 0, 8, 8, 8))
	}

	preview := &Preview{Canvas: canvas}
	preview.Player = lightshow.New(lightshow.NewBlobShow(parsed, 0, 1, 2, 3), canvas)
	preview.Player.AudioFrame = func(int) int { return preview.Level }
	return preview, nil
}

// Update advances the show when the current frame's time is over; now is a millisecond clock.
// It returns true when a new frame was drawn.
func (this *Preview) Update(now uint32) bool {
	if int32(now-this.next) < 0 {
		return false
	}
	rendered, wait := this.Player.Step(now)
	this.next = now + uint32(max(wait, 1))
	return rendered
}

// Restart makes the next Update step immediately (after a key or selection change).
func (this *Preview) Restart() { this.next = 0 }
