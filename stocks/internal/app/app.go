package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"

	"stocks/internal/config"
	"stocks/internal/db"
	"stocks/internal/kafka"
	"stocks/internal/log"
	"stocks/internal/log/zap"
	"stocks/internal/metrics"
	"stocks/internal/repository"
	"stocks/internal/server"
	"stocks/internal/trace"
	"stocks/internal/usecase"

	"github.com/jmoiron/sqlx"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sql/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

const serverCount = 4

func Run(envFile string) error {
	cfg, err := config.Load(envFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	shutdownTracer, err := trace.InitTracer("stocks", cfg.JaegerEndpoint)
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
		log.String("gateway_port", cfg.GatewayPort),
	)

	database, err := db.ConnectDB(cfg.PostgresConnStr(), "internal/db/migrations")
	if err != nil {
		logger.Error("failed to connect and migrate db", log.Error(err))
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
		logger.Error("failed to create kafka producer config", log.Error(err))
		return fmt.Errorf("failed to create kafka producer config: %w", err)
	}

	producer, err := kafka.NewProducer(producerConfig, logger)
	if err != nil {
		logger.Error("failed to create kafka producer", log.Error(err))
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			logger.Error("failed to close kafka producer: %v", log.Error(err))
		}
	}()

	useCase := usecase.NewStockUsecase(repo, txManager, producer, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsInstance := metrics.RegisterMetrics()

	errCh := make(chan error, serverCount)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(serverCount)

	go func() {
		defer wg.Done()
		logger.Info("Starting Prometheus metrics server on :9000")

		metrics.StartMetricsServer(":9000")
	}()

	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC server", log.String("port", cfg.GRPCPort))

		if err := server.StartGRPCServer(ctx, cfg, useCase, logger, metricsInstance); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC-Gateway server", log.String("port", cfg.GatewayPort))

		if err := server.StartGatewayServer(ctx, cfg, logger, metricsInstance); err != nil {
			errCh <- fmt.Errorf("gRPC-Gateway server failed: %w", err)
		}
	}()

	select {
	case sig := <-stop:
		logger.Info("Shutdown signal received", log.String("signal", sig.String()))
	case err := <-errCh:
		logger.Error("Server error received", log.Error(err))
		return err
	}

	wg.Wait()
	logger.Info("Stocks server gracefully stopped")

	return nil
}
