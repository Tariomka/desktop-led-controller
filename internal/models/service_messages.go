package models

import (
	"github.com/Tariomka/desktop-led-controller/internal/data"
)

type AddToBufferMessage struct{ Frame data.CubeFrame }
type RenameMessage struct{ Name string }
type LoadFrameMessage struct{ Index uint32 }

type FetchMessage struct{}
type LoadMessage struct{ Name string }
type SaveMessage struct{}
type ResetMessage struct{}
