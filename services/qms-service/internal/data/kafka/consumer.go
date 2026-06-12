package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/qms-service/internal/business/domain"
	"github.com/erp-system/qms-service/internal/business/service"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader    *kafkago.Reader
	publisher domain.EventPublisher
	planSvc   *service.InspectionPlanService
	execSvc   *service.InspectionExecutionService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
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
		reader:    reader,
		publisher: publisher,
		planSvc:   planSvc,
		execSvc:   execSvc,
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
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("[QMS-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicScmReceiptStaged:
		var ev struct {
			EventID         string `json:"event_id"`
			LegalEntityID   string `json:"legal_entity_id"`
			PurchaseOrderID string `json:"purchase_order_id"`
			MaterialID      string `json:"material_id"`
			Quantity        float64 `json:"quantity"`
			Timestamp       string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[QMS-CONSUMER] SCM goods receipt staged: material %s, quantity %f. Staging quality inspection.", ev.MaterialID, ev.Quantity)
		// Try to stage inspection under a default plan
		_, err := c.execSvc.StageInspection(ctx, ev.LegalEntityID, "plan_default", domain.InspectionTriggerTypeINBOUND_RECEIPT, ev.PurchaseOrderID)
		return err

	case domain.TopicMfgYieldProduced:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			WorkOrderID   string `json:"work_order_id"`
			MaterialID    string `json:"material_id"`
			YieldQuantity float64 `json:"yield_quantity"`
			Timestamp     string `json:"timestamp"`
		}
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[QMS-CONSUMER] Manufacturing yield produced: material %s, quantity %f. Staging quality inspection.", ev.MaterialID, ev.YieldQuantity)
		_, err := c.execSvc.StageInspection(ctx, ev.LegalEntityID, "plan_default", domain.InspectionTriggerTypePRODUCTION_YIELD, ev.WorkOrderID)
		return err

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
		log.Printf("[QMS-CONSUMER] Syncing inspector %s", ev.EmployeeID)
		return nil
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
