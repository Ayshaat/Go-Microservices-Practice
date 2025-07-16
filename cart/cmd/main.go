package main

import (
	"cart/internal/app"
	"log"
	"os"
)

func main() {
	file := os.Getenv("ENV")
	switch file {
	case "production":
		file = "/app/.env.docker"
	case "local":
		file = ".env.local"
	default:
		log.Fatalf("Environment variable ENV is not set.")
	}

	log.Printf("Loading environment variables from %s", file)

	if err := app.Run(file); err != nil {
		log.Fatalf("app failed: %v", err)
	}
}
