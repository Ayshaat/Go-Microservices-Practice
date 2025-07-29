package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ayshaat/metrics-consumer/internal/config"
	"github.com/ayshaat/metrics-consumer/internal/kafka"
	"github.com/ayshaat/metrics-consumer/internal/log"
	"github.com/ayshaat/metrics-consumer/internal/log/zap"
	"github.com/ayshaat/metrics-consumer/internal/metrics"
	"github.com/ayshaat/metrics-consumer/internal/trace"

	"github.com/Shopify/sarama"
)

func main() {
	envFile := ".env.docker"

	logger, cleanup, err := zap.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	cfg, err := config.Load(envFile)
	if err != nil {
		logger.Error("Failed to load config", log.Error(err))
	}

	shutdown, err := trace.InitTracer("metrics-consumer", cfg.JaegerEndpoint)
	if err != nil {
		logger.Error("Failed to initialize tracer", log.Error(err))
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer", log.Error(err))
		}
	}()

	configSarama := sarama.NewConfig()
	configSarama.Version = sarama.V2_8_0_0
	configSarama.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	configSarama.Consumer.Offsets.Initial = sarama.OffsetNewest

	metricsInstance := metrics.RegisterMetrics()
	metrics.StartMetricsServer(":9095")
	consumer := kafka.NewConsumer(logger, metricsInstance)

	ctx, cancel := context.WithCancel(context.Background())

	client, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.ConsumerGroup, configSarama)
	if err != nil {
		logger.Error("Error creating consumer group client", log.Error(err))
	}

	defer func() {
		if err := client.Close(); err != nil {
			logger.Error("Error closing client", log.Error(err))
		}
	}()

	go func() {
		for {
			if err := client.Consume(ctx, []string{cfg.Topic}, consumer); err != nil {
				logger.Error("Error from consumer", log.Error(err))
			}

			if ctx.Err() != nil {
				return
			}
			consumer = kafka.NewConsumer(logger, metricsInstance)
		}
	}()

	<-consumer.Ready

	logger.Info("Metrics Consumer started, consuming topic", log.String("topic", cfg.Topic))

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigterm

	logger.Info("Terminating: via signal", log.String("signal", sig.String()))
	cancel()
}
