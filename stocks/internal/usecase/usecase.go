package usecase

import (
	"context"
	"stocks/internal/models"
)

//go:generate mockgen -source=internal/usecase/usecase.go -destination=internal/usecase/mocks/stockusecase_mock.go -package=mocks

type StockUseCase interface {
	Add(ctx context.Context, item models.StockItem) error
	Delete(ctx context.Context, sku uint32) error
	GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error)
	ListByLocation(ctx context.Context, location string, pageSize, currentPage int64) ([]models.StockItem, error)
}
