package usecase

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log"
	"stocks/internal/errors"
	"stocks/internal/kafka"
	"stocks/internal/models"
	"stocks/internal/repository"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type stockUseCase struct {
	repo      repository.StockRepository
	txManager trm.Manager
	producer  kafka.ProducerInterface
}

func NewStockUsecase(repo repository.StockRepository, txManager trm.Manager, producer kafka.ProducerInterface) StockUseCase {
	return &stockUseCase{
		repo:      repo,
		txManager: txManager,
		producer:  producer,
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

			err = u.repo.InsertStockItem(ctx, item)
			if err == nil {
				if err := u.producer.SendSKUCreated(fmt.Sprint(item.SKU), item.Price, int(item.Count)); err != nil {
					log.Printf("failed to send SKUCreated event: %v", err)
				}
			}

			return err
		}

		if existingItem.UserID != item.UserID {
			return errors.ErrOwnershipViolation
		}

		existingItem.Count += item.Count

		err = u.repo.UpdateCount(ctx, existingItem.UserID, existingItem.SKU, existingItem.Count, item.Price)
		if err == nil {
			if err := u.producer.SendStockChanged(fmt.Sprint(existingItem.SKU), int(existingItem.Count), existingItem.Price); err != nil {
				log.Printf("failed to send StockChanged event: %v", err)
			}
		}

		return err
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
