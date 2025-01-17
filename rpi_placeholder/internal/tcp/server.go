package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Server interface {
	Start()
	Stop()
}

type LedServer struct {
	listener  net.Listener
	waitGroup *sync.WaitGroup
	conns     map[*Connection]bool
	mutex     sync.RWMutex
}

func NewServer(port uint16) (Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		return nil, err
	}

	return &LedServer{
		listener:  listener,
		waitGroup: &sync.WaitGroup{},
		conns:     make(map[*Connection]bool, 4),
		mutex:     sync.RWMutex{},
	}, nil
}

func (ls *LedServer) Start() {
	for {
		connection, err := ls.listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v\n", err)
			break
		}

		var connWrapper *Connection
		if len(ls.conns) > 3 {
			log.Printf("too many connections, can't add '%v'\n", connection)
			continue
		}
		ls.withLock(func() {
			connWrapper = NewConnection(connection, ls.waitGroup)
			ls.conns[connWrapper] = true
			log.Printf("new connection aquired: %v\n", connWrapper)
			log.Printf("current capacity: %d\n", len(ls.conns))
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
				log.Printf("user disconnected: %v\n", connection)
			default:
				log.Printf("failed to read data from connection: %v\n", err)
			}
			break
		}

		packet.Marshall()
		// log.Printf("received data: %v\n", packet)
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
	// incorrect indexes when multiple connections are added and removed
	log.Printf("removing connection: %v\n", conn)
	conn.Close()
	ls.withLock(func() { delete(ls.conns, conn) })
}

func (ls *LedServer) withLock(callback func()) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	callback()
}

type LedClient struct {
	address string
}

func NewClient() *LedClient {
	return &LedClient{
		address: ":42069",
	}
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
	for i := 0; i < 3; i++ {
		n, err = connection.Write(payload)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Printf("client error: %v\n", err)
	}

	log.Printf("client sent %d bytes, content: %s\n", n, payload)
}
