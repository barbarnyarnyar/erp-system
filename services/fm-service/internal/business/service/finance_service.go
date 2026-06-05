package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/erp-system/fm-service/internal/business/domain"
	"github.com/shopspring/decimal"
)

// FinanceService orchestrates the domain models and repositories
type FinanceService struct {
	accounts     domain.AccountRepository
	entries      domain.JournalEntryRepository
	invoices     domain.InvoiceRepository
	payments     domain.PaymentRepository
	vendors      domain.VendorRepository
	budgets      domain.BudgetRepository
}

// NewFinanceService creates a new finance application service
func NewFinanceService(
	accounts domain.AccountRepository,
	entries domain.JournalEntryRepository,
	invoices domain.InvoiceRepository,
	payments domain.PaymentRepository,
	vendors domain.VendorRepository,
	budgets domain.BudgetRepository,
) *FinanceService {
	return &FinanceService{
		accounts:     accounts,
		entries:      entries,
		invoices:     invoices,
		payments:     payments,
		vendors:      vendors,
		budgets:      budgets,
	}
}

// -----------------------------------------------------------------
// Account Management
// -----------------------------------------------------------------

func (s *FinanceService) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	return s.accounts.List(ctx)
}

func (s *FinanceService) CreateAccount(ctx context.Context, accNum, name, accType, parentID, currency string) (*domain.Account, error) {
	if accNum == "" || name == "" || accType == "" {
		return nil, errors.New("account number, name, and type are required")
	}

	id := fmt.Sprintf("acc_%d", time.Now().UnixNano())
	acc := &domain.Account{
		ID:            id,
		AccountNumber: accNum,
		Name:          name,
		Type:          domain.AccountType(accType),
		Balance:       decimal.Zero,
		Currency:      currency,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if parentID != "" {
		acc.ParentID = &parentID
	}

	err := s.accounts.Create(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *FinanceService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.GetByID(ctx, id)
}

func (s *FinanceService) UpdateAccount(ctx context.Context, id, name, accType, parentID string, isActive bool) (*domain.Account, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	acc.Name = name
	acc.Type = domain.AccountType(accType)
	acc.IsActive = isActive
	acc.UpdatedAt = time.Now()

	if parentID != "" {
		acc.ParentID = &parentID
	} else {
		acc.ParentID = nil
	}

	err = s.accounts.Update(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *FinanceService) DeleteAccount(ctx context.Context, id string) error {
	return s.accounts.Delete(ctx, id)
}

func (s *FinanceService) GetAccountBalance(ctx context.Context, id string) (decimal.Decimal, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return decimal.Zero, err
	}
	return acc.Balance, nil
}

// -----------------------------------------------------------------
// Journal Entries (GL)
// -----------------------------------------------------------------

func (s *FinanceService) ListJournalEntries(ctx context.Context) ([]domain.JournalEntry, error) {
	return s.entries.List(ctx)
}

func (s *FinanceService) CreateJournalEntry(ctx context.Context, ref, desc string, lines []domain.JournalEntryLine) (*domain.JournalEntry, error) {
	if len(lines) < 2 {
		return nil, errors.New("a journal entry must have at least 2 lines")
	}

	// Validate balance (Debits must equal Credits)
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero
	for _, l := range lines {
		totalDebits = totalDebits.Add(l.DebitAmount)
		totalCredits = totalCredits.Add(l.CreditAmount)
	}

	if !totalDebits.Equal(totalCredits) {
		return nil, fmt.Errorf("journal entry is unbalanced: debits=%s, credits=%s", totalDebits, totalCredits)
	}

	id := fmt.Sprintf("je_%d", time.Now().UnixNano())
	entry := &domain.JournalEntry{
		ID:          id,
		Reference:   ref,
		Date:        time.Now(),
		Description: desc,
		Status:      "POSTED", // Auto-post for simplicity in mock
		CreatedBy:   "system",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Post updates to account balances
	for i, l := range lines {
		lines[i].ID = fmt.Sprintf("jel_%d_%d", time.Now().UnixNano(), i)
		lines[i].EntryID = id

		acc, err := s.accounts.GetByID(ctx, l.AccountID)
		if err != nil {
			return nil, fmt.Errorf("account not found: %s", l.AccountID)
		}

		// Update balance depending on account type
		// Assets/Expenses: Debits increase, Credits decrease
		// Liabilities/Equity/Revenue: Debits decrease, Credits increase
		isDebitIncreaseType := acc.Type == "ASSET" || acc.Type == "EXPENSE"
		if isDebitIncreaseType {
			acc.Balance = acc.Balance.Add(l.DebitAmount).Sub(l.CreditAmount)
		} else {
			acc.Balance = acc.Balance.Sub(l.DebitAmount).Add(l.CreditAmount)
		}

		err = s.accounts.Update(ctx, acc)
		if err != nil {
			return nil, fmt.Errorf("failed to update balance for account %s: %w", acc.ID, err)
		}
	}

	err := s.entries.Create(ctx, entry, lines)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *FinanceService) GetJournalEntry(ctx context.Context, id string) (*domain.JournalEntry, []domain.JournalEntryLine, error) {
	return s.entries.GetByID(ctx, id)
}

// -----------------------------------------------------------------
// Invoicing & Payments
// -----------------------------------------------------------------

func (s *FinanceService) ListInvoices(ctx context.Context) ([]domain.Invoice, error) {
	return s.invoices.List(ctx)
}

func (s *FinanceService) CreateInvoice(ctx context.Context, customerID string, dueDate time.Time, lines []domain.InvoiceLine) (*domain.Invoice, error) {
	id := fmt.Sprintf("inv_%d", time.Now().UnixNano())
	invNum := fmt.Sprintf("INV-%d", time.Now().Unix())
	
	total := decimal.Zero
	for _, l := range lines {
		total = total.Add(l.LineTotal)
	}

	inv := &domain.Invoice{
		ID:             id,
		CustomerID:     customerID,
		InvoiceNumber:  invNum,
		IssueDate:      time.Now(),
		DueDate:        dueDate,
		TotalAmount:    total,
		Status:         "SENT",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.invoices.Create(ctx, inv, lines)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *FinanceService) RecordPayment(ctx context.Context, invoiceID, billID string, amount decimal.Decimal, method string) (*domain.Payment, error) {
	id := fmt.Sprintf("pay_%d", time.Now().UnixNano())
	payNum := fmt.Sprintf("PAY-%d", time.Now().Unix())

	payment := &domain.Payment{
		ID:            id,
		PaymentNumber: payNum,
		PaymentDate:   time.Now(),
		Amount:        amount,
		PaymentMethod: method,
		Status:        "COMPLETED",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if invoiceID != "" {
		payment.InvoiceID = &invoiceID
		inv, _, err := s.invoices.GetByID(ctx, invoiceID)
		if err != nil {
			return nil, err
		}
		inv.Status = "PAID"
		_ = s.invoices.Update(ctx, inv)
	}

	err := s.payments.Create(ctx, payment)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

// -----------------------------------------------------------------
// Reporting & Budgets
// -----------------------------------------------------------------

func (s *FinanceService) GetBalanceSheet(ctx context.Context) (map[string]interface{}, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}

	assets := make(map[string]decimal.Decimal)
	liabilities := make(map[string]decimal.Decimal)
	equity := make(map[string]decimal.Decimal)
	
	var totalAssets, totalLiabilities, totalEquity decimal.Decimal

	for _, a := range accs {
		if !a.IsActive {
			continue
		}
		switch a.Type {
		case "ASSET":
			assets[a.Name] = a.Balance
			totalAssets = totalAssets.Add(a.Balance)
		case "LIABILITY":
			liabilities[a.Name] = a.Balance
			totalLiabilities = totalLiabilities.Add(a.Balance)
		case "EQUITY":
			equity[a.Name] = a.Balance
			totalEquity = totalEquity.Add(a.Balance)
		}
	}

	return map[string]interface{}{
		"assets":            assets,
		"total_assets":      totalAssets,
		"liabilities":       liabilities,
		"total_liabilities": totalLiabilities,
		"equity":            equity,
		"total_equity":      totalEquity,
	}, nil
}
