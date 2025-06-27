package repository

import (
	"context"
	"stocks/internal/models"
)

type StockRepository interface {
	Delete(ctx context.Context, sku uint32) error
	GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error)
	ListByLocation(ctx context.Context, location string, pageSize, currentPage int64) ([]models.StockItem, error)
	GetSKUInfo(ctx context.Context, sku uint32) (string, string, error)
	GetByUserSKU(ctx context.Context, userID int64, sku uint32) (models.StockItem, error)
	InsertStockItem(ctx context.Context, item models.StockItem) error
	UpdateCount(ctx context.Context, userID int64, sku uint32, newCount uint16) error
}
