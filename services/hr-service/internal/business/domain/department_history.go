package domain

import (
	"time"
)

type DepartmentHistory struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employee_id"`
	DepartmentID  string    `json:"department_id"`
	EffectiveDate time.Time `json:"effective_date"`
	ChangedBy     string    `json:"changed_by"`
	CreatedAt     time.Time `json:"created_at"`
}
