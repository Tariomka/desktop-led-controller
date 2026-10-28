package models

type TCPConnectMessage struct{}

type TCPDisconnectMessage struct{}

// TCPConfigureMessage changes the Pico address used by the next connection.
type TCPConfigureMessage struct {
	IP   string
	Port uint16
}

// TCPCommandMessage sends a cube console command line (e.g. "Key 2") through the Pico.
type TCPCommandMessage struct{ Line string }

// TCPUploadMessage uploads compiled show data to the cube.
type TCPUploadMessage struct{ Data []byte }

// TCPUploadCancelMessage aborts a running upload by closing the connection (the Pico aborts too).
type TCPUploadCancelMessage struct{}
