package service

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
	"github.com/erp-system/hr-service/internal/data/memory"
)

// MockPublisher records published events without contacting Kafka.
type MockPublisher struct{}

func (m *MockPublisher) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func newTrainingService(t *testing.T) (*TrainingService, *memory.MemoryTrainingProgramRepo, *memory.MemoryTrainingEnrollmentRepo) {
	t.Helper()
	tpRepo := memory.NewMemoryTrainingProgramRepo()
	teRepo := memory.NewMemoryTrainingEnrollmentRepo()
	svc := NewTrainingService(tpRepo, teRepo, &MockPublisher{})
	return svc, tpRepo, teRepo
}

// TestEnrollEmployee_PreventsDuplicate is the regression test for Phase S4.6.
// Previously EnrollEmployee called Create() directly with no existence check,
// allowing the same employee to be enrolled in the same training program
// multiple times.
func TestEnrollEmployee_PreventsDuplicate(t *testing.T) {
	svc, tpRepo, _ := newTrainingService(t)
	ctx := context.Background()

	// Seed a training program.
	if err := tpRepo.Create(ctx, &domain.TrainingProgram{
		ID:        "train_1",
		Title:     "Go Concurrency",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		Status:    "SCHEDULED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed training: %v", err)
	}

	// First enrollment: should succeed.
	first, err := svc.EnrollEmployee(ctx, "train_1", "emp_42")
	if err != nil {
		t.Fatalf("first enroll: %v", err)
	}
	if first.Status != "ENROLLED" {
		t.Errorf("expected status ENROLLED, got %s", first.Status)
	}

	// Second enrollment for same (training, employee): must fail.
	_, err = svc.EnrollEmployee(ctx, "train_1", "emp_42")
	if err == nil {
		t.Errorf("expected error on duplicate enrollment, got nil")
	}

	// Different employee, same training: must succeed.
	_, err = svc.EnrollEmployee(ctx, "train_1", "emp_99")
	if err != nil {
		t.Errorf("different employee should be allowed, got: %v", err)
	}

	// Same employee, different training: must succeed.
	if err := tpRepo.Create(ctx, &domain.TrainingProgram{
		ID:        "train_2",
		Title:     "Kubernetes Basics",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		Status:    "SCHEDULED",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed train_2: %v", err)
	}
	_, err = svc.EnrollEmployee(ctx, "train_2", "emp_42")
	if err != nil {
		t.Errorf("same employee in different training should be allowed, got: %v", err)
	}
}

// TestEnrollEmployee_AllowsReEnrollmentAfterCancellation verifies the
// re-enrollment policy: once a prior enrollment is CANCELLED or COMPLETED,
// the employee can enroll again.
func TestEnrollEmployee_AllowsReEnrollmentAfterCancellation(t *testing.T) {
	svc, _, teRepo := newTrainingService(t)
	ctx := context.Background()

	// Seed via service so we go through EnrollEmployee.
	first, err := svc.EnrollEmployee(ctx, "train_1", "emp_42")
	if err != nil {
		t.Fatalf("first enroll: %v", err)
	}

	// Manually mark prior enrollment as CANCELLED via the repo.
	first.Status = "CANCELLED"
	if err := teRepo.Update(ctx, first); err != nil {
		t.Fatalf("update to cancelled: %v", err)
	}

	// Re-enroll: should succeed because prior is CANCELLED.
	second, err := svc.EnrollEmployee(ctx, "train_1", "emp_42")
	if err != nil {
		t.Errorf("re-enroll after cancellation should be allowed, got: %v", err)
	}
	if second.ID == first.ID {
		t.Errorf("expected new enrollment ID, got same ID %s", second.ID)
	}
}
