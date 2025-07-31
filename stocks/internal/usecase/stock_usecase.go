package usecase

import (
	"context"
	stdErrors "errors"
	"stocks/internal/errors"
	"stocks/internal/kafka"
	"stocks/internal/log"
	"stocks/internal/models"
	"stocks/internal/repository"
	"strconv"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type stockUseCase struct {
	repo      repository.StockRepository
	txManager trm.Manager
	producer  kafka.ProducerInterface
	logger    log.Logger
}

func NewStockUsecase(repo repository.StockRepository, txManager trm.Manager, producer kafka.ProducerInterface, logger log.Logger) StockUseCase {
	return &stockUseCase{
		repo:      repo,
		txManager: txManager,
		producer:  producer,
		logger:    logger,
	}
}

func (u *stockUseCase) sendSKUCreatedEvent(ctx context.Context, sku uint32, price float64, count int) {
	err := u.producer.SendSKUCreated(ctx, strconv.FormatUint(uint64(sku), 10), price, count)
	if err != nil {
		u.logger.Error("failed to send SKUCreated event", log.Error(err))
	}
}

func (u *stockUseCase) sendStockChangedEvent(ctx context.Context, sku uint32, count int, price float64) {
	err := u.producer.SendStockChanged(ctx, strconv.FormatUint(uint64(sku), 10), count, price)
	if err != nil {
		u.logger.Error("failed to send StockChanged event", log.Error(err))
	}
}

func (u *stockUseCase) Add(ctx context.Context, item models.StockItem) error {
	tracer := otel.Tracer("stocks-usecase")
	ctx, span := tracer.Start(ctx, "Add")
	defer span.End()

	span.SetAttributes(
		attribute.Int64("user.id", item.UserID),
		attribute.Int64("item.sku", int64(item.SKU)),
		attribute.Int64("item.count", int64(item.Count)),
		attribute.Float64("item.price", item.Price),
	)

	return u.txManager.Do(ctx, func(ctx context.Context) error {
		_, _, err := u.repo.GetSKUInfo(ctx, item.SKU)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid SKU")
			return errors.ErrInvalidSKU
		}

		existingItem, err := u.repo.GetByUserSKU(ctx, item.UserID, item.SKU)
		if err != nil {
			if !stdErrors.Is(err, errors.ErrItemNotFound) {
				span.RecordError(err)
				span.SetStatus(codes.Error, "db error fetching item")
				return err
			}

			err = u.repo.InsertStockItem(ctx, item)
			if err == nil {
				u.sendSKUCreatedEvent(ctx, item.SKU, item.Price, int(item.Count))
			} else {
				u.logger.Error("failed to insert stock item", log.Error(err))
				span.RecordError(err)
				span.SetStatus(codes.Error, "insert failed")
			}

			return err
		}

		if existingItem.UserID != item.UserID {
			span.SetStatus(codes.Error, "ownership violation")
			return errors.ErrOwnershipViolation
		}

		existingItem.Count += item.Count

		err = u.repo.UpdateCount(ctx, existingItem.UserID, existingItem.SKU, existingItem.Count, item.Price)
		if err == nil {
			u.sendStockChangedEvent(ctx, existingItem.SKU, int(existingItem.Count), existingItem.Price)
		} else {
			u.logger.Error("failed to update stock count", log.Error(err))
			span.RecordError(err)
			span.SetStatus(codes.Error, "update failed")
		}

		return err
	})
}

func (u *stockUseCase) Delete(ctx context.Context, sku uint32) error {
	tracer := otel.Tracer("stocks-usecase")
	ctx, span := tracer.Start(ctx, "Delete")
	defer span.End()

	span.SetAttributes(attribute.Int64("item.sku", int64(sku)))

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
