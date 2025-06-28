package delivery

import "stocks/internal/models"

type StockItemDTO struct {
	UserID   int64   `json:"userId"`
	SKU      uint32  `json:"sku"`
	Price    float64 `json:"price"`
	Count    uint16  `json:"count"`
	Location string  `json:"location"`
}

func (dto *StockItemDTO) ToModel() models.StockItem {
	return models.StockItem{
		UserID:   dto.UserID,
		SKU:      dto.SKU,
		Price:    dto.Price,
		Count:    dto.Count,
		Location: dto.Location,
	}
}
