package global

import "github.com/Tariomka/desktop-led-controller/internal/models"

var (
	tcpClientChannel chan<- any
	uiChannel        chan<- any
)

func SetTcpClientChannel(channel chan<- any) { tcpClientChannel = channel }

func SendToClient(message any) {
	if tcpClientChannel == nil {
		SendToUi(models.DisconnectedMessage{})
		return
	}

	tcpClientChannel <- message
}

func SetUiChannel(channel chan<- any) { uiChannel = channel }

func SendToUi(message any) {
	if uiChannel == nil {
		return
	}

	uiChannel <- message
}
