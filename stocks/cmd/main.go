package main

import (
	"log"
	"net/http"
	"os"
	"stocks/internal/db"
	"stocks/internal/delivery"
	"stocks/internal/repository"
	"stocks/internal/usecase"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")

	sqlDB, err := db.ConnectDB(connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	repo := repository.NewPostgresStockRepo(sqlDB)
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

	log.Println("Starting server on :8080...")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("stocks server failed: %v", err)
	}
}
