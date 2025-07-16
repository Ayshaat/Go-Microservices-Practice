package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ayshaat/metrics-consumer/internal/event"

	"github.com/Shopify/sarama"
)

type Consumer struct {
	Ready chan bool
}

func NewConsumer() *Consumer {
	return &Consumer{
		Ready: make(chan bool),
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
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%s partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)

		var event event.KafkaMessage

		err := json.Unmarshal(msg.Value, &event)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			sess.MarkMessage(msg, "")

			continue
		}

		eventBytes, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal event: %v", err)
		} else {
			fmt.Printf("Consumed event:\n%s\n\n", string(eventBytes))
		}

		sess.MarkMessage(msg, "")
	}

	return nil
}
