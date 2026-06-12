package service

import (
	"context"
	"erp-system/shared/utils"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CashManagementService struct {
	payments   domain.PaymentRepository
	invoices   domain.ArInvoiceRepository
	statements domain.BankStatementRepository
	outbox     domain.TransactionalOutboxRepository
	tm         domain.TransactionManager
}

func NewCashManagementService(
	payments domain.PaymentRepository,
	invoices domain.ArInvoiceRepository,
	statements domain.BankStatementRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *CashManagementService {
	return &CashManagementService{
		payments:   payments,
		invoices:   invoices,
		statements: statements,
		outbox:     outbox,
		tm:         tm,
	}
}

func (s *CashManagementService) ListPayments(ctx context.Context) ([]domain.Payment, error) {
	return s.payments.List(ctx)
}

func (s *CashManagementService) RecordPayment(ctx context.Context, invoiceID, billID, bankAccountID string, amount decimal.Decimal, method string) (*domain.Payment, error) {
	id := utils.NewID("pay")
	payNum := fmt.Sprintf("PAY-%d", time.Now().Unix())

	payment := &domain.Payment{
		ID:            id,
		PaymentNumber: payNum,
		PaymentDate:   time.Now(),
		Amount:        amount,
		PaymentMethod: method,
		Status:        "COMPLETED",
		BankAccountID: nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if bankAccountID != "" {
		payment.BankAccountID = &bankAccountID
	}
	if invoiceID != "" {
		payment.InvoiceID = &invoiceID
	}
	if billID != "" {
		payment.BillID = &billID
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if invoiceID != "" {
			inv, err := s.invoices.GetByID(txCtx, invoiceID)
			if err != nil {
				return err
			}
			inv.Status = domain.PaymentStatusPAID
			err = s.invoices.Update(txCtx, inv)
			if err != nil {
				return err
			}

			// Write invoice paid to outbox
			outboxRec := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   string(domain.TopicFmInvoicePaid),
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
			if err := s.outbox.Create(txCtx, outboxRec); err != nil {
				return err
			}
		}

		err := s.payments.Create(txCtx, payment)
		if err != nil {
			return err
		}

		// Write payment received to outbox
		outboxRecReceived := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmPaymentReceived),
			AggregateID: payment.ID,
			Payload: domain.PaymentEventPayload{
				ID:            payment.ID,
				InvoiceID:     payment.InvoiceID,
				BillID:        payment.BillID,
				PaymentNumber: payment.PaymentNumber,
				Amount:        payment.Amount,
				PaymentMethod: payment.PaymentMethod,
				Status:        payment.Status,
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		if err := s.outbox.Create(txCtx, outboxRecReceived); err != nil {
			return err
		}

		// Write payment processed to outbox
		outboxRecProcessed := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmPaymentProcessed),
			AggregateID: payment.ID,
			Payload: domain.PaymentEventPayload{
				ID:            payment.ID,
				InvoiceID:     payment.InvoiceID,
				BillID:        payment.BillID,
				PaymentNumber: payment.PaymentNumber,
				Amount:        payment.Amount,
				PaymentMethod: payment.PaymentMethod,
				Status:        payment.Status,
				Timestamp:     time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		if err := s.outbox.Create(txCtx, outboxRecProcessed); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Publish payment failed event in a separate transaction
		_ = s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
			outboxRecFailed := &domain.TransactionalOutbox{
				ID:          utils.NewID("outbox"),
				EventType:   string(domain.TopicFmPaymentFailed),
				AggregateID: payment.ID,
				Payload: domain.PaymentEventPayload{
					ID:            payment.ID,
					InvoiceID:     payment.InvoiceID,
					BillID:        payment.BillID,
					PaymentNumber: payment.PaymentNumber,
					Amount:        payment.Amount,
					PaymentMethod: payment.PaymentMethod,
					Status:        "FAILED",
					Timestamp:     time.Now(),
				},
				Status:    domain.OutboxStatusPENDING,
				CreatedAt: time.Now(),
			}
			return s.outbox.Create(txCtx, outboxRecFailed)
		})
		return nil, err
	}

	return payment, nil
}

func (s *CashManagementService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.payments.GetByID(ctx, id)
}

func (s *CashManagementService) ReconcileBankStatement(ctx context.Context, statementID string) error {
	return nil
}

func (s *CashManagementService) GetCashFlowForecast(ctx context.Context, monthsAhead int) (map[string]interface{}, error) {
	return map[string]interface{}{
		"forecast_period_months": monthsAhead,
		"projected_cash_inflow":  decimal.NewFromInt(125000),
		"projected_cash_outflow": decimal.NewFromInt(80000),
		"net_cash_flow":          decimal.NewFromInt(45000),
	}, nil
}

func (s *CashManagementService) GetBankStatement(ctx context.Context, id string) (*domain.BankStatement, []domain.BankStatementLine, error) {
	if s.statements == nil {
		return nil, nil, fmt.Errorf("bank statement repository not initialized")
	}
	return s.statements.GetByID(ctx, id)
}
