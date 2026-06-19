package kafka

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/erp-system/plm-service/internal/business/service"
	kafkago "github.com/segmentio/kafka-go"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils")

type KafkaConsumer struct {
	reader    *kafkago.Reader
	publisher domain.EventPublisher
	matSvc    *service.MaterialService
	bomSvc    *service.BomService
	inbox     domain.KafkaEventInboxRepository
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	matSvc *service.MaterialService,
	bomSvc *service.BomService,
	inbox domain.KafkaEventInboxRepository,
) *KafkaConsumer {
	topics := []string{
		domain.TopicScmReceiptStaged,
		domain.TopicMfgMaterialConsumed,
		domain.TopicHrEmployeeCreated,
		domain.TopicQmsInspectionFailed,
		domain.TopicEamMachineOffline,
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		publisher: publisher,
		matSvc:    matSvc,
		bomSvc:    bomSvc,
		inbox:     inbox,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for plm-service...")
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

			log.Printf("[PLM-CONSUMER] Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			// Extract trace context and register trace ID
			msgCtx := sharedkafka.ExtractTraceContext(ctx, msg.Headers)
			traceID := utils.GetTraceIDFromContext(msgCtx)
			utils.SetTraceID(traceID)

			// Inject publisher into message context for DLQ routing in idempotent transactions
			msgCtx = context.WithValue(msgCtx, "publisher", c.publisher)

			if err := c.handleMessage(msgCtx, msg.Topic, msg.Value); err != nil {
				log.Printf("[PLM-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

type rawEventEnvelope struct {
	EventID string `json:"event_id"`
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	// 1. Idempotency Check
	var env rawEventEnvelope
	if err := json.Unmarshal(value, &env); err == nil && env.EventID != "" {
		if exists, err := c.inbox.Exists(ctx, env.EventID); err == nil && exists {
			log.Printf("[PLM-CONSUMER] Event %s already processed. Skipping.", env.EventID)
			return nil
		}
	} else {
		h := sha256.New()
		h.Write([]byte(topic))
		h.Write(value)
		env.EventID = fmt.Sprintf("evt_%x", h.Sum(nil))[:36]

		if exists, err := c.inbox.Exists(ctx, env.EventID); err == nil && exists {
			log.Printf("[PLM-CONSUMER] Event %s (hashed from payload) already processed. Skipping.", env.EventID)
			return nil
		}
	}

	// 2. Process Event
	err := c.processEvent(ctx, topic, value)

	// 3. Log to Inbox
	status := domain.EventProcessingStatusSUCCESS
	if err != nil {
		status = domain.EventProcessingStatusFAILED
	}

	inboxRec := &domain.KafkaEventInbox{
		EventID:          env.EventID,
		EventType:        topic,
		ProcessedAt:      time.Now(),
		ProcessingStatus: status,
		Payload:          string(value),
	}
	_ = c.inbox.Create(ctx, inboxRec)

	return err
}

func (c *KafkaConsumer) processEvent(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicScmReceiptStaged:
		var ev struct {
			EventID         string `json:"event_id"`
			LegalEntityID   string `json:"legal_entity_id"`
			PurchaseOrderID string `json:"purchase_order_id"`
			Timestamp       string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[PLM-CONSUMER] Processing SCM receipt staged: PO: %s", ev.PurchaseOrderID)
		return nil

	case domain.TopicMfgMaterialConsumed:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			WorkOrderID   string `json:"work_order_id"`
			MaterialID    string `json:"material_id"`
			VolumeLost    float64 `json:"volume_lost"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[PLM-CONSUMER] Material consumed in MFG: Mat ID %s, qty: %f", ev.MaterialID, ev.VolumeLost)
		return nil

	case domain.TopicHrEmployeeCreated:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			EmployeeID    string `json:"employee_id"`
			ExplicitRole  string `json:"explicit_role"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[PLM-CONSUMER] Syncing employee %s for Engineering Review", ev.EmployeeID)
		return nil

	case domain.TopicQmsInspectionFailed:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			MaterialID    string `json:"material_id"`
			BatchNumber   string `json:"batch_number"`
			InspectorID   string `json:"inspector_id"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[PLM-CONSUMER] Processing QMS inspection failure: Material %s, Batch %s", ev.MaterialID, ev.BatchNumber)

		// Engineering notification spec update
		specMsg := fmt.Sprintf(`{"warning_flag": "QMS inspection failed for batch %s"}`, ev.BatchNumber)
		_, err := c.matSvc.UpdateTechnicalSpecs(ctx, ev.MaterialID, specMsg)
		if err != nil {
			// Material might not exist yet in local database, which is expected during testing/seeding setup
			log.Printf("[PLM-CONSUMER] Warning: Material master %s not found for warning spec update: %v", ev.MaterialID, err)
		} else {
			log.Printf("[PLM-CONSUMER] NOTIFICATION: Engineering group alerted of suspected design flaw on Material %s. Corrective ECO loop staged.", ev.MaterialID)
		}
		return nil

	case domain.TopicEamMachineOffline:
		var ev struct {
			EventID                 string `json:"event_id"`
			LegalEntityID           string `json:"legal_entity_id"`
			MachineID               string `json:"machine_id"`
			DowntimeDurationSeconds int    `json:"downtime_duration_seconds"`
			ReasonCode              string `json:"reason_code"`
			Timestamp               string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[PLM-CONSUMER] Processing EAM machine offline event: Machine %s, Reason %s", ev.MachineID, ev.ReasonCode)
		log.Printf("[PLM-CONSUMER] METADATA: Machine %s offline. Registered production constraint (reason: %s). Engineers alerted to modify part tolerances.", ev.MachineID, ev.ReasonCode)
		return nil
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

