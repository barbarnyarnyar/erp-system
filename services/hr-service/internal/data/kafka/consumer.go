package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/segmentio/kafka-go"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils")

type DeadLetterMessage struct {
	OriginalTopic string      `json:"original_topic"`
	OriginalKey   string      `json:"original_key,omitempty"`
	Payload       interface{} `json:"payload"`
	Error         string      `json:"error"`
	FailedAt      time.Time   `json:"failed_at"`
	ServiceName   string      `json:"service_name"`
}

type KafkaConsumer struct {
	reader      *kafka.Reader
	publisher   domain.EventPublisher
	expenseSvc  service.ExpenseService
	reliableSvc service.ReliableMessagingService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	expenseSvc service.ExpenseService,
	reliableSvc service.ReliableMessagingService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicPrjTimeLogged,
		domain.TopicFmVendorPaid,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:      reader,
		publisher:   publisher,
		expenseSvc:  expenseSvc,
		reliableSvc: reliableSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for hr-service...")
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer due to context cancellation...")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error reading message: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			log.Printf("Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			// Extract trace context and register trace ID
			msgCtx := sharedkafka.ExtractTraceContext(ctx, msg.Headers)
			traceID := utils.GetTraceIDFromContext(msgCtx)
			utils.SetTraceID(traceID)

			// Inject publisher into message context for DLQ routing in idempotent transactions
			msgCtx = context.WithValue(msgCtx, "publisher", c.publisher)

			if err := c.handleMessage(msgCtx, msg.Topic, msg.Value); err != nil {
				log.Printf("Failed to process event %s: %v", msg.Topic, err)
				c.publishToDLQ(msgCtx, msg.Topic, string(msg.Key), msg.Value, err)
			}
			utils.ClearTraceID()
		}
	}
}

func (c *KafkaConsumer) publishToDLQ(ctx context.Context, topic string, key string, value []byte, err error) {
	dlqMsg := DeadLetterMessage{
		OriginalTopic: topic,
		OriginalKey:   key,
		Payload:       string(value),
		Error:         err.Error(),
		FailedAt:      time.Now(),
		ServiceName:   "hr-service",
	}
	dlqTopic := topic + ".dead-letter"
	if dlqErr := c.publisher.Publish(ctx, dlqTopic, key, dlqMsg); dlqErr != nil {
		log.Printf("ERROR: failed to publish DLQ message for topic %s: %v", topic, dlqErr)
	} else {
		log.Printf("ERROR: consumer handler failed for topic %s: %v — sent to DLQ topic %s", topic, err, dlqTopic)
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicPrjTimeLogged:
		var ev domain.PrjTimeLoggedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, domain.TopicPrjTimeLogged, ev, func(txCtx context.Context) error {
			log.Printf("[Idempotent] Processing PrjTimeLoggedEvent: Project %s, Total Hours: %s, Number of Logs: %d",
				ev.ProjectID, ev.TotalAccumulatedHours.String(), len(ev.Details))
			return nil
		})

	case domain.TopicFmVendorPaid:
		var ev domain.FmVendorPaidEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, domain.TopicFmVendorPaid, ev, func(txCtx context.Context) error {
			log.Printf("[Idempotent] Processing FmVendorPaidEvent: Bill %s, Target Document (Expense Claim) %s Paid",
				ev.BillID, ev.TargetDocumentID)
			return c.expenseSvc.ClearClaimForPayment(txCtx, ev.TargetDocumentID)
		})
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
