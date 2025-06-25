package main

import (
	"log"
	"stocks/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("stocks app failed: %v", err)
	}
}
