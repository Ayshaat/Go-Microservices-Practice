package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cart/internal/config"
	"cart/internal/log"
	"cart/internal/log/zap"
	"cart/internal/metrics"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartpb "cart/pkg/api/cart"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

func NewGatewayMux(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()
	err := cartpb.RegisterCartServiceHandlerFromEndpoint(ctx, mux, cfg.GRPCEndpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	return mux, nil
}

func StartGatewayServer(ctx context.Context, cfg *config.Config, logger *zap.Logger, m *metrics.Metrics) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := cartpb.RegisterCartServiceHandlerFromEndpoint(context.Background(), mux, cfg.GRPCPort, opts)
	if err != nil {
		logger.Error("failed to register gRPC-Gateway endpoint",
			log.String("error", err.Error()),
		)

		return fmt.Errorf("failed to register gateway: %w", err)
	}

	logger.Info("gRPC-Gateway HTTP server listening",
		log.String("http_port", cfg.HTTPPort),
		log.String("grpc_endpoint", cfg.GRPCEndpoint),
	)

	handler := metrics.MetricsMiddleware(m, logger)(mux)
	handler = loggingMiddleware(logger, handler)

	handler = otelhttp.NewHandler(handler, "grpc-gateway")

	srv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down gRPC-Gateway HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("gRPC-Gateway server error", log.Error(err))
			return fmt.Errorf("gRPC-Gateway server failed: %w", err)
		}
	}

	return nil
}
