package repository

import (
	"bufio"
	"fmt"
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
	return &FileManager{
		filePath:     filePath,
		dfltFilePath: dfltFilePath,
		logger:       logger,
	}
}

func (m *FileManager) getAbsPath(filePath string) (string, error) {
	absPath, err := helper.GetAbsFilePath(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for `%s`: %w", filePath, err)
	}
	return absPath, nil
}

func (m *FileManager) open(useDefault bool, flag int) (*os.File, error) {
	if useDefault {
		m.logger.Info("Using default file path")
		m.filePath = m.dfltFilePath
	}
	absPath, err := m.getAbsPath(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for `%s`: %w", m.filePath, err)
	}
	file, err := os.OpenFile(absPath, flag, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file by path `%s`: %w", absPath, err)
	}
	m.file = file
	m.writer = bufio.NewWriter(file)
	return m.file, nil
}

func (m *FileManager) openForAppend(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_APPEND)
}

func (m *FileManager) openForWrite(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

func (m *FileManager) close() error {
	if m.file != nil {
		return m.file.Close()
	}
	return nil
}

func (m *FileManager) writeData(data []byte) error {
	if _, err := m.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}
	if err := m.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to add line break to file: %w", err)
	}
	return m.writer.Flush()
}
