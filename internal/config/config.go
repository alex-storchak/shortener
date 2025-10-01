package config

import (
	"flag"

	dbCfg "github.com/alex-storchak/shortener/internal/db/config"
	handlerCfg "github.com/alex-storchak/shortener/internal/handler/config"
	loggerCfg "github.com/alex-storchak/shortener/internal/logger/config"
	mwCfg "github.com/alex-storchak/shortener/internal/middleware/config"
	repoCfg "github.com/alex-storchak/shortener/internal/repository/config"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Handler    handlerCfg.Config
	Logger     loggerCfg.Config
	Repository repoCfg.Config
	DB         dbCfg.Config
	Middleware mwCfg.Config
}

func ParseConfig() (*Config, error) {
	cfg := Config{}
	parseFlags(&cfg)
	err := parseEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Handler.ServerAddr, "a", handlerCfg.DefaultServerAddr, "address of HTTP server")
	flag.StringVar(&cfg.Handler.BaseURL, "b", handlerCfg.DefaultBaseURL, "base URL of short url service")
	flag.StringVar(&cfg.Logger.LogLevel, "l", loggerCfg.DefaultLogLevel, "log level")
	flag.StringVar(&cfg.Repository.FileStorage.Path, "f", repoCfg.DefaultFileStoragePath, "db storage file path")
	flag.StringVar(&cfg.DB.DSN, "d", dbCfg.DefaultDatabaseDSN, "postgres database DSN")

	flag.StringVar(&cfg.Middleware.AuthCookieName, "auth-cookie-name", mwCfg.DefaultAuthCookieName, "auth cookie name")
	flag.DurationVar(&cfg.Middleware.AuthTokenMaxAge, "auth-token-max-age", mwCfg.DefaultAuthTokenMaxAge, "auth token max age in hours")
	flag.DurationVar(&cfg.Middleware.AuthRefreshThreshold, "auth-refresh-threshold", mwCfg.DefaultAuthRefreshThreshold, "auth refresh threshold in hours")
	flag.StringVar(&cfg.Middleware.SecretKey, "auth-secret-key", mwCfg.DefaultAuthSecretKey, "auth JWT secret key")

	flag.Parse()
}

func parseEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}
	return nil
}
