package tcp

import (
	"log"
	"net"
	"time"
)

type LedClient struct {
	address string
}

func NewClient() *LedClient {
	return &LedClient{
		address: "192.168.0.169:42069",
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
	for range 3 {
		n, err = connection.Write(payload)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Printf("client error: %v\n", err)
	}

	log.Printf("client sent %d bytes, content: %s\n", n, payload)
}
