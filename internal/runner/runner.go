package runner

import (
	"github.com/Tariomka/desktop-led-controller/internal/processor"
	"github.com/Tariomka/desktop-led-controller/internal/tcp"
	"github.com/Tariomka/desktop-led-controller/internal/ui"
)

type IRunner interface {
	Start()
	Stop()
}

type LedClientRunner struct {
	Window       *ui.Window
	Client       *tcp.LedClient
	LayoutWorker processor.LayoutWorker

	config RunnerConfig
}

func NewRunner(config RunnerConfig) IRunner {
	return &LedClientRunner{
		Window:       ui.NewWindow(),
		Client:       tcp.NewClient(config.IP, config.Port),
		LayoutWorker: &processor.LedLayout{},
		config:       config,
	}
}

func (this *LedClientRunner) Start() {
	this.Window.Start()

	// TODO: create a channel to communicate between menu panel and TCP client
	// TODO: start a Goroutine for the client?
	// Channels?:

	this.Window.Render()
}

func (this *LedClientRunner) Stop() {
	defer this.Window.Stop()
}
