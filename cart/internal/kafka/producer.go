package kafka

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"cart/internal/event"

	"github.com/IBM/sarama"
)

const maxProducerRetry = 5

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

func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = maxProducerRetry
	config.Producer.Partitioner = func(topic string) sarama.Partitioner {
		return sarama.NewManualPartitioner(topic)
	}

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

func (p *Producer) SendCartItemAdded(cartId, sku string, count int, status string) error {
	payload := event.CartItemAddedPayload{
		CartID: cartId,
		SKU:    sku,
		Count:  count,
		Status: status,
	}

	return p.send("cart_item_added", payload)
}

func (p *Producer) SendCartItemFailed(cartId, sku string, count int, status, reason string) error {
	payload := event.CartItemFailedPayload{
		CartID: cartId,
		SKU:    sku,
		Count:  count,
		Status: status,
		Reason: reason,
	}

	return p.send("cart_item_failed", payload)
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

func NewProducerConfigFromEnv() (*ProducerConfig, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS env var is not set")
	}
	brokers := strings.Split(brokersEnv, ",")

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "metrics"
	}

	partitionStr := os.Getenv("PARTITION")
	if partitionStr == "" {
		partitionStr = "0"
	}

	partitionInt, err := strconv.Atoi(partitionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PARTITION: %w", err)
	}
	if partitionInt < math.MinInt32 || partitionInt > math.MaxInt32 {
		return nil, fmt.Errorf("partition value %d out of int32 range", partitionInt)
	}

	partition := int32(partitionInt)

	service := os.Getenv("SERVICE_NAME")
	if service == "" {
		service = "cart-service"
	}

	return &ProducerConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: int32(partition),
		Service:   service,
	}, nil
}
