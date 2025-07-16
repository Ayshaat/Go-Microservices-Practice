package kafka

type ProducerInterface interface {
	SendSKUCreated(sku string, price float64, count int) error
	SendStockChanged(sku string, count int, price float64) error
	Close() error
}
