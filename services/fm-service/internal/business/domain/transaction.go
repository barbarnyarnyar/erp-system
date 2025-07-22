package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionPending  TransactionStatus = "PENDING"
	TransactionPosted   TransactionStatus = "POSTED"
	TransactionReversed TransactionStatus = "REVERSED"
)

type Transaction struct {
	ID          string               `json:"id"`
	Reference   string               `json:"reference"`
	Date        time.Time            `json:"date"`
	Description string               `json:"description"`
	Status      TransactionStatus    `json:"status"`
	CreatedBy   string               `json:"created_by"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Lines       []TransactionLine    `json:"lines"`
}

type TransactionLine struct {
	ID            string          `json:"id"`
	TransactionID string          `json:"transaction_id"`
	AccountID     string          `json:"account_id"`
	DebitAmount   decimal.Decimal `json:"debit_amount"`
	CreditAmount  decimal.Decimal `json:"credit_amount"`
	Description   string          `json:"description"`
	CreatedAt     time.Time       `json:"created_at"`
}

func NewTransaction(reference, description, createdBy string) *Transaction {
	return &Transaction{
		Reference:   reference,
		Date:        time.Now(),
		Description: description,
		Status:      TransactionPending,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Lines:       make([]TransactionLine, 0),
	}
}

func (t *Transaction) AddLine(accountID string, amount decimal.Decimal, isDebit bool, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}

	line := TransactionLine{
		TransactionID: t.ID,
		AccountID:     accountID,
		Description:   description,
		CreatedAt:     time.Now(),
	}

	if isDebit {
		line.DebitAmount = amount
		line.CreditAmount = decimal.Zero
	} else {
		line.DebitAmount = decimal.Zero
		line.CreditAmount = amount
	}

	t.Lines = append(t.Lines, line)
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) IsBalanced() bool {
	totalDebits := decimal.Zero
	totalCredits := decimal.Zero

	for _, line := range t.Lines {
		totalDebits = totalDebits.Add(line.DebitAmount)
		totalCredits = totalCredits.Add(line.CreditAmount)
	}

	return totalDebits.Equal(totalCredits)
}

func (t *Transaction) Post() error {
	if t.Status != TransactionPending {
		return errors.New("only pending transactions can be posted")
	}

	if !t.IsBalanced() {
		return errors.New("transaction must be balanced before posting")
	}

	if len(t.Lines) < 2 {
		return errors.New("transaction must have at least two lines")
	}

	t.Status = TransactionPosted
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) Reverse() error {
	if t.Status != TransactionPosted {
		return errors.New("only posted transactions can be reversed")
	}

	t.Status = TransactionReversed
	t.UpdatedAt = time.Now()
	return nil
}