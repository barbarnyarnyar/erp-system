import journal_entry_id;

type JournalEntry struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Reference   string            `json:"reference" gorm:"unique"`
	Description string            `json:"description"`
	Date        time.Time         `json:"date"`
	TotalDebit  float64           `json:"total_debit"`
	TotalCredit float64           `json:"total_credit"`
	Status      string            `json:"status"` // Draft, Posted, Reversed
	SourceType  string            `json:"source_type"` // HR, SCM, M, CRM, PM
	SourceID    string            `json:"source_id"`
	LineItems   []JournalLineItem `json:"line_items" gorm:"foreignKey:JournalEntryID"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}