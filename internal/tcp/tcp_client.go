package tcp

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reflect"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/led-common-lib/pkg/network"
)

type TCPConfig struct {
	IP     string
	Port   uint16
	Logger *slog.Logger
}

type TCPClient struct {
	logger  *slog.Logger
	address string // Temporary until UI is updated to accept user input server address

	connected   bool
	connection  net.Conn
	sendChannel chan any

	channel chan any // Channel for receiving messages from other components
}

func NewClient(config TCPConfig) *TCPClient {
	// TODO: validate config
	config.Logger = common.EnsureLoggerExists(config.Logger)

	client := &TCPClient{
		logger:      config.Logger,
		address:     fmt.Sprintf("%s:%d", config.IP, config.Port),
		channel:     make(chan any, 1),
		sendChannel: make(chan any),
	}

	go client.channelLoop()
	global.RegisterMessageReceiver(
		constants.TCPClient,
		func(message any) { client.channel <- message })

	return client
}

func (this *TCPClient) Connect() {
	connection, err := net.Dial("tcp", this.address)
	if err != nil {
		this.logger.Error("Failed to dial remote address", "error", err)
		global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
		return
	}

	this.connection = connection
	this.connected = true
	this.logger.Info(
		"Successfully made a connection",
		"connected from (local)", this.connection.LocalAddr().String(),
		"connected to (remote)", this.connection.RemoteAddr().String(),
		"connection type", this.connection.RemoteAddr().Network())

	go this.receive()
	go this.send()

	global.SendMessage(constants.UIMenuPanel, models.ConnectedMessage{})
}

func (this *TCPClient) Disconnect() {
	if !this.connected {
		this.logger.Debug("No connection to disconnect from")
		return
	}

	this.connected = false
	if this.connection != nil {
		address := this.connection.RemoteAddr().String()

		this.connection.Close()
		this.connection = nil

		this.logger.Info("Disconnected and cleaned up", "remote address", address)
	}
	global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
}

func (this *TCPClient) receive() {
	for this.connected {
		buffer := make([]byte, 1024*4)
		n, err := this.connection.Read(buffer)
		// TODO: Handle errors and data
		// TODO: Handle and send to UI when server closes
		if err != nil {
			switch {
			case errors.Is(err, net.ErrClosed):
				this.logger.Info("Client disconnected from server")
			case errors.Is(err, io.EOF):
				this.logger.Info("Server has shutdown")
			default:
				this.logger.Error("Failed to read packet", "error", err.Error())
			}
			this.Disconnect()
			return
		}

		this.logger.Debug("Packet received", "content length", n, "content", buffer)
	}
}

func (this *TCPClient) send() {
	for this.connected { // Infinite loop while connected to a server
		var packets []network.Packet

		switch message := (<-this.sendChannel).(type) { // Blocking opperation while channel is emtpy
		case string:
			packets = append(packets, network.NewMessagePacket(message))
		case []byte:
			// TODO: either get global state of data type or add it to the message
			packets = append(packets, network.NewLedPacket(network.RGB8x8, message))
		case [][]byte:
			for _, packet := range message {
				packets = append(packets, network.NewLedPacket(network.RGB8x8, packet))
			}
		default:
			this.logger.Warn(
				"Received message is not supported for sending to server",
				"message type", reflect.TypeOf(message),
				"message content", message)
		}

		if len(packets) == 0 {
			this.logger.Debug("Message received but no packets were created")
		}

		for i, packet := range packets {
			n, err := this.connection.Write(packet.Marshall())

			this.logger.Debug(
				"Packet has been sent",
				"packet index in the batch", i,
				"byte count that should have been sent", len(packet.Marshall()),
				"byte count that actually was sent", n,
				"any error?", err)
		}
	}
}

// Blocking message loop
func (this *TCPClient) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.TCPConnectMessage:
			this.Connect()
		case models.TCPDisconnectMessage:
			this.Disconnect()
		case models.TCPSendPacketMessage:
			if !this.connected {
				this.logger.Warn("Packet cannot be sent because tcp connection has not been established")
				continue
			}

			this.sendChannel <- message.Data
		}
	}
}
