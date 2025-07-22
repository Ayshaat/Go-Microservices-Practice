package service

import (
	"fmt"
	"math"
	"strconv"

	"cart/internal/models"
	cartpb "cart/pkg/api/cart"
)

func ParseUserID(userIDStr string) (int64, error) {
	return strconv.ParseInt(userIDStr, 10, 64)
}

func ParseSKU(skuStr string) (uint32, error) {
	skuUint64, err := strconv.ParseUint(skuStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(skuUint64), nil
}

func CartItemFromAddRequest(req *cartpb.AddItemRequest) (models.CartItem, error) {
	userID, err := ParseUserID(req.UserId)
	if err != nil {
		return models.CartItem{}, fmt.Errorf("invalid user_id: %w", err)
	}

	sku, err := ParseSKU(req.Sku)
	if err != nil {
		return models.CartItem{}, fmt.Errorf("invalid sku: %w", err)
	}

	if req.Count < math.MinInt16 || req.Count > math.MaxInt16 {
		return models.CartItem{}, fmt.Errorf("count value %d out of int16 range", req.Count)
	}
	count := int16(req.Count)

	return models.CartItem{
		UserID: userID,
		SKU:    sku,
		Count:  count,
	}, nil
}
