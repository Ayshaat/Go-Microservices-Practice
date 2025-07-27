package app

import (
	"cart/internal/config"
	"cart/internal/db"
	"cart/internal/kafka"
	"cart/internal/repository"
	"cart/internal/server"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
)

const serverCount = 3

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	log.Println(cfg.GRPCPort)
	log.Println(cfg.HTTPPort)
	database, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	cartRepo := repository.NewPostgresCartRepo(database)

	stockClient, err := stockclient.NewGRPCClient(os.Getenv("STOCK_SERVICE_URL"))
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

	errCh := make(chan error, serverCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(serverCount)

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC server on port", cfg.GRPCPort)

		if err := server.StartGRPCServer(ctx, cfg, cartUseCase); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC-Gateway server on port", cfg.HTTPPort)

		if err := server.StartGatewayServer(ctx, cfg); err != nil {
			errCh <- fmt.Errorf("gateway server failed: %w", err)
		}
	}()

	select {
	case <-stop:
		log.Println("Shutdown signal received")
	case err := <-errCh:
		return err
	}

	log.Println("Shutting down servers...")

	log.Println("Server gracefully stopped")

	return nil
}
