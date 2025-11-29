// Package file provides file management utilities.
// It handles file operations, path resolution, and buffered writing for storage persistence.
//
// # Core Components
//
// Main file management functionality:
//   - Manager: central file handler with buffered I/O operations
//   - Path resolution with absolute path conversion
//   - Support for default fallback paths
//
// # File Operations
//
// Comprehensive file handling capabilities:
//   - OpenForAppend: open files for appending data (used for incremental updates)
//   - OpenForWrite: open files for writing (truncates existing content)
//   - WriteData: buffered writing with automatic line termination
//   - Close: safe file closure with state tracking
//
// # Path Management
//
// Path handling features:
//   - Absolute path resolution for consistent file access
//   - Default path fallback when primary paths are unavailable
//   - Configurable primary and fallback paths
package file

import (
	"bufio"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/helper/filepath"
)

// Manager provides file management operations.
// It handles file opening, closing, writing, and path resolution with support
// for fallback to default paths when primary paths are unavailable.
//
// The manager maintains file state and provides buffered writing for efficiency.
type Manager struct {
	path     string
	defPath  string
	file     *os.File
	isClosed bool
	logger   *zap.Logger
	writer   *bufio.Writer
}

// NewManager creates a new file manager instance.
//
// Parameters:
//   - path: primary file path for operations
//   - defPath: fallback default file path
//   - logger: structured logger for logging operations
//
// Returns:
//   - *Manager: configured file manager
func NewManager(path, defPath string, logger *zap.Logger) *Manager {
	return &Manager{
		path:    path,
		defPath: defPath,
		logger:  logger,
	}
}

// getAbsPath converts a relative file path to an absolute path.
func (m *Manager) getAbsPath(filePath string) (string, error) {
	absPath, err := filepath.GetAbsPath(filePath)
	if err != nil {
		return "", fmt.Errorf("get absolute path for `%s`: %w", filePath, err)
	}
	return absPath, nil
}

// open opens a file with the specified flags, with optional fallback to default path.
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

// OpenForAppend opens a file for appending data.
// If useDefault is `true`, it uses the default fallback path instead of primary path.
//
// Parameters:
//   - useDefault: whether to use the default fallback path
//
// Returns:
//   - *os.File: opened file handle
//   - error: nil on success, or error if file opening fails
func (m *Manager) OpenForAppend(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_APPEND)
}

// OpenForWrite opens a file for writing, truncating existing content.
// If useDefault is `true`, it uses the default fallback path instead of primary path.
//
// Parameters:
//   - useDefault: whether to use the default fallback path
//
// Returns:
//   - *os.File: opened file handle
//   - error: nil on success, or error if file opening fails
func (m *Manager) OpenForWrite(useDefault bool) (*os.File, error) {
	return m.open(useDefault, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
}

// Close closes the managed file if it's currently open.
// Implements safe closing with state tracking to prevent double closure.
//
// Returns:
//   - error: nil on success, or error if file closure fails
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

// WriteData writes data to the managed file with automatic line termination.
// Uses buffered writing for efficiency.
//
// Parameters:
//   - data: byte data to write to the file
//
// Returns:
//   - error: nil on success, or error if write operation fails
func (m *Manager) WriteData(data []byte) error {
	if _, err := m.writer.Write(data); err != nil {
		return fmt.Errorf("write data to file: %w", err)
	}
	if err := m.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("add line break to file: %w", err)
	}
	return m.writer.Flush()
}
