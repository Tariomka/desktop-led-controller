package main

import "github.com/Tariomka/desktop-led-controller/internal/runner"

func main() {
	runner := runner.NewRunner(runner.NewConfig())
	runner.Start()
	defer runner.Stop()
}

// import (
// 	"fmt"
// 	"time"

// 	"github.com/Tariomka/desktop-led-controller/internal/tcp"
// )

// func main() {
// 	println("test")
// 	go func() {
// 		for i := range 10 {
// 			client := tcp.NewClient()
// 			go client.Start(fmt.Appendf(nil, "test packet numero %d", 1+i))
// 			time.Sleep(500 * time.Millisecond)
// 		}
// 	}()

// 	time.Sleep(10 * time.Second)
// }
