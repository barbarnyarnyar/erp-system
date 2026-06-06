package service

import (
	"log"
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type PerformanceService struct {
	repo      domain.PerformanceReviewRepository
	publisher domain.EventPublisher
}

func NewPerformanceService(repo domain.PerformanceReviewRepository, publisher domain.EventPublisher) *PerformanceService {
	return &PerformanceService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *PerformanceService) ListPerformanceReviews(ctx context.Context) ([]domain.PerformanceReview, error) {
	return s.repo.List(ctx)
}

func (s *PerformanceService) CreatePerformanceReview(ctx context.Context, employeeID, reviewerID string, reviewDate time.Time, periodStart, periodEnd time.Time, rating int, feedback string) (*domain.PerformanceReview, error) {
	id := fmt.Sprintf("review_%d", time.Now().UnixNano())

	pr := &domain.PerformanceReview{
		ID:           id,
		EmployeeID:   employeeID,
		ReviewerID:   reviewerID,
		ReviewDate:   reviewDate,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Rating:       rating,
		Feedback:     feedback,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.repo.Create(ctx, pr)
	if err != nil {
		return nil, err
	}

	// Publish performance review completed event
	if err := s.publisher.Publish(ctx, domain.TopicHrPerformanceReviewCompleted, pr.ID, domain.PerformanceReviewCompletedEvent{
		ReviewID:   pr.ID,
		EmployeeID: pr.EmployeeID,
		ReviewerID: pr.ReviewerID,
		Rating:     pr.Rating,
		Timestamp:  time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrPerformanceReviewCompleted, err)
	}

	// Publish performance improvement needed if rating is low (e.g., < 3)
	if pr.Rating < 3 {
		if err := s.publisher.Publish(ctx, domain.TopicHrPerformanceImprovementNeeded, pr.EmployeeID, domain.PerformanceImprovementNeededEvent{
			EmployeeID: pr.EmployeeID,
			ReviewID:    pr.ID,
			Details:     fmt.Sprintf("Low performance review rating of %d: %s", pr.Rating, pr.Feedback),
			Timestamp:   time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrPerformanceImprovementNeeded, err)
		}
	}

	return pr, nil
}

func (s *PerformanceService) GetPerformanceReview(ctx context.Context, id string) (*domain.PerformanceReview, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PerformanceService) UpdatePerformanceReview(ctx context.Context, id string, rating int, feedback string) (*domain.PerformanceReview, error) {
	pr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pr.Rating = rating
	pr.Feedback = feedback
	pr.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, pr)
	if err != nil {
		return nil, err
	}

	// Publish performance review completed event again for updates
	if err := s.publisher.Publish(ctx, domain.TopicHrPerformanceReviewCompleted, pr.ID, domain.PerformanceReviewCompletedEvent{
		ReviewID:   pr.ID,
		EmployeeID: pr.EmployeeID,
		ReviewerID: pr.ReviewerID,
		Rating:     pr.Rating,
		Timestamp:  time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrPerformanceReviewCompleted, err)
	}

	// Publish performance improvement needed if updated rating is low
	if pr.Rating < 3 {
		if err := s.publisher.Publish(ctx, domain.TopicHrPerformanceImprovementNeeded, pr.EmployeeID, domain.PerformanceImprovementNeededEvent{
			EmployeeID: pr.EmployeeID,
			ReviewID:    pr.ID,
			Details:     fmt.Sprintf("Low performance review rating updated to %d: %s", pr.Rating, pr.Feedback),
			Timestamp:   time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicHrPerformanceImprovementNeeded, err)
		}
	}

	return pr, nil
}
