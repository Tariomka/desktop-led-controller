package services

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

type FileService struct {
	logger *slog.Logger
}

func NewFileService(logger *slog.Logger) *FileService {
	return &FileService{
		logger: common.EnsureLoggerExists(logger),
	}
}

func (this *FileService) AppendToFile(filePath string, payload []byte) error {
	file, err := getOrCreateFile(filePath)
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] file fetching failed",
			"relative file path", filePath,
			"error", err)
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
		this.logger.Error(
			"[FILE_SERVICE] file fetching failed",
			"relative file path", filePath,
			"error", err)
		return []byte{}
	}

	defer file.Close()
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			this.logger.Debug("[FILE_SERVICE] end of file")
			return []byte{}
		}

		this.logger.Error(
			"[FILE_SERVICE] file reading failed",
			"relative file path", filePath,
			"error", err)
		return []byte{}
	}

	return buffer[:n]
}

func (this *FileService) FindFiles(globPath string) {

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
