package logger

import (
	"github.com/alex-storchak/shortener/internal/logger/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetInstance(cfg *config.Config) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	zcfg := zap.NewProductionConfig()
	zcfg.Level = lvl
	zcfg.Encoding = "console"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zl, err := zcfg.Build()
	if err != nil {
		return nil, err
	}

	return zl, nil
}
