package main

import (
	"log"

	"app/config"
	"app/internal/app"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
