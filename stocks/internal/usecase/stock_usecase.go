package usecase

import (
	"context"
	stdErrors "errors"
	"stocks/internal/errors"
	"stocks/internal/models"
	"stocks/internal/repository"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type stockUseCase struct {
	repo      repository.StockRepository
	txManager trm.Manager
}

func NewStockUsecase(repo repository.StockRepository, txManager trm.Manager) StockUseCase {
	return &stockUseCase{
		repo:      repo,
		txManager: txManager,
	}
}

func (u *stockUseCase) Add(ctx context.Context, item models.StockItem) error {
	return u.txManager.Do(ctx, func(ctx context.Context) error {
		_, _, err := u.repo.GetSKUInfo(ctx, item.SKU)
		if err != nil {
			return errors.ErrInvalidSKU
		}

		existingItem, err := u.repo.GetByUserSKU(ctx, item.UserID, item.SKU)
		if err != nil {
			if !stdErrors.Is(err, errors.ErrItemNotFound) {
				return err
			}

			return u.repo.InsertStockItem(ctx, item)
		}

		if existingItem.UserID != item.UserID {
			return errors.ErrOwnershipViolation
		}

		existingItem.Count += item.Count

		return u.repo.UpdateCount(ctx, existingItem.UserID, existingItem.SKU, existingItem.Count)
	})
}

func (u *stockUseCase) Delete(ctx context.Context, sku uint32) error {
	return u.txManager.Do(ctx, func(ctx context.Context) error {
		return u.repo.Delete(ctx, sku)
	})
}

func (u *stockUseCase) GetBySKU(ctx context.Context, sku uint32) (models.StockItem, error) {
	return u.repo.GetBySKU(ctx, sku)
}

func (u *stockUseCase) ListByLocation(ctx context.Context, location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	return u.repo.ListByLocation(ctx, location, pageSize, currentPage)
}
