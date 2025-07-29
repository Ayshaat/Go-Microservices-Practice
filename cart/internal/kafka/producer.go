package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"cart/internal/event"
	"cart/internal/log"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
	logger    log.Logger
}

func NewProducer(cfg *ProducerConfig, logger log.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = maxProducerRetry
	config.Producer.Partitioner = func(topic string) sarama.Partitioner {
		return sarama.NewManualPartitioner(topic)
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		logger.Error("failed to create Kafka producer", log.Error(err))
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	logger.Info("Kafka producer created",
		log.Strings("brokers", cfg.Brokers),
		log.String("topic", cfg.Topic),
		log.Int32("partition", cfg.Partition),
		log.String("service", cfg.Service),
	)

	return &Producer{
		producer:  producer,
		topic:     cfg.Topic,
		partition: cfg.Partition,
		service:   cfg.Service,
		logger:    logger,
	}, nil
}

func (p *Producer) SendCartItemAdded(ctx context.Context, cartId, sku string, count int, status string) error {
	tr := otel.Tracer("kafka-producer")
	ctx, span := tr.Start(ctx, "SendCartItemAdded")
	defer span.End()

	span.SetAttributes(
		attribute.String("cart_id", cartId),
		attribute.String("sku", sku),
		attribute.Int("count", count),
		attribute.String("status", status),
	)

	payload := event.CartItemAddedPayload{
		CartID: cartId,
		SKU:    sku,
		Count:  count,
		Status: status,
	}

	p.logger.Info("Sending cart_item_added event",
		log.String("cart_id", cartId),
		log.String("sku", sku),
		log.Int("count", count),
		log.String("status", status),
	)

	return p.send(ctx, "cart_item_added", payload)
}

func (p *Producer) SendCartItemFailed(ctx context.Context, cartId, sku string, count int, status, reason string) error {
	tr := otel.Tracer("kafka-producer")
	ctx, span := tr.Start(ctx, "SendCartItemFailed")
	defer span.End()

	span.SetAttributes(
		attribute.String("cart_id", cartId),
		attribute.String("sku", sku),
		attribute.Int("count", count),
		attribute.String("status", status),
		attribute.String("reason", reason),
	)

	payload := event.CartItemFailedPayload{
		CartID: cartId,
		SKU:    sku,
		Count:  count,
		Status: status,
		Reason: reason,
	}

	p.logger.Info("Sending cart_item_failed event",
		log.String("cart_id", cartId),
		log.String("sku", sku),
		log.Int("count", count),
		log.String("status", status),
		log.String("reason", reason),
	)

	return p.send(ctx, "cart_item_failed", payload)
}

func (p *Producer) send(ctx context.Context, eventType string, payload interface{}) error {
	msg := event.KafkaMessage{
		Type:      eventType,
		Service:   p.service,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   payload,
	}

	valueBytes, err := json.Marshal(msg)
	if err != nil {
		p.logger.Error("Failed to marshal Kafka message",
			log.String("event_type", eventType),
			log.Error(err),
		)
		return fmt.Errorf("failed to marshal Kafka message: %w", err)
	}

	producerMsg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Partition: p.partition,
		Value:     sarama.ByteEncoder(valueBytes),
		Key:       sarama.StringEncoder(fmt.Sprintf("%s-%d", p.service, time.Now().UnixNano())),
	}

	partition, offset, err := p.producer.SendMessage(producerMsg)
	if err != nil {
		p.logger.Error("Failed to send Kafka message",
			log.String("event_type", eventType),
			log.Error(err),
		)
		return fmt.Errorf("failed to send Kafka message: %w", err)
	}

	p.logger.Info("Kafka message sent successfully",
		log.String("event_type", eventType),
		log.Int32("partition", partition),
		log.Int64("offset", offset),
	)

	return nil
}

func (p *Producer) Close() error {
	err := p.producer.Close()
	if err != nil {
		p.logger.Error("failed to close Kafka producer", log.Error(err))
	} else {
		p.logger.Info("Kafka producer closed")
	}
	return err
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
