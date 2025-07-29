package server

import (
	"context"
	"fmt"
	"net"

	"stocks/internal/config"
	service "stocks/internal/delivery"
	"stocks/internal/log"
	"stocks/internal/log/zap"
	"stocks/internal/metrics"
	"stocks/internal/usecase"
	stockpb "stocks/pkg/api/stocks"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

func LoggingInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		fields := []log.Field{
			log.String("method", info.FullMethod),
		}

		if p, ok := peer.FromContext(ctx); ok {
			fields = append(fields, log.String("peer", p.Addr.String()))
		}

		traceID := traceIDFromCtx(ctx)
		if traceID != "" {
			fields = append(fields, log.String("trace_id", traceID))
		}

		if err != nil {
			fields = append(fields, log.Error(err))
			logger.Error("gRPC request failed", fields...)
		} else {
			logger.Info("gRPC request handled", fields...)
		}

		return resp, err
	}
}

func traceIDFromCtx(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

func TracingInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("cart-grpc-server")

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		resp, err := handler(ctx, req)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}

		return resp, err
	}
}

func StartGRPCServer(ctx context.Context, cfg *config.Config, stockUC usecase.StockUseCase, logger *zap.Logger, m *metrics.Metrics) error {
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen on port", log.String("port", cfg.GRPCPort), log.Error(err))
		return fmt.Errorf("failed to listen on port %s: %w", cfg.GRPCPort, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(m),
			TracingInterceptor(),
			LoggingInterceptor(logger)),
	)

	stockpb.RegisterStockServiceServer(grpcServer, service.NewStockServer(stockUC, logger))

	reflection.Register(grpcServer)

	logger.Info("gRPC server is running", log.String("port", cfg.GRPCPort))

	errCh := make(chan error, 1)

	go func() {
		errCh <- grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down gRPC server gracefully...")
		grpcServer.GracefulStop()

		return nil
	case err := <-errCh:
		logger.Error("gRPC server failed", log.Error(err))
		return fmt.Errorf("gRPC server failed: %w", err)
	}
}
