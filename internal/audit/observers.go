package audit

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/audit/observer"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/file"
)

// InitObservers initializes and configures all audit observers based on the configuration.
// It creates HTTP observers for remote servers and file observers for local file logging
// as specified in the audit configuration.
//
// Parameters:
//   - cfg: Audit configuration specifying which observers to enable
//   - l: Structured logger for logging operations
//
// Returns:
//   - []Observer: List of initialized and started observers
//   - error: nil on success, or error if observer creation fails
//
// Note: HTTP observers are started immediately, file observers are initialized but not started.
func InitObservers(cfg config.Audit, l *zap.Logger) ([]Observer, error) {
	observers := make([]Observer, 0, 2)
	if cfg.URL != "" {
		ho := observer.NewHTTP(cfg, l)
		ho.Start()
		observers = append(observers, ho)
	}
	if cfg.File != "" {
		fm := file.NewManager(cfg.File, "", l)
		fo, err := observer.NewFile(fm, l)
		if err != nil {
			return nil, fmt.Errorf("create file observer: %w", err)
		}
		observers = append(observers, fo)
	}
	return observers, nil
}
