type Account struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"unique;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Type        string    `json:"type"` // Asset, Liability, Equity, Revenue, Expense
	ParentID    *uint     `json:"parent_id"`
	Balance     float64   `json:"balance" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}