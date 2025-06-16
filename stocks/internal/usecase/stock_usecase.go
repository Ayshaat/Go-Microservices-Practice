package usecase

import (
	"errors"
	"stocks/internal/models"
	"stocks/internal/repository"
)

var (
	ErrInvalidSKU = errors.New("invalid SKU — not registered")
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
	skuInfo, ok := models.SKUDetails[item.SKU]
	if !ok {
		return ErrInvalidSKU
	}

	item.Name = skuInfo.Name
	item.Type = skuInfo.Type

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
