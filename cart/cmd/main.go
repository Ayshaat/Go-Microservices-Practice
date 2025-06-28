package main

import (
	"cart/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("app failed: %v", err)
	}
}
