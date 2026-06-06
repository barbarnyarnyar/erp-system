package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/scm-service/internal/business/domain"
	"github.com/erp-system/scm-service/internal/business/service"
	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type KafkaConsumer struct {
	reader    *kafka.Reader
	poSvc     *service.PurchaseOrderService
	invSvc    *service.InventoryService
	demandSvc *service.DemandPlanningService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	poSvc *service.PurchaseOrderService,
	invSvc *service.InventoryService,
	demandSvc *service.DemandPlanningService,
) *KafkaConsumer {
	topics := []string{
		domain.TopicCrmSalesOrderCreated,
		domain.TopicCrmCustomerDemandForecast,
		domain.TopicMfgMaterialRequired,
		domain.TopicMfgMaterialConsumed,
		domain.TopicMfgProductionCompleted,
		// TODO: connect when fin/fm publishes fin.vendor.payment.processed
		// domain.TopicFinVendorPaymentProcessed,
		domain.TopicPrjMaterialRequested,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		poSvc:     poSvc,
		invSvc:    invSvc,
		demandSvc: demandSvc,
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
		log.Printf("[SCM-CONSUMER] Processing Sales Order Created: creating pick list for Order %s, Customer: %s", ev.OrderNumber, ev.CustomerID)
		return nil

	case domain.TopicCrmCustomerDemandForecast:
		var ev domain.CustomerDemandForecastEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Customer Demand Forecast: Product %s, forecast date: %s, quantity: %d", ev.ProductID, ev.ForecastDate.String(), ev.ForecastQuantity)
		confidenceDec := decimal.NewFromFloat(ev.ConfidenceLevel)
		_, err := c.demandSvc.CreateForecast(ctx, ev.ProductID, ev.ForecastDate, ev.ForecastQuantity, confidenceDec, "Auto-created from customer demand forecast event")
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

	// TODO: connect when fin/fm publishes fin.vendor.payment.processed
	/*
	case domain.TopicFinVendorPaymentProcessed:
		var ev domain.VendorPaymentProcessedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("[SCM-CONSUMER] Processing Vendor Payment Processed: Vendor ID %s, payment amount: %s, status: %s", ev.VendorID, ev.AmountPaid.String(), ev.Status)
		return nil
	*/

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
