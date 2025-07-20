package kafka

//go:generate mockgen -source=producer_interface.go -destination=../usecase/mocks/mock_producer.go -package=mocks ProducerInterface

type ProducerInterface interface {
	SendCartItemAdded(cartId, sku string, count int, status string) error
	SendCartItemFailed(cartId, sku string, count int, status, reason string) error
	Close() error
}
