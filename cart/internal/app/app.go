package app

import (
	"cart/internal/config"
	"cart/internal/db"
	"cart/internal/kafka"
	"cart/internal/log"
	"cart/internal/log/zap"
	"cart/internal/metrics"
	"cart/internal/repository"
	"cart/internal/server"
	"cart/internal/stockclient"
	"cart/internal/trace"
	"cart/internal/usecase"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
)

const serverCount = 4

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	shutdownTracer, err := trace.InitTracer("cart", cfg.JaegerEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}
	defer shutdownTracer(context.Background())

	logger, cleanup, err := zap.NewLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer cleanup()

	logger.Info("Config loaded",
		log.String("grpc_port", cfg.GRPCPort),
		log.String("http_port", cfg.HTTPPort),
	)

	database, err := db.ConnectDB(cfg.PostgresConnStr())
	if err != nil {
		logger.Errorf("failed to connect to database: %v", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	cartRepo := repository.NewPostgresCartRepo(database)

	metricsInstance := metrics.RegisterMetrics()

	stockClient, err := stockclient.NewGRPCClient(os.Getenv("STOCK_SERVICE_URL"), logger, metricsInstance)
	if err != nil {
		logger.Errorf("failed to create stock client: %v", err)
		return fmt.Errorf("failed to create stock client: %w", err)
	}

	producerConfig, err := kafka.NewProducerConfigFromEnv()
	if err != nil {
		logger.Errorf("failed to create kafka producer config: %v", err)
		return fmt.Errorf("failed to create kafka producer config: %w", err)
	}

	producer, err := kafka.NewProducer(producerConfig, logger)
	if err != nil {
		logger.Errorf("failed to create kafka producer: %v", err)
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}
	defer producer.Close()

	cartUseCase := usecase.NewCartUsecase(cartRepo, stockClient, producer, logger)

	errCh := make(chan error, serverCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(serverCount)

	go func() {
		defer wg.Done()
		logger.Info("Starting Prometheus metrics server on :9090")

		metrics.StartMetricsServer(":9090")
	}()

	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC server", log.String("port", cfg.GRPCPort))

		if err := server.StartGRPCServer(ctx, cfg, cartUseCase, logger, metricsInstance); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC-Gateway server", log.String("port", cfg.HTTPPort))

		if err := server.StartGatewayServer(ctx, cfg, logger, metricsInstance); err != nil {
			errCh <- fmt.Errorf("gateway server failed: %w", err)
		}
	}()

	select {
	case <-stop:
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		return err
	}

	logger.Info("Shutting down servers...")
	cancel()
	wg.Wait()

	logger.Info("Server gracefully stopped")

	return nil
}
