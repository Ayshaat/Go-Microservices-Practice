package server

import (
	"fmt"
	"log"
	"net"

	"cart/internal/usecase"

	"google.golang.org/grpc"

	cartpb "cart/pkg/api"
	service "cart/pkg/internal/service"
)

const grpcPort = ":9090"

func StartGRPCServer(cartUC usecase.CartUseCase) error {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	cartpb.RegisterCartServiceServer(grpcServer, service.NewCartServer(cartUC))

	log.Printf("gRPC server is running on %s...", grpcPort)

	return grpcServer.Serve(lis)
}
