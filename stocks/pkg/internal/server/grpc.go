package server

import (
	"fmt"
	"log"
	"net"

	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"
	service "stocks/pkg/internal/service"

	"google.golang.org/grpc"
)

const grpcPort = ":9090"

func StartGRPCServer(stockUC usecase.StockUseCase) error {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	stockpb.RegisterStockServiceServer(grpcServer, service.NewStockServer(stockUC))

	log.Printf("gRPC server is running on %s...", grpcPort)

	return grpcServer.Serve(lis)
}
