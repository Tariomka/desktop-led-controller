package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/data"
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

	headerSize  = 2
	byteSize8x8 = 192
	// byteSize8x8Mono = 64
	// byteSize16x16   = 768
	// byteSize8x32    = 768
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
	lightShow    led.LightShow
	framesBuffer []data.CubeFrame

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
	files := this.fileService.FindFiles(filepath.Join(lightShowDir, "*"+lightShowExtension))
	this.logger.Debug("[LED_PROC_SERVICE] files fetched", "files", files)

	var names []string
	for _, filePath := range files {
		filename := filepath.Base(filePath)
		name := strings.TrimSuffix(filename, lightShowExtension)
		names = append(names, name)
	}
	global.SendMessage(constants.UIEditPanel, models.SetLightShowsMessage{Names: names})
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

func (this *LedProcService) Load(name string) {
	relativeFilePath := filepath.Join(lightShowDir, name+lightShowExtension)

	this.logger.Info("[LED_PROC_SERVICE] loading file content", "file path", relativeFilePath)

	content := this.fileService.PeakFileContent(relativeFilePath, headerSize)
	if len(content) < headerSize {
		this.logger.Error("[LED_PROC_SERVICE] file is empty", "file path", relativeFilePath)
		return
	}

	if content[0] != byte(v1) || content[1] != byte(network.RGB8x8) {
		this.logger.Error("[LED_PROC_SERVICE] unsupported file", "file path", relativeFilePath)
		this.logger.Debug(
			"[LED_PROC_SERVICE] unsupported file",
			"header version", content[0],
			"header type", content[1])
		return
	}

	content = this.fileService.ReadFileContent(relativeFilePath)
	content = content[headerSize:]
	this.logger.Debug(
		"[LED_PROC_SERVICE] loaded file content",
		"file path", relativeFilePath,
		"light show name", name,
		"life content length", len(content),
		"frames", len(content)/byteSize8x8)

	this.logger.Warn("[LED_PROC_SERVICE] file content loading is unfinished")
	return
	// this.name = name
	this.logger.Info(
		"[LED_PROC_SERVICE] file content successfully loaded",
		"file path", relativeFilePath,
		"light show name", name)
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
	var payloads [][]byte
	payloads = append(payloads, []byte{byte(v1), byte(network.RGB8x8)})

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

func (this *LedProcService) loadIntoLayout(frame data.CubeFrame) {
	setLayout := led.Frame(func(lw led.LayoutWorker) {
		for index, cube := range frame.IterateWithIndex() {
			lw.SetSingle(index.X, index.Y, index.Z, led.RGBAToColor(cube.Color))
		}
	})

	setLayout(this.layoutWorker)
}

func (this *LedProcService) loadIntoLightShow() {

}

func (this *LedProcService) loadIntoFrameBuffer() {

}

// Blocking message loop
func (this *LedProcService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.RenameMessage:
			this.name = message.Name
		case models.AddToBufferMessage:
			// TODO: handle addition when currently selected not last frame
			this.framesBuffer = append(this.framesBuffer, message.Frame)
			global.TotalFrameCount = uint32(len(this.framesBuffer))
			global.SelectedFrame = global.TotalFrameCount
		case models.LoadMessage:
			this.Load(message.Name)
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
