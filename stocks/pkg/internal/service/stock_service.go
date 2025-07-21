package server

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"stocks/internal/models"
	"stocks/internal/usecase"
	stockpb "stocks/pkg/api"
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
	if req.GetSku() == "" || req.GetLocation() == "" || req.GetCount() <= 0 {
		return nil, fmt.Errorf("invalid input: sku, location must be non-empty and count > 0")
	}

	skuUint64, err := strconv.ParseUint(req.GetSku(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid sku format: %w", err)
	}
	sku := uint32(skuUint64)

	item := models.StockItem{
		SKU:      sku,
		Location: req.GetLocation(),
		Count:    uint16(req.GetCount()),
	}

	err = s.usecase.Add(ctx, item)
	if err != nil {
		log.Printf("AddItem error: %v", err)
		return nil, err
	}

	return &stockpb.StockResponse{Message: "item added"}, nil
}

func (s *StockServer) DeleteItem(ctx context.Context, req *stockpb.DeleteItemRequest) (*stockpb.StockResponse, error) {
	if req.GetSku() == "" {
		return nil, fmt.Errorf("invalid input: sku must be non-empty")
	}

	skuUint64, err := strconv.ParseUint(req.GetSku(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid sku format: %w", err)
	}
	sku := uint32(skuUint64)

	err = s.usecase.Delete(ctx, sku)
	if err != nil {
		log.Printf("DeleteItem error: %v", err)
		return nil, err
	}

	return &stockpb.StockResponse{Message: "item deleted"}, nil
}

func (s *StockServer) GetItem(ctx context.Context, req *stockpb.GetItemRequest) (*stockpb.StockItem, error) {
	if req.GetSku() == "" {
		return nil, fmt.Errorf("invalid input: sku must be non-empty")
	}

	skuUint64, err := strconv.ParseUint(req.GetSku(), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid sku format: %w", err)
	}
	sku := uint32(skuUint64)

	item, err := s.usecase.GetBySKU(ctx, sku)
	if err != nil {
		log.Printf("GetItem error: %v", err)
		return nil, err
	}

	return &stockpb.StockItem{
		Sku:      strconv.FormatUint(uint64(item.SKU), 10),
		Location: item.Location,
		Count:    int32(item.Count),
	}, nil
}

func (s *StockServer) ListByLocation(ctx context.Context, req *stockpb.ListByLocationRequest) (*stockpb.ListByLocationResponse, error) {
	if req.GetLocation() == "" {
		return nil, fmt.Errorf("invalid input: location must be non-empty")
	}

	const (
		defaultPageSize   = 100
		defaultPageNumber = 1
	)

	items, err := s.usecase.ListByLocation(ctx, req.GetLocation(), defaultPageSize, defaultPageNumber)
	if err != nil {
		log.Printf("ListByLocation error: %v", err)
		return nil, err
	}

	pbItems := make([]*stockpb.StockItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &stockpb.StockItem{
			Sku:      strconv.FormatUint(uint64(item.SKU), 10),
			Location: item.Location,
			Count:    int32(item.Count),
		})
	}

	return &stockpb.ListByLocationResponse{
		Location: req.GetLocation(),
		Items:    pbItems,
	}, nil
}
