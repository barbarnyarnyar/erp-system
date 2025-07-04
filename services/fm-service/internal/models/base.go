package models

import (
	"time"
	"fmt"
	"gorm.io/gorm"
)


// Database table names (following convention)
func (Account) TableName() string          { return "accounts" }
func (Customer) TableName() string         { return "customers" }
func (Vendor) TableName() string           { return "vendors" }
func (Invoice) TableName() string          { return "invoices" }
func (JournalEntry) TableName() string     { return "journal_entries" }
func (JournalLineItem) TableName() string  { return "journal_line_items" }

// Standard Chart of Accounts Seeder (can be used for initial setup)
func GetStandardChartOfAccounts() []Account {
	return []Account{
		// Assets (1000-1999)
		{Code: "1000", Name: "Cash", Type: AccountTypeAsset},
		{Code: "1010", Name: "Checking Account", Type: AccountTypeAsset, ParentCode: stringPtr("1000")},
		{Code: "1020", Name: "Savings Account", Type: AccountTypeAsset, ParentCode: stringPtr("1000")},
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

		// Other (6000-6999)
		{Code: "6000", Name: "Interest Expense", Type: AccountTypeOther},
		{Code: "6100", Name: "Tax Expense", Type: AccountTypeOther},
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}