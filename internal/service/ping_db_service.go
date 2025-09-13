package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type IPingDBService interface {
	Ping() error
}

type DBPinger interface {
	PingContext(ctx context.Context) error
}

type PingDBService struct {
	db     DBPinger
	logger *zap.Logger
}

func NewPingDBService(db DBPinger, logger *zap.Logger) *PingDBService {
	logger = logger.With(zap.String("package", "ping_db_service"))
	return &PingDBService{
		db:     db,
		logger: logger,
	}
}

func (s *PingDBService) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		s.logger.Error("failed to ping database", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrFailedToPingDB, err)
	}
	return nil
}

var (
	ErrFailedToPingDB = errors.New("failed to ping database")
)
