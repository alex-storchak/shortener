package main

import (
	"errors"
	"log"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	cfg := config.GetConfig()

	shortener, err := service.NewShortener()
	if err != nil {
		return errors.New("failed to instantiate shortener: " + err.Error())
	}

	return handler.Serve(cfg.Handler, shortener)
}
