package ui

import (
	"github.com/Tariomka/desktop-led-controller/internal/config"
	"github.com/Tariomka/desktop-led-controller/internal/ui/component"
	"github.com/Tariomka/desktop-led-controller/internal/ui/style"
	"github.com/Tariomka/desktop-led-controller/internal/ui/utils"
	raylib "github.com/gen2brain/raylib-go/raylib"
)

const (
	topBarHeight    = 44
	statusBarHeight = 24
	stripSize       = 24 // collapsed panel strip
	splitterSize    = 6
	minLeftWidth    = 180
	maxLeftWidth    = 480
	minRightWidth   = 260
	maxRightWidth   = 560
	minBottomHeight = 120
	maxBottomHeight = 420
)

type splitter int

const (
	splitNone splitter = iota
	splitLeft
	splitRight
	splitBottom
)

// Layout computes the screen regions from the panel settings and handles the collapse handles
// and the draggable splitters between panels.
type Layout struct {
	TopBar, StatusBar, Left, Center, Right, Bottom raylib.Rectangle

	dragging splitter
}

// Update recomputes the regions for the window size and applies mouse interaction with the
// handles and splitters; it reports whether the panel settings changed.
func (this *Layout) Update(panels *config.Panels, width, height float32, mouseFree bool) bool {
	changed := this.drag(panels, width, height, mouseFree)
	this.compute(panels, width, height)
	if mouseFree && this.handles(panels) {
		this.compute(panels, width, height)
		changed = true
	}
	return changed
}

func (this *Layout) compute(panels *config.Panels, width, height float32) {
	this.TopBar = raylib.NewRectangle(0, 0, width, topBarHeight)
	this.StatusBar = raylib.NewRectangle(0, height-statusBarHeight, width, statusBarHeight)
	top, bottom := float32(topBarHeight), height-statusBarHeight

	left := float32(stripSize)
	if panels.LeftOpen {
		left = panels.LeftWidth
	}
	right := float32(stripSize)
	if panels.RightOpen {
		right = panels.RightWidth
	}
	bottomHeight := float32(stripSize)
	if panels.BottomOpen {
		bottomHeight = panels.BottomHeight
	}

	this.Left = raylib.NewRectangle(0, top, left, bottom-top-bottomHeight)
	this.Right = raylib.NewRectangle(width-right, top, right, bottom-top-bottomHeight)
	this.Bottom = raylib.NewRectangle(0, bottom-bottomHeight, width, bottomHeight)
	this.Center = raylib.NewRectangle(left, top, width-left-right, bottom-top-bottomHeight)
}

// splitterBounds returns the drag band of an open panel.
func (this *Layout) splitterBounds(which splitter) raylib.Rectangle {
	switch which {
	case splitLeft:
		return raylib.NewRectangle(this.Left.X+this.Left.Width-splitterSize/2, this.Left.Y, splitterSize, this.Left.Height)
	case splitRight:
		return raylib.NewRectangle(this.Right.X-splitterSize/2, this.Right.Y, splitterSize, this.Right.Height)
	case splitBottom:
		return raylib.NewRectangle(this.Bottom.X, this.Bottom.Y-splitterSize/2, this.Bottom.Width, splitterSize)
	}
	return raylib.Rectangle{}
}

func (this *Layout) drag(panels *config.Panels, width, height float32, mouseFree bool) bool {
	mouse := raylib.GetMousePosition()
	if this.dragging == splitNone {
		if !mouseFree || !raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
			return false
		}
		for _, which := range []splitter{splitLeft, splitRight, splitBottom} {
			open := map[splitter]bool{splitLeft: panels.LeftOpen, splitRight: panels.RightOpen, splitBottom: panels.BottomOpen}[which]
			if open && raylib.CheckCollisionPointRec(mouse, this.splitterBounds(which)) {
				this.dragging = which
			}
		}
		return false
	}

	if raylib.IsMouseButtonReleased(raylib.MouseLeftButton) {
		this.dragging = splitNone
		return true
	}
	switch this.dragging {
	case splitLeft:
		panels.LeftWidth = clamp(mouse.X, minLeftWidth, min(maxLeftWidth, width/2))
	case splitRight:
		panels.RightWidth = clamp(width-mouse.X, minRightWidth, min(maxRightWidth, width/2))
	case splitBottom:
		panels.BottomHeight = clamp(height-statusBarHeight-mouse.Y, minBottomHeight, min(maxBottomHeight, height/2))
	}
	return false
}

// Dragging reports whether a splitter is being dragged (the view must ignore the mouse).
func (this *Layout) Dragging() bool { return this.dragging != splitNone }

// handleBounds returns the collapse/expand button of a panel.
func (this *Layout) handleBounds(which splitter) raylib.Rectangle {
	switch which {
	case splitLeft:
		return raylib.NewRectangle(this.Left.X+this.Left.Width-16, this.Left.Y+6, 14, 22)
	case splitRight:
		return raylib.NewRectangle(this.Right.X+2, this.Right.Y+6, 14, 22)
	case splitBottom:
		return raylib.NewRectangle(this.Bottom.X+this.Bottom.Width/2-12, this.Bottom.Y+2, 24, 12)
	}
	return raylib.Rectangle{}
}

func (this *Layout) handles(panels *config.Panels) bool {
	if !raylib.IsMouseButtonPressed(raylib.MouseLeftButton) {
		return false
	}
	mouse := raylib.GetMousePosition()
	switch {
	case raylib.CheckCollisionPointRec(mouse, this.handleBounds(splitLeft)):
		panels.LeftOpen = !panels.LeftOpen
	case raylib.CheckCollisionPointRec(mouse, this.handleBounds(splitRight)):
		panels.RightOpen = !panels.RightOpen
	case raylib.CheckCollisionPointRec(mouse, this.handleBounds(splitBottom)):
		panels.BottomOpen = !panels.BottomOpen
	default:
		return false
	}
	return true
}

// DrawChrome draws the collapse handles, the collapsed strips and the splitter hover.
func (this *Layout) DrawChrome(panels *config.Panels) {
	mouse := raylib.GetMousePosition()
	for _, which := range []splitter{splitLeft, splitRight, splitBottom} {
		open := map[splitter]bool{splitLeft: panels.LeftOpen, splitRight: panels.RightOpen, splitBottom: panels.BottomOpen}[which]
		if !open {
			this.drawStrip(which)
		}
		if open && (this.dragging == which || raylib.CheckCollisionPointRec(mouse, this.splitterBounds(which))) {
			raylib.DrawRectangleRec(this.splitterBounds(which), raylib.Fade(component.AccentColor, 0.5))
			if which == splitBottom {
				raylib.SetMouseCursor(int32(raylib.MouseCursorResizeNS))
			} else {
				raylib.SetMouseCursor(int32(raylib.MouseCursorResizeEW))
			}
		}
		this.drawHandle(which, open)
	}
}

func (this *Layout) drawStrip(which splitter) {
	var bounds raylib.Rectangle
	label := ""
	switch which {
	case splitLeft:
		bounds, label = this.Left, "Assets"
	case splitRight:
		bounds, label = this.Right, "Show  Props  Tools  Cube"
	case splitBottom:
		bounds, label = this.Bottom, "Timeline  Console"
	}
	raylib.DrawRectangleRec(bounds, component.StripColor)
	raylib.DrawRectangleLinesEx(bounds, 1, style.BorderColor)
	if which == splitBottom {
		utils.RenderText(label, raylib.NewRectangle(bounds.X+40, bounds.Y, 300, bounds.Height), component.MutedColor)
		return
	}
	raylib.DrawTextPro(style.GuiFont, label,
		raylib.NewVector2(bounds.X+bounds.Width/2+style.TextSize/2, bounds.Y+60),
		raylib.NewVector2(0, 0), 90, style.TextSize, style.TextSpacing, component.MutedColor)
}

func (this *Layout) drawHandle(which splitter, open bool) {
	bounds := this.handleBounds(which)
	hover := raylib.CheckCollisionPointRec(raylib.GetMousePosition(), bounds)
	fill := style.BorderColor
	if hover {
		fill = component.AccentColor
	}
	raylib.DrawRectangleRounded(bounds, 0.3, 4, fill)
	arrow := ""
	switch {
	case which == splitLeft && open, which == splitRight && !open:
		arrow = "<"
	case which == splitLeft && !open, which == splitRight && open:
		arrow = ">"
	case which == splitBottom && open:
		arrow = "v"
	default:
		arrow = "^"
	}
	utils.RenderText(arrow, bounds, raylib.White)
}

func clamp(value, low, high float32) float32 { return max(low, min(value, high)) }
