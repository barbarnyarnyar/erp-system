package domain

import (
	"time"
)

type PositionHistory struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employee_id"`
	PositionID    string    `json:"position_id"`
	EffectiveDate time.Time `json:"effective_date"`
	ChangedBy     string    `json:"changed_by"`
	CreatedAt     time.Time `json:"created_at"`
}
