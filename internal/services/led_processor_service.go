package services

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/led-common-lib/pkg/led"
)

const (
	secureDirMode  = os.FileMode(0755)
	secureFileMode = os.FileMode(0644)

	lightShowDir       = "light_shows"
	lightShowExtension = ".ls"
)

type LedProcConfig struct {
	Logger *slog.Logger
}

type ProcessorService interface {
	AddToBuffer(layout *common.CubeLayout)
	Fetch()
	Save()
	Load()
}

type LedProcService struct {
	name         string
	layoutWorker led.LayoutWorker
	framesBuffer led.LightShow

	fileService *FileService
	logger      *slog.Logger
	// TODO: add messenger and channel to communicate from and to UI (for saving, loading, fetching, etc.)

	channel chan any
}

func NewLedProcService(config LedProcConfig) ProcessorService {
	service := &LedProcService{
		fileService:  NewFileService(config.Logger),
		logger:       common.EnsureLoggerExists(config.Logger),
		name:         "placeholder.ls",
		layoutWorker: &led.LedLayout{},
		channel:      make(chan any),
	}

	go service.channelLoop()
	global.RegisterMessageReceiver(
		constants.ServiceLedProcessor,
		func(message any) { service.channel <- message })

	return service
}

func (this *LedProcService) AddToBuffer(layout *common.CubeLayout) {
	// In the UI when "Next Frame" is clicked, the currect CubeGrid state will be saved and
	// CubeGrid cubes will be reset

}

func (this *LedProcService) Fetch() {
	// TODO: add logic to fetch all "*.ls" files from "light_shows/" directory
	// This will be used to query saved cube configurations and the output will be sent to UI
	// to select which configuration to Load in.
}

func (this *LedProcService) Save() {
	relativeFilePath := filepath.Join(lightShowDir, this.name)

	this.logger.Info("[LED_PROC_SERVICE] saving", "path", relativeFilePath)
}

func (this *LedProcService) Load() {

}

func (this *LedProcService) SetName(name string) {
	this.name = name + lightShowExtension
}

// Blocking state loop
func (this *LedProcService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		default:
			this.logger.Info("[LED_PROC_SERVICE] received", "message", message)
		}
	}
}
