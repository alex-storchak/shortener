package config

import (
	handlerConfig "github.com/alex-storchak/shortener/internal/handler/config"
)

type Config struct {
	Handler handlerConfig.Config
}

func GetConfig() Config {
	return Config{
		Handler: handlerConfig.Config{
			ServerHost: "localhost",
			ServerPort: "8080",
		},
	}
}
