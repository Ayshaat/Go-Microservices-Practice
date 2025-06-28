package models

type StockItem struct {
	SKU      uint32
	Name     string
	Location string
	Type     string
	Price    float64
	Count    int16
}
