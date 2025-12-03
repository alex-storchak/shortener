package observer

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/alex-storchak/shortener/internal/file"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type File struct {
	fm     *file.Manager
	mu     *sync.Mutex
	logger *zap.Logger
	f      *os.File
}

func NewFile(fm *file.Manager, l *zap.Logger) (*File, error) {
	return &File{
		fm:     fm,
		mu:     &sync.Mutex{},
		logger: l,
	}, nil
}

func (o *File) Name() string {
	return "audit_file"
}

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

func (o *File) Close(_ context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if err := o.fm.Close(); err != nil {
		return fmt.Errorf("closing file manager: %w", err)
	}
	o.logger.Info("audit file observer closed")
	return nil
}
