package delivery

import "stocks/internal/models"

type StockItemDTO struct {
	SKU      uint32 `json:"sku"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Price    uint32 `json:"price"`
	Count    uint16 `json:"count"`
	Location string `json:"location"`
}

func (dto *StockItemDTO) ToModel() models.StockItem {
	return models.StockItem{
		SKU:      dto.SKU,
		Name:     dto.Name,
		Type:     dto.Type,
		Price:    dto.Price,
		Count:    dto.Count,
		Location: dto.Location,
	}
}

func ToDTO(item models.StockItem) StockItemDTO {
	return StockItemDTO{
		SKU:      item.SKU,
		Name:     item.Name,
		Type:     item.Type,
		Price:    item.Price,
		Count:    item.Count,
		Location: item.Location,
	}
}
