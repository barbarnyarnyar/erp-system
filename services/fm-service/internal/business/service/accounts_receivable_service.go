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
	invoices domain.ArInvoiceRepository
	credits  domain.CustomerCreditRepository
	outbox   domain.TransactionalOutboxRepository
	tm       domain.TransactionManager
}

func NewAccountsReceivableService(
	invoices domain.ArInvoiceRepository,
	credits domain.CustomerCreditRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *AccountsReceivableService {
	return &AccountsReceivableService{
		invoices: invoices,
		credits:  credits,
		outbox:   outbox,
		tm:       tm,
	}
}

func (s *AccountsReceivableService) ListInvoices(ctx context.Context) ([]domain.ArInvoice, error) {
	return s.invoices.List(ctx)
}

func (s *AccountsReceivableService) CreateInvoice(ctx context.Context, legalEntityID, customerID, salesOrderID string, totalAmount, taxAmount decimal.Decimal, dueDate time.Time) (*domain.ArInvoice, error) {
	id := utils.NewID("inv")
	invNum := fmt.Sprintf("INV-%d", time.Now().Unix())

	inv := &domain.ArInvoice{
		ID:            id,
		LegalEntityID: legalEntityID,
		InvoiceNumber: invNum,
		CustomerID:    customerID,
		SalesOrderID:  salesOrderID,
		TotalAmount:   totalAmount,
		TaxAmount:     taxAmount,
		DueDate:       dueDate,
		Status:        domain.PaymentStatusOPEN,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.invoices.Create(txCtx, inv)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmInvoiceCreated),
			AggregateID: inv.ID,
			Payload: domain.InvoiceEventPayload{
				ID:            inv.ID,
				CustomerID:    inv.CustomerID,
				InvoiceNumber: inv.InvoiceNumber,
				TotalAmount:   inv.TotalAmount,
				Status:        string(inv.Status),
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *AccountsReceivableService) GetInvoice(ctx context.Context, id string) (*domain.ArInvoice, error) {
	return s.invoices.GetByID(ctx, id)
}

func (s *AccountsReceivableService) UpdateInvoice(ctx context.Context, id string, fields map[string]interface{}) (*domain.ArInvoice, error) {
	var inv *domain.ArInvoice
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		inv, err = s.invoices.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		if statusStr, ok := fields["status"].(string); ok {
			inv.Status = domain.PaymentStatus(statusStr)
		}
		inv.UpdatedAt = time.Now()

		err = s.invoices.Update(txCtx, inv)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmInvoiceUpdated),
			AggregateID: inv.ID,
			Payload: domain.InvoiceEventPayload{
				ID:            inv.ID,
				CustomerID:    inv.CustomerID,
				InvoiceNumber: inv.InvoiceNumber,
				TotalAmount:   inv.TotalAmount,
				Status:        string(inv.Status),
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (s *AccountsReceivableService) DeleteInvoice(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		return s.invoices.Delete(txCtx, id)
	})
}

func (s *AccountsReceivableService) SendInvoice(ctx context.Context, id string) (bool, error) {
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		inv, err := s.invoices.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		inv.Status = domain.PaymentStatusOPEN
		err = s.invoices.Update(txCtx, inv)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmInvoiceSent),
			AggregateID: inv.ID,
			Payload: domain.InvoiceEventPayload{
				ID:            inv.ID,
				CustomerID:    inv.CustomerID,
				InvoiceNumber: inv.InvoiceNumber,
				TotalAmount:   inv.TotalAmount,
				Status:        string(inv.Status),
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *AccountsReceivableService) CheckCustomerCredit(ctx context.Context, customerID string, orderValue decimal.Decimal) (bool, error) {
	limit := decimal.NewFromInt(100000)
	if orderValue.GreaterThan(limit) {
		return false, nil
	}
	return true, nil
}

func (s *AccountsReceivableService) MarkInvoiceOverdue(ctx context.Context, id string) error {
	return s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		inv, err := s.invoices.GetByID(txCtx, id)
		if err != nil {
			return err
		}

		inv.Status = domain.PaymentStatusOPEN // or define an overdue status if in domain. But in CDD PaymentStatus is OPEN, PARTIAL, PAID.
		inv.UpdatedAt = time.Now()
		err = s.invoices.Update(txCtx, inv)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmInvoiceOverdue),
			AggregateID: inv.ID,
			Payload: domain.InvoiceEventPayload{
				ID:            inv.ID,
				CustomerID:    inv.CustomerID,
				InvoiceNumber: inv.InvoiceNumber,
				TotalAmount:   inv.TotalAmount,
				Status:        string(inv.Status),
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})
}

func (s *AccountsReceivableService) GetCustomerCredit(ctx context.Context, customerID string) (*domain.CustomerCredit, error) {
	cc, err := s.credits.GetByCustomerID(ctx, customerID)
	if err != nil || cc == nil {
		cc = &domain.CustomerCredit{
			ID:             utils.NewID("cc"),
			CustomerID:     customerID,
			CreditLimit:    decimal.NewFromFloat(5000.00),
			CurrentBalance: decimal.Zero,
			IsOnHold:       false,
			Version:        1,
			UpdatedAt:      time.Now(),
		}
		if createErr := s.credits.Create(ctx, cc); createErr != nil {
			return nil, createErr
		}
	}
	return cc, nil
}

