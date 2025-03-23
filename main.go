package main

import "github.com/Tariomka/desktop-led-controller/internal/ui"

func main() {
	if window := ui.NewWindow(); window != nil {
		window.Start()
		defer window.Stop()
	}
}

// import (
// 	"fmt"
// 	"time"

// 	"github.com/Tariomka/desktop-led-controller/internal/tcp"
// )

// func main() {
// 	go func() {
// 		for i := range 10 {
// 			client := tcp.NewClient()
// 			go client.Start(fmt.Appendf(nil, "test packet numero %d", 1+i))
// 			time.Sleep(500 * time.Millisecond)
// 		}
// 	}()

// 	time.Sleep(5 * time.Second)
// }
