package main

import (
	"log"

	"app/config"
	"app/internal/app"
)

func main() {
	cfg := config.Load()
	log.Printf("server port: %s", cfg.ServerPort)
	log.Printf("database target: %s", cfg.DatabaseTarget())

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
