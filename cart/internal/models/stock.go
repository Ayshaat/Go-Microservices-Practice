package models

type StockItem struct {
	SKU      uint32
	Name     string
	Location string
	Type     string
	Price    uint32
	Count    int16
}
