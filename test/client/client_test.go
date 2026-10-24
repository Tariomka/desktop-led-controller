package client_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/Tariomka/desktop-led-controller/internal/client"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/led-common-lib/pkg/communication/tcp"
)

type fakePico struct {
	t          *testing.T
	connection net.Conn
	buffer     []byte
}

func (this *fakePico) read(want tcp.Type) tcp.Message {
	this.t.Helper()
	this.connection.SetReadDeadline(time.Now().Add(5 * time.Second))
	message, err := tcp.Read(this.connection, this.buffer)
	if err != nil {
		this.t.Fatal(err)
	}
	if message.Type != want {
		this.t.Fatalf("got message type %#x, want %#x", message.Type, want)
	}
	return message
}

func (this *fakePico) write(message tcp.Message) {
	this.t.Helper()
	if err := tcp.Write(this.connection, message); err != nil {
		this.t.Fatal(err)
	}
}

func receive[T any](t *testing.T, channel chan any) T {
	t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case message := <-channel:
			if typed, ok := message.(T); ok {
				return typed
			}
		case <-timeout:
			var zero T
			t.Fatalf("no %T received", zero)
			return zero
		}
	}
}

func TestClientSession(t *testing.T) {
	app := make(chan any, 256)
	global.RegisterMessageReceiver(constants.UIApp, func(message any) { app <- message })
	menu, console, showPanel := app, app, app

	desktop, picoSide := net.Pipe()
	defer picoSide.Close()
	pico := &fakePico{t: t, connection: picoSide, buffer: make([]byte, tcp.HeaderSize+tcp.MaxChunk)}
	go client.NewClient(client.Config{IP: "127.0.0.1", Port: 1}).Run(desktop)

	receive[models.ConnectedMessage](t, menu)
	if hello := pico.read(tcp.TypeHello); hello.Text() == "" {
		t.Fatal("empty hello")
	}

	pico.write(tcp.NewHello("rpi-led-cube"))
	for {
		if line := receive[models.ConsoleLineMessage](t, console); line.Source == models.SourceApp {
			break
		}
	}

	pico.write(tcp.NewCubeLine("Status:Mode=Text,Cartoon=1,Count=9"))
	for {
		if line := receive[models.ConsoleLineMessage](t, menu); line.Source == models.SourceCube {
			if line.Text != "Status:Mode=Text,Cartoon=1,Count=9" {
				t.Fatalf("cube line: got %+v", line)
			}
			break
		}
	}
	pico.write(tcp.NewPicoLine("INFO Desktop connected"))
	for {
		if line := receive[models.ConsoleLineMessage](t, console); line.Source == models.SourcePico {
			break
		}
	}

	pico.write(tcp.Message{Version: tcp.Version + 1, Type: tcp.TypeCubeLine, Payload: []byte("future")})
	pico.write(tcp.NewPing())
	pico.read(tcp.TypePong)

	global.SendMessage(constants.TCPClient, models.TCPCommandMessage{Line: "Key 2"})
	if command := pico.read(tcp.TypeCommand); command.Text() != "Key 2" {
		t.Fatalf("command: got %q", command.Text())
	}

	data := bytes.Repeat([]byte{1, 2, 3, 4, 5}, 1000)
	global.SendMessage(constants.TCPClient, models.TCPUploadMessage{Data: data})
	size, err := pico.read(tcp.TypeUploadBegin).Size()
	if err != nil || int(size) != len(data) {
		t.Fatalf("upload begin: size %d, %v", size, err)
	}
	var stored []byte
	for len(stored) < len(data) {
		offset, chunk, err := pico.read(tcp.TypeUploadChunk).Chunk()
		if err != nil || int(offset) != len(stored) || len(chunk) > tcp.MaxChunk {
			t.Fatalf("chunk at %d: offset %d, %d bytes, %v", len(stored), offset, len(chunk), err)
		}
		stored = append(stored, chunk...)
		pico.write(tcp.NewUploadProgress(uint32(len(stored)), size))
		if progress := receive[models.UploadProgressMessage](t, showPanel); progress.Acked != len(stored) {
			t.Fatalf("progress: got %+v", progress)
		}
	}
	pico.read(tcp.TypeUploadEnd)
	pico.write(tcp.NewUploadResult(true, "Show upload OK"))
	if result := receive[models.UploadResultMessage](t, showPanel); !result.OK {
		t.Fatalf("result: got %+v", result)
	}
	if !bytes.Equal(stored, data) {
		t.Fatal("uploaded data differs")
	}

	picoSide.Close()
	receive[models.DisconnectedMessage](t, menu)
}
