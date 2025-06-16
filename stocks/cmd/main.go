package main

import (
	"log"
	"net/http"
	"stocks/internal/delivery"
	"stocks/internal/repository"
	"stocks/internal/usecase"
)

func main() {
	repo := repository.NewInMemoryStockRepo()
	useCase := usecase.NewStockUsecase(repo)
	handler := delivery.NewHandler(useCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}

	log.Println("Starting server on :8080...")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("stocks server failed: %v", err)
	}
}
