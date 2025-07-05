// File: services/financial-management/models/invoice.go
package models

import (
	"fmt"
	"time"
)

// Invoice handles both customer invoices (receivables) and vendor bills (payables)
type Invoice struct {
	BaseModel
	Auditable
	Trackable
	Number       string        `json:"number" gorm:"unique;not null;size:50" validate:"required"`
	Type         InvoiceType   `json:"type" gorm:"not null" validate:"required"`
	PartyID      uint          `json:"party_id" gorm:"not null;index" validate:"required"`
	IssueDate    time.Time     `json:"issue_date" gorm:"not null" validate:"required"`
	DueDate      time.Time     `json:"due_date" gorm:"not null" validate:"required"`
	
	// Amounts
	SubtotalAmount float64 `json:"subtotal_amount" gorm:"not null" validate:"required,gt=0"`
	TaxAmount      float64 `json:"tax_amount" gorm:"default:0" validate:"min=0"`
	DiscountAmount float64 `json:"discount_amount" gorm:"default:0" validate:"min=0"`
	TotalAmount    float64 `json:"total_amount" gorm:"not null" validate:"required,gt=0"`
	PaidAmount     float64 `json:"paid_amount" gorm:"default:0" validate:"min=0"`
	Balance        float64 `json:"balance" gorm:"default:0"`
	
	// Status and currency
	Status       InvoiceStatus `json:"status" gorm:"default:'draft'"`
	Currency     string        `json:"currency" gorm:"default:'USD';size:3"`
	ExchangeRate float64       `json:"exchange_rate" gorm:"default:1.0"`
	
	// Optional fields
	Description string `json:"description" gorm:"size:500"`
	Notes       string `json:"notes" gorm:"size:1000"`
	Terms       string `json:"terms" gorm:"size:1000"` // Payment terms and conditions
	
	// Extension points
	ProjectID      *uint   `json:"project_id,omitempty" gorm:"index"`      // For project billing
	ContractID     *uint   `json:"contract_id,omitempty" gorm:"index"`     // For contract billing
	PurchaseOrderRef *string `json:"purchase_order_ref,omitempty" gorm:"size:100"` // PO reference
	
	// Tax information
	TaxRate        float64 `json:"tax_rate" gorm:"default:0"`              // Overall tax rate
	TaxExempt      bool    `json:"tax_exempt" gorm:"default:false"`        // Tax exemption flag
	
	// Dates for tracking
	SentDate       *time.Time `json:"sent_date,omitempty"`
	PaidDate       *time.Time `json:"paid_date,omitempty"`
	
	// Relationships
	Party     *Party             `json:"party,omitempty" gorm:"foreignKey:PartyID"`
	LineItems []InvoiceLineItem  `json:"line_items" gorm:"foreignKey:InvoiceID"`
	Payments  []Payment          `json:"-" gorm:"foreignKey:InvoiceID"`
	Entries   []JournalEntry     `json:"-" gorm:"foreignKey:SourceID;references:Number"`
}

// InvoiceType enum for invoice types
type InvoiceType string

const (
	InvoiceTypeReceivable InvoiceType = "receivable" // Customer invoice (AR)
	InvoiceTypePayable    InvoiceType = "payable"    // Vendor bill (AP)
)

// InvoiceStatus enum for invoice status
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusPartial   InvoiceStatus = "partial"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusVoid      InvoiceStatus = "void"
)

// Business methods for type checking
func (i *Invoice) IsReceivable() bool {
	return i.Type == InvoiceTypeReceivable
}

func (i *Invoice) IsPayable() bool {
	return i.Type == InvoiceTypePayable
}

// Status checking methods
func (i *Invoice) IsOverdue() bool {
	return time.Now().After(i.DueDate) && 
		   i.Status != InvoiceStatusPaid && 
		   i.Status != InvoiceStatusCancelled && 
		   i.Status != InvoiceStatusVoid
}

func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

func (i *Invoice) IsPartiallyPaid() bool {
	return i.Status == InvoiceStatusPartial
}

func (i *Invoice) CanBeModified() bool {
	return i.Status == InvoiceStatusDraft
}

// Amount calculation methods
func (i *Invoice) CalculateAmounts() {
	i.SubtotalAmount = 0
	for _, line := range i.LineItems {
		line.CalculateAmount() // Calculate line amount first
		i.SubtotalAmount += line.Amount
	}
	
	// Calculate total with tax and discount
	i.TotalAmount = i.SubtotalAmount + i.TaxAmount - i.DiscountAmount
	i.Balance = i.TotalAmount - i.PaidAmount
	i.updateStatus()
}

func (i *Invoice) updateStatus() {
	if i.Balance <= 0 {
		i.Status = InvoiceStatusPaid
		if i.PaidDate == nil {
			now := time.Now()
			i.PaidDate = &now
		}
	} else if i.PaidAmount > 0 {
		i.Status = InvoiceStatusPartial
	} else if i.IsOverdue() {
		i.Status = InvoiceStatusOverdue
	} else if i.Status == InvoiceStatusDraft {
		// Keep draft status
	} else {
		i.Status = InvoiceStatusPending
	}
}

// Payment methods
func (i *Invoice) ApplyPayment(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}
	if amount > i.Balance {
		return fmt.Errorf("payment amount (%.2f) cannot exceed balance (%.2f)", amount, i.Balance)
	}
	
	i.PaidAmount += amount
	i.CalculateAmounts()
	return nil
}

func (i *Invoice) ReversePayment(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("reversal amount must be positive")
	}
	if amount > i.PaidAmount {
		return fmt.Errorf("reversal amount (%.2f) cannot exceed paid amount (%.2f)", amount, i.PaidAmount)
	}
	
	i.PaidAmount -= amount
	i.CalculateAmounts()
	return nil
}

// Aging and analysis methods
func (i *Invoice) GetDaysOverdue() int {
	if !i.IsOverdue() {
		return 0
	}
	return int(time.Since(i.DueDate).Hours() / 24)
}

func (i *Invoice) GetAgingCategory() string {
	if !i.IsOverdue() {
		return "Current"
	}
	
	daysOverdue := i.GetDaysOverdue()
	switch {
	case daysOverdue <= 30:
		return "1-30 Days"
	case daysOverdue <= 60:
		return "31-60 Days"
	case daysOverdue <= 90:
		return "61-90 Days"
	default:
		return "Over 90 Days"
	}
}

func (i *Invoice) GetPaymentPercentage() float64 {
	if i.TotalAmount <= 0 {
		return 0
	}
	return (i.PaidAmount / i.TotalAmount) * 100
}

func (i *Invoice) GetDaysUntilDue() int {
	if time.Now().After(i.DueDate) {
		return 0
	}
	return int(i.DueDate.Sub(time.Now()).Hours() / 24)
}

// Tax methods
func (i *Invoice) CalculateTax() {
	if i.TaxExempt {
		i.TaxAmount = 0
		return
	}
	
	if i.TaxRate > 0 {
		i.TaxAmount = i.SubtotalAmount * (i.TaxRate / 100)
	} else {
		// Calculate from line items
		i.TaxAmount = 0
		for _, line := range i.LineItems {
			i.TaxAmount += line.GetTaxAmount()
		}
	}
}

// Line item methods
func (i *Invoice) AddLineItem(lineItem *InvoiceLineItem) {
	lineItem.InvoiceID = i.ID
	i.LineItems = append(i.LineItems, *lineItem)
	i.CalculateAmounts()
}

func (i *Invoice) RemoveLineItem(lineItemID uint) {
	for idx, line := range i.LineItems {
		if line.ID == lineItemID {
			i.LineItems = append(i.LineItems[:idx], i.LineItems[idx+1:]...)
			break
		}
	}
	i.CalculateAmounts()
}

func (i *Invoice) GetLineItemCount() int {
	return len(i.LineItems)
}

// Status transition methods
func (i *Invoice) Send() error {
	if i.Status != InvoiceStatusDraft {
		return fmt.Errorf("can only send draft invoices")
	}
	i.Status = InvoiceStatusSent
	now := time.Now()
	i.SentDate = &now
	return nil
}

func (i *Invoice) Cancel() error {
	if i.Status == InvoiceStatusPaid {
		return fmt.Errorf("cannot cancel paid invoices")
	}
	i.Status = InvoiceStatusCancelled
	return nil
}

func (i *Invoice) Void() error {
	if i.PaidAmount > 0 {
		return fmt.Errorf("cannot void invoices with payments")
	}
	i.Status = InvoiceStatusVoid
	return nil
}

// Implement Validator interface
func (i *Invoice) Validate() error {
	if i.Number == "" {
		return fmt.Errorf("invoice number is required")
	}
	if i.PartyID == 0 {
		return fmt.Errorf("party ID is required")
	}
	if i.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be greater than zero")
	}
	if i.IssueDate.After(i.DueDate) {
		return fmt.Errorf("due date cannot be before issue date")
	}
	if i.PaidAmount > i.TotalAmount {
		return fmt.Errorf("paid amount cannot exceed total amount")
	}
	return nil
}

// Implement Calculator interface
func (i *Invoice) Calculate() {
	i.CalculateAmounts()
}

func (Invoice) TableName() string {
	return "invoices"
}

// InvoiceLineItem represents individual items on an invoice
type InvoiceLineItem struct {
	BaseModel
	InvoiceID   uint    `json:"invoice_id" gorm:"not null;index" validate:"required"`
	LineNumber  int     `json:"line_number" gorm:"not null" validate:"required,gt=0"`
	Description string  `json:"description" gorm:"not null;size:500" validate:"required"`
	Quantity    float64 `json:"quantity" gorm:"not null" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" gorm:"not null" validate:"required,gt=0"`
	Amount      float64 `json:"amount" gorm:"not null" validate:"required,gt=0"`

	// Extension points
	ProductID     *uint   `json:"product_id,omitempty" gorm:"index"`      // Link to product catalog
	ServiceID     *uint   `json:"service_id,omitempty" gorm:"index"`      // Link to service catalog
	ProjectID     *uint   `json:"project_id,omitempty" gorm:"index"`      // For project billing
	AccountID     *uint   `json:"account_id,omitempty" gorm:"index"`      // For expense categorization
	CostCenterID  *uint   `json:"cost_center_id,omitempty" gorm:"index"`  // For cost tracking
	
	// Tax and discount
	TaxRate       float64 `json:"tax_rate" gorm:"default:0"`
	TaxAmount     float64 `json:"tax_amount" gorm:"default:0"`
	DiscountRate  float64 `json:"discount_rate" gorm:"default:0"`
	DiscountAmount float64 `json:"discount_amount" gorm:"default:0"`
	
	// Optional fields
	UnitOfMeasure string  `json:"unit_of_measure" gorm:"size:20"`
	Notes         string  `json:"notes" gorm:"size:500"`

	// Relationships
	Invoice *Invoice `json:"-" gorm:"foreignKey:InvoiceID"`
	Account *Account `json:"account,omitempty" gorm:"foreignKey:AccountID"`
}

// Amount calculation methods
func (ili *InvoiceLineItem) CalculateAmount() {
	// Calculate base amount
	baseAmount := ili.Quantity * ili.UnitPrice
	
	// Apply discount
	ili.DiscountAmount = baseAmount * (ili.DiscountRate / 100)
	discountedAmount := baseAmount - ili.DiscountAmount
	
	// Calculate tax
	ili.TaxAmount = discountedAmount * (ili.TaxRate / 100)
	
	// Final amount (excluding tax as it's handled at invoice level)
	ili.Amount = discountedAmount
}

func (ili *InvoiceLineItem) GetTotalWithTax() float64 {
	return ili.Amount + ili.TaxAmount
}

func (ili *InvoiceLineItem) GetTaxAmount() float64 {
	return ili.TaxAmount
}

func (ili *InvoiceLineItem) GetDiscountAmount() float64 {
	return ili.DiscountAmount
}

func (ili *InvoiceLineItem) GetBaseAmount() float64 {
	return ili.Quantity * ili.UnitPrice
}

// Business methods
func (ili *InvoiceLineItem) UpdateQuantity(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	ili.Quantity = quantity
	ili.CalculateAmount()
	return nil
}

func (ili *InvoiceLineItem) UpdateUnitPrice(price float64) error {
	if price <= 0 {
		return fmt.Errorf("unit price must be greater than zero")
	}
	ili.UnitPrice = price
	ili.CalculateAmount()
	return nil
}

func (ili *InvoiceLineItem) ApplyDiscount(discountRate float64) error {
	if discountRate < 0 || discountRate > 100 {
		return fmt.Errorf("discount rate must be between 0 and 100")
	}
	ili.DiscountRate = discountRate
	ili.CalculateAmount()
	return nil
}

func (ili *InvoiceLineItem) SetTaxRate(taxRate float64) error {
	if taxRate < 0 || taxRate > 100 {
		return fmt.Errorf("tax rate must be between 0 and 100")
	}
	ili.TaxRate = taxRate
	ili.CalculateAmount()
	return nil
}

// Implement Validator interface
func (ili *InvoiceLineItem) Validate() error {
	if ili.InvoiceID == 0 {
		return fmt.Errorf("invoice ID is required")
	}
	if ili.Description == "" {
		return fmt.Errorf("description is required")
	}
	if ili.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if ili.UnitPrice <= 0 {
		return fmt.Errorf("unit price must be greater than zero")
	}
	if ili.LineNumber <= 0 {
		return fmt.Errorf("line number must be greater than zero")
	}
	return nil
}

// Implement Calculator interface
func (ili *InvoiceLineItem) Calculate() {
	ili.CalculateAmount()
}

func (InvoiceLineItem) TableName() string {
	return "invoice_line_items"
}

// Factory methods for creating invoices
func NewCustomerInvoice(number string, partyID uint, issueDate time.Time, dueDate time.Time) *Invoice {
	return &Invoice{
		Number:       number,
		Type:         InvoiceTypeReceivable,
		PartyID:      partyID,
		IssueDate:    issueDate,
		DueDate:      dueDate,
		Status:       InvoiceStatusDraft,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
		LineItems:    make([]InvoiceLineItem, 0),
	}
}

func NewVendorBill(number string, partyID uint, issueDate time.Time, dueDate time.Time) *Invoice {
	return &Invoice{
		Number:       number,
		Type:         InvoiceTypePayable,
		PartyID:      partyID,
		IssueDate:    issueDate,
		DueDate:      dueDate,
		Status:       InvoiceStatusDraft,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
		LineItems:    make([]InvoiceLineItem, 0),
	}
}