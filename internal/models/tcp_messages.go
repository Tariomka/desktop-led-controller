package models

type TCPConnectMessage struct{}

type TCPDisconnectMessage struct{}

// TCPCommandMessage sends a cube console command line (e.g. "Key 2") through the Pico.
type TCPCommandMessage struct{ Line string }

// TCPUploadMessage uploads compiled show data to the cube.
type TCPUploadMessage struct{ Data []byte }
