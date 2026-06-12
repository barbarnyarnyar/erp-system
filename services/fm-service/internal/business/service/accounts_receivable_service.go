package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type AccountsReceivableService struct {
	invoices  domain.InvoiceRepository
	publisher domain.EventPublisher
}

func NewAccountsReceivableService(invoices domain.InvoiceRepository, publisher domain.EventPublisher) *AccountsReceivableService {
	return &AccountsReceivableService{
		invoices:  invoices,
		publisher: publisher,
	}
}

func (s *AccountsReceivableService) ListInvoices(ctx context.Context) ([]domain.Invoice, error) {
	return s.invoices.List(ctx)
}

func (s *AccountsReceivableService) CreateInvoice(ctx context.Context, customerID string, issueDate, dueDate time.Time, lines []domain.InvoiceLine) (*domain.Invoice, error) {
	id := utils.NewID("inv")
	invNum := fmt.Sprintf("INV-%d", time.Now().Unix())

	total := decimal.Zero
	for _, l := range lines {
		total = total.Add(l.LineTotal)
	}

	inv := &domain.Invoice{
		ID:            id,
		CustomerID:    customerID,
		InvoiceNumber: invNum,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		TotalAmount:   total,
		Status:        "SENT",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.invoices.Create(ctx, inv, lines)
	if err != nil {
		return nil, err
	}

	// Publish event
	if err := s.publisher.Publish(ctx, domain.TopicFmInvoiceCreated, inv.ID, domain.InvoiceEventPayload{
		ID:            inv.ID,
		CustomerID:    inv.CustomerID,
		InvoiceNumber: inv.InvoiceNumber,
		TotalAmount:   inv.TotalAmount,
		Status:        inv.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		utils.LogPublishErr("fm-service", domain.TopicFmInvoiceCreated, err)
	}

	return inv, nil
}

func (s *AccountsReceivableService) GetInvoice(ctx context.Context, id string) (*domain.Invoice, []domain.InvoiceLine, error) {
	return s.invoices.GetByID(ctx, id)
}

func (s *AccountsReceivableService) UpdateInvoice(ctx context.Context, id string, fields map[string]interface{}) (*domain.Invoice, error) {
	inv, _, err := s.invoices.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if status, ok := fields["status"].(string); ok {
		inv.Status = status
	}
	inv.UpdatedAt = time.Now()

	err = s.invoices.Update(ctx, inv)
	if err != nil {
		return nil, err
	}

	// Publish event
	if err := s.publisher.Publish(ctx, domain.TopicFmInvoiceUpdated, inv.ID, domain.InvoiceEventPayload{
		ID:            inv.ID,
		CustomerID:    inv.CustomerID,
		InvoiceNumber: inv.InvoiceNumber,
		TotalAmount:   inv.TotalAmount,
		Status:        inv.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		utils.LogPublishErr("fm-service", domain.TopicFmInvoiceUpdated, err)
	}

	return inv, nil
}

func (s *AccountsReceivableService) DeleteInvoice(ctx context.Context, id string) error {
	return s.invoices.Delete(ctx, id)
}

func (s *AccountsReceivableService) SendInvoice(ctx context.Context, id string) (bool, error) {
	inv, _, err := s.invoices.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	inv.Status = "SENT"
	err = s.invoices.Update(ctx, inv)
	if err != nil {
		return false, err
	}

	// Publish event
	if err := s.publisher.Publish(ctx, domain.TopicFmInvoiceSent, inv.ID, domain.InvoiceEventPayload{
		ID:            inv.ID,
		CustomerID:    inv.CustomerID,
		InvoiceNumber: inv.InvoiceNumber,
		TotalAmount:   inv.TotalAmount,
		Status:        inv.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		utils.LogPublishErr("fm-service", domain.TopicFmInvoiceSent, err)
	}

	return true, nil
}

func (s *AccountsReceivableService) CheckCustomerCredit(ctx context.Context, customerID string, orderValue decimal.Decimal) (bool, error) {
	// For mock purposes: allow all credit unless order value is abnormally high (e.g. > $100,000)
	limit := decimal.NewFromInt(100000)
	if orderValue.GreaterThan(limit) {
		return false, nil
	}
	return true, nil
}

func (s *AccountsReceivableService) MarkInvoiceOverdue(ctx context.Context, id string) error {
	inv, _, err := s.invoices.GetByID(ctx, id)
	if err != nil {
		return err
	}

	inv.Status = "OVERDUE"
	inv.UpdatedAt = time.Now()
	err = s.invoices.Update(ctx, inv)
	if err != nil {
		return err
	}

	if err := s.publisher.Publish(ctx, domain.TopicFmInvoiceOverdue, inv.ID, domain.InvoiceEventPayload{
		ID:            inv.ID,
		CustomerID:    inv.CustomerID,
		InvoiceNumber: inv.InvoiceNumber,
		TotalAmount:   inv.TotalAmount,
		Status:        inv.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		utils.LogPublishErr("fm-service", domain.TopicFmInvoiceOverdue, err)
	}
	return nil
}
