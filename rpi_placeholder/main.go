package main

import (
	"log"

	"github.com/Tariomka/desktop-led-controller/rpi_placeholder/internal/tcp"
)

func main() {
	server, err := tcp.NewServer(42069)
	if err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}

	server.Start()
	defer server.Stop()
}
