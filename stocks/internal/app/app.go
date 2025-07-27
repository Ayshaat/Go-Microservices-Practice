package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/kafka"
	"stocks/internal/repository"
	"stocks/internal/server"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

const serverCount = 3

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.ConnectDB(cfg.PostgresConnStr(), "internal/db/migrations")
	if err != nil {
		return fmt.Errorf("failed to connect and migrate db: %w", err)
	}
	defer database.Close()

	dbx := sqlx.NewDb(database, "postgres")

	defer dbx.Close()

	txFactory := trmsqlx.NewDefaultFactory(dbx.DB)
	txManager := manager.Must(txFactory)
	txCtxGetter := trmsqlx.DefaultCtxGetter

	repo := repository.NewPostgresStockRepo(dbx, txCtxGetter)

	producerConfig, err := kafka.NewProducerConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create kafka producer config: %w", err)
	}

	producer, err := kafka.NewProducer(producerConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("failed to close kafka producer: %v", err)
		}
	}()

	useCase := usecase.NewStockUsecase(repo, txManager, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, serverCount)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(serverCount)

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC server on port", cfg.GRPCPort)

		if err := server.StartGRPCServer(ctx, cfg, useCase); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC-Gateway server on port", cfg.GatewayPort)

		if err := server.StartGatewayServer(ctx, cfg); err != nil {
			errCh <- fmt.Errorf("gRPC-Gateway server failed: %w", err)
		}
	}()

	select {
	case sig := <-stop:
		log.Printf("Shutdown signal received: %v", sig)
	case err := <-errCh:
		return err
	}

	wg.Wait()
	log.Println("Stocks server gracefully stopped")

	return nil
}
