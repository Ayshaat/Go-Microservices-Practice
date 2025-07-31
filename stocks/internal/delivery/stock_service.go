package delivery

import (
	"context"
	"fmt"
	"strconv"

	"stocks/internal/errors"
	"stocks/internal/log"
	"stocks/internal/usecase"
	stockpb "stocks/pkg/api/stocks"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultPageSize   = 100
	defaultPageNumber = 1
)

type StockServer struct {
	stockpb.UnimplementedStockServiceServer
	usecase usecase.StockUseCase
	logger  log.Logger
}

func NewStockServer(u usecase.StockUseCase, logger log.Logger) stockpb.StockServiceServer {
	return &StockServer{
		usecase: u,
		logger:  logger,
	}
}

func (s *StockServer) AddItem(ctx context.Context, req *stockpb.AddItemRequest) (*stockpb.StockResponse, error) {
	tr := otel.Tracer("stocks-server")
	ctx, span := tr.Start(ctx, "AddItem")
	defer span.End()

	s.logger.Info("AddItem called",
		log.String("sku", req.GetSku()),
		log.Int32("count", req.GetCount()),
		log.String("location", req.GetLocation()),
	)

	item, err := ValidateAddItemRequest(req)
	if err != nil {
		s.logger.Error("Invalid AddItem request", log.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.usecase.Add(ctx, item)
	if err != nil {
		fields := []log.Field{log.Error(err)}

		switch err {
		case errors.ErrInvalidSKU:
			s.logger.Error("AddItem error: invalid SKU", fields...)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.ErrOwnershipViolation:
			s.logger.Error("AddItem error: ownership violation", fields...)
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			s.logger.Error("AddItem error: internal", fields...)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	s.logger.Info("Item added successfully",
		log.String("sku", req.GetSku()),
		log.Int32("count", req.GetCount()),
	)

	return &stockpb.StockResponse{Message: "Item added successfully"}, nil
}

func (s *StockServer) DeleteItem(ctx context.Context, req *stockpb.DeleteItemRequest) (*stockpb.StockResponse, error) {
	tr := otel.Tracer("stocks-server")
	ctx, span := tr.Start(ctx, "DeleteItem")
	defer span.End()

	s.logger.Info("DeleteItem called", log.String("sku", req.GetSku()))

	sku, err := ParseSKU(req.GetSku())
	if err != nil {
		s.logger.Error("DeleteItem invalid SKU", log.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.usecase.Delete(ctx, sku)
	if err != nil {
		s.logger.Error("DeleteItem failed", log.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete item")
	}

	s.logger.Info("Item deleted successfully", log.String("sku", req.GetSku()))

	return &stockpb.StockResponse{Message: "Item deleted successfully"}, nil
}

func (s *StockServer) GetItem(ctx context.Context, req *stockpb.GetItemRequest) (*stockpb.StockItem, error) {
	tr := otel.Tracer("stocks-server")
	ctx, span := tr.Start(ctx, "GetItem")
	defer span.End()

	s.logger.Info("GetItem called", log.String("sku", req.GetSku()))

	sku, err := ParseSKU(req.GetSku())
	if err != nil {
		s.logger.Error("GetItem invalid SKU", log.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	item, err := s.usecase.GetBySKU(ctx, sku)
	if err != nil {
		s.logger.Error("GetItem failed to get item", log.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.logger.Info("GetItem succeeded",
		log.String("sku", req.GetSku()),
		log.String("location", item.Location),
		log.Int("count", int(item.Count)),
		log.Float32("price", float32(item.Price)),
	)

	return &stockpb.StockItem{
		Sku:      strconv.FormatUint(uint64(item.SKU), 10),
		Location: item.Location,
		Count:    int32(item.Count),
		Price:    float32(item.Price),
	}, nil
}

func (s *StockServer) ListByLocation(ctx context.Context, req *stockpb.ListByLocationRequest) (*stockpb.ListByLocationResponse, error) {
	tr := otel.Tracer("stocks-server")
	ctx, span := tr.Start(ctx, "ListByLocation")
	defer span.End()

	s.logger.Info("ListByLocation called", log.String("location", req.GetLocation()))

	if req.GetLocation() == "" {
		err := fmt.Errorf("location must be non-empty")
		s.logger.Error("ListByLocation invalid location", log.Error(err))
		return nil, status.Error(codes.InvalidArgument, "location must be non-empty")
	}

	items, err := s.usecase.ListByLocation(ctx, req.GetLocation(), defaultPageSize, defaultPageNumber)
	if err != nil {
		s.logger.Error("ListByLocation failed", log.Error(err))
		return nil, status.Error(codes.Internal, "failed to list items")
	}

	s.logger.Info("ListByLocation succeeded", log.String("location", req.GetLocation()), log.Int("items_count", len(items)))

	return &stockpb.ListByLocationResponse{
		Location: req.GetLocation(),
		Items:    ToProtoList(items),
	}, nil
}
