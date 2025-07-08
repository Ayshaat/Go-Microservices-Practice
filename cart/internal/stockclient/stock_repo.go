package stockclient

import (
	"cart/internal/models"
	"context"
)

//go:generate mockgen -source=stock_repo.go -destination=../usecase/mocks/stock_repository_mock.go -package=mocks

type StockRepository interface {
	GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error)
}
