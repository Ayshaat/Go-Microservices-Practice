package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"stocks/internal/log"
	"stocks/internal/metrics"
	"time"

	"stocks/internal/config"
	stockpb "stocks/pkg/api/stocks"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	readTimeout       = 5 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 120 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func NewGatewayMux(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := stockpb.RegisterStockServiceHandlerFromEndpoint(ctx, mux, cfg.GRPCPort, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	return mux, nil
}

func StartGatewayServer(ctx context.Context, cfg *config.Config, logger log.Logger, m metrics.MetricsInterface) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := stockpb.RegisterStockServiceHandlerFromEndpoint(context.Background(), mux, cfg.GRPCPort, opts)
	if err != nil {
		logger.Error("failed to register gRPC-Gateway endpoint", log.Error(err))
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	handler := metrics.MetricsMiddleware(m, logger)(mux)
	handler = loggingMiddleware(logger, handler)

	handler = otelhttp.NewHandler(handler, "grpc-gateway")

	srv := &http.Server{
		Addr:              cfg.GatewayPort,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Info("gRPC-Gateway HTTP server listening", log.String("address", cfg.GatewayPort))

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down gRPC-Gateway server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), readTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		logger.Error("gRPC-Gateway server failed", log.Error(err))
		return fmt.Errorf("gRPC-Gateway server failed: %w", err)
	}
}
