package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cart/internal/usecase"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cartpb "cart/pkg/api"
)

const (
	grpcEndpoint = "localhost:9090"
	httpPort     = ":9091"
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

func StartGatewayServer(cartUC usecase.CartUseCase) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := cartpb.RegisterCartServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	log.Printf("gRPC-Gateway HTTP server listening on %s", httpPort)

	srv := &http.Server{
		Addr:         httpPort,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return srv.ListenAndServe()
}
