package show

import (
	"github.com/Tariomka/led-common-lib/pkg/lightshow"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/blob"
	"github.com/Tariomka/led-common-lib/pkg/lightshow/render"
)

type PreviewMode int

const (
	// ModeShow plays the whole compiled show with the cube's engine (keys, play model, ...).
	ModeShow PreviewMode = iota
	// ModeCartoon loops one cartoon frame by frame, with stepping and scrubbing.
	ModeCartoon
)

// Speed factors in tenths, as the cube's engine (index 4 = 1.0).
var speedTenths = [8]int{1, 2, 5, 7, 10, 15, 20, 30}

const defaultSpeed = 4

// Preview plays the compiled project with the cube's show engine, on a canvas with one display
// per cube of the show's display table.
type Preview struct {
	Player *lightshow.Player
	Canvas *render.Canvas
	show   *lightshow.BlobShow
	// Level is the audio level 0..63 used for Audio cartoons.
	Level int

	Mode    PreviewMode
	Paused  bool
	cartoon int // ModeCartoon: index in the compiled show
	frame   int
	speed   int
	next    uint32
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

	preview := &Preview{Canvas: canvas, speed: defaultSpeed}
	preview.show = lightshow.NewBlobShow(parsed, 0, 1, 2, 3)
	preview.Player = lightshow.New(preview.show, canvas)
	preview.Player.AudioFrame = func(int) int { return preview.Level }
	return preview, nil
}

// Count is the number of compiled cartoons.
func (this *Preview) Count() int { return this.show.Count() }

// Current is the index of the playing cartoon in the compiled show.
func (this *Preview) Current() int {
	if this.Mode == ModeCartoon {
		return this.cartoon
	}
	return this.Player.Current()
}

// Frame is the frame shown in ModeCartoon (unknown in ModeShow: -1).
func (this *Preview) Frame() int {
	if this.Mode == ModeCartoon {
		return this.frame
	}
	return -1
}

// FrameCount is the number of frames of the cartoon in ModeCartoon.
func (this *Preview) FrameCount() int {
	if this.Mode != ModeCartoon || this.cartoon >= this.Count() {
		return 0
	}
	cartoon := this.show.Cartoon(this.cartoon)
	if cartoon.Kind() == blob.Audio {
		return 64
	}
	return max(cartoon.Header().FrameCount, 1)
}

func (this *Preview) Speed() int { return this.speed }

func (this *Preview) SpeedText() string {
	tenths := speedTenths[this.speed]
	return itoa(tenths/10) + "." + itoa(tenths%10) + "0"
}

// AddSpeed changes the speed index by delta within 0..7 (both modes).
func (this *Preview) AddSpeed(delta int) {
	this.speed = max(0, min(this.speed+delta, len(speedTenths)-1))
	this.Player.SetSpeed(this.speed)
}

// PlayShow starts the whole show at cartoon index.
func (this *Preview) PlayShow(index int) {
	this.Mode, this.Paused = ModeShow, false
	this.Player.Select(index)
	this.Player.Paused = false
	this.Restart()
}

// PlayCartoon loops cartoon index (compiled position) frame by frame.
func (this *Preview) PlayCartoon(index int) {
	if index < 0 || index >= this.Count() {
		return
	}
	this.Mode, this.cartoon, this.frame, this.Paused = ModeCartoon, index, 0, false
	this.Restart()
}

// SetFrame shows one frame of the cartoon and pauses (scrubbing).
func (this *Preview) SetFrame(frame int) {
	if this.Mode != ModeCartoon {
		return
	}
	count := this.FrameCount()
	this.frame = ((frame % count) + count) % count
	this.Paused = true
	this.render()
}

// Step moves by delta frames and pauses.
func (this *Preview) Step(delta int) { this.SetFrame(this.frame + delta) }

func (this *Preview) TogglePause() {
	this.Paused = !this.Paused
	if this.Mode == ModeShow {
		this.Player.Paused = this.Paused
	}
	this.Restart()
}

// Update advances the show when the current frame's time is over; now is a millisecond clock.
// It returns true when a new frame was drawn.
func (this *Preview) Update(now uint32) bool {
	if int32(now-this.next) < 0 {
		return false
	}
	if this.Mode == ModeShow {
		rendered, wait := this.Player.Step(now)
		this.next = now + uint32(max(wait, 1))
		return rendered
	}

	if this.Paused || this.cartoon >= this.Count() {
		this.next = now + 100
		return false
	}
	this.render()
	delay := this.show.Cartoon(this.cartoon).Header().Delay
	this.next = now + uint32(max(delay*10/speedTenths[this.speed], 1))
	this.frame = (this.frame + 1) % this.FrameCount()
	return true
}

func (this *Preview) render() {
	cartoon := this.show.Cartoon(this.cartoon)
	frame := this.frame
	if cartoon.Kind() == blob.Audio {
		frame = this.Level
	}
	this.Canvas.Clear(0)
	cartoon.Render(frame, this.Canvas)
}

// Restart makes the next Update step immediately (after a key or selection change).
func (this *Preview) Restart() { this.next = 0 }
