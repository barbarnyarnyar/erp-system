// File: services/financial-management/models/journal.go
package models

import (
	"fmt"
	"time"
)

// JournalEntry represents the header of a financial transaction (double-entry bookkeeping)
type JournalEntry struct {
	BaseModel
	Auditable
	Trackable
	Number      string        `json:"number" gorm:"unique;not null;size:50" validate:"required"`
	Date        time.Time     `json:"date" gorm:"not null" validate:"required"`
	Description string        `json:"description" gorm:"not null;size:500" validate:"required"`
	TotalDebit  float64       `json:"total_debit" gorm:"default:0"`
	TotalCredit float64       `json:"total_credit" gorm:"default:0"`
	Status      JournalStatus `json:"status" gorm:"default:'draft'"`
	
	// Audit trail for posting and reversing
	PostedBy    string     `json:"posted_by" gorm:"size:50"`
	PostedAt    *time.Time `json:"posted_at,omitempty"`
	ReversedBy  string     `json:"reversed_by" gorm:"size:50"`
	ReversedAt  *time.Time `json:"reversed_at,omitempty"`
	
	// Optional fields
	Notes       string `json:"notes" gorm:"size:1000"`
	
	// Extension points
	PeriodID    *uint  `json:"period_id,omitempty" gorm:"index"`    // Accounting period
	BatchID     *uint  `json:"batch_id,omitempty" gorm:"index"`     // For batch processing
	
	// Currency support
	Currency     string  `json:"currency" gorm:"default:'USD';size:3"`
	ExchangeRate float64 `json:"exchange_rate" gorm:"default:1.0"`

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

// Status checking methods
func (je *JournalEntry) IsDraft() bool {
	return je.Status == JournalStatusDraft
}

func (je *JournalEntry) IsPosted() bool {
	return je.Status == JournalStatusPosted
}

func (je *JournalEntry) IsReversed() bool {
	return je.Status == JournalStatusReversed
}

func (je *JournalEntry) CanBeEdited() bool {
	return je.Status == JournalStatusDraft
}

func (je *JournalEntry) CanBePosted() bool {
	return je.Status == JournalStatusDraft && len(je.LineItems) >= 2 && je.IsBalanced()
}

func (je *JournalEntry) CanBeReversed() bool {
	return je.Status == JournalStatusPosted
}

// Balance checking methods
func (je *JournalEntry) IsBalanced() bool {
	// Allow for small rounding differences
	diff := je.TotalDebit - je.TotalCredit
	return diff > -0.01 && diff < 0.01
}

func (je *JournalEntry) GetBalanceDifference() float64 {
	return je.TotalDebit - je.TotalCredit
}

// Calculation methods
func (je *JournalEntry) CalculateTotals() {
	je.TotalDebit = 0
	je.TotalCredit = 0
	
	for _, line := range je.LineItems {
		je.TotalDebit += line.Debit
		je.TotalCredit += line.Credit
	}
}

func (je *JournalEntry) RecalculateAndValidate() error {
	je.CalculateTotals()
	return je.Validate()
}

// Line item management
func (je *JournalEntry) AddLineItem(lineItem *JournalLineItem) error {
	if err := lineItem.Validate(); err != nil {
		return fmt.Errorf("invalid line item: %w", err)
	}
	
	lineItem.JournalEntryID = je.ID
	je.LineItems = append(je.LineItems, *lineItem)
	je.CalculateTotals()
	return nil
}

func (je *JournalEntry) RemoveLineItem(lineItemID uint) error {
	for idx, line := range je.LineItems {
		if line.ID == lineItemID {
			je.LineItems = append(je.LineItems[:idx], je.LineItems[idx+1:]...)
			je.CalculateTotals()
			return nil
		}
	}
	return fmt.Errorf("line item with ID %d not found", lineItemID)
}

func (je *JournalEntry) GetLineItemCount() int {
	return len(je.LineItems)
}

func (je *JournalEntry) ClearLineItems() {
	je.LineItems = make([]JournalLineItem, 0)
	je.TotalDebit = 0
	je.TotalCredit = 0
}

// Status transition methods
func (je *JournalEntry) Post(postedBy string) error {
	if !je.CanBePosted() {
		return fmt.Errorf("journal entry cannot be posted in current state")
	}
	
	// Final validation before posting
	je.CalculateTotals()
	if !je.IsBalanced() {
		return fmt.Errorf("journal entry must be balanced before posting: debits=%.2f, credits=%.2f", 
			je.TotalDebit, je.TotalCredit)
	}
	
	if len(je.LineItems) < 2 {
		return fmt.Errorf("journal entry must have at least 2 line items")
	}
	
	// Validate each line item
	for i, line := range je.LineItems {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line item %d is invalid: %w", i+1, err)
		}
	}
	
	now := time.Now()
	je.Status = JournalStatusPosted
	je.PostedBy = postedBy
	je.PostedAt = &now
	
	return nil
}

func (je *JournalEntry) Reverse(reversedBy string, reason string) error {
	if !je.CanBeReversed() {
		return fmt.Errorf("journal entry cannot be reversed in current status: %s", je.Status)
	}
	
	now := time.Now()
	je.Status = JournalStatusReversed
	je.ReversedBy = reversedBy
	je.ReversedAt = &now
	
	if reason != "" {
		je.Notes = fmt.Sprintf("%s\n[REVERSED: %s]", je.Notes, reason)
	}
	
	return nil
}

// Analysis methods
func (je *JournalEntry) GetAffectedAccounts() []uint {
	accountIDs := make([]uint, 0, len(je.LineItems))
	seen := make(map[uint]bool)
	
	for _, line := range je.LineItems {
		if !seen[line.AccountID] {
			accountIDs = append(accountIDs, line.AccountID)
			seen[line.AccountID] = true
		}
	}
	
	return accountIDs
}

func (je *JournalEntry) GetDebitLineItems() []JournalLineItem {
	var debitLines []JournalLineItem
	for _, line := range je.LineItems {
		if line.IsDebit() {
			debitLines = append(debitLines, line)
		}
	}
	return debitLines
}

func (je *JournalEntry) GetCreditLineItems() []JournalLineItem {
	var creditLines []JournalLineItem
	for _, line := range je.LineItems {
		if line.IsCredit() {
			creditLines = append(creditLines, line)
		}
	}
	return creditLines
}

func (je *JournalEntry) GetSourceDescription() string {
	if je.SourceService == "" {
		return "Manual Entry"
	}
	return fmt.Sprintf("%s - %s", je.SourceService, je.SourceID)
}

// Timing methods
func (je *JournalEntry) GetPostingDelay() time.Duration {
	if je.PostedAt == nil {
		return 0
	}
	return je.PostedAt.Sub(je.CreatedAt)
}

func (je *JournalEntry) IsPostedSameDay() bool {
	if je.PostedAt == nil {
		return false
	}
	return je.Date.Format("2006-01-02") == je.PostedAt.Format("2006-01-02")
}

// Implement Validator interface
func (je *JournalEntry) Validate() error {
	if je.Number == "" {
		return fmt.Errorf("journal entry number is required")
	}
	if je.Description == "" {
		return fmt.Errorf("journal entry description is required")
	}
	if len(je.LineItems) < 2 {
		return fmt.Errorf("journal entry must have at least 2 line items")
	}
	
	je.CalculateTotals()
	if !je.IsBalanced() {
		return fmt.Errorf("journal entry must be balanced: debits (%.2f) must equal credits (%.2f)", 
			je.TotalDebit, je.TotalCredit)
	}
	
	// Validate all line items
	for i, line := range je.LineItems {
		if err := line.Validate(); err != nil {
			return fmt.Errorf("line item %d: %w", i+1, err)
		}
	}
	
	return nil
}

// Implement Calculator interface
func (je *JournalEntry) Calculate() {
	je.CalculateTotals()
}

func (JournalEntry) TableName() string {
	return "journal_entries"
}

// JournalLineItem represents individual debit/credit lines in a journal entry
type JournalLineItem struct {
	BaseModel
	JournalEntryID uint    `json:"journal_entry_id" gorm:"not null;index" validate:"required"`
	LineNumber     int     `json:"line_number" gorm:"not null" validate:"required,gt=0"`
	AccountID      uint    `json:"account_id" gorm:"not null;index" validate:"required"`
	Debit          float64 `json:"debit" gorm:"default:0" validate:"min=0"`
	Credit         float64 `json:"credit" gorm:"default:0" validate:"min=0"`
	Description    string  `json:"description" gorm:"size:500"`

	// Extension points for dimensional accounting
	CostCenterID *uint  `json:"cost_center_id,omitempty" gorm:"index"`    // For cost tracking
	ProjectID    *uint  `json:"project_id,omitempty" gorm:"index"`        // For project accounting
	DepartmentID string `json:"department_id,omitempty" gorm:"size:20;index"` // From HR service
	LocationID   *uint  `json:"location_id,omitempty" gorm:"index"`       // For multi-location businesses
	
	// Additional tracking
	Memo         string `json:"memo" gorm:"size:255"`                     // Additional notes
	DocumentRef  string `json:"document_ref" gorm:"size:100"`             // Reference to supporting document
	
	// Currency support (for multi-currency transactions)
	Currency         string  `json:"currency" gorm:"default:'USD';size:3"`
	ExchangeRate     float64 `json:"exchange_rate" gorm:"default:1.0"`
	OriginalDebit    float64 `json:"original_debit" gorm:"default:0"`      // Amount in original currency
	OriginalCredit   float64 `json:"original_credit" gorm:"default:0"`     // Amount in original currency

	// Relationships
	JournalEntry *JournalEntry `json:"-" gorm:"foreignKey:JournalEntryID"`
	Account      *Account      `json:"account,omitempty" gorm:"foreignKey:AccountID"`
}

// Amount and type checking methods
func (jli *JournalLineItem) IsDebit() bool {
	return jli.Debit > 0
}

func (jli *JournalLineItem) IsCredit() bool {
	return jli.Credit > 0
}

func (jli *JournalLineItem) GetAmount() float64 {
	if jli.IsDebit() {
		return jli.Debit
	}
	return jli.Credit
}

func (jli *JournalLineItem) GetType() string {
	if jli.IsDebit() {
		return "Debit"
	}
	return "Credit"
}

func (jli *JournalLineItem) GetFormattedAmount() string {
	amount := jli.GetAmount()
	if jli.IsDebit() {
		return fmt.Sprintf("Dr. %.2f", amount)
	}
	return fmt.Sprintf("Cr. %.2f", amount)
}

// Amount setting methods
func (jli *JournalLineItem) SetDebit(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("debit amount cannot be negative")
	}
	jli.Debit = amount
	jli.Credit = 0 // Clear credit when setting debit
	return nil
}

func (jli *JournalLineItem) SetCredit(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("credit amount cannot be negative")
	}
	jli.Credit = amount
	jli.Debit = 0 // Clear debit when setting credit
	return nil
}

func (jli *JournalLineItem) SetAmount(amount float64, isDebit bool) error {
	if amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}
	
	if isDebit {
		return jli.SetDebit(amount)
	}
	return jli.SetCredit(amount)
}

// Dimensional accounting methods
func (jli *JournalLineItem) HasDimensions() bool {
	return jli.CostCenterID != nil || 
		   jli.ProjectID != nil || 
		   jli.DepartmentID != "" || 
		   jli.LocationID != nil
}

func (jli *JournalLineItem) GetDimensionInfo() string {
	var dimensions []string
	
	if jli.CostCenterID != nil {
		dimensions = append(dimensions, fmt.Sprintf("CC:%d", *jli.CostCenterID))
	}
	if jli.ProjectID != nil {
		dimensions = append(dimensions, fmt.Sprintf("Proj:%d", *jli.ProjectID))
	}
	if jli.DepartmentID != "" {
		dimensions = append(dimensions, fmt.Sprintf("Dept:%s", jli.DepartmentID))
	}
	if jli.LocationID != nil {
		dimensions = append(dimensions, fmt.Sprintf("Loc:%d", *jli.LocationID))
	}
	
	if len(dimensions) == 0 {
		return "No Dimensions"
	}
	
	return fmt.Sprintf("[%s]", fmt.Sprintf("%v", dimensions))
}

func (jli *JournalLineItem) SetCostCenter(costCenterID uint) {
	jli.CostCenterID = UintPtr(costCenterID)
}

func (jli *JournalLineItem) SetProject(projectID uint) {
	jli.ProjectID = UintPtr(projectID)
}

func (jli *JournalLineItem) SetDepartment(departmentID string) {
	jli.DepartmentID = departmentID
}

func (jli *JournalLineItem) SetLocation(locationID uint) {
	jli.LocationID = UintPtr(locationID)
}

// Currency methods
func (jli *JournalLineItem) IsMultiCurrency() bool {
	return jli.Currency != DefaultCurrency || jli.ExchangeRate != DefaultExchangeRate
}

func (jli *JournalLineItem) SetOriginalCurrency(currency string, exchangeRate float64, originalAmount float64, isDebit bool) error {
	if exchangeRate <= 0 {
		return fmt.Errorf("exchange rate must be positive")
	}
	
	jli.Currency = currency
	jli.ExchangeRate = exchangeRate
	
	if isDebit {
		jli.OriginalDebit = originalAmount
		jli.OriginalCredit = 0
		jli.Debit = originalAmount * exchangeRate
		jli.Credit = 0
	} else {
		jli.OriginalCredit = originalAmount
		jli.OriginalDebit = 0
		jli.Credit = originalAmount * exchangeRate
		jli.Debit = 0
	}
	
	return nil
}

func (jli *JournalLineItem) GetOriginalAmount() float64 {
	if jli.OriginalDebit > 0 {
		return jli.OriginalDebit
	}
	return jli.OriginalCredit
}

// Description and memo methods
func (jli *JournalLineItem) GetFullDescription() string {
	if jli.Memo != "" {
		return fmt.Sprintf("%s - %s", jli.Description, jli.Memo)
	}
	return jli.Description
}

func (jli *JournalLineItem) SetMemo(memo string) {
	jli.Memo = memo
}

func (jli *JournalLineItem) SetDocumentReference(ref string) {
	jli.DocumentRef = ref
}

// Validation methods
func (jli *JournalLineItem) Validate() error {
	// Must have either debit or credit, but not both
	if jli.Debit > 0 && jli.Credit > 0 {
		return fmt.Errorf("line item cannot have both debit and credit amounts")
	}
	if jli.Debit == 0 && jli.Credit == 0 {
		return fmt.Errorf("line item must have either debit or credit amount")
	}
	if jli.JournalEntryID == 0 {
		return fmt.Errorf("journal entry ID is required")
	}
	if jli.AccountID == 0 {
		return fmt.Errorf("account ID is required")
	}
	if jli.LineNumber <= 0 {
		return fmt.Errorf("line number must be greater than zero")
	}
	
	// Currency validation
	if jli.ExchangeRate <= 0 {
		return fmt.Errorf("exchange rate must be positive")
	}
	
	return nil
}

func (jli *JournalLineItem) IsValid() bool {
	return jli.Validate() == nil
}

func (JournalLineItem) TableName() string {
	return "journal_line_items"
}

// Factory methods for creating journal entries
func NewJournalEntry(number string, date time.Time, description string) *JournalEntry {
	return &JournalEntry{
		Number:       number,
		Date:         date,
		Description:  description,
		Status:       JournalStatusDraft,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
		LineItems:    make([]JournalLineItem, 0),
	}
}

func NewJournalLineItem(accountID uint, amount float64, isDebit bool, description string) *JournalLineItem {
	lineItem := &JournalLineItem{
		AccountID:    accountID,
		Description:  description,
		Currency:     DefaultCurrency,
		ExchangeRate: DefaultExchangeRate,
	}
	
	if isDebit {
		lineItem.Debit = amount
	} else {
		lineItem.Credit = amount
	}
	
	return lineItem
}

// Helper method to create a simple two-line journal entry
func NewSimpleJournalEntry(number string, date time.Time, description string, 
	debitAccountID uint, creditAccountID uint, amount float64) *JournalEntry {
	
	entry := NewJournalEntry(number, date, description)
	
	// Add debit line
	debitLine := NewJournalLineItem(debitAccountID, amount, true, description)
	debitLine.LineNumber = 1
	entry.LineItems = append(entry.LineItems, *debitLine)
	
	// Add credit line
	creditLine := NewJournalLineItem(creditAccountID, amount, false, description)
	creditLine.LineNumber = 2
	entry.LineItems = append(entry.LineItems, *creditLine)
	
	entry.CalculateTotals()
	return entry
}