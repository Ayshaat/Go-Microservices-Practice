package models

type CartItem struct {
	UserID int64
	SKU    uint32
	Count  int16
	Price  float64
	Stock  int16
}
