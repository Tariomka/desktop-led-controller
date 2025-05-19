package services

import (
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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

	headerSize = 3
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
	this.logger.Debug("files fetched", "files", files)

	var names []string
	for _, filePath := range files {
		filename := filepath.Base(filePath)
		name := strings.TrimSuffix(filename, lightShowExtension)
		names = append(names, name)
	}
	global.SendMessage(constants.UIEditPanel, models.SetLightShowsMessage{Names: names})
}

func (this *LedProcService) Save() {
	this.logger.Info("Saving light show data")

	relativeFilePath := filepath.Join(lightShowDir, this.name+lightShowExtension)
	payload := append(this.getHeaderPayload(), this.getLightShowPayload()...)
	if err := this.fileService.SaveFile(relativeFilePath, payload...); err != nil {
		this.logger.Error(
			"Failed to save light show data",
			"file path", relativeFilePath,
			"error", err)
		return
	}

	this.logger.Info("Saving successfully finished", "file path", relativeFilePath)
}

func (this *LedProcService) Load(name string) {
	relativeFilePath := filepath.Join(lightShowDir, name+lightShowExtension)
	this.logger.Info("Loading file content", "file path", relativeFilePath)
	if !this.verifyFileHeader(relativeFilePath) {
		return
	}

	content := this.fileService.ReadFileContent(relativeFilePath)
	content = content[headerSize:]

	this.loadIntoLightShow(content)
	this.loadIntoFrameBuffer(this.lightShow)

	global.SelectedFrame = 0
	global.TotalFrameCount = uint32(len(this.framesBuffer))

	if len(this.framesBuffer) < 1 {
		global.SendMessage(constants.UICubeGrid, models.ResetMessage{})
		this.logger.Error(
			"Unexpected reset due to failed frame buffer loading - file might be emtpy or malformed",
			"file path", relativeFilePath)
		return
	}

	global.SendMessage(constants.UICubeGrid, models.SetFrameMessage{Frame: this.framesBuffer[0]})
	this.name = name
	this.logger.Info(
		"File content successfully loaded",
		"file path", relativeFilePath,
		"light show name", name)
}

func (this *LedProcService) AddToBuffer(cubeFrame data.CubeFrame, index uint32) {
	frame := led.Frame(func(lw led.LayoutWorker) {
		for index, cube := range cubeFrame.IterateWithIndex() {
			if err := lw.SetSingle(index.X, index.Y, index.Z, led.RGBAToColor(cube.Color)); err != nil {
				this.logger.Warn(
					"Unexpected error occured while loading frame - state might be malformed or not set",
					"index", index,
					"color", cube.Color,
					"error", err)
			}
		}
	})

	if index < uint32(len(this.framesBuffer)) && len(this.framesBuffer) > 0 {
		this.framesBuffer = slices.Insert(this.framesBuffer, int(index), cubeFrame)
		this.lightShow = slices.Insert(this.lightShow, int(index), frame)
	} else {
		this.framesBuffer = append(this.framesBuffer, cubeFrame)
		this.lightShow = append(this.lightShow, frame)
	}

	global.TotalFrameCount = uint32(len(this.framesBuffer))
	this.RenderFrame(index)
}

func (this *LedProcService) RemoveFromBuffer(index uint32) {
	if len(this.framesBuffer) < 1 {
		this.logger.Debug("Buffer is already empty")
		return
	}

	if index >= uint32(len(this.framesBuffer)) {
		this.logger.Warn(
			"Index out of range of buffer",
			"buffer length", len(this.framesBuffer),
			"index", index)
		return
	}

	this.framesBuffer = slices.Delete(this.framesBuffer, int(index), int(index+1))
	this.lightShow = slices.Delete(this.lightShow, int(index), int(index+1))

	global.TotalFrameCount = uint32(len(this.framesBuffer))
	this.RenderFrame(index)
}

func (this *LedProcService) RenderFrame(index uint32) {
	bufferSize := uint32(len(this.framesBuffer))
	if index >= bufferSize {
		global.SendMessage(constants.UICubeGrid, models.ResetMessage{})
		global.SelectedFrame = bufferSize
		return
	}

	global.SendMessage(constants.UICubeGrid, models.SetFrameMessage{Frame: this.framesBuffer[index]})
	global.SelectedFrame = index
}

func (this *LedProcService) SendCurrentBuffer() {
	global.SendMessage(
		constants.TCPClient,
		models.TCPSendPacketMessage{Data: this.getLightShowPayload()})
}

func (this *LedProcService) getHeaderPayload() [][]byte {
	return [][]byte{{byte(network.V1), byte(network.RGB8x8)}}
}

func (this *LedProcService) getLightShowPayload() [][]byte {
	var payload [][]byte

	this.layoutWorker.ResetBlock()
	for _, setLayout := range this.lightShow {
		setLayout(this.layoutWorker)
		var payloadFragment []byte

		for _, content := range this.layoutWorker.IterateSlices() {
			payloadFragment = append(payloadFragment, content...)
		}
		payload = append(payload, payloadFragment)

		this.layoutWorker.ResetBlock()
	}

	return payload
}

func (this *LedProcService) loadIntoLightShow(content []byte) {
	this.lightShow = nil

	// Only parsing full frame content, so if there is additional content, the file might be incorrect
	// Addition +1 is to accommodate '\n' (new line) character as well
	for index := 0; index+network.ByteSize8x8 < len(content); index += network.ByteSize8x8 + 1 {
		frameContent := content[index : index+network.ByteSize8x8] // 192 byte slice - equivalent to 8x24
		layout := led.LedLayout{}

		for i := range 8 {
			layout[i] = [24]byte(frameContent[i*24:])
		}

		this.lightShow = append(this.lightShow, func(lw led.LayoutWorker) {
			if err := lw.Overwrite(layout.IterateSlices()); err != nil {
				this.logger.Error("Incompatible frame was loaded", "error", err)
			}
		})
	}
}

func (this *LedProcService) loadIntoFrameBuffer(lightShow led.LightShow) {
	this.framesBuffer = nil

	this.layoutWorker.ResetBlock()
	for _, setLayout := range lightShow {
		setLayout(this.layoutWorker)
		// TODO: synchronize frame creation to be the same as in 'Window', as this will be volatile
		frame := data.NewCubeFrameWithDefaultSize(8, 8, 8)

		for index, color := range this.layoutWorker.IterateColors() {
			frame[index.Z][index.Y][index.X].Color = common.ColorToRGBA(color)
		}

		this.framesBuffer = append(this.framesBuffer, frame)
	}
}

func (this *LedProcService) verifyFileHeader(filePath string) bool {
	headerContent := this.fileService.PeakFileContent(filePath, headerSize)
	if len(headerContent) < headerSize {
		this.logger.Error("File is empty", "file path", filePath)
		return false
	}

	if headerContent[0] != byte(network.V1) ||
		headerContent[1] != byte(network.RGB8x8) ||
		headerContent[2] != '\n' {
		this.logger.Error("Unsupported file", "file path", filePath)
		this.logger.Debug(
			"Unsupported file",
			"header version", headerContent[0],
			"header type", headerContent[1])
		return false
	}

	return true
}

// Blocking message loop
func (this *LedProcService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.SendMessage:
			this.SendCurrentBuffer()
		case models.RenameMessage:
			this.name = message.Name
		case models.AddFrameMessage:
			this.AddToBuffer(message.Frame, message.Index)
		case models.RemoveFrameMessage:
			this.RemoveFromBuffer(message.Index)
		case models.RenderFrameMessage:
			this.RenderFrame(message.Index)
		case models.LoadMessage:
			this.Load(message.Name)
		case models.SaveMessage:
			this.Save()
		case models.FetchMessage:
			this.Fetch()
		case models.ResetMessage:
			this.framesBuffer = nil
			this.lightShow = nil
			global.TotalFrameCount = uint32(len(this.framesBuffer))
			global.SelectedFrame = global.TotalFrameCount
		default:
			this.logger.Warn(
				"Unproccessed message - either there was a mistake or need to handle this",
				"message", message,
				"type", reflect.TypeOf(message))
		}
	}
}
