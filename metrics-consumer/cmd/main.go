package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

<<<<<<< HEAD
	"github.com/ayshaat/metrics-consumer/internal/config"
	"github.com/ayshaat/metrics-consumer/internal/kafka"

	"github.com/Shopify/sarama"
=======
	"metrics-consumer/internal/config"
	"metrics-consumer/internal/kafka"

	"github.com/IBM/sarama"
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	configSarama := sarama.NewConfig()
	configSarama.Version = sarama.V2_8_0_0
	configSarama.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	configSarama.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer := kafka.NewConsumer()

	ctx, cancel := context.WithCancel(context.Background())
<<<<<<< HEAD

=======
>>>>>>> 06ad7f29756e466367a0284cadef04bc7c11f318
	client, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.ConsumerGroup, configSarama)
	if err != nil {
		log.Fatalf("Error creating consumer group client: %v", err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	go func() {
		for {
			if err := client.Consume(ctx, []string{cfg.Topic}, consumer); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
			consumer = kafka.NewConsumer()
		}
	}()

	<-consumer.Ready

	log.Println("Metrics Consumer started, consuming topic:", cfg.Topic)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Println("Terminating: via signal")
	cancel()
}
