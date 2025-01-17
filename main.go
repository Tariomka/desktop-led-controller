package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

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

func main() {
	go func() {
		for i := 0; i < 10; i++ {
			client := NewClient()
			go client.Start([]byte(fmt.Sprintf("test packet numero %d", 1+i)))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
}
