package main

import (
	"cart/internal/config"
	"cart/internal/db"
	"cart/internal/delivery"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	cartRepo := repository.NewPostgresCartRepo(database)

	stockClient, err := stockclient.New(os.Getenv("STOCK_SERVICE_URL"))
	if err != nil {
		log.Fatalf("failed to create stock client: %v", err)
	}
	cartUseCase := usecase.NewCartUsecase(cartRepo, stockClient)
	handler := delivery.NewHandler(cartUseCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8090"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting cart server on port", port)

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
