package repository

import "cart/internal/models"

type CartItemRow struct {
	UserID int64
	SKU    uint32
	Count  int16
}

func (r *CartItemRow) ToDomain() models.CartItem {
	return models.CartItem{
		UserID: r.UserID,
		SKU:    r.SKU,
		Count:  r.Count,
	}
}
