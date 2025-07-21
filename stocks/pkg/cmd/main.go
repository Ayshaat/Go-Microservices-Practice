package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/kafka"
	"stocks/internal/repository"
	"stocks/internal/usecase"
	server "stocks/pkg/internal/server"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with system env vars")
	}

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbConn, err := db.ConnectDB(cfg.PostgresConnStr(), "internal/db/migrations")
	if err != nil {
		log.Fatalf("failed to connect and migrate db: %v", err)
	}
	defer dbConn.Close()

	dbx := sqlx.NewDb(dbConn, "postgres")
	defer dbx.Close()

	txFactory := trmsqlx.NewDefaultFactory(dbx.DB)
	txManager := manager.Must(txFactory)
	txCtxGetter := trmsqlx.DefaultCtxGetter

	repo := repository.NewPostgresStockRepo(dbx, txCtxGetter)

	kafkaCfg, err := kafka.NewProducerConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to load kafka config: %v", err)
	}

	producer, err := kafka.NewProducer(kafkaCfg)
	if err != nil {
		log.Fatalf("failed to create Kafka producer: %v", err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("failed to close kafka producer: %v", err)
		}
	}()

	stockUC := usecase.NewStockUsecase(repo, txManager, producer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC server...")

		if err := server.StartGRPCServer(stockUC); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		log.Println("Starting gRPC-Gateway server...")

		if err := server.StartGatewayServer(stockUC); err != nil {
			log.Fatalf("gRPC-Gateway server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutdown signal received, stopping servers...")

	wg.Wait()

	log.Println("Servers stopped gracefully, exiting.")
}
