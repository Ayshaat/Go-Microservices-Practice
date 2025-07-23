package delivery

import (
	"context"
	"log"
	"strconv"

	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"

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
}

func NewStockServer(u usecase.StockUseCase) stockpb.StockServiceServer {
	return &StockServer{
		usecase: u,
	}
}

func (s *StockServer) AddItem(ctx context.Context, req *stockpb.AddItemRequest) (*stockpb.StockResponse, error) {
	item, err := ValidateAddItemRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.usecase.Add(ctx, item)
	if err != nil {
		log.Printf("AddItem error: %v", err)
		return nil, status.Error(codes.Internal, "failed to add item")
	}

	return &stockpb.StockResponse{Message: "item added"}, nil
}

func (s *StockServer) DeleteItem(ctx context.Context, req *stockpb.DeleteItemRequest) (*stockpb.StockResponse, error) {
	sku, err := ParseSKU(req.GetSku())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.usecase.Delete(ctx, sku)
	if err != nil {
		log.Printf("DeleteItem error: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete item")
	}

	return &stockpb.StockResponse{Message: "item deleted"}, nil
}

func (s *StockServer) GetItem(ctx context.Context, req *stockpb.GetItemRequest) (*stockpb.StockItem, error) {
	sku, err := ParseSKU(req.GetSku())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	item, err := s.usecase.GetBySKU(ctx, sku)
	if err != nil {
		log.Printf("GetItem error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &stockpb.StockItem{
		Sku:      strconv.FormatUint(uint64(item.SKU), 10),
		Location: item.Location,
		Count:    int32(item.Count),
	}, nil
}

func (s *StockServer) ListByLocation(ctx context.Context, req *stockpb.ListByLocationRequest) (*stockpb.ListByLocationResponse, error) {
	if req.GetLocation() == "" {
		return nil, status.Error(codes.InvalidArgument, "location must be non-empty")
	}

	items, err := s.usecase.ListByLocation(ctx, req.GetLocation(), defaultPageSize, defaultPageNumber)
	if err != nil {
		log.Printf("ListByLocation error: %v", err)
		return nil, status.Error(codes.Internal, "failed to list items")
	}
	return &stockpb.ListByLocationResponse{
		Location: req.GetLocation(),
		Items:    ToProtoList(items),
	}, nil
}
