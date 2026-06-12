package service

import (
	"context"
	"erp-system/shared/utils"
	"errors"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type AccountsPayableService struct {
	bills  domain.ApVendorBillRepository
	outbox domain.TransactionalOutboxRepository
	tm     domain.TransactionManager
}

func NewAccountsPayableService(
	bills domain.ApVendorBillRepository,
	outbox domain.TransactionalOutboxRepository,
	tm domain.TransactionManager,
) *AccountsPayableService {
	return &AccountsPayableService{
		bills:  bills,
		outbox: outbox,
		tm:     tm,
	}
}

func (s *AccountsPayableService) MatchPurchaseOrder(ctx context.Context, billID, poID, goodsReceiptID string) (bool, error) {
	if billID == "" || poID == "" || goodsReceiptID == "" {
		return false, errors.New("bill ID, PO ID, and Goods Receipt ID are required for 3-way matching")
	}
	return true, nil
}

func (s *AccountsPayableService) CreateVendorBill(ctx context.Context, legalEntityID, vendorID, billNum, poID string, dueDate time.Time, total, tax decimal.Decimal) (*domain.ApVendorBill, error) {
	id := utils.NewID("bill")

	bill := &domain.ApVendorBill{
		ID:              id,
		LegalEntityID:   legalEntityID,
		BillNumber:      billNum,
		VendorID:        vendorID,
		PurchaseOrderID: poID,
		TotalAmount:     total,
		TaxAmount:       tax,
		DueDate:         dueDate,
		Status:          domain.PaymentStatusOPEN,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		err := s.bills.Create(txCtx, bill)
		if err != nil {
			return err
		}

		// Write to outbox
		outboxRec := &domain.TransactionalOutbox{
			ID:          utils.NewID("outbox"),
			EventType:   string(domain.TopicFmVendorPaymentDue),
			AggregateID: bill.ID,
			Payload: domain.VendorBillEventPayload{
				ID:         bill.ID,
				VendorID:   bill.VendorID,
				BillNumber: bill.BillNumber,
				Amount:     bill.TotalAmount,
				DueDate:    bill.DueDate,
				Timestamp:  time.Now(),
			},
			Status:    domain.OutboxStatusPENDING,
			CreatedAt: time.Now(),
		}
		return s.outbox.Create(txCtx, outboxRec)
	})

	if err != nil {
		return nil, err
	}

	return bill, nil
}

func (s *AccountsPayableService) ListVendorBills(ctx context.Context) ([]domain.ApVendorBill, error) {
	return s.bills.List(ctx)
}

func (s *AccountsPayableService) GetVendorBill(ctx context.Context, id string) (*domain.ApVendorBill, error) {
	return s.bills.GetByID(ctx, id)
}
