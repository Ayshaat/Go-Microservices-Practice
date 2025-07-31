package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ayshaat/metrics-consumer/internal/event"
	"github.com/ayshaat/metrics-consumer/internal/log"
	"github.com/ayshaat/metrics-consumer/internal/metrics"

	"github.com/Shopify/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Consumer struct {
	Ready   chan bool
	Logger  log.Logger
	Tracer  trace.Tracer
	Metrics *metrics.Metrics
}

func NewConsumer(logger log.Logger, m *metrics.Metrics) *Consumer {
	return &Consumer{
		Ready:   make(chan bool),
		Logger:  logger,
		Tracer:  otel.Tracer("metrics-consumer"),
		Metrics: m,
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.Ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()

	for msg := range claim.Messages() {
		start := time.Now()
		_, span := c.Tracer.Start(ctx, "ConsumeKafkaMessage")
		span.SetAttributes(
			attribute.String("kafka.topic", msg.Topic),
			attribute.Int64("kafka.offset", msg.Offset),
			attribute.Int("kafka.partition", int(msg.Partition)),
		)

		c.Logger.Info("Kafka message received",
			log.String("topic", msg.Topic),
			log.Int32("partition", msg.Partition),
			log.Int64("offset", msg.Offset),
		)

		var event event.KafkaMessage

		err := json.Unmarshal(msg.Value, &event)

		duration := time.Since(start).Seconds()
		if c.Metrics != nil {
			c.Metrics.RequestsTotal.WithLabelValues("kafka_consumer_consume", "HTTP").Inc()
			c.Metrics.RequestDuration.WithLabelValues("kafka_consumer_consume", "HTTP").Observe(duration)
			if err != nil {
				c.Metrics.RequestErrors.WithLabelValues("kafka_consumer_consume", "HTTP").Inc()
			}
		}

		if err != nil {
			c.Logger.Error("Failed to unmarshal Kafka message", log.Error(err))
			span.RecordError(err)
			span.End()
			sess.MarkMessage(msg, "")

			continue
		}

		eventBytes, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			c.Logger.Error("Failed to pretty-print Kafka message", log.Error(err))
			span.RecordError(err)
		} else {
			c.Logger.Info("Consumed event", log.String("event", string(eventBytes)))
		}

		sess.MarkMessage(msg, "")
		span.End()
	}

	return nil
}
