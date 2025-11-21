package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

func Load() (*Config, error) {
	cfg := &Config{}
	parseFlags(cfg)
	err := parseEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("parse env vars: %w", err)
	}
	return cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Server.ServerAddr, "a", DefServerAddr, "address of HTTP server")
	flag.DurationVar(&cfg.Server.ShutdownWaitSecsDuration, "shutdown-wait-secs-duration", DefShutdownWaitSecsDuration, "server shutdown wait seconds duration")

	flag.StringVar(&cfg.Handler.BaseURL, "b", DefBaseURL, "base URL of short url service")

	flag.StringVar(&cfg.Logger.LogLevel, "l", DefLogLevel, "log level")

	flag.StringVar(&cfg.Repo.FileStoragePath, "f", DefFileStoragePath, "db storage file path")
	flag.StringVar(&cfg.DB.DSN, "d", DefDatabaseDSN, "postgres database DSN")

	flag.StringVar(&cfg.Auth.CookieName, "auth-cookie-name", DefAuthCookieName, "auth cookie name")
	flag.DurationVar(&cfg.Auth.TokenMaxAge, "auth-token-max-age", DefAuthTokenMaxAge, "auth token max age in hours")
	flag.DurationVar(&cfg.Auth.RefreshThreshold, "auth-refresh-threshold", DefAuthRefreshThreshold, "auth refresh threshold in hours")
	flag.StringVar(&cfg.Auth.SecretKey, "auth-secret-key", DefAuthSecretKey, "auth JWT secret key")

	flag.StringVar(&cfg.Audit.File, "audit-file", DefAuditFile, "audit log file path")
	flag.StringVar(&cfg.Audit.URL, "audit-url", DefAuditURL, "full URL of audit server")

	flag.Parse()
}

func parseEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("parse env with lib: %w", err)
	}
	return nil
}
