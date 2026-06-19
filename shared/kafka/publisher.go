package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
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

// KafkaHeaderCarrier implements propagation.TextMapCarrier for segmentio/kafka-go Headers.
type KafkaHeaderCarrier []kafka.Header

func (c *KafkaHeaderCarrier) Get(key string) string {
	for _, h := range *c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *KafkaHeaderCarrier) Set(key string, value string) {
	for i, h := range *c {
		if h.Key == key {
			(*c)[i].Value = []byte(value)
			return
		}
	}
	*c = append(*c, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c *KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = h.Key
	}
	return keys
}

// ExtractTraceContext extracts trace context from Kafka headers and returns a new context.
func ExtractTraceContext(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := KafkaHeaderCarrier(headers)
	return otel.GetTextMapPropagator().Extract(ctx, &carrier)
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

	// Inject OTel trace context into Kafka message headers
	headers := KafkaHeaderCarrier(msg.Headers)
	otel.GetTextMapPropagator().Inject(ctx, &headers)
	msg.Headers = []kafka.Header(headers)

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
