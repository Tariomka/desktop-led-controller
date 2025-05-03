package global

import "github.com/Tariomka/desktop-led-controller/internal/models"

var (
	tcpClientChannel    chan<- any
	uiConnectionChannel chan<- any
	uiMessageChannel    chan<- string
)

func SetTCPClientChannel(channel chan<- any) { tcpClientChannel = channel }

func SendToClient(message any) {
	if tcpClientChannel == nil {
		SendToUIConnection(models.DisconnectedMessage{})
		return
	}

	tcpClientChannel <- message
}

func SetUIConnectionChannel(channel chan<- any) { uiConnectionChannel = channel }

func SendToUIConnection(message any) {
	if uiConnectionChannel == nil {
		return
	}

	uiConnectionChannel <- message
}

func SetUIMessageChannel(channel chan<- string) { uiMessageChannel = channel }

func SendToUIMessage(message string) {
	if uiMessageChannel == nil {
		return
	}

	uiMessageChannel <- message
}
