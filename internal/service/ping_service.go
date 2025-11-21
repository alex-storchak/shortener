package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Pinger interface {
	IsReady(ctx context.Context) error
}

type PingService struct {
	pinger Pinger
	logger *zap.Logger
}

func NewPingService(pinger Pinger, logger *zap.Logger) *PingService {
	return &PingService{
		pinger: pinger,
		logger: logger,
	}
}

func (s *PingService) Process() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.pinger.IsReady(ctx); err != nil {
		return fmt.Errorf("ping service: %w", err)
	}
	return nil
}
