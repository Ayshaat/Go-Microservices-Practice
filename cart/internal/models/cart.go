package models

type CartItem struct {
	UserID int64  `json:"userID"`
	SKU    uint32 `json:"sku"`
	Count  uint16 `json:"count"`
}
