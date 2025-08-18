package main

import (
	"errors"
	"log"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/handler"
	"github.com/alex-storchak/shortener/internal/repository"
	"github.com/alex-storchak/shortener/internal/service"
	"github.com/teris-io/shortid"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	cfg := config.GetConfig()

	generator, err := shortid.New(1, shortid.DefaultABC, 1)
	if err != nil {
		return errors.New("failed to instantiate shortid generator")
	}
	shortIDGenerator := service.NewShortIDGenerator(generator)
	urlToShortStorage := repository.NewMapURLStorage()
	shortToURLStorage := repository.NewMapURLStorage()
	shortener := service.NewShortener(
		shortIDGenerator,
		urlToShortStorage,
		shortToURLStorage,
	)

	return handler.Serve(cfg.Handler, shortener)
}
