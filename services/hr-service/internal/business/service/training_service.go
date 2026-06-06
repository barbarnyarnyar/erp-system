package service

import (
	"log"
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
		if err := s.publisher.Publish(ctx, domain.TopicHrTrainingCompleted, tp.ID, domain.TrainingCompletedEvent{
			TrainingProgramID: tp.ID,
			EmployeeID:        "", // Broadcast or individual (left blank for general program completion)
			CompletionDate:    time.Now(),
			Timestamp:         time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrTrainingCompleted, err)
		}
	}

	return tp, nil
}

func (s *TrainingService) EnrollEmployee(ctx context.Context, trainingID string, employeeID string) (*domain.TrainingEnrollment, error) {
	// Guard: prevent duplicate active enrollments for the same (training, employee) pair.
	// An employee may be re-enrolled only if their previous enrollment is CANCELLED or
	// COMPLETED. This stops a real bug where the same employee could be enrolled in the
	// same training program multiple times (repo's GetByTrainingAndEmployee existed but
	// was never called by this service).
	existing, err := s.enrollments.GetByTrainingAndEmployee(ctx, trainingID, employeeID)
	if err == nil && existing != nil {
		switch existing.Status {
		case "ENROLLED", "IN_PROGRESS":
			return nil, fmt.Errorf("employee %s is already enrolled in training %s (enrollment %s, status %s)",
				employeeID, trainingID, existing.ID, existing.Status)
		}
		// For CANCELLED or COMPLETED, fall through and create a new enrollment.
	}

	id := fmt.Sprintf("enr_%d", time.Now().UnixNano())
	te := &domain.TrainingEnrollment{
		ID:         id,
		TrainingID: trainingID,
		EmployeeID: employeeID,
		EnrolledAt: time.Now(),
		Status:     "ENROLLED",
	}

	err = s.enrollments.Create(ctx, te)
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
	if err := s.publisher.Publish(ctx, domain.TopicHrTrainingCompleted, te.ID, domain.TrainingCompletedEvent{
		TrainingProgramID: te.TrainingID,
		EmployeeID:        te.EmployeeID,
		CompletionDate:    now,
		Timestamp:         now,
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrTrainingCompleted, err)
	}

	return te, nil
}

func (s *TrainingService) EarnCertification(ctx context.Context, employeeID, certificationName string, expiryDate time.Time) error {
	if err := s.publisher.Publish(ctx, domain.TopicHrCertificationEarned, employeeID, domain.CertificationEarnedEvent{
		EmployeeID:        employeeID,
		CertificationName: certificationName,
		ExpiryDate:        expiryDate,
		Timestamp:         time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrCertificationEarned, err)
		return err
	}
	return nil
}

func (s *TrainingService) AcquireSkill(ctx context.Context, employeeID, skillName, proficiency string) error {
	if err := s.publisher.Publish(ctx, domain.TopicHrSkillAcquired, employeeID, domain.SkillAcquiredEvent{
		EmployeeID:  employeeID,
		SkillName:   skillName,
		Proficiency: proficiency,
		Timestamp:   time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrSkillAcquired, err)
		return err
	}
	return nil
}

