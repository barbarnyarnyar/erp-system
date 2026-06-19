package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
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
	reliableSvc service.ReliableMessagingService
	execSvc     service.WorkOrderExecutionService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	reliableSvc service.ReliableMessagingService,
	execSvc service.WorkOrderExecutionService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicPlmBomReleased,
		domain.TopicQmsInspectionPassed,
		domain.TopicQmsInspectionFailed,
		domain.TopicEamMachineOffline,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:      reader,
		publisher:   publisher,
		reliableSvc: reliableSvc,
		execSvc:     execSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for mfg-service...")
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
		ServiceName:   "mfg-service",
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
	case domain.TopicPlmBomReleased:
		var ev domain.PlmBomReleasedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("Processing PLM BOM Released: BOM %s for Material %s", ev.BomHeaderID, ev.MaterialID)
			return c.execSvc.FreezeObsoleteWorkOrders(txCtx, ev.MaterialID, ev.BomHeaderID)
		})

	case domain.TopicQmsInspectionPassed:
		var ev domain.QmsInspectionPassedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("Processing QMS Inspection Passed: Inspection %s for Material %s", ev.InspectionID, ev.MaterialID)
			if ev.TriggerSource == "WORK_ORDER" && ev.SourceDocumentID != "" {
				_, err := c.execSvc.TransitionWorkOrderState(txCtx, ev.SourceDocumentID, domain.WorkOrderStateIN_PROGRESS, domain.WorkOrderStateCOMPLETED)
				if err != nil {
					log.Printf("Failed to transition work order %s to COMPLETED: %v", ev.SourceDocumentID, err)
				}
			}
			return nil
		})

	case domain.TopicQmsInspectionFailed:
		var ev domain.QmsInspectionFailedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("Processing QMS Inspection Failed: Inspection %s failed for Material %s. NC: %s", ev.InspectionID, ev.MaterialID, ev.NonConformanceID)
			if ev.TriggerSource == "WORK_ORDER" && ev.SourceDocumentID != "" {
				_, err := c.execSvc.TransitionWorkOrderState(txCtx, ev.SourceDocumentID, domain.WorkOrderStateIN_PROGRESS, domain.WorkOrderStateON_HOLD)
				if err != nil {
					log.Printf("Failed to transition work order %s to ON_HOLD: %v", ev.SourceDocumentID, err)
				}
			}
			return nil
		})

	case domain.TopicEamMachineOffline:
		var ev domain.EamMachineOfflineEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("Processing EAM Machine Offline: Equipment %s is offline. Work Order ID: %s", ev.EquipmentID, ev.WorkOrderID)
			if ev.WorkOrderID != "" {
				_, err := c.execSvc.TransitionWorkOrderState(txCtx, ev.WorkOrderID, domain.WorkOrderStateIN_PROGRESS, domain.WorkOrderStateON_HOLD)
				if err != nil {
					log.Printf("Failed to transition work order %s to ON_HOLD: %v", ev.WorkOrderID, err)
				}
			}
			return nil
		})
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
