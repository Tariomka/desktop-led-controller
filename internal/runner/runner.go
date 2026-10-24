package runner

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/client"
	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/config"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/services"
	"github.com/Tariomka/desktop-led-controller/internal/ui"
)

type IRunner interface {
	Start()
	Stop()
}

type LedClientRunner struct {
	window      *ui.Window
	client      *client.Client
	showService *services.ShowService
	logger      *slog.Logger
}

func NewRunner() IRunner {
	logger := common.NewStructuredLogger(os.Stdout, slog.LevelDebug)
	settings, err := config.Load()
	if err != nil {
		fmt.Printf("Using default settings: %s\n", err.Error())
	}
	global.Settings = settings

	return &LedClientRunner{
		window: ui.NewWindow(),
		client: client.NewClient(client.Config{
			Logger: logger,
			IP:     settings.IP,
			Port:   settings.Port,
		}),
		showService: services.NewShowService(logger),
		logger:      logger,
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
