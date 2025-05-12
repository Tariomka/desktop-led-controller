package runner

import (
	"log/slog"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/services"
	"github.com/Tariomka/desktop-led-controller/internal/tcp"
	"github.com/Tariomka/desktop-led-controller/internal/ui"
)

type IRunner interface {
	Start()
	Stop()
}

type LedClientRunner struct {
	window       *ui.Window
	tcpClient    *tcp.LedClient
	ledProcessor *services.LedProcService
	logger       *slog.Logger

	config RunnerConfig
}

func NewRunner(config RunnerConfig) IRunner {
	// TODO: Validate configuration
	logger := common.NewStructuredLogger(os.Stdout, slog.LevelDebug)

	return &LedClientRunner{
		window: ui.NewWindow(),
		tcpClient: tcp.NewClient(tcp.ClientConfig{
			Logger: logger,
			IP:     config.IP,
			Port:   config.Port,
		}),
		ledProcessor: services.NewLedProcService(services.LedProcConfig{
			Logger: logger,
		}),
		logger: logger,
		config: config,
	}
}

func (this *LedClientRunner) Start() {
	this.window.Start()
	this.window.Render()
}

func (this *LedClientRunner) Stop() {
	global.SendMessage(constants.TCPClient, models.TCPDisconnectMessage{})
	this.window.Stop()
}
