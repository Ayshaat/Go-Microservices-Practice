package delivery

import (
	"context"
	"fmt"
	"strconv"

	"cart/internal/log"
	"cart/internal/usecase"
	cartpb "cart/pkg/api/cart"

	"go.opentelemetry.io/otel"
)

type cartServer struct {
	cartpb.UnimplementedCartServiceServer
	useCase usecase.CartUseCase
	logger  log.Logger
}

func NewCartServer(useCase usecase.CartUseCase, logger log.Logger) cartpb.CartServiceServer {
	return &cartServer{
		useCase: useCase,
		logger:  logger,
	}
}

func (s *cartServer) AddItem(ctx context.Context, req *cartpb.AddItemRequest) (*cartpb.CartResponse, error) {
	tr := otel.Tracer("cart-server")
	ctx, span := tr.Start(ctx, "AddItem")
	defer span.End()

	s.logger.Info("AddItem called",
		log.String("user_id", req.UserId),
		log.String("sku", req.Sku),
		log.Int32("count", req.Count),
	)

	item, err := CartItemFromAddRequest(req)
	if err != nil {
		s.logger.Error("Invalid AddItem request", log.Error(err))
		return nil, err
	}

	err = s.useCase.Add(ctx, item)
	if err != nil {
		s.logger.Error("Failed to add item", log.Error(err))
		return nil, err
	}

	s.logger.Info("Item added successfully",
		log.String("user_id", req.UserId),
		log.String("sku", req.Sku),
	)

	return &cartpb.CartResponse{Message: "Item added successfully"}, nil
}

func (s *cartServer) DeleteItem(ctx context.Context, req *cartpb.DeleteItemRequest) (*cartpb.CartResponse, error) {
	tr := otel.Tracer("cart-server")
	ctx, span := tr.Start(ctx, "DeleteItem")
	defer span.End()

	s.logger.Info("DeleteItem called",
		log.String("user_id", req.UserId),
		log.String("sku", req.Sku),
	)

	userID, err := ParseUserID(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user_id in DeleteItem", log.Error(err))
		return nil, err
	}

	sku, err := ParseSKU(req.Sku)
	if err != nil {
		s.logger.Error("Invalid sku in DeleteItem", log.Error(err))
		return nil, err
	}

	err = s.useCase.Delete(ctx, userID, sku)
	if err != nil {
		s.logger.Error("Failed to delete item", log.Error(err))
		return nil, err
	}

	s.logger.Info("Item deleted successfully",
		log.String("user_id", req.UserId),
		log.String("sku", req.Sku),
	)

	return &cartpb.CartResponse{Message: "Item deleted successfully"}, nil
}

func (s *cartServer) ClearCart(ctx context.Context, req *cartpb.ClearCartRequest) (*cartpb.CartResponse, error) {
	tr := otel.Tracer("cart-server")
	ctx, span := tr.Start(ctx, "ClearCart")
	defer span.End()

	s.logger.Info("ClearCart called",
		log.String("user_id", req.UserId),
	)

	userID, err := ParseUserID(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user_id in ClearCart", log.Error(err))
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	err = s.useCase.Clear(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to clear cart", log.Error(err))
		return nil, err
	}

	s.logger.Info("Cart cleared successfully",
		log.String("user_id", req.UserId),
	)

	return &cartpb.CartResponse{Message: "Cart cleared successfully"}, nil
}

func (s *cartServer) ListCart(ctx context.Context, req *cartpb.ListCartRequest) (*cartpb.ListCartResponse, error) {
	tr := otel.Tracer("cart-server")
	ctx, span := tr.Start(ctx, "ListCart")
	defer span.End()

	s.logger.Info("ListCart called",
		log.String("user_id", req.UserId),
	)

	userID, err := ParseUserID(req.UserId)
	if err != nil {
		s.logger.Error("Invalid user_id in ListCart", log.Error(err))
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	items, err := s.useCase.List(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to list cart items", log.Error(err))
		return nil, err
	}

	var cartItems []*cartpb.CartItem
	for _, item := range items {
		cartItems = append(cartItems, &cartpb.CartItem{
			Sku:   strconv.FormatUint(uint64(item.SKU), 10),
			Count: int32(item.Count),
		})
	}

	if cartItems == nil {
		cartItems = []*cartpb.CartItem{}
	}

	s.logger.Info("ListCart succeeded",
		log.String("user_id", req.UserId),
		log.Int("item_count", len(cartItems)),
	)

	return &cartpb.ListCartResponse{
		UserId: req.UserId,
		Items:  cartItems,
	}, nil
}
