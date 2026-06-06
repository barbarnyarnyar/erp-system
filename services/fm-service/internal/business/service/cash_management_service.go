package service

import (
	"log"
	"context"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type CashManagementService struct {
	payments  domain.PaymentRepository
	invoices  domain.InvoiceRepository
	publisher domain.EventPublisher
}

func NewCashManagementService(payments domain.PaymentRepository, invoices domain.InvoiceRepository, publisher domain.EventPublisher) *CashManagementService {
	return &CashManagementService{
		payments:  payments,
		invoices:  invoices,
		publisher: publisher,
	}
}

func (s *CashManagementService) ListPayments(ctx context.Context) ([]domain.Payment, error) {
	return s.payments.List(ctx)
}

func (s *CashManagementService) RecordPayment(ctx context.Context, invoiceID, billID, bankAccountID string, amount decimal.Decimal, method string) (*domain.Payment, error) {
	id := fmt.Sprintf("pay_%d", time.Now().UnixNano())
	payNum := fmt.Sprintf("PAY-%d", time.Now().Unix())

	payment := &domain.Payment{
		ID:              id,
		PaymentNumber:   payNum,
		PaymentDate:     time.Now(),
		Amount:          amount,
		PaymentMethod:   method,
		Status:          "COMPLETED",
		BankAccountID:   nil,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if bankAccountID != "" {
		payment.BankAccountID = &bankAccountID
	}

	if invoiceID != "" {
		payment.InvoiceID = &invoiceID
		inv, _, err := s.invoices.GetByID(ctx, invoiceID)
		if err != nil {
			return nil, err
		}
		inv.Status = "PAID"
		_ = s.invoices.Update(ctx, inv)

		// Publish invoice paid event
		if err := s.publisher.Publish(ctx, domain.TopicFinInvoicePaid, inv.ID, domain.InvoiceEventPayload{
			ID:            inv.ID,
			CustomerID:     inv.CustomerID,
			InvoiceNumber:  inv.InvoiceNumber,
			TotalAmount:    inv.TotalAmount,
			Status:         inv.Status,
			Timestamp:      time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinInvoicePaid, err)
		}
	}
	if billID != "" {
		payment.BillID = &billID
	}

	err := s.payments.Create(ctx, payment)
	if err != nil {
		// Publish payment failed event
		if err := s.publisher.Publish(ctx, domain.TopicFinPaymentFailed, payment.ID, domain.PaymentEventPayload{
			ID:            payment.ID,
			InvoiceID:     payment.InvoiceID,
			BillID:        payment.BillID,
			PaymentNumber: payment.PaymentNumber,
			Amount:        payment.Amount,
			PaymentMethod: payment.PaymentMethod,
			Status:        "FAILED",
			Timestamp:     time.Now(),
		}); err != nil {
			log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinPaymentFailed, err)
		}
		return nil, err
	}

	// Publish payment received and processed events
	if err := s.publisher.Publish(ctx, domain.TopicFinPaymentReceived, payment.ID, domain.PaymentEventPayload{
		ID:            payment.ID,
		InvoiceID:     payment.InvoiceID,
		BillID:        payment.BillID,
		PaymentNumber: payment.PaymentNumber,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinPaymentReceived, err)
	}

	if err := s.publisher.Publish(ctx, domain.TopicFinPaymentProcessed, payment.ID, domain.PaymentEventPayload{
		ID:            payment.ID,
		InvoiceID:     payment.InvoiceID,
		BillID:        payment.BillID,
		PaymentNumber: payment.PaymentNumber,
		Amount:        payment.Amount,
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
		Timestamp:     time.Now(),
	}); err != nil {
		log.Printf("ERROR: failed to publish event %s: %v", domain.TopicFinPaymentProcessed, err)
	}

	return payment, nil
}

func (s *CashManagementService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.payments.GetByID(ctx, id)
}

func (s *CashManagementService) ReconcileBankStatement(ctx context.Context, statementID string) error {
	// Bank reconciliation logic
	return nil
}

func (s *CashManagementService) GetCashFlowForecast(ctx context.Context, monthsAhead int) (map[string]interface{}, error) {
	// Simple forecasting mock
	return map[string]interface{}{
		"forecast_period_months": monthsAhead,
		"projected_cash_inflow":  decimal.NewFromInt(125000),
		"projected_cash_outflow": decimal.NewFromInt(80000),
		"net_cash_flow":          decimal.NewFromInt(45000),
	}, nil
}
