package repository

import (
	"cart/internal/models"
	"context"
)

//go:generate mockgen -source=repo.go -destination=../../usecase/mocks/cart_repository_mock.go -package=mocks

type CartRepository interface {
	Add(ctx context.Context, item models.CartItem) error
	Delete(ctx context.Context, userID int64, sku uint32) error
	List(ctx context.Context, userID int64) ([]models.CartItem, error)
	Clear(ctx context.Context, userID int64) error
	Upsert(ctx context.Context, item models.CartItem) error
}
