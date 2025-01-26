package main

import "github.com/Tariomka/desktop-led-controller/internal/ui"

// import (
// 	"fmt"

// 	"github.com/Tariomka/desktop-led-controller/internal/tcp"
// )

// func main() {
// 	go func() {
// 		for i := 0; i < 10; i++ {
// 			client := tcp.NewClient()
// 			go client.Start([]byte(fmt.Sprintf("test packet numero %d", 1+i)))
// 			time.Sleep(500 * time.Millisecond)
// 		}
// 	}()

// 	time.Sleep(5 * time.Second)
// }

func main() {
	if window := ui.NewWindow(); window != nil {
		window.Start()
		defer window.Stop()
	}
}
