package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	kafkago "github.com/segmentio/kafka-go"

	sharedkafka "erp-system/shared/kafka"
	"erp-system/shared/utils")

type KafkaConsumer struct {
	reader      *kafkago.Reader
	publisher   domain.EventPublisher
	reliableSvc service.ReliableMessagingService
	planSvc     *service.InspectionPlanService
	execSvc     *service.InspectionExecutionService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	reliableSvc service.ReliableMessagingService,
	planSvc *service.InspectionPlanService,
	execSvc *service.InspectionExecutionService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicScmReceiptStaged,
		domain.TopicMfgYieldProduced,
		domain.TopicHrEmployeeCreated,
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:      reader,
		publisher:   publisher,
		reliableSvc: reliableSvc,
		planSvc:     planSvc,
		execSvc:     execSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for qms-service...")
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

			log.Printf("[QMS-CONSUMER] Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			// Extract trace context and register trace ID
			msgCtx := sharedkafka.ExtractTraceContext(ctx, msg.Headers)
			traceID := utils.GetTraceIDFromContext(msgCtx)
			utils.SetTraceID(traceID)

			// Inject publisher into message context for DLQ routing in idempotent transactions
			msgCtx = context.WithValue(msgCtx, "publisher", c.publisher)

			if err := c.handleMessage(msgCtx, msg.Topic, msg.Value); err != nil {
				log.Printf("[QMS-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicScmReceiptStaged:
		var ev struct {
			EventID         string  `json:"event_id"`
			LegalEntityID   string  `json:"legal_entity_id"`
			PurchaseOrderID string  `json:"purchase_order_id"`
			MaterialID      string  `json:"material_id"`
			Quantity        float64 `json:"quantity"`
			Timestamp       string  `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[QMS-CONSUMER] SCM goods receipt staged: material %s, quantity %f. Staging quality inspection.", ev.MaterialID, ev.Quantity)
			// Try to stage inspection under default plan or find matching plan
			plan, err := c.planSvc.ConfigurePlan(txCtx, ev.LegalEntityID, ev.MaterialID, "Receiving plan for "+ev.MaterialID)
			if err != nil {
				// Plan might already exist, try retrieval
				plan, err = c.planSvc.GetPlanByMaterial(txCtx, ev.LegalEntityID, ev.MaterialID)
				if err != nil {
					// Fallback to plan_default
					plan = &domain.InspectionPlan{ID: "plan_default"}
				}
			}
			_, err = c.execSvc.StageInspection(txCtx, ev.LegalEntityID, plan.ID, domain.InspectionTriggerTypeINBOUND_RECEIPT, ev.PurchaseOrderID)
			return err
		})

	case domain.TopicMfgYieldProduced:
		var ev struct {
			EventID       string  `json:"event_id"`
			LegalEntityID string  `json:"legal_entity_id"`
			WorkOrderID   string  `json:"work_order_id"`
			MaterialID    string  `json:"material_id"`
			YieldQuantity float64 `json:"yield_quantity"`
			Timestamp     string  `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[QMS-CONSUMER] Manufacturing yield produced: material %s, quantity %f. Staging quality inspection.", ev.MaterialID, ev.YieldQuantity)
			plan, err := c.planSvc.ConfigurePlan(txCtx, ev.LegalEntityID, ev.MaterialID, "Yield plan for "+ev.MaterialID)
			if err != nil {
				plan, err = c.planSvc.GetPlanByMaterial(txCtx, ev.LegalEntityID, ev.MaterialID)
				if err != nil {
					plan = &domain.InspectionPlan{ID: "plan_default"}
				}
			}
			_, err = c.execSvc.StageInspection(txCtx, ev.LegalEntityID, plan.ID, domain.InspectionTriggerTypePRODUCTION_YIELD, ev.WorkOrderID)
			return err
		})

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
		return c.reliableSvc.ExecuteIdempotentTransaction(ctx, ev.EventID, topic, ev, func(txCtx context.Context) error {
			log.Printf("[QMS-CONSUMER] Syncing inspector %s", ev.EmployeeID)
			return nil
		})
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
