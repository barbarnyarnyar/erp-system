package service

import (
	"context"
	"fmt"
	"time"

	"github.com/erp-system/hr-service/internal/business/domain"
)

type EmployeeDocumentService struct {
	repo domain.EmployeeDocumentRepository
}

func NewEmployeeDocumentService(repo domain.EmployeeDocumentRepository) *EmployeeDocumentService {
	return &EmployeeDocumentService{repo: repo}
}

func (s *EmployeeDocumentService) ListDocuments(ctx context.Context, employeeID string) ([]domain.EmployeeDocument, error) {
	return s.repo.ListByEmployeeID(ctx, employeeID)
}

func (s *EmployeeDocumentService) UploadDocument(ctx context.Context, employeeID string, docType string, fileName string, fileURL string) (*domain.EmployeeDocument, error) {
	id := fmt.Sprintf("doc_%d", time.Now().UnixNano())

	doc := &domain.EmployeeDocument{
		ID:         id,
		EmployeeID: employeeID,
		DocType:    docType,
		FileName:   fileName,
		FileUrl:    fileURL,
		UploadedAt: time.Now(),
	}

	err := s.repo.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *EmployeeDocumentService) DeleteDocument(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
