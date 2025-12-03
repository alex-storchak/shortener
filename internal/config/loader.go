package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Load initializes and loads the application configuration from all available sources.
// It parses command-line flags first, then environment variables, with environment
// variables taking precedence over flag values.
//
// Returns:
//   - *Config: Fully populated configuration structure
//   - error: nil on success, or error if environment parsing fails
func Load() (*Config, error) {
	cfg := &Config{}
	parseFlags(cfg)
	err := parseEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("parse env vars: %w", err)
	}
	return cfg, nil
}

// parseFlags parses command-line flags and sets configuration values.
// Flags provide the base configuration that can be overridden by environment variables.
//
// Parameters:
//   - cfg: Pointer to Config structure to populate with flag values
func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Server.ServerAddr, "a", DefServerAddr, "address of HTTP server")
	flag.DurationVar(&cfg.Server.ShutdownWaitSecsDuration, "shutdown-wait-secs-duration", DefShutdownWaitSecsDuration, "server shutdown wait seconds duration")

	flag.StringVar(&cfg.Handler.BaseURL, "b", DefBaseURL, "base URL of short url service")

	flag.StringVar(&cfg.Logger.LogLevel, "l", DefLogLevel, "log level")

	flag.StringVar(&cfg.Repo.FileStoragePath, "f", DefFileStoragePath, "db storage file path")

	flag.StringVar(&cfg.DB.DSN, "d", DefDatabaseDSN, "postgres database DSN")
	flag.StringVar(&cfg.DB.MigrationsPath, "m", DefMigrationsPath, "postgres database migrations path")

	flag.StringVar(&cfg.Auth.CookieName, "auth-cookie-name", DefAuthCookieName, "auth cookie name")
	flag.DurationVar(&cfg.Auth.TokenMaxAge, "auth-token-max-age", DefAuthTokenMaxAge, "auth token max age in hours")
	flag.DurationVar(&cfg.Auth.RefreshThreshold, "auth-refresh-threshold", DefAuthRefreshThreshold, "auth refresh threshold in hours")
	flag.StringVar(&cfg.Auth.SecretKey, "auth-secret-key", DefAuthSecretKey, "auth JWT secret key")

	flag.StringVar(&cfg.Audit.File, "audit-file", DefAuditFile, "audit log file path")
	flag.StringVar(&cfg.Audit.URL, "audit-url", DefAuditURL, "full URL of audit server")
	flag.IntVar(&cfg.Audit.EventChanSize, "audit-event-chan-size", DefAuditEventChanSize, "audit event chan size")
	flag.IntVar(&cfg.Audit.HTTPWorkersCount, "audit-http-workers-count", DefAuditHTTPWorkersCount, "audit http workers count")
	flag.DurationVar(&cfg.Audit.HTTPTimeout, "audit-http-timeout", DefAuditHTTPTimeout, "audit http timeout")

	flag.Parse()
}

// parseEnv parses environment variables and overrides any previously set flag values.
// Environment variables have the highest precedence in the configuration hierarchy.
//
// Parameters:
//   - cfg: Pointer to Config structure to populate with environment values
//
// Returns:
//   - error: nil on success, or error if environment parsing fails
func parseEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("parse env with lib: %w", err)
	}
	return nil
}
