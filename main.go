package main

import "github.com/Tariomka/desktop-led-controller/internal/runner"

func main() {
	runner := runner.NewRunner()
	runner.Start()
	defer runner.Stop()
}
