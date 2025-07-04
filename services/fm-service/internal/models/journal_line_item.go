package models

import JournalEntry "services/fm-service/internal/models/journal_entry"	
import Account "services/fm-service/internal/models/account"	
import "fmt"	
import "time"	
import "gorm.io/gorm"	

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