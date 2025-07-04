// File: services/financial-management/models/vendor.go
package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

// Vendor represents suppliers we owe money to
type Vendor struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:20" validate:"required,min=3,max=20"`
	Name         string         `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	Email        string         `json:"email" gorm:"size:100" validate:"omitempty,email"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Address      string         `json:"address" gorm:"size:500"`
	Balance      float64        `json:"balance" gorm:"default:0"` // Amount we owe to vendor
	PaymentTerms int            `json:"payment_terms" gorm:"default:30" validate:"min=0,max=365"` // Days
	TaxID        string         `json:"tax_id" gorm:"size:50"`
	BankAccount  string         `json:"bank_account" gorm:"size:100"`
	Category     string         `json:"category" gorm:"size:50"` // Office Supplies, Raw Materials, etc.
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Invoices []Invoice `json:"-" gorm:"foreignKey:VendorID"`
}

// Business logic methods for Vendor
func (v *Vendor) GetOwedAmount() float64 {
	return v.Balance
}

func (v *Vendor) GetPaymentStatus() string {
	if v.Balance <= 0 {
		return "Paid Up"
	} else if v.Balance > 0 {
		return "Outstanding"
	}
	return "Unknown"
}

func (v *Vendor) GetPaymentTermsDescription() string {
	if v.PaymentTerms == 0 {
		return "Due on Receipt"
	}
	return fmt.Sprintf("Net %d days", v.PaymentTerms)
}

func (v *Vendor) GetPaymentPriority() string {
	if v.Balance <= 0 {
		return "None"
	}
	if v.Balance >= 10000 {
		return "High"
	} else if v.Balance >= 5000 {
		return "Medium"
	}
	return "Low"
}

func (v *Vendor) GetCategoryDescription() string {
	if v.Category == "" {
		return "General Vendor"
	}
	return v.Category
}

// Database table name
func (Vendor) TableName() string {
	return "vendors"
}