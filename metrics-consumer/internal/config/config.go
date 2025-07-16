package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers  []string
	ConsumerGroup string
	Topic         string
}

func Load() (*Config, error) {
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

	brokers := []string{}
	brokers = append(brokers, splitAndTrim(brokersEnv, ",")...)

	return &Config{
		KafkaBrokers:  brokers,
		ConsumerGroup: consumerGroup,
		Topic:         topic,
	}, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
