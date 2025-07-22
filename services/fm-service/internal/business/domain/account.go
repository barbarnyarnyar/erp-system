package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type AccountType string

const (
	AssetAccount      AccountType = "ASSET"
	LiabilityAccount  AccountType = "LIABILITY"
	EquityAccount     AccountType = "EQUITY"
	RevenueAccount    AccountType = "REVENUE"
	ExpenseAccount    AccountType = "EXPENSE"
)

type Account struct {
	ID            string          `json:"id"`
	AccountNumber string          `json:"account_number"`
	Name          string          `json:"name"`
	Type          AccountType     `json:"type"`
	ParentID      *string         `json:"parent_id"`
	Balance       decimal.Decimal `json:"balance"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Children      []Account       `json:"children,omitempty"`
	Transactions  []Transaction   `json:"transactions,omitempty"`
}

func NewAccount(accountNumber, name string, accountType AccountType) *Account {
	return &Account{
		AccountNumber: accountNumber,
		Name:          name,
		Type:          accountType,
		Balance:       decimal.Zero,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (a *Account) Debit(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("debit amount must be positive")
	}

	switch a.Type {
	case AssetAccount, ExpenseAccount:
		a.Balance = a.Balance.Add(amount)
	case LiabilityAccount, EquityAccount, RevenueAccount:
		a.Balance = a.Balance.Sub(amount)
	}

	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Credit(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("credit amount must be positive")
	}

	switch a.Type {
	case AssetAccount, ExpenseAccount:
		a.Balance = a.Balance.Sub(amount)
	case LiabilityAccount, EquityAccount, RevenueAccount:
		a.Balance = a.Balance.Add(amount)
	}

	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Validate() error {
	if a.AccountNumber == "" {
		return errors.New("account number is required")
	}
	if a.Name == "" {
		return errors.New("account name is required")
	}
	if a.Type == "" {
		return errors.New("account type is required")
	}
	return nil
}