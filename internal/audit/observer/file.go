package observer

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/model"
)

// File is an observer that writes audit events to a local file.
// It ensures thread-safe file operations with proper locking and append semantics.
type File struct {
	fm     *file.Manager
	mu     *sync.Mutex
	logger *zap.Logger
	f      *os.File
}

// NewFile creates a new File observer with the specified file manager.
//
// Parameters:
//   - fm: File manager for handling file operations
//   - l: Structured logger for logging operations
//
// Returns:
//   - *File: Initialized file observer
//   - error: nil on success, or error if file operations fail
func NewFile(fm *file.Manager, l *zap.Logger) (*File, error) {
	return &File{
		fm:     fm,
		mu:     &sync.Mutex{},
		logger: l,
	}, nil
}

// Name returns human-readable identifier for the observer.
func (o *File) Name() string {
	return "audit_file"
}

// Notify writes an audit event to the file in JSON format.
// It handles file opening, writing, and closing for each event with proper error handling.
//
// Parameters:
//   - ctx: Context (currently unused but part of interface)
//   - e: AuditEvent to write to the file
func (o *File) Notify(_ context.Context, e model.AuditEvent) {
	o.logger.Debug("event received", zap.Any("event", e))
	b, err := e.ToJSON()
	if err != nil {
		o.logger.Error("error encoding audit event", zap.Error(err))
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, err = o.fm.OpenForAppend(false); err != nil {
		o.logger.Error("error opening file for append", zap.Error(err))
		return
	}
	defer func() {
		if cErr := o.fm.Close(); cErr != nil {
			o.logger.Error("error closing file after append data", zap.Error(cErr))
		}
	}()

	if err := o.fm.WriteData(b); err != nil {
		o.logger.Error("error writing audit event to file", zap.Error(err))
		return
	}
}

// Close releases file resources and ensures all data is flushed.
//
// Parameters:
//   - ctx: Context (currently unused but part of interface)
//
// Returns:
//   - error: nil on success, or error if file closing fails
func (o *File) Close(_ context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.fm.Close(); err != nil {
		return fmt.Errorf("closing file manager: %w", err)
	}
	o.logger.Info("audit file observer closed")
	return nil
}
