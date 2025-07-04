package models

import (
	"time"
	"fmt"
	"gorm.io/gorm"
)

// Account represents the chart of accounts with standardized numbering
type Account struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"unique;not null;size:10" validate:"required,min=4,max=10"`
	Name        string         `json:"name" gorm:"not null;size:100" validate:"required,min=3,max=100"`
	Type        AccountType    `json:"type" gorm:"not null" validate:"required"`
	ParentCode  *string        `json:"parent_code,omitempty" gorm:"size:10"` // For sub-accounts
	Balance     float64        `json:"balance" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Description string         `json:"description" gorm:"size:255"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	JournalLineItems []JournalLineItem `json:"-" gorm:"foreignKey:AccountID"`
	ParentAccount    *Account          `json:"parent_account,omitempty" gorm:"foreignKey:ParentCode;references:Code"`
	SubAccounts      []Account         `json:"sub_accounts,omitempty" gorm:"foreignKey:ParentCode;references:Code"`
}

// AccountType enum following standard accounting categories
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"     // 1000-1999
	AccountTypeLiability AccountType = "liability" // 2000-2999
	AccountTypeEquity    AccountType = "equity"    // 3000-3999
	AccountTypeRevenue   AccountType = "revenue"   // 4000-4999
	AccountTypeExpense   AccountType = "expense"   // 5000-5999
	AccountTypeOther     AccountType = "other"     // 6000-6999
)

// Business logic methods for Account
func (a *Account) GetBalance() float64 {
	return a.Balance
}

func (a *Account) IsDebitAccount() bool {
	return a.Type == AccountTypeAsset || a.Type == AccountTypeExpense
}

func (a *Account) IsCreditAccount() bool {
	return a.Type == AccountTypeLiability || a.Type == AccountTypeEquity || a.Type == AccountTypeRevenue
}

func (a *Account) GetAccountCategory() string {
	switch a.Type {
	case AccountTypeAsset:
		return "Assets"
	case AccountTypeLiability:
		return "Liabilities"
	case AccountTypeEquity:
		return "Equity"
	case AccountTypeRevenue:
		return "Revenue"
	case AccountTypeExpense:
		return "Expenses"
	default:
		return "Other"
	}
}

// ValidateAccountCode ensures account code follows standard numbering
func (a *Account) ValidateAccountCode() error {
	if len(a.Code) < 4 {
		return fmt.Errorf("account code must be at least 4 digits")
	}
	
	// Check if code matches account type range
	firstDigit := a.Code[0]
	switch a.Type {
	case AccountTypeAsset:
		if firstDigit != '1' {
			return fmt.Errorf("asset accounts must start with 1 (1000-1999)")
		}
	case AccountTypeLiability:
		if firstDigit != '2' {
			return fmt.Errorf("liability accounts must start with 2 (2000-2999)")
		}
	case AccountTypeEquity:
		if firstDigit != '3' {
			return fmt.Errorf("equity accounts must start with 3 (3000-3999)")
		}
	case AccountTypeRevenue:
		if firstDigit != '4' {
			return fmt.Errorf("revenue accounts must start with 4 (4000-4999)")
		}
	case AccountTypeExpense:
		if firstDigit != '5' {
			return fmt.Errorf("expense accounts must start with 5 (5000-5999)")
		}
	case AccountTypeOther:
		if firstDigit != '6' {
			return fmt.Errorf("other accounts must start with 6 (6000-6999)")
		}
	}
	return nil
}