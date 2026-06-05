package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type TrainingService struct {
	repo        domain.TrainingProgramRepository
	enrollments domain.TrainingEnrollmentRepository
	publisher   domain.EventPublisher
}

func NewTrainingService(repo domain.TrainingProgramRepository, enrollments domain.TrainingEnrollmentRepository, publisher domain.EventPublisher) *TrainingService {
	return &TrainingService{
		repo:        repo,
		enrollments: enrollments,
		publisher:   publisher,
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

func (s *TrainingService) EnrollEmployee(ctx context.Context, trainingID string, employeeID string) (*domain.TrainingEnrollment, error) {
	id := fmt.Sprintf("enr_%d", time.Now().UnixNano())
	te := &domain.TrainingEnrollment{
		ID:         id,
		TrainingID: trainingID,
		EmployeeID: employeeID,
		EnrolledAt: time.Now(),
		Status:     "ENROLLED",
	}

	err := s.enrollments.Create(ctx, te)
	if err != nil {
		return nil, err
	}
	return te, nil
}

func (s *TrainingService) CompleteTraining(ctx context.Context, enrollmentID string) (*domain.TrainingEnrollment, error) {
	te, err := s.enrollments.GetByID(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	te.CompletedAt = &now
	te.Status = "COMPLETED"

	err = s.enrollments.Update(ctx, te)
	if err != nil {
		return nil, err
	}

	// Publish training completed event for this specific enrollment
	_ = s.publisher.Publish(ctx, domain.TopicHrTrainingCompleted, te.ID, domain.TrainingCompletedEvent{
		TrainingProgramID: te.TrainingID,
		EmployeeID:        te.EmployeeID,
		CompletionDate:    now,
		Timestamp:         now,
	})

	return te, nil
}

