package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/business/service"
	"github.com/segmentio/kafka-go"
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
	TopicMfgProductionScheduledDeadLetter = domain.TopicMfgProductionScheduled + ".dead-letter"
	TopicScmTrainingRequiredDeadLetter    = domain.TopicScmTrainingRequired + ".dead-letter"
)

type KafkaConsumer struct {
	reader    *kafka.Reader
	publisher *KafkaPublisher
	training  *service.TrainingService
}

func NewKafkaConsumer(
	brokers []string,
	groupID string,
	publisher *KafkaPublisher,
	training *service.TrainingService,
) *KafkaConsumer {
	topics := []string{
		// TODO: connect when PM event handler is implemented (currently log-only)
		// domain.TopicPrjProjectCreated,
		// TODO: connect when PM event handler is implemented (currently log-only)
		// domain.TopicPrjTaskAssigned,
		// TODO: connect when fm/fin publishes fin.budget.allocated
		// domain.TopicFinBudgetAllocated,
		domain.TopicMfgProductionScheduled,
		domain.TopicScmTrainingRequired,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		GroupTopics: topics,
	})

	return &KafkaConsumer{
		reader:    reader,
		publisher: publisher,
		training:  training,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Println("Starting Kafka Event Consumer for hr-service...")
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
		ServiceName:   "hr-service",
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
	// TODO: connect when topic has real handler implementation
	/*
	case domain.TopicPrjProjectCreated:
		var ev domain.ProjectCreatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Project Created: assigning resource buffer for Project %s (%s)", ev.ProjectID, ev.Name)
		return nil

	case domain.TopicPrjTaskAssigned:
		var ev domain.TaskAssignedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Task Assigned: updating employee workload for Employee %s, Task %s, Workload %d hours", ev.EmployeeID, ev.TaskID, ev.Workload)
		return nil
	*/

	// TODO: connect when fm/fin publishes fin.budget.allocated
	/*
	case domain.TopicFinBudgetAllocated:
		var ev domain.BudgetAllocatedEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Budget Allocated: updating salary budgets for Dept %s, Allocated Amount: %s, Period: %s", ev.DepartmentID, ev.Amount.String(), ev.Period)
		return nil
	*/

	case domain.TopicMfgProductionScheduled:
		var ev domain.ProductionScheduledEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing Production Scheduled: scheduling workforce for workstation %s, required staff: %d", ev.Workstation, ev.RequiredStaff)
		return nil

	case domain.TopicScmTrainingRequired:
		var ev domain.SCMTrainingRequiredEvent
		if err := json.Unmarshal(value, &ev); err != nil {
			return err
		}
		log.Printf("Processing SCM Training Required: auto-scheduling training program for topic: %s, deadline: %s", ev.Topic, ev.Deadline.String())
		
		title := "SCM Required Training: " + ev.Topic
		description := "Automated mandatory training scheduled due to supply chain requirement for department " + ev.DepartmentID
		trainer := "SCM Technical Specialist"
		
		_, err := c.training.CreateTrainingProgram(ctx, title, description, trainer, time.Now(), ev.Deadline)
		return err
	}

	return nil
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
