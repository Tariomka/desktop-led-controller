package tcp

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

type ServerConfig struct {
	Address        string
	MaxConnections uint8
}

func NewConfig() ServerConfig {
	return ServerConfig{
		Address:        ":42069",
		MaxConnections: 4,
	}
}

type Server interface {
	Start()
	Stop()
	Send(message string)
}

type LedServer struct {
	listener net.Listener
	maxConns uint8

	waitGroup *sync.WaitGroup
	mutex     sync.RWMutex
	conns     map[*Connection]bool

	logger *slog.Logger
}

func NewServer(config ServerConfig) (Server, error) {
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, err
	}

	return &LedServer{
		listener:  listener,
		maxConns:  config.MaxConnections,
		waitGroup: &sync.WaitGroup{},
		mutex:     sync.RWMutex{},
		conns:     make(map[*Connection]bool, config.MaxConnections),
		logger: slog.New(NewLogHandler(
			func(message string) { fmt.Println(message) },
			&slog.HandlerOptions{Level: slog.LevelDebug})),
	}, nil
}

func (ls *LedServer) Start() {
	ls.logger.Debug("Starting up server")
	fmt.Println()
	for {
		connection, err := ls.listener.Accept()
		if err != nil {
			ls.logger.Error("failed to accept connection:", "error", err)
			break
		}

		if len(ls.conns) >= int(ls.maxConns) {
			ls.logger.Warn("can't accept any new connections, because of limited capacity", "connection", connection.LocalAddr())
			connection.Close()
			continue
		}

		var connWrapper *Connection
		ls.withLock(func() {
			connWrapper = NewConnection(connection, ls.waitGroup)
			ls.conns[connWrapper] = true
			ls.logger.Debug("new connection aquired:",
				"connection", connWrapper.connection.RemoteAddr(),
				"current capacity", len(ls.conns))
		})

		go ls.receive(connWrapper)
	}
	ls.waitGroup.Wait()
}

func (ls *LedServer) Stop() {
	for connection := range ls.conns {
		connection.Close()
	}
	ls.listener.Close()
}

func (ls *LedServer) Send(message string) {
	ls.broadcast(NewMessagePacket(message))
}

func (ls *LedServer) receive(connection *Connection) {
	defer ls.removeConnection(connection)
	for {
		packet, err := connection.ReadPacket()
		if err != nil {
			switch err {
			case io.EOF:
				ls.logger.Info("user disconnected:", "connection", connection.connection.RemoteAddr())
			default:
				ls.logger.Error("failed to read data from connection:", "error", err)
			}
			break
		}

		ls.logger.Debug("received data:",
			"version", packet.Version,
			"type", packet.Type,
			"data", string(packet.Data))
	}
}

func (ls *LedServer) broadcast(packet Packet) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()
	for connection := range ls.conns {
		connection.WritePacket(packet)
	}
}

func (ls *LedServer) removeConnection(conn *Connection) {
	ls.logger.Info("removing connection:", "connection", conn.connection.LocalAddr())
	conn.Close()
	ls.withLock(func() { delete(ls.conns, conn) })
}

func (ls *LedServer) withLock(callback func()) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	callback()
}
