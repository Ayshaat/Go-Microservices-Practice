package delivery

import "cart/internal/models"

type CartItemDTO struct {
	UserID int64  `json:"item_id"`
	SKU    uint32 `json:"name"`
	Count  int16  `json:"count"`
	Price  uint32 `json:"price"`
	Stock  int16  `json:"stock"`
}

func (dto CartItemDTO) ToModel() models.CartItem {
	return models.CartItem{
		UserID: dto.UserID,
		SKU:    dto.SKU,
		Count:  dto.Count,
		Price:  dto.Price,
		Stock:  dto.Stock,
	}
}

type StockItemDTO struct {
	SKU      uint32 `json:"sku"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Price    uint32 `json:"price"`
	Count    int16  `json:"count"`
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
