package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"erp-system/shared/utils"
	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/business/service"
	"github.com/erp-system/pm-service/internal/data/sql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	utils.InitLogger("prj-consumer-test")
}

type mockPublisher struct {
	failPublish bool
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	if m.failPublish {
		return fmt.Errorf("injected publish error")
	}
	return nil
}

type testEnv struct {
	db       *gorm.DB
	consumer *KafkaConsumer
}

func setupTestEnv(t *testing.T) *testEnv {
	// Use named in-memory sqlite unique to each test to isolate databases
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	err = db.AutoMigrate(
		&sql.Project{},
		&sql.WbsNode{},
		&sql.TimeLog{},
		&sql.TransactionalOutbox{},
		&sql.KafkaEventInbox{},
	)
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	projRepo := sql.NewSQLProjectRepository(db)
	inboxRepo := sql.NewSQLKafkaEventInboxRepository(db)

	publisher := &mockPublisher{}
	projTrackingSvc := service.NewProjectTrackingService(db, projRepo)
	reliableSvc := service.NewReliableMessagingService(db, inboxRepo)

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "prj-group", publisher, reliableSvc, projTrackingSvc)

	return &testEnv{
		db:       db,
		consumer: consumer,
	}
}

func TestConsumer_AllEvents(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	// 1. TopicHrEmployeeCreated
	evCreated := domain.HrEmployeeCreatedEvent{
		EventID:       "evt-created-1",
		LegalEntityID: "tenant-1",
		EmployeeID:    "emp-1",
		ExplicitRole:  "DEVELOPER",
		Timestamp:     time.Now(),
	}
	valCreated, _ := json.Marshal(evCreated)
	if err := env.consumer.handleMessage(ctx, domain.TopicHrEmployeeCreated, valCreated); err != nil {
		t.Fatalf("failed to process HrEmployeeCreated: %v", err)
	}

	// 2. TopicHrEmployeeTerminated
	evTerminated := domain.HrEmployeeTerminatedEvent{
		EventID:       "evt-terminated-1",
		LegalEntityID: "tenant-1",
		EmployeeID:    "emp-1",
		Timestamp:     time.Now(),
	}
	valTerminated, _ := json.Marshal(evTerminated)
	if err := env.consumer.handleMessage(ctx, domain.TopicHrEmployeeTerminated, valTerminated); err != nil {
		t.Fatalf("failed to process HrEmployeeTerminated: %v", err)
	}

	// 3. TopicCrmSalesOrderConfirmed
	evConfirmed := domain.CrmSalesOrderConfirmedEvent{
		EventID:       "evt-confirmed-1",
		LegalEntityID: "tenant-1",
		SalesOrderID:  "so-123",
		CustomerID:    "cust-123",
		Timestamp:     time.Now(),
	}
	valConfirmed, _ := json.Marshal(evConfirmed)
	if err := env.consumer.handleMessage(ctx, domain.TopicCrmSalesOrderConfirmed, valConfirmed); err != nil {
		t.Fatalf("failed to process CrmSalesOrderConfirmed: %v", err)
	}

	// Verify project was initialized in the GORM DB
	var fetchedProj sql.Project
	if err := env.db.First(&fetchedProj, "project_code = ?", "PRJ-so-123").Error; err != nil {
		t.Errorf("expected project to be initialized in DB, got error: %v", err)
	}

	// Test unmarshal failures for error paths
	topics := []string{
		domain.TopicHrEmployeeCreated,
		domain.TopicHrEmployeeTerminated,
		domain.TopicCrmSalesOrderConfirmed,
	}
	for _, topic := range topics {
		if err := env.consumer.handleMessage(ctx, topic, []byte("invalid-json")); err == nil {
			t.Errorf("expected error for invalid json on topic %s", topic)
		}
		// Malformed timestamp in valid JSON to fail specific unmarshal but pass generic
		badPayload := []byte(`{"event_id":"evt-err","timestamp":"invalid-date"}`)
		if err := env.consumer.handleMessage(ctx, topic, badPayload); err == nil {
			t.Errorf("expected error for bad timestamp json on topic %s", topic)
		}
	}

	// Test unknown topic
	if err := env.consumer.handleMessage(ctx, "unknown-topic", []byte("{}")); err != nil {
		t.Fatalf("expected nil error for unknown topic, got %v", err)
	}

	// Test DLQ Publish happy path
	env.consumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("test error"))

	// Test DLQ Publish error path
	failPub := &mockPublisher{failPublish: true}
	failConsumer := NewKafkaConsumer([]string{"localhost:9092"}, "prj-group", failPub, service.NewReliableMessagingService(env.db, sql.NewSQLKafkaEventInboxRepository(env.db)), service.NewProjectTrackingService(env.db, sql.NewSQLProjectRepository(env.db)))
	failConsumer.publishToDLQ(ctx, "test-topic", "test-key", []byte("test-val"), fmt.Errorf("test error"))
}

func TestConsumer_StartAndClose(t *testing.T) {
	env := setupTestEnv(t)

	// Test Start with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	env.consumer.Start(canceledCtx)

	// Test Start with delayed cancel
	ctx, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel2()
	}()
	env.consumer.Start(ctx)

	// Test Close
	_ = env.consumer.Close()
}
