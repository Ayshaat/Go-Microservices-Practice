package repository

import (
	"cart/internal/models"
	"errors"
	"sync"
)

const expectedCapacity = 10

var (
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrCartItemExists   = errors.New("cart item already exists")
)

type InMemoryCartRepo struct {
	mu    sync.RWMutex
	items map[int64]map[uint32]models.CartItem
}

func NewInMemoryCartRepo() *InMemoryCartRepo {
	return &InMemoryCartRepo{
		items: make(map[int64]map[uint32]models.CartItem),
	}
}

func (r *InMemoryCartRepo) Add(item models.CartItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[item.UserID]; !ok {
		r.items[item.UserID] = make(map[uint32]models.CartItem)
	}

	if _, exists := r.items[item.UserID][item.SKU]; exists {
		return ErrCartItemExists
	}

	r.items[item.UserID][item.SKU] = item

	return nil
}

func (r *InMemoryCartRepo) Delete(userID int64, sku uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[userID]; !ok {
		return ErrCartItemNotFound
	}

	if _, exists := r.items[userID][sku]; !exists {
		return ErrCartItemNotFound
	}

	delete(r.items[userID], sku)

	if len(r.items[userID]) == 0 {
		delete(r.items, userID)
	}

	return nil
}

func (r *InMemoryCartRepo) Get(userID int64, sku uint32) (models.CartItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.items[userID]; !ok {
		return models.CartItem{}, ErrCartItemNotFound
	}

	item, exists := r.items[userID][sku]
	if !exists {
		return models.CartItem{}, ErrCartItemNotFound
	}

	return item, nil
}

func (r *InMemoryCartRepo) List(userID int64) ([]models.CartItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userItems, ok := r.items[userID]
	if !ok {
		return []models.CartItem{}, nil
	}

	result := make([]models.CartItem, 0, expectedCapacity)
	for _, item := range userItems {
		result = append(result, item)
	}

	return result, nil
}

func (r *InMemoryCartRepo) Clear(userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[userID]; !ok {
		return ErrCartItemNotFound
	}

	delete(r.items, userID)

	return nil
}
