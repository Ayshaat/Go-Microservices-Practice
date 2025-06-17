package usecase

import (
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

func (u *stockUseCase) Add(item models.StockItem) error {
	name, typ, err := u.repo.GetSKUInfo(item.SKU)
	if err != nil {
		return errors.ErrInvalidSKU
	}

	item.Name = name
	item.Type = typ

	return u.repo.Add(item)
}

func (u *stockUseCase) Delete(sku uint32) error {
	return u.repo.Delete(sku)
}

func (u *stockUseCase) GetBySKU(sku uint32) (models.StockItem, error) {
	return u.repo.GetBySKU(sku)
}

func (u *stockUseCase) ListByLocation(location string, pageSize, currentPage int64) ([]models.StockItem, error) {
	return u.repo.ListByLocation(location, pageSize, currentPage)
}
