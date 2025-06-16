package main

import (
	"cart/internal/delivery"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"log"
	"net/http"
)

func main() {
	cartRepo := repository.NewInMemoryCartRepo()
	stockClient := stockclient.New("http://localhost:8080")
	cartUseCase := usecase.NewCartUsecase(cartRepo, stockClient)
	handler := delivery.NewHandler(cartUseCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":8090",
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Println("Starting cart server on :8090...")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
