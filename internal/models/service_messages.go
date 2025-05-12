package models

import "github.com/Tariomka/desktop-led-controller/internal/common"

type AddToBufferMessage struct{ Frame common.CubeFrame }
type RenameMessage struct{ Name string }
type LoadFrameMessage struct{ Index uint32 }

type FetchMessage struct{}
type LoadMessage struct{}
type SaveMessage struct{}
type ResetMessage struct{}
