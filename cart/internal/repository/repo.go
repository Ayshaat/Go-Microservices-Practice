package repository

import "cart/internal/models"

type CartRepository interface {
	Add(item models.CartItem) error
	Delete(userID int64, sku uint32) error
	List(userID int64) ([]models.CartItem, error)
	Clear(userID int64) error
}
