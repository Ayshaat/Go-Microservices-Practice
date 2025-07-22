package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"cart/internal/config"
	"cart/internal/usecase"

	"google.golang.org/grpc"

	service "cart/internal/service"
	cartpb "cart/pkg/api/cart"
)

func StartGRPCServer(ctx context.Context, cfg *config.Config, cartUC usecase.CartUseCase) error {
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer()

	cartpb.RegisterCartServiceServer(grpcServer, service.NewCartServer(cartUC))

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
