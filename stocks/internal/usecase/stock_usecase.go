package usecase

import (
	"context"
	"stocks/internal/errors"
	"stocks/internal/models"
	"stocks/internal/repository"
)

type stockUseCase struct {
	repo repository.StockRepository
}

func NewStockUsecase(repo repository.StockRepository) StockUseCase {
	return &stockUseCase{
		repo: repo,
	}
}

func (u *stockUseCase) Add(ctx context.Context, item models.StockItem) error {
	_, _, err := u.repo.GetSKUInfo(ctx, item.SKU)
	if err != nil {
		return errors.ErrInvalidSKU
	}

	return u.repo.Add(ctx, item)
}

func (u *stockUseCase) Delete(ctx context.Context, sku uint32) error {
	return u.repo.Delete(ctx, sku)
}

func (u *stockUseCase) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	return u.repo.GetBySKU(ctx, sku)
}

func (u *stockUseCase) ListByLocation(ctx context.Context, location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	return u.repo.ListByLocation(ctx, location, pageSize, currentPage)
}
