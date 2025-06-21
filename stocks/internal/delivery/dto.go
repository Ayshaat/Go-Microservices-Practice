package delivery

import "stocks/internal/models"

type StockItemDTO struct {
	UserID   int64  `json:"user_id"`
	SKU      uint32 `json:"sku"`
	Price    uint32 `json:"price"`
	Count    uint16 `json:"count"`
	Location string `json:"location"`
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
