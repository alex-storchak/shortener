package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/alex-storchak/shortener/internal/app"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	printBuildInfo()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("failed to run application: %v", err)
	}
}

func printBuildInfo() {
	valueOrNA := func(s string) string {
		if s == "" {
			return "N/A"
		}
		return s
	}
	fmt.Printf("Build version: %s\n", valueOrNA(buildVersion))
	fmt.Printf("Build date: %s\n", valueOrNA(buildDate))
	fmt.Printf("Build commit: %s\n", valueOrNA(buildCommit))
}
