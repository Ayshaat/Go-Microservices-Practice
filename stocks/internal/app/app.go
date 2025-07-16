package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/lib/pq"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/delivery"
	"stocks/internal/kafka"
	"stocks/internal/repository"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

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

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return fmt.Errorf("KAFKA_BROKERS env var is not set")
	}
	brokers := strings.Split(brokersEnv, ",")

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "stocks"
	}

	partition := int32(0)

	producerConfig := kafka.ProducerConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: partition,
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
	handler := delivery.NewHandler(useCase)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
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
		log.Println("Starting server on port", port)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("stocks server failed: %v", err)
		}
	}()

	<-stop

	log.Println("Shutting down stocks server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.WriteTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("stocks server shutdown failed: %w", err)
	}

	log.Println("Stocks server gracefully stopped")

	return nil
}
