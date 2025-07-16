package kafka

//go:generate mockgen -source=internal/kafka/producer_interface.go -destination=internal/usecase/mocks/mock_producer.go -package=mocks

type ProducerInterface interface {
	SendSKUCreated(sku string, price float64, count int) error
	SendStockChanged(sku string, count int, price float64) error
	Close() error
}
