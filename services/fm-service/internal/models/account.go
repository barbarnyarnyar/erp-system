// File: services/financial-management/models/account.go
package models

import "fmt"

// Account represents the chart of accounts
type Account struct {
	BaseModel
	Code        string      `json:"code" gorm:"unique;not null;size:10" validate:"required"`
	Name        string      `json:"name" gorm:"not null;size:100" validate:"required"`
	Type        AccountType `json:"type" gorm:"not null" validate:"required"`
	ParentCode  *string     `json:"parent_code,omitempty" gorm:"size:10"`
	Balance     float64     `json:"balance" gorm:"default:0"`
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	Description string      `json:"description" gorm:"size:255"`

	// Future extension points
	CurrencyCode *string `json:"currency_code,omitempty" gorm:"size:3;default:'USD'"`
	TaxCode      *string `json:"tax_code,omitempty" gorm:"size:10"`
	
	// Optional fields for advanced accounting
	BankAccountNumber *string `json:"bank_account_number,omitempty" gorm:"size:50"`
	BankName          *string `json:"bank_name,omitempty" gorm:"size:100"`

	// Relationships
	LineItems    []JournalLineItem `json:"-" gorm:"foreignKey:AccountID"`
	Parent       *Account          `json:"parent,omitempty" gorm:"foreignKey:ParentCode;references:Code"`
	SubAccounts  []Account         `json:"sub_accounts,omitempty" gorm:"foreignKey:ParentCode;references:Code"`
}

// AccountType enum following standard accounting categories
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"     // 1000-1999
	AccountTypeLiability AccountType = "liability" // 2000-2999
	AccountTypeEquity    AccountType = "equity"    // 3000-3999
	AccountTypeRevenue   AccountType = "revenue"   // 4000-4999
	AccountTypeExpense   AccountType = "expense"   // 5000-5999
)

// Business methods
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

// ValidateCode ensures account code follows standard numbering
func (a *Account) ValidateCode() error {
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
	}
	return nil
}

func (a *Account) IsParent() bool {
	return len(a.SubAccounts) > 0
}

func (a *Account) IsSubAccount() bool {
	return a.ParentCode != nil
}

func (a *Account) GetFullPath() string {
	if a.Parent != nil {
		return fmt.Sprintf("%s > %s", a.Parent.Name, a.Name)
	}
	return a.Name
}

func (a *Account) UpdateBalance(amount float64, isDebit bool) {
	if a.IsDebitAccount() {
		if isDebit {
			a.Balance += amount
		} else {
			a.Balance -= amount
		}
	} else {
		if isDebit {
			a.Balance -= amount
		} else {
			a.Balance += amount
		}
	}
}

// Implement Validator interface
func (a *Account) Validate() error {
	if a.Code == "" {
		return fmt.Errorf("account code is required")
	}
	if a.Name == "" {
		return fmt.Errorf("account name is required")
	}
	return a.ValidateCode()
}

func (Account) TableName() string {
	return "accounts"
}

// Standard Chart of Accounts Seeder
func GetStandardChartOfAccounts() []Account {
	return []Account{
		// Assets (1000-1999)
		{Code: "1000", Name: "Cash", Type: AccountTypeAsset},
		{Code: "1010", Name: "Checking Account", Type: AccountTypeAsset, ParentCode: StringPtr("1000")},
		{Code: "1020", Name: "Savings Account", Type: AccountTypeAsset, ParentCode: StringPtr("1000")},
		{Code: "1100", Name: "Accounts Receivable", Type: AccountTypeAsset},
		{Code: "1200", Name: "Inventory", Type: AccountTypeAsset},
		{Code: "1300", Name: "Prepaid Expenses", Type: AccountTypeAsset},
		{Code: "1500", Name: "Equipment", Type: AccountTypeAsset},
		{Code: "1600", Name: "Accumulated Depreciation", Type: AccountTypeAsset},

		// Liabilities (2000-2999)
		{Code: "2000", Name: "Accounts Payable", Type: AccountTypeLiability},
		{Code: "2100", Name: "Short-term Loans", Type: AccountTypeLiability},
		{Code: "2200", Name: "Accrued Expenses", Type: AccountTypeLiability},
		{Code: "2300", Name: "Salaries Payable", Type: AccountTypeLiability},
		{Code: "2500", Name: "Long-term Debt", Type: AccountTypeLiability},

		// Equity (3000-3999)
		{Code: "3000", Name: "Owner's Equity", Type: AccountTypeEquity},
		{Code: "3100", Name: "Retained Earnings", Type: AccountTypeEquity},

		// Revenue (4000-4999)
		{Code: "4000", Name: "Sales Revenue", Type: AccountTypeRevenue},
		{Code: "4100", Name: "Service Revenue", Type: AccountTypeRevenue},
		{Code: "4200", Name: "Interest Income", Type: AccountTypeRevenue},

		// Expenses (5000-5999)
		{Code: "5000", Name: "Cost of Goods Sold", Type: AccountTypeExpense},
		{Code: "5100", Name: "Salary Expense", Type: AccountTypeExpense},
		{Code: "5200", Name: "Rent Expense", Type: AccountTypeExpense},
		{Code: "5300", Name: "Utilities Expense", Type: AccountTypeExpense},
		{Code: "5400", Name: "Office Supplies", Type: AccountTypeExpense},
		{Code: "5500", Name: "Depreciation Expense", Type: AccountTypeExpense},
	}
}