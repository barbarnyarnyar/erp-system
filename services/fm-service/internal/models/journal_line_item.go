// File: services/financial-management/models/journal_line_item.go
package models

import (
	"fmt"
	"gorm.io/gorm"
)

// JournalLineItem represents individual debit/credit lines
type JournalLineItem struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	JournalEntryID uint    `json:"journal_entry_id" gorm:"not null;index" validate:"required"`
	AccountID      uint    `json:"account_id" gorm:"not null;index" validate:"required"`
	Debit          float64 `json:"debit" gorm:"default:0" validate:"min=0"`
	Credit         float64 `json:"credit" gorm:"default:0" validate:"min=0"`
	Description    string  `json:"description" gorm:"size:500"`
	Reference      string  `json:"reference" gorm:"size:100"`

	// Future extensions for dimensional accounting
	CostCenterID *uint  `json:"cost_center_id,omitempty" gorm:"index"` // For cost tracking
	ProjectID    *uint  `json:"project_id,omitempty" gorm:"index"`     // For project accounting
	DepartmentID string `json:"department_id,omitempty" gorm:"size:20;index"` // From HR service

	// Relationships
	JournalEntry *JournalEntry `json:"-" gorm:"foreignKey:JournalEntryID"`
	Account      *Account      `json:"account,omitempty" gorm:"foreignKey:AccountID"`
}

// Business logic methods for JournalLineItem
func (jli *JournalLineItem) Validate() error {
	// Must have either debit or credit, but not both
	if jli.Debit > 0 && jli.Credit > 0 {
		return fmt.Errorf("line item cannot have both debit and credit amounts")
	}
	if jli.Debit == 0 && jli.Credit == 0 {
		return fmt.Errorf("line item must have either debit or credit amount")
	}
	return nil
}

func (jli *JournalLineItem) IsDebit() bool {
	return jli.Debit > 0
}

func (jli *JournalLineItem) IsCredit() bool {
	return jli.Credit > 0
}

func (jli *JournalLineItem) GetAmount() float64 {
	if jli.IsDebit() {
		return jli.Debit
	}
	return jli.Credit
}

func (jli *JournalLineItem) GetType() string {
	if jli.IsDebit() {
		return "Debit"
	}
	return "Credit"
}

func (jli *JournalLineItem) GetFormattedAmount() string {
	amount := jli.GetAmount()
	if jli.IsDebit() {
		return fmt.Sprintf("Dr. %.2f", amount)
	}
	return fmt.Sprintf("Cr. %.2f", amount)
}

func (jli *JournalLineItem) SetDebit(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("debit amount cannot be negative")
	}
	jli.Debit = amount
	jli.Credit = 0 // Clear credit when setting debit
	return nil
}

func (jli *JournalLineItem) SetCredit(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("credit amount cannot be negative")
	}
	jli.Credit = amount
	jli.Debit = 0 // Clear debit when setting credit
	return nil
}

func (jli *JournalLineItem) HasDimensions() bool {
	return jli.CostCenterID != nil || jli.ProjectID != nil || jli.DepartmentID != ""
}

func (jli *JournalLineItem) GetDimensionInfo() string {
	var dimensions []string
	
	if jli.CostCenterID != nil {
		dimensions = append(dimensions, fmt.Sprintf("CC:%d", *jli.CostCenterID))
	}
	if jli.ProjectID != nil {
		dimensions = append(dimensions, fmt.Sprintf("Proj:%d", *jli.ProjectID))
	}
	if jli.DepartmentID != "" {
		dimensions = append(dimensions, fmt.Sprintf("Dept:%s", jli.DepartmentID))
	}
	
	if len(dimensions) == 0 {
		return "No Dimensions"
	}
	
	return fmt.Sprintf("[%s]", fmt.Sprintf("%v", dimensions))
}

// Database table name
func (JournalLineItem) TableName() string {
	return "journal_line_items"
}