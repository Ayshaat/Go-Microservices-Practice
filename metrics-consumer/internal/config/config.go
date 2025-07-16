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
<<<<<<< HEAD
	brokers = append(brokers, splitAndTrim(brokersEnv, ",")...)
=======
	for _, b := range splitAndTrim(brokersEnv, ",") {
		brokers = append(brokers, b)
	}
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318

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
<<<<<<< HEAD

=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
	return parts
}
