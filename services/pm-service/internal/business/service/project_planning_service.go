package service

import (
	"log"
	"context"
	"fmt"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
)

type ProjectPlanningService struct {
	portfolioRepo domain.PortfolioRepository
	projectRepo   domain.ProjectRepository
	publisher     domain.EventPublisher
}

func NewProjectPlanningService(
	portfolioRepo domain.PortfolioRepository,
	projectRepo domain.ProjectRepository,
	publisher domain.EventPublisher,
) *ProjectPlanningService {
	return &ProjectPlanningService{
		portfolioRepo: portfolioRepo,
		projectRepo:   projectRepo,
		publisher:     publisher,
	}
}

func (s *ProjectPlanningService) CreatePortfolio(ctx context.Context, name, description, managerID string) (*domain.Portfolio, error) {
	id := fmt.Sprintf("port_%d", time.Now().UnixNano())
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
	id := fmt.Sprintf("proj_%d", time.Now().UnixNano())
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
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicPrjProjectCreated, err)
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
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicPrjProjectStarted, err)
		}
	case "COMPLETED":
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectCompleted, id, domain.ProjectCompletedEvent{
			ProjectID: id,
			Timestamp: time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicPrjProjectCompleted, err)
		}
	case "CANCELLED":
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectCancelled, id, domain.ProjectCancelledEvent{
			ProjectID: id,
			Reason:    "Status changed to cancelled",
			Timestamp: time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicPrjProjectCancelled, err)
		}
	default:
		if err := s.publisher.Publish(ctx, domain.TopicPrjProjectUpdated, id, domain.ProjectUpdatedEvent{
			ProjectID: id,
			Status:    status,
			Timestamp: time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicPrjProjectUpdated, err)
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
