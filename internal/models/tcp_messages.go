package models

type TCPConnectMessage struct{}

type TCPDisconnectMessage struct{}

type TCPSendPacketMessage struct {
	Data []byte
}
