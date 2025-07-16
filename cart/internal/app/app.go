package app

import (
	"cart/internal/config"
	"cart/internal/db"
	"cart/internal/delivery"
	"cart/internal/kafka"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	cartRepo := repository.NewPostgresCartRepo(database)

	stockClient, err := stockclient.New(os.Getenv("STOCK_SERVICE_URL"))
	if err != nil {
		return fmt.Errorf("failed to create stock client: %w", err)
	}

	producerConfig, err := kafka.NewProducerConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create kafka producer config: %w", err)
	}

	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}
	defer producer.Close()

	cartUseCase := usecase.NewCartUsecase(cartRepo, stockClient, producer)
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
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Println("Server gracefully stopped")

	return nil
}
