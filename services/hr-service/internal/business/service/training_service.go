package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type TrainingService struct {
	repo      domain.TrainingProgramRepository
	publisher domain.EventPublisher
}

func NewTrainingService(repo domain.TrainingProgramRepository, publisher domain.EventPublisher) *TrainingService {
	return &TrainingService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *TrainingService) ListTrainingPrograms(ctx context.Context) ([]domain.TrainingProgram, error) {
	return s.repo.List(ctx)
}

func (s *TrainingService) CreateTrainingProgram(ctx context.Context, title, description, trainer string, start, end time.Time) (*domain.TrainingProgram, error) {
	id := fmt.Sprintf("train_%d", time.Now().UnixNano())

	tp := &domain.TrainingProgram{
		ID:          id,
		Title:       title,
		Description: description,
		Trainer:     trainer,
		StartDate:   start,
		EndDate:     end,
		Status:      "SCHEDULED",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.repo.Create(ctx, tp)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func (s *TrainingService) GetTrainingProgram(ctx context.Context, id string) (*domain.TrainingProgram, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TrainingService) UpdateTrainingProgram(ctx context.Context, id string, title, description, trainer string, start, end time.Time, status string) (*domain.TrainingProgram, error) {
	tp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := tp.Status

	tp.Title = title
	tp.Description = description
	tp.Trainer = trainer
	tp.StartDate = start
	tp.EndDate = end
	tp.Status = status
	tp.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, tp)
	if err != nil {
		return nil, err
	}

	// Publish training completed event if status changed to COMPLETED
	if status == "COMPLETED" && oldStatus != "COMPLETED" {
		_ = s.publisher.Publish(ctx, domain.TopicHrTrainingCompleted, tp.ID, domain.TrainingCompletedEvent{
			TrainingProgramID: tp.ID,
			EmployeeID:        "", // Broadcast or individual (left blank for general program completion)
			CompletionDate:    time.Now(),
			Timestamp:         time.Now(),
		})
	}

	return tp, nil
}
