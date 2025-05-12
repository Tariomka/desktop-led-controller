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
	framesBuffer []common.CubeFrame

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

func (this *LedProcService) Fetch() {
	// TODO: add logic to fetch all "*.ls" files from "light_shows/" directory
	// This will be used to query saved cube configurations and the output will be sent to UI
	// to select which configuration to Load in.
	panic(common.ErrNotImplemented)
}

func (this *LedProcService) Save() {
	this.logger.Debug("[LED_PROC_SERVICE] saving light show data")

	relativeFilePath := filepath.Join(lightShowDir, this.name+lightShowExtension)
	if err := this.fileService.SaveFile(
		relativeFilePath,
		this.getLightShowPayload()...); err != nil {
		this.logger.Error(
			"[LED_PROC_SERVICE] failed to save light show data",
			"file path", relativeFilePath,
			"error", err)
		return
	}

	this.logger.Debug(
		"[LED_PROC_SERVICE] saving successfully finished",
		"file path", relativeFilePath)
}

func (this *LedProcService) Load() {
	panic(common.ErrNotImplemented)
}

func (this *LedProcService) LoadFrame(index uint32) {
	bufferSize := uint32(len(this.framesBuffer))
	if index >= bufferSize {
		global.SendMessage(constants.UICubeGrid, models.ResetMessage{})
		global.SelectedFrame = bufferSize
		return
	}

	global.SendMessage(constants.UICubeGrid, models.SetFrameMessage{Frame: this.framesBuffer[index]})
	global.SelectedFrame = index
}

func (this *LedProcService) getLightShowPayload() [][]byte {
	header := []byte{byte(v1), byte(network.RGB8x8), byte(len(this.name))}
	header = append(header, []byte(this.name)...)

	var payloads [][]byte
	payloads = append(payloads, header)

	for _, frame := range this.framesBuffer {
		this.loadIntoLayout(frame)
		var payload []byte
		for _, content := range this.layoutWorker.IterateSlices() {
			payload = append(payload, content...)
		}
		payloads = append(payloads, payload)

		this.layoutWorker.ResetBlock()
	}

	return payloads
}

func (this *LedProcService) loadIntoLayout(frame common.CubeFrame) {
	setLayout := led.Frame(func(lw led.LayoutWorker) {
		for zIndex, z := range frame {
			for yIndex, y := range z {
				for xIndex, cube := range y {
					lw.SetSingle(
						uint8(xIndex),
						uint8(yIndex),
						uint8(zIndex),
						led.RGBAToColor(cube.Color))
				}
			}
		}
	})

	setLayout(this.layoutWorker)
}

// Blocking message loop
func (this *LedProcService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.RenameMessage:
			this.name = message.Name
		case models.AddToBufferMessage:
			this.framesBuffer = append(this.framesBuffer, common.DeepCloneLayout(message.Frame))
			global.TotalFrameCount = uint32(len(this.framesBuffer))
			global.SelectedFrame = global.TotalFrameCount
		case models.LoadMessage:
			this.Load()
		case models.SaveMessage:
			this.Save()
		case models.FetchMessage:
			this.Fetch()
		case models.LoadFrameMessage:
			this.LoadFrame(message.Index)
		case models.ResetMessage:
			this.framesBuffer = nil
			global.TotalFrameCount = uint32(len(this.framesBuffer))
			global.SelectedFrame = global.TotalFrameCount
		default:
			this.logger.Info("[LED_PROC_SERVICE] unproccessed message", "message", message)
		}
	}
}
