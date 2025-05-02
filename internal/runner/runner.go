package runner

import (
	"log/slog"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/tcp"
	"github.com/Tariomka/desktop-led-controller/internal/ui"
	"github.com/Tariomka/desktop-led-controller/internal/ui/global"
	"github.com/Tariomka/led-common-lib/pkg/led"
)

type IRunner interface {
	Start()
	Stop()
}

type LedClientRunner struct {
	Window       *ui.Window
	Client       *tcp.LedClient
	LayoutWorker led.LayoutWorker

	config RunnerConfig
}

func NewRunner(config RunnerConfig) IRunner {
	return &LedClientRunner{
		Window: ui.NewWindow(),
		Client: tcp.NewClient(tcp.ClientConfig{
			Logger: common.NewStructuredLogger(os.Stdout, slog.LevelDebug),
			IP:     config.IP,
			Port:   config.Port,
		}),
		LayoutWorker: &led.LedLayout{},
		config:       config,
	}
}

func (this *LedClientRunner) Start() {
	this.Window.Start()
	this.Window.Render()
}

func (this *LedClientRunner) Stop() {
	global.SendToClient(models.TCPDisconnectMessage{})
	this.Window.Stop()
}
