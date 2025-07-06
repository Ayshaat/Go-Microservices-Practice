package usecase

import (
	"cart/internal/models"
	"context"
)

//go:generate mockgen -source=usecase.go -destination=mocks/cart_usecase_mock.go -package=mocks

type CartUseCase interface {
	Add(ctx context.Context, item models.CartItem) error
	Delete(ctx context.Context, userID int64, sku uint32) error
	List(ctx context.Context, userID int64) ([]models.CartItem, error)
	Clear(ctx context.Context, userID int64) error
}
