// TODO: Seperate and rename after implementing
// Here lies logic for storing, extracting, seting and loading LedLayout data to/from UI CubeGrid
// As well as sending LedLayout data via TCP using LedClient
package services

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Tariomka/desktop-led-controller/internal/common"
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

type LedProcService struct {
	name         string
	layoutWorker led.LayoutWorker
	framesBuffer led.LightShow

	fileService *FileService
	logger      *slog.Logger
	// TODO: add messenger and channel to communicate from and to UI (for saving, loading, fetching, etc.)
}

func NewLedProcService(config LedProcConfig) *LedProcService {
	if config.Logger == nil {
		config.Logger = common.NewConsoleLogger(slog.LevelDebug)
	}

	return &LedProcService{
		fileService:  NewFileService(config.Logger),
		logger:       config.Logger,
		layoutWorker: &led.LedLayout{},
	}
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

}

func (this *LedProcService) Load() {

}

type FileService struct {
	logger *slog.Logger
}

func NewFileService(logger *slog.Logger) *FileService {
	if logger == nil {
		logger = common.NewConsoleLogger(slog.LevelDebug)
	}

	return &FileService{
		logger: logger,
	}
}

func (this *FileService) AppendToFile(filePath string, payload []byte) error {
	filePath = filepath.Join(lightShowDir, filePath)
	file, err := getOrCreateFile(filePath)
	if err != nil {
		fmt.Printf(
			"[FILE_SERVICE][ERROR] file fetching failed. Relative file path: %s. Error: %s\n",
			filePath, err)
		return err
	}

	defer file.Close()
	_, err = file.Write(payload)
	return err
}

func (this *FileService) ReadFileContent(filePath string) []byte {
	filePath = filepath.Join(lightShowDir, filePath)
	file, err := getOrCreateFile(filePath)
	if err != nil {
		fmt.Printf(
			"[FILE_SERVICE][ERROR] file fetching failed. Relative file path: %s. Error: %s\n",
			filePath, err)
		return []byte{}
	}

	defer file.Close()
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Printf("[FILE_SERVICE] end of file.\n")
			return []byte{}
		}

		fmt.Printf(
			"[FILE_SERVICE][ERROR] file reading failed. Relative file path: %s. Error: %s\n",
			filePath, err)
		return []byte{}
	}

	return buffer[:n]
}

func (this *FileService) FindFiles(glob string) {

}

func createFolderIfNotExists(directory string) (fullPath string, err error) {
	fullPath, err = common.GetFullPathFromRelativePath(directory)
	if err != nil {
		// TODO: remove logging here as it is redundant?
		fmt.Printf("[ERROR] failed to read full path: %s\n", err)
		return "", err
	}

	err = os.MkdirAll(fullPath, secureDirMode)
	if err != nil {
		fmt.Printf("[ERROR] failed to make directory: %s\n", err)
		return "", err
	}

	return fullPath, err
}

func getOrCreateFile(filePath string) (*os.File, error) {
	relativeDir := common.GetRelativeDirFromRelativePath(filePath)
	_, err := createFolderIfNotExists(relativeDir)
	if err != nil {
		return nil, err
	}

	fullPath, err := common.GetFullPathFromRelativePath(filePath)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, secureFileMode)
}
