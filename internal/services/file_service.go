package services

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

const (
	bufferSize       = 4096
	maxFileSize      = 100_000_000 // 100MB
	backupFileSuffix = "_bak"
)

// FileService requires relative paths for actions to work correctly
type FileService struct {
	logger *slog.Logger
}

func NewFileService(logger *slog.Logger) *FileService {
	return &FileService{
		logger: common.EnsureLoggerExists(logger),
	}
}

func (this *FileService) AppendToFile(filePath string, payloads ...[]byte) error {
	file, err := this.getOrCreateFile(filePath)
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

func (this *FileService) PeakFileContent(filePath string, length uint32) []byte {
	file, err := this.getOrCreateFile(filePath)
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] file fetching failed",
			"relative file path", filePath,
			"error", err)
		return []byte{}
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] failed to read file information",
			"relative file path", filePath,
			"error", err)
		return []byte{}
	}

	if info.Size() > maxFileSize {
		this.logger.Warn(
			"[FILE_SERVICE] file exceeds maximum allowed size - file will not be read fully",
			"relative file path", filePath,
			"maximum file size (MB)", maxFileSize/1_000_000)
	}
	if info.Size() < int64(length) {
		length = uint32(info.Size()) + 1 // Additional byte to trigger end of file
	}

	content := make([]byte, 0, length+1)
	for {
		n, err := file.Read(content[len(content):cap(content)])
		content = content[:len(content)+n]
		if err != nil {
			if err == io.EOF {
				this.logger.Debug("[FILE_SERVICE] end of file")
			} else {
				this.logger.Error(
					"[FILE_SERVICE] failed to read file content",
					"relative file path", filePath,
					"error", err)
			}
			return content
		}

		// Short circuit when reading before end of file
		if n == 0 {
			return content
		}
	}
}

func (this *FileService) ReadFileContent(filePath string) []byte {
	return this.PeakFileContent(filePath, 1_000_000) // For now only ready up to 1MB of content
}

func (this *FileService) SaveFile(filePath string, payloads ...[]byte) error {
	err := this.copyFile(filePath, filePath+backupFileSuffix)
	if err != nil {
		return err
	}

	err = this.deleteFile(filePath)
	if err != nil {
		return err
	}

	err = this.AppendToFile(filePath, payloads...)
	if err != nil {
		this.logger.Warn("[FILE_SERVICE] failed to save new file, restoring backup")
		if innerErr := this.deleteFile(filePath); innerErr != nil {
			this.logger.Error(
				"[FILE_SERVICE] error while deleting malformed file",
				"inner error", innerErr)
		}
		if innerErr := this.copyFile(filePath+backupFileSuffix, filePath); innerErr != nil {
			this.logger.Error(
				"[FILE_SERVICE] failed to restore backed up file",
				"backed up file path", filePath+backupFileSuffix,
				"inner error", innerErr)
		}
		return err
	}

	this.deleteFile(filePath + backupFileSuffix) // intentionally ignored error
	return nil
}

func (this *FileService) FindFiles(globPath string) []string {
	files, err := this.getFiles(globPath)
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] failed to find files",
			"pattern", globPath,
			"error", err)
	}
	return files
}

func (this *FileService) copyFile(sourcePath, destinationPath string) error {
	source, err := this.getOrCreateFile(sourcePath)
	if err != nil {
		return err
	}

	destination, err := this.getOrCreateFile(destinationPath)
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

func (this *FileService) createFolderIfNotExists(directory string) (fullPath string, err error) {
	fullPath, err = common.GetFullPathFromRelativePath(directory)
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] failed to read full path",
			"relative path", directory,
			"error", err)
		return "", err
	}

	err = os.MkdirAll(fullPath, secureDirMode)
	if err != nil {
		this.logger.Error(
			"[FILE_SERVICE] failed to make directory",
			"absolute path", fullPath,
			"error", err)
		return "", err
	}

	return fullPath, nil
}

func (this *FileService) getOrCreateFile(filePath string) (*os.File, error) {
	relativeDir := common.GetRelativeDirFromRelativePath(filePath)
	_, err := this.createFolderIfNotExists(relativeDir)
	if err != nil {
		return nil, err
	}

	fullPath, err := common.GetFullPathFromRelativePath(filePath)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, secureFileMode)
}

func (this *FileService) deleteFile(filePath string) error {
	fullPath, err := common.GetFullPathFromRelativePath(filePath)
	if err != nil {
		return err
	}

	return os.Remove(fullPath)
}

func (this *FileService) getFiles(globPath string) ([]string, error) {
	var files []string
	fullPath, err := common.GetFullPathFromRelativePath(globPath)
	if err != nil {
		return files, err
	}

	files, err = filepath.Glob(fullPath)
	return files, err
}
