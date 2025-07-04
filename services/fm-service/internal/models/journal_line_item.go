type JournalLineItem struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	JournalEntryID  uint    `json:"journal_entry_id"`
	AccountID       uint    `json:"account_id"`
	Debit          float64 `json:"debit" gorm:"default:0"`
	Credit         float64 `json:"credit" gorm:"default:0"`
	Description    string  `json:"description"`
	Account        Account `json:"account" gorm:"foreignKey:AccountID"`
}