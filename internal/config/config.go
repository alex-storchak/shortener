package config

import (
	"flag"

	handlerCfg "github.com/alex-storchak/shortener/internal/handler/config"
	loggerCfg "github.com/alex-storchak/shortener/internal/logger/config"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Handler handlerCfg.Config
	Logger  loggerCfg.Config
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
	flag.Parse()
}

func parseEnv(cfg *Config) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}
	return nil
}
