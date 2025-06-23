package repository

import "stocks/internal/models"

type StockItemRow struct {
	SKU      uint32
	Name     string
	Type     string
	Price    float64
	Count    uint16
	Location string
}

func (r *StockItemRow) ToDomain() models.StockItem {
	return models.StockItem{
		SKU:      r.SKU,
		Name:     r.Name,
		Type:     r.Type,
		Price:    r.Price,
		Count:    r.Count,
		Location: r.Location,
	}
}
