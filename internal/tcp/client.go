package tcp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/led-common-lib/pkg/network"
)

type ClientConfig struct {
	IP     string
	Port   uint16
	Logger *slog.Logger
}

type LedClient struct {
	logger  *slog.Logger
	address string // Temporary until UI is updated to accept user input server address

	channel chan any

	connected   bool
	connection  net.Conn
	sendChannel chan string
}

func NewClient(config ClientConfig) *LedClient {
	// TODO: validate config

	if config.Logger == nil {
		config.Logger = common.NewConsoleLogger(slog.LevelDebug)
	}

	client := &LedClient{
		logger:      config.Logger,
		address:     fmt.Sprintf("%s:%d", config.IP, config.Port),
		channel:     make(chan any, 1),
		sendChannel: make(chan string),
	}

	go client.channelLoop()
	global.RegisterMessageReceiver(
		constants.TCPClient,
		func(message any) { client.channel <- message })

	return client
}

func (this *LedClient) Start(data []byte) { // Isn"t used
	connection, err := net.Dial("tcp", this.address)
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
		this.logger.Error("tcp client dial failure", "error", err)
		global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
		return
	}

	this.connection = connection
	this.connected = true
	this.logger.Debug(
		"action taken - Connect(ed)",
		"connected from", this.connection.LocalAddr().String(),
		"connected to", this.connection.RemoteAddr().String(),
		"connection type", this.connection.RemoteAddr().Network())

	go this.receive()
	go this.send()

	global.SendMessage(constants.UIMenuPanel, models.ConnectedMessage{})
}

func (this *LedClient) Disconnect() {
	if !this.connected {
		return
	}

	this.connected = false
	if this.connection != nil {
		this.connection.Close()
		this.connection = nil
	}
	global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
	this.logger.Debug("action taken - Disconnect(ed)")
}

// Blocking state loop
func (this *LedClient) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.TCPConnectMessage:
			this.logger.Debug("received connect")
			this.Connect()
		case models.TCPDisconnectMessage:
			this.logger.Debug("received disconnect")
			this.Disconnect()
		case models.TCPSendPacketMessage:
			if this.connected {
				this.logger.Debug("received message", "message", message)
				this.sendChannel <- message.Data
			}
		}
	}
}

func (this *LedClient) receive() {
	for this.connected {
		buffer := make([]byte, 1024)
		n, err := this.connection.Read(buffer)
		// TODO: Handle errors and data
		// TODO: Handle and send to UI when server closes
		if err != nil {
			switch {
			case errors.Is(err, net.ErrClosed):
				this.logger.Info("tcp client disconnected from server")
			case errors.Is(err, io.EOF):
				this.logger.Info("server has shutdown")
				this.Disconnect()
			default:
				this.logger.Error("tcp client read failure", "error", err.Error())
				this.Disconnect()
			}
			return
		}

		print("Data length: ")
		println(n)
		print("Data: ")
		println(buffer)
	}
}

func (this *LedClient) send() {
	for this.connected {
		packet := network.NewMessagePacket(<-this.sendChannel)
		this.connection.Write(packet.Marshall())
	}
}
