package models

import (
	"github.com/Tariomka/desktop-led-controller/internal/data"
)

type ConnectedMessage struct{}

type DisconnectedMessage struct{}

type FillVisibleCubesMessage struct{}

type SetFrameMessage struct{ Frame data.CubeFrame }

type SetLightShowsMessage struct{ Names []string }
