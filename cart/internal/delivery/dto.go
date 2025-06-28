package delivery

import "cart/internal/models"

type CartItemDTO struct {
	UserID int64  `json:"userId"`
	SKU    uint32 `json:"sku"`
	Count  int16  `json:"count"`
}

func (dto CartItemDTO) ToModel() models.CartItem {
	return models.CartItem{
		UserID: dto.UserID,
		SKU:    dto.SKU,
		Count:  dto.Count,
	}
}

type StockItemDTO struct {
	SKU      uint32  `json:"sku"`
	Name     string  `json:"name"`
	Location string  `json:"location"`
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Count    int16   `json:"count"`
}

func (dto StockItemDTO) ToModel() models.StockItem {
	return models.StockItem{
		SKU:      dto.SKU,
		Name:     dto.Name,
		Location: dto.Location,
		Type:     dto.Type,
		Price:    dto.Price,
		Count:    dto.Count,
	}
}
