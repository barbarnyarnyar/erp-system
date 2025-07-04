// File: services/financial-management/models/invoice.go
package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

// Invoice represents both customer invoices (AR) and vendor bills (AP)
type Invoice struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Number         string         `json:"number" gorm:"unique;not null;size:50" validate:"required"`
	Type           InvoiceType    `json:"type" gorm:"not null" validate:"required"`
	CustomerID     *uint          `json:"customer_id,omitempty" gorm:"index"` // For receivables
	VendorID       *uint          `json:"vendor_id,omitempty" gorm:"index"`   // For payables
	Amount         float64        `json:"amount" gorm:"not null" validate:"required,gt=0"`
	TaxAmount      float64        `json:"tax_amount" gorm:"default:0" validate:"min=0"`
	TotalAmount    float64        `json:"total_amount" gorm:"not null"` // Amount + TaxAmount
	PaidAmount     float64        `json:"paid_amount" gorm:"default:0" validate:"min=0"`
	Balance        float64        `json:"balance" gorm:"default:0"`
	Status         InvoiceStatus  `json:"status" gorm:"default:'pending'"`
	IssueDate      time.Time      `json:"issue_date" gorm:"not null" validate:"required"`
	DueDate        time.Time      `json:"due_date" gorm:"not null" validate:"required"`
	Description    string         `json:"description" gorm:"size:500"`
	Reference      string         `json:"reference" gorm:"size:100"` // PO number, SO number, etc.
	SourceService  string         `json:"source_service" gorm:"size:20"` // HR, SCM, CRM, PM, M
	SourceID       string         `json:"source_id" gorm:"size:50"` // External service ID
	Currency       string         `json:"currency" gorm:"default:'USD';size:3"`
	ExchangeRate   float64        `json:"exchange_rate" gorm:"default:1.0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Customer       *Customer      `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Vendor         *Vendor        `json:"vendor,omitempty" gorm:"foreignKey:VendorID"`
	JournalEntries []JournalEntry `json:"-" gorm:"foreignKey:SourceID;references:Number"`
}

// InvoiceType enum for invoice types
type InvoiceType string

const (
	InvoiceTypeReceivable InvoiceType = "receivable" // Customer owes us (AR)
	InvoiceTypePayable    InvoiceType = "payable"    // We owe vendor (AP)
)

// InvoiceStatus enum for invoice status
type InvoiceStatus string

const (
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusPartial   InvoiceStatus = "partial"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusVoid      InvoiceStatus = "void"
)

// Business logic methods for Invoice
func (i *Invoice) IsOverdue() bool {
	return time.Now().After(i.DueDate) && i.Status != InvoiceStatusPaid && i.Status != InvoiceStatusCancelled
}

func (i *Invoice) UpdateBalance() {
	i.Balance = i.TotalAmount - i.PaidAmount
	i.updateStatus()
}

func (i *Invoice) updateStatus() {
	if i.Balance <= 0 {
		i.Status = InvoiceStatusPaid
	} else if i.PaidAmount > 0 {
		i.Status = InvoiceStatusPartial
	} else if i.IsOverdue() {
		i.Status = InvoiceStatusOverdue
	} else {
		i.Status = InvoiceStatusPending
	}
}

func (i *Invoice) CalculateDaysOverdue() int {
	if !i.IsOverdue() {
		return 0
	}
	return int(time.Since(i.DueDate).Hours() / 24)
}

func (i *Invoice) ApplyPayment(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}
	if amount > i.Balance {
		return fmt.Errorf("payment amount cannot exceed balance")
	}
	
	i.PaidAmount += amount
	i.UpdateBalance()
	return nil
}

func (i *Invoice) CalculateTotalAmount() {
	i.TotalAmount = i.Amount + i.TaxAmount
	i.Balance = i.TotalAmount - i.PaidAmount
}

func (i *Invoice) GetAgingCategory() string {
	if !i.IsOverdue() {
		return "Current"
	}
	
	daysOverdue := i.CalculateDaysOverdue()
	if daysOverdue <= 30 {
		return "1-30 Days"
	} else if daysOverdue <= 60 {
		return "31-60 Days"
	} else if daysOverdue <= 90 {
		return "61-90 Days"
	}
	return "Over 90 Days"
}

func (i *Invoice) GetPaymentPercentage() float64 {
	if i.TotalAmount <= 0 {
		return 0
	}
	return (i.PaidAmount / i.TotalAmount) * 100
}

func (i *Invoice) IsReceivable() bool {
	return i.Type == InvoiceTypeReceivable
}

func (i *Invoice) IsPayable() bool {
	return i.Type == InvoiceTypePayable
}

// Database table name
func (Invoice) TableName() string {
	return "invoices"
}