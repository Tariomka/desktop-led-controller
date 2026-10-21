package services

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Tariomka/desktop-led-controller/internal/common"
	"github.com/Tariomka/desktop-led-controller/internal/common/constants"
	"github.com/Tariomka/desktop-led-controller/internal/global"
	"github.com/Tariomka/desktop-led-controller/internal/models"
	"github.com/Tariomka/desktop-led-controller/internal/show"
)

const (
	projectDir       = "light_shows"
	projectExtension = ".ls"
)

var (
	ErrProjectName = errors.New("invalid project name")
	ErrFileType    = errors.New("unsupported file type (use .ls, .gc, .hex or .bin)")
)

// ShowService stores and reads project files. Loaded projects are handed to the UI, which
// owns them; saving receives an already encoded project.
type ShowService struct {
	fileService *FileService
	logger      *slog.Logger

	channel chan any
}

func NewShowService(logger *slog.Logger) *ShowService {
	service := &ShowService{
		fileService: NewFileService(logger),
		logger:      common.EnsureLoggerExists(logger),
		channel:     make(chan any, 4),
	}

	go service.channelLoop()
	global.RegisterMessageReceiver(
		constants.ServiceShow,
		func(message any) { service.channel <- message })

	return service
}

func (this *ShowService) Fetch() {
	files := this.fileService.FindFiles(filepath.Join(projectDir, "*"+projectExtension))
	var names []string
	for _, file := range files {
		names = append(names, strings.TrimSuffix(filepath.Base(file), projectExtension))
	}
	global.SendMessage(constants.UIShowPanel, models.ProjectListMessage{Names: names})
}

func (this *ShowService) Load(name string) {
	path, err := projectPath(name)
	if err == nil {
		this.Import(path)
		return
	}
	this.report(err, "Loading failed")
}

// Import reads a project, vendor project, firmware image or show data file.
func (this *ShowService) Import(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		this.report(err, "Reading failed")
		return
	}

	var project *show.Project
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case projectExtension:
		project, err = show.LoadProject(data)
	case ".gc":
		project, err = show.FromGC(data)
	case ".hex":
		if data, err = show.ShowDataFromHex(data); err == nil {
			project, err = show.FromShowData(data)
		}
	case ".bin":
		project, err = show.FromShowData(data)
	default:
		err = ErrFileType
	}
	if err != nil {
		this.report(err, "Import of "+filepath.Base(path)+" failed")
		return
	}

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	this.logger.Info("Project loaded", "file", path)
	global.SendMessage(constants.UIShowPanel, models.ProjectLoadedMessage{Name: name, Project: project})
	this.status(fmt.Sprintf("Loaded %s", filepath.Base(path)), false)
}

func (this *ShowService) Save(name string, data []byte) {
	path, err := projectPath(name)
	if err == nil {
		relative := filepath.Join(projectDir, name+projectExtension)
		err = this.fileService.SaveFile(relative, data)
		if err == nil {
			this.logger.Info("Project saved", "file", path)
			this.status("Saved "+name+projectExtension, false)
			this.Fetch()
			return
		}
	}
	this.report(err, "Saving failed")
}

func (this *ShowService) report(err error, text string) {
	this.logger.Error(text, "error", err)
	this.status(text+": "+err.Error(), true)
}

func (this *ShowService) status(text string, isError bool) {
	global.SendMessage(constants.UIShowPanel, models.StatusTextMessage{Text: text, Error: isError})
}

func projectPath(name string) (string, error) {
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `/\:*?"<>|`) || name == ".." {
		return "", ErrProjectName
	}
	return common.GetFullPathFromRelativePath(projectDir, name+projectExtension)
}

// Blocking message loop
func (this *ShowService) channelLoop() {
	for {
		switch message := (<-this.channel).(type) {
		case models.FetchMessage:
			this.Fetch()
		case models.LoadMessage:
			this.Load(message.Name)
		case models.ImportMessage:
			this.Import(message.Path)
		case models.SaveMessage:
			this.Save(message.Name, message.Data)
		default:
			this.logger.Warn("Unhandled message", "type", reflect.TypeOf(message))
		}
	}
}
