package processor

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/service"
)

// Ping provides health check functionality for service monitoring.
// It handles the business logic for the '/ping' endpoint.
type Ping struct {
	pinger service.Pinger
	logger *zap.Logger
}

// NewPing creates a new Ping processor instance.
//
// Parameters:
//   - pinger: Service that implements health checking
//   - logger: Structured logger for logging operations
//
// Returns: configured Ping processor
func NewPing(pinger service.Pinger, logger *zap.Logger) *Ping {
	return &Ping{
		pinger: pinger,
		logger: logger,
	}
}

// Process performs a health check by pinging the service dependencies.
// It uses a 1-second timeout context for the health check.
//
// Returns:
//   - error: nil if service is healthy, or error if health check fails
func (s *Ping) Process() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.pinger.IsReady(ctx); err != nil {
		return fmt.Errorf("ping service: %w", err)
	}
	return nil
}
