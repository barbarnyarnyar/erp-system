package service

import (
	"context"
	"testing"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
	"github.com/erp-system/pm-service/internal/data/memory"
)

type silentPub struct{}

func (s *silentPub) Publish(ctx context.Context, topic string, key string, payload interface{}) error {
	return nil
}

func newPlanningService(t *testing.T) (*ProjectPlanningService, *memory.ProjectRepository, *memory.MilestoneRepository) {
	t.Helper()
	portRepo := memory.NewPortfolioRepository()
	projRepo := memory.NewProjectRepository()
	msRepo := memory.NewMilestoneRepository()
	svc := NewProjectPlanningService(portRepo, projRepo, msRepo, &silentPub{})
	return svc, projRepo, msRepo
}

// TestCreateMilestone_DefaultsToPending is the happy path: a new milestone
// starts in PENDING status with the given target date and timestamps.
func TestCreateMilestone_DefaultsToPending(t *testing.T) {
	svc, projRepo, _ := newPlanningService(t)
	ctx := context.Background()

	// Seed a project so the FK check passes.
	if err := projRepo.Create(ctx, &domain.Project{
		ID:        "proj_1",
		Name:      "Test Project",
		StartDate: time.Now(),
		Status:    "PLANNING",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	target := time.Now().Add(7 * 24 * time.Hour)
	m, err := svc.CreateMilestone(ctx, "proj_1", "MVP", "Initial release", &target)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if m.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %s", m.Status)
	}
	if m.ProjectID != "proj_1" {
		t.Errorf("expected project_id proj_1, got %s", m.ProjectID)
	}
	if m.TargetDate == nil || !m.TargetDate.Equal(target) {
		t.Errorf("expected target_date=%v, got %v", target, m.TargetDate)
	}
	if m.CompletionDate != nil {
		t.Errorf("expected completion_date to be nil, got %v", m.CompletionDate)
	}
}

// TestCompleteMilestone_SetsStatusAndDate verifies that completing a
// milestone transitions it to ACHIEVED with a completion date set.
func TestCompleteMilestone_SetsStatusAndDate(t *testing.T) {
	svc, projRepo, _ := newPlanningService(t)
	ctx := context.Background()

	_ = projRepo.Create(ctx, &domain.Project{
		ID: "proj_2", Name: "P2", StartDate: time.Now(), Status: "ACTIVE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	target := time.Now().Add(24 * time.Hour)
	m, err := svc.CreateMilestone(ctx, "proj_2", "Alpha", "", &target)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	completedAt := time.Now()
	updated, err := svc.CompleteMilestone(ctx, m.ID, completedAt)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if updated.Status != "ACHIEVED" {
		t.Errorf("expected status ACHIEVED, got %s", updated.Status)
	}
	if updated.CompletionDate == nil {
		t.Fatalf("expected completion_date to be set")
	}
	if !updated.CompletionDate.Equal(completedAt) {
		t.Errorf("expected completion_date=%v, got %v", completedAt, updated.CompletionDate)
	}
}

// TestCompleteMilestone_IdempotencyCheck verifies that completing an
// already-achieved milestone returns an error (idempotent rejection).
func TestCompleteMilestone_IdempotencyCheck(t *testing.T) {
	svc, projRepo, _ := newPlanningService(t)
	ctx := context.Background()

	_ = projRepo.Create(ctx, &domain.Project{
		ID: "proj_3", Name: "P3", StartDate: time.Now(), Status: "ACTIVE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	m, _ := svc.CreateMilestone(ctx, "proj_3", "Beta", "", nil)
	_, err := svc.CompleteMilestone(ctx, m.ID, time.Now())
	if err != nil {
		t.Fatalf("first complete should succeed, got: %v", err)
	}
	_, err = svc.CompleteMilestone(ctx, m.ID, time.Now())
	if err == nil {
		t.Errorf("second complete should fail, got nil")
	}
}

// TestDelayMilestone_SetsNewTargetDate verifies that delaying a milestone
// updates its target date and sets status to DELAYED.
func TestDelayMilestone_SetsNewTargetDate(t *testing.T) {
	svc, projRepo, _ := newPlanningService(t)
	ctx := context.Background()

	_ = projRepo.Create(ctx, &domain.Project{
		ID: "proj_4", Name: "P4", StartDate: time.Now(), Status: "ACTIVE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	orig := time.Now().Add(7 * 24 * time.Hour)
	m, _ := svc.CreateMilestone(ctx, "proj_4", "Gamma", "", &orig)

	newTarget := time.Now().Add(14 * 24 * time.Hour)
	updated, err := svc.DelayMilestone(ctx, m.ID, newTarget)
	if err != nil {
		t.Fatalf("delay: %v", err)
	}
	if updated.Status != "DELAYED" {
		t.Errorf("expected status DELAYED, got %s", updated.Status)
	}
	if updated.TargetDate == nil || !updated.TargetDate.Equal(newTarget) {
		t.Errorf("expected target_date=%v, got %v", newTarget, updated.TargetDate)
	}
}

// TestListMilestonesByProject_FiltersByProject verifies that the list
// endpoint only returns milestones for the requested project.
func TestListMilestonesByProject_FiltersByProject(t *testing.T) {
	svc, projRepo, _ := newPlanningService(t)
	ctx := context.Background()

	_ = projRepo.Create(ctx, &domain.Project{
		ID: "proj_A", Name: "A", StartDate: time.Now(), Status: "ACTIVE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	_ = projRepo.Create(ctx, &domain.Project{
		ID: "proj_B", Name: "B", StartDate: time.Now(), Status: "ACTIVE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})

	_, _ = svc.CreateMilestone(ctx, "proj_A", "A1", "", nil)
	_, _ = svc.CreateMilestone(ctx, "proj_A", "A2", "", nil)
	_, _ = svc.CreateMilestone(ctx, "proj_B", "B1", "", nil)

	list, err := svc.ListMilestonesByProject(ctx, "proj_A")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 milestones for proj_A, got %d", len(list))
	}
}
