package usecase

import (
	"cart/internal/models"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"errors"
)

var (
	ErrInvalidSKU = errors.New("invalid SKU — not registered")
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

func (u *cartUseCase) Add(item models.CartItem) error {
	_, err := u.stockRepo.GetBySKU(item.SKU)
	if err != nil {
		return ErrInvalidSKU
	}

	return u.repo.Add(item)
}

func (u *cartUseCase) Delete(userID int64, sku uint32) error {
	return u.repo.Delete(userID, sku)
}

func (u *cartUseCase) List(userID int64) ([]models.CartItem, error) {
	return u.repo.List(userID)
}

func (u *cartUseCase) Clear(userID int64) error {
	return u.repo.Clear(userID)
}
