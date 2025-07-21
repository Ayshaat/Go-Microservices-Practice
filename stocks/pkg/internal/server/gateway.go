package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcEndpoint = "localhost:9090"
	httpPort     = ":8080"
)

func StartGatewayServer(stockUC usecase.StockUseCase) error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	mux := runtime.NewServeMux()

	err := stockpb.RegisterStockServiceHandlerFromEndpoint(context.Background(), mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	log.Printf("gRPC-Gateway HTTP server listening on %s", httpPort)

	return http.ListenAndServe(httpPort, mux)
}
