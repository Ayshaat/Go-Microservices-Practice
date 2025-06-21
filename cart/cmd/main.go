package main

import (
	"cart/internal/config"
	"cart/internal/delivery"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	stockClient := stockclient.New(os.Getenv("STOCK_SERVICE_URL"))
	cartUseCase := usecase.NewCartUsecase(cartRepo, stockClient)
	handler := delivery.NewHandler(cartUseCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + os.Getenv("HTTP_PORT"),
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting cart server on port", os.Getenv("HTTP_PORT"))

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down cart server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.WriteTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
