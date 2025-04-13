package tcp

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
)

type LedClient struct {
	address string

	channel chan any
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

	time.Sleep(100 * time.Millisecond)
	if rand.Intn(3) == 0 {
		global.SendToUi(models.ConnectedMessage{})
	} else {
		global.SendToUi(models.DisconnectedMessage{})
	}
}

func (this *LedClient) Disconnect() {
	time.Sleep(100 * time.Millisecond)
	global.SendToUi(models.DisconnectedMessage{})
}

// Blocking state loop
func (this *LedClient) channelLoop() {
	for {
		select {
		case message := <-this.channel:
			switch message.(type) {
			case models.ConnectRequestMessage:
				println("received connect")
				this.Connect()
			case models.DisconnectRequestMessage:
				println("received disconnect")
				this.Disconnect()
			}
		}
	}
}
