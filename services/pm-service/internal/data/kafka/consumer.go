package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
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
	TopicCrmSalesOrderReceivedDeadLetter = domain.TopicCrmSalesOrderReceived + ".dead-letter"
)

type KafkaConsumer struct {
	reader      *kafka.Reader
	publisher   domain.EventPublisher
	planningSvc *service.ProjectPlanningService
	taskSvc     *service.TaskManagementService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher domain.EventPublisher,
	planningSvc *service.ProjectPlanningService,
	taskSvc *service.TaskManagementService,
) *KafkaConsumer {
	topics := []string{
		// TODO: connect when handler does real work (currently log-only)
		// domain.TopicHrEmployeeAvailable,
		// TODO: connect when handler does real work (currently log-only)
		// domain.TopicHrEmployeeSkillsUpdated,
		// TODO: connect when fm/fin publishes fm.budget.approved
		// domain.TopicFinBudgetApproved,
		// TODO: connect when handler does real work (currently log-only)
		// domain.TopicFinPaymentReceived,
		domain.TopicCrmSalesOrderReceived,
		// TODO: connect when handler does real work (currently log-only)
		// domain.TopicScmMaterialDelivered,
		// TODO: connect when mfg/m publishes mfg.custom.production.completed
		// domain.TopicMfgCustomProductionCompleted,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:      reader,
		publisher:   publisher,
		planningSvc: planningSvc,
		taskSvc:     taskSvc,
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
	switch topic {
	// TODO: connect when handler does real work (currently log-only)
	/*
		case domain.TopicHrEmployeeAvailable:
			var ev domain.EmployeeAvailableEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("Processing HR Employee Available: Employee %s is %s. Updating resource scheduling options.", ev.EmployeeID, ev.Status)
			return nil

		case domain.TopicHrEmployeeSkillsUpdated:
			var ev domain.EmployeeSkillsUpdatedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("Processing HR Employee Skills Updated: Employee %s skills updated to %v. Re-mapping project resource capabilities.", ev.EmployeeID, ev.Skills)
			return nil
	*/

	// TODO: connect when fm/fin publishes fin.budget.approved
	/*
		case domain.TopicFinBudgetApproved:
			var ev domain.BudgetApprovedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("Processing Finance Budget Approved: Project %s budget approved for amount %s. Updating project planning budget ceiling.", ev.ProjectID, ev.TotalBudget.String())
			return nil
	*/

	// TODO: connect when handler does real work (currently log-only)
	/*
		case domain.TopicFinPaymentReceived:
			var ev domain.PaymentReceivedEvent
			if err := json.Unmarshal(value, &ev); err != nil {
				return err
			}
			log.Printf("Processing Finance Payment Received: Project %s received payment of %s on Invoice %s. Updating billing summary.", ev.ProjectID, ev.AmountPaid.String(), ev.InvoiceID)
			return nil
	*/

	case domain.TopicCrmSalesOrderReceived:
		var ev domain.SalesOrderReceivedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing CRM Sales Order Received: Order %s for Customer %s. Automanaging custom project creation.", ev.SalesOrderID, ev.CustomerID)

		// Create a custom project automatically for this sales order
		projName := "Order Delivery Project - " + ev.SalesOrderID
		projDesc := "Automatically generated project to fulfill sales order " + ev.SalesOrderID
		startDate := time.Now()
		endDate := startDate.AddDate(0, 1, 0) // 1 month duration

		proj, err := c.planningSvc.CreateProject(ctx, projName, projDesc, startDate, &endDate, "", "")
		if err != nil {
			log.Printf("Failed to auto-create custom project: %v", err)
			return err
		}
		log.Printf("Successfully auto-created project %s (ID: %s) for Sales Order %s", proj.Name, proj.ID, ev.SalesOrderID)

		// Auto-create initial kick-off task
		_, _ = c.taskSvc.CreateTask(ctx, proj.ID, "", "Project Kick-off & Alignment", "Confirm requirements and resources for Sales Order "+ev.SalesOrderID, "", &startDate, &startDate, decimal.NewFromInt(0))
		return nil

		// TODO: connect when handler does real work (currently log-only)
		/*
			case domain.TopicScmMaterialDelivered:
				var ev domain.MaterialDeliveredEvent
				if err := json.Unmarshal(value, &ev); err != nil {
					return err
				}
				log.Printf("Processing SCM Material Delivered: Material delivered for project %s, task %s (Shipment: %s). Updating task resource status.", ev.ProjectID, ev.TaskID, ev.ShipmentID)
				return nil
		*/

		// TODO: connect when mfg/m publishes mfg.custom.production.completed
		/*
			case domain.TopicMfgCustomProductionCompleted:
				var ev domain.CustomProductionCompletedEvent
				if err := json.Unmarshal(value, &ev); err != nil {
					return err
				}
				log.Printf("Processing Manufacturing Custom Production Completed: Custom production completed for project %s, Item %s. Marking production order %s resolved.", ev.ProjectID, ev.CustomItemID, ev.ProductionOrderID)
				return nil
		*/
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
