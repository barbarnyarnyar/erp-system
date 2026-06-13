package kafka

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type DeadLetterMessage struct {
	OriginalTopic string      `json:"original_topic"`
	OriginalKey   string      `json:"original_key,omitempty"`
	Payload       interface{} `json:"payload"`
	Error         string      `json:"error"`
	FailedAt      time.Time   `json:"failed_at"`
	ServiceName   string      `json:"service_name"`
}

const (
	TopicCrmCustomerDemandForecastDeadLetter = domain.TopicCrmCustomerDemandForecast + ".dead-letter"
	TopicMfgMaterialRequiredDeadLetter       = domain.TopicMfgMaterialRequired + ".dead-letter"
	TopicMfgMaterialConsumedDeadLetter       = domain.TopicMfgMaterialConsumed + ".dead-letter"
	TopicMfgProductionCompletedDeadLetter    = domain.TopicMfgProductionCompleted + ".dead-letter"
	TopicPrjMaterialRequestedDeadLetter      = domain.TopicPrjMaterialRequested + ".dead-letter"
)

type KafkaConsumer struct {
	reader    *kafka.Reader
	publisher domain.EventPublisher
	poSvc     *service.PurchaseOrderService
	invSvc    *service.InventoryService
	demandSvc *service.DemandPlanningService
	inbox     domain.KafkaEventInboxRepository
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	poSvc *service.PurchaseOrderService,
	invSvc *service.InventoryService,
	demandSvc *service.DemandPlanningService,
	inbox domain.KafkaEventInboxRepository,
) *KafkaConsumer {
	topics := []string{
		domain.TopicCrmSalesOrderCreated,
		domain.TopicCrmCustomerDemandForecast,
		domain.TopicMfgMaterialRequired,
		domain.TopicMfgMaterialConsumed,
		domain.TopicMfgProductionCompleted,
		domain.TopicFinVendorPaymentProcessed,
		domain.TopicPrjMaterialRequested,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		publisher: publisher,
		poSvc:     poSvc,
		invSvc:    invSvc,
		demandSvc: demandSvc,
		inbox:     inbox,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for scm-service...")
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

			log.Printf("[SCM-CONSUMER] Received event on topic %s, key %s", msg.Topic, string(msg.Key))
			if err := c.handleMessage(ctx, msg.Topic, msg.Value); err != nil {
				log.Printf("[SCM-CONSUMER] Failed to process event %s: %v", msg.Topic, err)
				c.publishToDLQ(ctx, msg.Topic, string(msg.Key), msg.Value, err)
			}
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
		ServiceName:   "scm-service",
	}
	dlqTopic := topic + ".dead-letter"
	if dlqErr := c.publisher.Publish(ctx, dlqTopic, key, dlqMsg); dlqErr != nil {
		log.Printf("ERROR: failed to publish DLQ message for topic %s: %v", topic, dlqErr)
	} else {
		log.Printf("ERROR: consumer handler failed for topic %s: %v — sent to DLQ topic %s", topic, err, dlqTopic)
	}
}

type rawEventEnvelope struct {
	EventID string `json:"event_id"`
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, topic string, value []byte) error {
	// 1. Idempotency Check
	var env rawEventEnvelope
	if err := json.Unmarshal(value, &env); err == nil && env.EventID != "" {
		if existing, err := c.inbox.GetByID(ctx, env.EventID); err == nil && existing != nil {
			log.Printf("[SCM-CONSUMER] Event %s already processed. Skipping.", env.EventID)
			return nil
		}
	} else {
		h := sha256.New()
		h.Write([]byte(topic))
		h.Write(value)
		env.EventID = fmt.Sprintf("evt_%x", h.Sum(nil))[:36]

		if existing, err := c.inbox.GetByID(ctx, env.EventID); err == nil && existing != nil {
			log.Printf("[SCM-CONSUMER] Event %s (hashed from payload) already processed. Skipping.", env.EventID)
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

	if inboxErr := c.inbox.Create(ctx, inboxRec); inboxErr != nil {
		log.Printf("[SCM-CONSUMER] Failed to record event %s in inbox: %v", env.EventID, inboxErr)
	}

	return err
}

func (c *KafkaConsumer) processEvent(ctx context.Context, topic string, value []byte) error {
	switch topic {
	case domain.TopicCrmSalesOrderCreated:
		var ev domain.SalesOrderCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Sales Order Created: creating pick list for Order %s, Customer: %s", ev.OrderNumber, ev.CustomerID)
		return nil

	case domain.TopicCrmCustomerDemandForecast:
		var ev domain.CustomerDemandForecastEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Customer Demand Forecast: Product %s, forecast date: %s, quantity: %d", ev.ProductID, ev.ForecastDate.String(), ev.ForecastQuantity)
		_, err := c.demandSvc.CreateForecast(ctx, ev.ProductID, ev.ForecastDate, ev.ForecastQuantity, ev.ConfidenceLevel, "Auto-created from customer demand forecast event")
		return err

	case domain.TopicMfgMaterialRequired:
		var ev domain.MaterialRequiredEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Manufacturing Material Required: Product %s, required qty: %d", ev.MaterialID, ev.Quantity)
		// Generate auto purchase requisition
		line := service.RequisitionLineInput{
			ProductID:          ev.MaterialID,
			QuantityRequested:  ev.Quantity,
			EstimatedUnitPrice: decimal.NewFromFloat(50.00),
		}
		_, err := c.poSvc.CreatePurchaseRequisition(ctx, "mfg-system", ev.RequiredBy, "Auto-generated from mfg.material.required event", []service.RequisitionLineInput{line})
		return err

	case domain.TopicMfgMaterialConsumed:
		var ev domain.MaterialConsumedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Material Consumed (WIP issue): Product %s, quantity consumed: %s", ev.ProductID, ev.Quantity.String())
		qtyInt := int(ev.Quantity.IntPart())
		if qtyInt == 0 && !ev.Quantity.IsZero() {
			qtyInt = 1
		}
		_, err := c.invSvc.AdjustInventory(ctx, ev.ProductID, "loc_default", qtyInt, "ISSUE", "Raw material issued for manufacturing production order "+ev.ProductionOrderID)
		return err

	case domain.TopicMfgProductionCompleted:
		var ev domain.ProductionCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Production Completed: Product %s, quantity produced: %d", ev.ProductID, ev.QuantityProduced)
		// Receive finished goods into inventory (default location loc_default)
		_, err := c.invSvc.AdjustInventory(ctx, ev.ProductID, "loc_default", ev.QuantityProduced, "RECEIPT", "Finished goods receipt from manufacturing completed")
		return err

	case domain.TopicFinVendorPaymentProcessed:
		var ev domain.VendorPaymentProcessedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Vendor Payment Processed: Vendor ID %s, payment amount: %s, status: %s", ev.VendorID, ev.AmountPaid.String(), ev.Status)
		return nil

	case domain.TopicPrjMaterialRequested:
		var ev domain.MaterialRequestedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Project Material Requested: Project %s, Task %s, Product %s, qty: %d", ev.ProjectID, ev.TaskID, ev.ProductID, ev.QtyRequired)
		// Reserve/deduct materials for project request
		_, err := c.invSvc.AdjustInventory(ctx, ev.ProductID, "loc_default", ev.QtyRequired, "ISSUE", "Reserve project materials")
		return err
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
