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

const (
	bufferSize       = 4096
	backupFileSuffix = "_bak"
)

type FileService struct {
	logger *slog.Logger
}

func NewFileService(logger *slog.Logger) *FileService {
	return &FileService{
		logger: common.EnsureLoggerExists(logger),
	}
}

func (this *FileService) AppendToFile(filePath string, payloads ...[]byte) error {
	file, err := getOrCreateFile(filePath)
	if err != nil {
		return err
	}

	defer file.Close()
	for _, payload := range payloads {
		payload = append(payload, '\n')
		if _, err = file.Write(payload); err != nil {
			return err
		}
	}

	return nil
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
	buffer := make([]byte, bufferSize) // TODO: keep reading?
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

func (this *FileService) SaveFile(filePath string, payloads ...[]byte) error {
	err := copyFile(filePath, filePath+backupFileSuffix)
	if err != nil {
		return err
	}

	err = deleteFile(filePath)
	if err != nil {
		return err
	}

	err = this.AppendToFile(filePath, payloads...)
	if err != nil {
		this.logger.Warn("[FILE_SERVICE] failed to save new file, restoring backup")
		if innerErr := deleteFile(filePath); innerErr != nil {
			this.logger.Error(
				"[FILE_SERVICE] error while deleting malformed file",
				"inner error", innerErr)
		}
		if innerErr := copyFile(filePath+backupFileSuffix, filePath); innerErr != nil {
			this.logger.Error(
				"[FILE_SERVICE] failed to restore backed up file",
				"backed up file path", filePath+backupFileSuffix,
				"inner error", innerErr)
		}
		return err
	}

	deleteFile(filePath + backupFileSuffix) // intentionally ignored error
	return nil
}

func (this *FileService) FindFiles(globPath string) {
	panic(common.ErrNotImplemented)
}

func copyFile(sourcePath, destinationPath string) error {
	source, err := getOrCreateFile(sourcePath)
	if err != nil {
		return err
	}

	destination, err := getOrCreateFile(destinationPath)
	if err != nil {
		return err
	}

	buffer := make([]byte, bufferSize)
	for {
		n, err := source.Read(buffer)
		if errors.Is(err, io.EOF) || n == 0 {
			break
		}

		if err != nil {
			return err
		}

		if _, err := destination.Write(buffer[:n]); err != nil {
			return err
		}
	}

	return nil
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

func deleteFile(filePath string) error {
	fullPath, err := common.GetFullPathFromRelativePath(filePath)
	if err != nil {
		return err
	}

	return os.Remove(fullPath)
}
