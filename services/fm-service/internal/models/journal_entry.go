package models

import (
	"time"
	"fmt"
	"gorm.io/gorm"
)
import JournalLineItem "services/fm-service/internal/models/journal_line_item"

// JournalEntry represents the header of a financial transaction
type JournalEntry struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Number      string            `json:"number" gorm:"unique;not null;size:50" validate:"required"`
	Date        time.Time         `json:"date" gorm:"not null" validate:"required"`
	Description string            `json:"description" gorm:"not null;size:500" validate:"required,min=3"`
	Reference   string            `json:"reference" gorm:"size:100"` // External reference
	TotalDebit  float64           `json:"total_debit" gorm:"default:0"`
	TotalCredit float64           `json:"total_credit" gorm:"default:0"`
	Status      JournalStatus     `json:"status" gorm:"default:'draft'"`
	SourceType  string            `json:"source_type" gorm:"size:20"` // Manual, HR, SCM, CRM, PM, M
	SourceID    string            `json:"source_id" gorm:"size:50"`   // External ID
	CreatedBy   string            `json:"created_by" gorm:"size:50"`  // User ID
	PostedBy    string            `json:"posted_by" gorm:"size:50"`   // User who posted
	PostedAt    *time.Time        `json:"posted_at,omitempty"`
	ReversedAt  *time.Time        `json:"reversed_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"-" gorm:"index"`

	// Relationships
	LineItems []JournalLineItem `json:"line_items" gorm:"foreignKey:JournalEntryID"`
}

// JournalStatus enum for journal entry status
type JournalStatus string

const (
	JournalStatusDraft    JournalStatus = "draft"
	JournalStatusPosted   JournalStatus = "posted"
	JournalStatusReversed JournalStatus = "reversed"
)

// Business logic methods for JournalEntry
func (je *JournalEntry) IsBalanced() bool {
	return je.TotalDebit == je.TotalCredit
}

func (je *JournalEntry) CalculateTotals() {
	je.TotalDebit = 0
	je.TotalCredit = 0
	for _, line := range je.LineItems {
		je.TotalDebit += line.Debit
		je.TotalCredit += line.Credit
	}
}

func (je *JournalEntry) Post(postedBy string) error {
	if je.Status != JournalStatusDraft {
		return fmt.Errorf("can only post draft journal entries")
	}
	if !je.IsBalanced() {
		return fmt.Errorf("journal entry must be balanced (debits = credits)")
	}
	if len(je.LineItems) < 2 {
		return fmt.Errorf("journal entry must have at least 2 line items")
	}
	
	now := time.Now()
	je.Status = JournalStatusPosted
	je.PostedBy = postedBy
	je.PostedAt = &now
	return nil
}

func (je *JournalEntry) Reverse(reversedBy string) error {
	if je.Status != JournalStatusPosted {
		return fmt.Errorf("can only reverse posted journal entries")
	}
	
	now := time.Now()
	je.Status = JournalStatusReversed
	je.ReversedAt = &now
	return nil
}

func (je *JournalEntry) Validate() error {
	if len(je.LineItems) < 2 {
		return fmt.Errorf("journal entry must have at least 2 line items")
	}
	
	je.CalculateTotals()
	if !je.IsBalanced() {
		return fmt.Errorf("journal entry must be balanced: debits (%.2f) must equal credits (%.2f)", 
			je.TotalDebit, je.TotalCredit)
	}
	
	return nil
}