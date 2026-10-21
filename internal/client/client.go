// Package client is the desktop side of the link to the Pico W bridge (protocol
// github.com/Tariomka/led-common-lib/pkg/communication/tcp).
package client

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	libcommon "github.com/Tariomka/led-common-lib/pkg/common"
	"github.com/Tariomka/led-common-lib/pkg/communication/tcp"
)

const (
	helloName   = "desktop-led-controller"
	dialTimeout = 5 * time.Second
	// UploadTimeout is the longest wait for an upload reply: a chunk is up to 16 cube records
	// with a 1 s timeout and one retry each.
	UploadTimeout = 40 * time.Second
)

var (
	ErrNotConnected  = errors.New("not connected")
	ErrUploadBusy    = errors.New("an upload is already running")
	ErrUploadTimeout = errors.New("no reply from the Pico")
)

type Config struct {
	IP     string
	Port   uint16
	Logger *slog.Logger
}

type Client struct {
	logger  *slog.Logger
	address string

	mutex      sync.Mutex
	connection net.Conn
	uploading  bool
	replies    chan tcp.Message // upload progress and results while uploading

	channel chan any
}

func NewClient(config Config) *Client {
	client := &Client{
		logger:  common.EnsureLoggerExists(config.Logger),
		address: fmt.Sprintf("%s:%d", config.IP, config.Port),
		channel: make(chan any, 16),
	}

	go client.channelLoop()
	global.RegisterMessageReceiver(
		constants.TCPClient,
		func(message any) { client.channel <- message })

	return client
}

// Run serves an established connection (e.g. a net.Pipe in tests) until it closes.
func (this *Client) Run(connection net.Conn) {
	this.mutex.Lock()
	this.connection = connection
	this.mutex.Unlock()

	global.SendMessage(constants.UIMenuPanel, models.ConnectedMessage{})
	if err := this.write(tcp.NewHello(helloName)); err != nil {
		this.logger.Error("Failed to send hello", "error", err)
	}
	this.receive(connection)
}

func (this *Client) Connect() {
	if this.connected() {
		return
	}

	connection, err := net.DialTimeout("tcp", this.address, dialTimeout)
	if err != nil {
		this.logger.Error("Failed to dial remote address", "address", this.address, "error", err)
		this.console(models.SourceApp, "Connection to "+this.address+" failed: "+err.Error())
		global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
		return
	}

	this.logger.Info("Connected", "remote address", connection.RemoteAddr().String())
	go this.Run(connection)
}

func (this *Client) Disconnect() {
	this.mutex.Lock()
	connection := this.connection
	this.mutex.Unlock()
	if connection != nil {
		connection.Close()
	}
}

func (this *Client) connected() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.connection != nil
}

func (this *Client) write(message tcp.Message) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.connection == nil {
		return ErrNotConnected
	}
	return tcp.Write(this.connection, message)
}

func (this *Client) receive(connection net.Conn) {
	buffer := make([]byte, tcp.HeaderSize+tcp.MaxChunk)
	for {
		message, err := tcp.Read(connection, buffer)
		switch {
		case errors.Is(err, libcommon.ErrUnsupportedVersion):
			this.logger.Warn("Message of another protocol version ignored", "version", message.Version, "type", message.Type)
			continue
		case errors.Is(err, libcommon.ErrPayloadTooLarge):
			this.logger.Warn("Oversized message skipped", "type", message.Type)
			continue
		case err != nil:
			this.closed(connection, err)
			return
		}
		this.handle(message)
	}
}

func (this *Client) closed(connection net.Conn, err error) {
	connection.Close()
	this.mutex.Lock()
	if this.connection == connection {
		this.connection = nil
	}
	if this.uploading {
		select {
		case this.replies <- tcp.NewUploadResult(false, "connection lost"):
		default:
		}
	}
	this.mutex.Unlock()

	switch {
	case errors.Is(err, net.ErrClosed):
		this.logger.Info("Disconnected")
	case errors.Is(err, io.EOF):
		this.logger.Info("The Pico closed the connection")
		this.console(models.SourceApp, "The Pico closed the connection")
	default:
		this.logger.Error("Connection lost", "error", err)
		this.console(models.SourceApp, "Connection lost: "+err.Error())
	}
	global.SendMessage(constants.UIMenuPanel, models.DisconnectedMessage{})
}

func (this *Client) handle(message tcp.Message) {
	switch message.Type {
	case tcp.TypeHello:
		this.console(models.SourceApp, "Connected to "+message.Text())
	case tcp.TypeCubeLine:
		line := models.ConsoleLineMessage{Source: models.SourceCube, Text: message.Text()}
		global.SendMessage(constants.UIConsolePanel, line)
		global.SendMessage(constants.UIMenuPanel, line)
	case tcp.TypePicoLine:
		this.console(models.SourcePico, message.Text())
	case tcp.TypeUploadProgress, tcp.TypeUploadResult:
		reply := tcp.Message{Type: message.Type, Payload: append([]byte(nil), message.Payload...)}
		this.mutex.Lock()
		if this.uploading {
			select {
			case this.replies <- reply:
			default:
			}
		}
		this.mutex.Unlock()
	case tcp.TypePing:
		if err := this.write(tcp.NewPong()); err != nil {
			this.logger.Error("Failed to answer ping", "error", err)
		}
	case tcp.TypePong:
	default:
		this.logger.Warn("Unexpected message", "type", message.Type)
	}
}

func (this *Client) console(source models.ConsoleSource, text string) {
	global.SendMessage(constants.UIConsolePanel, models.ConsoleLineMessage{Source: source, Text: text})
}

func (this *Client) command(line string) {
	if err := this.write(tcp.NewCommand(line)); err != nil {
		this.logger.Error("Failed to send command", "command", line, "error", err)
		this.console(models.SourceApp, "Command not sent: "+err.Error())
		return
	}
	this.console(models.SourceApp, "> "+line)
}

// Upload sends show data in chunks, each after the Pico reported the previous one stored, and
// reports progress and the result to the show panel.
func (this *Client) Upload(data []byte) {
	err := this.upload(data)
	if err != nil {
		this.logger.Error("Upload failed", "error", err)
		global.SendMessage(constants.UIShowPanel, models.UploadResultMessage{Text: err.Error()})
	}
}

func (this *Client) upload(data []byte) error {
	this.mutex.Lock()
	if this.uploading {
		this.mutex.Unlock()
		return ErrUploadBusy
	}
	this.uploading, this.replies = true, make(chan tcp.Message, 4)
	replies := this.replies
	this.mutex.Unlock()
	defer func() {
		this.mutex.Lock()
		this.uploading = false
		this.mutex.Unlock()
	}()

	if err := this.write(tcp.NewUploadBegin(uint32(len(data)))); err != nil {
		return err
	}
	for offset := 0; offset < len(data); {
		end := min(offset+tcp.MaxChunk, len(data))
		if err := this.write(tcp.NewUploadChunk(uint32(offset), data[offset:end])); err != nil {
			return err
		}
		for {
			reply, err := this.reply(replies)
			if err != nil {
				return err
			}
			if reply.Type == tcp.TypeUploadResult {
				this.result(reply)
				return nil
			}
			acked, total, err := reply.Progress()
			if err != nil {
				return err
			}
			global.SendMessage(constants.UIShowPanel, models.UploadProgressMessage{Acked: int(acked), Total: int(total)})
			if int(acked) >= end {
				break
			}
		}
		offset = end
	}

	if err := this.write(tcp.NewUploadEnd()); err != nil {
		return err
	}
	for {
		reply, err := this.reply(replies)
		if err != nil {
			return err
		}
		if reply.Type == tcp.TypeUploadResult {
			this.result(reply)
			return nil
		}
	}
}

func (this *Client) reply(replies chan tcp.Message) (tcp.Message, error) {
	select {
	case reply := <-replies:
		return reply, nil
	case <-time.After(UploadTimeout):
		if !this.connected() {
			return tcp.Message{}, ErrNotConnected
		}
		return tcp.Message{}, ErrUploadTimeout
	}
}

func (this *Client) result(reply tcp.Message) {
	ok, text, err := reply.Result()
	if err != nil {
		text = err.Error()
	}
	this.logger.Info("Upload finished", "ok", ok, "message", text)
	global.SendMessage(constants.UIShowPanel, models.UploadResultMessage{OK: ok, Text: text})
}

// Blocking message loop
func (this *Client) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.TCPConnectMessage:
			this.Connect()
		case models.TCPDisconnectMessage:
			this.Disconnect()
		case models.TCPCommandMessage:
			this.command(message.Line)
		case models.TCPUploadMessage:
			go this.Upload(message.Data)
		}
	}
}
