package usecase

import (
	"cart/internal/errors"
	"cart/internal/kafka"
	"cart/internal/log"
	"cart/internal/models"
	"cart/internal/repository"
	"cart/internal/stockclient"
	"context"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type cartUseCase struct {
	repo      repository.CartRepository
	stockRepo stockclient.StockRepository
	producer  kafka.ProducerInterface
	logger    log.Logger
}

func NewCartUsecase(repo repository.CartRepository, stockRepo stockclient.StockRepository, producer kafka.ProducerInterface, logger log.Logger) CartUseCase {
	return &cartUseCase{
		repo:      repo,
		stockRepo: stockRepo,
		producer:  producer,
		logger:    logger,
	}
}

func (u *cartUseCase) sendFailedEvent(ctx context.Context, userID int64, sku uint32, count int16, reason string) {
	err := u.producer.SendCartItemFailed(
		ctx,
		strconv.FormatInt(userID, 10),
		strconv.FormatUint(uint64(sku), 10),
		int(count),
		"failed",
		reason,
	)

	if err != nil {
		u.logger.Error("failed to send CartItemFailed event: %v", log.Error(err))
	}
}

func (u *cartUseCase) Add(ctx context.Context, item models.CartItem) error {
	tracer := otel.Tracer("cart-usecase")
	ctx, span := tracer.Start(ctx, "Add")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", item.UserID),
		attribute.Int64("item.sku", int64(item.SKU)),
		attribute.Int64("item.count", int64(item.Count)),
	)

	stockItem, err := u.stockRepo.GetBySKU(ctx, item.SKU)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid SKU")
		u.sendFailedEvent(ctx, item.UserID, item.SKU, item.Count, "invalid SKU - not registered")
		u.logger.Error("stockRepo.GetBySKU failed", log.Error(err))
		return errors.ErrInvalidSKU
	}

	if item.Count > stockItem.Count {
		reason := "not enough stock"
		span.SetStatus(codes.Error, reason)
		u.sendFailedEvent(ctx, item.UserID, item.SKU, item.Count, "not enough stock available")
		u.logger.Warn("not enough stock", log.Int16("requested", item.Count), log.Int16("available", stockItem.Count))
		return errors.ErrNotEnoughStock
	}
	item.Price = stockItem.Price

	err = u.repo.Upsert(ctx, item)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		u.sendFailedEvent(ctx, item.UserID, item.SKU, item.Count, "db error")
		u.logger.Error("repo.Upsert failed", log.Error(err))
		return err
	}

	err = u.producer.SendCartItemAdded(
		ctx,
		strconv.FormatInt(item.UserID, 10),
		strconv.FormatUint(uint64(item.SKU), 10),
		int(item.Count),
		"success",
	)
	if err != nil {
		u.logger.Error("failed to send CartItemAdded event: %v", log.Error(err))
		span.RecordError(err)
	}
	span.SetStatus(codes.Ok, "success")
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
