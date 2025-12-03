package file

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/filepath"
)

type Manager struct {
	path     string
	defPath  string
	file     *os.File
	isClosed bool
	logger   *zap.Logger
	writer   *bufio.Writer
}

func NewManager(path, defPath string, logger *zap.Logger) *Manager {
	return &Manager{
		path:    path,
		defPath: defPath,
		logger:  logger,
	}
}

func (m *Manager) getAbsPath(filePath string) (string, error) {
	absPath, err := filepath.GetAbsPath(filePath)
	if err != nil {
		return "", fmt.Errorf("get absolute path for `%s`: %w", filePath, err)
	}
	return absPath, nil
}

func (m *Manager) open(useDefault bool, flag int) (*os.File, error) {
	if useDefault {
		m.logger.Info("Using default file path")
		m.path = m.defPath
	}
	absPath, err := m.getAbsPath(m.path)
	if err != nil {
		return nil, fmt.Errorf("get absolute path for `%s`: %w", m.path, err)
	}
	file, err := os.OpenFile(absPath, flag, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file by path `%s`: %w", absPath, err)
	}
	m.file = file
	m.isClosed = false
	m.writer = bufio.NewWriter(file)
	return m.file, nil
}

func (m *Manager) OpenForAppend(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_APPEND)
}

func (m *Manager) OpenForWrite(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

func (m *Manager) Close() error {
	if m.file != nil && !m.isClosed {
		if err := m.file.Close(); err != nil {
			return fmt.Errorf("close file: %w", err)
		}
		m.isClosed = true
		m.logger.Debug("file closed",
			zap.String("path", m.path),
			zap.Bool("isClosed", m.isClosed),
		)
	}
	return nil
}

func (m *Manager) WriteData(data []byte) error {
	if _, err := m.writer.Write(data); err != nil {
		return fmt.Errorf("write data to file: %w", err)
	}
	if err := m.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("add line break to file: %w", err)
	}
	return m.writer.Flush()
}
