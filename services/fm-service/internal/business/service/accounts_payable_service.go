package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

type AccountsPayableService struct {
	vendors   domain.VendorRepository
	bills     domain.VendorBillRepository
	publisher domain.EventPublisher
}

func NewAccountsPayableService(vendors domain.VendorRepository, bills domain.VendorBillRepository, publisher domain.EventPublisher) *AccountsPayableService {
	return &AccountsPayableService{
		vendors:   vendors,
		bills:     bills,
		publisher: publisher,
	}
}

func (s *AccountsPayableService) ListVendors(ctx context.Context) ([]domain.Vendor, error) {
	return s.vendors.List(ctx)
}

func (s *AccountsPayableService) CreateVendor(ctx context.Context, code, name, contact, email, phone string) (*domain.Vendor, error) {
	if code == "" || name == "" {
		return nil, errors.New("vendor code and name are required")
	}

	id := fmt.Sprintf("ven_%d", time.Now().UnixNano())
	vendor := &domain.Vendor{
		ID:          id,
		VendorCode:  code,
		VendorName:  name,
		ContactName: contact,
		Email:       email,
		Phone:       phone,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.vendors.Create(ctx, vendor)
	if err != nil {
		return nil, err
	}

	// Publish event
	_ = s.publisher.Publish(ctx, "fin.vendor.created", vendor.ID, domain.VendorEventPayload{
		ID:         vendor.ID,
		VendorCode: vendor.VendorCode,
		VendorName: vendor.VendorName,
		Email:      vendor.Email,
		Timestamp:  time.Now(),
	})

	return vendor, nil
}

func (s *AccountsPayableService) GetVendor(ctx context.Context, id string) (*domain.Vendor, error) {
	return s.vendors.GetByID(ctx, id)
}

func (s *AccountsPayableService) UpdateVendor(ctx context.Context, id string, fields map[string]interface{}) (*domain.Vendor, error) {
	vendor, err := s.vendors.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := fields["name"].(string); ok {
		vendor.VendorName = name
	}
	if contact, ok := fields["contact_name"].(string); ok {
		vendor.ContactName = contact
	}
	if email, ok := fields["email"].(string); ok {
		vendor.Email = email
	}
	if phone, ok := fields["phone"].(string); ok {
		vendor.Phone = phone
	}
	vendor.UpdatedAt = time.Now()

	err = s.vendors.Update(ctx, vendor)
	if err != nil {
		return nil, err
	}

	// Publish event
	_ = s.publisher.Publish(ctx, "fin.vendor.updated", vendor.ID, domain.VendorEventPayload{
		ID:         vendor.ID,
		VendorCode: vendor.VendorCode,
		VendorName: vendor.VendorName,
		Email:      vendor.Email,
		Timestamp:  time.Now(),
	})

	return vendor, nil
}

func (s *AccountsPayableService) DeleteVendor(ctx context.Context, id string) error {
	return s.vendors.Delete(ctx, id)
}

func (s *AccountsPayableService) MatchPurchaseOrder(ctx context.Context, billID, poID, goodsReceiptID string) (bool, error) {
	// 3-way matching logic placeholder
	// Verifies that quantities and unit prices match across SCM PO, Warehouse Goods Receipt, and FM Bill.
	if billID == "" || poID == "" || goodsReceiptID == "" {
		return false, errors.New("bill ID, PO ID, and Goods Receipt ID are required for 3-way matching")
	}
	return true, nil
}

func (s *AccountsPayableService) CreateVendorBill(ctx context.Context, vendorID, billNum, poID string, issueDate, dueDate time.Time, total decimal.Decimal, lines []domain.VendorBillLine) (*domain.VendorBill, error) {
	id := fmt.Sprintf("bill_%d", time.Now().UnixNano())
	
	bill := &domain.VendorBill{
		ID:          id,
		VendorID:    vendorID,
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
	_ = s.publisher.Publish(ctx, "fin.vendor.payment.due", bill.ID, domain.VendorBillEventPayload{
		ID:         bill.ID,
		VendorID:   bill.VendorID,
		BillNumber: bill.BillNumber,
		Amount:     bill.TotalAmount,
		DueDate:    bill.DueDate,
		Timestamp:  time.Now(),
	})

	return bill, nil
}

