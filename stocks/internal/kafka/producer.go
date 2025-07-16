package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"stocks/internal/event"

	"github.com/Shopify/sarama"
)

<<<<<<< HEAD
const maxProducerRetry = 5

=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
type ProducerConfig struct {
	Brokers   []string
	Topic     string
	Partition int32
	Service   string
}

type Producer struct {
	producer  sarama.SyncProducer
	topic     string
	partition int32
	service   string
}

func NewProducer(cfg ProducerConfig) (*Producer, error) {
	config := sarama.NewConfig()
<<<<<<< HEAD
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = maxProducerRetry
=======
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer:  producer,
		topic:     cfg.Topic,
		partition: cfg.Partition,
		service:   cfg.Service,
	}, nil
}

func (p *Producer) SendSKUCreated(sku string, price float64, count int) error {
	payload := event.SKUCreatedPayload{
		SKU:   sku,
		Price: price,
		Count: count,
	}
<<<<<<< HEAD

=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
	return p.send("sku_created", payload)
}

func (p *Producer) SendStockChanged(sku string, count int, price float64) error {
	payload := event.StockChangedPayload{
		SKU:   sku,
		Count: count,
		Price: price,
	}
<<<<<<< HEAD

=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
	return p.send("stock_changed", payload)
}

func (p *Producer) send(eventType string, payload interface{}) error {
	msg := event.KafkaMessage{
		Type:      eventType,
		Service:   p.service,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
	}

	valueBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Kafka message: %w", err)
	}

	producerMsg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Partition: p.partition,
		Value:     sarama.ByteEncoder(valueBytes),
		Key:       sarama.StringEncoder(fmt.Sprintf("%s-%d", p.service, time.Now().UnixNano())),
	}

	_, _, err = p.producer.SendMessage(producerMsg)
	if err != nil {
		return fmt.Errorf("failed to send Kafka message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
