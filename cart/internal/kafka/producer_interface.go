package kafka

type ProducerInterface interface {
	SendCartItemAdded(cartId, sku string, count int, status string) error
	SendCartItemFailed(cartId, sku string, count int, status, reason string) error
	Close() error
}
