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

	"stocks/internal/event"
	"stocks/internal/log"

	"github.com/Shopify/sarama"
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

func (p *Producer) SendSKUCreated(ctx context.Context, sku string, price float64, count int) error {
	tr := otel.Tracer("kafka-producer")
	ctx, span := tr.Start(ctx, "SendSKUCreated")
	defer span.End()

	span.SetAttributes(
		attribute.String("sku", sku),
		attribute.Float64("price", price),
		attribute.Int("count", count),
	)

	payload := event.SKUCreatedPayload{
		SKU:   sku,
		Price: price,
		Count: count,
	}

	p.logger.Info("Sending sku_created event",
		log.String("sku", sku),
		log.Float64("price", price),
		log.Int("count", count),
	)

	return p.send(ctx, "sku_created", payload)
}

func (p *Producer) SendStockChanged(ctx context.Context, sku string, count int, price float64) error {
	tr := otel.Tracer("kafka-producer")
	ctx, span := tr.Start(ctx, "SendStockChanged")
	defer span.End()

	span.SetAttributes(
		attribute.String("sku", sku),
		attribute.Int("count", count),
		attribute.Float64("price", price),
	)

	payload := event.StockChangedPayload{
		SKU:   sku,
		Count: count,
		Price: price,
	}

	p.logger.Info("Sending stock_changed event",
		log.String("sku", sku),
		log.Int("count", count),
		log.Float64("price", price),
	)

	return p.send(ctx, "stock_changed", payload)
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
		p.logger.Error("failed to marshal Kafka message", log.Error(err))
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
		p.logger.Error("failed to send Kafka message", log.Error(err))
		return fmt.Errorf("failed to send Kafka message: %w", err)
	}

	p.logger.Info("Kafka message sent",
		log.String("event_type", eventType),
		log.String("topic", p.topic),
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
		partitionStr = "1"
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
		service = "stocks-service"
	}

	return &ProducerConfig{
		Brokers:   brokers,
		Topic:     topic,
		Partition: int32(partition),
		Service:   service,
	}, nil
}
