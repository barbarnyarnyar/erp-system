package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/segmentio/kafka-go"
	"erp-system/shared/utils"

	sharedkafka "erp-system/shared/kafka")

type DeadLetterMessage struct {
	OriginalTopic string      `json:"original_topic"`
	OriginalKey   string      `json:"original_key,omitempty"`
	Payload       interface{} `json:"payload"`
	Error         string      `json:"error"`
	FailedAt      time.Time   `json:"failed_at"`
	ServiceName   string      `json:"service_name"`
}

type KafkaConsumer struct {
	reader          *kafka.Reader
	publisher       domain.EventPublisher
	reliableMsgSvc  service.ReliableMessagingService
	projTrackingSvc service.ProjectTrackingService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	reliableMsgSvc service.ReliableMessagingService,
	projTrackingSvc service.ProjectTrackingService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicHrEmployeeCreated,
		domain.TopicHrEmployeeTerminated,
		domain.TopicCrmSalesOrderConfirmed,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:          reader,
		publisher:       publisher,
		reliableMsgSvc:  reliableMsgSvc,
		projTrackingSvc: projTrackingSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for pm-service...")
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
		ServiceName:   "pm-service",
	}
	dlqTopic := topic + ".dead-letter"
	if dlqErr := c.publisher.Publish(ctx, dlqTopic, key, dlqMsg); dlqErr != nil {
		log.Printf("ERROR: failed to publish DLQ message for topic %s: %v", topic, dlqErr)
	} else {
		log.Printf("ERROR: consumer handler failed for topic %s: %v — sent to DLQ topic %s", topic, err, dlqTopic)
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	var genericEv struct {
		EventID string `json:"event_id"`
	}
	if err := json.Unmarshal(value, &genericEv); err != nil {
		return err
	}
	if genericEv.EventID == "" {
		genericEv.EventID = utils.NewID("missing-evt")
	}

	return c.reliableMsgSvc.ExecuteIdempotentTransaction(ctx, genericEv.EventID, topic, value, func(txCtx context.Context) error {
		switch topic {
		case domain.TopicHrEmployeeCreated:
			var ev domain.HrEmployeeCreatedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("[Idempotent Inbox] Processed HR Employee Created: ID=%s, LegalEntity=%s, Role=%s", ev.EmployeeID, ev.LegalEntityID, ev.ExplicitRole)
			return nil

		case domain.TopicHrEmployeeTerminated:
			var ev domain.HrEmployeeTerminatedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("[Idempotent Inbox] Processed HR Employee Terminated: ID=%s, LegalEntity=%s", ev.EmployeeID, ev.LegalEntityID)
			return nil

		case domain.TopicCrmSalesOrderConfirmed:
			var ev domain.CrmSalesOrderConfirmedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("[Idempotent Inbox] Processed CRM Sales Order Confirmed: SalesOrder=%s, Customer=%s, LegalEntity=%s", ev.SalesOrderID, ev.CustomerID, ev.LegalEntityID)

			_, err := c.projTrackingSvc.InitializeProject(
				txCtx,
				ev.LegalEntityID,
				ev.CustomerID,
				"PRJ-"+ev.SalesOrderID,
				"Project Fulfilling Sales Order "+ev.SalesOrderID,
				domain.BillingMethodTIME_AND_MATERIALS,
				time.Now(),
			)
			return err
		}
		return nil
	})
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
