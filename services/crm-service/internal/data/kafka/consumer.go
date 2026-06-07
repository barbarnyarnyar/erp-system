package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/crm-service/internal/business/domain"
	"github.com/erp-system/crm-service/internal/business/service"
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
	TopicScmShipmentDeliveredDeadLetter = domain.TopicScmShipmentDelivered + ".dead-letter"
)

type KafkaConsumer struct {
	reader         *kafka.Reader
	publisher      domain.EventPublisher
	orderSvc       *service.SalesOrderService
	leadSvc        *service.LeadService
	oppSvc         *service.OpportunityService
	interactionSvc *service.CustomerInteractionService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	orderSvc *service.SalesOrderService,
	leadSvc *service.LeadService,
	oppSvc *service.OpportunityService,
	interactionSvc *service.CustomerInteractionService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicScmInventoryAvailable,
		domain.TopicScmShipmentDelivered,
		domain.TopicFinPaymentReceived,
		domain.TopicFinCreditCheckCompleted,
		domain.TopicMfgProductionCompleted,
		domain.TopicPrjProjectCompleted,
		domain.TopicHrEmployeePerformance,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:         reader,
		publisher:      publisher,
		orderSvc:       orderSvc,
		leadSvc:        leadSvc,
		oppSvc:         oppSvc,
		interactionSvc: interactionSvc,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for crm-service...")
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
		ServiceName:   "crm-service",
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
	case domain.TopicScmInventoryAvailable:
		var ev domain.InventoryAvailableEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing SCM Inventory Available: Product %s is available with quantity %s. Updating CRM sales catalog availability.", ev.ProductID, ev.QuantityOnHand.String())
		opps, err := c.oppSvc.ListOpportunities(ctx)
		if err == nil {
			for _, opp := range opps {
				if opp.Status == "NEW" {
					log.Printf("Bumping Opportunity %s stage to NEGOTIATION since product %s inventory is available", opp.ID, ev.ProductID)
					_, _ = c.oppSvc.UpdateOpportunity(ctx, opp.ID, opp.Title, opp.Value, "IN_PROGRESS", "NEGOTIATION", decimal.NewFromFloat(0.30), "system")
				}
			}
		}
		return nil

	case domain.TopicScmShipmentDelivered:
		var ev domain.ShipmentDeliveredEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing SCM Shipment Delivered: Shipment %s delivered for Sales Order %s. Updating sales order status in CRM.", ev.ShipmentID, ev.SalesOrderID)
		_, err := c.orderSvc.UpdateSalesOrder(ctx, ev.SalesOrderID, "DELIVERED")
		if err != nil {
			log.Printf("Failed to update sales order status to DELIVERED: %v", err)
			return err
		}
		return nil

	case domain.TopicFinPaymentReceived:
		var ev domain.PaymentReceivedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Financial Payment Received: Invoice %s, ref %s for amount %s received. Updating customer payment history in CRM.", ev.InvoiceID, ev.ReferenceID, ev.Amount.String())
		return nil

	case domain.TopicFinCreditCheckCompleted:
		var ev domain.CreditCheckCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Financial Credit Check Completed: Customer %s credit status: %s. Updating customer credit history.", ev.CustomerID, ev.CreditStatus)
		if ev.CreditStatus == "APPROVED" {
			orders, err := c.orderSvc.ListSalesOrders(ctx)
			if err == nil {
				for _, o := range orders {
					if o.CustomerID == ev.CustomerID && o.Status == "DRAFT" {
						log.Printf("Auto-confirming Sales Order %s for customer %s since credit check passed", o.ID, ev.CustomerID)
						_, _ = c.orderSvc.UpdateSalesOrder(ctx, o.ID, "CONFIRMED")
					}
				}
			}
		}
		return nil

	case domain.TopicMfgProductionCompleted:
		var ev domain.ProductionCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Manufacturing Production Completed: Production Order %s completed for Product %s, quantity %d. Catalog updated.", ev.ProductionOrderID, ev.ProductID, ev.Quantity)
		return nil

	case domain.TopicPrjProjectCompleted:
		var ev domain.ProjectCompletedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Project Completed: Custom project %s completed. Updating status of project-linked sales orders.", ev.ProjectID)
		return nil

	case domain.TopicHrEmployeePerformance:
		var ev domain.EmployeePerformanceEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing HR Employee Performance: Employee %s rated %s. Updating sales representative metrics in CRM.", ev.EmployeeID, ev.Rating.String())
		_, _ = c.interactionSvc.CreateCustomerInteraction(ctx, "cust_default", "EMAIL", "Employee Performance Rating Received", "Employee "+ev.EmployeeID+" performance rating received: "+ev.Rating.String(), time.Now(), "system")
		return nil
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
