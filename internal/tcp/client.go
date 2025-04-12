package tcp

import (
	"fmt"
	"log"
	"net"
	"time"
)

type LedClient struct{ address string }

func NewClient(ip string, port uint16) *LedClient {
	return &LedClient{address: fmt.Sprintf("%s:%d", ip, port)}
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
