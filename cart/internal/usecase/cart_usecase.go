package usecase

import (
	"cart/internal/errors"
	"cart/internal/kafka"
	"cart/internal/models"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"context"
	"log"
	"strconv"
)

type cartUseCase struct {
	repo      repository.CartRepository
	stockRepo stockclient.StockRepository
	producer  kafka.ProducerInterface
}

func NewCartUsecase(repo repository.CartRepository, stockRepo stockclient.StockRepository, producer kafka.ProducerInterface) CartUseCase {
	return &cartUseCase{
		repo:      repo,
		stockRepo: stockRepo,
		producer:  producer,
	}
}

func (u *cartUseCase) sendFailedEvent(userID int64, sku uint32, count int16, reason string) {
	err := u.producer.SendCartItemFailed(
		strconv.FormatInt(userID, 10),
		strconv.FormatUint(uint64(sku), 10),
		int(count),
		"failed",
		reason,
	)

	if err != nil {
		log.Printf("failed to send CartItemFailed event: %v", err)
	}
}

func (u *cartUseCase) Add(ctx context.Context, item models.CartItem) error {
	stockItem, err := u.stockRepo.GetBySKU(ctx, item.SKU)
	if err != nil {
		u.sendFailedEvent(item.UserID, item.SKU, item.Count, "invalid SKU")
		return errors.ErrInvalidSKU
	}

	if item.Count > stockItem.Count {
		u.sendFailedEvent(item.UserID, item.SKU, item.Count, "not enough stock")
		return errors.ErrNotEnoughStock
	}
	item.Price = stockItem.Price

	err = u.repo.Upsert(ctx, item)
	if err != nil {
		u.sendFailedEvent(item.UserID, item.SKU, item.Count, "db error")
		return err
	}

	err = u.producer.SendCartItemAdded(
		strconv.FormatInt(item.UserID, 10),
		strconv.FormatUint(uint64(item.SKU), 10),
		int(item.Count),
		"success",
	)
	if err != nil {
		log.Printf("failed to send CartItemAdded event: %v", err)
	}

	return nil
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
