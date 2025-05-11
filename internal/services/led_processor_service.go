package services

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/led-common-lib/pkg/led"
	"github.com/Tariomka/led-common-lib/pkg/network"
)

const (
	secureDirMode  = os.FileMode(0755)
	secureFileMode = os.FileMode(0644)

	lightShowDir       = "light_shows"
	lightShowExtension = ".ls"

	headerSize = 3
)

type version byte

const (
	v1 version = iota + 1
)

type LedProcConfig struct {
	Logger *slog.Logger
}

type LedProcService struct {
	name         string
	layoutWorker led.LayoutWorker
	framesBuffer led.LightShow

	fileService *FileService
	logger      *slog.Logger

	channel chan any
}

func NewLedProcService(config LedProcConfig) *LedProcService {
	service := &LedProcService{
		fileService:  NewFileService(config.Logger),
		logger:       common.EnsureLoggerExists(config.Logger),
		name:         "placeholder",
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
	this.framesBuffer = append(this.framesBuffer, func(lw led.LayoutWorker) {
		for zIndex, z := range *layout {
			for yIndex, y := range z {
				for xIndex, cube := range y {
					lw.SetSingle(uint8(xIndex), uint8(yIndex), uint8(zIndex), led.RGBAToColor(cube.Color))
				}
			}
		}
	})
}

func (this *LedProcService) Fetch() {
	// TODO: add logic to fetch all "*.ls" files from "light_shows/" directory
	// This will be used to query saved cube configurations and the output will be sent to UI
	// to select which configuration to Load in.
}

func (this *LedProcService) Save() {
	var payload []byte
	payload = append(payload, byte(v1), byte(network.RGB8x8), byte(len(this.name)))
	payload = append(payload, []byte(this.name)...)
	payload = append(payload, '\n')

	for _, setFrame := range this.framesBuffer {
		setFrame(this.layoutWorker)
		for _, content := range this.layoutWorker.IterateSlices() {
			payload = append(payload, content...)
		}
		payload = append(payload, '\n')

		this.layoutWorker.ResetBlock()
	}

	relativeFilePath := filepath.Join(lightShowDir, this.name+lightShowExtension)
	this.logger.Debug(
		"[LED_PROC_SERVICE] saving",
		"path", relativeFilePath,
		"payload", string(payload))

	this.fileService.AppendToFile(relativeFilePath, payload)
}

func (this *LedProcService) Load() {

}

func (this *LedProcService) SetName(name string) {
	this.name = name
}

// Blocking message loop
func (this *LedProcService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.RenameMessage:
			this.SetName(message.Name)
		case models.AddToBufferMessage:
			this.AddToBuffer(message.Layout)
		case models.LoadMessage:
			this.Load()
		case models.SaveMessage:
			this.Save()
		case models.FetchMessage:
			this.Fetch()
		default:
			this.logger.Info("[LED_PROC_SERVICE] unproccessed message", "message", message)
		}
	}
}
