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
	shortIDGenerator := service.NewShortidIDGenerator(generator)
	shortener := service.NewShortener(
		shortIDGenerator,
		repository.NewMapURLStorage(),
		repository.NewMapURLStorage(),
	)

	return handler.Serve(cfg.Handler, shortener)
}
