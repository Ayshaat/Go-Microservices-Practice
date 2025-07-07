package main

import (
	"log"
	"os"
	"stocks/internal/app"
)

func main() {
	file := os.Getenv("ENV")
	switch file {
	case "production":
		file = ".env.local"
	case "testing":
		file = ".env.docker"
	default:
		log.Println("Environment variable ENV is not set.")
		file = ".env.local"
	}

	if err := app.Run(file); err != nil {
		log.Fatalf("stocks app failed: %v", err)
	}
}
