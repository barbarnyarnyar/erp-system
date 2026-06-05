package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/m-service/internal/business/domain"
	"github.com/erp-system/m-service/internal/business/service"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	prod   *service.ProductionService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	prod *service.ProductionService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicCrmSalesOrderCreated,
		domain.TopicScmMaterialReceived,
		domain.TopicScmInventoryUpdated,
		domain.TopicFinCostBudgetAllocated,
		domain.TopicHrEmployeeScheduled,
		domain.TopicPrjCustomOrderCreated,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader: reader,
		prod:   prod,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for m-service...")
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
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("Failed to process event %s: %v", msg.Topic, err)
			}
		}
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicCrmSalesOrderCreated:
		var ev domain.SalesOrderCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Sales Order Created: Auto-scheduling production order for Product: %s, Quantity: %d", ev.ProductID, ev.Quantity)

		// Auto-schedule production order using a default BOM (or fallback/mock logic)
		bomID := "bom_default"
		_, err := c.prod.CreateProductionOrder(ctx, bomID, ev.Quantity, time.Now().AddDate(0, 0, 7))
		if err != nil {
			log.Printf("Failed to auto-schedule production order for Sales Order %s: %v", ev.SalesOrderID, err)
		}
		return nil

	case domain.TopicScmMaterialReceived:
		var ev domain.SCMMaterialReceivedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing SCM Material Received: Material Product %s received for PO %s, quantity: %s. Updating material availability.", ev.ProductID, ev.PurchaseOrderID, ev.Quantity.String())
		return nil

	case domain.TopicScmInventoryUpdated:
		var ev domain.SCMInventoryUpdatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing SCM Inventory Updated: Product %s changed by type %s at location %s. New QOH: %s. Updating production material status.", ev.ProductID, ev.ChangeType, ev.LocationID, ev.QuantityOnHand.String())
		return nil

	case domain.TopicFinCostBudgetAllocated:
		var ev domain.FinCostBudgetAllocatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Financial Cost Budget Allocated: Allocated budget amount: %s to Project: %s (Dept: %s).", ev.Amount.String(), ev.ProjectID, ev.DepartmentID)
		return nil

	case domain.TopicHrEmployeeScheduled:
		var ev domain.HREmployeeScheduledEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing HR Employee Scheduled: Employee %s scheduled for Work Center %s from %s to %s. Updating labor capacity.", ev.EmployeeID, ev.WorkCenterID, ev.ShiftStart, ev.ShiftEnd)
		return nil

	case domain.TopicPrjCustomOrderCreated:
		var ev domain.PrjCustomOrderCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Project Custom Order Created: Scheduling custom production order for item: %s, quantity: %d.", ev.CustomItemID, ev.Quantity)
		
		// Auto-schedule production order for custom product
		bomID := "bom_default"
		_, err := c.prod.CreateProductionOrder(ctx, bomID, ev.Quantity, ev.RequiredBy)
		if err != nil {
			log.Printf("Failed to auto-schedule production order for Project Custom Order %s: %v", ev.ProjectID, err)
		}
		return nil
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
