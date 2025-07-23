package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"cart/internal/config"
	"cart/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

func StartGatewayServer(ctx context.Context, cfg *config.Config, cartUC usecase.CartUseCase) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := cartpb.RegisterCartServiceHandlerFromEndpoint(context.Background(), mux, cfg.GRPCEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	log.Printf("gRPC-Gateway HTTP server listening on %s", cfg.HTTPPort)

	srv := &http.Server{
		Addr:         cfg.HTTPPort,
		Handler:      mux,
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
		log.Println("Shutting down gRPC-Gateway HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("gRPC-Gateway server failed: %w", err)
		}
	}

	return nil
}
