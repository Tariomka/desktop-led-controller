package global

import "github.com/Tariomka/desktop-led-controller/internal/models"

var (
	tcpClientChannel chan<- any
	uiChannel        chan<- any
)

func SetTCPClientChannel(channel chan<- any) { tcpClientChannel = channel }

func SendToClient(message any) {
	if tcpClientChannel == nil {
		SendToUI(models.DisconnectedMessage{})
		return
	}

	tcpClientChannel <- message
}

func SetUIChannel(channel chan<- any) { uiChannel = channel }

func SendToUI(message any) {
	if uiChannel == nil {
		return
	}

	uiChannel <- message
}
