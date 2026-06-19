package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/segmentio/kafka-go"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils")

// KafkaConsumer subscribes to HR offboarding events and translates them
// into Auth user deactivations. This closes the loop: when HR terminates an
// employee, the corresponding Auth user account is automatically deactivated
// (which bumps the security_stamp and invalidates any in-flight JWTs).
type KafkaConsumer struct {
	reader    *kafka.Reader
	publisher domain.EventPublisher
	userSvc   *service.UserService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	userSvc *service.UserService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicHrEmployeeTerminated,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		publisher: publisher,
		userSvc:   userSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for auth-service...")
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

			log.Printf("[AUTH-CONSUMER] Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			// Extract trace context and register trace ID
			msgCtx := sharedkafka.ExtractTraceContext(ctx, msg.Headers)
			traceID := utils.GetTraceIDFromContext(msgCtx)
			utils.SetTraceID(traceID)

			// Inject publisher into message context for DLQ routing in idempotent transactions
			msgCtx = context.WithValue(msgCtx, "publisher", c.publisher)

			if err := c.handleMessage(msgCtx, msg.Topic, msg.Value); err != nil {
				log.Printf("[AUTH-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
			utils.ClearTraceID()
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicHrEmployeeTerminated:
		var ev domain.HREmployeeTerminatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return fmt.Errorf("failed to unmarshal HREmployeeTerminatedEvent: %w", err)
		}
		log.Printf("[AUTH-CONSUMER] Processing HR employee termination: deactivating user %s (reason: %s)",
			ev.EmployeeID, ev.Reason)

		// Treat EmployeeID as the Auth User ID per cross-service @reference
		// convention (master PRD 2.10). If the user does not exist, treat
		// as idempotent success — there is nothing to deactivate.
		if err := c.userSvc.DeactivateUser(ctx, ev.EmployeeID); err != nil {
			return fmt.Errorf("failed to deactivate user %s: %w", ev.EmployeeID, err)
		}
		log.Printf("[AUTH-CONSUMER] User %s deactivated due to HR termination", ev.EmployeeID)
		return nil
	}
	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
