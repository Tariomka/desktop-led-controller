package models

import "github.com/Tariomka/desktop-led-controller/internal/common"

type ConnectedMessage struct{}

type DisconnectedMessage struct{}

type FillVisibleCubesMessage struct{}

type SetFrameMessage struct{ Frame common.CubeFrame }
