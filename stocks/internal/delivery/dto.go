package delivery

type StockItemDTO struct {
	SKU      uint32 `json:"sku"`
	Price    uint32 `json:"price"`
	Count    uint16 `json:"count"`
	Location string `json:"location"`
}
