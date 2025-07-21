package main

import (
	"log"
	"os"
	"sync"

	"cart/internal/db"
	"cart/internal/kafka"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"cart/internal/usecase"
	server "cart/pkg/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, continuing with environment variables")
	}

	connStr :=
		"host=" + os.Getenv("DB_HOST") +
			" port=" + os.Getenv("DB_PORT") +
			" user=" + os.Getenv("DB_USER") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" sslmode=disable"

	dbConn, err := db.ConnectDB(connStr)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	repo := repository.NewPostgresCartRepo(dbConn)

	stockServiceURL := os.Getenv("STOCK_SERVICE_URL")

	stockRepo, err := stockclient.New(stockServiceURL)
	if err != nil {
		log.Fatalf("failed to create stock client: %v", err)
	}

	cfg, err := kafka.NewProducerConfigFromEnv()
	if err != nil {
		log.Fatalf("failed to load kafka producer config: %v", err)
	}

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer producer.Close()

	cartUC := usecase.NewCartUsecase(repo, stockRepo, producer)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err := server.StartGRPCServer(cartUC); err != nil {
			log.Fatalf("failed to start gRPC server: %v", err)
		}
	}()

	go func() {
		defer wg.Done()

		if err := server.StartGatewayServer(cartUC); err != nil {
			log.Fatalf("failed to start gRPC-Gateway server: %v", err)
		}
	}()

	wg.Wait()
}
