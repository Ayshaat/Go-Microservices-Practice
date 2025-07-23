package delivery

import (
	"errors"
	"fmt"
	"strconv"

	"stocks/internal/models"
	stockpb "stocks/pkg/api"
)

func ParseSKU(sku string) (uint32, error) {
	if sku == "" {
		return 0, errors.New("sku must be non-empty")
	}

	skuUint64, err := strconv.ParseUint(sku, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid sku format: %w", err)
	}

	return uint32(skuUint64), nil
}

func ValidateAddItemRequest(req *stockpb.AddItemRequest) (models.StockItem, error) {
	if req.GetLocation() == "" {
		return models.StockItem{}, errors.New("location must be non-empty")
	}

	sku, err := ParseSKU(req.GetSku())
	if err != nil {
		return models.StockItem{}, err
	}

	if req.GetCount() <= 0 {
		return models.StockItem{}, errors.New("count must be greater than zero")
	}

	return models.StockItem{
		SKU:      sku,
		Location: req.GetLocation(),
		Count:    uint16(req.GetCount()),
	}, nil
}

func ToProto(item models.StockItem) *stockpb.StockItem {
	return &stockpb.StockItem{
		Sku:      strconv.FormatUint(uint64(item.SKU), 10),
		Location: item.Location,
		Count:    int32(item.Count),
	}
}

func ToProtoList(items []models.StockItem) []*stockpb.StockItem {
	pbItems := make([]*stockpb.StockItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, ToProto(item))
	}
	return pbItems
}
