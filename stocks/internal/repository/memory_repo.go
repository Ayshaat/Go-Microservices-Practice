package repository

import (
	"errors"
	"stocks/internal/models"
	"sync"
)

var (
	ErrItemNotFound = errors.New("stock item not found")
	ErrItemExists   = errors.New("stock item already exists")
)

type InMemoryStockRepo struct {
	mu    sync.RWMutex
	items map[uint32]models.StockItem
}

func NewInMemoryStockRepo() *InMemoryStockRepo {
	return &InMemoryStockRepo{
		items: make(map[uint32]models.StockItem),
	}
}

func (r *InMemoryStockRepo) Add(item models.StockItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[item.SKU]; exists {
		return ErrItemExists
	}
	r.items[item.SKU] = item

	return nil
}

func (r *InMemoryStockRepo) Delete(sku uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[sku]; !exists {
		return ErrItemNotFound
	}

	delete(r.items, sku)

	return nil
}

func (r *InMemoryStockRepo) GetBySKU(sku uint32) (models.StockItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[sku]
	if !exists {
		return models.StockItem{}, ErrItemNotFound
	}

	return item, nil
}

func (r *InMemoryStockRepo) ListByLocation(location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []models.StockItem

	for _, item := range r.items {
		if item.Location == location {
			filtered = append(filtered, item)
		}
	}

	start := (currentPage - 1) * pageSize
	end := start + pageSize

	if start >= int64(len(filtered)) {
		return []models.StockItem{}, nil
	}

	if end > int64(len(filtered)) {
		end = int64(len(filtered))
	}

	return filtered[start:end], nil
}
