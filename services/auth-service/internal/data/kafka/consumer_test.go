package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/erp-system/auth-service/internal/business/domain"
	"github.com/erp-system/auth-service/internal/business/service"
	"github.com/erp-system/auth-service/internal/data/memory"
)

func newConsumerWithUser(t *testing.T, userID string) (*KafkaConsumer, *memory.UserRepository) {
	t.Helper()
	userRepo := memory.NewUserRepository()
	usRepo := memory.NewUserStoreRepository()
	urRepo := memory.NewUserRoleRepository()
	userSvc := service.NewUserService(userRepo, usRepo, urRepo, &silentPub{})

	// Create a user that maps to the HR employee.
	u := &domain.User{
		ID:            userID,
		Username:      "frank",
		Email:         "frank@example.com",
		PasswordHash:  "password",
		FirstName:     "Frank",
		LastName:      "F",
		Status:        domain.UserStatusACTIVE,
		SecurityStamp: "ss_initial",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	publisher := &silentPub{}
	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "auth-service-test", publisher, userSvc)
	return consumer, userRepo
}

type silentPub struct{}

func (s *silentPub) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

// TestConsumer_HREmployeeTerminated_DeactivatesUser is the regression test
// for Phase S4.8: when HR publishes a termination event, Auth must
// deactivate the user (which also bumps the security_stamp).
func TestConsumer_HREmployeeTerminated_DeactivatesUser(t *testing.T) {
	consumer, userRepo := newConsumerWithUser(t, "user_frank_42")
	ctx := context.Background()

	ev := domain.HREmployeeTerminatedEvent{
		EmployeeID: "user_frank_42",
		TermDate:   time.Now(),
		Reason:     "Voluntary departure",
		Timestamp:  time.Now(),
	}
	value, _ := json.Marshal(ev)

	if err := consumer.handleMessage(ctx, domain.TopicHrEmployeeTerminated, value); err != nil {
		t.Fatalf("handleMessage: %v", err)
	}

	fresh, err := userRepo.GetByID(ctx, "user_frank_42")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if fresh.Status == domain.UserStatusACTIVE {
		t.Errorf("expected IsActive=false after HR termination event, got true")
	}
	if fresh.SecurityStamp == "ss_initial" {
		t.Errorf("expected security_stamp to be bumped, still %q", fresh.SecurityStamp)
	}
}

// TestConsumer_HREmployeeTerminated_UnknownUserReturnsError verifies that
// if the HR event references a user Auth has never seen, the consumer
// returns an error (so the message can be retried or DLQ'd).
func TestConsumer_HREmployeeTerminated_UnknownUserReturnsError(t *testing.T) {
	consumer, _ := newConsumerWithUser(t, "user_existing")
	ctx := context.Background()

	ev := domain.HREmployeeTerminatedEvent{
		EmployeeID: "user_does_not_exist",
		TermDate:   time.Now(),
		Reason:     "Test",
		Timestamp:  time.Now(),
	}
	value, _ := json.Marshal(ev)

	if err := consumer.handleMessage(ctx, domain.TopicHrEmployeeTerminated, value); err == nil {
		t.Errorf("expected error for unknown user, got nil")
	}
}

func TestConsumer_StartAndClose(t *testing.T) {
	consumer, _ := newConsumerWithUser(t, "user_existing")

	// Test Start with already canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	consumer.Start(canceledCtx)

	// Test Start with context canceled after a short delay, triggering ReadMessage and ctx cancellation
	ctx, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(1 * time.Second)
		cancel2()
	}()
	consumer.Start(ctx)

	// Test Close
	_ = consumer.Close()
}

func TestConsumer_EdgeCases(t *testing.T) {
	consumer, _ := newConsumerWithUser(t, "user_existing")
	ctx := context.Background()

	// 1. Invalid JSON payload
	if err := consumer.handleMessage(ctx, domain.TopicHrEmployeeTerminated, []byte("invalid-json")); err == nil {
		t.Errorf("expected error for invalid json, got nil")
	}

	// 2. Unknown topic
	if err := consumer.handleMessage(ctx, "unknown-topic", []byte("{}")); err != nil {
		t.Errorf("expected no error for unknown topic, got %v", err)
	}
}
