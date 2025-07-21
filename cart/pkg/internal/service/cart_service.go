package service

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"cart/internal/models"
	"cart/internal/usecase"
	cartpb "cart/pkg/api"
)

type cartServer struct {
	cartpb.UnimplementedCartServiceServer
	useCase usecase.CartUseCase
}

func NewCartServer(useCase usecase.CartUseCase) cartpb.CartServiceServer {
	return &cartServer{useCase: useCase}
}

func (s *cartServer) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*cartpb.CartResponse, error) {
	log.Printf("AddItem called: %+v", req)

	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	skuUint64, err := strconv.ParseUint(req.Sku, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid sku: %w", err)
	}
	sku := uint32(skuUint64)

	item := models.CartItem{
		UserID: userID,
		SKU:    sku,
		Count:  int16(req.Count),
	}

	err = s.useCase.Add(ctx, item)
	if err != nil {
		return nil, err
	}

	return &cartpb.CartResponse{Message: "Item added successfully"}, nil
}

func (s *cartServer) DeleteItem(ctx context.Context, req *cartpb.DeleteItemRequest) (*cartpb.CartResponse, error) {
	log.Printf("DeleteItem called: %+v", req)

	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	skuUint64, err := strconv.ParseUint(req.Sku, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid sku: %w", err)
	}
	sku := uint32(skuUint64)

	err = s.useCase.Delete(ctx, userID, sku)
	if err != nil {
		return nil, err
	}

	return &cartpb.CartResponse{Message: "Item deleted successfully"}, nil
}

func (s *cartServer) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.CartResponse, error) {
	log.Printf("ClearCart called: %+v", req)

	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	err = s.useCase.Clear(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &cartpb.CartResponse{Message: "Cart cleared successfully"}, nil
}

func (s *cartServer) ListCart(ctx context.Context, req *cartpb.ListCartRequest) (*cartpb.ListCartResponse, error) {
	log.Printf("ListCart called: %+v", req)

	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	items, err := s.useCase.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	var cartItems []*cartpb.CartItem
	for _, item := range items {
		cartItems = append(cartItems, &cartpb.CartItem{
			Sku:   strconv.FormatUint(uint64(item.SKU), 10), // convert uint32 -> string
			Count: int32(item.Count),
		})
	}

	return &cartpb.ListCartResponse{
		UserId: req.UserId,
		Items:  cartItems,
	}, nil
}
