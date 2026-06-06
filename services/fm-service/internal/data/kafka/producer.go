package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher implements domain.EventPublisher
type KafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher initializes a new Kafka publisher
func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &KafkaPublisher{writer: writer}
}

// Publish serializes the payload to JSON and writes it to the specified topic
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
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

// Close releases the writer resources
func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
