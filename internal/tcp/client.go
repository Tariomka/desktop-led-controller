package tcp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
)

type LedClient struct {
	address string

	channel chan any

	connected   bool
	connection  net.Conn
	sendChannel chan []byte
}

func NewClient(ip string, port uint16) *LedClient {
	client := &LedClient{
		address: fmt.Sprintf("%s:%d", ip, port),
		channel: make(chan any, 1),
	}

	go client.channelLoop()
	global.SetTcpClientChannel(client.channel)

	return client
}

func (lc *LedClient) Start(data []byte) {
	connection, err := net.Dial("tcp", lc.address)
	// add retry policy
	if err != nil {
		log.Fatal(err) // change to printf
	}
	defer connection.Close()

	payload := make([]byte, 0, 1+1+len(data))
	payload = append(payload, 1, 0)
	payload = append(payload, data...)

	var n int
	for range 3 {
		n, err = connection.Write(payload)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Printf("client error: %v\n", err)
	}

	log.Printf("client sent %d bytes, content: %s\n", n, payload)
}

func (this *LedClient) Connect() {
	connection, err := net.Dial("tcp", this.address)
	if err != nil {
		println(err.Error())
		global.SendToUi(models.DisconnectedMessage{})
		return
	}

	this.connection = connection
	this.connected = true
	// defer this.Disconnect()

	go this.receive()
	go this.send()

	global.SendToUi(models.ConnectedMessage{})
	// for this.connected {
	// Just to block
	// }
}

func (this *LedClient) Disconnect() {
	this.connected = false
	if this.connection != nil {
		this.connection.Close()
		this.connection = nil
	}
	global.SendToUi(models.DisconnectedMessage{})
}

// Blocking state loop
func (this *LedClient) channelLoop() {
	for {
		select {
		case message := <-this.channel:
			switch message.(type) {
			case models.TCPConnectMessage:
				println("received connect")
				this.Connect()
			case models.TCPDisconnectMessage:
				println("received disconnect")
				this.Disconnect()
			case models.TCPSendPacketMessage:
				this.sendChannel <- message.(models.TCPSendPacketMessage).Data
			}
		}
	}
}

func (this *LedClient) receive() {
	for this.connected {

	}
}

func (this *LedClient) send() {
	for this.connected {
		data := <-this.sendChannel
		this.connection.Write(data)
	}
}
