package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type RecruitmentService struct {
	postings     domain.JobPostingRepository
	applications domain.JobApplicationRepository
}

func NewRecruitmentService(postings domain.JobPostingRepository, applications domain.JobApplicationRepository) *RecruitmentService {
	return &RecruitmentService{
		postings:     postings,
		applications: applications,
	}
}

func (s *RecruitmentService) ListJobPostings(ctx context.Context) ([]domain.JobPosting, error) {
	return s.postings.List(ctx)
}

func (s *RecruitmentService) CreateJobPosting(ctx context.Context, title, description, deptID, location, salaryRange string) (*domain.JobPosting, error) {
	id := fmt.Sprintf("job_%d", time.Now().UnixNano())

	jp := &domain.JobPosting{
		ID:           id,
		Title:        title,
		Description:  description,
		DepartmentID: deptID,
		Location:     location,
		SalaryRange:  salaryRange,
		Status:       "OPEN",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := s.postings.Create(ctx, jp)
	if err != nil {
		return nil, err
	}

	return jp, nil
}

func (s *RecruitmentService) GetJobPosting(ctx context.Context, id string) (*domain.JobPosting, error) {
	return s.postings.GetByID(ctx, id)
}

func (s *RecruitmentService) UpdateJobPosting(ctx context.Context, id string, title, description, location, salaryRange, status string) (*domain.JobPosting, error) {
	jp, err := s.postings.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	jp.Title = title
	jp.Description = description
	jp.Location = location
	jp.SalaryRange = salaryRange
	jp.Status = status
	jp.UpdatedAt = time.Now()

	err = s.postings.Update(ctx, jp)
	if err != nil {
		return nil, err
	}

	return jp, nil
}

func (s *RecruitmentService) DeleteJobPosting(ctx context.Context, id string) error {
	return s.postings.Delete(ctx, id)
}

func (s *RecruitmentService) ListApplications(ctx context.Context) ([]domain.JobApplication, error) {
	return s.applications.List(ctx)
}

func (s *RecruitmentService) CreateApplication(ctx context.Context, jobPostingID, applicantName, email, phone, resumeURL string) (*domain.JobApplication, error) {
	id := fmt.Sprintf("app_%d", time.Now().UnixNano())

	ja := &domain.JobApplication{
		ID:            id,
		JobPostingID:  jobPostingID,
		ApplicantName: applicantName,
		Email:         email,
		Phone:         phone,
		ResumeUrl:     resumeURL,
		Status:        "APPLIED",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.applications.Create(ctx, ja)
	if err != nil {
		return nil, err
	}

	return ja, nil
}

func (s *RecruitmentService) GetApplication(ctx context.Context, id string) (*domain.JobApplication, error) {
	return s.applications.GetByID(ctx, id)
}

func (s *RecruitmentService) UpdateApplication(ctx context.Context, id, status string) (*domain.JobApplication, error) {
	ja, err := s.applications.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ja.Status = status
	ja.UpdatedAt = time.Now()

	err = s.applications.Update(ctx, ja)
	if err != nil {
		return nil, err
	}

	return ja, nil
}
