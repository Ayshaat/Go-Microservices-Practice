package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"stocks/internal/config"
	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	readTimeout       = 5 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 120 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func StartGatewayServer(ctx context.Context, cfg *config.Config, stockUC usecase.StockUseCase) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := stockpb.RegisterStockServiceHandlerFromEndpoint(context.Background(), mux, cfg.GRPCPort, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.GatewayPort,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("gRPC-Gateway HTTP server listening on %s", cfg.GatewayPort)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down gRPC-Gateway server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), readTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("gRPC-Gateway server failed: %w", err)
	}
}
