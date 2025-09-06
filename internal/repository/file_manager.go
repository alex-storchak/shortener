package repository

import (
	"bufio"
	"os"

	"github.com/alex-storchak/shortener/internal/helper"
	"go.uber.org/zap"
)

type FileManager struct {
	filePath     string
	dfltFilePath string
	file         *os.File
	logger       *zap.Logger
	writer       *bufio.Writer
}

func NewFileManager(filePath, dfltFilePath string, logger *zap.Logger) *FileManager {
	logger = logger.With(
		zap.String("component", "file manager"),
	)
	return &FileManager{
		filePath:     filePath,
		dfltFilePath: dfltFilePath,
		logger:       logger,
	}
}

func (m *FileManager) getAbsPath(filePath string) (string, error) {
	absPath, err := helper.GetAbsFilePath(filePath)
	if err != nil {
		m.logger.Error("Failed to get absolute path", zap.String("file_path", filePath), zap.Error(err))
		return "", err
	}
	return absPath, nil
}

func (m *FileManager) open(useDefault bool) (*os.File, error) {
	if useDefault {
		m.logger.Info("Using default file path")
		m.filePath = m.dfltFilePath
	}
	absPath, err := m.getAbsPath(m.filePath)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		m.logger.Error("Can't open file by path", zap.String("path", absPath), zap.Error(err))
		return nil, err
	}
	m.file = file
	m.writer = bufio.NewWriter(file)
	return m.file, nil
}

func (m *FileManager) close() error {
	if m.file != nil {
		return m.file.Close()
	}
	return nil
}

func (m *FileManager) writeData(data []byte) error {
	if _, err := m.writer.Write(data); err != nil {
		m.logger.Error("Can't write data to file", zap.Error(err))
		return err
	}
	if err := m.writer.WriteByte('\n'); err != nil {
		m.logger.Error("Can't add line break in file", zap.Error(err))
		return err
	}
	return m.writer.Flush()
}
