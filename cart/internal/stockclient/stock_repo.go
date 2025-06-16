package stockclient

import "cart/internal/models"

type StockRepository interface {
	GetBySKU(sku uint32) (models.StockItem, error)
}
