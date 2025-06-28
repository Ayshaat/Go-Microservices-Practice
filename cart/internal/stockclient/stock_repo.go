package stockclient

import (
	"cart/internal/models"
	"context"
)

type StockRepository interface {
	GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error)
}
