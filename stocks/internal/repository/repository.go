package repository

import "stocks/internal/models"

type StockRepository interface {
	Add(item models.StockItem) error
	Delete(sku uint32) error
	GetBySKU(sku uint32) (models.StockItem, error)
	ListByLocation(location string, pageSize, currentPage int64) ([]models.StockItem, error)
}
