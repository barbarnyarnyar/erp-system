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
		ID:           userID,
		Username:     "frank",
		Email:        "frank@example.com",
		PasswordHash: "password",
		FirstName:    "Frank",
		LastName:     "F",
		IsActive:     true,
		SecurityStamp: "ss_initial",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	publisher := NewKafkaPublisher([]string{"localhost:9092"})
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
	if fresh.IsActive {
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
