package main

import (
	"log"
	"stocks/internal/app"
)

func main() {
	if err := app.Run(".env.docker"); err != nil {
		log.Fatalf("stocks app failed: %v", err)
	}
}
