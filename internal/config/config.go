package config

import (
	"flag"

	handlerConfig "github.com/alex-storchak/shortener/internal/handler/config"
)

type Config struct {
	Handler handlerConfig.Config
}

func GetConfig() Config {
	cfg := Config{}
	parseFlags(&cfg)
	return cfg
}

func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.Handler.ServerAddr, "a", "localhost:8080", "address of HTTP server")
	flag.StringVar(
		&cfg.Handler.ShortURLBaseAddr,
		"b",
		"http://localhost:8080",
		"base address of short url service",
	)

	flag.Parse()
}
