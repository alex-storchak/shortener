// Package logger provides structured logging configuration and initialization
// for the URL shortener service using the Uber Zap logging framework.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/alex-storchak/shortener/internal/config"
)

// New creates and configures a new Zap logger instance with the specified settings.
// The logger is optimized for development and production use with console encoding
// and ISO8601 timestamp formatting for better readability.
//
// Parameters:
//   - cfg: Logger configuration containing the desired log level
//
// Returns:
//   - *zap.Logger: Configured logger instance ready for use
//   - error: Configuration or initialization error if any occurs
//
// Example:
//
//	cfg := &config.Logger{LogLevel: "debug"}
//	logger, err := New(cfg)
//	if err != nil {
//	    // handle error
//	}
//	defer logger.Sync()
//
//	logger.Info("application started")
//
// The function validates the log level and returns an error if an invalid
// level is specified. The returned logger should be properly closed with
// Sync() when the application shuts down to flush any buffered log entries.
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
