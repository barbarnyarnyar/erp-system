// File: services/financial-management/models/party.go
package models

import "fmt"

// Party represents any entity we do business with (customers, vendors, employees)
// This unified model replaces separate Customer and Vendor models
type Party struct {
	BaseModel
	Auditable
	Code         string    `json:"code" gorm:"unique;not null;size:20" validate:"required"`
	Name         string    `json:"name" gorm:"not null;size:100" validate:"required"`
	Type         PartyType `json:"type" gorm:"not null" validate:"required"`
	Email        string    `json:"email" gorm:"size:100" validate:"omitempty,email"`
	Phone        string    `json:"phone" gorm:"size:20"`
	Address      string    `json:"address" gorm:"size:500"`
	TaxID        string    `json:"tax_id" gorm:"size:50"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`

	// Financial settings
	CreditLimit  *float64 `json:"credit_limit,omitempty" validate:"omitempty,min=0"`
	PaymentTerms *int     `json:"payment_terms,omitempty" validate:"omitempty,min=0,max=365"` // Days
	Balance      float64  `json:"balance" gorm:"default:0"` // Outstanding balance

	// Extension points for different party types
	Metadata PartyMetadata `json:"metadata" gorm:"type:jsonb"` // PostgreSQL JSONB for flexible data

	// Optional fields that can be used based on party type
	BankAccount     *string `json:"bank_account,omitempty" gorm:"size:100"`
	Category        *string `json:"category,omitempty" gorm:"size:50"`        // Vendor category
	Industry        *string `json:"industry,omitempty" gorm:"size:50"`        // Customer industry
	CompanySize     *string `json:"company_size,omitempty" gorm:"size:20"`    // Small, Medium, Large
	PreferredCurrency *string `json:"preferred_currency,omitempty" gorm:"size:3"`

	// Relationships
	Invoices []Invoice `json:"-" gorm:"foreignKey:PartyID"`
	Payments []Payment `json:"-" gorm:"foreignKey:PartyID"`
}

// PartyType enum for different types of business relationships
type PartyType string

const (
	PartyTypeCustomer PartyType = "customer"
	PartyTypeVendor   PartyType = "vendor"
	PartyTypeEmployee PartyType = "employee"
	PartyTypeBank     PartyType = "bank"
	PartyTypeBoth     PartyType = "both" // Can be both customer and vendor
)

// PartyMetadata allows flexible extension without schema changes
// Examples: {"industry": "tech", "annual_revenue": 1000000, "contact_person": "John Doe"}
type PartyMetadata map[string]interface{}

// Business methods for party type checking
func (p *Party) IsCustomer() bool {
	return p.Type == PartyTypeCustomer || p.Type == PartyTypeBoth
}

func (p *Party) IsVendor() bool {
	return p.Type == PartyTypeVendor || p.Type == PartyTypeBoth
}

func (p *Party) IsEmployee() bool {
	return p.Type == PartyTypeEmployee
}

func (p *Party) IsBank() bool {
	return p.Type == PartyTypeBank
}

// Credit management methods
func (p *Party) HasCreditLimit() bool {
	return p.CreditLimit != nil && *p.CreditLimit > 0
}

func (p *Party) GetCreditAvailable() float64 {
	if !p.HasCreditLimit() {
		return 0 // No limit set
	}
	available := *p.CreditLimit - p.Balance
	if available < 0 {
		return 0
	}
	return available
}

func (p *Party) IsOverCreditLimit() bool {
	if !p.HasCreditLimit() {
		return false
	}
	return p.Balance > *p.CreditLimit
}

func (p *Party) GetCreditUtilization() float64 {
	if !p.HasCreditLimit() {
		return 0
	}
	return (p.Balance / *p.CreditLimit) * 100
}

// Payment terms methods
func (p *Party) GetPaymentTermsDescription() string {
	if p.PaymentTerms == nil || *p.PaymentTerms == 0 {
		return "Due on Receipt"
	}
	return fmt.Sprintf("Net %d days", *p.PaymentTerms)
}

func (p *Party) GetPaymentTermsDays() int {
	if p.PaymentTerms == nil {
		return 30 // Default to 30 days
	}
	return *p.PaymentTerms
}

// Status and priority methods
func (p *Party) GetCreditStatus() string {
	if !p.HasCreditLimit() {
		return "No Limit"
	}
	if p.IsOverCreditLimit() {
		return "Over Limit"
	}
	utilization := p.GetCreditUtilization()
	switch {
	case utilization >= 90:
		return "Near Limit"
	case utilization >= 75:
		return "High Usage"
	case utilization >= 50:
		return "Moderate Usage"
	default:
		return "Good Standing"
	}
}

func (p *Party) GetPaymentPriority() string {
	if p.Balance <= 0 {
		return "None"
	}
	switch {
	case p.Balance >= 10000:
		return "High"
	case p.Balance >= 5000:
		return "Medium"
	default:
		return "Low"
	}
}

// Metadata helper methods
func (p *Party) SetMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(PartyMetadata)
	}
	p.Metadata[key] = value
}

func (p *Party) GetMetadata(key string) (interface{}, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, exists := p.Metadata[key]
	return value, exists
}

func (p *Party) GetMetadataString(key string) string {
	if value, exists := p.GetMetadata(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (p *Party) GetMetadataFloat64(key string) float64 {
	if value, exists := p.GetMetadata(key); exists {
		if num, ok := value.(float64); ok {
			return num
		}
	}
	return 0
}

// Business logic methods
func (p *Party) CanMakeCredit(amount float64) bool {
	if !p.IsCustomer() {
		return false
	}
	if !p.HasCreditLimit() {
		return true // No limit means unlimited credit
	}
	return (p.Balance + amount) <= *p.CreditLimit
}

func (p *Party) UpdateBalance(amount float64, isReceivable bool) {
	if isReceivable {
		p.Balance += amount // Customer owes us more
	} else {
		p.Balance -= amount // We owe vendor less / receive payment
	}
}

// Implement Validator interface
func (p *Party) Validate() error {
	if p.Code == "" {
		return fmt.Errorf("party code is required")
	}
	if p.Name == "" {
		return fmt.Errorf("party name is required")
	}
	if p.Type == "" {
		return fmt.Errorf("party type is required")
	}
	if p.CreditLimit != nil && *p.CreditLimit < 0 {
		return fmt.Errorf("credit limit cannot be negative")
	}
	if p.PaymentTerms != nil && (*p.PaymentTerms < 0 || *p.PaymentTerms > 365) {
		return fmt.Errorf("payment terms must be between 0 and 365 days")
	}
	return nil
}

func (Party) TableName() string {
	return "parties"
}

// Factory methods for creating specific party types
func NewCustomer(code, name string) *Party {
	return &Party{
		Code:         code,
		Name:         name,
		Type:         PartyTypeCustomer,
		IsActive:     true,
		CreditLimit:  Float64Ptr(10000), // Default credit limit
		PaymentTerms: IntPtr(30),         // Default payment terms
		Metadata:     make(PartyMetadata),
	}
}

func NewVendor(code, name string) *Party {
	return &Party{
		Code:         code,
		Name:         name,
		Type:         PartyTypeVendor,
		IsActive:     true,
		PaymentTerms: IntPtr(30), // Default payment terms
		Metadata:     make(PartyMetadata),
	}
}

func NewEmployee(code, name string) *Party {
	return &Party{
		Code:     code,
		Name:     name,
		Type:     PartyTypeEmployee,
		IsActive: true,
		Metadata: make(PartyMetadata),
	}
}