package models

type CartItem struct {
	UserID int64
	SKU    uint32
	Count  int16
	Price  uint32
	Stock  int16
}
