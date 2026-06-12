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
	bills     domain.VendorBillRepository
	publisher domain.EventPublisher
}

func NewAccountsPayableService(bills domain.VendorBillRepository, publisher domain.EventPublisher) *AccountsPayableService {
	return &AccountsPayableService{
		bills:     bills,
		publisher: publisher,
	}
}

func (s *AccountsPayableService) MatchPurchaseOrder(ctx context.Context, billID, poID, goodsReceiptID string) (bool, error) {
	// 3-way matching logic placeholder
	// Verifies that quantities and unit prices match across SCM PO, Warehouse Goods Receipt, and FM Bill.
	if billID == "" || poID == "" || goodsReceiptID == "" {
		return false, errors.New("bill ID, PO ID, and Goods Receipt ID are required for 3-way matching")
	}
	return true, nil
}

func (s *AccountsPayableService) CreateVendorBill(ctx context.Context, supplierID, billNum, poID string, issueDate, dueDate time.Time, total decimal.Decimal, lines []domain.VendorBillLine) (*domain.VendorBill, error) {
	id := utils.NewID("bill")

	bill := &domain.VendorBill{
		ID:          id,
		SupplierID:  supplierID,
		BillNumber:  billNum,
		IssueDate:   issueDate,
		DueDate:     dueDate,
		TotalAmount: total,
		Status:      "DRAFT",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if poID != "" {
		bill.PurchaseOrderID = &poID
	}

	err := s.bills.Create(ctx, bill, lines)
	if err != nil {
		return nil, err
	}

	// Publish event
	if err := s.publisher.Publish(ctx, domain.TopicFmVendorPaymentDue, bill.ID, domain.VendorBillEventPayload{
		ID:         bill.ID,
		VendorID:   bill.SupplierID,
		BillNumber: bill.BillNumber,
		Amount:     bill.TotalAmount,
		DueDate:    bill.DueDate,
		Timestamp:  time.Now(),
	}); err != nil {
		utils.LogPublishErr("fm-service", domain.TopicFmVendorPaymentDue, err)
	}

	return bill, nil
}

func (s *AccountsPayableService) ListVendorBills(ctx context.Context) ([]domain.VendorBill, error) {
	return s.bills.List(ctx)
}

func (s *AccountsPayableService) GetVendorBill(ctx context.Context, id string) (*domain.VendorBill, []domain.VendorBillLine, error) {
	return s.bills.GetByID(ctx, id)
}
