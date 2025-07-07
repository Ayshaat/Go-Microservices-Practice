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
		file = ".env.docker"
	case "testing":
		file = ".env.docker"
	default:
		log.Println("Environment variable ENV is not set.")
		file = ".env.local"
	}

	if err := app.Run(file); err != nil {
		log.Fatalf("app failed: %v", err)
	}
}
