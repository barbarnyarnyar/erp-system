package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/plm-service/internal/business/domain"
	"github.com/erp-system/plm-service/internal/business/service"
	kafkago "github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader    *kafkago.Reader
	publisher domain.EventPublisher
	matSvc    *service.MaterialService
	bomSvc    *service.BomService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	matSvc *service.MaterialService,
	bomSvc *service.BomService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicScmReceiptStaged,
		domain.TopicMfgMaterialConsumed,
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
		matSvc:    matSvc,
		bomSvc:    bomSvc,
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
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("[PLM-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicScmReceiptStaged:
		var ev struct {
			EventID       string `json:"event_id"`
			LegalEntityID string `json:"legal_entity_id"`
			PurchaseOrderID string `json:"purchase_order_id"`
			Timestamp     string `json:"timestamp"`
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
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
