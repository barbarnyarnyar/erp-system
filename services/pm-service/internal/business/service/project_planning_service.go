package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
)

type ProjectPlanningService struct {
	portfolioRepo domain.PortfolioRepository
	projectRepo   domain.ProjectRepository
	milestoneRepo domain.MilestoneRepository
	publisher     domain.EventPublisher
}

func NewProjectPlanningService(
	portfolioRepo domain.PortfolioRepository,
	projectRepo domain.ProjectRepository,
	milestoneRepo domain.MilestoneRepository,
	publisher domain.EventPublisher,
) *ProjectPlanningService {
	return &ProjectPlanningService{
		portfolioRepo: portfolioRepo,
		projectRepo:   projectRepo,
		milestoneRepo: milestoneRepo,
		publisher:     publisher,
	}
}

func (s *ProjectPlanningService) CreatePortfolio(ctx context.Context, name, description, managerID string) (*domain.Portfolio, error) {
	id := utils.NewID("port")
	port := &domain.Portfolio{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if managerID != "" {
		port.ManagerID = &managerID
	}

	err := s.portfolioRepo.Create(ctx, port)
	if err != nil {
		return nil, err
	}
	return port, nil
}

func (s *ProjectPlanningService) ListPortfolios(ctx context.Context) ([]domain.Portfolio, error) {
	return s.portfolioRepo.List(ctx)
}

func (s *ProjectPlanningService) GetPortfolio(ctx context.Context, id string) (*domain.Portfolio, error) {
	return s.portfolioRepo.GetByID(ctx, id)
}

func (s *ProjectPlanningService) CreateProject(ctx context.Context, name, description string, startDate time.Time, endDate *time.Time, portfolioID string, budgetID string) (*domain.Project, error) {
	id := utils.NewID("proj")
	proj := &domain.Project{
		ID:          id,
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      "PLANNING",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if portfolioID != "" {
		proj.PortfolioID = &portfolioID
	}
	if budgetID != "" {
		proj.BudgetID = &budgetID
	}

	err := s.projectRepo.Create(ctx, proj)
	if err != nil {
		return nil, err
	}

	// Publish Event
	if err := s.publisher.Publish(ctx, domain.TopicPrjProjectCreated, id, domain.ProjectCreatedEvent{
		ProjectID:   id,
		ProjectName: name,
		ManagerID:   "",
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjProjectCreated, err)
	}

	return proj, nil
}

func (s *ProjectPlanningService) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	return s.projectRepo.GetByID(ctx, id)
}

func (s *ProjectPlanningService) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return s.projectRepo.List(ctx)
}

func (s *ProjectPlanningService) UpdateProjectStatus(ctx context.Context, id string, status string) (*domain.Project, error) {
	proj, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	proj.Status = status
	proj.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, proj)
	if err != nil {
		return nil, err
	}

	switch status {
	case "ACTIVE":
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectStarted, id, domain.ProjectStartedEvent{
			ProjectID: id,
			Timestamp: time.Now(),
		}); err != nil {
			utils.LogPublishErr("pm-service", domain.TopicPrjProjectStarted, err)
		}
	case "COMPLETED":
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectCompleted, id, domain.ProjectCompletedEvent{
			ProjectID: id,
			Timestamp: time.Now(),
		}); err != nil {
			utils.LogPublishErr("pm-service", domain.TopicPrjProjectCompleted, err)
		}
	case "CANCELLED":
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectCancelled, id, domain.ProjectCancelledEvent{
			ProjectID: id,
			Reason:    "Status changed to cancelled",
			Timestamp: time.Now(),
		}); err != nil {
			utils.LogPublishErr("pm-service", domain.TopicPrjProjectCancelled, err)
		}
	default:
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectUpdated, id, domain.ProjectUpdatedEvent{
			ProjectID: id,
			Status:    status,
			Timestamp: time.Now(),
		}); err != nil {
			utils.LogPublishErr("pm-service", domain.TopicPrjProjectUpdated, err)
		}
	}

	return proj, nil
}

func (s *ProjectPlanningService) RequestCustomOrder(ctx context.Context, projectID, customItemID string, quantity int, requiredBy time.Time) error {
	return s.publisher.Publish(ctx, domain.TopicPrjCustomOrderCreated, projectID, domain.PrjCustomOrderCreatedEvent{
		ProjectID:    projectID,
		CustomItemID: customItemID,
		Quantity:     quantity,
		RequiredBy:   requiredBy,
		Timestamp:    time.Now(),
	})
}

func (s *ProjectPlanningService) DelayProject(ctx context.Context, projectID string, delayDays int) (*domain.Project, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	proj.Status = "DELAYED"
	proj.UpdatedAt = time.Now()

	err = s.projectRepo.Update(ctx, proj)
	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicPrjProjectDelayed, projectID, domain.ProjectDelayedEvent{
		ProjectID: projectID,
		DelayDays: delayDays,
		Timestamp: time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjProjectDelayed, err)
	}

	return proj, nil
}

// CreateMilestone creates a new milestone on a project (Phase 2.21).
// Status starts as PENDING; transitions to IN_PROGRESS, ACHIEVED, DELAYED,
// or CANCELLED as the project evolves.
func (s *ProjectPlanningService) CreateMilestone(ctx context.Context, projectID, name, description string, targetDate *time.Time) (*domain.Milestone, error) {
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	id := utils.NewID("ms")
	m := &domain.Milestone{
		ID:         id,
		ProjectID:  projectID,
		Name:       name,
		Status:     "PENDING",
		TargetDate: targetDate,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if description != "" {
		m.Description = &description
	}

	if err := s.milestoneRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetMilestone retrieves a milestone by ID.
func (s *ProjectPlanningService) GetMilestone(ctx context.Context, id string) (*domain.Milestone, error) {
	return s.milestoneRepo.GetByID(ctx, id)
}

// ListMilestonesByProject returns all milestones for a given project.
func (s *ProjectPlanningService) ListMilestonesByProject(ctx context.Context, projectID string) ([]domain.Milestone, error) {
	return s.milestoneRepo.ListByProject(ctx, projectID)
}

// CompleteMilestone marks a milestone as ACHIEVED, records the completion
// date, and fires the prj.milestone.achieved event. Previously this method
// published the event without persisting any state — the event payload had
// no entity behind it. Now the milestone record is updated atomically with
// the event fire.
func (s *ProjectPlanningService) CompleteMilestone(ctx context.Context, milestoneID string, completionDate time.Time) (*domain.Milestone, error) {
	m, err := s.milestoneRepo.GetByID(ctx, milestoneID)
	if err != nil {
		return nil, err
	}
	if m.Status == "ACHIEVED" {
		return nil, fmt.Errorf("milestone %s is already achieved", milestoneID)
	}

	now := time.Now()
	m.Status = "ACHIEVED"
	m.CompletionDate = &completionDate
	m.UpdatedAt = now

	if err := s.milestoneRepo.Update(ctx, m); err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicPrjMilestoneAchieved, milestoneID, domain.MilestoneAchievedEvent{
		ProjectID:      m.ProjectID,
		MilestoneID:    milestoneID,
		Name:           m.Name,
		CompletionDate: completionDate,
		Timestamp:      now,
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjMilestoneAchieved, err)
	}

	return m, nil
}

// DelayMilestone marks a milestone as DELAYED with a new target date and
// fires the prj.milestone.delayed event. Like CompleteMilestone, this
// replaces the previous event-only stub with a full entity lifecycle.
func (s *ProjectPlanningService) DelayMilestone(ctx context.Context, milestoneID string, newTargetDate time.Time) (*domain.Milestone, error) {
	m, err := s.milestoneRepo.GetByID(ctx, milestoneID)
	if err != nil {
		return nil, err
	}

	m.Status = "DELAYED"
	m.TargetDate = &newTargetDate
	m.UpdatedAt = time.Now()

	if err := s.milestoneRepo.Update(ctx, m); err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(ctx, domain.TopicPrjMilestoneDelayed, milestoneID, domain.MilestoneDelayedEvent{
		ProjectID:   m.ProjectID,
		MilestoneID: milestoneID,
		Name:        m.Name,
		TargetDate:  newTargetDate,
		Timestamp:   time.Now(),
	}); err != nil {
		utils.LogPublishErr("pm-service", domain.TopicPrjMilestoneDelayed, err)
	}

	return m, nil
}
