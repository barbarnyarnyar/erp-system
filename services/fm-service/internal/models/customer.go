package models

import (
	"time"
	"fmt"
	"gorm.io/gorm"
)

// Customer represents customers who owe us money
type Customer struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Code         string         `json:"code" gorm:"unique;not null;size:20" validate:"required,min=3,max=20"`
	Name         string         `json:"name" gorm:"not null;size:100" validate:"required,min=2,max=100"`
	Email        string         `json:"email" gorm:"size:100" validate:"omitempty,email"`
	Phone        string         `json:"phone" gorm:"size:20"`
	Address      string         `json:"address" gorm:"size:500"`
	CreditLimit  float64        `json:"credit_limit" gorm:"default:0" validate:"min=0"`
	Balance      float64        `json:"balance" gorm:"default:0"` // Outstanding amount owed by customer
	PaymentTerms int            `json:"payment_terms" gorm:"default:30" validate:"min=0,max=365"` // Days
	TaxID        string         `json:"tax_id" gorm:"size:50"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Invoices []Invoice `json:"-" gorm:"foreignKey:CustomerID"`
}

// Business logic methods for Customer
func (c *Customer) GetOutstandingBalance() float64 {
	return c.Balance
}

func (c *Customer) HasCreditAvailable(amount float64) bool {
	if c.CreditLimit <= 0 {
		return true // No credit limit set
	}
	return (c.Balance + amount) <= c.CreditLimit
}

func (c *Customer) IsOverCreditLimit() bool {
	if c.CreditLimit <= 0 {
		return false // No credit limit set
	}
	return c.Balance > c.CreditLimit
}

func (c *Customer) GetCreditUtilization() float64 {
	if c.CreditLimit <= 0 {
		return 0
	}
	return (c.Balance / c.CreditLimit) * 100
}

func (c *Customer) GetPaymentTermsDescription() string {
	if c.PaymentTerms == 0 {
		return "Due on Receipt"
	}
	return fmt.Sprintf("Net %d days", c.PaymentTerms)
}
