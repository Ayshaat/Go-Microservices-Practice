package models

type StockItem struct {
	SKU      uint32 `json:"sku"`
	Name     string `json:"name"`
	Location string `json:"location"`
}
