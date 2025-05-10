package runner

import (
	"log/slog"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/services"
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
	logger       *slog.Logger

	config RunnerConfig
}

func NewRunner(config RunnerConfig) IRunner {
	// TODO: Validate configuration
	logger := common.NewStructuredLogger(os.Stdout, slog.LevelDebug)

	return &LedClientRunner{
		Window: ui.NewWindow(),
		Client: tcp.NewClient(tcp.ClientConfig{
			Logger: logger,
			IP:     config.IP,
			Port:   config.Port,
		}),
		LayoutWorker: &led.LedLayout{},
		logger:       logger,
		config:       config,
	}
}

func (this *LedClientRunner) Start() {
	// TODO: THIS iS TEMP, CLEAN LATER
	fs := &services.FileService{}
	content := fs.ReadFileContent("aahhh")
	this.logger.Debug("Read some stuff from file", "content", string(content))

	this.Window.Start()
	this.Window.Render()
}

func (this *LedClientRunner) Stop() {
	global.SendMessage(constants.TCPClient, models.TCPDisconnectMessage{})
	this.Window.Stop()
}
