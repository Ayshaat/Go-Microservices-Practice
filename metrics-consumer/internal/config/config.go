package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers   []string
	ConsumerGroup  string
	Topic          string
	JaegerEndpoint string
}

func Load(envFile string) (*Config, error) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}

	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "metrics-consumer-group"
	}

	topic := os.Getenv("TOPIC")
	if topic == "" {
		topic = "metrics"
	}

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces" // or log an error if it's required
	}

	brokers := []string{}
	brokers = append(brokers, splitAndTrim(brokersEnv, ",")...)

	return &Config{
		KafkaBrokers:   brokers,
		ConsumerGroup:  consumerGroup,
		Topic:          topic,
		JaegerEndpoint: jaegerEndpoint,
	}, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}
