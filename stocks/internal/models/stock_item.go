package models

type StockItem struct {
	SKU      uint32
	Name     string
	Type     string
	Price    uint32
	Count    uint16
	Location string
}
