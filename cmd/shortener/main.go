package main

import (
	"log"

	"github.com/alex-storchak/shortener/internal/application"
)

func main() {
	app, err := application.NewApp()
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			log.Fatalf("failed to close application: %v", err)
		}
	}()
	if err := app.Run(); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}
