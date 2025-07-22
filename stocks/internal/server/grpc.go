package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"stocks/internal/config"
	service "stocks/internal/service"
	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"

	"google.golang.org/grpc"
)

func StartGRPCServer(ctx context.Context, cfg *config.Config, stockUC usecase.StockUseCase) error {
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()

	stockpb.RegisterStockServiceServer(grpcServer, service.NewStockServer(stockUC))

	log.Printf("gRPC server is running on %s...", cfg.GRPCPort)

	errCh := make(chan error, 1)

	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down gRPC server gracefully...")
		grpcServer.GracefulStop()

		return nil
	case err := <-errCh:
		return fmt.Errorf("gRPC server failed: %w", err)
	}
}
