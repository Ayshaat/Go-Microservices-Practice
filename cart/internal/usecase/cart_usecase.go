package usecase

import (
	"cart/internal/errors"
	"cart/internal/models"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"context"
)

type cartUseCase struct {
	repo      repository.CartRepository
	stockRepo stockclient.StockRepository
}

func NewCartUsecase(repo repository.CartRepository, stockRepo stockclient.StockRepository) CartUseCase {
	return &cartUseCase{
		repo:      repo,
		stockRepo: stockRepo,
	}
}

func (u *cartUseCase) Add(ctx context.Context, item models.CartItem) error {
	_, err := u.stockRepo.GetBySKU(ctx, item.SKU)
	if err != nil {
		return errors.ErrInvalidSKU
	}

	return u.repo.Add(ctx, item)
}

func (u *cartUseCase) Delete(ctx context.Context, userID int64, sku uint32) error {
	return u.repo.Delete(ctx, userID, sku)
}

func (u *cartUseCase) List(ctx context.Context, userID int64) ([]models.CartItem, error) {
	items, err := u.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range items {
		stockItem, err := u.stockRepo.GetBySKU(ctx, items[i].SKU)
		if err != nil {
			return nil, err
		}
		items[i].Price = stockItem.Price
		items[i].Count = stockItem.Count
	}

	return items, nil
}

func (u *cartUseCase) Clear(ctx context.Context, userID int64) error {
	return u.repo.Clear(ctx, userID)
}
