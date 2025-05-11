package models

import "github.com/Tariomka/desktop-led-controller/internal/common"

type AddToBufferMessage struct{ Layout *common.CubeLayout }

type FetchMessage struct{}
type LoadMessage struct{}
type SaveMessage struct{}

type RenameMessage struct{ Name string }
