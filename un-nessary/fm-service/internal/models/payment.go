// File: services/financial-management/models/payment.go
package models

import (
	"fmt"
	"time"
)

// Payment represents money received from customers or paid to vendors
type Payment struct {
	BaseModel
	Auditable
	Trackable
	Number       string        `json:"number" gorm:"unique;not null;size:50" validate:"required"`
	Type         PaymentType   `json:"type" gorm:"not null" validate:"required"`
	PartyID      uint          `json:"party_id" gorm:"not null;index" validate:"required"`
	InvoiceID    *uint         `json:"invoice_id,omitempty" gorm:"index"` // Optional - for invoice payments
	Amount       float64       `json:"amount" gorm:"not null" validate:"required,gt=0"`
	PaymentDate  time.Time     `json:"payment_date" gorm:"not null" validate:"required"`
	Method       PaymentMethod `json:"method" gorm:"not null" validate:"required"`
	Status       PaymentStatus `json:"status" gorm:"default:'pending'"`
	
	// Payment details
	Currency        string `json:"currency" gorm:"default:'USD';size:3"`
	ExchangeRate    float64 `json:"exchange_rate" gorm:"default:1.0"`
	BankAccount     string `json:"bank_account" gorm:"size:100"`
	CheckNumber     string `json:"check_number" gorm:"size:50"`
	TransactionRef  string `json:"transaction_ref" gorm:"size:100"`
	ConfirmationRef string `json:"confirmation_ref" gorm:"size:100"`
	
	// Processing details
	ProcessedDate   *time.Time `json:"processed_date,omitempty"`
	ClearedDate     *time.Time `json:"cleared_date,omitempty"`
	FailureReason   string     `json:"failure_reason" gorm:"size:500"`
	
	// Fees and charges
	ProcessingFee   float64 `json:"processing_fee" gorm:"default:0"`
	BankCharges     float64 `json:"bank_charges" gorm:"default:0"`
	NetAmount       float64 `json:"net_amount" gorm:"default:0"` // Amount - fees
	
	// Optional fields
	Description string `json:"description" gorm:"size:500"`
	Notes       string `json:"notes" gorm:"size:1000"`

	// Relationships
	Party   *Party   `json:"party,omitempty" gorm:"foreignKey:PartyID"`
	Invoice *Invoice `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	Entries []JournalEntry `json:"-" gorm:"foreignKey:SourceID;references:Number"`
}

// PaymentType enum for payment direction
type PaymentType string

const (
	PaymentTypeReceived PaymentType = "received" // Money received from customers
	PaymentTypePaid     PaymentType = "paid"     // Money paid to vendors
)

// PaymentMethod enum for payment methods
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodWire         PaymentMethod = "wire"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodOnline       PaymentMethod = "online"
	PaymentMethodMobileApp    PaymentMethod = "mobile_app"
	PaymentMethodCrypto       PaymentMethod = "cryptocurrency"
)

// PaymentStatus enum for payment status
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusCleared    PaymentStatus = "cleared"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusReversed   PaymentStatus = "reversed"
)

// Business methods for type checking
func (p *Payment) IsReceived() bool {
	return p.Type == PaymentTypeReceived
}

func (p *Payment) IsPaid() bool {
	return p.Type == PaymentTypePaid
}

// Status checking methods
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted || p.Status == PaymentStatusCleared
}

func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

func (p *Payment) IsReversed() bool {
	return p.Status == PaymentStatusReversed
}

func (p *Payment) CanBeModified() bool {
	return p.Status == PaymentStatusPending
}

func (p *Payment) CanBeCancelled() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusProcessing
}

func (p *Payment) CanBeReversed() bool {
	return p.Status == PaymentStatusCompleted || p.Status == PaymentStatusCleared
}

// Amount calculation methods
func (p *Payment) CalculateNetAmount() {
	p.NetAmount = p.Amount - p.ProcessingFee - p.BankCharges
}

func (p *Payment) GetTotalFees() float64 {
	return p.ProcessingFee + p.BankCharges
}

func (p *Payment) GetEffectiveAmount() float64 {
	// For received payments, subtract fees. For paid payments, add fees to total cost
	if p.IsReceived() {
		return p.NetAmount
	}
	return p.Amount + p.GetTotalFees()
}

// Payment method information
func (p *Payment) IsElectronic() bool {
	electronic := []PaymentMethod{
		PaymentMethodBankTransfer,
		PaymentMethodWire,
		PaymentMethodCreditCard,
		PaymentMethodDebitCard,
		PaymentMethodOnline,
		PaymentMethodMobileApp,
		PaymentMethodCrypto,
	}
	
	for _, method := range electronic {
		if p.Method == method {
			return true
		}
	}
	return false
}

func (p *Payment) RequiresClearing() bool {
	return p.Method == PaymentMethodCheck || p.Method == PaymentMethodBankTransfer
}

func (p *Payment) GetMethodDescription() string {
	switch p.Method {
	case PaymentMethodCash:
		return "Cash Payment"
	case PaymentMethodCheck:
		return fmt.Sprintf("Check #%s", p.CheckNumber)
	case PaymentMethodBankTransfer:
		return "Bank Transfer"
	case PaymentMethodWire:
		return "Wire Transfer"
	case PaymentMethodCreditCard:
		return "Credit Card"
	case PaymentMethodDebitCard:
		return "Debit Card"
	case PaymentMethodOnline:
		return "Online Payment"
	case PaymentMethodMobileApp:
		return "Mobile App Payment"
	case PaymentMethodCrypto:
		return "Cryptocurrency"
	default:
		return string(p.Method)
	}
}

// Status transition methods
func (p *Payment) Process() error {
	if p.Status != PaymentStatusPending {
		return fmt.Errorf("can only process pending payments")
	}
	
	p.Status = PaymentStatusProcessing
	now := time.Now()
	p.ProcessedDate = &now
	p.CalculateNetAmount()
	return nil
}

func (p *Payment) Complete() error {
	if p.Status != PaymentStatusProcessing {
		return fmt.Errorf("can only complete processing payments")
	}
	
	p.Status = PaymentStatusCompleted
	return nil
}

func (p *Payment) Clear() error {
	if p.Status != PaymentStatusCompleted {
		return fmt.Errorf("can only clear completed payments")
	}
	
	p.Status = PaymentStatusCleared
	now := time.Now()
	p.ClearedDate = &now
	return nil
}

func (p *Payment) Fail(reason string) error {
	if p.Status == PaymentStatusCompleted || p.Status == PaymentStatusCleared {
		return fmt.Errorf("cannot fail completed or cleared payments")
	}
	
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	return nil
}

func (p *Payment) Cancel(reason string) error {
	if !p.CanBeCancelled() {
		return fmt.Errorf("payment cannot be cancelled in current status: %s", p.Status)
	}
	
	p.Status = PaymentStatusCancelled
	p.FailureReason = reason
	return nil
}

func (p *Payment) Reverse(reason string) error {
	if !p.CanBeReversed() {
		return fmt.Errorf("payment cannot be reversed in current status: %s", p.Status)
	}
	
	p.Status = PaymentStatusReversed
	p.FailureReason = reason
	return nil
}

// Fee management methods
func (p *Payment) AddProcessingFee(fee float64) error {
	if fee < 0 {
		return fmt.Errorf("processing fee cannot be negative")
	}
	p.ProcessingFee = fee
	p.CalculateNetAmount()
	return nil
}

func (p *Payment) AddBankCharges(charges float64) error {
	if charges < 0 {
		return fmt.Errorf("bank charges cannot be negative")
	}
	p.BankCharges = charges
	p.CalculateNetAmount()
	return nil
}

// Invoice association methods
func (p *Payment) AssociateWithInvoice(invoiceID uint) {
	p.InvoiceID = &invoiceID
}

func (p *Payment) RemoveInvoiceAssociation() {
	p.InvoiceID = nil
}

func (p *Payment) IsAssociatedWithInvoice() bool {
	return p.InvoiceID != nil
}

// Timing methods
func (p *Payment) GetProcessingDays() int {
	if p.ProcessedDate == nil {
		return 0
	}
	return int(p.ProcessedDate.Sub(p.PaymentDate).Hours() / 24)
}

func (p *Payment) GetClearingDays() int {
	if p.ClearedDate == nil || p.ProcessedDate == nil {
		return 0
	}
	return int(p.ClearedDate.Sub(*p.ProcessedDate).Hours() / 24)
}

func (p *Payment) GetTotalProcessingDays() int {
	if p.ClearedDate == nil {
		return 0
	}
	return int(p.ClearedDate.Sub(p.PaymentDate).Hours() / 24)
}

// Implement Validator interface
func (p *Payment) Validate() error {
	if p.Number == "" {
		return fmt.Errorf("payment number is required")
	}
	if p.PartyID == 0 {
		return fmt.Errorf("party ID is required")
	}
	if p.Amount <= 0 {
		return fmt.Errorf("payment amount must be greater than zero")
	}
	if p.PaymentDate.After(time.Now()) {
		return fmt.Errorf("payment date cannot be in the future")
	}
	if p.ProcessingFee < 0 {
		return fmt.Errorf("processing fee cannot be negative")
	}
	if p.BankCharges < 0 {
		return fmt.Errorf("bank charges cannot be negative")
	}
	
	// Method-specific validations
	if p.Method == PaymentMethodCheck && p.CheckNumber == "" {
		return fmt.Errorf("check number is required for check payments")
	}
	
	return nil
}

// Implement Calculator interface
func (p *Payment) Calculate() {
	p.CalculateNetAmount()
}

func (Payment) TableName() string {
	return "payments"
}

// Factory methods for creating payments
func NewCustomerPayment(number string, partyID uint, amount float64, paymentDate time.Time, method PaymentMethod) *Payment {
	payment := &Payment{
		Number:       number,
		Type:         PaymentTypeReceived,
		PartyID:      partyID,
		Amount:       amount,
		PaymentDate:  paymentDate,
		Method:       method,
		Status:       PaymentStatusPending,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
	}
	payment.CalculateNetAmount()
	return payment
}

func NewVendorPayment(number string, partyID uint, amount float64, paymentDate time.Time, method PaymentMethod) *Payment {
	payment := &Payment{
		Number:       number,
		Type:         PaymentTypePaid,
		PartyID:      partyID,
		Amount:       amount,
		PaymentDate:  paymentDate,
		Method:       method,
		Status:       PaymentStatusPending,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
	}
	payment.CalculateNetAmount()
	return payment
}