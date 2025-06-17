package usecase

import (
	"cart/internal/models"
	"context"
)

type CartUseCase interface {
	Add(ctx context.Context, item models.CartItem) error
	Delete(userID int64, sku uint32) error
	List(ctx context.Context, userID int64) ([]models.CartItem, error)
	Clear(userID int64) error
}
