package service

import (
	"context"
	"erp-system/shared/utils"
	"time"

	"github.com/erp-system/pm-service/internal/business/domain"
)

type CollaborationService struct {
	docRepo    domain.ProjectDocumentRepository
	issueRepo  domain.ProjectIssueRepository
	changeRepo domain.ChangeRequestRepository
	publisher  domain.EventPublisher
}

func NewCollaborationService(
	docRepo domain.ProjectDocumentRepository,
	issueRepo domain.ProjectIssueRepository,
	changeRepo domain.ChangeRequestRepository,
	publisher domain.EventPublisher,
) *CollaborationService {
	return &CollaborationService{
		docRepo:    docRepo,
		issueRepo:  issueRepo,
		changeRepo: changeRepo,
		publisher:  publisher,
	}
}

func (s *CollaborationService) UploadDocument(ctx context.Context, projectID, name, filePath string, fileSize int, uploadedBy string) (*domain.ProjectDocument, error) {
	id := utils.NewID("doc")
	doc := &domain.ProjectDocument{
		ID:         id,
		ProjectID:  projectID,
		Name:       name,
		FilePath:   filePath,
		FileSize:   fileSize,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *CollaborationService) ListDocuments(ctx context.Context, projectID string) ([]domain.ProjectDocument, error) {
	return s.docRepo.ListByProject(ctx, projectID)
}

func (s *CollaborationService) LogIssue(ctx context.Context, projectID, title, description, severity, raisedBy string) (*domain.ProjectIssue, error) {
	id := utils.NewID("iss")
	issue := &domain.ProjectIssue{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      "OPEN",
		RaisedBy:    raisedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.issueRepo.Create(ctx, issue)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *CollaborationService) ResolveIssue(ctx context.Context, issueID string, assignedTo string) (*domain.ProjectIssue, error) {
	issue, err := s.issueRepo.GetByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	issue.Status = "RESOLVED"
	if assignedTo != "" {
		issue.AssignedTo = &assignedTo
	}
	issue.UpdatedAt = time.Now()

	err = s.issueRepo.Update(ctx, issue)
	if err != nil {
		return nil, err
	}
	return issue, nil
}

func (s *CollaborationService) ListIssues(ctx context.Context, projectID string) ([]domain.ProjectIssue, error) {
	return s.issueRepo.ListByProject(ctx, projectID)
}

func (s *CollaborationService) CreateChangeRequest(ctx context.Context, projectID, title, description, reason, impact string, requestedBy string) (*domain.ChangeRequest, error) {
	id := utils.NewID("cr")
	req := &domain.ChangeRequest{
		ID:             id,
		ProjectID:      projectID,
		Title:          title,
		Description:    description,
		Reason:         reason,
		ImpactAnalysis: impact,
		Status:         "PENDING",
		RequestedBy:    requestedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.changeRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *CollaborationService) ApproveChangeRequest(ctx context.Context, requestID string, approvedBy string) (*domain.ChangeRequest, error) {
	req, err := s.changeRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	req.Status = "APPROVED"
	req.ApprovedBy = &approvedBy
	req.UpdatedAt = time.Now()

	err = s.changeRepo.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *CollaborationService) ListChangeRequests(ctx context.Context, projectID string) ([]domain.ChangeRequest, error) {
	return s.changeRepo.ListByProject(ctx, projectID)
}
