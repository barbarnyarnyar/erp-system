package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// Publisher handles event publishing to Kafka brokers.
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher initializes a new Kafka publisher writer with least bytes balancer and auto-topic creation.
func NewPublisher(brokers []string) *Publisher {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &Publisher{writer: writer}
}

// Publish serializes the payload to JSON and writes the message to the specified Kafka topic.
func (p *Publisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: body,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("Failed to publish message to topic %s: %v", topic, err)
		return fmt.Errorf("failed to write message to topic %s: %w", topic, err)
	}

	log.Printf("Successfully published event to topic %s with key %s", topic, key)
	return nil
}

// Close closes the underlying writer connection.
func (p *Publisher) Close() error {
	return p.writer.Close()
}
