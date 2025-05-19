package models

import (
	"github.com/Tariomka/desktop-led-controller/internal/data"
)

type AddFrameMessage struct {
	Frame data.CubeFrame
	Index uint32
}

type RemoveFrameMessage struct{ Index uint32 }
type RenderFrameMessage struct{ Index uint32 }

type SendMessage struct{}
type RenameMessage struct{ Name string }

type FetchMessage struct{}
type LoadMessage struct{ Name string }
type SaveMessage struct{}
type ResetMessage struct{}
