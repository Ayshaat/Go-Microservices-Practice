package models

type StockItem struct {
	UserID   int64
	SKU      uint32
	Name     string
	Type     string
	Price    float64
	Count    uint16
	Location string
}
