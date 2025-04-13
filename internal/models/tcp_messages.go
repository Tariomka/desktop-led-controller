package models

type TCPConnectMessage struct{}

type TCPDisconnectMessage struct{}

type TCPSendPacketMessage struct {
	Data []byte
	// If this a good idea? Should it be any?
	// Or maybe a Packet interface/struct with Marshal and Unmarshal should be here?
}
