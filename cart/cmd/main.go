package main

import (
	"cart/internal/config"
	"cart/internal/delivery"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"database/sql"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.PostgresConnStr())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	cartRepo := repository.NewPostgresCartRepo(db)
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
