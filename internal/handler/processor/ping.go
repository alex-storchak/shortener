package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/alex-storchak/shortener/internal/service"
	"go.uber.org/zap"
)

type Ping struct {
	pinger service.Pinger
	logger *zap.Logger
}

func NewPing(pinger service.Pinger, logger *zap.Logger) *Ping {
	return &Ping{
		pinger: pinger,
		logger: logger,
	}
}

func (s *Ping) Process() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.pinger.IsReady(ctx); err != nil {
		return fmt.Errorf("ping service: %w", err)
	}
	return nil
}
