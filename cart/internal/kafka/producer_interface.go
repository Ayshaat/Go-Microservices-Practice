package kafka

import (
	"context"
)

//go:generate mockgen -source=producer_interface.go -destination=../usecase/mocks/mock_producer.go -package=mocks ProducerInterface

type ProducerInterface interface {
	SendCartItemAdded(ctx context.Context, cartId, sku string, count int, status string) error
	SendCartItemFailed(ctx context.Context, cartId, sku string, count int, status, reason string) error
	Close() error
}
