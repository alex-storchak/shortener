package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/alex-storchak/shortener/internal/config"
)

func New(cfg *config.Logger) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	zcfg := zap.NewProductionConfig()
	zcfg.Level = lvl
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zl, err := zcfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return zl, nil
}
