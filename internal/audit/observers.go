package audit

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/audit/observer"
	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/file"
)

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
